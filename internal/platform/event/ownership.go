package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

var ErrReplicaOwnership = errors.New("event replica ownership is unavailable or stale")

var ownerTablePath = regexp.MustCompile(`^(/[a-zA-Z0-9_][a-zA-Z0-9_.-]*)+$`)
var kubernetesName = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

// ReplicaSlot identifies a Pod name within a particular StatefulSet incarnation.
// All values are trusted deployment configuration.
type ReplicaSlot struct {
	Namespace      string
	StatefulSetUID string
	PodName        string
	TablePath      string
}

func (s ReplicaSlot) validate() error {
	if len(s.Namespace) > 63 || len(s.PodName) > 253 ||
		!kubernetesName.MatchString(s.Namespace) || !kubernetesName.MatchString(s.PodName) ||
		!validPodUID(s.StatefulSetUID) || !ownerTablePath.MatchString(s.TablePath) {
		return ErrInvalidConfiguration
	}
	return nil
}

func validPodUID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}

func (s ReplicaSlot) key() string {
	return s.Namespace + "/" + s.StatefulSetUID + "/" + s.PodName
}

const declareReplicaKey = "DECLARE $replica AS Utf8; "

func (s ReplicaSlot) read(ctx context.Context, tx query.TxActor) (uid string, generation uint64, err error) {
	row, err := tx.QueryRow(ctx, declareReplicaKey+"SELECT pod_uid, generation FROM `"+s.TablePath+"` WHERE replica = $replica;",
		query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(s.key()).Build()))
	if errors.Is(err, query.ErrNoRows) {
		return "", 0, nil
	}
	if err != nil {
		return "", 0, err
	}
	if err := row.Scan(&uid, &generation); err != nil {
		return "", 0, err
	}
	if !validPodUID(uid) || generation == 0 {
		return "", 0, ErrReplicaOwnership
	}
	return uid, generation, nil
}

// Reconcile reads the current Pod from Kubernetes and assigns its generation.
// Run it with a dedicated Kubernetes reader and the only YDB credentials
// allowed to update the ownership table. The HTTP client supplies TLS and
// authentication.
func (s ReplicaSlot) Reconcile(ctx context.Context, owner *Transactor, client *http.Client, apiServer string) error {
	if err := s.validate(); err != nil {
		return err
	}
	base, err := url.Parse(apiServer)
	if err != nil || base.Scheme != "https" || base.Hostname() == "" || base.User != nil ||
		base.Path != "" && base.Path != "/" || base.RawQuery != "" || base.Fragment != "" || client == nil {
		return ErrInvalidConfiguration
	}
	endpoint := base.JoinPath("api", "v1", "namespaces", s.Namespace, "pods", s.PodName).String()
	// Observe YDB before Kubernetes. An older Kubernetes observation cannot
	// overwrite a newer, committed handoff: the YDB read is invalidated and
	// the SDK retry performs a fresh Kubernetes read.
	return owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		prior, generation, err := s.read(ctx, tx)
		if err != nil {
			return err
		}
		current, err := s.currentPodUID(ctx, client, endpoint)
		if err != nil {
			return err
		}
		if current == prior {
			return nil
		}
		if generation == math.MaxUint64 {
			return ErrReplicaOwnership
		}
		return tx.Exec(ctx, declareReplicaKey+"DECLARE $uid AS Utf8; DECLARE $gen AS Uint64; "+
			"UPSERT INTO `"+s.TablePath+"` (replica, pod_uid, generation) VALUES ($replica, $uid, $gen);",
			query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(s.key()).
				Param("$uid").Text(current).Param("$gen").Uint64(generation+1).Build()))
	})
}

func (s ReplicaSlot) currentPodUID(ctx context.Context, client *http.Client, endpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	kubeClient := *client
	kubeClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := kubeClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("read Kubernetes Pod: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: Kubernetes Pod returned HTTP %d", ErrReplicaOwnership, resp.StatusCode)
	}
	var pod struct {
		Metadata struct {
			UID, Name, Namespace string
			DeletionTimestamp    *string
			OwnerReferences      []struct {
				UID, Kind, APIVersion string
				Controller            bool
			}
		}
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&pod); err != nil {
		return "", fmt.Errorf("decode Kubernetes Pod: %w", err)
	}
	m := pod.Metadata
	if m.Name != s.PodName || m.Namespace != s.Namespace || m.DeletionTimestamp != nil || !validPodUID(m.UID) {
		return "", ErrReplicaOwnership
	}
	for _, ref := range m.OwnerReferences {
		if ref.Controller && ref.APIVersion == "apps/v1" &&
			ref.Kind == "StatefulSet" && ref.UID == s.StatefulSetUID {
			return m.UID, nil
		}
	}
	return "", ErrReplicaOwnership
}

// PodOwnership is an immutable generation bound to the Pod UID supplied by
// the Downward API. An old process can never bind to a different UID.
type PodOwnership struct {
	slot       ReplicaSlot
	uid        string
	generation uint64
}

// Bind reads ownership using publisher credentials. The podUID must come
// from the publisher's own Downward API, fixed at process startup.
func (s ReplicaSlot) Bind(ctx context.Context, reader *Transactor, podUID string) (PodOwnership, error) {
	if err := s.validate(); err != nil || !validPodUID(podUID) {
		return PodOwnership{}, ErrInvalidConfiguration
	}
	var bound PodOwnership
	err := reader.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		uid, generation, err := s.read(ctx, tx)
		if err != nil {
			return err
		}
		if uid != podUID {
			return ErrReplicaOwnership
		}
		bound = PodOwnership{slot: s, uid: uid, generation: generation}
		return nil
	})
	if err != nil {
		return PodOwnership{}, err
	}
	return bound, nil
}

// Check must run in every publishing transaction attempt before any writer
// opens. Keep the same ownership value across retries.
func (o PodOwnership) Check(ctx context.Context, tx query.TxActor) error {
	if o.generation == 0 || tx == nil {
		return ErrReplicaOwnership
	}
	uid, generation, err := o.slot.read(ctx, tx)
	if err != nil {
		return err
	}
	if uid != o.uid || generation != o.generation {
		return ErrReplicaOwnership
	}
	return nil
}

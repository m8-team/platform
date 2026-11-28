package event

import (
	"context"
	"errors"
	"math"
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

const retiredReplica = "retired"

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
	if generation == 0 || (uid != "" && uid != retiredReplica && !validPodUID(uid)) {
		return "", 0, ErrReplicaOwnership
	}
	return uid, generation, nil
}

// Reconcile applies an authoritative observation using the ownership writer.
// observe runs once per SDK attempt after reading the current generation.
// An empty UID marks a vacant slot. A retired incarnation is permanent.
func (s ReplicaSlot) Reconcile(ctx context.Context, owner *Transactor, observe func(context.Context) (uid string, retired bool, err error)) error {
	if err := s.validate(); err != nil || observe == nil {
		return ErrInvalidConfiguration
	}
	var handoff, retirement bool
	err := owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		handoff, retirement = false, false
		prior, generation, err := s.read(ctx, tx)
		if err != nil {
			return err
		}
		if prior == retiredReplica {
			return nil
		}
		current, retired, err := observe(ctx)
		if err != nil {
			return err
		}
		if current != "" && !validPodUID(current) {
			return ErrReplicaOwnership
		}
		if retired {
			current = retiredReplica
		}
		if current == prior {
			return nil
		}
		if generation == math.MaxUint64 {
			return ErrReplicaOwnership
		}
		handoff = validPodUID(prior) && validPodUID(current)
		retirement = current == retiredReplica || current == ""
		return tx.Exec(ctx, declareReplicaKey+"DECLARE $uid AS Utf8; DECLARE $gen AS Uint64; "+
			"UPSERT INTO `"+s.TablePath+"` (replica, pod_uid, generation) VALUES ($replica, $uid, $gen);",
			query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(s.key()).
				Param("$uid").Text(current).Param("$gen").Uint64(generation+1).Build()))
	})
	if err == nil && handoff {
		owner.counters.handoffs.Add(1)
	}
	if err == nil && retirement {
		owner.counters.retirements.Add(1)
	}
	return err
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

// check must run in every publishing transaction attempt before any writer
// opens. Keep the same ownership value across retries.
func (o PodOwnership) check(ctx context.Context, tx query.TxActor) error {
	if !validPodUID(o.uid) || o.generation == 0 || tx == nil {
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

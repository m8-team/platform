package event

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
)

var ErrReplicaOwnership = errors.New("event replica ownership is unavailable or stale")

const (
	ownersTable = "/replica_owners"
	guardsTable = "/replica_publication_guards"
)

var ownerTablePath = regexp.MustCompile(`^(/[a-zA-Z0-9_][a-zA-Z0-9_.-]*)+$`)
var kubernetesName = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

// ReplicaSlot identifies a Pod name within a particular StatefulSet incarnation.
// All values are trusted deployment configuration.
type ReplicaSlot struct {
	Namespace      string
	StatefulSetUID string
	PodName        string
	tablePath      string
}

func (s ReplicaSlot) validate() error {
	if len(s.Namespace) > 63 || len(s.PodName) > 253 ||
		!kubernetesName.MatchString(s.Namespace) || !kubernetesName.MatchString(s.PodName) ||
		!validPodUID(s.StatefulSetUID) {
		return ErrInvalidConfiguration
	}
	return nil
}

func (s ReplicaSlot) withOwnershipTable(owner *Transactor) (ReplicaSlot, error) {
	if s.validate() != nil || owner == nil || owner.db == nil {
		return ReplicaSlot{}, ErrInvalidConfiguration
	}
	if s.tablePath == "" {
		database := owner.db.Name()
		if !ownerTablePath.MatchString(database) {
			return ReplicaSlot{}, ErrInvalidConfiguration
		}
		s.tablePath = database + ownersTable
	}
	if !ownerTablePath.MatchString(s.tablePath) {
		return ReplicaSlot{}, ErrInvalidConfiguration
	}
	return s, nil
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
	row, err := tx.QueryRow(ctx, declareReplicaKey+"SELECT pod_uid, generation FROM `"+s.tablePath+"` WHERE replica = $replica;",
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

var ErrPredecessorInventory = errors.New("incomplete event predecessor inventory")

// ActivateReplicaSet records the incarnation authorized to reconcile slots. On
// a transition it verifies the retirement barrier against canonical YDB rows.
func (t *Transactor) ActivateReplicaSet(ctx context.Context, namespace, name, uid string, bootstrap bool, predecessors map[string]int, isCurrent func(context.Context) (bool, error)) error {
	if isCurrent == nil || !kubernetesName.MatchString(name) || !kubernetesName.MatchString(namespace) || !validPodUID(uid) {
		return ErrInvalidConfiguration
	}
	controller, err := (ReplicaSlot{Namespace: namespace}).ownershipController(t, name)
	if err != nil {
		return err
	}
	return t.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
		previous, generation, err := controller.read(ctx, tx)
		if err != nil || previous == uid {
			return err
		}
		if previous == "" && !bootstrap {
			return fmt.Errorf("%w: canonical ownership history has not been attested for bootstrap", ErrPredecessorInventory)
		}
		if previous != "" && predecessors[previous] <= 0 {
			return fmt.Errorf("%w: previous incarnation missing", ErrPredecessorInventory)
		}
		if err := controller.checkRetirement(ctx, tx, namespace, name, uid, predecessors); err != nil {
			return err
		}
		current, err := isCurrent(ctx)
		if err != nil {
			return err
		}
		if !current || generation == math.MaxUint64 {
			return ErrReplicaOwnership
		}
		return controller.write(ctx, tx, uid, generation+1)
	})
}

// The empty UID component distinguishes controller rows from Pod slot rows.
func (s ReplicaSlot) ownershipController(t *Transactor, name string) (ReplicaSlot, error) {
	if t == nil || t.db == nil || !ownerTablePath.MatchString(t.db.Name()) {
		return ReplicaSlot{}, ErrInvalidConfiguration
	}
	return ReplicaSlot{Namespace: s.Namespace, PodName: name, tablePath: t.db.Name() + ownersTable}, nil
}

func (s ReplicaSlot) checkRetirement(ctx context.Context, tx query.TxActor, namespace, name, uid string, predecessors map[string]int) error {
	rows, err := tx.QueryResultSet(ctx, "DECLARE $from AS Utf8; DECLARE $until AS Utf8; "+
		"SELECT replica, pod_uid FROM `"+s.tablePath+"` WHERE replica >= $from AND replica < $until;",
		query.WithParameters(ydb.ParamsBuilder().Param("$from").Text(namespace+"/").Param("$until").Text(namespace+"0").Build()))
	if err != nil {
		return err
	}
	for {
		row, err := rows.NextRow(ctx)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		var key, binding string
		if err := row.Scan(&key, &binding); err != nil {
			return err
		}
		parts := strings.Split(key, "/")
		if len(parts) != 3 {
			continue
		}
		podSet, ordinal, err := statefulSetOrdinal(parts[2])
		if podSet != name {
			continue
		}
		if err != nil || !validPodUID(parts[1]) {
			return fmt.Errorf("%w: invalid stored slot %q", ErrPredecessorInventory, parts[2])
		}
		if parts[1] == uid {
			if binding == retiredReplica {
				return ErrReplicaOwnership
			}
			continue
		}
		slots, listed := predecessors[parts[1]]
		if binding != retiredReplica || (listed && ordinal >= slots) {
			return fmt.Errorf("%w: historical slot %q is unretired or outside the configured range", ErrPredecessorInventory, parts[2])
		}
	}
}

func statefulSetOrdinal(pod string) (string, int, error) {
	last := strings.LastIndexByte(pod, '-')
	if last <= 0 {
		return "", 0, ErrInvalidConfiguration
	}
	name := pod[:last]
	value, err := strconv.Atoi(pod[last+1:])
	if err != nil || value < 0 || strconv.Itoa(value) != pod[last+1:] {
		return name, 0, ErrInvalidConfiguration
	}
	return name, value, nil
}

func (s ReplicaSlot) write(ctx context.Context, tx query.TxActor, uid string, generation uint64) error {
	return tx.Exec(ctx, declareReplicaKey+"DECLARE $uid AS Utf8; DECLARE $gen AS Uint64; "+
		"UPSERT INTO `"+s.tablePath+"` (replica, pod_uid, generation) VALUES ($replica, $uid, $gen);",
		query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(s.key()).
			Param("$uid").Text(uid).Param("$gen").Uint64(generation).Build()))
}

// Reconcile applies an authoritative observation using the ownership writer.
// observe runs once per SDK attempt after reading the current generation.
// An empty UID marks a vacant slot. A retired incarnation is permanent.
func (s ReplicaSlot) Reconcile(ctx context.Context, owner *Transactor, observe func(context.Context) (uid string, retired bool, err error)) error {
	canonical := s.tablePath == ""
	s, err := s.withOwnershipTable(owner)
	if err != nil || observe == nil {
		return ErrInvalidConfiguration
	}
	var handoff, retirement bool
	err = owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
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
		if canonical && !retired {
			name, _, err := statefulSetOrdinal(s.PodName)
			if err != nil {
				return err
			}
			controller, err := s.ownershipController(owner, name)
			if err != nil {
				return err
			}
			active, _, err := controller.read(ctx, tx)
			if err != nil {
				return err
			}
			if active != s.StatefulSetUID {
				return ErrReplicaOwnership
			}
		}
		if current == prior {
			return nil
		}
		if generation == math.MaxUint64 || (generation == math.MaxUint64-1 && current != retiredReplica) {
			return ErrReplicaOwnership
		}
		handoff = validPodUID(prior) && validPodUID(current)
		retirement = current == retiredReplica || current == ""
		return s.write(ctx, tx, current, generation+1)
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
	s, err := s.withOwnershipTable(reader)
	if err != nil || !validPodUID(podUID) {
		return PodOwnership{}, ErrInvalidConfiguration
	}
	var bound PodOwnership
	err = reader.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
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
	if !validPodUID(o.uid) || o.generation == 0 || o.generation == math.MaxUint64 || tx == nil {
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

// Package filecheckpoint is a durable single-host checkpoint implementation.
// It intentionally does not claim HA safety across hosts or independent disks.
package filecheckpoint

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"time"

	"github.com/m8-team/platform/internal/access/app/projection"
)

var ErrCorrupt = errors.New("corrupt access projection checkpoint")

const maxCheckpointBytes = 4 << 20

type Store struct{ directory string }

func New(directory string) (*Store, error) {
	if directory == "" {
		return nil, errors.New("checkpoint directory required")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	return &Store{directory: directory}, nil
}

var _ projection.CheckpointStore = (*Store)(nil)
var _ projection.CheckpointReader = (*Store)(nil)

// LoadCheckpoint acquires the same lock used by writers so readers never
// observe an incomplete checkpoint.
func (s *Store) LoadCheckpoint(ctx context.Context, key projection.Key) (projection.Checkpoint, error) {
	var observed projection.Checkpoint
	err := s.WithLocked(ctx, key, func(cp projection.Checkpoint) (projection.Checkpoint, error) {
		observed = cp
		return cp, nil
	})
	return observed, err
}

func (s *Store) WithLocked(ctx context.Context, key projection.Key, fn func(projection.Checkpoint) (projection.Checkpoint, error)) error {
	if s == nil || fn == nil {
		return projection.ErrInvalidCheckpoint
	}
	if err := key.Validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	hash := sha256.Sum256([]byte(key.String()))
	name := hex.EncodeToString(hash[:])
	lock, err := os.OpenFile(filepath.Join(s.directory, name+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	for {
		err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if err != syscall.EWOULDBLOCK && err != syscall.EAGAIN {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	path := filepath.Join(s.directory, name+".json")
	cp := projection.Checkpoint{Key: key}
	file, err := os.Open(path)
	if err == nil {
		limited := io.LimitReader(file, maxCheckpointBytes+1)
		bytes, readErr := io.ReadAll(limited)
		closeErr := file.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		if len(bytes) > maxCheckpointBytes {
			return ErrCorrupt
		}
		if err := json.Unmarshal(bytes, &cp); err != nil {
			return fmt.Errorf("%w: %v", ErrCorrupt, err)
		}
		if cp.Key != key || cp.Validate() != nil {
			return ErrCorrupt
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	next, err := fn(cp)
	if err != nil {
		return err
	}
	if next.Key != key || next.Validate() != nil || next.Revision < cp.Revision {
		return projection.ErrInvalidCheckpoint
	}
	if reflect.DeepEqual(next, cp) {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	encoded, err := json.Marshal(next)
	if err != nil {
		return err
	}
	if len(encoded) > maxCheckpointBytes {
		return projection.ErrInvalidCheckpoint
	}
	tmp, err := os.CreateTemp(s.directory, ".checkpoint-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(encoded); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	dir, err := os.Open(s.directory)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

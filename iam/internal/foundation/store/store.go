package store

import (
	"context"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"github.com/m8platform/platform/iam/internal/foundation/protokit"
	"google.golang.org/protobuf/proto"
)

var ErrNotFound = errors.New("document not found")

const DefaultPageSize = 50

type StoredDocument struct {
	ID        string
	TenantID  string
	Payload   []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DocumentStore interface {
	GetDocument(ctx context.Context, table string, id string) (StoredDocument, error)
	UpsertDocument(ctx context.Context, table string, doc StoredDocument) error
	DeleteDocument(ctx context.Context, table string, id string) error
	ListDocuments(ctx context.Context, table string, tenantID string, offset int, limit int) ([]StoredDocument, string, error)
}

func EncodePageToken(offset int) string {
	if offset <= 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func DecodePageToken(token string) int {
	if token == "" {
		return 0
	}
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	offset, err := strconv.Atoi(string(decoded))
	if err != nil || offset < 0 {
		return 0
	}
	return offset
}

func LoadProto(ctx context.Context, store DocumentStore, table string, id string, target proto.Message) error {
	document, err := store.GetDocument(ctx, table, id)
	if err != nil {
		return err
	}
	return protokit.Unmarshal(document.Payload, target)
}

func SaveProto(ctx context.Context, store DocumentStore, table string, id string, tenantID string, message proto.Message, now time.Time) error {
	payload, err := protokit.Marshal(message)
	if err != nil {
		return err
	}
	return store.UpsertDocument(ctx, table, StoredDocument{
		ID:        id,
		TenantID:  tenantID,
		Payload:   payload,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	})
}

func ListProto[T proto.Message](ctx context.Context, store DocumentStore, table string, tenantID string, pageSize int, pageToken string, newItem func() T) ([]T, string, error) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	offset := DecodePageToken(pageToken)
	documents, next, err := store.ListDocuments(ctx, table, tenantID, offset, pageSize)
	if err != nil {
		return nil, "", err
	}

	items := make([]T, 0, len(documents))
	for _, document := range documents {
		item := newItem()
		if err := protokit.Unmarshal(document.Payload, item); err != nil {
			return nil, "", err
		}
		items = append(items, item)
	}
	return items, next, nil
}

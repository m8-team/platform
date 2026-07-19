package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/app/query"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
)

const (
	minimumPageTokenKeyLength = 32
	maximumPageTokenLength    = 1024
	pageTokenVersion          = 1
)

var (
	ErrInvalidOrganizationPageSize  = errors.New("invalid organization page size")
	ErrInvalidOrganizationPageToken = errors.New("invalid organization page token")
	ErrInvalidOrganizationFilter    = errors.New("invalid organization filter")
	ErrInvalidOrganizationOrderBy   = errors.New("invalid organization order_by")
)

type normalizedListRequest struct {
	PageSize           int                      `json:"page_size"`
	Filter             ports.OrganizationFilter `json:"filter"`
	Order              ports.OrganizationOrder  `json:"order"`
	ShowDeleted        bool                     `json:"show_deleted"`
	AuthorizationScope string                   `json:"authorization_scope"`
}

func normalizeListQuery(
	q query.ListOrganizations,
	authorizationScope string,
) (ports.ListOrganizationsOptions, string, error) {
	if q.PageSize < 0 || q.PageSize > ports.MaxOrganizationPageSize {
		return ports.ListOrganizationsOptions{}, "", fmt.Errorf(
			"%w: must be between 0 and %d",
			ErrInvalidOrganizationPageSize,
			ports.MaxOrganizationPageSize,
		)
	}
	if len(q.PageToken) > maximumPageTokenLength {
		return ports.ListOrganizationsOptions{}, "", fmt.Errorf(
			"%w: exceeds %d characters",
			ErrInvalidOrganizationPageToken,
			maximumPageTokenLength,
		)
	}

	filter, err := parseOrganizationFilter(q.Filter)
	if err != nil {
		return ports.ListOrganizationsOptions{}, "", err
	}
	filter.ShowDeleted = q.ShowDeleted
	order, err := parseOrganizationOrder(q.OrderBy)
	if err != nil {
		return ports.ListOrganizationsOptions{}, "", err
	}

	options := ports.ListOrganizationsOptions{
		Filter:   filter,
		Order:    order,
		PageSize: q.PageSize,
	}.WithDefaults()
	if err := options.Validate(); err != nil {
		return ports.ListOrganizationsOptions{}, "", err
	}

	canonical, err := json.Marshal(normalizedListRequest{
		PageSize:           options.PageSize,
		Filter:             options.Filter,
		Order:              options.Order,
		ShowDeleted:        q.ShowDeleted,
		AuthorizationScope: authorizationScope,
	})
	if err != nil {
		return ports.ListOrganizationsOptions{}, "", fmt.Errorf("marshal list request: %w", err)
	}
	digest := sha256.Sum256(canonical)

	return options, hex.EncodeToString(digest[:]), nil
}

func parseOrganizationOrder(raw string) (ports.OrganizationOrder, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ports.OrganizationOrder{
			Field:     ports.OrganizationOrderFieldID,
			Direction: ports.SortDirectionAscending,
		}, nil
	}
	if utf8.RuneCountInString(raw) > 128 {
		return ports.OrganizationOrder{}, fmt.Errorf("%w: exceeds 128 characters", ErrInvalidOrganizationOrderBy)
	}
	if strings.Contains(raw, ",") {
		return ports.OrganizationOrder{}, fmt.Errorf("%w: only one field is supported", ErrInvalidOrganizationOrderBy)
	}
	parts := strings.Fields(raw)
	if len(parts) < 1 || len(parts) > 2 {
		return ports.OrganizationOrder{}, fmt.Errorf("%w: expected field [asc|desc]", ErrInvalidOrganizationOrderBy)
	}

	field := ports.OrganizationOrderField(parts[0])
	if !field.IsValid() {
		return ports.OrganizationOrder{}, fmt.Errorf("%w: unsupported field %q", ErrInvalidOrganizationOrderBy, parts[0])
	}
	direction := ports.SortDirectionAscending
	if len(parts) == 2 {
		direction = ports.SortDirection(strings.ToLower(parts[1]))
		if !direction.IsValid() {
			return ports.OrganizationOrder{}, fmt.Errorf("%w: unsupported direction %q", ErrInvalidOrganizationOrderBy, parts[1])
		}
	}

	return ports.OrganizationOrder{Field: field, Direction: direction}, nil
}

type pageTokenCodec struct {
	key []byte
}

type pageTokenPayload struct {
	Version     int       `json:"v"`
	RequestHash string    `json:"request_hash"`
	ID          string    `json:"id"`
	Name        string    `json:"name,omitempty"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

func newPageTokenCodec(key []byte) pageTokenCodec {
	return pageTokenCodec{key: append([]byte(nil), key...)}
}

func (c pageTokenCodec) encode(cursor ports.OrganizationListCursor, requestHash string) (string, error) {
	payload, err := json.Marshal(pageTokenPayload{
		Version:     pageTokenVersion,
		RequestHash: requestHash,
		ID:          cursor.ID.String(),
		Name:        cursor.Name,
		CreateTime:  cursor.CreateTime,
		UpdateTime:  cursor.UpdateTime,
	})
	if err != nil {
		return "", err
	}
	signature := c.sign(payload)
	token := base64.RawURLEncoding.EncodeToString(payload) + "." +
		base64.RawURLEncoding.EncodeToString(signature)
	if len(token) > maximumPageTokenLength {
		return "", fmt.Errorf("page token exceeds %d characters", maximumPageTokenLength)
	}
	return token, nil
}

func (c pageTokenCodec) decode(token string, requestHash string) (*ports.OrganizationListCursor, error) {
	if len(token) == 0 || len(token) > maximumPageTokenLength {
		return nil, ErrInvalidOrganizationPageToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidOrganizationPageToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: malformed payload", ErrInvalidOrganizationPageToken)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, c.sign(payload)) {
		return nil, fmt.Errorf("%w: signature mismatch", ErrInvalidOrganizationPageToken)
	}

	var decoded pageTokenPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, fmt.Errorf("%w: malformed payload", ErrInvalidOrganizationPageToken)
	}
	if decoded.Version != pageTokenVersion || decoded.RequestHash != requestHash {
		return nil, fmt.Errorf("%w: token does not match list request", ErrInvalidOrganizationPageToken)
	}
	id, err := organization.ParseID(decoded.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid cursor id", ErrInvalidOrganizationPageToken)
	}
	if decoded.CreateTime.IsZero() || decoded.UpdateTime.IsZero() {
		return nil, fmt.Errorf("%w: invalid cursor timestamp", ErrInvalidOrganizationPageToken)
	}

	return &ports.OrganizationListCursor{
		ID:         id,
		Name:       decoded.Name,
		CreateTime: decoded.CreateTime,
		UpdateTime: decoded.UpdateTime,
	}, nil
}

func (c pageTokenCodec) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, c.key)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

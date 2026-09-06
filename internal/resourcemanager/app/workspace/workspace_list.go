package workspaceapp

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

	platformfilter "github.com/m8-team/platform/internal/platform/filter"

	"github.com/m8-team/platform/internal/resourcemanager/app/ports"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

var (
	ErrInvalidWorkspacePageSize  = errors.New("invalid workspace page size")
	ErrInvalidWorkspacePageToken = errors.New("invalid workspace page token")
	ErrInvalidWorkspaceFilter    = errors.New("invalid workspace filter")
	ErrInvalidWorkspaceOrderBy   = errors.New("invalid workspace order_by")
)

type normalizedWorkspaceListRequest struct {
	OrganizationID     string                `json:"organization_id"`
	PageSize           int                   `json:"page_size"`
	Filter             ports.WorkspaceFilter `json:"filter"`
	Order              ports.WorkspaceOrder  `json:"order"`
	ShowDeleted        bool                  `json:"show_deleted"`
	AuthorizationScope string                `json:"authorization_scope"`
}

func normalizeWorkspaceListQuery(
	parser *platformfilter.CELParser,
	q ListWorkspaces,
	authorizationScope string,
) (ports.ListWorkspacesOptions, string, error) {
	if q.PageSize < 0 || q.PageSize > ports.MaxWorkspacePageSize {
		return ports.ListWorkspacesOptions{}, "", fmt.Errorf(
			"%w: must be between 0 and %d",
			ErrInvalidWorkspacePageSize,
			ports.MaxWorkspacePageSize,
		)
	}
	if len(q.PageToken) > maximumPageTokenLength {
		return ports.ListWorkspacesOptions{}, "", fmt.Errorf(
			"%w: exceeds %d characters",
			ErrInvalidWorkspacePageToken,
			maximumPageTokenLength,
		)
	}

	filter, err := parseWorkspaceFilter(parser, q.Filter)
	if err != nil {
		return ports.ListWorkspacesOptions{}, "", err
	}
	filter.ShowDeleted = q.ShowDeleted
	order, err := parseWorkspaceOrder(q.OrderBy)
	if err != nil {
		return ports.ListWorkspacesOptions{}, "", err
	}

	options := ports.ListWorkspacesOptions{
		OrganizationID: q.OrganizationID,
		Filter:         filter,
		Order:          order,
		PageSize:       q.PageSize,
	}.WithDefaults()
	if err := options.Validate(); err != nil {
		return ports.ListWorkspacesOptions{}, "", err
	}

	canonical, err := json.Marshal(normalizedWorkspaceListRequest{
		OrganizationID:     q.OrganizationID.String(),
		PageSize:           options.PageSize,
		Filter:             options.Filter,
		Order:              options.Order,
		ShowDeleted:        q.ShowDeleted,
		AuthorizationScope: authorizationScope,
	})
	if err != nil {
		return ports.ListWorkspacesOptions{}, "", fmt.Errorf("marshal list request: %w", err)
	}
	digest := sha256.Sum256(canonical)

	return options, hex.EncodeToString(digest[:]), nil
}

func parseWorkspaceOrder(raw string) (ports.WorkspaceOrder, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ports.WorkspaceOrder{
			Field:     ports.WorkspaceOrderFieldID,
			Direction: ports.SortDirectionAscending,
		}, nil
	}
	if utf8.RuneCountInString(raw) > 128 {
		return ports.WorkspaceOrder{}, fmt.Errorf("%w: exceeds 128 characters", ErrInvalidWorkspaceOrderBy)
	}
	if strings.Contains(raw, ",") {
		return ports.WorkspaceOrder{}, fmt.Errorf("%w: only one field is supported", ErrInvalidWorkspaceOrderBy)
	}
	parts := strings.Fields(raw)
	if len(parts) < 1 || len(parts) > 2 {
		return ports.WorkspaceOrder{}, fmt.Errorf("%w: expected field [asc|desc]", ErrInvalidWorkspaceOrderBy)
	}

	field := ports.WorkspaceOrderField(parts[0])
	if !field.IsValid() {
		return ports.WorkspaceOrder{}, fmt.Errorf("%w: unsupported field %q", ErrInvalidWorkspaceOrderBy, parts[0])
	}
	direction := ports.SortDirectionAscending
	if len(parts) == 2 {
		direction = ports.SortDirection(strings.ToLower(parts[1]))
		if !direction.IsValid() {
			return ports.WorkspaceOrder{}, fmt.Errorf("%w: unsupported direction %q", ErrInvalidWorkspaceOrderBy, parts[1])
		}
	}

	return ports.WorkspaceOrder{Field: field, Direction: direction}, nil
}

type workspacePageTokenCodec struct {
	key []byte
}

type workspacePageTokenPayload struct {
	Version     int       `json:"v"`
	RequestHash string    `json:"request_hash"`
	ID          string    `json:"id"`
	Name        string    `json:"name,omitempty"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

func newWorkspacePageTokenCodec(key []byte) workspacePageTokenCodec {
	return workspacePageTokenCodec{key: append([]byte(nil), key...)}
}

func (c workspacePageTokenCodec) encode(cursor ports.WorkspaceListCursor, requestHash string) (string, error) {
	payload, err := json.Marshal(workspacePageTokenPayload{
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

func (c workspacePageTokenCodec) decode(token string, requestHash string) (*ports.WorkspaceListCursor, error) {
	if len(token) == 0 || len(token) > maximumPageTokenLength {
		return nil, ErrInvalidWorkspacePageToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidWorkspacePageToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: malformed payload", ErrInvalidWorkspacePageToken)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, c.sign(payload)) {
		return nil, fmt.Errorf("%w: signature mismatch", ErrInvalidWorkspacePageToken)
	}

	var decoded workspacePageTokenPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, fmt.Errorf("%w: malformed payload", ErrInvalidWorkspacePageToken)
	}
	if decoded.Version != pageTokenVersion || decoded.RequestHash != requestHash {
		return nil, fmt.Errorf("%w: token does not match list request", ErrInvalidWorkspacePageToken)
	}
	id, err := workspace.ParseID(decoded.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid cursor id", ErrInvalidWorkspacePageToken)
	}
	if decoded.CreateTime.IsZero() || decoded.UpdateTime.IsZero() {
		return nil, fmt.Errorf("%w: invalid cursor timestamp", ErrInvalidWorkspacePageToken)
	}

	return &ports.WorkspaceListCursor{
		ID:         id,
		Name:       decoded.Name,
		CreateTime: decoded.CreateTime,
		UpdateTime: decoded.UpdateTime,
	}, nil
}

func (c workspacePageTokenCodec) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, c.key)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

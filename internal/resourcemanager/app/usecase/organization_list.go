package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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

func parseOrganizationFilter(raw string) (ports.OrganizationFilter, error) {
	filter := ports.OrganizationFilter{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return filter, nil
	}
	if utf8.RuneCountInString(raw) > maximumPageTokenLength {
		return filter, fmt.Errorf("%w: exceeds 1024 characters", ErrInvalidOrganizationFilter)
	}

	clauses, err := splitFilterClauses(raw)
	if err != nil {
		return filter, err
	}
	seenName := false
	filter.LabelsEqual = make(map[string]string)

	for _, clause := range clauses {
		field, rawValue, ok := strings.Cut(clause, "=")
		if !ok {
			return filter, fmt.Errorf("%w: expected field = value in %q", ErrInvalidOrganizationFilter, clause)
		}
		field = strings.TrimSpace(field)
		rawValue = strings.TrimSpace(rawValue)
		if field == "" || rawValue == "" {
			return filter, fmt.Errorf("%w: incomplete clause %q", ErrInvalidOrganizationFilter, clause)
		}

		switch {
		case field == "state":
			if len(filter.States) != 0 {
				return filter, fmt.Errorf("%w: duplicate state clause", ErrInvalidOrganizationFilter)
			}
			state, err := parseOrganizationState(rawValue)
			if err != nil {
				return filter, err
			}
			filter.States = append(filter.States, state)
		case field == "name":
			if seenName {
				return filter, fmt.Errorf("%w: duplicate name clause", ErrInvalidOrganizationFilter)
			}
			name, err := parseQuotedFilterValue(rawValue)
			if err != nil {
				return filter, fmt.Errorf("%w: name: %v", ErrInvalidOrganizationFilter, err)
			}
			seenName = true
			filter.NameEquals = &name
		case strings.HasPrefix(field, "labels."):
			key := strings.TrimPrefix(field, "labels.")
			if strings.TrimSpace(key) == "" {
				return filter, fmt.Errorf("%w: label key is empty", ErrInvalidOrganizationFilter)
			}
			if _, duplicate := filter.LabelsEqual[key]; duplicate {
				return filter, fmt.Errorf("%w: duplicate label %q", ErrInvalidOrganizationFilter, key)
			}
			value, err := parseQuotedFilterValue(rawValue)
			if err != nil {
				return filter, fmt.Errorf("%w: label %q: %v", ErrInvalidOrganizationFilter, key, err)
			}
			filter.LabelsEqual[key] = value
		default:
			return filter, fmt.Errorf("%w: unsupported field %q", ErrInvalidOrganizationFilter, field)
		}
	}

	if len(filter.LabelsEqual) == 0 {
		filter.LabelsEqual = nil
	}
	return filter, nil
}

func splitFilterClauses(raw string) ([]string, error) {
	clauses := make([]string, 0, 2)
	start := 0
	inQuotes := false
	escaped := false

	for i := 0; i < len(raw); i++ {
		current := raw[i]
		if inQuotes {
			if escaped {
				escaped = false
				continue
			}
			if current == '\\' {
				escaped = true
				continue
			}
			if current == '"' {
				inQuotes = false
			}
			continue
		}
		if current == '"' {
			inQuotes = true
			continue
		}
		if i > start && i+5 <= len(raw) && isASCIISpace(raw[i]) &&
			strings.EqualFold(raw[i+1:i+4], "and") && isASCIISpace(raw[i+4]) {
			clause := strings.TrimSpace(raw[start:i])
			if clause == "" {
				return nil, fmt.Errorf("%w: empty clause", ErrInvalidOrganizationFilter)
			}
			clauses = append(clauses, clause)
			i += 4
			for i < len(raw) && isASCIISpace(raw[i]) {
				i++
			}
			start = i
			i--
		}
	}
	if inQuotes || escaped {
		return nil, fmt.Errorf("%w: unterminated quoted value", ErrInvalidOrganizationFilter)
	}
	last := strings.TrimSpace(raw[start:])
	if last == "" {
		return nil, fmt.Errorf("%w: empty clause", ErrInvalidOrganizationFilter)
	}
	clauses = append(clauses, last)
	return clauses, nil
}

func isASCIISpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\r' || value == '\n'
}

func parseQuotedFilterValue(value string) (string, error) {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return "", errors.New("value must be double-quoted")
	}
	parsed, err := strconv.Unquote(value)
	if err != nil {
		return "", fmt.Errorf("invalid quoted value: %w", err)
	}
	return parsed, nil
}

func parseOrganizationState(value string) (organization.State, error) {
	if strings.HasPrefix(value, "\"") || strings.HasSuffix(value, "\"") {
		parsed, err := parseQuotedFilterValue(value)
		if err != nil {
			return organization.StateUnspecified, fmt.Errorf("%w: state: %v", ErrInvalidOrganizationFilter, err)
		}
		value = parsed
	}
	switch strings.ToUpper(value) {
	case "CREATING":
		return organization.StateCreating, nil
	case "ACTIVE":
		return organization.StateActive, nil
	case "SUSPENDED":
		return organization.StateSuspended, nil
	case "DELETING":
		return organization.StateDeleting, nil
	case "DELETED":
		return organization.StateDeleted, nil
	case "FAILED":
		return organization.StateFailed, nil
	default:
		return organization.StateUnspecified, fmt.Errorf(
			"%w: unsupported state %q",
			ErrInvalidOrganizationFilter,
			value,
		)
	}
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

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	authzmodel "github.com/m8platform/platform/iam/internal/module/authz/model"
)

type AccessDecisionCache struct {
	cache         *Cache
	policyVersion string
}

func NewAccessDecisionCache(cache *Cache, policyVersion string) *AccessDecisionCache {
	return &AccessDecisionCache{
		cache:         cache,
		policyVersion: policyVersion,
	}
}

func (c *AccessDecisionCache) GetAccessDecision(ctx context.Context, query authzmodel.AccessCheckQuery) (authzmodel.AccessCheckResult, bool, error) {
	if c == nil || c.cache == nil {
		return authzmodel.AccessCheckResult{}, false, nil
	}
	payload, ok, err := c.cache.Get(ctx, c.cacheKey(query))
	if err != nil || !ok {
		return authzmodel.AccessCheckResult{}, ok, err
	}

	var result authzmodel.AccessCheckResult
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		return authzmodel.AccessCheckResult{}, false, err
	}
	return result, true, nil
}

func (c *AccessDecisionCache) SaveAccessDecision(ctx context.Context, query authzmodel.AccessCheckQuery, result authzmodel.AccessCheckResult) error {
	if c == nil || c.cache == nil {
		return nil
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return c.cache.Set(ctx, c.cacheKey(query), string(payload), 30*time.Second)
}

func (c *AccessDecisionCache) cacheKey(query authzmodel.AccessCheckQuery) string {
	return fmt.Sprintf(
		"authz:%s:%s:%s:%s:%s:%s",
		query.Subject.Type,
		query.Subject.ID,
		query.Resource.Type,
		query.Resource.ID,
		query.Permission,
		c.policyVersion,
	)
}

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	legacyredis "github.com/m8platform/platform/iam/internal/storage/redis"
	"github.com/m8platform/platform/iam/internal/usecase/model"
)

type AccessDecisionCache struct {
	cache         *legacyredis.Cache
	policyVersion string
}

func NewAccessDecisionCache(cache *legacyredis.Cache, policyVersion string) *AccessDecisionCache {
	return &AccessDecisionCache{
		cache:         cache,
		policyVersion: policyVersion,
	}
}

func (c *AccessDecisionCache) GetAccessDecision(ctx context.Context, query model.AccessCheckQuery) (model.AccessCheckResult, bool, error) {
	if c == nil || c.cache == nil {
		return model.AccessCheckResult{}, false, nil
	}
	payload, ok, err := c.cache.Get(ctx, c.cacheKey(query))
	if err != nil || !ok {
		return model.AccessCheckResult{}, ok, err
	}

	var result model.AccessCheckResult
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		return model.AccessCheckResult{}, false, err
	}
	return result, true, nil
}

func (c *AccessDecisionCache) SaveAccessDecision(ctx context.Context, query model.AccessCheckQuery, result model.AccessCheckResult) error {
	if c == nil || c.cache == nil {
		return nil
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return c.cache.Set(ctx, c.cacheKey(query), string(payload), 30*time.Second)
}

func (c *AccessDecisionCache) cacheKey(query model.AccessCheckQuery) string {
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

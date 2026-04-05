package redis

import (
	"context"
	"fmt"
	"time"

	authzv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/authz/v1"
	"github.com/m8platform/platform/iam/internal/foundation/config"
	goredis "github.com/redis/go-redis/v9"
)

type Cache struct {
	client *goredis.Client
}

func NewCache(cfg config.RedisConfig) *Cache {
	return &Cache{
		client: goredis.NewClient(&goredis.Options{
			Addr:     cfg.Address,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
	}
}

func (c *Cache) Get(ctx context.Context, key string) (string, bool, error) {
	value, err := c.client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (c *Cache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func BuildCheckAccessCacheKey(subject *authzv1.SubjectRef, resource *authzv1.ResourceRef, permission string, policyVersion string) string {
	return fmt.Sprintf(
		"authz:%s:%s:%s:%s:%s:%s",
		subject.GetType().String(),
		subject.GetId(),
		resource.GetType().String(),
		resource.GetId(),
		permission,
		policyVersion,
	)
}

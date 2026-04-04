package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName string
	Environment string
	Development bool
	GRPC        GRPCConfig
	HTTP        HTTPConfig
	YDB         YDBConfig
	Redis       RedisConfig
	Keycloak    KeycloakConfig
	SpiceDB     SpiceDBConfig
	Temporal    TemporalConfig
	Topics      TopicsConfig
}

type GRPCConfig struct {
	Address string
}

type HTTPConfig struct {
	Address    string
	OpenAPIDir string
}

type YDBConfig struct {
	DSN      string
	Database string
}

type RedisConfig struct {
	Address        string
	Password       string
	DB             int
	DefaultTTL     time.Duration
	PolicyVersion  string
	CheckAccessTTL time.Duration
}

type KeycloakConfig struct {
	BaseURL      string
	Realm        string
	ClientID     string
	ClientSecret string
}

type SpiceDBConfig struct {
	Endpoint     string
	Token        string
	SchemaPath   string
	PreSharedKey string
	Consistency  string
}

type TemporalConfig struct {
	Address   string
	Namespace string
	TaskQueue string
	Enabled   bool
}

type TopicsConfig struct {
	IdentityUsers       string
	IdentityMemberships string
	IdentityGroups      string
	ServiceAccounts     string
	OAuthClients        string
	SupportGrants       string
	Relationships       string
	AuditEvents         string
	Operations          string
}

func Load() Config {
	return Config{
		ServiceName: envString("IAM_SERVICE_NAME", "m8-platform-iam"),
		Environment: envString("IAM_ENVIRONMENT", "dev"),
		Development: envBool("IAM_DEVELOPMENT", true),
		GRPC: GRPCConfig{
			Address: envString("IAM_GRPC_ADDRESS", ":8080"),
		},
		HTTP: HTTPConfig{
			Address:    envString("IAM_HTTP_ADDRESS", ":8082"),
			OpenAPIDir: envString("IAM_OPENAPI_DIR", "gen/openapi"),
		},
		YDB: YDBConfig{
			DSN:      envString("IAM_YDB_DSN", ""),
			Database: envString("IAM_YDB_DATABASE", ""),
		},
		Redis: RedisConfig{
			Address:        envString("IAM_REDIS_ADDRESS", "127.0.0.1:6379"),
			Password:       envString("IAM_REDIS_PASSWORD", ""),
			DB:             envInt("IAM_REDIS_DB", 0),
			DefaultTTL:     envDuration("IAM_REDIS_DEFAULT_TTL", 5*time.Minute),
			CheckAccessTTL: envDuration("IAM_REDIS_CHECK_ACCESS_TTL", 30*time.Second),
			PolicyVersion:  envString("IAM_POLICY_VERSION", "v1"),
		},
		Keycloak: KeycloakConfig{
			BaseURL:      strings.TrimRight(envString("IAM_KEYCLOAK_BASE_URL", ""), "/"),
			Realm:        envString("IAM_KEYCLOAK_REALM", "m8"),
			ClientID:     envString("IAM_KEYCLOAK_CLIENT_ID", ""),
			ClientSecret: envString("IAM_KEYCLOAK_CLIENT_SECRET", ""),
		},
		SpiceDB: SpiceDBConfig{
			Endpoint:     envString("IAM_SPICEDB_ENDPOINT", ""),
			Token:        envString("IAM_SPICEDB_TOKEN", ""),
			SchemaPath:   envString("IAM_SPICEDB_SCHEMA_PATH", "docs/spicedb/schema.zed"),
			PreSharedKey: envString("IAM_SPICEDB_PRESHARED_KEY", ""),
			Consistency:  envString("IAM_SPICEDB_CONSISTENCY", "at_least_as_fresh"),
		},
		Temporal: TemporalConfig{
			Address:   envString("IAM_TEMPORAL_ADDRESS", "127.0.0.1:7233"),
			Namespace: envString("IAM_TEMPORAL_NAMESPACE", "default"),
			TaskQueue: envString("IAM_TEMPORAL_TASK_QUEUE", "iam-task-queue"),
			Enabled:   envBool("IAM_TEMPORAL_ENABLED", true),
		},
		Topics: TopicsConfig{
			IdentityUsers:       envString("IAM_TOPIC_IDENTITY_USERS", "iam.identity.users.v1"),
			IdentityMemberships: envString("IAM_TOPIC_IDENTITY_MEMBERSHIPS", "iam.identity.memberships.v1"),
			IdentityGroups:      envString("IAM_TOPIC_IDENTITY_GROUPS", "iam.identity.groups.v1"),
			ServiceAccounts:     envString("IAM_TOPIC_SERVICE_ACCOUNTS", "iam.identity.service_accounts.v1"),
			OAuthClients:        envString("IAM_TOPIC_OAUTH_CLIENTS", "iam.oauth.clients.v1"),
			SupportGrants:       envString("IAM_TOPIC_SUPPORT_GRANTS", "iam.support.grants.v1"),
			Relationships:       envString("IAM_TOPIC_RELATIONSHIPS", "iam.authz.relationships.v1"),
			AuditEvents:         envString("IAM_TOPIC_AUDIT_EVENTS", "iam.audit.events.v1"),
			Operations:          envString("IAM_TOPIC_OPERATIONS", "iam.operations.v1"),
		},
	}
}

func envString(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func envBool(key string, fallback bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

package config

import legacyconfig "github.com/m8platform/platform/iam/internal/config"

type Config = legacyconfig.Config
type GRPCConfig = legacyconfig.GRPCConfig
type HTTPConfig = legacyconfig.HTTPConfig
type YDBConfig = legacyconfig.YDBConfig
type RedisConfig = legacyconfig.RedisConfig
type KeycloakConfig = legacyconfig.KeycloakConfig
type SpiceDBConfig = legacyconfig.SpiceDBConfig
type TemporalConfig = legacyconfig.TemporalConfig
type TopicsConfig = legacyconfig.TopicsConfig

func Load() Config {
	return legacyconfig.Load()
}

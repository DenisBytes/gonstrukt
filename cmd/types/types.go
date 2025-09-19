package types

type ServiceType string

const (
	ServiceTypeGateway ServiceType = "gateway"
	ServiceTypeAuth    ServiceType = "auth"
)

func (s ServiceType) String() string {
	return string(s)
}

type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "psql"
)

func (d DatabaseType) String() string {
	return string(d)
}

type CacheType string

const (
	CacheTypeMemory CacheType = "memory"
	CacheTypeRedis  CacheType = "redis"
	CacheTypeValkey CacheType = "valkey"
)

func (c CacheType) String() string {
	return string(c)
}

type ConfigType string

const (
	ConfigTypeLocalYAML ConfigType = "yaml"
	ConfigTypeVault     ConfigType = "vault"
)

func (c ConfigType) String() string {
	return string(c)
}

type RateLimiterType string

const (
	RateLimiterTokenBucket            RateLimiterType = "token-bucket"
	RateLimiterApproximatedSlidingWin RateLimiterType = "approximated-sliding-window"
)

func (r RateLimiterType) String() string {
	return string(r)
}

type ObservabilityType string

const (
	ObservabilityOTLP ObservabilityType = "otlp"
	ObservabilityNone ObservabilityType = "none"
)

func (o ObservabilityType) String() string {
	return string(o)
}

type ServiceConfig struct {
	Name          string
	ServiceType   ServiceType
	Database      DatabaseType
	Cache         *CacheType
	Config        ConfigType
	Observability ObservabilityType
	RateLimiter   *RateLimiterType
}

func ValidServiceTypes() []string {
	return []string{string(ServiceTypeGateway), string(ServiceTypeAuth)}
}

func ValidDatabaseTypes() []string {
	return []string{string(DatabaseTypePostgreSQL)}
}

func ValidCacheTypes() []string {
	return []string{string(CacheTypeMemory), string(CacheTypeRedis), string(CacheTypeValkey)}
}

func ValidConfigTypes() []string {
	return []string{string(ConfigTypeLocalYAML), string(ConfigTypeVault)}
}

func ValidRateLimiterTypes() []string {
	return []string{string(RateLimiterTokenBucket), string(RateLimiterApproximatedSlidingWin)}
}

func ValidObservabilityTypes() []string {
	return []string{string(ObservabilityOTLP), string(ObservabilityNone)}
}

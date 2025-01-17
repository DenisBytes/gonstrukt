package generator

// DependencyMap maps feature selections to their required Go dependencies
var DependencyMap = map[string][]string{
	// Databases
	"database:postgres": {
		"github.com/jackc/pgx/v5",
		"github.com/jackc/pgx/v5/pgxpool",
	},
	"database:mysql": {
		"github.com/go-sql-driver/mysql",
	},
	"database:sqlite": {
		"modernc.org/sqlite",
	},
	"database:mongodb": {
		"go.mongodb.org/mongo-driver/mongo",
		"go.mongodb.org/mongo-driver/bson",
	},
	"database:arangodb": {
		"github.com/arangodb/go-driver/v2/arangodb",
		"github.com/arangodb/go-driver/v2/connection",
	},

	// Cache
	"cache:redis": {
		"github.com/redis/go-redis/v9",
	},
	"cache:valkey": {
		"github.com/redis/go-redis/v9",
	},

	// Config
	"config:yaml": {
		"gopkg.in/yaml.v3",
	},
	"config:vault": {
		"github.com/hashicorp/vault/api",
	},

	// Observability
	"observability:otlp": {
		"go.opentelemetry.io/otel",
		"go.opentelemetry.io/otel/sdk/trace",
		"go.opentelemetry.io/otel/sdk/metric",
		"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc",
		"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc",
		"go.opentelemetry.io/otel/trace",
		"go.opentelemetry.io/otel/metric",
		"go.opentelemetry.io/otel/attribute",
		"go.opentelemetry.io/otel/propagation",
		"go.opentelemetry.io/otel/semconv/v1.24.0",
	},

	// Auth
	"auth:jwt": {
		"github.com/golang-jwt/jwt/v5",
	},
	"auth:rbac": {
		"github.com/casbin/casbin/v2",
	},
	"auth:crypto": {
		"golang.org/x/crypto",
	},
	"auth:mfa": {
		"github.com/pquerna/otp",
		"github.com/pquerna/otp/totp",
	},

	// OAuth
	"oauth:google": {
		"golang.org/x/oauth2",
		"golang.org/x/oauth2/google",
	},
	"oauth:microsoft": {
		"golang.org/x/oauth2",
	},
	"oauth:apple": {
		"github.com/Timothylock/go-signin-with-apple/apple",
	},

	// Email
	"email:ses": {
		"github.com/aws/aws-sdk-go-v2/service/ses",
		"github.com/aws/aws-sdk-go-v2/config",
	},
	// email:smtp uses stdlib net/smtp, no external deps

	// Common
	"common": {
		"go.uber.org/zap",
		"github.com/google/uuid",
	},
}

// GetDependencies returns all dependencies for the given features
func GetDependencies(features []string) []string {
	depSet := make(map[string]struct{})

	// Always include common dependencies
	for _, dep := range DependencyMap["common"] {
		depSet[dep] = struct{}{}
	}

	// Add feature-specific dependencies
	for _, feature := range features {
		if deps, ok := DependencyMap[feature]; ok {
			for _, dep := range deps {
				depSet[dep] = struct{}{}
			}
		}
	}

	// Convert to slice
	deps := make([]string, 0, len(depSet))
	for dep := range depSet {
		deps = append(deps, dep)
	}

	return deps
}

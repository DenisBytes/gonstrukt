package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/generator/writers"
)

// Generator orchestrates the project generation
type Generator struct {
	config *config.ProjectConfig
	data   *config.TemplateData
	writer *writers.FileWriter
}

// NewGenerator creates a new generator
func NewGenerator(cfg *config.ProjectConfig) *Generator {
	data := config.NewTemplateData(cfg)

	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = cfg.ProjectName
	}

	return &Generator{
		config: cfg,
		data:   data,
		writer: writers.NewFileWriter(outputDir, data),
	}
}

// Generate generates the project
func (g *Generator) Generate(ctx context.Context) error {
	// Create output directory
	if err := os.MkdirAll(g.writer.OutputDir(), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate based on service type
	switch g.config.ServiceType {
	case config.ServiceGateway:
		if err := g.generateGateway(); err != nil {
			return fmt.Errorf("failed to generate gateway: %w", err)
		}
	case config.ServiceAuth:
		if err := g.generateAuth(); err != nil {
			return fmt.Errorf("failed to generate auth service: %w", err)
		}
	case config.ServiceBoth:
		if err := g.generateMonorepo(); err != nil {
			return fmt.Errorf("failed to generate monorepo: %w", err)
		}
	}

	// Generate static files
	if err := g.generateStaticFiles(); err != nil {
		return fmt.Errorf("failed to generate static files: %w", err)
	}

	// Generate frontend if configured
	if len(g.config.Frontends) > 0 {
		if err := g.generateFrontend(); err != nil {
			return fmt.Errorf("failed to generate frontend: %w", err)
		}
	}

	// Generate K8s manifests if enabled
	if g.config.EnableK8s {
		if err := g.generateK8s(); err != nil {
			return fmt.Errorf("failed to generate k8s manifests: %w", err)
		}
	}

	// Run go mod tidy and format code
	if g.config.ServiceType == config.ServiceBoth {
		// For monorepo, run on each service directory
		gatewayDir := filepath.Join(g.writer.OutputDir(), "services", "gateway")
		authDir := filepath.Join(g.writer.OutputDir(), "services", "auth_service")

		if err := g.runGoModTidyIn(gatewayDir); err != nil {
			return fmt.Errorf("failed to run go mod tidy for gateway: %w", err)
		}
		if err := g.runGoModTidyIn(authDir); err != nil {
			return fmt.Errorf("failed to run go mod tidy for auth: %w", err)
		}
		if err := g.runGoFmtIn(gatewayDir); err != nil {
			return fmt.Errorf("failed to format gateway code: %w", err)
		}
		if err := g.runGoFmtIn(authDir); err != nil {
			return fmt.Errorf("failed to format auth code: %w", err)
		}
	} else {
		if err := g.runGoModTidy(); err != nil {
			return fmt.Errorf("failed to run go mod tidy: %w", err)
		}
		if err := g.runGoFmt(); err != nil {
			return fmt.Errorf("failed to format code: %w", err)
		}
	}

	return nil
}

// generateGateway generates a gateway service
func (g *Generator) generateGateway() error {
	data := g.data.ForGateway()
	writer := writers.NewFileWriter(g.writer.OutputDir(), data)

	// Create directory structure
	dirs := []string{
		"cmd/gateway",
		"internals/services",
		"internals/middleware",
		"internals/config",
		"internals/utils",
	}

	if g.config.Cache != nil {
		dirs = append(dirs, "internals/cache")
	}
	if g.config.RateLimiter != nil {
		dirs = append(dirs, "internals/ratelimiter")
	}
	if g.config.Observability {
		dirs = append(dirs, "internals/telemetry")
	}

	for _, dir := range dirs {
		if err := writer.EnsureDir(dir); err != nil {
			return err
		}
	}

	// Generate main.go
	if err := writer.WriteTemplate("gateway/cmd/main.go.tmpl", "cmd/gateway/main.go"); err != nil {
		return err
	}

	// Generate services
	if err := writer.WriteTemplate("gateway/services/gateway.go.tmpl", "internals/services/gateway.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("gateway/services/health.go.tmpl", "internals/services/health.go"); err != nil {
		return err
	}

	// Generate cache
	if g.config.Cache != nil {
		if err := writer.WriteTemplate("gateway/cache/interface.go.tmpl", "internals/cache/interface.go"); err != nil {
			return err
		}
		cacheFile := fmt.Sprintf("gateway/cache/%s.go.tmpl", *g.config.Cache)
		if err := writer.WriteTemplate(cacheFile, fmt.Sprintf("internals/cache/%s.go", *g.config.Cache)); err != nil {
			return err
		}
	}

	// Generate rate limiter
	if g.config.RateLimiter != nil {
		if err := writer.WriteTemplate("gateway/ratelimiter/interface.go.tmpl", "internals/ratelimiter/interface.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate("gateway/ratelimiter/middleware.go.tmpl", "internals/ratelimiter/middleware.go"); err != nil {
			return err
		}

		var rlFile, rlOut string
		switch *g.config.RateLimiter {
		case config.RateLimiterTokenBucket:
			rlFile = "gateway/ratelimiter/token_bucket.go.tmpl"
			rlOut = "internals/ratelimiter/token_bucket.go"
		case config.RateLimiterSlidingWindow:
			rlFile = "gateway/ratelimiter/sliding_window.go.tmpl"
			rlOut = "internals/ratelimiter/sliding_window.go"
		case config.RateLimiterLeakyBucket:
			rlFile = "gateway/ratelimiter/leaky_bucket.go.tmpl"
			rlOut = "internals/ratelimiter/leaky_bucket.go"
		case config.RateLimiterFixedWindow:
			rlFile = "gateway/ratelimiter/fixed_window.go.tmpl"
			rlOut = "internals/ratelimiter/fixed_window.go"
		default:
			rlFile = "gateway/ratelimiter/token_bucket.go.tmpl"
			rlOut = "internals/ratelimiter/token_bucket.go"
		}
		if err := writer.WriteTemplate(rlFile, rlOut); err != nil {
			return err
		}
	}

	// Generate middleware
	if err := writer.WriteTemplate("gateway/middleware/cors.go.tmpl", "internals/middleware/cors.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("gateway/middleware/csrf.go.tmpl", "internals/middleware/csrf.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("gateway/middleware/chain.go.tmpl", "internals/middleware/chain.go"); err != nil {
		return err
	}

	// Generate security middleware
	if err := writer.WriteTemplate("gateway/middleware/host_validation.go.tmpl", "internals/middleware/host_validation.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("gateway/middleware/path_traversal.go.tmpl", "internals/middleware/path_traversal.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("gateway/middleware/malicious_payload.go.tmpl", "internals/middleware/malicious_payload.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("gateway/middleware/request_size.go.tmpl", "internals/middleware/request_size.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("gateway/middleware/security_logging.go.tmpl", "internals/middleware/security_logging.go"); err != nil {
		return err
	}

	// Generate auth cache service if enabled
	if g.config.AuthCache {
		if err := writer.WriteTemplate("gateway/services/auth_cache.go.tmpl", "internals/services/auth_cache.go"); err != nil {
			return err
		}
	}

	// Generate test files
	if err := g.generateGatewayTests(writer); err != nil {
		return err
	}

	// Generate common files
	if err := g.generateCommon(writer); err != nil {
		return err
	}

	// Generate gateway-specific config types
	if err := writer.WriteTemplate("gateway/config/types.go.tmpl", "internals/config/types.go"); err != nil {
		return err
	}

	// Generate go.mod
	if err := writer.WriteTemplate("gateway/go.mod.tmpl", "go.mod"); err != nil {
		return err
	}

	// Generate Dockerfile
	if err := writer.WriteTemplate("gateway/Dockerfile.tmpl", "Dockerfile"); err != nil {
		return err
	}

	return nil
}

// generateAuth generates an auth service
func (g *Generator) generateAuth() error {
	data := g.data.ForAuth()
	writer := writers.NewFileWriter(g.writer.OutputDir(), data)

	// Create directory structure
	dirs := []string{
		"cmd/auth_service",
		"internals/services",
		"internals/middleware",
		"internals/config",
		"internals/utils",
		"internals/types",
		"internals/db",
	}

	if g.config.Observability {
		dirs = append(dirs, "internals/telemetry")
	}

	// Add cache directory if needed (optional gateway feature for auth)
	if g.config.Cache != nil {
		dirs = append(dirs, "internals/cache")
	}

	// Add rate limiter directory if needed (optional gateway feature for auth)
	if g.config.RateLimiter != nil {
		dirs = append(dirs, "internals/ratelimiter")
	}

	// Add OAuth directory if needed
	if len(g.config.OAuthProviders) > 0 {
		dirs = append(dirs, "internals/services/oauth")
	}

	// Add email directory if needed
	if g.config.EmailService != nil {
		dirs = append(dirs, "internals/services/email")
	}

	// Add crypto directory if OAuth is enabled (for token encryption)
	if len(g.config.OAuthProviders) > 0 {
		dirs = append(dirs, "internals/crypto")
	}

	for _, dir := range dirs {
		if err := writer.EnsureDir(dir); err != nil {
			return err
		}
	}

	// Generate main.go
	if err := writer.WriteTemplate("auth/cmd/main.go.tmpl", "cmd/auth_service/main.go"); err != nil {
		return err
	}

	// Generate services
	if err := writer.WriteTemplate("auth/services/auth_service.go.tmpl", "internals/services/auth_service.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("auth/services/handlers.go.tmpl", "internals/services/handlers.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("auth/services/key_manager.go.tmpl", "internals/services/key_manager.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("auth/services/token_blacklist.go.tmpl", "internals/services/token_blacklist.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("auth/services/health.go.tmpl", "internals/services/health.go"); err != nil {
		return err
	}

	// Generate GDPR service if any GDPR features are enabled
	if len(g.config.GDPRFeatures) > 0 {
		if err := writer.WriteTemplate("auth/services/gdpr.go.tmpl", "internals/services/gdpr.go"); err != nil {
			return err
		}
	}

	// Generate MFA service if enabled
	if g.config.EnableMFA {
		if err := writer.WriteTemplate("auth/services/mfa.go.tmpl", "internals/services/mfa.go"); err != nil {
			return err
		}
	}

	// Generate OAuth providers
	for _, provider := range g.config.OAuthProviders {
		tmplFile := fmt.Sprintf("auth/services/oauth/%s.go.tmpl", provider)
		outFile := fmt.Sprintf("internals/services/oauth/%s.go", provider)
		if err := writer.WriteTemplate(tmplFile, outFile); err != nil {
			return err
		}
	}

	// Generate email service
	if g.config.EmailService != nil {
		tmplFile := fmt.Sprintf("auth/services/email/%s.go.tmpl", *g.config.EmailService)
		outFile := fmt.Sprintf("internals/services/email/%s.go", *g.config.EmailService)
		if err := writer.WriteTemplate(tmplFile, outFile); err != nil {
			return err
		}
	}

	// Generate crypto package for OAuth token encryption
	if len(g.config.OAuthProviders) > 0 {
		if err := writer.WriteTemplate("auth/services/crypto/encryption.go.tmpl", "internals/crypto/encryption.go"); err != nil {
			return err
		}
	}

	// Generate OAuth revocation service (requires OAuth + GDPR data deletion)
	if len(g.config.OAuthProviders) > 0 && g.data.HasGDPRDataDeletion {
		if err := writer.WriteTemplate("auth/services/oauth_revocation.go.tmpl", "internals/services/oauth_revocation.go"); err != nil {
			return err
		}
	}

	// Generate IP hash utility
	if err := writer.WriteTemplate("auth/services/ip_hash.go.tmpl", "internals/services/ip_hash.go"); err != nil {
		return err
	}

	// Generate tenancy handlers if enabled
	if g.config.EnableTenancy {
		if err := writer.WriteTemplate("auth/services/handlers_workspace.go.tmpl", "internals/services/handlers_workspace.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate("auth/services/handlers_invitation.go.tmpl", "internals/services/handlers_invitation.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate("auth/services/handlers_tenant.go.tmpl", "internals/services/handlers_tenant.go"); err != nil {
			return err
		}
	}

	// Generate types
	if err := writer.WriteTemplate("auth/types/types.go.tmpl", "internals/types/types.go"); err != nil {
		return err
	}

	// Generate GDPR types if any GDPR features are enabled
	if len(g.config.GDPRFeatures) > 0 {
		if err := writer.WriteTemplate("auth/types/gdpr.go.tmpl", "internals/types/gdpr.go"); err != nil {
			return err
		}
	}

	// Generate tenancy types if enabled
	if g.config.EnableTenancy {
		if err := writer.WriteTemplate("auth/types/tenancy.go.tmpl", "internals/types/tenancy.go"); err != nil {
			return err
		}
	}

	// Generate OAuth token types if OAuth enabled
	if len(g.config.OAuthProviders) > 0 {
		if err := writer.WriteTemplate("auth/types/oauth_token.go.tmpl", "internals/types/oauth_token.go"); err != nil {
			return err
		}
	}

	// Generate middleware
	if err := writer.WriteTemplate("auth/middleware/auth.go.tmpl", "internals/middleware/auth.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("auth/middleware/request_validation.go.tmpl", "internals/middleware/request_validation.go"); err != nil {
		return err
	}

	// Generate RBAC middleware if enabled
	if g.config.EnableRBAC {
		if err := writer.WriteTemplate("auth/middleware/rbac.go.tmpl", "internals/middleware/rbac.go"); err != nil {
			return err
		}
	}

	// Generate cache (optional gateway feature for auth - reuses gateway templates)
	if g.config.Cache != nil {
		if err := writer.WriteTemplate("gateway/cache/interface.go.tmpl", "internals/cache/interface.go"); err != nil {
			return err
		}
		cacheFile := fmt.Sprintf("gateway/cache/%s.go.tmpl", *g.config.Cache)
		if err := writer.WriteTemplate(cacheFile, fmt.Sprintf("internals/cache/%s.go", *g.config.Cache)); err != nil {
			return err
		}
	}

	// Generate rate limiter (optional gateway feature for auth - reuses gateway templates)
	if g.config.RateLimiter != nil {
		if err := writer.WriteTemplate("gateway/ratelimiter/interface.go.tmpl", "internals/ratelimiter/interface.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate("gateway/ratelimiter/middleware.go.tmpl", "internals/ratelimiter/middleware.go"); err != nil {
			return err
		}

		var rlFile, rlOut string
		switch *g.config.RateLimiter {
		case config.RateLimiterTokenBucket:
			rlFile = "gateway/ratelimiter/token_bucket.go.tmpl"
			rlOut = "internals/ratelimiter/token_bucket.go"
		case config.RateLimiterSlidingWindow:
			rlFile = "gateway/ratelimiter/sliding_window.go.tmpl"
			rlOut = "internals/ratelimiter/sliding_window.go"
		case config.RateLimiterLeakyBucket:
			rlFile = "gateway/ratelimiter/leaky_bucket.go.tmpl"
			rlOut = "internals/ratelimiter/leaky_bucket.go"
		case config.RateLimiterFixedWindow:
			rlFile = "gateway/ratelimiter/fixed_window.go.tmpl"
			rlOut = "internals/ratelimiter/fixed_window.go"
		default:
			rlFile = "gateway/ratelimiter/token_bucket.go.tmpl"
			rlOut = "internals/ratelimiter/token_bucket.go"
		}
		if err := writer.WriteTemplate(rlFile, rlOut); err != nil {
			return err
		}
	}

	// Generate database layer
	if g.config.Database != nil {
		if err := g.generateDatabase(writer); err != nil {
			return err
		}
	}

	// Generate test files
	if err := g.generateAuthTests(writer); err != nil {
		return err
	}

	// Generate common files
	if err := g.generateCommon(writer); err != nil {
		return err
	}

	// Generate auth-specific config types
	if err := writer.WriteTemplate("auth/config/types.go.tmpl", "internals/config/types.go"); err != nil {
		return err
	}

	// Copy rbac_model.conf if RBAC is enabled
	if g.config.EnableRBAC {
		if err := writer.CopyStatic("auth/rbac_model.conf", "rbac_model.conf"); err != nil {
			return err
		}
	}

	// Generate go.mod
	if err := writer.WriteTemplate("auth/go.mod.tmpl", "go.mod"); err != nil {
		return err
	}

	// Generate Dockerfile
	if err := writer.WriteTemplate("auth/Dockerfile.tmpl", "Dockerfile"); err != nil {
		return err
	}

	return nil
}

// generateMonorepo generates both services in a monorepo
func (g *Generator) generateMonorepo() error {
	// Create services directory
	if err := g.writer.EnsureDir("services"); err != nil {
		return err
	}

	// Generate gateway with unique module path
	gatewayDir := filepath.Join(g.writer.OutputDir(), "services", "gateway")
	gatewayConfig := *g.config
	gatewayConfig.OutputDir = gatewayDir
	gatewayConfig.ServiceType = config.ServiceGateway
	gatewayConfig.ModuleName = g.config.ModuleName + "/services/gateway"

	gatewayGen := NewGenerator(&gatewayConfig)
	if err := gatewayGen.generateGateway(); err != nil {
		return fmt.Errorf("failed to generate gateway: %w", err)
	}

	// Generate auth with unique module path
	// When gateway exists, auth must NOT have cache or rate limiter
	// (those belong to the gateway service only)
	authDir := filepath.Join(g.writer.OutputDir(), "services", "auth_service")
	authConfig := *g.config
	authConfig.OutputDir = authDir
	authConfig.ServiceType = config.ServiceAuth
	authConfig.ModuleName = g.config.ModuleName + "/services/auth_service"
	authConfig.Cache = nil
	authConfig.RateLimiter = nil
	authConfig.AuthCache = false

	authGen := NewGenerator(&authConfig)
	if err := authGen.generateAuth(); err != nil {
		return fmt.Errorf("failed to generate auth: %w", err)
	}

	// Generate go.work file
	goWork := fmt.Sprintf(`go %s

use (
	./services/gateway
	./services/auth_service
)
`, g.data.GoVersion)

	if err := g.writer.WriteFile("go.work", []byte(goWork)); err != nil {
		return err
	}

	return nil
}

// generateCommon generates common files (logger, telemetry, config)
func (g *Generator) generateCommon(writer *writers.FileWriter) error {
	// Logger
	if err := writer.WriteTemplate("common/utils/logger.go.tmpl", "internals/utils/logger.go"); err != nil {
		return err
	}

	// TLS configuration
	if err := writer.WriteTemplate("common/utils/tls.go.tmpl", "internals/utils/tls.go"); err != nil {
		return err
	}

	// Telemetry (if enabled)
	if g.config.Observability {
		if err := writer.WriteTemplate("common/telemetry/telemetry.go.tmpl", "internals/telemetry/telemetry.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate("common/telemetry/http_metrics.go.tmpl", "internals/telemetry/http_metrics.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate("common/telemetry/http_tracing.go.tmpl", "internals/telemetry/http_tracing.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate("common/telemetry/db_observability.go.tmpl", "internals/telemetry/db_observability.go"); err != nil {
			return err
		}
	}

	// Config loader
	configFile := fmt.Sprintf("common/config/%s.go.tmpl", g.config.ConfigSource)
	if err := writer.WriteTemplate(configFile, fmt.Sprintf("internals/config/%s.go", g.config.ConfigSource)); err != nil {
		return err
	}

	// Common middleware
	if err := writer.WriteTemplate("common/middleware/recovery.go.tmpl", "internals/middleware/recovery.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("common/middleware/logging.go.tmpl", "internals/middleware/logging.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("common/middleware/security_headers.go.tmpl", "internals/middleware/security_headers.go"); err != nil {
		return err
	}
	if err := writer.WriteTemplate("common/middleware/chain.go.tmpl", "internals/middleware/chain.go"); err != nil {
		return err
	}

	return nil
}

// generateDatabase generates database layer files
func (g *Generator) generateDatabase(writer *writers.FileWriter) error {
	dbType := string(*g.config.Database)

	// Generate interface
	if err := writer.WriteTemplate("database/interface.go.tmpl", "internals/db/interface.go"); err != nil {
		return err
	}

	// Generate client
	clientFile := fmt.Sprintf("database/%s/client.go.tmpl", dbType)
	if err := writer.WriteTemplate(clientFile, "internals/db/client.go"); err != nil {
		return err
	}

	// For SQL databases, also copy the specific repository files
	if g.config.Database.IsSQL() {
		if err := writer.WriteTemplate(fmt.Sprintf("database/%s/user.go.tmpl", dbType), "internals/db/user.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate(fmt.Sprintf("database/%s/session.go.tmpl", dbType), "internals/db/session.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate(fmt.Sprintf("database/%s/consent.go.tmpl", dbType), "internals/db/consent.go"); err != nil {
			return err
		}
		if err := writer.WriteTemplate(fmt.Sprintf("database/%s/mfa.go.tmpl", dbType), "internals/db/mfa.go"); err != nil {
			return err
		}
		if g.data.HasGDPRProcessingLogs {
			if err := writer.WriteTemplate(fmt.Sprintf("database/%s/data_processing_log.go.tmpl", dbType), "internals/db/data_processing_log.go"); err != nil {
				return err
			}
		}

		// Generate tenancy database files if enabled
		if g.config.EnableTenancy {
			if err := writer.WriteTemplate(fmt.Sprintf("database/%s/tenant.go.tmpl", dbType), "internals/db/tenant.go"); err != nil {
				return err
			}
			if err := writer.WriteTemplate(fmt.Sprintf("database/%s/invitation.go.tmpl", dbType), "internals/db/invitation.go"); err != nil {
				return err
			}
			if err := writer.WriteTemplate(fmt.Sprintf("database/%s/user_tenant.go.tmpl", dbType), "internals/db/user_tenant.go"); err != nil {
				return err
			}
		}

		// Generate OAuth token database file if OAuth enabled
		if g.data.HasOAuth {
			if err := writer.WriteTemplate(fmt.Sprintf("database/%s/oauth_token.go.tmpl", dbType), "internals/db/oauth_token.go"); err != nil {
				return err
			}
		}

		// Copy migrations
		if err := writer.EnsureDir("internals/db/migrations"); err != nil {
			return err
		}

		// Generate core migration files
		migrations := []string{"001_users.sql", "002_sessions.sql", "003_consents.sql", "004_mfa.sql"}
		for _, mig := range migrations {
			tmplFile := fmt.Sprintf("database/%s/migrations/%s.tmpl", dbType, mig)
			outFile := fmt.Sprintf("internals/db/migrations/%s", mig)
			if err := writer.WriteTemplate(tmplFile, outFile); err != nil {
				return err
			}
		}

		// Generate GDPR-related migrations
		gdprMigrations := []struct {
			file      string
			condition bool
		}{
			{"005_data_processing_logs.sql", g.data.HasGDPRProcessingLogs},
			{"006_email_verification.sql", g.data.HasEmail},
			{"007_password_reset.sql", g.data.HasEmail},
			{"008_casbin_rules.sql", g.data.EnableRBAC},
			{"009_oauth.sql", g.data.HasOAuth},
			{"010_tenants.sql", g.data.HasTenancy},
		}

		for _, mig := range gdprMigrations {
			if mig.condition {
				tmplFile := fmt.Sprintf("database/%s/migrations/%s.tmpl", dbType, mig.file)
				outFile := fmt.Sprintf("internals/db/migrations/%s", mig.file)
				if err := writer.WriteTemplate(tmplFile, outFile); err != nil {
					return err
				}
			}
		}
	} else {
		// For non-SQL databases (MongoDB, ArangoDB), copy repository files
		// These may genuinely not exist for all NoSQL implementations
		if err := writer.WriteTemplate(fmt.Sprintf("database/%s/user.go.tmpl", dbType), "internals/db/user.go"); err != nil {
			// NoSQL template might not exist
		}
		if err := writer.WriteTemplate(fmt.Sprintf("database/%s/session.go.tmpl", dbType), "internals/db/session.go"); err != nil {
		}
		if err := writer.WriteTemplate(fmt.Sprintf("database/%s/consent.go.tmpl", dbType), "internals/db/consent.go"); err != nil {
		}
		if err := writer.WriteTemplate(fmt.Sprintf("database/%s/mfa.go.tmpl", dbType), "internals/db/mfa.go"); err != nil {
		}
		if g.data.HasGDPRProcessingLogs {
			if err := writer.WriteTemplate(fmt.Sprintf("database/%s/data_processing_log.go.tmpl", dbType), "internals/db/data_processing_log.go"); err != nil {
			}
		}
	}

	return nil
}

// generateStaticFiles generates static files
func (g *Generator) generateStaticFiles() error {
	// Copy .gitignore
	if err := g.writer.CopyStatic("static/.gitignore", ".gitignore"); err != nil {
		return err
	}

	// Generate Makefile
	if err := g.writer.WriteTemplate("static/Makefile.tmpl", "Makefile"); err != nil {
		return err
	}

	// Generate docker-compose.yml
	if err := g.writer.WriteTemplate("static/docker-compose.yml.tmpl", "docker-compose.yml"); err != nil {
		return err
	}

	// Generate README.md
	if err := g.writer.WriteTemplate("static/README.md.tmpl", "README.md"); err != nil {
		return err
	}

	// Generate test infrastructure (if using docker and not gateway-only)
	if g.data.HasDockerTests && g.data.ServiceType != "gateway" {
		if err := g.writer.WriteTemplate("static/docker-compose.test.yml.tmpl", "docker-compose.test.yml"); err != nil {
			return err
		}
	}

	// Generate test script
	if err := g.writer.EnsureDir("scripts"); err != nil {
		return err
	}
	if err := g.writer.WriteTemplate("static/scripts/run-tests.sh.tmpl", "scripts/run-tests.sh"); err != nil {
		return err
	}
	// Make script executable
	scriptPath := filepath.Join(g.writer.OutputDir(), "scripts", "run-tests.sh")
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make run-tests.sh executable: %w", err)
	}

	return nil
}

// generateK8s generates k3s-based Kubernetes manifests for local dev environment
func (g *Generator) generateK8s() error {
	// Create directory structure
	dirs := []string{
		"k8s",
		"k8s/ingress",
		"k8s/services",
		"k8s/scripts",
	}

	if g.data.HasCache {
		dirs = append(dirs, "k8s/redis")
	}

	if g.data.HasObservability {
		dirs = append(dirs,
			"k8s/observability",
			"k8s/observability/jaeger",
			"k8s/observability/otel-collector",
			"k8s/observability/prometheus",
			"k8s/observability/fluent-bit",
			"k8s/observability/grafana",
		)
	}

	for _, dir := range dirs {
		if err := g.writer.EnsureDir(dir); err != nil {
			return err
		}
	}

	// Generate namespace
	if err := g.writer.WriteTemplate("k8s/namespace.yaml.tmpl", "k8s/namespace.yaml"); err != nil {
		return err
	}

	// Generate ingress
	if err := g.writer.WriteTemplate("k8s/ingress/helmchart.yaml.tmpl", "k8s/ingress/helmchart.yaml"); err != nil {
		return err
	}
	if err := g.writer.WriteTemplate("k8s/ingress/ingress.yaml.tmpl", "k8s/ingress/ingress.yaml"); err != nil {
		return err
	}
	if err := g.writer.WriteTemplate("k8s/ingress/tls-secret.yaml.tmpl", "k8s/ingress/tls-secret.yaml"); err != nil {
		// Optional - created by setup.sh
	}

	// Generate external services
	if err := g.writer.WriteTemplate("k8s/services/gateway-ext.yaml.tmpl", "k8s/services/gateway-ext.yaml"); err != nil {
		return err
	}
	if err := g.writer.WriteTemplate("k8s/services/auth-ext.yaml.tmpl", "k8s/services/auth-ext.yaml"); err != nil {
		return err
	}
	if g.data.HasFrontend {
		if err := g.writer.WriteTemplate("k8s/services/frontend-ext.yaml.tmpl", "k8s/services/frontend-ext.yaml"); err != nil {
			return err
		}
	}

	// Generate Redis manifests if cache uses Redis/Valkey
	if g.data.HasCache {
		redisFiles := map[string]string{
			"k8s/redis/deployment.yaml.tmpl": "k8s/redis/deployment.yaml",
			"k8s/redis/pvc.yaml.tmpl":        "k8s/redis/pvc.yaml",
			"k8s/redis/service.yaml.tmpl":    "k8s/redis/service.yaml",
		}
		for tmpl, out := range redisFiles {
			if err := g.writer.WriteTemplate(tmpl, out); err != nil {
				return err
			}
		}
	}

	// Generate observability stack if enabled
	if g.data.HasObservability {
		// Jaeger
		jaegerFiles := map[string]string{
			"k8s/observability/jaeger/configmap.yaml.tmpl":  "k8s/observability/jaeger/configmap.yaml",
			"k8s/observability/jaeger/deployment.yaml.tmpl": "k8s/observability/jaeger/deployment.yaml",
			"k8s/observability/jaeger/pvc.yaml.tmpl":        "k8s/observability/jaeger/pvc.yaml",
			"k8s/observability/jaeger/service.yaml.tmpl":    "k8s/observability/jaeger/service.yaml",
		}
		for tmpl, out := range jaegerFiles {
			if err := g.writer.WriteTemplate(tmpl, out); err != nil {
				return err
			}
		}

		// OTel Collector
		otelFiles := map[string]string{
			"k8s/observability/otel-collector/configmap.yaml.tmpl":  "k8s/observability/otel-collector/configmap.yaml",
			"k8s/observability/otel-collector/deployment.yaml.tmpl": "k8s/observability/otel-collector/deployment.yaml",
			"k8s/observability/otel-collector/service.yaml.tmpl":    "k8s/observability/otel-collector/service.yaml",
		}
		for tmpl, out := range otelFiles {
			if err := g.writer.WriteTemplate(tmpl, out); err != nil {
				return err
			}
		}

		// Prometheus (kube-prometheus-stack HelmChart)
		if err := g.writer.WriteTemplate("k8s/observability/prometheus/helmchart.yaml.tmpl", "k8s/observability/prometheus/helmchart.yaml"); err != nil {
			return err
		}
		if err := g.writer.WriteTemplate("k8s/observability/prometheus/redis-servicemonitor.yaml.tmpl", "k8s/observability/prometheus/redis-servicemonitor.yaml"); err != nil {
			// Optional - only useful when Redis is deployed
		}

		// Fluent Bit
		if err := g.writer.WriteTemplate("k8s/observability/fluent-bit/helmchart.yaml.tmpl", "k8s/observability/fluent-bit/helmchart.yaml"); err != nil {
			return err
		}

		// Grafana dashboards
		if err := g.writer.WriteTemplate("k8s/observability/grafana/dashboards-configmap.yaml.tmpl", "k8s/observability/grafana/dashboards-configmap.yaml"); err != nil {
			return err
		}
	}

	// Generate scripts
	if err := g.writer.WriteTemplate("k8s/scripts/setup.sh.tmpl", "k8s/scripts/setup.sh"); err != nil {
		return err
	}
	if err := g.writer.WriteTemplate("k8s/scripts/teardown.sh.tmpl", "k8s/scripts/teardown.sh"); err != nil {
		return err
	}

	// Make scripts executable
	setupPath := filepath.Join(g.writer.OutputDir(), "k8s", "scripts", "setup.sh")
	if err := os.Chmod(setupPath, 0755); err != nil {
		return fmt.Errorf("failed to make setup.sh executable: %w", err)
	}
	teardownPath := filepath.Join(g.writer.OutputDir(), "k8s", "scripts", "teardown.sh")
	if err := os.Chmod(teardownPath, 0755); err != nil {
		return fmt.Errorf("failed to make teardown.sh executable: %w", err)
	}

	return nil
}

// runGoModTidy runs go mod tidy in the output directory
func (g *Generator) runGoModTidy() error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = g.writer.OutputDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runGoFmt runs gofmt on the output directory
func (g *Generator) runGoFmt() error {
	cmd := exec.Command("gofmt", "-w", ".")
	cmd.Dir = g.writer.OutputDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runGoModTidyIn runs go mod tidy in a specific directory
func (g *Generator) runGoModTidyIn(dir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runGoFmtIn runs gofmt in a specific directory
func (g *Generator) runGoFmtIn(dir string) error {
	cmd := exec.Command("gofmt", "-w", ".")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// generateGatewayTests generates test files for gateway service
func (g *Generator) generateGatewayTests(writer *writers.FileWriter) error {
	// Generate service test helpers
	if err := writer.WriteTemplate("gateway/services/testhelpers_test.go.tmpl", "internals/services/testhelpers_test.go"); err != nil {
		// Test helpers might not exist yet
	}

	// Generate cache tests if cache is enabled
	if g.config.Cache != nil {
		if err := writer.WriteTemplate("gateway/cache/cache_test.go.tmpl", "internals/cache/cache_test.go"); err != nil {
			// Cache tests might not exist
		}
	}

	// Generate rate limiter tests if rate limiter is enabled
	if g.config.RateLimiter != nil {
		if err := writer.WriteTemplate("gateway/ratelimiter/ratelimiter_test.go.tmpl", "internals/ratelimiter/ratelimiter_test.go"); err != nil {
			// Rate limiter tests might not exist
		}
	}

	return nil
}

// generateAuthTests generates test files for auth service
func (g *Generator) generateAuthTests(writer *writers.FileWriter) error {
	// Generate test helpers
	if err := writer.WriteTemplate("auth/services/testhelpers_test.go.tmpl", "internals/services/testhelpers_test.go"); err != nil {
		// Test helpers might not exist yet
	}

	// Generate auth service unit tests
	if err := writer.WriteTemplate("auth/services/auth_service_test.go.tmpl", "internals/services/auth_service_test.go"); err != nil {
		// Tests might not exist
	}

	// Generate handler tests
	if err := writer.WriteTemplate("auth/services/handlers_test.go.tmpl", "internals/services/handlers_test.go"); err != nil {
		// Tests might not exist
	}

	// Generate integration tests
	if err := writer.WriteTemplate("auth/services/integration_test.go.tmpl", "internals/services/integration_test.go"); err != nil {
		// Tests might not exist
	}

	// Generate cache tests if cache is enabled (optional gateway feature)
	if g.config.Cache != nil {
		if err := writer.WriteTemplate("gateway/cache/cache_test.go.tmpl", "internals/cache/cache_test.go"); err != nil {
			// Cache tests might not exist
		}
	}

	// Generate rate limiter tests if rate limiter is enabled (optional gateway feature)
	if g.config.RateLimiter != nil {
		if err := writer.WriteTemplate("gateway/ratelimiter/ratelimiter_test.go.tmpl", "internals/ratelimiter/ratelimiter_test.go"); err != nil {
			// Rate limiter tests might not exist
		}
	}

	// Generate database tests
	if g.config.Database != nil {
		if err := g.generateDatabaseTests(writer); err != nil {
			// DB tests might not exist
		}
	}

	return nil
}

// generateDatabaseTests generates test files for database layer
func (g *Generator) generateDatabaseTests(writer *writers.FileWriter) error {
	dbType := string(*g.config.Database)

	// Generate database test helpers
	tmplFile := fmt.Sprintf("database/%s/db_test.go.tmpl", dbType)
	if err := writer.WriteTemplate(tmplFile, "internals/db/db_test.go"); err != nil {
		// DB tests might not exist for all DB types
	}

	return nil
}

// generateFrontend generates frontend applications
func (g *Generator) generateFrontend() error {
	if len(g.config.Frontends) == 0 {
		return nil
	}

	for _, frontend := range g.config.Frontends {
		var frontendDir string
		switch frontend {
		case config.FrontendWeb:
			frontendDir = "frontend/web"
		case config.FrontendMobile:
			frontendDir = "frontend/mobile"
		}

		// If only one frontend, use simpler directory structure
		if len(g.config.Frontends) == 1 {
			frontendDir = "frontend"
		}

		writer := writers.NewFileWriter(filepath.Join(g.writer.OutputDir(), frontendDir), g.data)

		// Create base directory
		if err := os.MkdirAll(writer.OutputDir(), 0755); err != nil {
			return fmt.Errorf("failed to create frontend directory: %w", err)
		}

		switch frontend {
		case config.FrontendWeb:
			if err := g.generateWebFrontend(writer); err != nil {
				return err
			}
		case config.FrontendMobile:
			if err := g.generateMobileFrontend(writer); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateWebFrontend generates web frontend based on framework choice
func (g *Generator) generateWebFrontend(writer *writers.FileWriter) error {
	if g.config.WebFramework == nil {
		return fmt.Errorf("web framework is required for web frontend generation")
	}
	framework := *g.config.WebFramework

	// Create directory structure
	dirs := []string{
		"src",
		"src/components",
		"src/components/auth",
		"src/components/settings",
		"src/components/layout",
		"src/components/ui",
		"src/routes",
		"src/routes/settings",
		"src/lib",
		"src/lib/api",
		"src/lib/hooks",
		"src/lib/types",
		"src/lib/stores",
		"public",
	}

	for _, dir := range dirs {
		if err := writer.EnsureDir(dir); err != nil {
			return err
		}
	}

	// Determine framework path
	var frameworkPath string
	switch framework {
	case config.FrameworkReact:
		frameworkPath = "frontend/web/react-vite"
	case config.FrameworkNext:
		frameworkPath = "frontend/web/nextjs"
	case config.FrameworkTanStack:
		frameworkPath = "frontend/web/tanstack"
	default:
		frameworkPath = "frontend/web/react-vite"
	}

	// Generate package.json
	if err := writer.WriteTemplate(frameworkPath+"/package.json.tmpl", "package.json"); err != nil {
		return err
	}

	// Generate config files
	configFiles := map[string]string{
		frameworkPath + "/tsconfig.json.tmpl":   "tsconfig.json",
		frameworkPath + "/.env.example.tmpl":    ".env.example",
	}

	// Framework-specific config files
	switch framework {
	case config.FrameworkReact:
		configFiles[frameworkPath+"/vite.config.ts.tmpl"] = "vite.config.ts"
		configFiles[frameworkPath+"/index.html.tmpl"] = "index.html"
	case config.FrameworkNext:
		configFiles[frameworkPath+"/next.config.js.tmpl"] = "next.config.js"
	case config.FrameworkTanStack:
		configFiles[frameworkPath+"/app.config.ts.tmpl"] = "app.config.ts"
	}

	for tmpl, out := range configFiles {
		if err := writer.WriteTemplate(tmpl, out); err != nil {
			// Config file might not exist, continue
		}
	}

	// Generate shared library files (API client, types, hooks)
	sharedFiles := map[string]string{
		"frontend/shared/api/client.ts.tmpl":       "src/lib/api/client.ts",
		"frontend/shared/api/auth.ts.tmpl":         "src/lib/api/auth.ts",
		"frontend/shared/types/auth.ts.tmpl":       "src/lib/types/auth.ts",
		"frontend/shared/hooks/use-auth.ts.tmpl":   "src/lib/hooks/use-auth.ts",
		"frontend/shared/hooks/use-sessions.ts.tmpl": "src/lib/hooks/use-sessions.ts",
	}

	if g.data.HasGDPR {
		sharedFiles["frontend/shared/hooks/use-gdpr.ts.tmpl"] = "src/lib/hooks/use-gdpr.ts"
	}

	if g.data.HasTenancy {
		sharedFiles["frontend/shared/api/tenancy.ts.tmpl"] = "src/lib/api/tenancy.ts"
		sharedFiles["frontend/shared/types/tenancy.ts.tmpl"] = "src/lib/types/tenancy.ts"
	}

	for tmpl, out := range sharedFiles {
		if err := writer.WriteTemplate(tmpl, out); err != nil {
			// Shared file might not exist, continue
		}
	}

	// Generate state management files
	if err := g.generateStateManagement(writer); err != nil {
		return err
	}

	// Generate UI library files
	if err := g.generateUILibrary(writer); err != nil {
		return err
	}

	// Generate framework-specific source files
	sourceFiles := map[string]string{
		frameworkPath + "/src/main.tsx.tmpl": "src/main.tsx",
		frameworkPath + "/src/App.tsx.tmpl":  "src/App.tsx",
	}

	// Layout components
	sourceFiles[frameworkPath+"/src/components/layout/AuthLayout.tsx.tmpl"] = "src/components/layout/AuthLayout.tsx"
	sourceFiles[frameworkPath+"/src/components/layout/DashboardLayout.tsx.tmpl"] = "src/components/layout/DashboardLayout.tsx"
	sourceFiles[frameworkPath+"/src/components/layout/Navbar.tsx.tmpl"] = "src/components/layout/Navbar.tsx"

	// Auth components
	sourceFiles[frameworkPath+"/src/components/auth/LoginForm.tsx.tmpl"] = "src/components/auth/LoginForm.tsx"
	sourceFiles[frameworkPath+"/src/components/auth/RegisterForm.tsx.tmpl"] = "src/components/auth/RegisterForm.tsx"
	if g.data.HasOAuth {
		sourceFiles[frameworkPath+"/src/components/auth/OAuthButtons.tsx.tmpl"] = "src/components/auth/OAuthButtons.tsx"
	}

	// Settings components
	sourceFiles[frameworkPath+"/src/components/settings/SessionList.tsx.tmpl"] = "src/components/settings/SessionList.tsx"
	sourceFiles[frameworkPath+"/src/components/settings/PasswordChange.tsx.tmpl"] = "src/components/settings/PasswordChange.tsx"

	// Route pages
	sourceFiles[frameworkPath+"/src/routes/index.tsx.tmpl"] = "src/routes/index.tsx"
	sourceFiles[frameworkPath+"/src/routes/login.tsx.tmpl"] = "src/routes/login.tsx"
	sourceFiles[frameworkPath+"/src/routes/register.tsx.tmpl"] = "src/routes/register.tsx"
	sourceFiles[frameworkPath+"/src/routes/forgot-password.tsx.tmpl"] = "src/routes/forgot-password.tsx"
	sourceFiles[frameworkPath+"/src/routes/settings.tsx.tmpl"] = "src/routes/settings.tsx"
	sourceFiles[frameworkPath+"/src/routes/settings/profile.tsx.tmpl"] = "src/routes/settings/profile.tsx"
	sourceFiles[frameworkPath+"/src/routes/settings/security.tsx.tmpl"] = "src/routes/settings/security.tsx"
	if g.data.HasGDPR {
		sourceFiles[frameworkPath+"/src/routes/settings/privacy.tsx.tmpl"] = "src/routes/settings/privacy.tsx"
	}

	// Tenancy routes
	if g.data.HasTenancy {
		sourceFiles[frameworkPath+"/src/routes/workspace-selector.tsx.tmpl"] = "src/routes/workspace-selector.tsx"
	}

	// MFA routes
	if g.data.EnableMFA {
		sourceFiles[frameworkPath+"/src/routes/mfa-setup.tsx.tmpl"] = "src/routes/mfa-setup.tsx"
		sourceFiles[frameworkPath+"/src/routes/mfa-verify.tsx.tmpl"] = "src/routes/mfa-verify.tsx"
	}

	// OAuth callback route
	if g.data.HasOAuth {
		sourceFiles[frameworkPath+"/src/routes/oauth-callback.tsx.tmpl"] = "src/routes/oauth-callback.tsx"
	}

	// Lib utils
	sourceFiles[frameworkPath+"/src/lib/utils.ts.tmpl"] = "src/lib/utils.ts"

	for tmpl, out := range sourceFiles {
		if err := writer.WriteTemplate(tmpl, out); err != nil {
			// Source file might not exist, continue
		}
	}

	// Generate CSS if using Shadcn (Tailwind)
	if g.data.HasShadcn {
		if err := g.generateTailwindCSS(writer); err != nil {
			return err
		}
	}

	// Generate test files (React + Vite uses Vitest)
	if framework == config.FrameworkReact {
		if err := g.generateWebTests(writer, frameworkPath); err != nil {
			return err
		}
	}

	// Generate E2E test files if enabled
	if g.data.HasCypress || g.data.HasPlaywright {
		if err := g.generateE2ETests(writer, frameworkPath); err != nil {
			return err
		}
	}

	return nil
}

// generateWebTests generates test files for web frontend
func (g *Generator) generateWebTests(writer *writers.FileWriter, frameworkPath string) error {
	// Create test directories
	testDirs := []string{
		"src/test",
		"src/test/mocks",
	}
	for _, dir := range testDirs {
		if err := writer.EnsureDir(dir); err != nil {
			return err
		}
	}

	// Generate test config and setup files
	testFiles := map[string]string{
		frameworkPath + "/vitest.config.ts.tmpl":           "vitest.config.ts",
		frameworkPath + "/src/test/setup.ts.tmpl":          "src/test/setup.ts",
		frameworkPath + "/src/test/test-utils.tsx.tmpl":    "src/test/test-utils.tsx",
		frameworkPath + "/src/test/mocks/handlers.ts.tmpl": "src/test/mocks/handlers.ts",
		frameworkPath + "/src/test/mocks/server.ts.tmpl":   "src/test/mocks/server.ts",
	}

	for tmpl, out := range testFiles {
		if err := writer.WriteTemplate(tmpl, out); err != nil {
			// Test file might not exist, continue
		}
	}

	// Generate component test files
	componentTests := map[string]string{
		frameworkPath + "/src/components/auth/LoginForm.test.tsx.tmpl": "src/components/auth/LoginForm.test.tsx",
	}

	for tmpl, out := range componentTests {
		if err := writer.WriteTemplate(tmpl, out); err != nil {
			// Test file might not exist, continue
		}
	}

	return nil
}

// generateE2ETests generates E2E test files for web frontend
func (g *Generator) generateE2ETests(writer *writers.FileWriter, frameworkPath string) error {
	// Create E2E directories
	if err := writer.EnsureDir("e2e/support"); err != nil {
		return err
	}

	if g.data.HasCypress {
		// Generate Cypress config and support files
		cypressFiles := map[string]string{
			frameworkPath + "/cypress.config.ts.tmpl":       "cypress.config.ts",
			frameworkPath + "/e2e/support/e2e.ts.tmpl":      "e2e/support/e2e.ts",
			frameworkPath + "/e2e/support/commands.ts.tmpl": "e2e/support/commands.ts",
			frameworkPath + "/e2e/auth.cy.ts.tmpl":          "e2e/auth.cy.ts",
		}

		// Add MFA tests if MFA is enabled
		if g.data.EnableMFA {
			cypressFiles[frameworkPath+"/e2e/mfa.cy.ts.tmpl"] = "e2e/mfa.cy.ts"
		}

		// Add GDPR tests if GDPR is enabled
		if g.data.HasGDPR {
			cypressFiles[frameworkPath+"/e2e/gdpr.cy.ts.tmpl"] = "e2e/gdpr.cy.ts"
		}

		for tmpl, out := range cypressFiles {
			if err := writer.WriteTemplate(tmpl, out); err != nil {
				// E2E file might not exist, continue
			}
		}
	}

	if g.data.HasPlaywright {
		// Generate Playwright config and test files
		playwrightFiles := map[string]string{
			frameworkPath + "/playwright.config.ts.tmpl": "playwright.config.ts",
			frameworkPath + "/e2e/auth.spec.ts.tmpl":     "e2e/auth.spec.ts",
		}

		// Add MFA tests if MFA is enabled
		if g.data.EnableMFA {
			playwrightFiles[frameworkPath+"/e2e/mfa.spec.ts.tmpl"] = "e2e/mfa.spec.ts"
		}

		// Add GDPR tests if GDPR is enabled
		if g.data.HasGDPR {
			playwrightFiles[frameworkPath+"/e2e/gdpr.spec.ts.tmpl"] = "e2e/gdpr.spec.ts"
		}

		for tmpl, out := range playwrightFiles {
			if err := writer.WriteTemplate(tmpl, out); err != nil {
				// E2E file might not exist, continue
			}
		}
	}

	return nil
}

// generateMobileFrontend generates React Native Expo frontend
func (g *Generator) generateMobileFrontend(writer *writers.FileWriter) error {
	// Create directory structure
	dirs := []string{
		"app",
		"app/(auth)",
		"app/(tabs)",
		"app/(tabs)/settings",
		"components",
		"components/auth",
		"components/settings",
		"components/ui",
		"lib",
		"lib/api",
		"lib/hooks",
		"lib/types",
		"lib/stores",
	}

	for _, dir := range dirs {
		if err := writer.EnsureDir(dir); err != nil {
			return err
		}
	}

	// Generate package.json
	if err := writer.WriteTemplate("frontend/mobile/expo/package.json.tmpl", "package.json"); err != nil {
		return err
	}

	// Generate config files
	configFiles := map[string]string{
		"frontend/mobile/expo/tsconfig.json.tmpl":    "tsconfig.json",
		"frontend/mobile/expo/app.json.tmpl":         "app.json",
		"frontend/mobile/expo/babel.config.js.tmpl":  "babel.config.js",
		"frontend/mobile/expo/eas.json.tmpl":         "eas.json",
	}

	for tmpl, out := range configFiles {
		if err := writer.WriteTemplate(tmpl, out); err != nil {
			// Config file might not exist, continue
		}
	}

	// Generate shared library files
	sharedFiles := map[string]string{
		"frontend/shared/api/client.ts.tmpl":       "lib/api/client.ts",
		"frontend/shared/api/auth.ts.tmpl":         "lib/api/auth.ts",
		"frontend/shared/types/auth.ts.tmpl":       "lib/types/auth.ts",
		"frontend/shared/hooks/use-auth.ts.tmpl":   "lib/hooks/use-auth.ts",
		"frontend/shared/hooks/use-sessions.ts.tmpl": "lib/hooks/use-sessions.ts",
	}

	if g.data.HasGDPR {
		sharedFiles["frontend/shared/hooks/use-gdpr.ts.tmpl"] = "lib/hooks/use-gdpr.ts"
	}

	if g.data.HasTenancy {
		sharedFiles["frontend/shared/api/tenancy.ts.tmpl"] = "lib/api/tenancy.ts"
		sharedFiles["frontend/shared/types/tenancy.ts.tmpl"] = "lib/types/tenancy.ts"
	}

	for tmpl, out := range sharedFiles {
		if err := writer.WriteTemplate(tmpl, out); err != nil {
			// Shared file might not exist, continue
		}
	}

	// Generate state management files
	if err := g.generateStateManagement(writer); err != nil {
		return err
	}

	// Generate mobile test files
	if err := g.generateMobileTests(writer); err != nil {
		return err
	}

	return nil
}

// generateMobileTests generates test files for mobile frontend
func (g *Generator) generateMobileTests(writer *writers.FileWriter) error {
	// Create test directories
	testDirs := []string{
		"__tests__",
		"__tests__/components",
		"__tests__/components/auth",
	}
	for _, dir := range testDirs {
		if err := writer.EnsureDir(dir); err != nil {
			return err
		}
	}

	// Generate test files
	testFiles := map[string]string{
		"frontend/mobile/expo/__tests__/components/auth/LoginForm.test.tsx.tmpl": "__tests__/components/auth/LoginForm.test.tsx",
	}

	for tmpl, out := range testFiles {
		if err := writer.WriteTemplate(tmpl, out); err != nil {
			// Test file might not exist, continue
		}
	}

	return nil
}

// generateStateManagement generates state management files based on choice
func (g *Generator) generateStateManagement(writer *writers.FileWriter) error {
	if g.config.StateManagement == nil {
		return nil
	}

	switch *g.config.StateManagement {
	case config.StateMgmtTanStack:
		// TanStack Query + Zustand
		files := map[string]string{
			"frontend/state/tanstack/providers/query-provider.tsx.tmpl": "src/lib/providers/query-provider.tsx",
			"frontend/state/tanstack/stores/auth-store.ts.tmpl":         "src/lib/stores/auth-store.ts",
		}
		// Ensure providers directory exists
		writer.EnsureDir("src/lib/providers")

		for tmpl, out := range files {
			if err := writer.WriteTemplate(tmpl, out); err != nil {
				// File might not exist, continue
			}
		}

	case config.StateMgmtRedux:
		// Redux Toolkit + RTK Query
		files := map[string]string{
			"frontend/state/redux/store/index.ts.tmpl":            "src/lib/store/index.ts",
			"frontend/state/redux/store/auth-slice.ts.tmpl":       "src/lib/store/auth-slice.ts",
			"frontend/state/redux/providers/store-provider.tsx.tmpl": "src/lib/providers/store-provider.tsx",
		}
		// Ensure directories exist
		writer.EnsureDir("src/lib/store")
		writer.EnsureDir("src/lib/providers")

		for tmpl, out := range files {
			if err := writer.WriteTemplate(tmpl, out); err != nil {
				// File might not exist, continue
			}
		}
	}

	return nil
}

// generateUILibrary generates UI library files based on choice
func (g *Generator) generateUILibrary(writer *writers.FileWriter) error {
	if g.config.UILibrary == nil {
		return nil
	}

	switch *g.config.UILibrary {
	case config.UILibShadcn:
		// ShadcnUI components
		files := map[string]string{
			"frontend/ui/shadcn/components/ui/button.tsx.tmpl":    "src/components/ui/button.tsx",
			"frontend/ui/shadcn/components/ui/input.tsx.tmpl":     "src/components/ui/input.tsx",
			"frontend/ui/shadcn/components/ui/label.tsx.tmpl":     "src/components/ui/label.tsx",
			"frontend/ui/shadcn/components/ui/card.tsx.tmpl":      "src/components/ui/card.tsx",
			"frontend/ui/shadcn/components/ui/dropdown-menu.tsx.tmpl": "src/components/ui/dropdown-menu.tsx",
			"frontend/ui/shadcn/components/ui/switch.tsx.tmpl":    "src/components/ui/switch.tsx",
		}
		for tmpl, out := range files {
			if err := writer.WriteTemplate(tmpl, out); err != nil {
				// File might not exist, continue
			}
		}

	case config.UILibBaseUI:
		// BaseUI doesn't need component files - uses npm package directly
		// Just generate theme customization if needed
		if err := writer.WriteTemplate("frontend/ui/baseui/lib/theme.ts.tmpl", "src/lib/theme.ts"); err != nil {
			// File might not exist, continue
		}
	}

	return nil
}

// generateTailwindCSS generates Tailwind CSS configuration and files
func (g *Generator) generateTailwindCSS(writer *writers.FileWriter) error {
	// Generate tailwind.config.js
	tailwindConfig := `/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
    },
  },
  plugins: [],
}
`
	if err := writer.WriteFile("tailwind.config.js", []byte(tailwindConfig)); err != nil {
		return err
	}

	// Generate postcss.config.js
	postcssConfig := `export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
`
	if err := writer.WriteFile("postcss.config.js", []byte(postcssConfig)); err != nil {
		return err
	}

	// Generate index.css
	indexCSS := `@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --muted: 210 40% 96.1%;
    --muted-foreground: 215.4 16.3% 46.9%;
    --accent: 210 40% 96.1%;
    --accent-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 222.2 84% 4.9%;
    --radius: 0.5rem;
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;
    --primary: 210 40% 98%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 212.7 26.8% 83.9%;
  }
}

@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground;
  }
}
`
	if err := writer.WriteFile("src/index.css", []byte(indexCSS)); err != nil {
		return err
	}

	return nil
}

package config

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
)

// ServiceType represents the type of service to generate
type ServiceType string

const (
	ServiceGateway ServiceType = "gateway"
	ServiceAuth    ServiceType = "auth"
	ServiceBoth    ServiceType = "both"
)

func (s ServiceType) String() string {
	return string(s)
}

// ValidServiceTypes returns all valid service types
func ValidServiceTypes() []string {
	return []string{
		string(ServiceGateway),
		string(ServiceAuth),
		string(ServiceBoth),
	}
}

// DatabaseType represents the database backend
type DatabaseType string

const (
	DBPostgres DatabaseType = "postgres"
	DBMySQL    DatabaseType = "mysql"
	DBSQLite   DatabaseType = "sqlite"
	DBMongoDB  DatabaseType = "mongodb"
	DBArangoDB DatabaseType = "arangodb"
)

func (d DatabaseType) String() string {
	return string(d)
}

// IsSQL returns true if the database is a SQL database
func (d DatabaseType) IsSQL() bool {
	switch d {
	case DBPostgres, DBMySQL, DBSQLite:
		return true
	default:
		return false
	}
}

// IsNoSQL returns true if the database is a NoSQL database
func (d DatabaseType) IsNoSQL() bool {
	switch d {
	case DBMongoDB, DBArangoDB:
		return true
	default:
		return false
	}
}

// ValidDatabaseTypes returns all valid database types
func ValidDatabaseTypes() []string {
	return []string{
		string(DBPostgres),
		string(DBMySQL),
		string(DBSQLite),
		string(DBMongoDB),
		string(DBArangoDB),
	}
}

// CacheType represents the caching backend
type CacheType string

const (
	CacheRedis  CacheType = "redis"
	CacheValkey CacheType = "valkey"
	CacheMemory CacheType = "memory"
)

func (c CacheType) String() string {
	return string(c)
}

// ValidCacheTypes returns all valid cache types
func ValidCacheTypes() []string {
	return []string{
		string(CacheRedis),
		string(CacheValkey),
		string(CacheMemory),
	}
}

// ConfigSource represents where configuration is loaded from
type ConfigSource string

const (
	ConfigYAML  ConfigSource = "yaml"
	ConfigEnv   ConfigSource = "env"
	ConfigVault ConfigSource = "vault"
)

func (c ConfigSource) String() string {
	return string(c)
}

// ValidConfigSources returns all valid config sources
func ValidConfigSources() []string {
	return []string{
		string(ConfigYAML),
		string(ConfigEnv),
		string(ConfigVault),
	}
}

// RateLimiterType represents the rate limiting algorithm
type RateLimiterType string

const (
	RateLimiterTokenBucket   RateLimiterType = "token-bucket"
	RateLimiterSlidingWindow RateLimiterType = "sliding-window"
	RateLimiterLeakyBucket   RateLimiterType = "leaky-bucket"
	RateLimiterFixedWindow   RateLimiterType = "fixed-window"
)

func (r RateLimiterType) String() string {
	return string(r)
}

// ValidRateLimiterTypes returns all valid rate limiter types
func ValidRateLimiterTypes() []string {
	return []string{
		string(RateLimiterTokenBucket),
		string(RateLimiterSlidingWindow),
		string(RateLimiterLeakyBucket),
		string(RateLimiterFixedWindow),
	}
}

// OAuthProvider represents an OAuth provider
type OAuthProvider string

const (
	OAuthGoogle    OAuthProvider = "google"
	OAuthMicrosoft OAuthProvider = "microsoft"
	OAuthApple     OAuthProvider = "apple"
)

func (o OAuthProvider) String() string {
	return string(o)
}

// ValidOAuthProviders returns all valid OAuth providers
func ValidOAuthProviders() []string {
	return []string{
		string(OAuthGoogle),
		string(OAuthMicrosoft),
		string(OAuthApple),
	}
}

// GDPRFeature represents a GDPR compliance feature
type GDPRFeature string

const (
	GDPRConsent        GDPRFeature = "consent"
	GDPRDataExport     GDPRFeature = "data-export"
	GDPRDataDeletion   GDPRFeature = "data-deletion"
	GDPRProcessingLogs GDPRFeature = "processing-logs"
)

func (g GDPRFeature) String() string {
	return string(g)
}

// ValidGDPRFeatures returns all valid GDPR features
func ValidGDPRFeatures() []string {
	return []string{
		string(GDPRConsent),
		string(GDPRDataExport),
		string(GDPRDataDeletion),
		string(GDPRProcessingLogs),
	}
}

// EmailService represents an email service provider
type EmailService string

const (
	EmailSES  EmailService = "ses"
	EmailSMTP EmailService = "smtp"
)

func (e EmailService) String() string {
	return string(e)
}

// ValidEmailServices returns all valid email services
func ValidEmailServices() []string {
	return []string{
		string(EmailSES),
		string(EmailSMTP),
	}
}

// FrontendType represents the type of frontend to generate
type FrontendType string

const (
	FrontendWeb    FrontendType = "web"
	FrontendMobile FrontendType = "mobile"
)

func (f FrontendType) String() string {
	return string(f)
}

// ValidFrontendTypes returns all valid frontend types
func ValidFrontendTypes() []string {
	return []string{
		string(FrontendWeb),
		string(FrontendMobile),
	}
}

// WebFramework represents the web framework choice
type WebFramework string

const (
	FrameworkReact    WebFramework = "react"    // React + Vite
	FrameworkNext     WebFramework = "next"     // Next.js
	FrameworkTanStack WebFramework = "tanstack" // TanStack Start
)

func (w WebFramework) String() string {
	return string(w)
}

// ValidWebFrameworks returns all valid web frameworks
func ValidWebFrameworks() []string {
	return []string{
		string(FrameworkReact),
		string(FrameworkNext),
		string(FrameworkTanStack),
	}
}

// UILibrary represents the UI component library choice
type UILibrary string

const (
	UILibShadcn UILibrary = "shadcn"
	UILibBaseUI UILibrary = "baseui"
)

func (u UILibrary) String() string {
	return string(u)
}

// ValidUILibraries returns all valid UI libraries
func ValidUILibraries() []string {
	return []string{
		string(UILibShadcn),
		string(UILibBaseUI),
	}
}

// StateManagement represents the state management choice
type StateManagement string

const (
	StateMgmtTanStack StateManagement = "tanstack" // TanStack Query + Zustand
	StateMgmtRedux    StateManagement = "redux"    // Redux Toolkit + RTK Query
)

func (s StateManagement) String() string {
	return string(s)
}

// ValidStateManagements returns all valid state management options
func ValidStateManagements() []string {
	return []string{
		string(StateMgmtTanStack),
		string(StateMgmtRedux),
	}
}

// TestInfraType represents the test infrastructure choice
type TestInfraType string

const (
	TestInfraDocker         TestInfraType = "docker"
	TestInfraTestcontainers TestInfraType = "testcontainers"
)

func (t TestInfraType) String() string {
	return string(t)
}

// ValidTestInfraTypes returns all valid test infrastructure types
func ValidTestInfraTypes() []string {
	return []string{
		string(TestInfraDocker),
		string(TestInfraTestcontainers),
	}
}

// E2EFrameworkType represents the E2E testing framework choice
type E2EFrameworkType string

const (
	E2EFrameworkCypress    E2EFrameworkType = "cypress"
	E2EFrameworkPlaywright E2EFrameworkType = "playwright"
)

func (e E2EFrameworkType) String() string {
	return string(e)
}

// ValidE2EFrameworkTypes returns all valid E2E framework types
func ValidE2EFrameworkTypes() []string {
	return []string{
		string(E2EFrameworkCypress),
		string(E2EFrameworkPlaywright),
	}
}

// ProjectConfig holds all configuration for project generation
type ProjectConfig struct {
	ModuleName  string // e.g., github.com/user/project
	ProjectName string // e.g., project (extracted from module)
	OutputDir   string // Where to generate
	ServiceType ServiceType

	// Gateway features (required for gateway, optional for auth)
	Cache       *CacheType       // Required for gateway, optional for auth
	RateLimiter *RateLimiterType // Required for gateway, optional for auth
	AuthCache   bool             // Enable auth response caching (gateway only)

	// Auth-specific
	Database       *DatabaseType   // Required for auth
	OAuthProviders []OAuthProvider // OAuth providers (google, microsoft, apple)
	EnableMFA      bool            // Enable MFA/TOTP support
	EnableRBAC     bool            // Enable Casbin RBAC
	GDPRFeatures   []GDPRFeature   // GDPR compliance features
	EmailService   *EmailService   // Email service (ses or smtp) - required if GDPR features selected

	// Shared
	ConfigSource  ConfigSource // yaml, env, or vault
	Observability bool         // Enable OTLP observability

	// Frontend options (optional add-on)
	Frontends       []FrontendType   // web, mobile, or both
	WebFramework    *WebFramework    // react, next, tanstack (only for web frontend)
	UILibrary       *UILibrary       // shadcn or baseui
	StateManagement *StateManagement // tanstack or redux

	// Frontend analytics/monitoring (optional)
	EnablePostHog bool // Enable PostHog analytics
	EnableSentry  bool // Enable Sentry error tracking

	// Testing options
	TestInfra    *TestInfraType    // docker (default) or testcontainers
	E2EFramework *E2EFrameworkType // cypress (default) or playwright - only when frontend enabled
}

// ExtractProjectName extracts the project name from a module path
func ExtractProjectName(moduleName string) string {
	return path.Base(moduleName)
}

// Validate validates the project configuration
func (p *ProjectConfig) Validate() error {
	if p.ModuleName == "" {
		return fmt.Errorf("module name is required")
	}

	if !isValidModuleName(p.ModuleName) {
		return fmt.Errorf("invalid module name format: %s", p.ModuleName)
	}

	if p.ServiceType == "" {
		return fmt.Errorf("service type is required")
	}

	// Validate gateway requirements
	if p.ServiceType == ServiceGateway || p.ServiceType == ServiceBoth {
		if p.Cache == nil {
			return fmt.Errorf("cache is required for gateway service")
		}
		if p.RateLimiter == nil {
			return fmt.Errorf("rate limiter is required for gateway service")
		}
	}

	// Validate auth requirements
	if p.ServiceType == ServiceAuth || p.ServiceType == ServiceBoth {
		if p.Database == nil {
			return fmt.Errorf("database is required for auth service")
		}
	}

	if p.ConfigSource == "" {
		return fmt.Errorf("config source is required")
	}

	// Validate GDPR email requirement
	if len(p.GDPRFeatures) > 0 && p.EmailService == nil {
		return fmt.Errorf("email service is required when GDPR features are enabled")
	}

	// Validate frontend options
	if len(p.Frontends) > 0 {
		// Frontend is only allowed with auth or both services
		if p.ServiceType == ServiceGateway {
			return fmt.Errorf("frontend is only available for auth or both service types")
		}

		// Check if web frontend is selected
		hasWeb := false
		for _, f := range p.Frontends {
			if f == FrontendWeb {
				hasWeb = true
				break
			}
		}

		// Web frontend requires web framework
		if hasWeb && p.WebFramework == nil {
			return fmt.Errorf("web framework is required for web frontend")
		}

		// Only React+Vite is currently implemented
		if hasWeb && p.WebFramework != nil && *p.WebFramework != FrameworkReact {
			return fmt.Errorf("web framework %q is not yet implemented, only %q is currently available", *p.WebFramework, FrameworkReact)
		}

		// UI library is required when frontend is selected
		if p.UILibrary == nil {
			return fmt.Errorf("UI library is required when frontend is selected")
		}

		// State management is required when frontend is selected
		if p.StateManagement == nil {
			return fmt.Errorf("state management is required when frontend is selected")
		}
	}

	return nil
}

func isValidModuleName(name string) bool {
	// Simple validation for Go module names
	pattern := `^[a-zA-Z0-9][a-zA-Z0-9._-]*(/[a-zA-Z0-9][a-zA-Z0-9._-]*)*$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}

// TemplateData holds all data passed to templates during generation
type TemplateData struct {
	// Project info
	ModuleName  string // e.g., "github.com/user/project"
	ProjectName string // e.g., "project" (extracted)

	// Service configuration
	ServiceType string // "gateway", "auth", or "both"
	ServiceName string // e.g., "gateway" or "auth_service"

	// Feature selections
	Database        string  // "postgres", "mysql", "sqlite", "mongodb", "arangodb"
	Cache           *string // nil, "redis", "valkey", "memory" (pointer for nil check)
	CacheType       string  // "redis", "valkey", "memory" (direct string for templates)
	Config          string  // "yaml", "env", "vault"
	RateLimiter     *string // nil, "token-bucket", "sliding-window", "leaky-bucket", "fixed-window"
	RateLimiterType string  // direct string for templates

	// Auth features
	OAuthProviders  []string // "google", "microsoft", "apple"
	EnableMFA       bool     // MFA/TOTP support
	EnableRBAC      bool     // Casbin RBAC
	GDPRFeatures    []string // "consent", "data-export", "data-deletion", "processing-logs"
	EmailService    string   // "ses" or "smtp"
	AuthCache       bool     // Gateway auth response caching

	// Computed helpers (used in templates)
	HasCache         bool
	HasRateLimiter   bool
	HasObservability bool
	IsSQL            bool // postgres, mysql, sqlite
	IsNoSQL          bool // mongodb, arangodb

	// OAuth helpers
	HasOAuth          bool
	HasGoogleOAuth    bool
	HasMicrosoftOAuth bool
	HasAppleOAuth     bool

	// GDPR helpers
	HasGDPR               bool
	HasGDPRConsent        bool
	HasGDPRDataExport     bool
	HasGDPRDataDeletion   bool
	HasGDPRProcessingLogs bool

	// Email helpers
	HasEmail     bool
	HasSESEmail  bool
	HasSMTPEmail bool

	// Frontend
	HasFrontend      bool
	IsWebFrontend    bool
	IsMobileFrontend bool
	FrontendType     string // "web" or "mobile"
	WebFramework     string // "react", "next", "tanstack"
	UILibrary        string // "shadcn", "baseui"
	StateManagement  string // "tanstack", "redux"

	// Frontend computed helpers
	HasShadcn        bool
	HasBaseUI        bool
	HasTanStackQuery bool
	HasRedux         bool
	HasZustand       bool
	IsReactVite      bool
	IsNextJS         bool
	IsTanStackStart  bool

	// Frontend analytics/monitoring
	HasPostHog bool
	HasSentry  bool

	// Testing
	TestInfra         string // "docker" or "testcontainers"
	E2EFramework      string // "cypress" or "playwright"
	HasDockerTests    bool
	HasTestcontainers bool
	HasCypress        bool
	HasPlaywright     bool

	// API connection for frontend
	APIBaseURL string

	// Metadata
	Year      int
	GoVersion string
}

// NewTemplateData creates TemplateData from ProjectConfig
func NewTemplateData(cfg *ProjectConfig) *TemplateData {
	data := &TemplateData{
		ModuleName:       cfg.ModuleName,
		ProjectName:      ExtractProjectName(cfg.ModuleName),
		ServiceType:      string(cfg.ServiceType),
		Config:           string(cfg.ConfigSource),
		HasObservability: cfg.Observability,
		Year:             time.Now().Year(),
		GoVersion:        "1.24.0",
		EnableMFA:        cfg.EnableMFA,
		EnableRBAC:       cfg.EnableRBAC,
		AuthCache:        cfg.AuthCache,
		APIBaseURL:       "http://localhost:8080", // Default API URL for frontend
	}

	if cfg.Database != nil {
		data.Database = string(*cfg.Database)
		data.IsSQL = cfg.Database.IsSQL()
		data.IsNoSQL = cfg.Database.IsNoSQL()
	}

	if cfg.Cache != nil {
		cacheStr := string(*cfg.Cache)
		data.Cache = &cacheStr
		data.CacheType = cacheStr
		data.HasCache = true
	}

	if cfg.RateLimiter != nil {
		rlStr := string(*cfg.RateLimiter)
		data.RateLimiter = &rlStr
		data.RateLimiterType = rlStr
		data.HasRateLimiter = true
	}

	// OAuth providers
	if len(cfg.OAuthProviders) > 0 {
		data.HasOAuth = true
		data.OAuthProviders = make([]string, len(cfg.OAuthProviders))
		for i, p := range cfg.OAuthProviders {
			data.OAuthProviders[i] = string(p)
			switch p {
			case OAuthGoogle:
				data.HasGoogleOAuth = true
			case OAuthMicrosoft:
				data.HasMicrosoftOAuth = true
			case OAuthApple:
				data.HasAppleOAuth = true
			}
		}
	}

	// GDPR features
	if len(cfg.GDPRFeatures) > 0 {
		data.HasGDPR = true
		data.GDPRFeatures = make([]string, len(cfg.GDPRFeatures))
		for i, f := range cfg.GDPRFeatures {
			data.GDPRFeatures[i] = string(f)
			switch f {
			case GDPRConsent:
				data.HasGDPRConsent = true
			case GDPRDataExport:
				data.HasGDPRDataExport = true
			case GDPRDataDeletion:
				data.HasGDPRDataDeletion = true
			case GDPRProcessingLogs:
				data.HasGDPRProcessingLogs = true
			}
		}
	}

	// Email service
	if cfg.EmailService != nil {
		data.HasEmail = true
		data.EmailService = string(*cfg.EmailService)
		switch *cfg.EmailService {
		case EmailSES:
			data.HasSESEmail = true
		case EmailSMTP:
			data.HasSMTPEmail = true
		}
	}

	// Frontend
	if len(cfg.Frontends) > 0 {
		data.HasFrontend = true

		for _, f := range cfg.Frontends {
			switch f {
			case FrontendWeb:
				data.IsWebFrontend = true
			case FrontendMobile:
				data.IsMobileFrontend = true
			}
		}

		// Set FrontendType for backward compatibility (use first one)
		if len(cfg.Frontends) == 1 {
			data.FrontendType = string(cfg.Frontends[0])
		} else {
			data.FrontendType = "both"
		}

		if cfg.WebFramework != nil {
			data.WebFramework = string(*cfg.WebFramework)
			switch *cfg.WebFramework {
			case FrameworkReact:
				data.IsReactVite = true
			case FrameworkNext:
				data.IsNextJS = true
			case FrameworkTanStack:
				data.IsTanStackStart = true
			}
		}

		if cfg.UILibrary != nil {
			data.UILibrary = string(*cfg.UILibrary)
			switch *cfg.UILibrary {
			case UILibShadcn:
				data.HasShadcn = true
			case UILibBaseUI:
				data.HasBaseUI = true
			}
		}

		if cfg.StateManagement != nil {
			data.StateManagement = string(*cfg.StateManagement)
			switch *cfg.StateManagement {
			case StateMgmtTanStack:
				data.HasTanStackQuery = true
				data.HasZustand = true
			case StateMgmtRedux:
				data.HasRedux = true
			}
		}

		// Analytics/monitoring
		data.HasPostHog = cfg.EnablePostHog
		data.HasSentry = cfg.EnableSentry
	}

	// Testing options
	if cfg.TestInfra != nil {
		data.TestInfra = string(*cfg.TestInfra)
		data.HasDockerTests = *cfg.TestInfra == TestInfraDocker
		data.HasTestcontainers = *cfg.TestInfra == TestInfraTestcontainers
	} else {
		// Default to docker
		data.TestInfra = string(TestInfraDocker)
		data.HasDockerTests = true
	}

	if cfg.E2EFramework != nil {
		data.E2EFramework = string(*cfg.E2EFramework)
		data.HasCypress = *cfg.E2EFramework == E2EFrameworkCypress
		data.HasPlaywright = *cfg.E2EFramework == E2EFrameworkPlaywright
	} else if data.HasFrontend {
		// Default to cypress when frontend is enabled
		data.E2EFramework = string(E2EFrameworkCypress)
		data.HasCypress = true
	}

	return data
}

// ForGateway returns a copy of TemplateData configured for gateway service
func (t *TemplateData) ForGateway() *TemplateData {
	copy := *t
	copy.ServiceName = "gateway"
	return &copy
}

// ForAuth returns a copy of TemplateData configured for auth service
func (t *TemplateData) ForAuth() *TemplateData {
	copy := *t
	copy.ServiceName = "auth_service"
	return &copy
}

// DatabaseDriver returns the database driver import path
func (t *TemplateData) DatabaseDriver() string {
	switch t.Database {
	case "postgres":
		return "github.com/jackc/pgx/v5"
	case "mysql":
		return "github.com/go-sql-driver/mysql"
	case "sqlite":
		return "modernc.org/sqlite"
	case "mongodb":
		return "go.mongodb.org/mongo-driver/mongo"
	case "arangodb":
		return "github.com/arangodb/go-driver/v2"
	default:
		return ""
	}
}

// CachePackage returns the cache package based on cache type
func (t *TemplateData) CachePackage() string {
	if t.Cache == nil {
		return ""
	}
	switch *t.Cache {
	case "redis", "valkey":
		return "github.com/redis/go-redis/v9"
	default:
		return ""
	}
}

// Title returns a title-cased string (helper for templates)
func Title(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

package config

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"slices"
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
)

func (d DatabaseType) String() string {
	return string(d)
}

// ValidDatabaseTypes returns all valid database types
func ValidDatabaseTypes() []string {
	return []string{
		string(DBPostgres),
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
	SkipTidy    bool   // Skip `go mod tidy` after generation (faster audit iterations)
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

	// Multi-tenancy (auth-first model)
	EnableTenancy bool // Enable multi-tenant workspaces

	// K8s dev environment (k3s)
	EnableK8s bool   // Generate k3s-based local dev environment
	Domain    string // Local dev domain (e.g., "myapp.dev") - required when EnableK8s is true

	// Testing options
	TestInfra    *TestInfraType    // docker (default) or testcontainers
	E2EFramework *E2EFrameworkType // cypress (default) or playwright - only when frontend enabled
}

// ExtractProjectName extracts the project name from a module path
func ExtractProjectName(moduleName string) string {
	return path.Base(moduleName)
}

// Validate is the single source of truth for configuration validity. It is used
// by both the interactive wizard and the flag-driven command path. All violations
// are collected and returned together via errors.Join so the caller can surface
// every problem at once rather than one at a time.
func (p *ProjectConfig) Validate() error {
	var errs []error

	// Required identity fields
	if p.ModuleName == "" {
		errs = append(errs, errors.New("module name is required"))
	} else if !IsValidModuleName(p.ModuleName) {
		errs = append(errs, fmt.Errorf("invalid module name format: %s", p.ModuleName))
	}

	if p.ServiceType == "" {
		errs = append(errs, errors.New("service type is required"))
	} else if !slices.Contains(ValidServiceTypes(), string(p.ServiceType)) {
		errs = append(errs, fmt.Errorf("invalid service type %q, valid options: %s", p.ServiceType, strings.Join(ValidServiceTypes(), ", ")))
	}

	if p.ConfigSource == "" {
		errs = append(errs, errors.New("config source is required"))
	} else if !slices.Contains(ValidConfigSources(), string(p.ConfigSource)) {
		errs = append(errs, fmt.Errorf("invalid config source %q, valid options: %s", p.ConfigSource, strings.Join(ValidConfigSources(), ", ")))
	}

	isGateway := p.ServiceType == ServiceGateway
	isAuth := p.ServiceType == ServiceAuth
	isBoth := p.ServiceType == ServiceBoth

	// Gateway requirements
	if isGateway || isBoth {
		if p.Cache == nil {
			errs = append(errs, errors.New("cache is required for gateway service"))
		}
		if p.RateLimiter == nil {
			errs = append(errs, errors.New("rate limiter is required for gateway service"))
		}
	}
	if p.Cache != nil && !slices.Contains(ValidCacheTypes(), string(*p.Cache)) {
		errs = append(errs, fmt.Errorf("invalid cache type %q, valid options: %s", *p.Cache, strings.Join(ValidCacheTypes(), ", ")))
	}
	if p.RateLimiter != nil && !slices.Contains(ValidRateLimiterTypes(), string(*p.RateLimiter)) {
		errs = append(errs, fmt.Errorf("invalid rate limiter %q, valid options: %s", *p.RateLimiter, strings.Join(ValidRateLimiterTypes(), ", ")))
	}

	// Auth requirements
	if isAuth || isBoth {
		if p.Database == nil {
			errs = append(errs, errors.New("database is required for auth service"))
		}
	}
	if p.Database != nil && !slices.Contains(ValidDatabaseTypes(), string(*p.Database)) {
		errs = append(errs, fmt.Errorf("invalid database type %q, valid options: %s", *p.Database, strings.Join(ValidDatabaseTypes(), ", ")))
	}

	// AuthCache is only valid for gateway or both
	if p.AuthCache && isAuth {
		errs = append(errs, errors.New("auth cache is only available for gateway or both service types"))
	}

	// Auth-specific features require auth or both service
	if isGateway {
		if len(p.OAuthProviders) > 0 {
			errs = append(errs, errors.New("OAuth providers require auth or both service type"))
		}
		if p.EnableMFA {
			errs = append(errs, errors.New("MFA requires auth or both service type"))
		}
		if p.EnableRBAC {
			errs = append(errs, errors.New("RBAC requires auth or both service type"))
		}
		if len(p.GDPRFeatures) > 0 {
			errs = append(errs, errors.New("GDPR features require auth or both service type"))
		}
	}

	for _, prov := range p.OAuthProviders {
		if !slices.Contains(ValidOAuthProviders(), string(prov)) {
			errs = append(errs, fmt.Errorf("invalid OAuth provider %q, valid options: %s", prov, strings.Join(ValidOAuthProviders(), ", ")))
		}
	}
	for _, f := range p.GDPRFeatures {
		if !slices.Contains(ValidGDPRFeatures(), string(f)) {
			errs = append(errs, fmt.Errorf("invalid GDPR feature %q, valid options: %s", f, strings.Join(ValidGDPRFeatures(), ", ")))
		}
	}
	if p.EmailService != nil && !slices.Contains(ValidEmailServices(), string(*p.EmailService)) {
		errs = append(errs, fmt.Errorf("invalid email service %q, valid options: %s", *p.EmailService, strings.Join(ValidEmailServices(), ", ")))
	}
	if len(p.GDPRFeatures) > 0 && p.EmailService == nil {
		errs = append(errs, errors.New("email service is required when GDPR features are enabled"))
	}

	// Tenancy requires auth or both service
	if p.EnableTenancy && isGateway {
		errs = append(errs, errors.New("tenancy requires auth or both service type"))
	}

	// K8s requires a valid domain, and a domain without K8s is an orphan
	if p.EnableK8s && p.Domain == "" {
		errs = append(errs, errors.New("domain is required when k8s dev environment is enabled"))
	}
	if p.Domain != "" {
		if !isValidDomain(p.Domain) {
			errs = append(errs, fmt.Errorf("invalid domain format %q: must be a valid domain (e.g., myapp.dev)", p.Domain))
		}
		if !p.EnableK8s {
			errs = append(errs, fmt.Errorf("domain %q is set but k8s dev environment is not enabled", p.Domain))
		}
	}

	errs = append(errs, p.validateFrontend()...)

	return errors.Join(errs...)
}

// validateFrontend collects all frontend, analytics, and testing-related violations.
func (p *ProjectConfig) validateFrontend() []error {
	var errs []error
	hasFrontend := len(p.Frontends) > 0

	if hasFrontend {
		if p.ServiceType == ServiceGateway {
			errs = append(errs, errors.New("frontend is only available for auth or both service types"))
		}

		hasWeb := false
		for _, f := range p.Frontends {
			if !slices.Contains(ValidFrontendTypes(), string(f)) {
				errs = append(errs, fmt.Errorf("invalid frontend type %q, valid options: %s", f, strings.Join(ValidFrontendTypes(), ", ")))
			}
			if f == FrontendWeb {
				hasWeb = true
			}
		}

		if hasWeb {
			switch {
			case p.WebFramework == nil:
				errs = append(errs, errors.New("web framework is required for web frontend"))
			case !slices.Contains(ValidWebFrameworks(), string(*p.WebFramework)):
				errs = append(errs, fmt.Errorf("invalid web framework %q, valid options: %s", *p.WebFramework, strings.Join(ValidWebFrameworks(), ", ")))
			case *p.WebFramework != FrameworkReact:
				errs = append(errs, fmt.Errorf("web framework %q is not yet implemented, only %q is currently available", *p.WebFramework, FrameworkReact))
			}
		}

		if p.UILibrary == nil {
			errs = append(errs, errors.New("UI library is required when frontend is selected"))
		} else if !slices.Contains(ValidUILibraries(), string(*p.UILibrary)) {
			errs = append(errs, fmt.Errorf("invalid UI library %q, valid options: %s", *p.UILibrary, strings.Join(ValidUILibraries(), ", ")))
		}

		if p.StateManagement == nil {
			errs = append(errs, errors.New("state management is required when frontend is selected"))
		} else if !slices.Contains(ValidStateManagements(), string(*p.StateManagement)) {
			errs = append(errs, fmt.Errorf("invalid state management %q, valid options: %s", *p.StateManagement, strings.Join(ValidStateManagements(), ", ")))
		}
	} else {
		// Frontend-dependent flags set without any frontend are orphans
		if p.WebFramework != nil {
			errs = append(errs, errors.New("web framework requires a frontend to be selected"))
		}
		if p.UILibrary != nil {
			errs = append(errs, errors.New("UI library requires a frontend to be selected"))
		}
		if p.StateManagement != nil {
			errs = append(errs, errors.New("state management requires a frontend to be selected"))
		}
		if p.EnablePostHog {
			errs = append(errs, errors.New("PostHog analytics requires a frontend to be selected"))
		}
		if p.EnableSentry {
			errs = append(errs, errors.New("Sentry error tracking requires a frontend to be selected"))
		}
	}

	if p.TestInfra != nil && !slices.Contains(ValidTestInfraTypes(), string(*p.TestInfra)) {
		errs = append(errs, fmt.Errorf("invalid test infrastructure %q, valid options: %s", *p.TestInfra, strings.Join(ValidTestInfraTypes(), ", ")))
	}
	if p.E2EFramework != nil {
		if !slices.Contains(ValidE2EFrameworkTypes(), string(*p.E2EFramework)) {
			errs = append(errs, fmt.Errorf("invalid E2E framework %q, valid options: %s", *p.E2EFramework, strings.Join(ValidE2EFrameworkTypes(), ", ")))
		}
		if !hasFrontend {
			errs = append(errs, errors.New("E2E framework requires a frontend to be selected"))
		}
	}

	return errs
}

// IsValidModuleName reports whether name is a syntactically valid Go module path.
func IsValidModuleName(name string) bool {
	pattern := `^[a-zA-Z0-9][a-zA-Z0-9._-]*(/[a-zA-Z0-9][a-zA-Z0-9._-]*)*$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}

func isValidDomain(domain string) bool {
	// Domain must have at least one dot and only valid characters
	pattern := `^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)+$`
	matched, _ := regexp.MatchString(pattern, domain)
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
	Database        string  // "postgres"
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

	// Multi-tenancy
	HasTenancy bool

	// K8s dev environment
	HasK8s bool
	Domain string

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
		GoVersion:        "1.25.0",
		EnableMFA:        cfg.EnableMFA,
		EnableRBAC:       cfg.EnableRBAC,
		AuthCache:        cfg.AuthCache,
		APIBaseURL:       "http://localhost:8080", // Default API URL for frontend
	}

	if cfg.Database != nil {
		data.Database = string(*cfg.Database)
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

	// Multi-tenancy
	data.HasTenancy = cfg.EnableTenancy

	// K8s dev environment
	data.HasK8s = cfg.EnableK8s
	data.Domain = cfg.Domain

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

// Title returns a title-cased string (helper for templates)
func Title(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

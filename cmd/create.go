package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"

	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/generator"
	"github.com/DenisBytes/gonstrukt/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func CreateCmd() *cobra.Command {
	var (
		serviceTypeStr    string
		databaseStr       string
		cacheStr          string
		configStr         string
		rateLimiterStr    string
		observabilityBool bool
		interactive       bool
		oauthProviders    []string
		enableMFA         bool
		enableRBAC        bool
		gdprFeatures      []string
		emailServiceStr   string
		authCache         bool
		frontends         []string
		webFrameworkStr   string
		uiLibraryStr      string
		stateMgmtStr      string
		enablePostHog     bool
		enableSentry      bool
		testInfraStr      string
		e2eFrameworkStr   string
		enableTenancy     bool
		enableK8s         bool
		domainStr         string
	)

	cmd := &cobra.Command{
		Use:   "create [module]",
		Short: "Create a new Go service with specified configuration",
		Long: `Create a new Go service (gateway, auth, or both) with database, caching,
configuration, observability, and rate limiting options.

Without flags, launches an interactive TUI wizard to configure the project.
With flags, creates the project directly without prompts.

Examples:
  # Interactive mode (TUI wizard)
  gonstrukt create

  # Non-interactive mode with flags
  gonstrukt create github.com/user/myproject -s gateway --cache redis -r token-bucket --config yaml

  # Create auth service with PostgreSQL
  gonstrukt create github.com/user/myauth -s auth -d postgres --config vault

  # Create both services (monorepo)
  gonstrukt create github.com/user/myapp -s both -d postgres --cache redis -r token-bucket --config vault`,
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine if we should use interactive mode
			isTTY := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

			// If no flags provided and it's a TTY, use interactive mode
			useInteractive := interactive && isTTY && !hasAnyFlag(cmd)

			if useInteractive {
				return runInteractive()
			}

			// Non-interactive mode - validate and run
			return runNonInteractive(cmd, args, serviceTypeStr, databaseStr, cacheStr, configStr, rateLimiterStr, observabilityBool, oauthProviders, enableMFA, enableRBAC, gdprFeatures, emailServiceStr, authCache, frontends, webFrameworkStr, uiLibraryStr, stateMgmtStr, enablePostHog, enableSentry, testInfraStr, e2eFrameworkStr, enableTenancy, enableK8s, domainStr)
		},
	}

	cmd.Flags().StringVarP(&serviceTypeStr, "service", "s", "", "Service type (gateway, auth, both)")
	cmd.Flags().StringVarP(&databaseStr, "database", "d", "", "Database type (postgres, mysql, sqlite, mongodb, arangodb)")
	cmd.Flags().StringVar(&cacheStr, "cache", "", "Cache type (redis, valkey, memory)")
	cmd.Flags().StringVarP(&configStr, "config", "c", "", "Configuration source (yaml, env, vault)")
	cmd.Flags().StringVarP(&rateLimiterStr, "rate-limiter", "r", "", "Rate limiting algorithm (token-bucket, sliding-window, leaky-bucket, fixed-window)")
	cmd.Flags().BoolVarP(&observabilityBool, "observability", "o", true, "Enable OTLP observability")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", true, "Use interactive TUI wizard")
	cmd.Flags().StringSliceVar(&oauthProviders, "oauth", nil, "OAuth providers (google, microsoft, apple)")
	cmd.Flags().BoolVar(&enableMFA, "mfa", false, "Enable MFA/TOTP support")
	cmd.Flags().BoolVar(&enableRBAC, "rbac", false, "Enable Casbin RBAC")
	cmd.Flags().StringSliceVar(&gdprFeatures, "gdpr", nil, "GDPR features (consent, data-export, data-deletion, processing-logs)")
	cmd.Flags().StringVar(&emailServiceStr, "email", "", "Email service (ses, smtp)")
	cmd.Flags().BoolVar(&authCache, "auth-cache", false, "Enable auth response caching (gateway)")
	cmd.Flags().StringSliceVar(&frontends, "frontend", nil, "Frontend types (web, mobile) - can specify both")
	cmd.Flags().StringVar(&webFrameworkStr, "web-framework", "", "Web framework (react, next, tanstack) - required for web frontend")
	cmd.Flags().StringVar(&uiLibraryStr, "ui-lib", "", "UI library (shadcn, baseui)")
	cmd.Flags().StringVar(&stateMgmtStr, "state-mgmt", "", "State management (tanstack, redux)")
	cmd.Flags().BoolVar(&enablePostHog, "posthog", false, "Enable PostHog analytics (requires --frontend)")
	cmd.Flags().BoolVar(&enableSentry, "sentry", false, "Enable Sentry error tracking (requires --frontend)")
	cmd.Flags().StringVar(&testInfraStr, "test-infra", "docker", "Test infrastructure (docker, testcontainers)")
	cmd.Flags().StringVar(&e2eFrameworkStr, "e2e-framework", "cypress", "E2E test framework (cypress, playwright) - only with --frontend")
	cmd.Flags().BoolVar(&enableTenancy, "tenancy", false, "Enable auth-first multi-tenancy")
	cmd.Flags().BoolVar(&enableK8s, "k8s", false, "Generate k3s-based local dev environment")
	cmd.Flags().StringVar(&domainStr, "domain", "", "Local dev domain for k8s (e.g., myapp.dev)")

	cmd.RegisterFlagCompletionFunc("service", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidServiceTypes(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("database", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidDatabaseTypes(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("cache", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidCacheTypes(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("config", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidConfigSources(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("rate-limiter", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidRateLimiterTypes(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("frontend", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidFrontendTypes(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("web-framework", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidWebFrameworks(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("ui-lib", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidUILibraries(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("state-mgmt", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidStateManagements(), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

// hasAnyFlag checks if any relevant flags were set
func hasAnyFlag(cmd *cobra.Command) bool {
	flags := []string{"service", "database", "cache", "config", "rate-limiter", "frontend", "web-framework", "ui-lib", "state-mgmt", "oauth", "mfa", "rbac", "gdpr", "email", "auth-cache", "observability", "posthog", "sentry", "test-infra", "e2e-framework", "tenancy", "k8s", "domain"}
	for _, name := range flags {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

// runInteractive runs the TUI wizard
func runInteractive() error {
	wizard := tui.NewWizard()

	p := tea.NewProgram(wizard, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Get the final wizard state
	w, ok := finalModel.(*tui.Wizard)
	if !ok {
		return fmt.Errorf("unexpected model type")
	}

	cfg := w.Config()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Generate the project
	fmt.Println("\nGenerating project...")

	gen := generator.NewGenerator(cfg)
	if err := gen.Generate(context.Background()); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Printf("\n✓ Project generated successfully at: %s\n", cfg.ProjectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", cfg.ProjectName)
	fmt.Println("  go build ./...")

	return nil
}

// runNonInteractive runs with command-line flags
func runNonInteractive(cmd *cobra.Command, args []string, serviceTypeStr, databaseStr, cacheStr, configStr, rateLimiterStr string, observability bool, oauthProviders []string, enableMFA, enableRBAC bool, gdprFeatures []string, emailServiceStr string, authCache bool, frontends []string, webFrameworkStr, uiLibraryStr, stateMgmtStr string, enablePostHog, enableSentry bool, testInfraStr, e2eFrameworkStr string, enableTenancy, enableK8s bool, domainStr string) error {
	var validationErrors []error

	// Module name is required in non-interactive mode
	if len(args) == 0 {
		validationErrors = append(validationErrors, errors.New("module name is required"))
	}

	var moduleName string
	if len(args) > 0 {
		moduleName = args[0]
		if err := validateModuleName(moduleName); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	// Service type is required
	if serviceTypeStr == "" {
		validationErrors = append(validationErrors, errors.New("--service flag is required"))
	} else if !slices.Contains(config.ValidServiceTypes(), serviceTypeStr) {
		validationErrors = append(validationErrors, fmt.Errorf("invalid service type '%s', valid options: %v", serviceTypeStr, config.ValidServiceTypes()))
	}

	// Config source is required
	if configStr == "" {
		validationErrors = append(validationErrors, errors.New("--config flag is required"))
	} else if !slices.Contains(config.ValidConfigSources(), configStr) {
		validationErrors = append(validationErrors, fmt.Errorf("invalid config source '%s', valid options: %v", configStr, config.ValidConfigSources()))
	}

	// Validate based on service type
	serviceType := config.ServiceType(serviceTypeStr)

	// AuthCache only valid for gateway or both
	if authCache && serviceType == config.ServiceAuth {
		validationErrors = append(validationErrors, errors.New("--auth-cache is only available for gateway or both service types"))
	}

	// Gateway or both requires cache and rate limiter
	if serviceType == config.ServiceGateway || serviceType == config.ServiceBoth {
		if cacheStr == "" {
			validationErrors = append(validationErrors, errors.New("--cache is required for gateway service"))
		}
		if rateLimiterStr == "" {
			validationErrors = append(validationErrors, errors.New("--rate-limiter is required for gateway service"))
		}
	}

	// Validate cache type if provided (required for gateway, optional for auth)
	if cacheStr != "" && !slices.Contains(config.ValidCacheTypes(), cacheStr) {
		validationErrors = append(validationErrors, fmt.Errorf("invalid cache type '%s', valid options: %v", cacheStr, config.ValidCacheTypes()))
	}

	// Validate rate limiter if provided (required for gateway, optional for auth)
	if rateLimiterStr != "" && !slices.Contains(config.ValidRateLimiterTypes(), rateLimiterStr) {
		validationErrors = append(validationErrors, fmt.Errorf("invalid rate limiter '%s', valid options: %v", rateLimiterStr, config.ValidRateLimiterTypes()))
	}

	// Auth or both requires database
	if serviceType == config.ServiceAuth || serviceType == config.ServiceBoth {
		if databaseStr == "" {
			validationErrors = append(validationErrors, errors.New("--database is required for auth service"))
		} else if !slices.Contains(config.ValidDatabaseTypes(), databaseStr) {
			validationErrors = append(validationErrors, fmt.Errorf("invalid database type '%s', valid options: %v", databaseStr, config.ValidDatabaseTypes()))
		}
	}

	// Validate OAuth providers
	for _, p := range oauthProviders {
		if !slices.Contains(config.ValidOAuthProviders(), p) {
			validationErrors = append(validationErrors, fmt.Errorf("invalid OAuth provider '%s', valid options: %v", p, config.ValidOAuthProviders()))
		}
	}

	// Validate GDPR features
	for _, f := range gdprFeatures {
		if !slices.Contains(config.ValidGDPRFeatures(), f) {
			validationErrors = append(validationErrors, fmt.Errorf("invalid GDPR feature '%s', valid options: %v", f, config.ValidGDPRFeatures()))
		}
	}

	// Validate email service
	if emailServiceStr != "" && !slices.Contains(config.ValidEmailServices(), emailServiceStr) {
		validationErrors = append(validationErrors, fmt.Errorf("invalid email service '%s', valid options: %v", emailServiceStr, config.ValidEmailServices()))
	}

	// Email service is required if GDPR features are selected
	if len(gdprFeatures) > 0 && emailServiceStr == "" {
		validationErrors = append(validationErrors, errors.New("--email is required when GDPR features are selected"))
	}

	// Tenancy requires auth or both
	if enableTenancy && serviceType == config.ServiceGateway {
		validationErrors = append(validationErrors, errors.New("--tenancy requires auth or both service type"))
	}

	// K8s requires domain
	if enableK8s && domainStr == "" {
		validationErrors = append(validationErrors, errors.New("--domain is required when --k8s is enabled"))
	}

	// Domain requires k8s
	if domainStr != "" && !enableK8s {
		validationErrors = append(validationErrors, errors.New("--domain requires --k8s to be enabled"))
	}

	// Validate frontend options
	if len(frontends) > 0 {
		hasWeb := false
		for _, f := range frontends {
			if !slices.Contains(config.ValidFrontendTypes(), f) {
				validationErrors = append(validationErrors, fmt.Errorf("invalid frontend type '%s', valid options: %v", f, config.ValidFrontendTypes()))
			}
			if f == string(config.FrontendWeb) {
				hasWeb = true
			}
		}

		// Frontend is only available for auth or both services
		if serviceType == config.ServiceGateway {
			validationErrors = append(validationErrors, errors.New("--frontend is only available for auth or both service types"))
		}

		// Web frontend requires web framework
		if hasWeb {
			if webFrameworkStr == "" {
				validationErrors = append(validationErrors, errors.New("--web-framework is required for web frontend"))
			} else if !slices.Contains(config.ValidWebFrameworks(), webFrameworkStr) {
				validationErrors = append(validationErrors, fmt.Errorf("invalid web framework '%s', valid options: %v", webFrameworkStr, config.ValidWebFrameworks()))
			} else if webFrameworkStr != string(config.FrameworkReact) {
				validationErrors = append(validationErrors, fmt.Errorf("web framework %q is not yet implemented, only %q is currently available", webFrameworkStr, config.FrameworkReact))
			}
		}

		// UI library is required
		if uiLibraryStr == "" {
			validationErrors = append(validationErrors, errors.New("--ui-lib is required when --frontend is specified"))
		} else if !slices.Contains(config.ValidUILibraries(), uiLibraryStr) {
			validationErrors = append(validationErrors, fmt.Errorf("invalid UI library '%s', valid options: %v", uiLibraryStr, config.ValidUILibraries()))
		}

		// State management is required
		if stateMgmtStr == "" {
			validationErrors = append(validationErrors, errors.New("--state-mgmt is required when --frontend is specified"))
		} else if !slices.Contains(config.ValidStateManagements(), stateMgmtStr) {
			validationErrors = append(validationErrors, fmt.Errorf("invalid state management '%s', valid options: %v", stateMgmtStr, config.ValidStateManagements()))
		}
	} else {
		// If frontend is not specified, warn if related flags are provided
		if webFrameworkStr != "" {
			validationErrors = append(validationErrors, errors.New("--web-framework requires --frontend to be specified"))
		}
		if uiLibraryStr != "" {
			validationErrors = append(validationErrors, errors.New("--ui-lib requires --frontend to be specified"))
		}
		if stateMgmtStr != "" {
			validationErrors = append(validationErrors, errors.New("--state-mgmt requires --frontend to be specified"))
		}
		if enablePostHog {
			validationErrors = append(validationErrors, errors.New("--posthog requires --frontend to be specified"))
		}
		if enableSentry {
			validationErrors = append(validationErrors, errors.New("--sentry requires --frontend to be specified"))
		}
		if cmd.Flags().Changed("e2e-framework") {
			validationErrors = append(validationErrors, errors.New("--e2e-framework requires --frontend to be specified"))
		}
	}

	if len(validationErrors) > 0 {
		joinedErr := errors.Join(validationErrors...)
		usage := cmd.UsageString()
		return NewCliError(joinedErr, usage)
	}

	// Build configuration
	cfg := &config.ProjectConfig{
		ModuleName:    moduleName,
		ProjectName:   config.ExtractProjectName(moduleName),
		ServiceType:   serviceType,
		ConfigSource:  config.ConfigSource(configStr),
		Observability: observability,
		EnableMFA:     enableMFA,
		EnableRBAC:    enableRBAC,
		AuthCache:     authCache,
		EnableTenancy: enableTenancy,
		EnableK8s:     enableK8s,
		Domain:        domainStr,
	}

	if databaseStr != "" {
		db := config.DatabaseType(databaseStr)
		cfg.Database = &db
	}

	if cacheStr != "" {
		cache := config.CacheType(cacheStr)
		cfg.Cache = &cache
	}

	if rateLimiterStr != "" {
		rl := config.RateLimiterType(rateLimiterStr)
		cfg.RateLimiter = &rl
	}

	// OAuth providers
	for _, p := range oauthProviders {
		cfg.OAuthProviders = append(cfg.OAuthProviders, config.OAuthProvider(p))
	}

	// GDPR features
	for _, f := range gdprFeatures {
		cfg.GDPRFeatures = append(cfg.GDPRFeatures, config.GDPRFeature(f))
	}

	// Email service
	if emailServiceStr != "" {
		email := config.EmailService(emailServiceStr)
		cfg.EmailService = &email
	}

	// Frontend
	if len(frontends) > 0 {
		for _, f := range frontends {
			cfg.Frontends = append(cfg.Frontends, config.FrontendType(f))
		}

		if webFrameworkStr != "" {
			framework := config.WebFramework(webFrameworkStr)
			cfg.WebFramework = &framework
		}

		if uiLibraryStr != "" {
			uiLib := config.UILibrary(uiLibraryStr)
			cfg.UILibrary = &uiLib
		}

		if stateMgmtStr != "" {
			stateMgmt := config.StateManagement(stateMgmtStr)
			cfg.StateManagement = &stateMgmt
		}

		// Analytics/monitoring options
		cfg.EnablePostHog = enablePostHog
		cfg.EnableSentry = enableSentry
	}

	// Test infrastructure
	if testInfraStr != "" {
		if !slices.Contains(config.ValidTestInfraTypes(), testInfraStr) {
			return fmt.Errorf("invalid test infrastructure: %s (valid: %v)", testInfraStr, config.ValidTestInfraTypes())
		}
		testInfra := config.TestInfraType(testInfraStr)
		cfg.TestInfra = &testInfra
	}

	// E2E framework (only valid when frontend is set)
	if e2eFrameworkStr != "" && len(frontends) > 0 {
		if !slices.Contains(config.ValidE2EFrameworkTypes(), e2eFrameworkStr) {
			return fmt.Errorf("invalid E2E framework: %s (valid: %v)", e2eFrameworkStr, config.ValidE2EFrameworkTypes())
		}
		e2eFramework := config.E2EFrameworkType(e2eFrameworkStr)
		cfg.E2EFramework = &e2eFramework
	}

	// Generate the project
	fmt.Println("Generating project...")

	gen := generator.NewGenerator(cfg)
	if err := gen.Generate(context.Background()); err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	fmt.Printf("\n✓ Project generated successfully at: %s\n", cfg.ProjectName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", cfg.ProjectName)
	fmt.Println("  go build ./...")

	return nil
}

// validateModuleName validates the Go module name format
func validateModuleName(name string) error {
	if name == "" {
		return errors.New("module name is required")
	}

	pattern := `^[a-zA-Z0-9][a-zA-Z0-9._-]*(/[a-zA-Z0-9][a-zA-Z0-9._-]*)*$`
	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		return fmt.Errorf("failed to validate module name: %w", err)
	}
	if !matched {
		return fmt.Errorf("invalid module name format: %s (expected format: domain.com/user/project)", name)
	}

	return nil
}

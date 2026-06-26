package cmd

import (
	"context"
	"fmt"
	"os"

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
		outputDir         string
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
			return runNonInteractive(cmd, args, serviceTypeStr, databaseStr, cacheStr, configStr, rateLimiterStr, observabilityBool, oauthProviders, enableMFA, enableRBAC, gdprFeatures, emailServiceStr, authCache, frontends, webFrameworkStr, uiLibraryStr, stateMgmtStr, enablePostHog, enableSentry, testInfraStr, e2eFrameworkStr, enableTenancy, enableK8s, domainStr, outputDir)
		},
	}

	cmd.Flags().StringVarP(&serviceTypeStr, "service", "s", "", "Service type (gateway, auth, both)")
	cmd.Flags().StringVarP(&databaseStr, "database", "d", "", "Database type (postgres)")
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
	cmd.Flags().StringVar(&webFrameworkStr, "web-framework", "", "Web framework for web frontend (only react is currently implemented)")
	cmd.Flags().StringVar(&uiLibraryStr, "ui-lib", "", "UI library (shadcn, baseui)")
	cmd.Flags().StringVar(&stateMgmtStr, "state-mgmt", "", "State management (tanstack, redux)")
	cmd.Flags().BoolVar(&enablePostHog, "posthog", false, "Enable PostHog analytics (requires --frontend)")
	cmd.Flags().BoolVar(&enableSentry, "sentry", false, "Enable Sentry error tracking (requires --frontend)")
	cmd.Flags().StringVar(&testInfraStr, "test-infra", "docker", "Test infrastructure (docker, testcontainers)")
	cmd.Flags().StringVar(&e2eFrameworkStr, "e2e-framework", "cypress", "E2E test framework (cypress, playwright) - only with --frontend")
	cmd.Flags().BoolVar(&enableTenancy, "tenancy", false, "Enable auth-first multi-tenancy")
	cmd.Flags().BoolVar(&enableK8s, "k8s", false, "Generate k3s-based local dev environment")
	cmd.Flags().StringVar(&domainStr, "domain", "", "Local dev domain for k8s (e.g., myapp.dev)")
	cmd.Flags().StringVarP(&outputDir, "output", "O", "", "Output directory (default: project name)")

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
		// Only React is implemented today, so don't tab-complete the unimplemented options.
		return []string{string(config.FrameworkReact)}, cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("ui-lib", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidUILibraries(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("state-mgmt", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidStateManagements(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("oauth", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidOAuthProviders(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("gdpr", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidGDPRFeatures(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("email", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidEmailServices(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("test-infra", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidTestInfraTypes(), cobra.ShellCompDirectiveNoFileComp
	})
	cmd.RegisterFlagCompletionFunc("e2e-framework", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidE2EFrameworkTypes(), cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

// hasAnyFlag checks if any relevant flags were set
func hasAnyFlag(cmd *cobra.Command) bool {
	flags := []string{"service", "database", "cache", "config", "rate-limiter", "frontend", "web-framework", "ui-lib", "state-mgmt", "oauth", "mfa", "rbac", "gdpr", "email", "auth-cache", "observability", "posthog", "sentry", "test-infra", "e2e-framework", "tenancy", "k8s", "domain", "output"}
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
func runNonInteractive(cmd *cobra.Command, args []string, serviceTypeStr, databaseStr, cacheStr, configStr, rateLimiterStr string, observability bool, oauthProviders []string, enableMFA, enableRBAC bool, gdprFeatures []string, emailServiceStr string, authCache bool, frontends []string, webFrameworkStr, uiLibraryStr, stateMgmtStr string, enablePostHog, enableSentry bool, testInfraStr, e2eFrameworkStr string, enableTenancy, enableK8s bool, domainStr, outputDir string) error {
	var moduleName string
	if len(args) > 0 {
		moduleName = args[0]
	}

	// Parse flags into a configuration. All validation — required-field presence,
	// enum validity, and cross-field rules — is delegated to cfg.Validate, the
	// single source of truth shared with the interactive wizard.
	cfg := &config.ProjectConfig{
		ModuleName:    moduleName,
		ProjectName:   config.ExtractProjectName(moduleName),
		OutputDir:     outputDir,
		ServiceType:   config.ServiceType(serviceTypeStr),
		ConfigSource:  config.ConfigSource(configStr),
		Observability: observability,
		EnableMFA:     enableMFA,
		EnableRBAC:    enableRBAC,
		AuthCache:     authCache,
		EnableTenancy: enableTenancy,
		EnableK8s:     enableK8s,
		Domain:        domainStr,
		EnablePostHog: enablePostHog,
		EnableSentry:  enableSentry,
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
	for _, p := range oauthProviders {
		cfg.OAuthProviders = append(cfg.OAuthProviders, config.OAuthProvider(p))
	}
	for _, f := range gdprFeatures {
		cfg.GDPRFeatures = append(cfg.GDPRFeatures, config.GDPRFeature(f))
	}
	if emailServiceStr != "" {
		email := config.EmailService(emailServiceStr)
		cfg.EmailService = &email
	}
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
	if testInfraStr != "" {
		testInfra := config.TestInfraType(testInfraStr)
		cfg.TestInfra = &testInfra
	}
	// E2E framework defaults to a non-empty value, so only attach it when a frontend
	// is selected or the user explicitly set it — letting Validate flag an orphan
	// --e2e-framework used without --frontend.
	if e2eFrameworkStr != "" && (len(frontends) > 0 || cmd.Flags().Changed("e2e-framework")) {
		e2eFramework := config.E2EFrameworkType(e2eFrameworkStr)
		cfg.E2EFramework = &e2eFramework
	}

	if err := cfg.Validate(); err != nil {
		return NewCliError(err, cmd.UsageString())
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

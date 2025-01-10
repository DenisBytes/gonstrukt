package tui

import (
	"fmt"
	"strings"

	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/DenisBytes/gonstrukt/internal/tui/steps"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WizardState represents the current state of the wizard
type WizardState int

const (
	StateSelectingServiceType WizardState = iota
	StateEnteringProjectName
	StateSelectingDatabase
	StateSelectingCache
	StateSelectingConfigSource
	StateSelectingRateLimiter
	StateSelectingOAuthProviders
	StateSelectingMFA
	StateSelectingRBAC
	StateSelectingGDPRFeatures
	StateSelectingEmailService
	StateSelectingGatewayFeatures // Auth-only: optional gateway features
	StateSelectingAuthCache
	StateSelectingFrontend        // Frontend type (web, mobile, none)
	StateSelectingWebFramework    // Web framework (react, next, tanstack)
	StateSelectingUILibrary       // UI library (shadcn, baseui)
	StateSelectingStateManagement // State management (tanstack, redux)
	StateSelectingAnalytics       // Analytics (posthog, sentry)
	StateSelectingObservability
	StateShowingSummary
	StateGenerating
	StateComplete
	StateError
)

// stepInfo holds information about a wizard step
type stepInfo struct {
	name        string
	title       string
	description string
	required    func(*config.ProjectConfig) bool
}

// Wizard is the main TUI model for the project generation wizard
type Wizard struct {
	state   WizardState
	config  *config.ProjectConfig
	styles  components.Styles
	keys    KeyMap
	width   int
	height  int
	err     error

	// Steps
	serviceTypeStep      *steps.ServiceTypeStep
	projectNameStep      *steps.ProjectNameStep
	databaseStep         *steps.DatabaseStep
	cacheStep            *steps.CacheStep
	configSourceStep     *steps.ConfigSourceStep
	rateLimiterStep      *steps.RateLimiterStep
	oauthStep            *steps.OAuthStep
	mfaStep              *steps.MFAStep
	rbacStep             *steps.RBACStep
	gdprStep             *steps.GDPRStep
	emailServiceStep     *steps.EmailServiceStep
	gatewayFeaturesStep  *steps.GatewayFeaturesStep
	authCacheStep        *steps.AuthCacheStep
	frontendStep         *steps.FrontendStep
	webFrameworkStep     *steps.WebFrameworkStep
	uiLibraryStep        *steps.UILibraryStep
	stateManagementStep  *steps.StateManagementStep
	analyticsStep        *steps.AnalyticsStep
	observabilityStep    *steps.ObservabilityStep
	summaryStep          *steps.SummaryStep

	// Progress
	progress components.Progress

	// Current step info for display
	stepInfos []stepInfo
}

// NewWizard creates a new wizard instance
func NewWizard() *Wizard {
	cfg := &config.ProjectConfig{}

	w := &Wizard{
		state:  StateSelectingServiceType,
		config: cfg,
		styles: components.DefaultStyles(),
		keys:   DefaultKeyMap(),

		serviceTypeStep:      steps.NewServiceTypeStep(),
		projectNameStep:      steps.NewProjectNameStep(),
		databaseStep:         steps.NewDatabaseStep(),
		cacheStep:            steps.NewCacheStep(),
		configSourceStep:     steps.NewConfigSourceStep(),
		rateLimiterStep:      steps.NewRateLimiterStep(),
		oauthStep:            steps.NewOAuthStep(),
		mfaStep:              steps.NewMFAStep(),
		rbacStep:             steps.NewRBACStep(),
		gdprStep:             steps.NewGDPRStep(),
		emailServiceStep:     steps.NewEmailServiceStep(),
		gatewayFeaturesStep:  steps.NewGatewayFeaturesStep(),
		authCacheStep:        steps.NewAuthCacheStep(),
		frontendStep:         steps.NewFrontendStep(),
		webFrameworkStep:     steps.NewWebFrameworkStep(),
		uiLibraryStep:        steps.NewUILibraryStep(),
		stateManagementStep:  steps.NewStateManagementStep(),
		analyticsStep:        steps.NewAnalyticsStep(),
		observabilityStep:    steps.NewObservabilityStep(),
		summaryStep:          steps.NewSummaryStep(cfg),

		progress: components.NewProgress([]string{
			"Creating project structure",
			"Generating configuration",
			"Generating service code",
			"Generating database layer",
			"Generating middleware",
			"Running go mod tidy",
			"Formatting code",
		}),
	}

	w.stepInfos = []stepInfo{
		{
			name:        "service_type",
			title:       "Service Type",
			description: "Choose the type of service to generate",
			required:    func(*config.ProjectConfig) bool { return true },
		},
		{
			name:        "project_name",
			title:       "Module Name",
			description: "Enter your Go module path",
			required:    func(*config.ProjectConfig) bool { return true },
		},
		{
			name:        "database",
			title:       "Database",
			description: "Select the database backend",
			required: func(cfg *config.ProjectConfig) bool {
				return cfg.ServiceType == config.ServiceAuth || cfg.ServiceType == config.ServiceBoth
			},
		},
		{
			name:        "cache",
			title:       "Cache",
			description: "Select the caching backend",
			required: func(cfg *config.ProjectConfig) bool {
				return cfg.ServiceType == config.ServiceGateway || cfg.ServiceType == config.ServiceBoth
			},
		},
		{
			name:        "config_source",
			title:       "Configuration",
			description: "Choose where configuration will be loaded from",
			required:    func(*config.ProjectConfig) bool { return true },
		},
		{
			name:        "rate_limiter",
			title:       "Rate Limiter",
			description: "Choose the rate limiting algorithm",
			required: func(cfg *config.ProjectConfig) bool {
				return cfg.ServiceType == config.ServiceGateway || cfg.ServiceType == config.ServiceBoth
			},
		},
		{
			name:        "oauth_providers",
			title:       "OAuth Providers",
			description: "Select OAuth providers for social login",
			required: func(cfg *config.ProjectConfig) bool {
				return cfg.ServiceType == config.ServiceAuth || cfg.ServiceType == config.ServiceBoth
			},
		},
		{
			name:        "mfa",
			title:       "Multi-Factor Auth",
			description: "Enable TOTP-based MFA",
			required: func(cfg *config.ProjectConfig) bool {
				return cfg.ServiceType == config.ServiceAuth || cfg.ServiceType == config.ServiceBoth
			},
		},
		{
			name:        "rbac",
			title:       "Role-Based Access Control",
			description: "Enable Casbin RBAC",
			required: func(cfg *config.ProjectConfig) bool {
				return cfg.ServiceType == config.ServiceAuth || cfg.ServiceType == config.ServiceBoth
			},
		},
		{
			name:        "gdpr_features",
			title:       "GDPR Features",
			description: "Select GDPR compliance features",
			required: func(cfg *config.ProjectConfig) bool {
				return cfg.ServiceType == config.ServiceAuth || cfg.ServiceType == config.ServiceBoth
			},
		},
		{
			name:        "email_service",
			title:       "Email Service",
			description: "Select email service provider",
			required: func(cfg *config.ProjectConfig) bool {
				// Only required if GDPR features are selected
				return (cfg.ServiceType == config.ServiceAuth || cfg.ServiceType == config.ServiceBoth) && len(cfg.GDPRFeatures) > 0
			},
		},
		{
			name:        "gateway_features",
			title:       "Gateway Features",
			description: "Enable caching and rate limiting",
			required: func(cfg *config.ProjectConfig) bool {
				// Only for auth-only service
				return cfg.ServiceType == config.ServiceAuth
			},
		},
		{
			name:        "auth_cache",
			title:       "Auth Caching",
			description: "Enable auth response caching",
			required: func(cfg *config.ProjectConfig) bool {
				return cfg.ServiceType == config.ServiceGateway || cfg.ServiceType == config.ServiceBoth
			},
		},
		{
			name:        "frontend",
			title:       "Frontend",
			description: "Add a frontend to your auth service",
			required: func(cfg *config.ProjectConfig) bool {
				return cfg.ServiceType == config.ServiceAuth || cfg.ServiceType == config.ServiceBoth
			},
		},
		{
			name:        "web_framework",
			title:       "Web Framework",
			description: "Choose the web framework",
			required: func(cfg *config.ProjectConfig) bool {
				for _, f := range cfg.Frontends {
					if f == config.FrontendWeb {
						return true
					}
				}
				return false
			},
		},
		{
			name:        "ui_library",
			title:       "UI Library",
			description: "Choose the UI component library",
			required: func(cfg *config.ProjectConfig) bool {
				return len(cfg.Frontends) > 0
			},
		},
		{
			name:        "state_management",
			title:       "State Management",
			description: "Choose the state management approach",
			required: func(cfg *config.ProjectConfig) bool {
				return len(cfg.Frontends) > 0
			},
		},
		{
			name:        "analytics",
			title:       "Analytics & Monitoring",
			description: "Add PostHog and/or Sentry to your frontend",
			required: func(cfg *config.ProjectConfig) bool {
				return len(cfg.Frontends) > 0
			},
		},
		{
			name:        "observability",
			title:       "Observability",
			description: "Enable OpenTelemetry tracing and metrics",
			required:    func(*config.ProjectConfig) bool { return true },
		},
		{
			name:        "summary",
			title:       "Summary",
			description: "Review and confirm your selections",
			required:    func(*config.ProjectConfig) bool { return true },
		},
	}

	return w
}

// Init initializes the wizard
func (w *Wizard) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (w *Wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		return w, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, w.keys.Quit):
			return w, tea.Quit
		}

	case steps.StepCompleteMsg:
		return w.handleStepComplete(msg)

	case steps.StepBackMsg:
		return w.handleStepBack()

	case components.ProgressMsg:
		return w.handleProgress(msg)

	case GenerationCompleteMsg:
		w.state = StateComplete
		w.progress.Complete()
		return w, nil

	case GenerationErrorMsg:
		w.state = StateError
		w.err = msg.Error
		return w, nil
	}

	// Delegate to current step
	return w.updateCurrentStep(msg)
}

// handleStepComplete processes step completion
func (w *Wizard) handleStepComplete(msg steps.StepCompleteMsg) (tea.Model, tea.Cmd) {
	switch msg.StepName {
	case "service_type":
		w.config.ServiceType = msg.Value.(config.ServiceType)
		w.state = StateEnteringProjectName
		return w, w.projectNameStep.Init()

	case "project_name":
		w.config.ModuleName = msg.Value.(string)
		w.config.ProjectName = config.ExtractProjectName(w.config.ModuleName)
		return w.nextStepFromProjectName()

	case "database":
		db := msg.Value.(config.DatabaseType)
		w.config.Database = &db
		return w.nextStepFromDatabase()

	case "cache":
		cache := msg.Value.(config.CacheType)
		w.config.Cache = &cache
		// For Auth with gateway features, config was already selected, go to rate limiter
		if w.config.ServiceType == config.ServiceAuth && w.config.ConfigSource != "" {
			w.state = StateSelectingRateLimiter
		} else {
			w.state = StateSelectingConfigSource
		}
		return w, nil

	case "config_source":
		w.config.ConfigSource = msg.Value.(config.ConfigSource)
		return w.nextStepFromConfigSource()

	case "rate_limiter":
		rl := msg.Value.(config.RateLimiterType)
		w.config.RateLimiter = &rl
		return w.nextStepFromRateLimiter()

	case "oauth_providers":
		w.config.OAuthProviders = msg.Value.([]config.OAuthProvider)
		w.state = StateSelectingMFA
		return w, nil

	case "mfa":
		w.config.EnableMFA = msg.Value.(bool)
		w.state = StateSelectingRBAC
		return w, nil

	case "rbac":
		w.config.EnableRBAC = msg.Value.(bool)
		w.state = StateSelectingGDPRFeatures
		return w, nil

	case "gdpr_features":
		w.config.GDPRFeatures = msg.Value.([]config.GDPRFeature)
		return w.nextStepFromGDPR()

	case "email_service":
		email := msg.Value.(config.EmailService)
		w.config.EmailService = &email
		return w.nextStepFromEmailService()

	case "gateway_features":
		enableGateway := msg.Value.(bool)
		if enableGateway {
			// User wants gateway features: go to cache selection
			w.state = StateSelectingCache
		} else {
			// No gateway features: go to frontend selection
			w.state = StateSelectingFrontend
		}
		return w, nil

	case "auth_cache":
		w.config.AuthCache = msg.Value.(bool)
		return w.nextStepFromAuthCache()

	case "frontend":
		w.config.Frontends = msg.Value.([]config.FrontendType)
		return w.nextStepFromFrontend()

	case "web_framework":
		framework := msg.Value.(config.WebFramework)
		w.config.WebFramework = &framework
		w.state = StateSelectingUILibrary
		return w, nil

	case "ui_library":
		uiLib := msg.Value.(config.UILibrary)
		w.config.UILibrary = &uiLib
		w.state = StateSelectingStateManagement
		return w, nil

	case "state_management":
		stateMgmt := msg.Value.(config.StateManagement)
		w.config.StateManagement = &stateMgmt
		w.state = StateSelectingAnalytics
		return w, nil

	case "analytics":
		selection := msg.Value.(steps.AnalyticsSelection)
		w.config.EnablePostHog = selection.PostHog
		w.config.EnableSentry = selection.Sentry
		w.state = StateSelectingObservability
		return w, nil

	case "observability":
		w.config.Observability = msg.Value.(bool)
		w.summaryStep.UpdateConfig(w.config)
		w.state = StateShowingSummary
		return w, nil

	case "summary":
		w.state = StateGenerating
		return w, w.startGeneration()
	}

	return w, nil
}

// nextStepFromProjectName determines the next step after project name
func (w *Wizard) nextStepFromProjectName() (tea.Model, tea.Cmd) {
	switch w.config.ServiceType {
	case config.ServiceAuth, config.ServiceBoth:
		w.state = StateSelectingDatabase
	case config.ServiceGateway:
		w.state = StateSelectingCache
	}
	return w, nil
}

// nextStepFromDatabase determines the next step after database selection
func (w *Wizard) nextStepFromDatabase() (tea.Model, tea.Cmd) {
	if w.config.ServiceType == config.ServiceBoth {
		w.state = StateSelectingCache
	} else {
		w.state = StateSelectingConfigSource
	}
	return w, nil
}

// nextStepFromConfigSource determines the next step after config source
func (w *Wizard) nextStepFromConfigSource() (tea.Model, tea.Cmd) {
	if w.config.ServiceType == config.ServiceGateway || w.config.ServiceType == config.ServiceBoth {
		w.state = StateSelectingRateLimiter
	} else {
		// Auth-only: go to OAuth selection
		w.state = StateSelectingOAuthProviders
	}
	return w, nil
}

// nextStepFromRateLimiter determines the next step after rate limiter selection
func (w *Wizard) nextStepFromRateLimiter() (tea.Model, tea.Cmd) {
	if w.config.ServiceType == config.ServiceBoth {
		// Both services: go to auth options
		w.state = StateSelectingOAuthProviders
	} else if w.config.ServiceType == config.ServiceAuth {
		// Auth with gateway features: go to frontend selection
		w.state = StateSelectingFrontend
	} else {
		// Gateway-only: go to auth cache
		w.state = StateSelectingAuthCache
	}
	return w, nil
}

// nextStepFromGDPR determines the next step after GDPR features
func (w *Wizard) nextStepFromGDPR() (tea.Model, tea.Cmd) {
	if len(w.config.GDPRFeatures) > 0 {
		// GDPR features selected, need email service
		w.state = StateSelectingEmailService
	} else {
		return w.nextStepFromEmailService()
	}
	return w, nil
}

// nextStepFromEmailService determines the next step after email service
func (w *Wizard) nextStepFromEmailService() (tea.Model, tea.Cmd) {
	if w.config.ServiceType == config.ServiceBoth {
		// Both services: go to auth cache
		w.state = StateSelectingAuthCache
	} else {
		// Auth-only: offer optional gateway features
		w.state = StateSelectingGatewayFeatures
	}
	return w, nil
}

// nextStepFromAuthCache determines the next step after auth cache
func (w *Wizard) nextStepFromAuthCache() (tea.Model, tea.Cmd) {
	// Both gateway and both services can have frontend (if auth is involved)
	if w.config.ServiceType == config.ServiceBoth {
		w.state = StateSelectingFrontend
	} else {
		// Gateway-only: skip frontend, go to observability
		w.state = StateSelectingObservability
	}
	return w, nil
}

// nextStepFromFrontend determines the next step after frontend selection
func (w *Wizard) nextStepFromFrontend() (tea.Model, tea.Cmd) {
	if len(w.config.Frontends) == 0 {
		// No frontend selected: go to observability
		w.state = StateSelectingObservability
	} else {
		// Check if web frontend is selected
		hasWeb := false
		for _, f := range w.config.Frontends {
			if f == config.FrontendWeb {
				hasWeb = true
				break
			}
		}
		if hasWeb {
			// Web frontend: need web framework selection
			w.state = StateSelectingWebFramework
		} else {
			// Mobile only: skip web framework, go to UI library
			w.state = StateSelectingUILibrary
		}
	}
	return w, nil
}

// handleStepBack goes back to the previous step
func (w *Wizard) handleStepBack() (tea.Model, tea.Cmd) {
	switch w.state {
	case StateEnteringProjectName:
		w.serviceTypeStep.Reset()
		w.state = StateSelectingServiceType

	case StateSelectingDatabase:
		w.projectNameStep.Reset()
		w.state = StateEnteringProjectName

	case StateSelectingCache:
		if w.config.ServiceType == config.ServiceBoth {
			w.databaseStep.Reset()
			w.state = StateSelectingDatabase
		} else if w.config.ServiceType == config.ServiceAuth {
			// Auth with gateway features: go back to gateway features toggle
			w.gatewayFeaturesStep.Reset()
			w.state = StateSelectingGatewayFeatures
		} else {
			// Gateway: go back to project name
			w.projectNameStep.Reset()
			w.state = StateEnteringProjectName
		}

	case StateSelectingConfigSource:
		if w.config.ServiceType == config.ServiceGateway || w.config.ServiceType == config.ServiceBoth {
			w.cacheStep.Reset()
			w.state = StateSelectingCache
		} else {
			w.databaseStep.Reset()
			w.state = StateSelectingDatabase
		}

	case StateSelectingRateLimiter:
		// For Auth with gateway features, go back to cache
		// For Gateway/Both, go back to config source
		if w.config.ServiceType == config.ServiceAuth {
			w.cacheStep.Reset()
			w.state = StateSelectingCache
		} else {
			w.configSourceStep.Reset()
			w.state = StateSelectingConfigSource
		}

	case StateSelectingOAuthProviders:
		if w.config.ServiceType == config.ServiceBoth {
			w.rateLimiterStep.Reset()
			w.state = StateSelectingRateLimiter
		} else {
			w.configSourceStep.Reset()
			w.state = StateSelectingConfigSource
		}

	case StateSelectingMFA:
		w.oauthStep.Reset()
		w.state = StateSelectingOAuthProviders

	case StateSelectingRBAC:
		w.mfaStep.Reset()
		w.state = StateSelectingMFA

	case StateSelectingGDPRFeatures:
		w.rbacStep.Reset()
		w.state = StateSelectingRBAC

	case StateSelectingEmailService:
		w.gdprStep.Reset()
		w.state = StateSelectingGDPRFeatures

	case StateSelectingGatewayFeatures:
		// Go back to email service or GDPR
		if len(w.config.GDPRFeatures) > 0 {
			w.emailServiceStep.Reset()
			w.state = StateSelectingEmailService
		} else {
			w.gdprStep.Reset()
			w.state = StateSelectingGDPRFeatures
		}

	case StateSelectingAuthCache:
		if w.config.ServiceType == config.ServiceBoth {
			if len(w.config.GDPRFeatures) > 0 {
				w.emailServiceStep.Reset()
				w.state = StateSelectingEmailService
			} else {
				w.gdprStep.Reset()
				w.state = StateSelectingGDPRFeatures
			}
		} else {
			w.rateLimiterStep.Reset()
			w.state = StateSelectingRateLimiter
		}

	case StateSelectingFrontend:
		// Go back based on service type and gateway features
		if w.config.ServiceType == config.ServiceBoth {
			w.authCacheStep.Reset()
			w.state = StateSelectingAuthCache
		} else if w.config.ServiceType == config.ServiceAuth {
			if w.config.Cache != nil || w.config.RateLimiter != nil {
				// Gateway features enabled: go back to rate limiter
				w.rateLimiterStep.Reset()
				w.state = StateSelectingRateLimiter
			} else {
				// No gateway features: go back to gateway features toggle
				w.gatewayFeaturesStep.Reset()
				w.state = StateSelectingGatewayFeatures
			}
		}

	case StateSelectingWebFramework:
		w.frontendStep.Reset()
		w.state = StateSelectingFrontend

	case StateSelectingUILibrary:
		// Check if web frontend is selected
		hasWeb := false
		for _, f := range w.config.Frontends {
			if f == config.FrontendWeb {
				hasWeb = true
				break
			}
		}
		if hasWeb {
			w.webFrameworkStep.Reset()
			w.state = StateSelectingWebFramework
		} else {
			// Mobile only: go back to frontend
			w.frontendStep.Reset()
			w.state = StateSelectingFrontend
		}

	case StateSelectingStateManagement:
		w.uiLibraryStep.Reset()
		w.state = StateSelectingUILibrary

	case StateSelectingAnalytics:
		w.stateManagementStep.Reset()
		w.state = StateSelectingStateManagement

	case StateSelectingObservability:
		// Go back based on frontend selection
		if len(w.config.Frontends) > 0 {
			w.analyticsStep.Reset()
			w.state = StateSelectingAnalytics
		} else if w.config.ServiceType == config.ServiceGateway {
			w.authCacheStep.Reset()
			w.state = StateSelectingAuthCache
		} else if w.config.ServiceType == config.ServiceBoth {
			w.frontendStep.Reset()
			w.state = StateSelectingFrontend
		} else if w.config.ServiceType == config.ServiceAuth {
			w.frontendStep.Reset()
			w.state = StateSelectingFrontend
		}

	case StateShowingSummary:
		w.observabilityStep.Reset()
		w.state = StateSelectingObservability
	}

	return w, nil
}

// handleProgress handles progress updates during generation
func (w *Wizard) handleProgress(msg components.ProgressMsg) (tea.Model, tea.Cmd) {
	switch msg.State {
	case components.ProgressInProgress:
		w.progress.StartStep(msg.Step, msg.Message)
	case components.ProgressComplete:
		w.progress.CompleteStep(msg.Step)
	case components.ProgressError:
		w.progress.FailStep(msg.Step, msg.Error)
	}

	var cmd tea.Cmd
	w.progress, cmd = w.progress.Update(msg)
	return w, cmd
}

// updateCurrentStep delegates update to the current step
func (w *Wizard) updateCurrentStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch w.state {
	case StateSelectingServiceType:
		model, c := w.serviceTypeStep.Update(msg)
		w.serviceTypeStep = model.(*steps.ServiceTypeStep)
		cmd = c

	case StateEnteringProjectName:
		model, c := w.projectNameStep.Update(msg)
		w.projectNameStep = model.(*steps.ProjectNameStep)
		cmd = c

	case StateSelectingDatabase:
		model, c := w.databaseStep.Update(msg)
		w.databaseStep = model.(*steps.DatabaseStep)
		cmd = c

	case StateSelectingCache:
		model, c := w.cacheStep.Update(msg)
		w.cacheStep = model.(*steps.CacheStep)
		cmd = c

	case StateSelectingConfigSource:
		model, c := w.configSourceStep.Update(msg)
		w.configSourceStep = model.(*steps.ConfigSourceStep)
		cmd = c

	case StateSelectingRateLimiter:
		model, c := w.rateLimiterStep.Update(msg)
		w.rateLimiterStep = model.(*steps.RateLimiterStep)
		cmd = c

	case StateSelectingOAuthProviders:
		model, c := w.oauthStep.Update(msg)
		w.oauthStep = model.(*steps.OAuthStep)
		cmd = c

	case StateSelectingMFA:
		model, c := w.mfaStep.Update(msg)
		w.mfaStep = model.(*steps.MFAStep)
		cmd = c

	case StateSelectingRBAC:
		model, c := w.rbacStep.Update(msg)
		w.rbacStep = model.(*steps.RBACStep)
		cmd = c

	case StateSelectingGDPRFeatures:
		model, c := w.gdprStep.Update(msg)
		w.gdprStep = model.(*steps.GDPRStep)
		cmd = c

	case StateSelectingEmailService:
		model, c := w.emailServiceStep.Update(msg)
		w.emailServiceStep = model.(*steps.EmailServiceStep)
		cmd = c

	case StateSelectingGatewayFeatures:
		model, c := w.gatewayFeaturesStep.Update(msg)
		w.gatewayFeaturesStep = model.(*steps.GatewayFeaturesStep)
		cmd = c

	case StateSelectingAuthCache:
		model, c := w.authCacheStep.Update(msg)
		w.authCacheStep = model.(*steps.AuthCacheStep)
		cmd = c

	case StateSelectingFrontend:
		model, c := w.frontendStep.Update(msg)
		w.frontendStep = model.(*steps.FrontendStep)
		cmd = c

	case StateSelectingWebFramework:
		model, c := w.webFrameworkStep.Update(msg)
		w.webFrameworkStep = model.(*steps.WebFrameworkStep)
		cmd = c

	case StateSelectingUILibrary:
		model, c := w.uiLibraryStep.Update(msg)
		w.uiLibraryStep = model.(*steps.UILibraryStep)
		cmd = c

	case StateSelectingStateManagement:
		model, c := w.stateManagementStep.Update(msg)
		w.stateManagementStep = model.(*steps.StateManagementStep)
		cmd = c

	case StateSelectingAnalytics:
		model, c := w.analyticsStep.Update(msg)
		w.analyticsStep = model.(*steps.AnalyticsStep)
		cmd = c

	case StateSelectingObservability:
		model, c := w.observabilityStep.Update(msg)
		w.observabilityStep = model.(*steps.ObservabilityStep)
		cmd = c

	case StateShowingSummary:
		model, c := w.summaryStep.Update(msg)
		w.summaryStep = model.(*steps.SummaryStep)
		cmd = c

	case StateGenerating:
		w.progress, cmd = w.progress.Update(msg)
	}

	return w, cmd
}

// View renders the wizard
func (w *Wizard) View() string {
	var b strings.Builder

	// Header
	b.WriteString(w.renderHeader())
	b.WriteString("\n")

	// Step indicator
	b.WriteString(w.renderStepIndicator())
	b.WriteString("\n\n")

	// Current step content
	b.WriteString(w.renderCurrentStep())
	b.WriteString("\n")

	// Help
	if w.state != StateGenerating && w.state != StateComplete && w.state != StateError {
		b.WriteString(w.renderHelp())
	}

	return w.styles.Container.Render(b.String())
}

// renderHeader renders the wizard header
func (w *Wizard) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.Primary).
		Render("Gonstrukt")

	subtitle := lipgloss.NewStyle().
		Foreground(components.Muted).
		Render(" - Go Service Generator")

	return title + subtitle
}

// renderStepIndicator renders the step progress indicator
func (w *Wizard) renderStepIndicator() string {
	currentStep := w.getCurrentStepIndex()
	totalSteps := w.getTotalSteps()

	if w.state == StateGenerating || w.state == StateComplete || w.state == StateError {
		return ""
	}

	return w.styles.StepCounter.Render(
		fmt.Sprintf("Step %d of %d", currentStep+1, totalSteps),
	)
}

// renderCurrentStep renders the current step's content
func (w *Wizard) renderCurrentStep() string {
	switch w.state {
	case StateSelectingServiceType:
		return w.renderStepWithTitle(w.serviceTypeStep)
	case StateEnteringProjectName:
		return w.renderStepWithTitle(w.projectNameStep)
	case StateSelectingDatabase:
		return w.renderStepWithTitle(w.databaseStep)
	case StateSelectingCache:
		return w.renderStepWithTitle(w.cacheStep)
	case StateSelectingConfigSource:
		return w.renderStepWithTitle(w.configSourceStep)
	case StateSelectingRateLimiter:
		return w.renderStepWithTitle(w.rateLimiterStep)
	case StateSelectingOAuthProviders:
		return w.renderStepWithTitle(w.oauthStep)
	case StateSelectingMFA:
		return w.renderStepWithTitle(w.mfaStep)
	case StateSelectingRBAC:
		return w.renderStepWithTitle(w.rbacStep)
	case StateSelectingGDPRFeatures:
		return w.renderStepWithTitle(w.gdprStep)
	case StateSelectingEmailService:
		return w.renderStepWithTitle(w.emailServiceStep)
	case StateSelectingGatewayFeatures:
		return w.renderStepWithTitle(w.gatewayFeaturesStep)
	case StateSelectingAuthCache:
		return w.renderStepWithTitle(w.authCacheStep)
	case StateSelectingFrontend:
		return w.renderStepWithTitle(w.frontendStep)
	case StateSelectingWebFramework:
		return w.renderStepWithTitle(w.webFrameworkStep)
	case StateSelectingUILibrary:
		return w.renderStepWithTitle(w.uiLibraryStep)
	case StateSelectingStateManagement:
		return w.renderStepWithTitle(w.stateManagementStep)
	case StateSelectingAnalytics:
		return w.renderStepWithTitle(w.analyticsStep)
	case StateSelectingObservability:
		return w.renderStepWithTitle(w.observabilityStep)
	case StateShowingSummary:
		return w.summaryStep.View()
	case StateGenerating:
		return w.progress.View()
	case StateComplete:
		return w.renderComplete()
	case StateError:
		return w.renderError()
	default:
		return ""
	}
}

// renderStepWithTitle renders a step with its title
func (w *Wizard) renderStepWithTitle(step steps.Step) string {
	var b strings.Builder

	b.WriteString(w.styles.Title.Render(step.Title()))
	b.WriteString("\n")
	b.WriteString(w.styles.Description.Render(step.Description()))
	b.WriteString("\n\n")
	b.WriteString(step.View())

	return b.String()
}

// renderComplete renders the completion message
func (w *Wizard) renderComplete() string {
	var b strings.Builder

	b.WriteString(w.progress.View())
	b.WriteString("\n\n")

	successStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.Success)

	b.WriteString(successStyle.Render("Project generated successfully!"))
	b.WriteString("\n\n")

	infoStyle := lipgloss.NewStyle().Foreground(components.Muted)
	b.WriteString(infoStyle.Render("Next steps:"))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render(fmt.Sprintf("  cd %s", w.config.ProjectName)))
	b.WriteString("\n")
	b.WriteString(infoStyle.Render("  go build ./..."))
	b.WriteString("\n")

	return b.String()
}

// renderError renders the error message
func (w *Wizard) renderError() string {
	var b strings.Builder

	errorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.Error)

	b.WriteString(errorStyle.Render("Error generating project"))
	b.WriteString("\n\n")

	if w.err != nil {
		b.WriteString(w.styles.Error.Render(w.err.Error()))
	}

	return b.String()
}

// renderHelp renders the help footer
func (w *Wizard) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(components.Muted)

	keys := []string{"↑/↓ navigate", "enter select", "esc back", "q quit"}
	return helpStyle.Render(strings.Join(keys, " • "))
}

// getCurrentStepIndex returns the current step index
func (w *Wizard) getCurrentStepIndex() int {
	activeSteps := w.getActiveSteps()
	switch w.state {
	case StateSelectingServiceType:
		return 0
	case StateEnteringProjectName:
		return 1
	case StateSelectingDatabase:
		return indexOf(activeSteps, "database")
	case StateSelectingCache:
		return indexOf(activeSteps, "cache")
	case StateSelectingConfigSource:
		return indexOf(activeSteps, "config_source")
	case StateSelectingRateLimiter:
		return indexOf(activeSteps, "rate_limiter")
	case StateSelectingOAuthProviders:
		return indexOf(activeSteps, "oauth_providers")
	case StateSelectingMFA:
		return indexOf(activeSteps, "mfa")
	case StateSelectingRBAC:
		return indexOf(activeSteps, "rbac")
	case StateSelectingGDPRFeatures:
		return indexOf(activeSteps, "gdpr_features")
	case StateSelectingEmailService:
		return indexOf(activeSteps, "email_service")
	case StateSelectingGatewayFeatures:
		return indexOf(activeSteps, "gateway_features")
	case StateSelectingAuthCache:
		return indexOf(activeSteps, "auth_cache")
	case StateSelectingFrontend:
		return indexOf(activeSteps, "frontend")
	case StateSelectingWebFramework:
		return indexOf(activeSteps, "web_framework")
	case StateSelectingUILibrary:
		return indexOf(activeSteps, "ui_library")
	case StateSelectingStateManagement:
		return indexOf(activeSteps, "state_management")
	case StateSelectingAnalytics:
		return indexOf(activeSteps, "analytics")
	case StateSelectingObservability:
		return indexOf(activeSteps, "observability")
	case StateShowingSummary:
		return indexOf(activeSteps, "summary")
	default:
		return 0
	}
}

// getTotalSteps returns the total number of active steps
func (w *Wizard) getTotalSteps() int {
	return len(w.getActiveSteps())
}

// getActiveSteps returns the list of active step names based on config
func (w *Wizard) getActiveSteps() []string {
	var active []string
	for _, info := range w.stepInfos {
		if info.required(w.config) {
			active = append(active, info.name)
		}
	}
	return active
}

// indexOf returns the index of a string in a slice
func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}

// startGeneration initiates the project generation
func (w *Wizard) startGeneration() tea.Cmd {
	return func() tea.Msg {
		return StartGenerationMsg{Config: w.config}
	}
}

// Config returns the current configuration
func (w *Wizard) Config() *config.ProjectConfig {
	return w.config
}

// StartGenerationMsg signals that generation should begin
type StartGenerationMsg struct {
	Config *config.ProjectConfig
}

// GenerationCompleteMsg signals that generation completed successfully
type GenerationCompleteMsg struct{}

// GenerationErrorMsg signals that generation failed
type GenerationErrorMsg struct {
	Error error
}

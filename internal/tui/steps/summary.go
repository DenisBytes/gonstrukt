package steps

import (
	"fmt"
	"strings"

	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SummaryStep displays the final configuration summary
type SummaryStep struct {
	config    *config.ProjectConfig
	confirmed bool
	styles    components.Styles
}

// NewSummaryStep creates a new summary step
func NewSummaryStep(cfg *config.ProjectConfig) *SummaryStep {
	return &SummaryStep{
		config: cfg,
		styles: components.DefaultStyles(),
	}
}

// Init initializes the step
func (s *SummaryStep) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (s *SummaryStep) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
			return s, func() tea.Msg { return StepBackMsg{} }
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "y"))):
			s.confirmed = true
			return s, func() tea.Msg {
				return StepCompleteMsg{
					StepName: "summary",
					Value:    true,
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
			return s, func() tea.Msg { return StepBackMsg{} }
		}
	}

	return s, nil
}

// View renders the step
func (s *SummaryStep) View() string {
	var b strings.Builder

	// Header box
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(components.Primary).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("Configuration Summary"))
	b.WriteString("\n")
	b.WriteString(components.Divider(50))
	b.WriteString("\n\n")

	// Configuration items
	type summaryItem struct{ label, value string }
	var items []summaryItem
	add := func(label, value string) {
		items = append(items, summaryItem{label, value})
	}

	add("Module Name", s.config.ModuleName)
	add("Project Name", config.ExtractProjectName(s.config.ModuleName))
	add("Service Type", formatServiceType(s.config.ServiceType))

	isAuth := s.config.ServiceType == config.ServiceAuth || s.config.ServiceType == config.ServiceBoth
	isGateway := s.config.ServiceType == config.ServiceGateway || s.config.ServiceType == config.ServiceBoth

	if isAuth && s.config.Database != nil {
		add("Database", formatDatabase(*s.config.Database))
	}
	if s.config.Cache != nil {
		add("Cache", formatCache(*s.config.Cache))
	}
	if s.config.RateLimiter != nil {
		add("Rate Limiter", formatRateLimiter(*s.config.RateLimiter))
	}

	add("Config Source", formatConfigSource(s.config.ConfigSource))
	add("Observability", formatObservability(s.config.Observability))

	// Auth feature selections
	if len(s.config.OAuthProviders) > 0 {
		providers := make([]string, len(s.config.OAuthProviders))
		for i, p := range s.config.OAuthProviders {
			providers[i] = string(p)
		}
		add("OAuth Providers", strings.Join(providers, ", "))
	}
	if s.config.EnableMFA {
		add("MFA", "Enabled")
	}
	if s.config.EnableRBAC {
		add("RBAC", "Enabled")
	}
	if len(s.config.GDPRFeatures) > 0 {
		features := make([]string, len(s.config.GDPRFeatures))
		for i, f := range s.config.GDPRFeatures {
			features[i] = string(f)
		}
		add("GDPR Features", strings.Join(features, ", "))
	}
	if s.config.EmailService != nil {
		add("Email Service", string(*s.config.EmailService))
	}
	if isGateway && s.config.AuthCache {
		add("Auth Cache", "Enabled")
	}

	// Frontend
	if len(s.config.Frontends) > 0 {
		fronts := make([]string, len(s.config.Frontends))
		for i, f := range s.config.Frontends {
			fronts[i] = string(f)
		}
		add("Frontend", strings.Join(fronts, ", "))
		if s.config.WebFramework != nil {
			add("Web Framework", string(*s.config.WebFramework))
		}
		if s.config.UILibrary != nil {
			add("UI Library", string(*s.config.UILibrary))
		}
		if s.config.StateManagement != nil {
			add("State Management", string(*s.config.StateManagement))
		}
		if s.config.EnablePostHog {
			add("PostHog Analytics", "Enabled")
		}
		if s.config.EnableSentry {
			add("Sentry Monitoring", "Enabled")
		}
	}

	// Multi-tenancy and k8s
	if s.config.EnableTenancy {
		add("Multi-Tenancy", "Enabled")
	}
	if s.config.EnableK8s {
		add("K8s Dev Env", "Enabled")
		if s.config.Domain != "" {
			add("Domain", s.config.Domain)
		}
	}

	// Testing
	if s.config.TestInfra != nil {
		add("Test Infra", string(*s.config.TestInfra))
	}
	if s.config.E2EFramework != nil {
		add("E2E Framework", string(*s.config.E2EFramework))
	}

	for _, item := range items {
		b.WriteString(s.styles.SummaryLabel.Render(item.label + ":"))
		b.WriteString("  ")
		b.WriteString(s.styles.SummaryValue.Render(item.value))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(components.Divider(50))
	b.WriteString("\n\n")

	// Confirmation prompt
	promptStyle := lipgloss.NewStyle().Foreground(components.Muted)
	b.WriteString(promptStyle.Render("Press "))
	b.WriteString(s.styles.Success.Render("Enter/Y"))
	b.WriteString(promptStyle.Render(" to generate, "))
	b.WriteString(s.styles.Warning.Render("N/Esc"))
	b.WriteString(promptStyle.Render(" to go back"))

	return b.String()
}

// Title returns the step title
func (s *SummaryStep) Title() string {
	return "Review Configuration"
}

// Description returns the step description
func (s *SummaryStep) Description() string {
	return "Review your selections before generating"
}

// IsComplete returns true if the step is complete
func (s *SummaryStep) IsComplete() bool {
	return s.confirmed
}

// Value returns the confirmation status
func (s *SummaryStep) Value() interface{} {
	return s.confirmed
}

// Reset resets the step
func (s *SummaryStep) Reset() {
	s.confirmed = false
}

// UpdateConfig updates the configuration to display
func (s *SummaryStep) UpdateConfig(cfg *config.ProjectConfig) {
	s.config = cfg
}

// Helper functions for formatting values
func formatServiceType(st config.ServiceType) string {
	switch st {
	case config.ServiceGateway:
		return "Gateway"
	case config.ServiceAuth:
		return "Auth Service"
	case config.ServiceBoth:
		return "Both (Monorepo)"
	default:
		return string(st)
	}
}

func formatDatabase(db config.DatabaseType) string {
	switch db {
	case config.DBPostgres:
		return "PostgreSQL"
	case config.DBMySQL:
		return "MySQL"
	default:
		return string(db)
	}
}

func formatCache(c config.CacheType) string {
	switch c {
	case config.CacheRedis:
		return "Redis"
	case config.CacheValkey:
		return "Valkey"
	case config.CacheMemory:
		return "In-Memory"
	default:
		return string(c)
	}
}

func formatConfigSource(cs config.ConfigSource) string {
	switch cs {
	case config.ConfigYAML:
		return "YAML File"
	case config.ConfigEnv:
		return "Environment Variables"
	case config.ConfigVault:
		return "HashiCorp Vault"
	default:
		return string(cs)
	}
}

func formatRateLimiter(rl config.RateLimiterType) string {
	switch rl {
	case config.RateLimiterTokenBucket:
		return "Token Bucket"
	case config.RateLimiterSlidingWindow:
		return "Sliding Window"
	case config.RateLimiterLeakyBucket:
		return "Leaky Bucket"
	case config.RateLimiterFixedWindow:
		return "Fixed Window"
	default:
		return string(rl)
	}
}

func formatObservability(enabled bool) string {
	if enabled {
		return fmt.Sprintf("%s OTLP Enabled", components.Checkbox(true))
	}
	return fmt.Sprintf("%s Disabled", components.Checkbox(false))
}

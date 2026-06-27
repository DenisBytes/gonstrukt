//go:build audit

// Package audit is a heavy, opt-in combinatorial test harness for the gonstrukt
// generator. It enumerates a pairwise-complete + edge-case set of ProjectConfigs,
// generates each into a temp dir, and verifies the result compiles (go build/vet)
// and, for frontend combos, builds the frontend.
//
// It is gated behind the `audit` build tag so it never runs as part of the normal
// `go test ./...`. Run it explicitly:
//
//	go test -tags audit ./audit -run TestAudit -timeout 60m -v
package audit

import (
	"fmt"
	"sort"

	"github.com/DenisBytes/gonstrukt/internal/config"
)

// Combo is a single generator input to be exercised by the harness.
type Combo struct {
	Name        string
	Cfg         *config.ProjectConfig
	ExpectValid bool // false => Validate() is expected to reject this config
	Integration bool // true => also run the full docker-backed test suite for this combo
}

// param is one test dimension with its discrete value domain.
type param struct {
	name   string
	values []string
}

// ---------------------------------------------------------------------------
// Generic greedy all-pairs (pairwise) generator.
//
// Guarantees that for every pair of parameters, every combination of their
// values appears together in at least one generated test vector. This catches
// the overwhelming majority of template-conditional interaction bugs at a tiny
// fraction of the full-cartesian cost.
// ---------------------------------------------------------------------------

// pairData is a single uncovered (param_i=val, param_j=val) pair, retained
// alongside its canonical key so the seed never needs to be re-parsed from a
// string (Go's fmt.Sscanf has no scanset verb, which made the old parse silently
// fail and loop forever).
type pairData struct {
	i, j   int
	vi, vj string
}

func pairKey(i, j int, vi, vj string) string {
	if i > j {
		i, j = j, i
		vi, vj = vj, vi
	}
	// NUL separators can't appear in param names/values, so the key is unambiguous.
	return fmt.Sprintf("%d\x00%s\x00%d\x00%s", i, vi, j, vj)
}

func pairwise(params []param) []map[string]string {
	// Seed the full set of uncovered (param_i=val, param_j=val) pairs.
	uncovered := map[string]pairData{}
	for i := 0; i < len(params); i++ {
		for j := i + 1; j < len(params); j++ {
			for _, vi := range params[i].values {
				for _, vj := range params[j].values {
					uncovered[pairKey(i, j, vi, vj)] = pairData{i, j, vi, vj}
				}
			}
		}
	}

	// Deterministic seed order: sort uncovered keys each round and take the first.
	var tests []map[string]string
	for len(uncovered) > 0 {
		keys := make([]string, 0, len(uncovered))
		for k := range uncovered {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		seed := uncovered[keys[0]]
		assign := make(map[int]string, len(params))
		assign[seed.i] = seed.vi
		assign[seed.j] = seed.vj

		// Greedily fill the remaining params, each time picking the value that
		// covers the most still-uncovered pairs against already-assigned params.
		for idx := 0; idx < len(params); idx++ {
			if _, ok := assign[idx]; ok {
				continue
			}
			bestVal, bestGain := params[idx].values[0], -1
			for _, v := range params[idx].values {
				gain := 0
				for a, av := range assign {
					if _, still := uncovered[pairKey(a, idx, av, v)]; still {
						gain++
					}
				}
				if gain > bestGain {
					bestGain, bestVal = gain, v
				}
			}
			assign[idx] = bestVal
		}

		// Mark every pair this vector covers.
		for a := 0; a < len(params); a++ {
			for b := a + 1; b < len(params); b++ {
				delete(uncovered, pairKey(a, b, assign[a], assign[b]))
			}
		}

		m := make(map[string]string, len(params))
		for idx, v := range assign {
			m[params[idx].name] = v
		}
		tests = append(tests, m)
	}
	return tests
}

// ---------------------------------------------------------------------------
// Dimension definitions per service type.
// ---------------------------------------------------------------------------

var commonParams = []param{
	{"config", []string{"yaml", "env", "vault"}},
	{"obs", []string{"on", "off"}},
	{"k8s", []string{"on", "off"}},
}

var gatewayParams = append([]param{
	{"cache", []string{"redis", "valkey", "memory"}},
	{"rl", []string{"token-bucket", "sliding-window", "leaky-bucket", "fixed-window"}},
	{"authcache", []string{"on", "off"}},
}, commonParams...)

var authFeatureParams = []param{
	{"db", []string{"postgres"}},
	{"cache", []string{"none", "redis", "valkey", "memory"}},
	{"rl", []string{"none", "token-bucket", "sliding-window", "leaky-bucket", "fixed-window"}},
	{"mfa", []string{"on", "off"}},
	{"rbac", []string{"on", "off"}},
	{"oauth", []string{"none", "google", "microsoft", "apple", "google+microsoft", "all"}},
	{"gdpr", []string{"none", "consent", "all"}},
	{"email", []string{"ses", "smtp"}},
	{"tenancy", []string{"on", "off"}},
	{"frontend", []string{"none", "web", "mobile", "web+mobile"}},
	{"uilib", []string{"shadcn", "baseui"}},
	{"state", []string{"tanstack", "redux"}},
	{"e2e", []string{"none", "cypress", "playwright"}},
	{"posthog", []string{"on", "off"}},
	{"sentry", []string{"on", "off"}},
	{"testinfra", []string{"docker", "testcontainers"}},
}

var authParams = append(append([]param{}, authFeatureParams...), commonParams...)

var bothParams = append(append([]param{
	{"cache", []string{"redis", "valkey", "memory"}},
	{"rl", []string{"token-bucket", "sliding-window", "leaky-bucket", "fixed-window"}},
	{"authcache", []string{"on", "off"}},
}, dropParams(authFeatureParams, "cache", "rl")...), commonParams...)

func dropParams(in []param, names ...string) []param {
	skip := map[string]bool{}
	for _, n := range names {
		skip[n] = true
	}
	var out []param
	for _, p := range in {
		if !skip[p.name] {
			out = append(out, p)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Builders: map[param]value -> *config.ProjectConfig.
//
// Constraints (GDPR=>email, frontend=>uilib/state/webframework, k8s=>domain) are
// satisfied here so every generated pairwise combo is expected to be valid.
// ---------------------------------------------------------------------------

func ptr[T any](v T) *T { return &v }

func applyCommon(cfg *config.ProjectConfig, m map[string]string) {
	cfg.ConfigSource = config.ConfigSource(m["config"])
	cfg.Observability = m["obs"] == "on"
	if m["k8s"] == "on" {
		cfg.EnableK8s = true
		cfg.Domain = "audit.dev"
	}
}

func applyCache(cfg *config.ProjectConfig, v string) {
	if v != "" && v != "none" {
		cfg.Cache = ptr(config.CacheType(v))
	}
}

func applyRateLimiter(cfg *config.ProjectConfig, v string) {
	if v != "" && v != "none" {
		cfg.RateLimiter = ptr(config.RateLimiterType(v))
	}
}

func applyOAuth(cfg *config.ProjectConfig, v string) {
	switch v {
	case "google":
		cfg.OAuthProviders = []config.OAuthProvider{config.OAuthGoogle}
	case "microsoft":
		cfg.OAuthProviders = []config.OAuthProvider{config.OAuthMicrosoft}
	case "apple":
		cfg.OAuthProviders = []config.OAuthProvider{config.OAuthApple}
	case "google+microsoft":
		cfg.OAuthProviders = []config.OAuthProvider{config.OAuthGoogle, config.OAuthMicrosoft}
	case "all":
		cfg.OAuthProviders = []config.OAuthProvider{config.OAuthGoogle, config.OAuthMicrosoft, config.OAuthApple}
	}
}

func applyGDPR(cfg *config.ProjectConfig, gdpr, email string) {
	switch gdpr {
	case "consent":
		cfg.GDPRFeatures = []config.GDPRFeature{config.GDPRConsent}
	case "all":
		cfg.GDPRFeatures = []config.GDPRFeature{
			config.GDPRConsent, config.GDPRDataExport,
			config.GDPRDataDeletion, config.GDPRProcessingLogs,
		}
	}
	if len(cfg.GDPRFeatures) > 0 {
		es := config.EmailService(email)
		if es != config.EmailSES && es != config.EmailSMTP {
			es = config.EmailSES
		}
		cfg.EmailService = ptr(es)
	}
}

func applyFrontend(cfg *config.ProjectConfig, m map[string]string) {
	switch m["frontend"] {
	case "web":
		cfg.Frontends = []config.FrontendType{config.FrontendWeb}
	case "mobile":
		cfg.Frontends = []config.FrontendType{config.FrontendMobile}
	case "web+mobile":
		cfg.Frontends = []config.FrontendType{config.FrontendWeb, config.FrontendMobile}
	}
	if len(cfg.Frontends) == 0 {
		return
	}
	hasWeb := false
	for _, f := range cfg.Frontends {
		if f == config.FrontendWeb {
			hasWeb = true
		}
	}
	if hasWeb {
		cfg.WebFramework = ptr(config.FrameworkReact)
	}
	cfg.UILibrary = ptr(config.UILibrary(m["uilib"]))
	cfg.StateManagement = ptr(config.StateManagement(m["state"]))
	cfg.EnablePostHog = m["posthog"] == "on"
	cfg.EnableSentry = m["sentry"] == "on"
	if e := m["e2e"]; e != "" && e != "none" {
		cfg.E2EFramework = ptr(config.E2EFrameworkType(e))
	}
}

func applyAuthFeatures(cfg *config.ProjectConfig, m map[string]string) {
	cfg.Database = ptr(config.DatabaseType(m["db"]))
	cfg.EnableMFA = m["mfa"] == "on"
	cfg.EnableRBAC = m["rbac"] == "on"
	cfg.EnableTenancy = m["tenancy"] == "on"
	applyOAuth(cfg, m["oauth"])
	applyGDPR(cfg, m["gdpr"], m["email"])
	applyFrontend(cfg, m)
	if ti := m["testinfra"]; ti != "" {
		cfg.TestInfra = ptr(config.TestInfraType(ti))
	}
}

func buildGateway(m map[string]string) *config.ProjectConfig {
	cfg := &config.ProjectConfig{ServiceType: config.ServiceGateway}
	applyCache(cfg, m["cache"])
	applyRateLimiter(cfg, m["rl"])
	cfg.AuthCache = m["authcache"] == "on"
	applyCommon(cfg, m)
	return cfg
}

func buildAuth(m map[string]string) *config.ProjectConfig {
	cfg := &config.ProjectConfig{ServiceType: config.ServiceAuth}
	applyCache(cfg, m["cache"])
	applyRateLimiter(cfg, m["rl"])
	applyAuthFeatures(cfg, m)
	applyCommon(cfg, m)
	return cfg
}

func buildBoth(m map[string]string) *config.ProjectConfig {
	cfg := &config.ProjectConfig{ServiceType: config.ServiceBoth}
	applyCache(cfg, m["cache"])
	applyRateLimiter(cfg, m["rl"])
	cfg.AuthCache = m["authcache"] == "on"
	applyAuthFeatures(cfg, m)
	applyCommon(cfg, m)
	return cfg
}

// ---------------------------------------------------------------------------
// Edge cases: hand-picked combos, including ones expected to be REJECTED.
// ---------------------------------------------------------------------------

func edgeCombos() []Combo {
	mk := func(name string, expectValid bool, cfg *config.ProjectConfig) Combo {
		return Combo{Name: "edge/" + name, Cfg: cfg, ExpectValid: expectValid}
	}
	return []Combo{
		mk("apple-only-oauth", true, &config.ProjectConfig{
			ServiceType: config.ServiceAuth, ConfigSource: config.ConfigYAML,
			Database: ptr(config.DBPostgres), OAuthProviders: []config.OAuthProvider{config.OAuthApple},
		}),
		mk("google-microsoft-no-apple", true, &config.ProjectConfig{
			ServiceType: config.ServiceAuth, ConfigSource: config.ConfigYAML,
			Database:       ptr(config.DBPostgres),
			OAuthProviders: []config.OAuthProvider{config.OAuthGoogle, config.OAuthMicrosoft},
		}),
		mk("web-and-mobile", true, &config.ProjectConfig{
			ServiceType: config.ServiceAuth, ConfigSource: config.ConfigYAML,
			Database:        ptr(config.DBPostgres),
			Frontends:       []config.FrontendType{config.FrontendWeb, config.FrontendMobile},
			WebFramework:    ptr(config.FrameworkReact),
			UILibrary:       ptr(config.UILibShadcn),
			StateManagement: ptr(config.StateMgmtTanStack),
		}),
		mk("gdpr-all-no-email", false, &config.ProjectConfig{
			ServiceType: config.ServiceAuth, ConfigSource: config.ConfigYAML,
			Database: ptr(config.DBPostgres),
			GDPRFeatures: []config.GDPRFeature{
				config.GDPRConsent, config.GDPRDataExport,
				config.GDPRDataDeletion, config.GDPRProcessingLogs,
			},
		}),
		mk("auth-with-cache-and-rl", true, &config.ProjectConfig{
			ServiceType: config.ServiceAuth, ConfigSource: config.ConfigEnv,
			Database: ptr(config.DBPostgres), Cache: ptr(config.CacheRedis),
			RateLimiter: ptr(config.RateLimiterTokenBucket),
		}),
		mk("gateway-minimal-memory", true, &config.ProjectConfig{
			ServiceType: config.ServiceGateway, ConfigSource: config.ConfigYAML,
			Cache: ptr(config.CacheMemory), RateLimiter: ptr(config.RateLimiterFixedWindow),
		}),
		mk("both-k8s-tenancy-obs-frontend", true, &config.ProjectConfig{
			ServiceType: config.ServiceBoth, ConfigSource: config.ConfigYAML,
			Database: ptr(config.DBPostgres), Cache: ptr(config.CacheRedis),
			RateLimiter: ptr(config.RateLimiterTokenBucket), Observability: true,
			EnableTenancy: true, EnableK8s: true, Domain: "audit.dev",
			Frontends: []config.FrontendType{config.FrontendWeb}, WebFramework: ptr(config.FrameworkReact),
			UILibrary: ptr(config.UILibShadcn), StateManagement: ptr(config.StateMgmtRedux),
		}),
		mk("postgres-oauth-off", true, &config.ProjectConfig{
			ServiceType: config.ServiceAuth, ConfigSource: config.ConfigYAML,
			Database: ptr(config.DBPostgres),
		}),
		mk("postgres-oauth-on", true, &config.ProjectConfig{
			ServiceType: config.ServiceAuth, ConfigSource: config.ConfigYAML,
			Database: ptr(config.DBPostgres), OAuthProviders: []config.OAuthProvider{config.OAuthGoogle},
		}),
		mk("vault-config", true, &config.ProjectConfig{
			ServiceType: config.ServiceAuth, ConfigSource: config.ConfigVault,
			Database: ptr(config.DBPostgres),
		}),
		mk("kitchen-sink-auth", true, &config.ProjectConfig{
			ServiceType: config.ServiceAuth, ConfigSource: config.ConfigYAML,
			Database: ptr(config.DBPostgres), Cache: ptr(config.CacheRedis),
			RateLimiter: ptr(config.RateLimiterSlidingWindow), Observability: true,
			EnableMFA: true, EnableRBAC: true, EnableTenancy: true,
			OAuthProviders: []config.OAuthProvider{config.OAuthGoogle, config.OAuthMicrosoft, config.OAuthApple},
			GDPRFeatures: []config.GDPRFeature{
				config.GDPRConsent, config.GDPRDataExport,
				config.GDPRDataDeletion, config.GDPRProcessingLogs,
			},
			EmailService: ptr(config.EmailSMTP),
			Frontends:    []config.FrontendType{config.FrontendWeb, config.FrontendMobile},
			WebFramework: ptr(config.FrameworkReact), UILibrary: ptr(config.UILibShadcn),
			StateManagement: ptr(config.StateMgmtTanStack), EnablePostHog: true, EnableSentry: true,
			E2EFramework: ptr(config.E2EFrameworkPlaywright),
		}),
		mk("tenancy-on-gateway", false, &config.ProjectConfig{
			ServiceType: config.ServiceGateway, ConfigSource: config.ConfigYAML,
			Cache: ptr(config.CacheRedis), RateLimiter: ptr(config.RateLimiterTokenBucket),
			EnableTenancy: true,
		}),
	}
}

// AllCombos returns the full pairwise + edge-case set. Every combo gets a unique
// module name so generated temp modules never collide.
func AllCombos() []Combo {
	var combos []Combo
	add := func(prefix string, builder func(map[string]string) *config.ProjectConfig, vectors []map[string]string) {
		for i, v := range vectors {
			cfg := builder(v)
			combos = append(combos, Combo{
				Name:        fmt.Sprintf("%s/%03d", prefix, i),
				Cfg:         cfg,
				ExpectValid: true,
			})
		}
	}
	add("gateway", buildGateway, pairwise(gatewayParams))
	add("auth", buildAuth, pairwise(authParams))
	add("both", buildBoth, pairwise(bothParams))
	combos = append(combos, edgeCombos()...)

	// Assign module name + project name from the combo name.
	for i := range combos {
		slug := sanitize(combos[i].Name)
		combos[i].Cfg.ModuleName = "github.com/audit/" + slug
	}

	markIntegration(combos)
	return combos
}

// integrationNames are combos that additionally run the full docker-backed test
// suite (real Postgres/Redis/Valkey). The generated code is largely identical
// across combos of a given feature set, so a curated, high-coverage subset proves
// the wired runtime behavior without standing up infra for every vector.
var integrationNames = map[string]bool{
	"gateway/000":                     true, // a gateway with cache + rate limiter
	"auth/000":                        true, // first pairwise auth vector
	"both/000":                        true, // first pairwise monorepo vector
	"edge/kitchen-sink-auth":          true, // every auth feature at once
	"edge/both-k8s-tenancy-obs-frontend": true, // monorepo + tenancy + k8s + obs
	"edge/auth-with-cache-and-rl":     true, // auth opting into gateway features
	"edge/gateway-minimal-memory":     true, // in-memory cache path
	"edge/vault-config":               true, // vault config source
}

// markIntegration flags the curated integration subset on the assembled combos.
func markIntegration(combos []Combo) {
	for i := range combos {
		if integrationNames[combos[i].Name] {
			combos[i].Integration = true
		}
	}
}

func sanitize(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			out = append(out, r)
		case r == '/' || r == '-' || r == '+':
			out = append(out, '-')
		}
	}
	return string(out)
}

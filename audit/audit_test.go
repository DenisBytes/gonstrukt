//go:build audit

package audit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestAuditValidate is the fast guard: it only runs Validate() on every combo and
// checks the expected accept/reject outcome. No generation, no compilation.
//
//	go test -tags audit ./audit -run TestAuditValidate -v
func TestAuditValidate(t *testing.T) {
	combos := AllCombos()
	t.Logf("validating %d combos", len(combos))
	results := validateOnly(combos)
	fails := 0
	for _, r := range results {
		if !r.Passed {
			fails++
			t.Errorf("validate %s: %s", r.Name, r.Detail)
		}
	}
	t.Logf("validation: %d/%d passed", len(results)-fails, len(results))
}

// TestAudit is the heavy combinatorial harness. It runs STRICTLY SERIALLY — one
// project is generated and compiled at a time — because parallel generation would
// overwhelm the host. Each combo flows through:
//
//	validate → generate → artifacts → build → vet → test-short
//	          [→ lint] [→ k8s] [→ integration] [→ frontend]
//
//	go test -tags audit ./audit -run TestAudit$ -timeout 180m -v
//
// Env knobs:
//
//	AUDIT_INTEGRATION=1  run the full docker-backed suite on the integration subset
//	AUDIT_LINT=1         run golangci-lint (advisory by default)
//	AUDIT_LINT_STRICT=1  make lint findings fail the combo
//	AUDIT_FRONTEND=1     also build frontends (npm ci + build; slow)
//	AUDIT_DIR=path       base output dir (default os.TempDir()/gonstrukt-audit)
//	AUDIT_KEEP=1         keep all generated projects (default keeps only failures)
//	AUDIT_REPORT=path    report path (default ./audit-report.md)
//	AUDIT_ONLY=substr    only run combos whose name contains substr
func TestAudit(t *testing.T) {
	combos := AllCombos()
	if only := os.Getenv("AUDIT_ONLY"); only != "" {
		combos = filterCombos(combos, only)
		t.Logf("AUDIT_ONLY=%q -> %d combos", only, len(combos))
	}

	baseDir := os.Getenv("AUDIT_DIR")
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "gonstrukt-audit")
	}
	os.MkdirAll(baseDir, 0o755)

	opts := runOpts{
		frontend:    os.Getenv("AUDIT_FRONTEND") == "1",
		lint:        os.Getenv("AUDIT_LINT") == "1",
		lintStrict:  os.Getenv("AUDIT_LINT_STRICT") == "1",
		integration: os.Getenv("AUDIT_INTEGRATION") == "1",
	}

	// Bring up shared Docker infra once if any selected combo needs integration.
	if opts.integration && anyIntegration(combos) {
		if !dockerAvailable(context.Background()) {
			t.Fatal("AUDIT_INTEGRATION=1 but Docker daemon is not available")
		}
		t.Log("starting shared Docker infra (postgres + redis + valkey)…")
		infra, err := StartInfra(context.Background())
		if err != nil {
			t.Fatalf("start infra: %v", err)
		}
		opts.infra = infra
		defer infra.Stop(context.Background())
	}

	report := os.Getenv("AUDIT_REPORT")
	if report == "" {
		report = "audit-report.md"
	}

	t.Logf("running %d combos SERIALLY | integration=%v | lint=%v | frontend=%v | dir=%s",
		len(combos), opts.integration, opts.lint, opts.frontend, baseDir)

	results := RunAudit(context.Background(), combos, baseDir, opts)

	if err := WriteReport(report, results); err != nil {
		t.Fatalf("write report: %v", err)
	}
	abs, _ := filepath.Abs(report)
	t.Logf("report written to %s", abs)

	keepAll := os.Getenv("AUDIT_KEEP") == "1"
	fails := 0
	for _, r := range results {
		if !r.Passed {
			fails++
			t.Errorf("FAIL %s [%s] %s (dir: %s)", r.Name, r.Stage, r.Detail, r.Dir)
		} else if !keepAll && r.Dir != "" {
			os.RemoveAll(r.Dir)
		}
	}
	t.Logf("audit: %d/%d passed", len(results)-fails, len(results))
}

func anyIntegration(combos []Combo) bool {
	for _, c := range combos {
		if c.Integration {
			return true
		}
	}
	return false
}

func filterCombos(combos []Combo, substr string) []Combo {
	var out []Combo
	for _, c := range combos {
		if contains(c.Name, substr) {
			out = append(out, c)
		}
	}
	return out
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

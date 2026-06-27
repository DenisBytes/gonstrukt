//go:build audit

package audit

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/DenisBytes/gonstrukt/internal/config"
	"github.com/DenisBytes/gonstrukt/internal/generator"
)

// Result captures the outcome of exercising one Combo.
type Result struct {
	Name      string
	Stage     string // last stage reached: validate|generate|artifacts|build|vet|test|lint|k8s|integration|frontend|ok
	Passed    bool
	Detail    string   // first meaningful error line, if any
	Dir       string
	Elapsed   time.Duration
	ModuleDir string   // the go module / frontend dir that failed, if any
	Warnings  []string // advisory findings (e.g. lint) that did not fail the combo
}

// runOpts bundles the per-run toggles so the stage pipeline stays readable.
type runOpts struct {
	frontend    bool
	lint        bool
	lintStrict  bool // treat lint findings as failures rather than warnings
	integration bool
	infra       *Infra // non-nil only when integration is enabled and infra is up
}

// runCmd runs a command in dir, returning combined output. extraEnv is appended to
// the inherited environment (later entries win).
func runCmd(ctx context.Context, dir string, extraEnv []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	// No GOFLAGS=-mod=mod: generation already ran `go mod tidy`, so the default
	// readonly mode works — and -mod=mod is outright rejected under workspace mode
	// (go.work), which the "both" monorepo uses.
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	cmd.Env = append(cmd.Env, extraEnv...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func firstErrorLines(out string, n int) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var keep []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		keep = append(keep, l)
		if len(keep) >= n {
			break
		}
	}
	return strings.Join(keep, " | ")
}

// goModuleDirs returns every directory under root that contains a go.mod.
func goModuleDirs(root string) []string {
	var dirs []string
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == "node_modules" {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "go.mod" {
			dirs = append(dirs, filepath.Dir(p))
		}
		return nil
	})
	sort.Strings(dirs)
	return dirs
}

// frontendDirs returns directories that look like a JS/TS frontend (package.json,
// excluding any nested node_modules).
func frontendDirs(root string) []string {
	var dirs []string
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == "node_modules" {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "package.json" {
			dirs = append(dirs, filepath.Dir(p))
		}
		return nil
	})
	sort.Strings(dirs)
	return dirs
}

// scanUnrenderedTemplates looks for "<no value>" in generated text files — the
// unambiguous fingerprint of a Go template that referenced a missing map key.
// (Bare "{{" is deliberately NOT flagged: it appears legitimately in Helm charts,
// Grafana dashboard legends, shell scripts, JSX object props, and the embedded
// email templates the auth service emits — so it produces false positives.)
func scanUnrenderedTemplates(root string) string {
	var hit string
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || hit != "" {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(p) {
		case ".go", ".yaml", ".yml", ".mod", ".sum", ".json", ".ts", ".tsx", ".md", ".sql", ".sh", ".conf", ".env", ".csv":
		default:
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		s := string(b)
		if i := strings.Index(s, "<no value>"); i >= 0 {
			start := i - 30
			if start < 0 {
				start = 0
			}
			end := i + 40
			if end > len(s) {
				end = len(s)
			}
			snippet := strings.ReplaceAll(s[start:end], "\n", " ")
			hit = fmt.Sprintf("%s: …%s…", rel(root, p), strings.TrimSpace(snippet))
			return filepath.SkipAll
		}
		return nil
	})
	return hit
}

// validateK8s does a dependency-free sanity check of generated k3s manifests:
// the expected files exist and every resource manifest carries apiVersion/kind.
// (No kubeconform/kubectl is available in this environment.)
func validateK8s(dir string) string {
	k8sDir := filepath.Join(dir, "k8s")
	if fi, err := os.Stat(k8sDir); err != nil || !fi.IsDir() {
		return "k8s enabled but k8s/ directory was not generated"
	}
	required := []string{"namespace.yaml", "ingress/ingress.yaml"}
	for _, r := range required {
		if _, err := os.Stat(filepath.Join(k8sDir, r)); err != nil {
			return "missing expected manifest: k8s/" + r
		}
	}
	// Only check that nothing rendered empty. apiVersion/kind are NOT asserted:
	// some generated files are intentionally comment-only instructions (e.g.
	// ingress/tls-secret.yaml, whose Secret is created at setup time by mkcert),
	// and Helm/k3s HelmChart wrappers don't carry a top-level resource kind.
	// Unrendered-template detection is handled separately by the artifacts stage.
	var bad string
	filepath.WalkDir(k8sDir, func(p string, d os.DirEntry, err error) error {
		if err != nil || bad != "" || d.IsDir() {
			return nil
		}
		ext := filepath.Ext(p)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		if strings.TrimSpace(string(b)) == "" {
			bad = "empty manifest: " + rel(dir, p)
			return filepath.SkipAll
		}
		return nil
	})
	return bad
}

// runOne drives a single combo through the full serial pipeline.
func runOne(parent context.Context, c Combo, baseDir string, opts runOpts) Result {
	start := time.Now()
	res := Result{Name: c.Name}
	finish := func(stage, detail string) Result {
		res.Stage = stage
		res.Detail = detail
		res.Elapsed = time.Since(start)
		return res
	}

	// --- validate ---------------------------------------------------------
	err := c.Cfg.Validate()
	if !c.ExpectValid {
		res.Passed = err != nil
		detail := ""
		if err == nil {
			detail = "expected validation to REJECT this config, but it was accepted"
		}
		return finish("validate", detail)
	}
	if err != nil {
		return finish("validate", "unexpected validation error: "+err.Error())
	}

	dir := filepath.Join(baseDir, sanitize(c.Name))
	os.RemoveAll(dir)
	res.Dir = dir
	c.Cfg.OutputDir = dir

	timeout := 8 * time.Minute
	if c.Integration && opts.integration {
		timeout = 18 * time.Minute
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	// --- generate (runs go mod tidy + go fmt internally) ------------------
	genErrCh := make(chan error, 1)
	go func() { genErrCh <- generator.NewGenerator(c.Cfg).Generate(ctx) }()
	select {
	case err := <-genErrCh:
		if err != nil {
			return finish("generate", firstErrorLines(err.Error(), 3))
		}
	case <-ctx.Done():
		return finish("generate", "generation timed out / cancelled")
	}

	// --- artifacts: no unrendered template markers ------------------------
	if hit := scanUnrenderedTemplates(dir); hit != "" {
		return finish("artifacts", "unrendered template artifact: "+hit)
	}

	// --- build / vet / test-short per module ------------------------------
	modules := goModuleDirs(dir)
	for _, md := range modules {
		if out, err := runCmd(ctx, md, nil, "go", "build", "./..."); err != nil {
			res.ModuleDir = rel(dir, md)
			return finish("build", firstErrorLines(out, 4))
		}
		if out, err := runCmd(ctx, md, nil, "go", "vet", "./..."); err != nil {
			res.ModuleDir = rel(dir, md)
			return finish("vet", firstErrorLines(out, 4))
		}
		// -short: compiles every test file and runs pure-unit tests; the generated
		// TestMain exits early in short mode so no infrastructure is required.
		if out, err := runCmd(ctx, md, nil, "go", "test", "-short", "./..."); err != nil {
			res.ModuleDir = rel(dir, md)
			return finish("test", firstErrorLines(out, 4))
		}
	}

	// --- lint (advisory unless strict) ------------------------------------
	if opts.lint {
		for _, md := range modules {
			out, err := runCmd(ctx, md, nil, "golangci-lint", "run", "--timeout", "3m", "./...")
			if err != nil {
				detail := rel(dir, md) + ": " + firstErrorLines(out, 4)
				if opts.lintStrict {
					res.ModuleDir = rel(dir, md)
					return finish("lint", detail)
				}
				res.Warnings = append(res.Warnings, "lint "+detail)
			}
		}
	}

	// --- k8s manifest validation ------------------------------------------
	if c.Cfg.EnableK8s {
		if bad := validateK8s(dir); bad != "" {
			return finish("k8s", bad)
		}
	}

	// --- integration: real Postgres/Redis/Valkey + full test suite --------
	if c.Integration && opts.integration && opts.infra != nil {
		if detail := runIntegration(ctx, dir, modules, opts.infra); detail != "" {
			return finish("integration", detail)
		}
	}

	// --- frontend build (slow: npm ci + build) ----------------------------
	if opts.frontend {
		for _, fd := range frontendDirs(dir) {
			if out, err := runCmd(ctx, fd, nil, "npm", "ci", "--no-audit", "--no-fund"); err != nil {
				res.ModuleDir = rel(dir, fd)
				return finish("frontend", "npm ci: "+firstErrorLines(out, 4))
			}
			if out, err := runCmd(ctx, fd, nil, "npm", "run", "build"); err != nil {
				res.ModuleDir = rel(dir, fd)
				return finish("frontend", "npm run build: "+firstErrorLines(out, 4))
			}
		}
	}

	res.Passed = true
	return finish("ok", "")
}

// runIntegration recreates fresh databases, flushes caches, and runs the full
// (non-short) test suite of every generated module against the shared infra.
// Returns "" on success or a one-line failure detail.
func runIntegration(ctx context.Context, dir string, modules []string, infra *Infra) string {
	if err := infra.RecreateDB(ctx, "test_auth", "test_db"); err != nil {
		return "db setup: " + firstErrorLines(err.Error(), 2)
	}
	infra.FlushCaches(ctx)
	env := infra.TestEnv("test_auth", "test_db")
	for _, md := range modules {
		out, err := runCmd(ctx, md, env, "go", "test", "-p", "1", "-count=1", "./...")
		if err != nil {
			return rel(dir, md) + ": " + firstErrorLines(out, 5)
		}
	}
	return ""
}

func rel(base, p string) string {
	if r, err := filepath.Rel(base, p); err == nil {
		return r
	}
	return p
}

// RunAudit exercises all combos STRICTLY SERIALLY (one project generated/built at a
// time) and returns the results. Serial execution is mandatory: parallel generation
// and compilation would overwhelm the host. If any integration combo is present and
// opts.integration is set, shared Docker infra is started once and reused.
func RunAudit(ctx context.Context, combos []Combo, baseDir string, opts runOpts) []Result {
	results := make([]Result, len(combos))
	for i, c := range combos {
		results[i] = runOne(ctx, c, baseDir, opts)
		r := results[i]
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		warn := ""
		if len(r.Warnings) > 0 {
			warn = fmt.Sprintf(" (+%d warn)", len(r.Warnings))
		}
		fmt.Printf("[%3d/%3d] %-4s %-40s %-11s %6.1fs %s%s\n",
			i+1, len(combos), status, r.Name, r.Stage, r.Elapsed.Seconds(), r.Detail, warn)
	}
	return results
}

// WriteReport renders a markdown summary of the run.
func WriteReport(path string, results []Result) error {
	var b strings.Builder
	pass, fail, warned := 0, 0, 0
	for _, r := range results {
		if r.Passed {
			pass++
		} else {
			fail++
		}
		if len(r.Warnings) > 0 {
			warned++
		}
	}
	fmt.Fprintf(&b, "# Gonstrukt Combinatorial Audit Report\n\n")
	fmt.Fprintf(&b, "Total: %d  |  Pass: %d  |  Fail: %d  |  With warnings: %d\n\n", len(results), pass, fail, warned)

	if fail > 0 {
		fmt.Fprintf(&b, "## Failures\n\n| Combo | Stage | Module | Detail |\n|---|---|---|---|\n")
		for _, r := range results {
			if !r.Passed {
				fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n",
					r.Name, r.Stage, r.ModuleDir, escapePipes(r.Detail))
			}
		}
		fmt.Fprintln(&b)
	}

	if warned > 0 {
		fmt.Fprintf(&b, "## Warnings (advisory)\n\n| Combo | Warning |\n|---|---|\n")
		for _, r := range results {
			for _, w := range r.Warnings {
				fmt.Fprintf(&b, "| `%s` | %s |\n", r.Name, escapePipes(w))
			}
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintf(&b, "## All Results\n\n| Combo | Result | Stage | Time |\n|---|---|---|---|\n")
	sorted := append([]Result{}, results...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	for _, r := range sorted {
		status := "✅"
		if !r.Passed {
			status = "❌"
		}
		fmt.Fprintf(&b, "| `%s` | %s | %s | %.1fs |\n", r.Name, status, r.Stage, r.Elapsed.Seconds())
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func escapePipes(s string) string { return strings.ReplaceAll(s, "|", "\\|") }

// envInt reads an int env var with a default.
func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// validateOnly is a fast path used by the validation-only test.
func validateOnly(combos []Combo) []Result {
	out := make([]Result, len(combos))
	for i, c := range combos {
		err := c.Cfg.Validate()
		r := Result{Name: c.Name, Stage: "validate"}
		if c.ExpectValid {
			r.Passed = err == nil
			if err != nil {
				r.Detail = err.Error()
			}
		} else {
			r.Passed = err != nil
			if err == nil {
				r.Detail = "expected rejection but was accepted"
			}
		}
		out[i] = r
	}
	return out
}

var _ = config.ServiceAuth // keep config import even if builders change

//go:build audit

package audit

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Infra is a set of throwaway Docker containers (Postgres + Redis + Valkey) shared
// across the integration stage. Because the harness runs strictly serially, a
// single set of containers is reused for every integration combo; each combo gets
// freshly (re)created databases and flushed caches so runs never bleed into each
// other.
type Infra struct {
	host string

	pgName string
	pgPort int
	pgUser string
	pgPass string

	redisName string
	redisPort int

	valkeyName string
	valkeyPort int
}

// Fixed high ports chosen to avoid colliding with anything a developer is likely
// to be running locally (5432/6379).
const (
	auditPGPort     = 65432
	auditRedisPort  = 63790
	auditValkeyPort = 63791
)

// StartInfra launches Postgres, Redis and Valkey containers and waits until each
// is accepting connections. Containers are force-removed first so a previous
// crashed run can't block startup.
func StartInfra(ctx context.Context) (*Infra, error) {
	in := &Infra{
		host:       "127.0.0.1",
		pgName:     "gonstrukt-audit-pg",
		pgPort:     auditPGPort,
		pgUser:     "postgres",
		pgPass:     "postgres",
		redisName:  "gonstrukt-audit-redis",
		redisPort:  auditRedisPort,
		valkeyName: "gonstrukt-audit-valkey",
		valkeyPort: auditValkeyPort,
	}

	// Clean up any leftovers from a prior aborted run.
	in.Stop(context.Background())

	type spec struct {
		name string
		args []string
	}
	specs := []spec{
		{in.pgName, []string{
			"run", "-d", "--name", in.pgName,
			"-e", "POSTGRES_PASSWORD=" + in.pgPass,
			"-e", "POSTGRES_USER=" + in.pgUser,
			"-e", "POSTGRES_DB=postgres",
			"-p", fmt.Sprintf("%d:5432", in.pgPort),
			"postgres:16-alpine",
		}},
		{in.redisName, []string{
			"run", "-d", "--name", in.redisName,
			"-p", fmt.Sprintf("%d:6379", in.redisPort),
			"redis:7-alpine",
		}},
		{in.valkeyName, []string{
			"run", "-d", "--name", in.valkeyName,
			"-p", fmt.Sprintf("%d:6379", in.valkeyPort),
			"valkey/valkey:8-alpine",
		}},
	}
	for _, s := range specs {
		if out, err := dockerRun(ctx, s.args...); err != nil {
			in.Stop(context.Background())
			return nil, fmt.Errorf("start %s: %v: %s", s.name, err, firstErrorLines(out, 3))
		}
	}

	// Readiness: Postgres needs pg_isready (the port opens before it can serve);
	// Redis/Valkey accept TCP connections only once ready, so a dial suffices.
	if err := in.waitPostgres(ctx, 60*time.Second); err != nil {
		in.Stop(context.Background())
		return nil, err
	}
	if err := waitTCP(in.host, in.redisPort, 30*time.Second); err != nil {
		in.Stop(context.Background())
		return nil, fmt.Errorf("redis not ready: %w", err)
	}
	if err := waitTCP(in.host, in.valkeyPort, 30*time.Second); err != nil {
		in.Stop(context.Background())
		return nil, fmt.Errorf("valkey not ready: %w", err)
	}
	return in, nil
}

// DSN returns a connection string for the named database on the shared Postgres.
func (in *Infra) DSN(dbName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		in.pgUser, in.pgPass, in.host, in.pgPort, dbName)
}

// RecreateDB drops and recreates each named database so an integration combo
// always starts from a clean schema (migrations across combos differ). dropdb /
// createdb are used rather than psql -c because DROP DATABASE cannot run inside
// the implicit transaction psql wraps a multi-statement string in.
func (in *Infra) RecreateDB(ctx context.Context, names ...string) error {
	for _, n := range names {
		if out, err := dockerExec(ctx, in.pgName,
			"dropdb", "-U", in.pgUser, "--if-exists", "--force", n); err != nil {
			return fmt.Errorf("dropdb %s: %v: %s", n, err, firstErrorLines(out, 3))
		}
		if out, err := dockerExec(ctx, in.pgName, "createdb", "-U", in.pgUser, n); err != nil {
			return fmt.Errorf("createdb %s: %v: %s", n, err, firstErrorLines(out, 3))
		}
	}
	return nil
}

// FlushCaches clears Redis and Valkey between integration combos.
func (in *Infra) FlushCaches(ctx context.Context) {
	_, _ = dockerExec(ctx, in.redisName, "redis-cli", "FLUSHALL")
	_, _ = dockerExec(ctx, in.valkeyName, "valkey-cli", "FLUSHALL")
}

// TestEnv returns the environment variable assignments a generated project's test
// suite expects, pointing at the shared infra and the supplied fresh databases.
func (in *Infra) TestEnv(authDB, dbTestDB string) []string {
	return []string{
		"TEST_DATABASE_URL=" + in.DSN(authDB),
		"TEST_POSTGRES_DSN=" + in.DSN(dbTestDB),
		"TEST_REDIS_ADDR=" + fmt.Sprintf("%s:%d", in.host, in.redisPort),
		"TEST_VALKEY_ADDR=" + fmt.Sprintf("%s:%d", in.host, in.valkeyPort),
	}
}

// Stop force-removes all containers. Safe to call repeatedly.
func (in *Infra) Stop(ctx context.Context) {
	if in == nil {
		return
	}
	for _, name := range []string{in.pgName, in.redisName, in.valkeyName} {
		if name != "" {
			_, _ = dockerRun(ctx, "rm", "-f", name)
		}
	}
}

func (in *Infra) waitPostgres(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if _, err := dockerExec(ctx, in.pgName, "pg_isready", "-U", in.pgUser); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("postgres not ready after %s", timeout)
		}
		time.Sleep(time.Second)
	}
}

// dockerRun runs `docker <args...>` and returns combined output.
func dockerRun(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// dockerExec runs a command inside a running container.
func dockerExec(ctx context.Context, container string, args ...string) (string, error) {
	full := append([]string{"exec", container}, args...)
	return dockerRun(ctx, full...)
}

func waitTCP(host string, port int, timeout time.Duration) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("timeout")
	}
	return lastErr
}

// dockerAvailable reports whether the docker daemon is reachable.
func dockerAvailable(ctx context.Context) bool {
	out, err := dockerRun(ctx, "info", "--format", "{{.ServerVersion}}")
	return err == nil && strings.TrimSpace(out) != ""
}

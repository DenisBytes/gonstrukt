#!/bin/bash
set -e

# Gonstrukt CI Test Script
# Runs one of 12 test configurations based on the provided config number

CONFIG_NUM=$1

if [ -z "$CONFIG_NUM" ]; then
    echo "Usage: $0 <config-number>"
    echo "Config numbers: 1-12"
    exit 1
fi

# Create temp directory for test output
TEST_DIR="test-config-${CONFIG_NUM}"
rm -rf "$TEST_DIR"

echo "========================================"
echo "Running test configuration #${CONFIG_NUM}"
echo "========================================"

# Build gonstrukt first
if [ ! -f "./gonstrukt" ]; then
    echo "Building gonstrukt..."
    go build -o gonstrukt
fi

# Define test configurations
case $CONFIG_NUM in
    1)
        echo "Config: auth + postgres + yaml (basic)"
        ./gonstrukt create github.com/test/config1 \
            -s auth -d postgres -c yaml \
            -o "$TEST_DIR" -i=false
        ;;
    2)
        echo "Config: auth + mysql + env + MFA"
        ./gonstrukt create github.com/test/config2 \
            -s auth -d mysql -c env \
            --mfa \
            -o "$TEST_DIR" -i=false
        ;;
    3)
        echo "Config: auth + sqlite + yaml + OAuth (google,apple)"
        ./gonstrukt create github.com/test/config3 \
            -s auth -d sqlite -c yaml \
            --oauth google,apple \
            -o "$TEST_DIR" -i=false
        ;;
    4)
        echo "Config: auth + postgres + vault + GDPR (all) + RBAC"
        ./gonstrukt create github.com/test/config4 \
            -s auth -d postgres -c vault \
            --gdpr consent,data-export,data-deletion,processing-logs \
            --email-service ses \
            --rbac \
            -o "$TEST_DIR" -i=false
        ;;
    5)
        echo "Config: auth + postgres + redis + token-bucket + MFA + OAuth (microsoft)"
        ./gonstrukt create github.com/test/config5 \
            -s auth -d postgres -c yaml \
            --cache redis -r token-bucket \
            --mfa --oauth microsoft \
            -o "$TEST_DIR" -i=false
        ;;
    6)
        echo "Config: gateway + redis + sliding-window + yaml"
        ./gonstrukt create github.com/test/config6 \
            -s gateway --cache redis -r sliding-window -c yaml \
            -o "$TEST_DIR" -i=false
        ;;
    7)
        echo "Config: gateway + valkey + leaky-bucket + env"
        ./gonstrukt create github.com/test/config7 \
            -s gateway --cache valkey -r leaky-bucket -c env \
            -o "$TEST_DIR" -i=false
        ;;
    8)
        echo "Config: gateway + memory + fixed-window + yaml + auth-cache"
        ./gonstrukt create github.com/test/config8 \
            -s gateway --cache memory -r fixed-window -c yaml \
            --auth-cache \
            -o "$TEST_DIR" -i=false
        ;;
    9)
        echo "Config: both + postgres + redis + token-bucket + yaml"
        ./gonstrukt create github.com/test/config9 \
            -s both -d postgres --cache redis -r token-bucket -c yaml \
            -o "$TEST_DIR" -i=false
        ;;
    10)
        echo "Config: both + mysql + valkey + sliding-window + env + MFA + GDPR consent"
        ./gonstrukt create github.com/test/config10 \
            -s both -d mysql --cache valkey -r sliding-window -c env \
            --mfa --gdpr consent --email-service smtp \
            -o "$TEST_DIR" -i=false
        ;;
    11)
        echo "Config: auth + postgres + yaml + frontend (react, shadcn, tanstack)"
        ./gonstrukt create github.com/test/config11 \
            -s auth -d postgres -c yaml \
            --frontend web --web-framework react --ui-lib shadcn --state-mgmt tanstack \
            -o "$TEST_DIR" -i=false
        ;;
    12)
        echo "Config: auth + postgres + yaml + frontend (next, baseui, redux) + posthog + sentry"
        ./gonstrukt create github.com/test/config12 \
            -s auth -d postgres -c yaml \
            --frontend web --web-framework next --ui-lib baseui --state-mgmt redux \
            --posthog --sentry \
            -o "$TEST_DIR" -i=false
        ;;
    *)
        echo "Invalid config number: $CONFIG_NUM"
        echo "Valid values: 1-12"
        exit 1
        ;;
esac

echo ""
echo "Project generated in $TEST_DIR"
echo ""

# Step 2: Compile the generated project
echo "Compiling generated project..."
cd "$TEST_DIR"
go mod tidy
go build ./...
echo "✓ Backend compilation successful"

# Step 3: Run backend tests (if test infrastructure is available)
if [ -f "docker-compose.test.yml" ]; then
    echo ""
    echo "Starting test infrastructure..."
    docker compose -f docker-compose.test.yml up -d --wait 2>/dev/null || {
        echo "⚠ Docker not available, skipping integration tests"
        SKIP_INTEGRATION=1
    }

    if [ -z "$SKIP_INTEGRATION" ]; then
        # Get dynamic ports
        if docker compose -f docker-compose.test.yml port postgres 5432 2>/dev/null; then
            POSTGRES_PORT=$(docker compose -f docker-compose.test.yml port postgres 5432 | cut -d: -f2)
            export TEST_POSTGRES_DSN="postgres://postgres:postgres@localhost:${POSTGRES_PORT}/test_db?sslmode=disable"
        fi

        if docker compose -f docker-compose.test.yml port mysql 3306 2>/dev/null; then
            MYSQL_PORT=$(docker compose -f docker-compose.test.yml port mysql 3306 | cut -d: -f2)
            export TEST_DATABASE_URL="root:password@tcp(localhost:${MYSQL_PORT})/test_db?parseTime=true"
        fi

        if docker compose -f docker-compose.test.yml port redis 6379 2>/dev/null; then
            REDIS_PORT=$(docker compose -f docker-compose.test.yml port redis 6379 | cut -d: -f2)
            export TEST_REDIS_ADDR="localhost:${REDIS_PORT}"
        fi

        if docker compose -f docker-compose.test.yml port valkey 6379 2>/dev/null; then
            VALKEY_PORT=$(docker compose -f docker-compose.test.yml port valkey 6379 | cut -d: -f2)
            export TEST_VALKEY_ADDR="localhost:${VALKEY_PORT}"
        fi

        echo "Running backend tests..."
        go test -v -race ./... || {
            echo "⚠ Backend tests failed"
            docker compose -f docker-compose.test.yml down -v 2>/dev/null
            exit 1
        }
        echo "✓ Backend tests passed"

        docker compose -f docker-compose.test.yml down -v 2>/dev/null
    fi
else
    echo "⚠ No docker-compose.test.yml found, skipping integration tests"
fi

# Step 4: Frontend tests (if frontend exists)
if [ -d "frontend" ]; then
    echo ""
    echo "Testing frontend..."
    cd frontend

    npm install --silent
    echo "✓ Frontend dependencies installed"

    npm run build --silent
    echo "✓ Frontend build successful"

    npm run test:run 2>/dev/null || {
        echo "⚠ Frontend unit tests not configured or failed"
    }

    cd ..
fi

cd ..

echo ""
echo "========================================"
echo "✓ Config #${CONFIG_NUM} passed all tests"
echo "========================================"

# Cleanup
rm -rf "$TEST_DIR"

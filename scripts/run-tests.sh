#!/bin/bash
# Run integration tests with docker-compose test infrastructure
#
# Usage: ./scripts/run-tests.sh [test-args...]
# Example: ./scripts/run-tests.sh -v ./internals/db/...

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Start test infrastructure
echo "Starting test infrastructure..."
docker compose -f docker-compose.test.yml up -d --wait

# Get dynamic ports
POSTGRES_PORT=$(docker compose -f docker-compose.test.yml port postgres 5432 | cut -d: -f2)
MYSQL_PORT=$(docker compose -f docker-compose.test.yml port mysql 3306 | cut -d: -f2)
REDIS_PORT=$(docker compose -f docker-compose.test.yml port redis 6379 | cut -d: -f2)
VALKEY_PORT=$(docker compose -f docker-compose.test.yml port valkey 6380 | cut -d: -f2)
MONGO_PORT=$(docker compose -f docker-compose.test.yml port mongodb 27017 | cut -d: -f2)

echo "Test infrastructure ready:"
echo "  PostgreSQL: localhost:$POSTGRES_PORT"
echo "  MySQL:      localhost:$MYSQL_PORT"
echo "  Redis:      localhost:$REDIS_PORT"
echo "  Valkey:     localhost:$VALKEY_PORT"
echo "  MongoDB:    localhost:$MONGO_PORT"

# Export environment variables for tests
export TEST_POSTGRES_DSN="postgres://postgres:postgres@localhost:$POSTGRES_PORT/test_db?sslmode=disable"
export TEST_MYSQL_DSN="root:root@tcp(localhost:$MYSQL_PORT)/test_db?parseTime=true"
export TEST_REDIS_ADDR="localhost:$REDIS_PORT"
export TEST_VALKEY_ADDR="localhost:$VALKEY_PORT"
export TEST_MONGODB_URL="mongodb://localhost:$MONGO_PORT"
export TEST_DATABASE_URL="$TEST_POSTGRES_DSN"  # Default to postgres

# Run tests
echo ""
echo "Running tests..."
if [ $# -eq 0 ]; then
    go test ./...
else
    go test "$@"
fi
TEST_EXIT_CODE=$?

# Cleanup
echo ""
echo "Stopping test infrastructure..."
docker compose -f docker-compose.test.yml down -v

exit $TEST_EXIT_CODE

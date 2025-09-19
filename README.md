# Gonstrukt

A CLI tool to spawn Go microservices with different configurations.

## Overview

Gonstrukt generates Go microservices with configurable options for databases, caching, observability, and rate limiting. It supports creating gateway and authentication services with predefined architectural patterns.

## Installation

```bash
go build -o gonstrukt
```

## Usage

### Basic Command Structure

```bash
gonstrukt create <github.com/username/project> [flags]
```

### Required Parameters

- **Service Name**: Must follow format `github.com/<username/org>/<project_name>`
- **Service Type** (`-s`, `--service-type`): `gateway` or `auth`
- **Database** (`-d`, `--database`): `psql`
- **Config** (`--config`): `yaml` or `vault`

### Conditional Requirements

**Gateway Services** require:
- **Cache** (`--cache`): `memory`, `redis`, or `valkey`
- **Rate Limiter** (`-r`, `--rate-limiter`): `token-bucket` or `approximated-sliding-window`

**Auth Services** have optional:
- Cache and rate limiter (not required)

### Optional Parameters

- **Observability** (`-o`, `--observability`): `otlp` or `none` (defaults to `otlp`)

## Examples

### Gateway Service

```bash
gonstrukt create github.com/myorg/api-gateway \
  --service-type gateway \
  --database psql \
  --config yaml \
  --cache redis \
  --rate-limiter token-bucket \
  --observability otlp
```

### Auth Service

```bash
gonstrukt create github.com/myorg/auth-service \
  -s auth \
  -d psql \
  --config vault \
  -o none
```

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
gonstrukt completion bash > /etc/bash_completion.d/gonstrukt

# Zsh
gonstrukt completion zsh > "${fpath[1]}/_gonstrukt"

# Fish
gonstrukt completion fish > ~/.config/fish/completions/gonstrukt.fish
```

## Service Types

### Gateway
- **Purpose**: API gateway for routing and traffic management
- **Required**: Cache (for performance), Rate limiter (for traffic control)
- **Features**: Request routing, load balancing, rate limiting

### Auth
- **Purpose**: Authentication and authorization service
- **Optional**: Cache and rate limiter
- **Features**: User authentication, token management, authorization

## Configuration Options

### Databases
- **psql**: PostgreSQL database

### Cache Types
- **memory**: In-memory caching
- **redis**: Redis cache
- **valkey**: Valkey cache

### Config Sources
- **yaml**: Local YAML configuration files
- **vault**: HashiCorp Vault configuration

### Rate Limiters
- **token-bucket**: Token bucket algorithm
- **approximated-sliding-window**: Approximated sliding window algorithm

### Observability
- **otlp**: OpenTelemetry Protocol for traces and spans
- **none**: No observability (logs still generated)

## Error Handling

The CLI provides comprehensive error messages with suggestions for valid options:

```
error: cache is required for gateway services: caching is mandatory for performance
error: rate-limiter is required for gateway services: rate limiting is essential for gateway traffic control

Usage:
  gonstrukt create <git_repo_url> [flags]
...
```

## Development

### Building

```bash
go build -o gonstrukt
```

### Testing

```bash
go test ./...
```

### Dependencies

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Color](https://github.com/fatih/color) - Terminal colors
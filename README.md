# Gonstrukt

> "Don't Roll Your Own Auth, They Said. They Also Said 'Just Use Firebase.' I Wrote Something Better."

Generate production-ready, GDPR-compliant authentication services in Go — in one command.

## 30 Seconds to Auth

```bash
# Install
go install github.com/DenisBytes/gonstrukt@latest

# Generate a complete auth service
gonstrukt create github.com/you/myauth \
  -s auth -d postgres -c yaml \
  --oauth google,apple --mfa \
  --gdpr consent,data-export,data-deletion
```

**What you just got:**
- JWT authentication with refresh tokens
- OAuth (Google, Apple) + MFA/TOTP with backup codes
- GDPR compliance: consent management, data export, right to erasure
- PostgreSQL with migrations
- OpenTelemetry tracing & metrics
- Production-ready security headers, rate limiting, TLS config

No Firebase. No Auth0 bills. No vendor lock-in. Just your code, your database, your users.

## What Gets Generated

### Authentication
- JWT access/refresh tokens with ECDSA signing
- Session management with device tracking
- Password hashing (bcrypt) with secure defaults
- Token blacklisting for logout/revocation

### OAuth Providers
- Google, Microsoft, Apple Sign-In
- PKCE flow for security
- Automatic account linking

### Multi-Factor Authentication
- TOTP with authenticator apps
- 10 backup codes (bcrypt hashed)
- Rate limiting (5 attempts → 15min lockout)

### GDPR Compliance
- Versioned consent records with audit trail
- Data export (Article 20 - Portability)
- Account deletion with PII anonymization (Article 17)
- Processing logs with legal basis tracking

### Security
- Secure headers (HSTS, CSP, X-Frame-Options)
- TLS 1.2+ with forward secrecy
- Request validation & rate limiting
- IP hashing for privacy

### Observability
- OpenTelemetry traces & metrics
- Structured logging with Zap
- Health check endpoints

### Optional Frontend
- React/Next.js/TanStack Start
- React Native Expo
- PostHog analytics, Sentry error tracking

## Configuration

### Service Types
```bash
-s gateway    # API gateway with rate limiting & caching
-s auth       # Authentication service
-s both       # Monorepo with both services
```

### Databases
```bash
-d postgres   # PostgreSQL (recommended)
-d mysql      # MySQL
-d sqlite     # SQLite (dev/small projects)
```

### Caching & Rate Limiting
```bash
--cache redis              # Redis or Valkey
--cache memory             # In-memory (dev only)
-r token-bucket            # Rate limiter algorithm
-r sliding-window
```

### Auth Features
```bash
--oauth google,microsoft,apple   # OAuth providers
--mfa                            # TOTP multi-factor auth
--rbac                           # Role-based access control
--gdpr consent,data-export,data-deletion,processing-logs
```

### Frontend (Optional)
```bash
--frontend web,mobile            # Generate frontend apps
--web-framework react            # react | next | tanstack
--ui-lib shadcn                  # shadcn | baseui
--state-mgmt tanstack            # tanstack | redux
--posthog --sentry               # Analytics & error tracking
```

## Interactive Mode

Don't like flags? Just run:
```bash
gonstrukt create
```
A TUI wizard walks you through every option.

## Project Structure (Generated)

```
myauth/
├── cmd/auth_service/    # Entry point
├── internals/
│   ├── services/        # Auth, OAuth, MFA, GDPR handlers
│   ├── middleware/      # JWT validation, rate limiting, security
│   ├── db/              # Database layer with migrations
│   └── telemetry/       # OpenTelemetry setup
├── config.yaml          # Configuration
├── Dockerfile
└── docker-compose.yml
```

## Run It

```bash
cd myauth
cp config.yaml.example config.yaml  # Configure your secrets
docker-compose up -d                 # Start Postgres
go run ./cmd/auth_service            # Run the service
```

## License

MIT — Do whatever you want. Roll your own auth. I dare you.

---

<p align="center">
  <i>Built by someone who got mass amounts of dopamine from making auth service generation trivial.</i>
</p>

# AI Cost Control Platform

A service for quota management and consumption tracking of corporate AI resources (LLM tokens and other metrics) for development, analytics, testing, support, and HR teams.

AI Cost Control Platform enables real-time cost control for LLM usage, flexible budget allocation across departments, and transparent reporting for management.

---

## Features

- Budget management with support for multiple currencies: tokens, USD, RUB, and custom resources
- Flexible budget periods: monthly, quarterly, yearly — with no limits on the number of periods
- Real-time ingestion of spending events via Kafka
- Strict idempotency: duplicate event delivery never results in double spending
- Role-based access control (RBAC): owners, limit managers, spenders, and viewers
- Transparent consumption statistics by team, period, and currency
- Business and technical metrics exported to Prometheus
- Full observability: structured logs, latency tracking, processing statuses
- Integration with AI Gateway and other internal systems via gRPC and HTTP API

---

## Architecture

The service follows clean architecture principles:



### Components

| Component | Purpose |
|-----------|---------|
| `budget-api` | HTTP + gRPC server: budget management, authentication, statistics |
| `budget-consumer` | Kafka consumer: spending event processing, idempotency, DLQ |
| `PostgreSQL` | Data storage: budgets, periods, limits, balances, spendings, users |
| `Kafka` | Event bus for spending events from AI Gateway |

### Transports

| Transport | Purpose |
|-----------|---------|
| **Kafka** | Ingestion of spending events. Partitioning by budget ensures sequential processing without race conditions |
| **gRPC** | Budget management and operational data retrieval for internal services |
| **HTTP REST** | Statistics for dashboards, health checks, Prometheus metrics |

---

## Tech Stack

- **Go 1.22+**
- **PostgreSQL 16** — primary data store
- **Kafka** — event streaming platform
- **Echo v4** — HTTP framework
- **gRPC + Protobuf** — inter-service communication
- **pgx v5** — PostgreSQL driver
- **squirrel** — SQL query builder
- **goose** — database schema migrations
- **JWT** — authentication
- **bcrypt** — password hashing
- **Prometheus** — metrics
- **log/slog** — structured logging
- **Docker Compose** — local development and integration tests
- **testcontainers-go** — integration testing

---

## Data Model

Core entities:

- **Budget** — a budget container for a team or department
- **BudgetPeriod** — a time period during which a budget is active (month, quarter, year)
- **Currency** — a dynamic currency (TOKENS, USD, RUB, and others)
- **PeriodLimit** — a limit for a specific currency within a period
- **PeriodBalance** — a materialized remaining balance for a currency within a period
- **Spending** — a spending record with an idempotency key
- **User** — a system user
- **Role** — a user role (owner, spender, viewer, limit_manager)

All amounts are stored as `NUMERIC(38,18)` and processed using `decimal.Decimal`.

Periods cannot overlap for the same budget — this is enforced at the database level via an exclusion constraint.

---

## Security

- Passwords are stored only as bcrypt hashes (cost ≥ 10)
- User enumeration protection: identical response for invalid username and invalid password
- Timing attack protection: comparison against a dummy hash for non-existent users
- Roles are checked on every request and are not cached in JWT
- Kafka spending ingestion requires the `spender` role
- Validation errors and forbidden operations are logged without sensitive data

---

## Observability

### Business Metrics

- Remaining balances by budget and currency
- Volume of processed spendings
- Number of duplicates and errors
- Negative balance indicator

### Technical Metrics

- HTTP, gRPC, and Kafka handler latency
- Kafka partition lag
- DLQ message count

Metrics are exported to Prometheus at `/metrics`.

Logs are structured, in JSON format (in production), with request and user identifiers.

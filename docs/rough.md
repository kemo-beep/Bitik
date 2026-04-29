# Shopee-like Architecture

## Core Backend

- `Go`
- `Gin`
- `PostgreSQL( NEON DB)`
- `Redis`
- `Goose`
- Message queue

### Message Queue Plan

- `RabbitMQ` for jobs
- `Kafka` later for event streaming
- `NATS` later for real-time fanout

## Real-time

- WebSocket service in `Go`
- `NATS` for internal real-time messaging

## File Storage

- Cloudflare R2 (production) / MinIO (local development)
- Object storage for media

## CDN

- Cloudflare CDN

## Search & Supporting Components

- OpenSearch
- Background workers
- WebSocket/chat service

## Authentication & Users

- JWT access tokens
- Refresh tokens stored in DB/Redis
- OAuth login: Google, Apple, Facebook
- Phone OTP login
- Seller account verification
- RBAC for admin panel
- Device/session management

## Payments

### Pay on delivery (POD)

- Checkout option: customer pays when the order is delivered (cash or whatever delivery policy allows).
- Order lifecycle: confirm order → fulfill → capture payment on delivery; caps / allowed cities / fraud rules as needed.

### Pay via Wave (manual merchant confirmation)

- Customer chooses **Wave** at checkout and pays from their Wave wallet to **Bitik’s Wave business account** (amount + instructions shown in-app: e.g. exact amount, order reference / memo to include).
- **No automatic Wave webhook is assumed** for v1: the **Bitik owner** (or delegated staff) **checks incoming transfers in the Wave app** (or bank/Wave merchant statement), matches **amount + reference + time** to the pending Bitik order, then **approves the payment inside Bitik** (admin or seller ops UI).
- Platform state: order/payment stays **`pending_manual_wave_confirmation`** until approval; on approve → **`paid`** (audit: approver user id, timestamp, optional note). Reject / timeout path for stale or non-matching transfers.
- Later: optional upgrade to Wave **API** / automated notifications if/when available and contracted, without changing the core “human verify then approve” safety model if you want to keep it.

### In-app e-wallet (later)

- Users top up / hold balance inside Bitik and pay checkout from wallet; settlement, KYC, and float rules TBD. **Defer** until core checkout (POD + Wave manual) is stable.

### Cross-cutting

- Payment webhook service for any **future** automated providers; idempotency keys on checkout and payment mutations; clear payment + order state machine shared by checkout, admin approval, and jobs (timeouts, cancellation).

## Core Services

Split into modules first, then microservices later:

- Auth service
- User service
- Seller service
- Product service
- Category service
- Inventory service
- Cart service
- Checkout service
- Order service
- Payment service
- Shipping service
- Promotion/Voucher service
- Review service
- Notification service
- Chat service
- Admin service
- Analytics/event service

## Infrastructure

### Container

- Docker

### Deployment

- Dokploy for early stage
- Kubernetes later

### Reverse Proxy

- Caddy or Nginx

### CI/CD

- GitHub Actions

### Secrets

- Doppler / Infisical / AWS Secrets Manager

### Monitoring

- Prometheus
- Grafana
- Loki
- Sentry
- Uptime Kuma

### Tracing

- OpenTelemetry
- Jaeger / Tempo

## Background Jobs

Use RabbitMQ workers for:

- Send emails
- Send SMS/OTP
- Payment confirmation
- Invoice generation
- Image processing
- Search indexing
- Order timeout cancellation
- Refund processing
- Seller reports
- Notification fanout

Go libraries

Use:
Swagger UI
gin-gonic/gin HTTP framework
jackc/pgx PostgreSQL driver
sqlc type-safe SQL
goose migrations
redis/go-redis Redis
segmentio/kafka-go Kafka
rabbitmq/amqp091-go RabbitMQ
nats-io/nats.go NATS
golang-jwt/jwt JWT
casbin RBAC/permissions
zap or zerolog logging
viper config
validator/v10 validation
swaggo/gin-swagger Swagger docs
stripe/stripe-go Stripe

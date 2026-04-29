FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/bitik-api ./cmd/api

FROM alpine:3.21
WORKDIR /srv
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/bitik-api /usr/local/bin/bitik-api
COPY --from=builder /app/db /srv/db
COPY --from=builder /app/internal/openapi /srv/internal/openapi

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/bitik-api"]

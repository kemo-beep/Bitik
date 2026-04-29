package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"sync"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/jwtutil"
	"github.com/bitik/backend/internal/notify"
	"github.com/bitik/backend/internal/platform/db"
	"github.com/bitik/backend/internal/platform/observability"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger, err := observability.NewLogger(cfg.Observability)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	addr := valueOrDefault(os.Getenv("BITIK_WS_ADDR"), ":8081")

	var pgPool *pgxpool.Pool
	if pool, err := db.Connect(context.Background(), cfg.Database); err != nil {
		logger.Warn("ws_database_unavailable", zap.Error(err))
	} else {
		pgPool = pool
		defer pgPool.Close()
	}
	pub := notify.NewPostgresPublisher(pgPool)

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/chat", wsHandler(cfg, logger, upgrader, pub, "chat"))
	mux.HandleFunc("/ws/notifications", wsHandler(cfg, logger, upgrader, pub, "notifications"))
	mux.HandleFunc("/ws/order-status", wsHandler(cfg, logger, upgrader, pub, "order_status"))

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("websocket_service_starting", zap.String("addr", addr), zap.String("environment", cfg.App.Environment))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("websocket_service_failed", zap.Error(err))
		}
	}()

	waitForShutdown(srv, logger, 10*time.Second)
}

func wsHandler(cfg config.Config, logger *zap.Logger, upgrader websocket.Upgrader, pub notify.Publisher, channel string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, roles, err := authenticateWS(cfg, r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn("ws_upgrade_failed", zap.Error(err))
			return
		}
		defer conn.Close()

		var writeMu sync.Mutex
		writeJSON := func(v any) error {
			writeMu.Lock()
			defer writeMu.Unlock()
			return conn.WriteJSON(v)
		}
		writeMessage := func(mt int, data []byte) error {
			writeMu.Lock()
			defer writeMu.Unlock()
			return conn.WriteMessage(mt, data)
		}

		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		_ = writeJSON(map[string]any{
			"type":    "welcome",
			"channel": channel,
			"user_id": userID,
			"roles":   roles,
		})

		var sub notify.Subscription
		if pub != nil {
			sub = pub.Subscribe(userID, 128)
			defer sub.Cancel()
			go func() {
				for evt := range sub.Events {
					switch channel {
					case "chat":
						if evt.Type != notify.EventChatMessageCreated {
							continue
						}
					case "notifications":
						if evt.Type != notify.EventNotificationCreated {
							continue
						}
					case "order_status":
						if evt.Type != notify.EventOrderStatusChanged {
							continue
						}
					}
					_ = writeJSON(map[string]any{
						"type": string(evt.Type),
						"data": evt.Data,
					})
				}
			}()
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// Minimal v1 protocol: accept {"type":"ping"} or raw "ping".
			if strings.TrimSpace(string(msg)) == "ping" {
				_ = writeMessage(websocket.TextMessage, []byte("pong"))
				continue
			}
			var payload map[string]any
			if err := json.Unmarshal(msg, &payload); err == nil {
				if t, _ := payload["type"].(string); t == "ping" {
					_ = writeJSON(map[string]any{"type": "pong"})
					continue
				}
			}
			_ = writeJSON(map[string]any{"type": "ack"})
		}
	}
}

func authenticateWS(cfg config.Config, r *http.Request) (userID string, roles []string, _ error) {
	raw := strings.TrimSpace(r.Header.Get("Authorization"))
	if token, ok := strings.CutPrefix(raw, "Bearer "); ok && token != "" {
		claims, err := jwtutil.Parse(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, token)
		if err != nil {
			return "", nil, err
		}
		sub, err := jwtutil.SubjectUUID(claims)
		if err != nil {
			return "", nil, err
		}
		return sub.String(), claims.Roles, nil
	}
	// Allow token via query param for simple clients: ws://.../ws/chat?token=...
	if token := strings.TrimSpace(r.URL.Query().Get("token")); token != "" {
		claims, err := jwtutil.Parse(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, token)
		if err != nil {
			return "", nil, err
		}
		sub, err := jwtutil.SubjectUUID(claims)
		if err != nil {
			return "", nil, err
		}
		return sub.String(), claims.Roles, nil
	}
	return "", nil, errors.New("missing token")
}

func waitForShutdown(srv *http.Server, logger *zap.Logger, timeout time.Duration) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("websocket_service_stopping")
	_ = srv.Shutdown(ctx)
}

func valueOrDefault(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

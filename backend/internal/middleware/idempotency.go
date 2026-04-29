package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const IdempotencyContextKey = "idempotency_key"

type cachedHTTPResponse struct {
	StatusCode  int    `json:"status_code"`
	ContentType string `json:"content_type"`
	BodyB64     string `json:"body_b64"`
}

type teeResponseWriter struct {
	gin.ResponseWriter
	status int
	buf    *bytes.Buffer
	max    int
}

func newTeeResponseWriter(w gin.ResponseWriter, max int) *teeResponseWriter {
	return &teeResponseWriter{
		ResponseWriter: w,
		buf:            bytes.NewBuffer(nil),
		max:            max,
	}
}

func (t *teeResponseWriter) WriteHeader(code int) {
	t.status = code
	t.ResponseWriter.WriteHeader(code)
}

func (t *teeResponseWriter) Write(b []byte) (int, error) {
	if t.buf != nil {
		if t.buf.Len()+len(b) > t.max {
			t.buf = nil
		} else {
			_, _ = t.buf.Write(b)
		}
	}
	return t.ResponseWriter.Write(b)
}

func (t *teeResponseWriter) statusCode() int {
	if t.status != 0 {
		return t.status
	}
	return http.StatusOK
}

// Idempotency enforces Idempotency-Key on mutating checkout/payment/order routes,
// deduplicates concurrent and retried requests using Redis, and caches successful
// responses for replay.
func Idempotency(redisClient *redis.Client, cfg config.IdempotencyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requiresIdempotency(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}

		if redisClient == nil {
			apiresponse.Error(c, http.StatusServiceUnavailable, "idempotency_unavailable", "Redis is required for idempotent mutations.")
			c.Abort()
			return
		}

		key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if key == "" {
			apiresponse.Error(c, http.StatusBadRequest, "missing_idempotency_key", "Idempotency-Key header is required for this mutation.")
			c.Abort()
			return
		}
		if len(key) > 255 {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_idempotency_key", "Idempotency-Key is too long.")
			c.Abort()
			return
		}

		c.Set(IdempotencyContextKey, key)

		ctx := c.Request.Context()
		scopeDigest := idempotencyScopeDigest(c)
		dataKey := "idem:v1:data:" + scopeDigest + ":" + key
		lockKey := "idem:v1:lock:" + scopeDigest + ":" + key

		if payload, err := redisClient.Get(ctx, dataKey).Bytes(); err == nil {
			replayCachedResponse(c, payload)
			c.Abort()
			return
		}

		acquired, err := redisClient.SetNX(ctx, lockKey, "1", cfg.LockTTL).Result()
		if err != nil {
			apiresponse.Error(c, http.StatusServiceUnavailable, "idempotency_store_error", "Unable to acquire idempotency lock.")
			c.Abort()
			return
		}

		if !acquired {
			if replayAfterWait(ctx, redisClient, c, dataKey, cfg) {
				c.Abort()
				return
			}
			apiresponse.Error(c, http.StatusConflict, "idempotency_in_progress", "The same idempotent request is still being processed. Retry later.")
			c.Abort()
			return
		}

		tee := newTeeResponseWriter(c.Writer, int(cfg.MaxBodyBytes))
		c.Writer = tee

		defer func() {
			_ = redisClient.Del(ctx, lockKey).Err()
		}()

		c.Next()

		if c.IsAborted() {
			return
		}

		status := tee.statusCode()
		if !shouldCacheIdempotentResponse(status) {
			return
		}

		if tee.buf == nil {
			return
		}

		body := tee.buf.Bytes()
		contentType := c.Writer.Header().Get("Content-Type")
		if contentType == "" {
			contentType = "application/json; charset=utf-8"
		}

		cached := cachedHTTPResponse{
			StatusCode:  status,
			ContentType: contentType,
			BodyB64:     base64.StdEncoding.EncodeToString(body),
		}

		raw, err := json.Marshal(cached)
		if err != nil {
			return
		}

		_ = redisClient.Set(ctx, dataKey, raw, cfg.TTL).Err()
	}
}

func idempotencyScope(c *gin.Context) string {
	if value, ok := c.Get(UserIDKey); ok {
		if userID, ok := value.(string); ok && userID != "" {
			return "user:" + userID
		}
	}
	return "anon:" + c.ClientIP()
}

func idempotencyScopeDigest(c *gin.Context) string {
	sum := sha256.Sum256([]byte(idempotencyScope(c)))
	return hex.EncodeToString(sum[:])
}

func requiresIdempotency(method string, path string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return false
	}

	return strings.Contains(path, "/checkout/") ||
		strings.Contains(path, "/payments/") ||
		strings.Contains(path, "/orders/") ||
		strings.Contains(path, "/shipments/") ||
		strings.Contains(path, "/shipping/")
}

func shouldCacheIdempotentResponse(status int) bool {
	if status >= 500 {
		return false
	}
	switch status {
	case http.StatusRequestTimeout, http.StatusTooManyRequests:
		return false
	default:
		return true
	}
}

func replayAfterWait(ctx context.Context, redisClient *redis.Client, c *gin.Context, dataKey string, cfg config.IdempotencyConfig) bool {
	deadline := time.Now().Add(cfg.WaitTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return false
		case <-time.After(cfg.PollInterval):
		}

		payload, err := redisClient.Get(ctx, dataKey).Bytes()
		if err == nil {
			replayCachedResponse(c, payload)
			return true
		}
	}
	return false
}

func replayCachedResponse(c *gin.Context, payload []byte) {
	var cached cachedHTTPResponse
	if err := json.Unmarshal(payload, &cached); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "idempotency_corrupt", "Stored idempotent response is invalid.")
		c.Abort()
		return
	}

	body, err := base64.StdEncoding.DecodeString(cached.BodyB64)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "idempotency_corrupt", "Stored idempotent response is invalid.")
		c.Abort()
		return
	}

	if cached.ContentType != "" {
		c.Header("Content-Type", cached.ContentType)
	}
	c.Status(cached.StatusCode)
	_, _ = c.Writer.Write(body)
	c.Abort()
}

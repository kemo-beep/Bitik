package paymentsvc

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Service) HandleAdminListWebhookEvents(c *gin.Context) {
	limit := 50
	offset := 0
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, provider, event_id, event_type, processed, created_at, processed_at
		FROM payment_webhook_events
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list webhook events.")
		return
	}
	defer rows.Close()
	items := make([]gin.H, 0)
	for rows.Next() {
		var id uuid.UUID
		var provider, eventID string
		var eventType *string
		var processed bool
		var createdAt, processedAt any
		if err := rows.Scan(&id, &provider, &eventID, &eventType, &processed, &createdAt, &processedAt); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not parse webhook events.")
			return
		}
		items = append(items, gin.H{
			"id": id, "provider": provider, "event_id": eventID, "event_type": eventType,
			"processed": processed, "created_at": createdAt, "processed_at": processedAt,
		})
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminGetWebhookEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("event_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid event id.")
		return
	}
	var (
		id        uuid.UUID
		provider  string
		external  string
		eventType *string
		payload   []byte
		processed bool
		createdAt any
		processedAt any
	)
	err = s.pool.QueryRow(c.Request.Context(), `
		SELECT id, provider, event_id, event_type, payload, processed, created_at, processed_at
		FROM payment_webhook_events
		WHERE id = $1
	`, eventID).Scan(&id, &provider, &external, &eventType, &payload, &processed, &createdAt, &processedAt)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Webhook event not found.")
		return
	}
	apiresponse.OK(c, gin.H{
		"id": id, "provider": provider, "event_id": external, "event_type": eventType,
		"payload": json.RawMessage(payload), "processed": processed, "created_at": createdAt, "processed_at": processedAt,
	})
}

func (s *Service) HandleAdminReprocessWebhookEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("event_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid event id.")
		return
	}
	updated, err := s.pay.MarkPaymentWebhookProcessed(c.Request.Context(), pgxutil.UUID(eventID))
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Webhook event not found.")
		return
	}
	apiresponse.OK(c, updated)
}


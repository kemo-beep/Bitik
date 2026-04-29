package shippingsvc

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *Service) HandleAdminListShipments(c *gin.Context) {
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
		SELECT id, order_id, seller_id, provider_id, status, tracking_number, created_at, updated_at
		FROM shipments
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list shipments.")
		return
	}
	defer rows.Close()
	items := make([]gin.H, 0)
	for rows.Next() {
		var (
			id, orderID, sellerID uuid.UUID
			providerID             *uuid.UUID
			status                 any
			trackingNumber         *string
			createdAt, updatedAt   any
		)
		if err := rows.Scan(&id, &orderID, &sellerID, &providerID, &status, &trackingNumber, &createdAt, &updatedAt); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not parse shipments.")
			return
		}
		items = append(items, gin.H{
			"id": id, "order_id": orderID, "seller_id": sellerID, "provider_id": providerID,
			"status": statusString(status), "tracking_number": trackingNumber, "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminGetShipment(c *gin.Context) {
	shipmentID, err := uuid.Parse(c.Param("shipment_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid shipment id.")
		return
	}
	var (
		id, orderID, sellerID uuid.UUID
		providerID             *uuid.UUID
		status                 any
		trackingNumber         *string
		addressSnapshot, providerMetadata []byte
		createdAt, updatedAt   any
	)
	err = s.pool.QueryRow(c.Request.Context(), `
		SELECT id, order_id, seller_id, provider_id, status, tracking_number, address_snapshot, provider_metadata, created_at, updated_at
		FROM shipments
		WHERE id = $1
	`, shipmentID).Scan(&id, &orderID, &sellerID, &providerID, &status, &trackingNumber, &addressSnapshot, &providerMetadata, &createdAt, &updatedAt)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Shipment not found.")
		return
	}
	apiresponse.OK(c, gin.H{
		"id": id, "order_id": orderID, "seller_id": sellerID, "provider_id": providerID, "status": statusString(status),
		"tracking_number": trackingNumber, "address_snapshot": json.RawMessage(addressSnapshot), "provider_metadata": json.RawMessage(providerMetadata),
		"created_at": createdAt, "updated_at": updatedAt,
	})
}

func (s *Service) HandleAdminShipmentTracking(c *gin.Context) {
	shipmentID, err := uuid.Parse(c.Param("shipment_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid shipment id.")
		return
	}
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, shipment_id, status, event_time, raw_payload, created_at
		FROM shipment_tracking_events
		WHERE shipment_id = $1
		ORDER BY event_time DESC
	`, shipmentID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list tracking events.")
		return
	}
	defer rows.Close()
	items := make([]gin.H, 0)
	for rows.Next() {
		var id, shipID uuid.UUID
		var status string
		var eventTime, createdAt any
		var rawPayload []byte
		if err := rows.Scan(&id, &shipID, &status, &eventTime, &rawPayload, &createdAt); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not parse tracking events.")
			return
		}
		items = append(items, gin.H{
			"id": id, "shipment_id": shipID, "status": status, "event_time": eventTime, "raw_payload": json.RawMessage(rawPayload), "created_at": createdAt,
		})
	}
	apiresponse.OK(c, gin.H{"items": items})
}


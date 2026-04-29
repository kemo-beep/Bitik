package shippingsvc

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/shippingsvc/providers"
	orderstore "github.com/bitik/backend/internal/store/orders"
	shippingstore "github.com/bitik/backend/internal/store/shipping"
	systemstore "github.com/bitik/backend/internal/store/system"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func text(v string) pgtype.Text {
	v = strings.TrimSpace(v)
	return pgtype.Text{String: v, Valid: v != ""}
}

func jsonObject(v any) []byte {
	b, _ := json.Marshal(v)
	if len(b) == 0 {
		return []byte(`{}`)
	}
	return b
}

func notFoundOrInternal(c *gin.Context, err error, code, message string) {
	if err == nil {
		return
	}
	if err == pgx.ErrNoRows {
		apiresponse.Error(c, http.StatusNotFound, code, message)
		return
	}
	apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load resource.")
}

func (s *Service) HandleBuyerOrderShipments(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	orderID, ok := uuidParam(c, "order_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid order id.")
		return
	}
	if _, err := s.orderQ.GetOrderForUser(c.Request.Context(), orderstore.GetOrderForUserParams{ID: orderID, UserID: pgxutil.UUID(uid)}); err != nil {
		notFoundOrInternal(c, err, "not_found", "Order not found.")
		return
	}
	items, err := s.shipQ.ListShipmentsForOrder(c.Request.Context(), orderID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list shipments.")
		return
	}
	apiresponse.OK(c, items)
}

func (s *Service) HandleBuyerOrderTracking(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	orderID, ok := uuidParam(c, "order_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid order id.")
		return
	}
	if _, err := s.orderQ.GetOrderForUser(c.Request.Context(), orderstore.GetOrderForUserParams{ID: orderID, UserID: pgxutil.UUID(uid)}); err != nil {
		notFoundOrInternal(c, err, "not_found", "Order not found.")
		return
	}
	events, err := s.shipQ.ListTrackingEventsForOrder(c.Request.Context(), orderID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list tracking events.")
		return
	}
	apiresponse.OK(c, events)
}

func (s *Service) HandleSellerOrderShipments(c *gin.Context) {
	seller, ok := sellerFromContext(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing seller context.")
		return
	}
	orderID, ok := uuidParam(c, "order_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid order id.")
		return
	}
	items, err := s.shipQ.ListShipmentsForSellerOrder(c.Request.Context(), shippingstore.ListShipmentsForSellerOrderParams{
		OrderID:  orderID,
		SellerID: seller.ID,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list shipments.")
		return
	}
	apiresponse.OK(c, items)
}

type patchShipmentRequest struct {
	ProviderID       *string        `json:"provider_id"`
	TrackingNumber   *string        `json:"tracking_number"`
	ProviderMetadata map[string]any `json:"provider_metadata"`
}

func (s *Service) HandleSellerPatchShipment(c *gin.Context) {
	seller, ok := sellerFromContext(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing seller context.")
		return
	}
	shipmentID, ok := uuidParam(c, "shipment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid shipment id.")
		return
	}
	var req patchShipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}

	var providerID pgtype.UUID
	if req.ProviderID != nil && strings.TrimSpace(*req.ProviderID) != "" {
		if id, ok := parseUUID(*req.ProviderID); ok {
			providerID = id
		} else {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid provider id.")
			return
		}
	}

	updated, err := s.shipQ.UpdateShipmentSellerFields(c.Request.Context(), shippingstore.UpdateShipmentSellerFieldsParams{
		ID:               shipmentID,
		SellerID:         seller.ID,
		ProviderID:       providerID,
		TrackingNumber:   text(ptr(req.TrackingNumber)),
		ProviderMetadata: jsonObject(req.ProviderMetadata),
	})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Shipment not found.")
		return
	}
	apiresponse.OK(c, updated)
}

func (s *Service) updateStatusForSeller(c *gin.Context, status string) {
	seller, ok := sellerFromContext(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing seller context.")
		return
	}
	uid, _ := currentUserID(c)
	shipmentID, ok := uuidParam(c, "shipment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid shipment id.")
		return
	}
	current, err := s.shipQ.GetShipmentByID(c.Request.Context(), shipmentID)
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Shipment not found.")
		return
	}
	if current.SellerID != seller.ID {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Shipment does not belong to seller.")
		return
	}
	oldStatus := strings.TrimSpace(statusString(current.Status))
	if !canTransitionShipment(oldStatus, status) {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_transition", "Invalid shipment status transition.")
		return
	}
	updated, err := s.shipQ.UpdateShipmentStatusForSeller(c.Request.Context(), shippingstore.UpdateShipmentStatusForSellerParams{
		ID:       shipmentID,
		SellerID: seller.ID,
		Status:   status,
	})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Shipment not found.")
		return
	}
	_, _ = s.shipQ.CreateShipmentTrackingEvent(c.Request.Context(), shippingstore.CreateShipmentTrackingEventParams{
		ShipmentID: updated.ID,
		Status:     status,
		EventTime:  pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
	_, _ = s.systemQ.CreateAuditLog(c.Request.Context(), systemstore.CreateAuditLogParams{
		ActorUserID: pgxutil.UUID(uid),
		Action:      "shipment.status_update.seller",
		EntityType:  text("shipment"),
		EntityID:    updated.ID,
		OldValues:   jsonObject(map[string]any{"status": oldStatus}),
		NewValues:   jsonObject(map[string]any{"status": status}),
		UserAgent:   text(c.Request.UserAgent()),
		IpAddress:   ipAddrPtr(c),
	})
	apiresponse.OK(c, updated)
}

func (s *Service) HandleSellerMarkPacked(c *gin.Context)    { s.updateStatusForSeller(c, "packed") }
func (s *Service) HandleSellerMarkShipped(c *gin.Context)   { s.updateStatusForSeller(c, "shipped") }
func (s *Service) HandleSellerMarkInTransit(c *gin.Context) { s.updateStatusForSeller(c, "in_transit") }
func (s *Service) HandleSellerMarkDelivered(c *gin.Context) { s.updateStatusForSeller(c, "delivered") }

type createLabelRequest struct {
	Format   *string        `json:"format"`
	Metadata map[string]any `json:"metadata"`
}

func (s *Service) HandleSellerCreateLabel(c *gin.Context) {
	seller, ok := sellerFromContext(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing seller context.")
		return
	}
	uid, _ := currentUserID(c)
	shipmentID, ok := uuidParam(c, "shipment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid shipment id.")
		return
	}
	var req createLabelRequest
	_ = c.ShouldBindJSON(&req)

	shipment, err := s.shipQ.GetShipmentByID(c.Request.Context(), shipmentID)
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Shipment not found.")
		return
	}
	if shipment.SellerID != seller.ID {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Shipment does not belong to seller.")
		return
	}

	if !shipment.ProviderID.Valid {
		apiresponse.Error(c, http.StatusBadRequest, "labels_not_supported", "Shipment provider does not support labels.")
		return
	}
	p, err := s.shipQ.GetShippingProviderByID(c.Request.Context(), shipment.ProviderID)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "labels_not_supported", "Shipment provider does not support labels.")
		return
	}
	if strings.TrimSpace(strings.ToLower(p.Code)) != "local-courier" {
		apiresponse.Error(c, http.StatusBadRequest, "labels_not_supported", "Shipment provider does not support labels.")
		return
	}
	adapter := providers.ByCode(p.Code)
	labelURL, md, err := adapter.CreateLabel(c.Request.Context(), shipment, p)
	if err != nil || strings.TrimSpace(labelURL) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "labels_not_supported", "Shipment provider does not support labels.")
		return
	}
	meta := req.Metadata
	if meta == nil {
		meta = map[string]any{}
	}
	for k, v := range md {
		meta[k] = v
	}

	label, err := s.shipQ.CreateShipmentLabel(c.Request.Context(), shippingstore.CreateShipmentLabelParams{
		ShipmentID: shipmentID,
		LabelUrl:   labelURL,
		Format:     text(ptr(req.Format)),
		Metadata:   jsonObject(meta),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create label.")
		return
	}
	_, _ = s.systemQ.CreateAuditLog(c.Request.Context(), systemstore.CreateAuditLogParams{
		ActorUserID: pgxutil.UUID(uid),
		Action:      "shipment.label_create.seller",
		EntityType:  text("shipment_label"),
		EntityID:    label.ID,
		OldValues:   jsonObject(nil),
		NewValues:   jsonObject(map[string]any{"shipment_id": shipmentID.String(), "label_url": labelURL, "provider_code": p.Code}),
		UserAgent:   text(c.Request.UserAgent()),
		IpAddress:   ipAddrPtr(c),
	})
	apiresponse.OK(c, label)
}

type adminCreateProviderRequest struct {
	Name     string         `json:"name" binding:"required"`
	Code     string         `json:"code" binding:"required"`
	Metadata map[string]any `json:"metadata"`
	IsActive *bool          `json:"is_active"`
}

func (s *Service) HandleAdminListProviders(c *gin.Context) {
	items, err := s.shipQ.ListActiveShippingProviders(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list providers.")
		return
	}
	apiresponse.OK(c, items)
}

func (s *Service) HandleAdminCreateProvider(c *gin.Context) {
	actor, _ := currentUserID(c)
	var req adminCreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}
	created, err := s.shipQ.CreateShippingProvider(c.Request.Context(), shippingstore.CreateShippingProviderParams{
		Name:     strings.TrimSpace(req.Name),
		Code:     strings.TrimSpace(req.Code),
		Metadata: jsonObject(req.Metadata),
		IsActive: boolPtr(req.IsActive),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create provider.")
		return
	}
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "shipping_provider_created",
		EntityType:  text("shipping_provider"),
		EntityID:    created.ID,
		Metadata:    jsonObject(map[string]any{"code": created.Code}),
		UserAgent:   text(c.Request.UserAgent()),
		IpAddress:   ipAddrPtr(c),
	})
	apiresponse.OK(c, created)
}

type adminUpdateProviderRequest struct {
	Name     *string        `json:"name"`
	Metadata map[string]any `json:"metadata"`
	IsActive *bool          `json:"is_active"`
}

func (s *Service) HandleAdminUpdateProvider(c *gin.Context) {
	actor, _ := currentUserID(c)
	providerID, ok := uuidParam(c, "provider_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid provider id.")
		return
	}
	var req adminUpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}
	updated, err := s.shipQ.UpdateShippingProvider(c.Request.Context(), shippingstore.UpdateShippingProviderParams{
		ID:       providerID,
		Name:     text(ptr(req.Name)),
		Metadata: jsonObject(req.Metadata),
		IsActive: boolPtr(req.IsActive),
	})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Provider not found.")
		return
	}
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "shipping_provider_updated",
		EntityType:  text("shipping_provider"),
		EntityID:    updated.ID,
		Metadata:    jsonObject(map[string]any{"code": updated.Code}),
		UserAgent:   text(c.Request.UserAgent()),
		IpAddress:   ipAddrPtr(c),
	})
	apiresponse.OK(c, updated)
}

type adminUpdateShipmentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (s *Service) HandleAdminUpdateShipmentStatus(c *gin.Context) {
	actor, _ := currentUserID(c)
	shipmentID, ok := uuidParam(c, "shipment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid shipment id.")
		return
	}
	var req adminUpdateShipmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}
	updated, err := s.shipQ.UpdateShipmentStatusAdmin(c.Request.Context(), shippingstore.UpdateShipmentStatusAdminParams{
		ID:     shipmentID,
		Status: strings.TrimSpace(req.Status),
	})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Shipment not found.")
		return
	}
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "shipment_status_override",
		EntityType:  text("shipment"),
		EntityID:    shipmentID,
		Metadata:    jsonObject(map[string]any{"status": req.Status}),
		UserAgent:   text(c.Request.UserAgent()),
		IpAddress:   ipAddrPtr(c),
	})
	apiresponse.OK(c, updated)
}

func ptr(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func boolPtr(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func parseUUID(raw string) (pgtype.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func canTransitionShipment(old, next string) bool {
	if old == next {
		return true
	}
	switch old {
	case "pending":
		return next == "packed" || next == "shipped"
	case "packed":
		return next == "shipped"
	case "shipped":
		return next == "in_transit" || next == "delivered"
	case "in_transit":
		return next == "delivered"
	case "delivered":
		return false
	case "failed":
		return false
	case "returned":
		return false
	default:
		return false
	}
}

func statusString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}

func ipAddrPtr(c *gin.Context) *netip.Addr {
	ipStr := strings.TrimSpace(c.ClientIP())
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil
	}
	return &ip
}

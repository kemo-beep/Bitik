package ordersvc

import (
	"context"
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleReleaseExpiredInventory(c *gin.Context) {
	var req struct {
		Limit int32 `json:"limit"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Limit <= 0 {
		req.Limit = defaultReservationBatch
	}
	count, err := s.releaseExpiredReservations(c.Request.Context(), req.Limit)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "release_failed", "Could not release expired inventory.")
		return
	}
	apiresponse.OK(c, gin.H{"released": count})
}

func (s *Service) HandleExpireCheckout(c *gin.Context) {
	var req struct {
		Limit int32 `json:"limit"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Limit <= 0 {
		req.Limit = defaultReservationBatch
	}
	count, err := s.releaseExpiredReservations(c.Request.Context(), req.Limit)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "expire_checkout_failed", "Could not expire checkout sessions.")
		return
	}
	apiresponse.OK(c, gin.H{"expired_reservations_released": count})
}

func (s *Service) releaseExpiredReservations(ctx context.Context, limit int32) (int, error) {
	reservations, err := s.queries.ListExpiredReservations(ctx, limit)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, reservation := range reservations {
		if err := s.releaseReservation(ctx, reservation); err != nil {
			return released, err
		}
		released++
	}
	return released, nil
}

func (s *Service) releaseReservation(ctx context.Context, reservation orderstore.InventoryReservation) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	inv, err := q.LockInventoryByID(ctx, reservation.InventoryItemID)
	if err != nil {
		return err
	}
	beforeAvailable := inv.QuantityAvailable
	beforeReserved := inv.QuantityReserved

	updated, err := q.UpdateInventoryReserved(ctx, orderstore.UpdateInventoryReservedParams{ID: reservation.InventoryItemID, DeltaReserved: -reservation.Quantity})
	if err != nil {
		return err
	}
	if _, err := q.MarkReservationReleased(ctx, reservation.ID); err != nil {
		return err
	}
	_, err = q.CreateInventoryMovement(ctx, orderstore.CreateInventoryMovementParams{
		InventoryItemID: reservation.InventoryItemID,
		MovementType:    "release",
		Quantity:        reservation.Quantity,
		Reason:          text("expired checkout reservation"),
		ReferenceType:   text("inventory_reservation"),
		ReferenceID:     reservation.ID,
		BeforeAvailable: pgtype.Int4{Int32: beforeAvailable, Valid: true},
		AfterAvailable:  pgtype.Int4{Int32: updated.QuantityAvailable, Valid: true},
		BeforeReserved:  pgtype.Int4{Int32: beforeReserved, Valid: true},
		AfterReserved:   pgtype.Int4{Int32: updated.QuantityReserved, Valid: true},
	})
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) HandleCancelUnpaidOrders(c *gin.Context) {
	var req struct {
		Limit            int32 `json:"limit"`
		OlderThanMinutes int32 `json:"older_than_minutes"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.Limit <= 0 {
		req.Limit = defaultReservationBatch
	}
	if req.OlderThanMinutes <= 0 {
		req.OlderThanMinutes = int32(defaultCheckoutTTL.Minutes())
	}
	cancelled, released, err := s.cancelStaleUnpaidOrders(c.Request.Context(), req.Limit, req.OlderThanMinutes)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "cancel_unpaid_failed", "Could not cancel unpaid orders.")
		return
	}
	apiresponse.OK(c, gin.H{"cancelled_orders": cancelled, "released_reservations": released})
}

func (s *Service) cancelStaleUnpaidOrders(ctx context.Context, limit, olderThanMinutes int32) (int, int, error) {
	orders, err := s.queries.ListStaleUnpaidOrders(ctx, orderstore.ListStaleUnpaidOrdersParams{Limit: limit, OlderThanMinutes: olderThanMinutes})
	if err != nil {
		return 0, 0, err
	}
	cancelled := 0
	released := 0
	for _, order := range orders {
		r, err := s.cancelOrderAndReleaseReservations(ctx, order.ID, "Cancelled unpaid order", pgtype.UUID{})
		if err != nil {
			return cancelled, released, err
		}
		cancelled++
		released += r
	}
	return cancelled, released, nil
}

func (s *Service) cancelOrderAndReleaseReservations(ctx context.Context, orderID pgtype.UUID, note string, actor pgtype.UUID) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	order, err := q.GetOrderByID(ctx, orderID)
	if err != nil {
		return 0, err
	}
	if statusString(order.Status) != "pending_payment" {
		return 0, nil
	}
	updated, err := q.UpdateOrderStatus(ctx, orderstore.UpdateOrderStatusParams{ID: orderID, Status: "cancelled"})
	if err != nil {
		return 0, err
	}
	if _, err := q.InsertOrderStatusHistory(ctx, orderstore.InsertOrderStatusHistoryParams{OrderID: orderID, OldStatus: order.Status, NewStatus: "cancelled", Note: text(note), CreatedBy: actor}); err != nil {
		return 0, err
	}
	n, err := releaseOrderReservationsWithQueries(ctx, q, updated.ID, actor, "cancel_unpaid_order")
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return n, nil
}

func (s *Service) HandleGenerateInvoices(c *gin.Context) {
	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Order id is required.")
		return
	}
	id, ok := parseUUID(req.OrderID)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_id", "Invalid order id.")
		return
	}
	order, err := s.queries.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		notFoundOrInternal(c, err, "order_not_found", "Order not found.")
		return
	}
	invoice, err := s.createInvoice(c.Request.Context(), order)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "invoice_generation_failed", "Could not generate invoice.")
		return
	}
	apiresponse.OK(c, invoiceJSON(invoice))
}

func (s *Service) createInvoice(ctx context.Context, order orderstore.Order) (orderstore.OrderInvoice, error) {
	items, err := s.queries.ListOrderItems(ctx, order.ID)
	if err != nil {
		return orderstore.OrderInvoice{}, err
	}
	payload := gin.H{"order": orderJSON(order), "items": mapSlice(items, orderItemJSON)}
	return s.queries.CreateOrGetOrderInvoice(ctx, orderstore.CreateOrGetOrderInvoiceParams{OrderID: order.ID, InvoiceNumber: invoiceNumber(order), Payload: jsonObject(payload)})
}

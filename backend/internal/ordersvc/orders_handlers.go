package ordersvc

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/notify"
	"github.com/bitik/backend/internal/pgxutil"
	orderstore "github.com/bitik/backend/internal/store/orders"
	paymentstore "github.com/bitik/backend/internal/store/payments"
	systemstore "github.com/bitik/backend/internal/store/system"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleBuyerListOrders(c *gin.Context) {
	uid, _ := currentUserID(c)
	p := pagination(c)
	orders, err := s.queries.ListUserOrders(c.Request.Context(), orderstore.ListUserOrdersParams{UserID: pgxutil.UUID(uid), Status: optionalOrderStatus(c.Query("status")), Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load orders.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(orders, orderJSON), "pagination": gin.H{"page": p.Page, "per_page": p.PerPage}})
}

func (s *Service) HandleBuyerGetOrder(c *gin.Context) {
	uid, _ := currentUserID(c)
	orderID, ok := uuidParam(c, "order_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_id", "Invalid order id.")
		return
	}
	order, err := s.queries.GetOrderForUser(c.Request.Context(), orderstore.GetOrderForUserParams{ID: orderID, UserID: pgxutil.UUID(uid)})
	if err != nil {
		notFoundOrInternal(c, err, "order_not_found", "Order not found.")
		return
	}
	s.writeOrderDetail(c, order)
}

func (s *Service) HandleBuyerOrderItems(c *gin.Context) {
	order, ok := s.buyerOrder(c)
	if !ok {
		return
	}
	items, err := s.queries.ListOrderItems(c.Request.Context(), order.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load order items.")
		return
	}
	apiresponse.OK(c, mapSlice(items, orderItemJSON))
}

func (s *Service) HandleOrderStatusHistory(c *gin.Context) {
	order, ok := s.buyerOrder(c)
	if !ok {
		return
	}
	s.writeHistory(c, order.ID)
}

func (s *Service) HandleBuyerCancelOrder(c *gin.Context) {
	s.buyerTransition(c, "cancelled", "Buyer cancelled order")
}

func (s *Service) HandleBuyerConfirmReceived(c *gin.Context) {
	uid, _ := currentUserID(c)
	order, ok := s.buyerOrder(c)
	if !ok {
		return
	}
	if statusString(order.Status) == "completed" {
		apiresponse.OK(c, orderJSON(order))
		return
	}

	shipments, err := s.queries.ListShipmentsForOrder(c.Request.Context(), order.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load shipments.")
		return
	}
	if len(shipments) == 0 {
		apiresponse.Error(c, http.StatusBadRequest, "not_delivered", "Order has no shipments.")
		return
	}
	for _, sh := range shipments {
		if statusString(sh.Status) != "delivered" {
			apiresponse.Error(c, http.StatusBadRequest, "not_delivered", "All shipments must be delivered before confirming receipt.")
			return
		}
	}

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not confirm receipt.")
		return
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	// Lock latest order row.
	locked, err := q.GetOrderForUser(ctx, orderstore.GetOrderForUserParams{ID: order.ID, UserID: pgxutil.UUID(uid)})
	if err != nil {
		notFoundOrInternal(c, err, "order_not_found", "Order not found.")
		return
	}
	if statusString(locked.Status) == "completed" {
		_ = tx.Commit(ctx)
		apiresponse.OK(c, orderJSON(locked))
		return
	}

	// Re-check shipments inside transaction.
	sh2, err := q.ListShipmentsForOrder(ctx, locked.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load shipments.")
		return
	}
	for _, sh := range sh2 {
		if statusString(sh.Status) != "delivered" {
			apiresponse.Error(c, http.StatusBadRequest, "not_delivered", "All shipments must be delivered before confirming receipt.")
			return
		}
	}

	updated, err := s.transitionOrderWithQueries(ctx, q, locked, "completed", "Buyer confirmed receipt", pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_transition", err.Error())
		return
	}

	// Settlement trigger: credit sellers into pending wallet balance (idempotent by unique (wallet, order)).
	payQ := paymentstore.New(s.pool).WithTx(tx)
	credits, err := q.ListOrderCreditsBySeller(ctx, updated.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not compute settlement.")
		return
	}
	for _, cr := range credits {
		if cr.AmountCents <= 0 {
			continue
		}
		if _, err := payQ.GetOrCreateSellerWallet(ctx, paymentstore.GetOrCreateSellerWalletParams{SellerID: cr.SellerID}); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not settle wallet.")
			return
		}
		wallet, err := payQ.LockSellerWallet(ctx, cr.SellerID)
		if err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not settle wallet.")
			return
		}
		beforePending := wallet.PendingBalanceCents
		afterPending := wallet.PendingBalanceCents + cr.AmountCents
		if _, err := payQ.CreateWalletTransaction(ctx, paymentstore.CreateWalletTransactionParams{
			SellerWalletID:     wallet.ID,
			Type:               "credit",
			AmountCents:        cr.AmountCents,
			BalanceBeforeCents: beforePending,
			BalanceAfterCents:  afterPending,
			ReferenceType:      text("order"),
			ReferenceID:        updated.ID,
			Description:        text("order settlement (pending)"),
		}); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				continue
			}
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not settle wallet.")
			return
		}
		if _, err := payQ.UpdateWalletBalances(ctx, paymentstore.UpdateWalletBalancesParams{
			ID:                  wallet.ID,
			BalanceCents:        wallet.BalanceCents,
			PendingBalanceCents: afterPending,
		}); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not settle wallet.")
			return
		}
	}

	// Audit: completion is a financial boundary (settlement trigger).
	sysQ := systemstore.New(s.pool).WithTx(tx)
	_, _ = sysQ.CreateAuditLog(ctx, systemstore.CreateAuditLogParams{
		ActorUserID: pgxutil.UUID(uid),
		Action:      "order.confirm_received.buyer",
		EntityType:  text("order"),
		EntityID:    updated.ID,
		OldValues:   jsonObject(map[string]any{"status": statusString(locked.Status)}),
		NewValues:   jsonObject(map[string]any{"status": "completed"}),
		IpAddress:   auditIP(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not confirm receipt.")
		return
	}
	apiresponse.OK(c, orderJSON(updated))
}

func auditIP(raw string) *netip.Addr {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	ip, err := netip.ParseAddr(raw)
	if err != nil {
		return nil
	}
	return &ip
}

func (s *Service) HandleBuyerDispute(c *gin.Context) {
	s.buyerTransition(c, "disputed", "Buyer opened dispute")
}

func (s *Service) buyerTransition(c *gin.Context, next, note string) {
	uid, _ := currentUserID(c)
	order, ok := s.buyerOrder(c)
	if !ok {
		return
	}
	updated, err := s.transitionOrder(c.Request.Context(), order, next, note, pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_transition", err.Error())
		return
	}
	apiresponse.OK(c, orderJSON(updated))
}

func (s *Service) HandleBuyerRequestRefund(c *gin.Context) {
	uid, _ := currentUserID(c)
	order, ok := s.buyerOrder(c)
	if !ok {
		return
	}
	var req struct {
		Reason   string         `json:"reason"`
		Metadata map[string]any `json:"metadata"`
	}
	_ = c.ShouldBindJSON(&req)
	refund, err := s.queries.CreateRefundRequest(c.Request.Context(), orderstore.CreateRefundRequestParams{OrderID: order.ID, RequestedBy: pgxutil.UUID(uid), Reason: text(req.Reason), Metadata: jsonObject(req.Metadata)})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "refund_request_failed", "Could not request refund.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, refundJSON(refund), nil)
}

func (s *Service) HandleBuyerRequestReturn(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		OrderItemID string         `json:"order_item_id" binding:"required"`
		Reason      string         `json:"reason"`
		Quantity    int32          `json:"quantity"`
		Metadata    map[string]any `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid return request.")
		return
	}
	if req.Quantity < 1 {
		req.Quantity = 1
	}
	itemID, ok := parseUUID(req.OrderItemID)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_item_id", "Invalid order item id.")
		return
	}

	// Phase 7: returns only allowed after delivery.
	row, err := s.queries.GetOrderAndSellerForOrderItemForUser(c.Request.Context(), orderstore.GetOrderAndSellerForOrderItemForUserParams{ID: itemID, UserID: pgxutil.UUID(uid)})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "return_request_failed", "Could not request return.")
		return
	}
	count, err := s.queries.CountDeliveredShipmentsForOrderSeller(c.Request.Context(), orderstore.CountDeliveredShipmentsForOrderSellerParams{
		OrderID:  row.OrderID,
		SellerID: row.SellerID,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not validate delivery.")
		return
	}
	if count == 0 {
		apiresponse.Error(c, http.StatusBadRequest, "return_not_available", "Returns are available after delivery.")
		return
	}

	ret, err := s.queries.CreateReturnRequest(c.Request.Context(), orderstore.CreateReturnRequestParams{OrderItemID: itemID, RequestedBy: pgxutil.UUID(uid), Reason: text(req.Reason), Quantity: req.Quantity, Metadata: jsonObject(req.Metadata)})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "return_request_failed", "Could not request return.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, returnJSON(ret), nil)
}

func (s *Service) HandleBuyerInvoice(c *gin.Context) {
	uid, _ := currentUserID(c)
	order, ok := s.buyerOrder(c)
	if !ok {
		return
	}
	invoice, err := s.createInvoice(c.Request.Context(), order)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not generate invoice.")
		return
	}
	if _, err := s.queries.GetOrderInvoiceForUser(c.Request.Context(), orderstore.GetOrderInvoiceForUserParams{OrderID: order.ID, UserID: pgxutil.UUID(uid)}); err != nil {
		notFoundOrInternal(c, err, "invoice_not_found", "Invoice not found.")
		return
	}
	apiresponse.OK(c, invoiceJSON(invoice))
}

func (s *Service) HandleBuyerTracking(c *gin.Context) {
	order, ok := s.buyerOrder(c)
	if !ok {
		return
	}
	shipments, err := s.queries.ListShipmentsForOrder(c.Request.Context(), order.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load shipments.")
		return
	}
	events, err := s.queries.ListTrackingEventsForOrder(c.Request.Context(), order.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load tracking events.")
		return
	}
	apiresponse.OK(c, gin.H{"shipments": mapSlice(shipments, shipmentJSON), "events": mapSlice(events, trackingEventJSON)})
}

func (s *Service) buyerOrder(c *gin.Context) (orderstore.Order, bool) {
	uid, _ := currentUserID(c)
	orderID, ok := uuidParam(c, "order_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_id", "Invalid order id.")
		return orderstore.Order{}, false
	}
	order, err := s.queries.GetOrderForUser(c.Request.Context(), orderstore.GetOrderForUserParams{ID: orderID, UserID: pgxutil.UUID(uid)})
	if err != nil {
		notFoundOrInternal(c, err, "order_not_found", "Order not found.")
		return orderstore.Order{}, false
	}
	return order, true
}

func (s *Service) HandleSellerListOrders(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	p := pagination(c)
	orders, err := s.queries.ListSellerOrders(c.Request.Context(), orderstore.ListSellerOrdersParams{SellerID: seller.ID, Status: optionalOrderStatus(c.Query("status")), Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load seller orders.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(orders, orderJSON), "pagination": gin.H{"page": p.Page, "per_page": p.PerPage}})
}

func (s *Service) HandleSellerGetOrder(c *gin.Context) {
	order, ok := s.sellerOrder(c)
	if !ok {
		return
	}
	s.writeOrderDetail(c, order)
}

func (s *Service) HandleSellerOrderItems(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	order, ok := s.sellerOrder(c)
	if !ok {
		return
	}
	items, err := s.queries.ListSellerOrderItems(c.Request.Context(), orderstore.ListSellerOrderItemsParams{OrderID: order.ID, SellerID: seller.ID})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load seller order items.")
		return
	}
	apiresponse.OK(c, mapSlice(items, orderItemJSON))
}

func (s *Service) HandleSellerAcceptOrder(c *gin.Context) {
	s.sellerTransition(c, "paid", "Seller accepted order")
}
func (s *Service) HandleSellerRejectOrder(c *gin.Context) {
	s.sellerTransition(c, "cancelled", "Seller rejected order")
}
func (s *Service) HandleSellerMarkProcessing(c *gin.Context) {
	s.sellerTransition(c, "processing", "Seller marked order processing")
}
func (s *Service) HandleSellerPackOrder(c *gin.Context) {
	s.sellerTransition(c, "processing", "Seller packed order")
}
func (s *Service) HandleSellerCancelOrder(c *gin.Context) {
	s.sellerTransition(c, "cancelled", "Seller cancelled order")
}
func (s *Service) HandleSellerRefundOrder(c *gin.Context) {
	s.sellerTransition(c, "refunded", "Seller refunded order")
}

func (s *Service) HandleSellerShipOrder(c *gin.Context) {
	uid, _ := currentUserID(c)
	seller, _ := sellerFromContext(c)
	order, ok := s.sellerOrder(c)
	if !ok {
		return
	}
	var req struct {
		TrackingNumber string `json:"tracking_number"`
	}
	_ = c.ShouldBindJSON(&req)
	shipment, err := s.queries.UpdateShipmentForSeller(c.Request.Context(), orderstore.UpdateShipmentForSellerParams{OrderID: order.ID, SellerID: seller.ID, Status: "shipped", TrackingNumber: text(req.TrackingNumber)})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "shipment_update_failed", "Could not mark shipment as shipped.")
		return
	}
	updated, err := s.transitionOrder(c.Request.Context(), order, "shipped", "Seller shipped order", pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_transition", err.Error())
		return
	}
	_ = s.queries.UpdateSellerOrderItemsStatus(c.Request.Context(), orderstore.UpdateSellerOrderItemsStatusParams{OrderID: order.ID, SellerID: seller.ID, Status: "shipped"})
	apiresponse.OK(c, gin.H{"order": orderJSON(updated), "shipment": shipmentJSON(shipment)})
}

func (s *Service) sellerTransition(c *gin.Context, next, note string) {
	uid, _ := currentUserID(c)
	seller, _ := sellerFromContext(c)
	order, ok := s.sellerOrder(c)
	if !ok {
		return
	}
	updated, err := s.transitionOrder(c.Request.Context(), order, next, note, pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_transition", err.Error())
		return
	}
	_ = s.queries.UpdateSellerOrderItemsStatus(c.Request.Context(), orderstore.UpdateSellerOrderItemsStatusParams{OrderID: order.ID, SellerID: seller.ID, Status: next})
	apiresponse.OK(c, orderJSON(updated))
}

func (s *Service) HandleSellerApproveReturn(c *gin.Context) { s.sellerReviewReturn(c, "approved") }
func (s *Service) HandleSellerRejectReturn(c *gin.Context)  { s.sellerReviewReturn(c, "rejected") }

func (s *Service) HandleSellerMarkReturnReceived(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	order, ok := s.sellerOrder(c)
	if !ok {
		return
	}
	ret, err := s.queries.MarkReturnReceivedForSeller(c.Request.Context(), orderstore.MarkReturnReceivedForSellerParams{
		OrderID:  order.ID,
		SellerID: seller.ID,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "return_receive_failed", "Could not mark return received.")
		return
	}
	apiresponse.OK(c, returnJSON(ret))
}

func (s *Service) sellerReviewReturn(c *gin.Context, status string) {
	uid, _ := currentUserID(c)
	seller, _ := sellerFromContext(c)
	order, ok := s.sellerOrder(c)
	if !ok {
		return
	}
	ret, err := s.queries.ReviewReturnForSeller(c.Request.Context(), orderstore.ReviewReturnForSellerParams{OrderID: order.ID, SellerID: seller.ID, Status: status, ReviewedBy: pgxutil.UUID(uid)})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "return_review_failed", "Could not review return.")
		return
	}
	apiresponse.OK(c, returnJSON(ret))
}

func (s *Service) sellerOrder(c *gin.Context) (orderstore.Order, bool) {
	seller, _ := sellerFromContext(c)
	orderID, ok := uuidParam(c, "order_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_id", "Invalid order id.")
		return orderstore.Order{}, false
	}
	order, err := s.queries.GetSellerOrder(c.Request.Context(), orderstore.GetSellerOrderParams{ID: orderID, SellerID: seller.ID})
	if err != nil {
		notFoundOrInternal(c, err, "order_not_found", "Order not found.")
		return orderstore.Order{}, false
	}
	return order, true
}

func (s *Service) HandleAdminListOrders(c *gin.Context) {
	p := pagination(c)
	orders, err := s.queries.ListAdminOrders(c.Request.Context(), orderstore.ListAdminOrdersParams{Status: optionalOrderStatus(c.Query("status")), Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load orders.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(orders, orderJSON), "pagination": gin.H{"page": p.Page, "per_page": p.PerPage}})
}

func (s *Service) HandleAdminGetOrder(c *gin.Context) {
	order, ok := s.adminOrder(c)
	if !ok {
		return
	}
	s.writeOrderDetail(c, order)
}

func (s *Service) HandleAdminUpdateOrder(c *gin.Context) {
	uid, _ := currentUserID(c)
	order, ok := s.adminOrder(c)
	if !ok {
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || !validOrderStatus(req.Status) {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid order status.")
		return
	}
	updated, err := s.transitionOrder(c.Request.Context(), order, req.Status, req.Note, pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_transition", err.Error())
		return
	}
	apiresponse.OK(c, orderJSON(updated))
}

func (s *Service) HandleAdminCancelOrder(c *gin.Context) {
	s.adminTransition(c, "cancelled", "Admin cancelled order")
}
func (s *Service) HandleAdminRefundOrder(c *gin.Context) {
	s.adminTransition(c, "refunded", "Admin refunded order")
}

func (s *Service) adminTransition(c *gin.Context, next, note string) {
	uid, _ := currentUserID(c)
	order, ok := s.adminOrder(c)
	if !ok {
		return
	}
	updated, err := s.transitionOrder(c.Request.Context(), order, next, note, pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_transition", err.Error())
		return
	}
	apiresponse.OK(c, orderJSON(updated))
}

func (s *Service) HandleAdminStatusHistory(c *gin.Context) {
	order, ok := s.adminOrder(c)
	if !ok {
		return
	}
	s.writeHistory(c, order.ID)
}

func (s *Service) HandleAdminPayments(c *gin.Context) {
	order, ok := s.adminOrder(c)
	if !ok {
		return
	}
	payments, err := s.queries.ListPaymentsForOrder(c.Request.Context(), order.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load payments.")
		return
	}
	apiresponse.OK(c, mapSlice(payments, paymentJSON))
}

func (s *Service) HandleAdminShipments(c *gin.Context) {
	order, ok := s.adminOrder(c)
	if !ok {
		return
	}
	shipments, err := s.queries.ListShipmentsForOrder(c.Request.Context(), order.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load shipments.")
		return
	}
	apiresponse.OK(c, mapSlice(shipments, shipmentJSON))
}

func (s *Service) adminOrder(c *gin.Context) (orderstore.Order, bool) {
	orderID, ok := uuidParam(c, "order_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_id", "Invalid order id.")
		return orderstore.Order{}, false
	}
	order, err := s.queries.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		notFoundOrInternal(c, err, "order_not_found", "Order not found.")
		return orderstore.Order{}, false
	}
	return order, true
}

func (s *Service) writeOrderDetail(c *gin.Context, order orderstore.Order) {
	items, err := s.queries.ListOrderItems(c.Request.Context(), order.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load order items.")
		return
	}
	shipments, err := s.queries.ListShipmentsForOrder(c.Request.Context(), order.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load shipments.")
		return
	}
	apiresponse.OK(c, orderDetailJSON(order, items, shipments))
}

func (s *Service) writeHistory(c *gin.Context, orderID pgtype.UUID) {
	history, err := s.queries.ListOrderStatusHistory(c.Request.Context(), orderID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load status history.")
		return
	}
	apiresponse.OK(c, mapSlice(history, historyJSON))
}

func (s *Service) transitionOrder(ctx context.Context, order orderstore.Order, next, note string, actor pgtype.UUID) (orderstore.Order, error) {
	if !canTransition(statusString(order.Status), next) {
		return orderstore.Order{}, errInvalidTransition
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return orderstore.Order{}, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)

	if statusString(order.Status) == "pending_payment" && next == "paid" {
		if _, err := consumeOrderReservationsWithQueries(ctx, q, order.ID, actor); err != nil {
			return orderstore.Order{}, err
		}
	}
	if next == "cancelled" || next == "refunded" {
		if _, err := releaseOrderReservationsWithQueries(ctx, q, order.ID, actor, "order_status_transition"); err != nil {
			return orderstore.Order{}, err
		}
	}

	updated, err := q.UpdateOrderStatus(ctx, orderstore.UpdateOrderStatusParams{ID: order.ID, Status: next})
	if err != nil {
		return orderstore.Order{}, err
	}
	if _, err := q.InsertOrderStatusHistory(ctx, orderstore.InsertOrderStatusHistoryParams{OrderID: order.ID, OldStatus: order.Status, NewStatus: next, Note: text(note), CreatedBy: actor}); err != nil {
		return orderstore.Order{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return orderstore.Order{}, err
	}
	if s.pub != nil {
		if uid, ok := pgxutil.ToUUID(order.UserID); ok {
			s.pub.Publish(ctx, notify.Event{
				Type:   notify.EventOrderStatusChanged,
				UserID: uid.String(),
				Data: map[string]any{
					"order_id":   uuidString(order.ID),
					"old_status": statusString(order.Status),
					"new_status": next,
					"note":       note,
				},
			})
		}
	}
	return updated, nil
}

func consumeOrderReservationsWithQueries(ctx context.Context, q *orderstore.Queries, orderID, actor pgtype.UUID) (int, error) {
	reservations, err := q.ListReservationsForOrder(ctx, orderID)
	if err != nil {
		return 0, err
	}
	consumed := 0
	for _, res := range reservations {
		inv, err := q.LockInventoryByID(ctx, res.InventoryItemID)
		if err != nil {
			return consumed, err
		}
		beforeAvailable := inv.QuantityAvailable
		beforeReserved := inv.QuantityReserved
		updated, err := q.ConsumeInventoryReservation(ctx, orderstore.ConsumeInventoryReservationParams{ID: inv.ID, Quantity: res.Quantity})
		if err != nil {
			return consumed, err
		}
		if _, err := q.MarkReservationReleased(ctx, res.ID); err != nil {
			return consumed, err
		}
		_, err = q.CreateInventoryMovement(ctx, orderstore.CreateInventoryMovementParams{
			InventoryItemID: inv.ID,
			MovementType:    "stock_out",
			Quantity:        res.Quantity,
			Reason:          text("order paid"),
			ReferenceType:   text("order"),
			ReferenceID:     orderID,
			ActorUserID:     actor,
			BeforeAvailable: pgtype.Int4{Int32: beforeAvailable, Valid: true},
			AfterAvailable:  pgtype.Int4{Int32: updated.QuantityAvailable, Valid: true},
			BeforeReserved:  pgtype.Int4{Int32: beforeReserved, Valid: true},
			AfterReserved:   pgtype.Int4{Int32: updated.QuantityReserved, Valid: true},
		})
		if err != nil {
			return consumed, err
		}
		consumed++
	}
	return consumed, nil
}

func releaseOrderReservationsWithQueries(ctx context.Context, q *orderstore.Queries, orderID, actor pgtype.UUID, reason string) (int, error) {
	reservations, err := q.ListReservationsForOrder(ctx, orderID)
	if err != nil {
		return 0, err
	}
	released := 0
	for _, res := range reservations {
		inv, err := q.LockInventoryByID(ctx, res.InventoryItemID)
		if err != nil {
			return released, err
		}
		beforeAvailable := inv.QuantityAvailable
		beforeReserved := inv.QuantityReserved
		updated, err := q.UpdateInventoryReserved(ctx, orderstore.UpdateInventoryReservedParams{ID: inv.ID, DeltaReserved: -res.Quantity})
		if err != nil {
			return released, err
		}
		if _, err := q.MarkReservationReleased(ctx, res.ID); err != nil {
			return released, err
		}
		_, err = q.CreateInventoryMovement(ctx, orderstore.CreateInventoryMovementParams{
			InventoryItemID: inv.ID,
			MovementType:    "release",
			Quantity:        res.Quantity,
			Reason:          text(reason),
			ReferenceType:   text("order"),
			ReferenceID:     orderID,
			ActorUserID:     actor,
			BeforeAvailable: pgtype.Int4{Int32: beforeAvailable, Valid: true},
			AfterAvailable:  pgtype.Int4{Int32: updated.QuantityAvailable, Valid: true},
			BeforeReserved:  pgtype.Int4{Int32: beforeReserved, Valid: true},
			AfterReserved:   pgtype.Int4{Int32: updated.QuantityReserved, Valid: true},
		})
		if err != nil {
			return released, err
		}
		released++
	}
	return released, nil
}

var errInvalidTransition = errString("order status transition is not allowed")

type errString string

func (e errString) Error() string { return string(e) }

func validOrderStatus(status string) bool {
	switch status {
	case "pending_payment", "paid", "processing", "shipped", "delivered", "completed", "cancelled", "refunded", "disputed":
		return true
	default:
		return false
	}
}

func optionalOrderStatus(status string) any {
	if !validOrderStatus(status) {
		return nil
	}
	return status
}

func canTransition(current, next string) bool {
	if current == next || !validOrderStatus(next) {
		return false
	}
	if current == "cancelled" || current == "completed" || current == "refunded" {
		return false
	}
	switch next {
	case "cancelled":
		return current == "pending_payment" || current == "paid" || current == "processing"
	case "paid":
		return current == "pending_payment"
	case "processing":
		return current == "paid" || current == "pending_payment"
	case "shipped":
		return current == "processing" || current == "paid"
	case "delivered":
		return current == "shipped"
	case "completed":
		return current == "delivered" || current == "shipped"
	case "refunded":
		return current == "paid" || current == "processing" || current == "shipped" || current == "delivered" || current == "disputed"
	case "disputed":
		return current != "pending_payment"
	default:
		return false
	}
}

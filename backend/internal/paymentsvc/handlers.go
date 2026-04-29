package paymentsvc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/queue"
	orderstore "github.com/bitik/backend/internal/store/orders"
	paymentstore "github.com/bitik/backend/internal/store/payments"
	systemstore "github.com/bitik/backend/internal/store/system"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type createPaymentMethodRequest struct {
	Provider       string         `json:"provider" binding:"required"`
	Type           string         `json:"type" binding:"required"`
	DisplayName    *string        `json:"display_name"`
	TokenReference *string        `json:"token_reference"`
	IsDefault      bool           `json:"is_default"`
	Metadata       map[string]any `json:"metadata"`
}

func (s *Service) HandleListPaymentMethods(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	methods, err := s.pay.ListPaymentMethodsForUser(c.Request.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list payment methods.")
		return
	}
	apiresponse.OK(c, methods)
}

func (s *Service) HandleCreatePaymentMethod(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	var req createPaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create payment method.")
		return
	}
	defer tx.Rollback(ctx)
	q := s.pay.WithTx(tx)

	if req.IsDefault {
		if err := q.ClearDefaultPaymentMethods(ctx, pgtype.UUID{Bytes: userID, Valid: true}); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create payment method.")
			return
		}
	}

	method, err := q.CreatePaymentMethod(ctx, paymentstore.CreatePaymentMethodParams{
		UserID:         pgtype.UUID{Bytes: userID, Valid: true},
		Provider:       strings.TrimSpace(req.Provider),
		Type:           strings.TrimSpace(req.Type),
		DisplayName:    text(ptrString(req.DisplayName)),
		TokenReference: text(ptrString(req.TokenReference)),
		IsDefault:      req.IsDefault,
		Metadata:       jsonObject(req.Metadata),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create payment method.")
		return
	}
	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create payment method.")
		return
	}
	apiresponse.OK(c, method)
}

func (s *Service) HandleDeletePaymentMethod(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	methodID, ok := uuidParam(c, "method_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment method id.")
		return
	}
	if err := s.pay.DeletePaymentMethod(c.Request.Context(), paymentstore.DeletePaymentMethodParams{ID: methodID, UserID: pgtype.UUID{Bytes: userID, Valid: true}}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete payment method.")
		return
	}
	apiresponse.OK(c, gin.H{"deleted": true})
}

func (s *Service) HandleSetDefaultPaymentMethod(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	methodID, ok := uuidParam(c, "method_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment method id.")
		return
	}

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update payment method.")
		return
	}
	defer tx.Rollback(ctx)
	q := s.pay.WithTx(tx)
	if err := q.ClearDefaultPaymentMethods(ctx, pgtype.UUID{Bytes: userID, Valid: true}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update payment method.")
		return
	}
	method, err := q.SetDefaultPaymentMethod(ctx, paymentstore.SetDefaultPaymentMethodParams{ID: methodID, UserID: pgtype.UUID{Bytes: userID, Valid: true}})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Payment method not found.")
		return
	}
	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update payment method.")
		return
	}
	apiresponse.OK(c, method)
}

type createIntentRequest struct {
	OrderID         string `json:"order_id" binding:"required"`
	Provider        string `json:"provider" binding:"required"`
	PaymentMethodID string `json:"payment_method_id"`
	Currency        string `json:"currency"`
}

func (s *Service) HandleCreateIntent(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	var req createIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}
	u, parseErr := uuid.Parse(strings.TrimSpace(req.OrderID))
	if parseErr != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid order id.")
		return
	}
	orderID := pgtype.UUID{Bytes: u, Valid: true}

	idem := idempotencyKey(c)
	if strings.TrimSpace(idem) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "missing_idempotency_key", "Idempotency-Key header is required.")
		return
	}

	ctx := c.Request.Context()

	// If already created with this idempotency key, return it.
	existing, err := s.pay.GetPaymentByOrderAndIdempotencyKey(ctx, paymentstore.GetPaymentByOrderAndIdempotencyKeyParams{
		OrderID:        orderID,
		IdempotencyKey: text(idem),
	})
	if err == nil {
		apiresponse.OK(c, existing)
		return
	}
	if err != nil && err != pgx.ErrNoRows {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create payment intent.")
		return
	}

	// Validate order ownership.
	order, err := s.ord.GetOrderForUser(ctx, orderstore.GetOrderForUserParams{ID: orderID, UserID: pgtype.UUID{Bytes: userID, Valid: true}})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Order not found.")
		return
	}
	if statusString(order.Status) != "pending_payment" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_order_status", "Order is not pending payment.")
		return
	}

	provider := strings.TrimSpace(req.Provider)
	adapter, err := s.provider(provider)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_provider", "Unsupported payment provider.")
		return
	}
	if err := adapter.ValidateCreateIntent(order); err != nil {
		switch err {
		case errPODNotEligible:
			apiresponse.Error(c, http.StatusBadRequest, "pod_not_eligible", "Order is not eligible for pay on delivery.")
		case errOrderNotPendingPay:
			apiresponse.Error(c, http.StatusBadRequest, "invalid_order_status", "Order is not pending payment.")
		default:
			apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Payment intent cannot be created.")
		}
		return
	}

	var pmID pgtype.UUID
	if strings.TrimSpace(req.PaymentMethodID) != "" {
		u, parseErr := uuid.Parse(strings.TrimSpace(req.PaymentMethodID))
		if parseErr != nil {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment_method_id.")
			return
		}
		pmID = pgtype.UUID{Bytes: u, Valid: true}
	}

	payment, err := s.pay.CreatePayment(ctx, paymentstore.CreatePaymentParams{
		OrderID:         orderID,
		PaymentMethodID: pmID,
		Provider:        provider,
		Status:          nil,
		AmountCents:     order.TotalCents,
		Currency:        text(req.Currency),
		IdempotencyKey:  text(idem),
		Metadata:        jsonObject(map[string]any{"provider": provider}),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			existing, getErr := s.pay.GetPaymentByOrderAndIdempotencyKey(ctx, paymentstore.GetPaymentByOrderAndIdempotencyKeyParams{OrderID: orderID, IdempotencyKey: text(idem)})
			if getErr == nil {
				apiresponse.OK(c, existing)
				return
			}
		}
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create payment intent.")
		return
	}

	if provider == "wave_manual" {
		instructions := s.waveInstructions(ctx)
		apiresponse.OK(c, gin.H{"payment": payment, "instructions": instructions})
		return
	}
	apiresponse.OK(c, payment)
}

type confirmPaymentRequest struct {
	PaymentID   string         `json:"payment_id" binding:"required"`
	Reference   *string        `json:"reference"`
	SenderPhone *string        `json:"sender_phone"`
	Metadata    map[string]any `json:"metadata"`
}

func (s *Service) HandleConfirmPayment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	var req confirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}
	u, err := uuid.Parse(strings.TrimSpace(req.PaymentID))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment id.")
		return
	}
	ctx := c.Request.Context()
	paymentID := pgtype.UUID{Bytes: u, Valid: true}
	payment, err := s.pay.GetPaymentForUser(ctx, paymentstore.GetPaymentForUserParams{ID: paymentID, UserID: pgtype.UUID{Bytes: userID, Valid: true}})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Payment not found.")
		return
	}
	if payment.Provider == "wave_manual" {
		patch := map[string]any{
			"wave_confirm": map[string]any{
				"reference":    strings.TrimSpace(ptrString(req.Reference)),
				"sender_phone": strings.TrimSpace(ptrString(req.SenderPhone)),
				"metadata":     req.Metadata,
				"confirmed_at": time.Now().UTC().Format(time.RFC3339),
			},
		}
		updated, err := s.pay.UpdatePaymentMetadataMerge(ctx, paymentstore.UpdatePaymentMetadataMergeParams{
			ID:            paymentID,
			MetadataPatch: jsonObject(patch),
		})
		if err == nil {
			payment = updated
		}
	}
	apiresponse.OK(c, payment)
}

func (s *Service) HandleGetPayment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	paymentID, ok := uuidParam(c, "payment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment id.")
		return
	}
	payment, err := s.pay.GetPaymentForUser(c.Request.Context(), paymentstore.GetPaymentForUserParams{ID: paymentID, UserID: pgtype.UUID{Bytes: userID, Valid: true}})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Payment not found.")
		return
	}
	apiresponse.OK(c, payment)
}

func (s *Service) HandleRetryPayment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	paymentID, ok := uuidParam(c, "payment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment id.")
		return
	}
	ctx := c.Request.Context()
	payment, err := s.pay.GetPaymentForUser(ctx, paymentstore.GetPaymentForUserParams{ID: paymentID, UserID: pgtype.UUID{Bytes: userID, Valid: true}})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Payment not found.")
		return
	}
	if payment.Status == "paid" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_status", "Payment is already paid.")
		return
	}
	idem := idempotencyKey(c)
	if strings.TrimSpace(idem) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "missing_idempotency_key", "Idempotency-Key header is required.")
		return
	}
	existing, err := s.pay.GetPaymentByOrderAndIdempotencyKey(ctx, paymentstore.GetPaymentByOrderAndIdempotencyKeyParams{
		OrderID:        payment.OrderID,
		IdempotencyKey: text(idem),
	})
	if err == nil {
		apiresponse.OK(c, existing)
		return
	}
	if err != nil && err != pgx.ErrNoRows {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not retry payment.")
		return
	}

	order, err := s.ord.GetOrderForUser(ctx, orderstore.GetOrderForUserParams{ID: payment.OrderID, UserID: pgtype.UUID{Bytes: userID, Valid: true}})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Order not found.")
		return
	}
	adapter, err := s.provider(payment.Provider)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_provider", "Unsupported payment provider.")
		return
	}
	if err := adapter.ValidateCreateIntent(order); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Payment cannot be retried for this order.")
		return
	}

	newPayment, err := s.pay.CreatePayment(ctx, paymentstore.CreatePaymentParams{
		OrderID:         payment.OrderID,
		PaymentMethodID: payment.PaymentMethodID,
		Provider:        payment.Provider,
		Status:          nil,
		AmountCents:     payment.AmountCents,
		Currency:        text(payment.Currency),
		IdempotencyKey:  text(idem),
		Metadata:        payment.Metadata,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not retry payment.")
		return
	}
	apiresponse.OK(c, newPayment)
}

func (s *Service) HandleCancelPayment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	paymentID, ok := uuidParam(c, "payment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment id.")
		return
	}
	payment, err := s.pay.GetPaymentForUser(c.Request.Context(), paymentstore.GetPaymentForUserParams{ID: paymentID, UserID: pgtype.UUID{Bytes: userID, Valid: true}})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Payment not found.")
		return
	}
	if payment.Status == "paid" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_status", "Payment is already paid.")
		return
	}
	updated, err := s.pay.UpdatePaymentStatus(c.Request.Context(), paymentstore.UpdatePaymentStatusParams{
		ID:     paymentID,
		Status: "cancelled",
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not cancel payment.")
		return
	}
	apiresponse.OK(c, updated)
}

type listPendingWaveRequest struct {
	Page    int32 `form:"page"`
	PerPage int32 `form:"per_page"`
}

func (s *Service) HandleListPendingWave(c *gin.Context) {
	var q listPendingWaveRequest
	_ = c.ShouldBindQuery(&q)
	limit := int32(50)
	offset := int32(0)
	if q.PerPage > 0 && q.PerPage <= 100 {
		limit = q.PerPage
	}
	if q.Page > 1 {
		offset = (q.Page - 1) * limit
	}
	items, err := s.pay.ListPendingWaveManualPayments(c.Request.Context(), paymentstore.ListPendingWaveManualPaymentsParams{Limit: limit, Offset: offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list pending payments.")
		return
	}
	apiresponse.OK(c, items)
}

type waveDecisionRequest struct {
	Reference   string  `json:"reference" binding:"required"`
	SenderPhone *string `json:"sender_phone"`
	Note        *string `json:"note"`
	Currency    *string `json:"currency"`
}

func (s *Service) HandleApproveWave(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	paymentID, ok := uuidParam(c, "payment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment id.")
		return
	}
	var req waveDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}
	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not approve payment.")
		return
	}
	defer tx.Rollback(ctx)
	pq := s.pay.WithTx(tx)
	sq := s.sys.WithTx(tx)

	payment, err := pq.GetPaymentByID(ctx, paymentID)
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Payment not found.")
		return
	}
	if payment.Provider != "wave_manual" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_provider", "Payment is not wave_manual.")
		return
	}
	if payment.Status == "paid" {
		apiresponse.OK(c, payment)
		return
	}
	if _, err := pq.GetManualWaveDecisionForPayment(ctx, paymentID); err == nil {
		apiresponse.Error(c, http.StatusConflict, "already_decided", "Wave payment already has a decision.")
		return
	} else if err != nil && err != pgx.ErrNoRows {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not approve payment.")
		return
	}

	if _, err := pq.CreateManualWaveApproval(ctx, paymentstore.CreateManualWaveApprovalParams{
		PaymentID:   paymentID,
		Reference:   strings.TrimSpace(req.Reference),
		SenderPhone: text(ptrString(req.SenderPhone)),
		AmountCents: payment.AmountCents,
		Currency:    text(valueOrDefault(ptrString(req.Currency), payment.Currency)),
		ApprovedBy:  pgtype.UUID{Bytes: actor, Valid: true},
		Note:        text(ptrString(req.Note)),
	}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not approve payment.")
		return
	}

	updated, err := pq.UpdatePaymentStatus(ctx, paymentstore.UpdatePaymentStatusParams{ID: paymentID, Status: "paid"})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not approve payment.")
		return
	}

	if _, err := s.orders.TransitionOrder(ctx, payment.OrderID, "paid", "payment approved: wave_manual", pgtype.UUID{Bytes: actor, Valid: true}); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "order_transition_failed", "Could not transition order to paid.")
		return
	}

	_, _ = sq.CreateAuditLog(ctx, systemstore.CreateAuditLogParams{
		ActorUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "wave_manual_approved",
		EntityType:  text("payment"),
		EntityID:    paymentID,
		OldValues:   jsonObject(map[string]any{"status": payment.Status}),
		NewValues:   jsonObject(map[string]any{"status": "paid"}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})
	_, _ = sq.CreateAdminActivityLog(ctx, systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "wave_manual_approved",
		EntityType:  text("payment"),
		EntityID:    paymentID,
		Metadata:    jsonObject(map[string]any{"reference": strings.TrimSpace(req.Reference)}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not approve payment.")
		return
	}
	s.enqueuePostPaymentJobs(c.Request.Context(), payment, updated)
	apiresponse.OK(c, updated)
}

func (s *Service) HandleRejectWave(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	paymentID, ok := uuidParam(c, "payment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment id.")
		return
	}
	var req waveDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reject payment.")
		return
	}
	defer tx.Rollback(ctx)
	pq := s.pay.WithTx(tx)
	sq := s.sys.WithTx(tx)

	payment, err := pq.GetPaymentByID(ctx, paymentID)
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Payment not found.")
		return
	}
	if payment.Provider != "wave_manual" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_provider", "Payment is not wave_manual.")
		return
	}
	if payment.Status == "paid" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_status", "Payment is already paid.")
		return
	}
	if _, err := pq.GetManualWaveDecisionForPayment(ctx, paymentID); err == nil {
		apiresponse.Error(c, http.StatusConflict, "already_decided", "Wave payment already has a decision.")
		return
	} else if err != nil && err != pgx.ErrNoRows {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reject payment.")
		return
	}

	if _, err := pq.CreateManualWaveRejection(ctx, paymentstore.CreateManualWaveRejectionParams{
		PaymentID:   paymentID,
		Reference:   strings.TrimSpace(req.Reference),
		SenderPhone: text(ptrString(req.SenderPhone)),
		AmountCents: payment.AmountCents,
		Currency:    text(valueOrDefault(ptrString(req.Currency), payment.Currency)),
		RejectedBy:  pgtype.UUID{Bytes: actor, Valid: true},
		Note:        text(ptrString(req.Note)),
	}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reject payment.")
		return
	}

	updated, err := pq.UpdatePaymentStatus(ctx, paymentstore.UpdatePaymentStatusParams{ID: paymentID, Status: "failed"})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reject payment.")
		return
	}

	_, _ = sq.CreateAuditLog(ctx, systemstore.CreateAuditLogParams{
		ActorUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "wave_manual_rejected",
		EntityType:  text("payment"),
		EntityID:    paymentID,
		OldValues:   jsonObject(map[string]any{"status": payment.Status}),
		NewValues:   jsonObject(map[string]any{"status": "failed"}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})
	_, _ = sq.CreateAdminActivityLog(ctx, systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "wave_manual_rejected",
		EntityType:  text("payment"),
		EntityID:    paymentID,
		Metadata:    jsonObject(map[string]any{"reference": strings.TrimSpace(req.Reference)}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reject payment.")
		return
	}
	s.enqueuePostPaymentJobs(c.Request.Context(), payment, updated)
	apiresponse.OK(c, updated)
}

func (s *Service) enqueuePostPaymentJobs(ctx context.Context, before, after paymentstore.Payment) {
	if s.queue == nil || after.Status != "paid" {
		return
	}
	orderID := uuidText(after.OrderID)
	paymentID := uuidText(after.ID)
	_ = s.queue.PublishJob(ctx, queue.JobGenerateInvoice, "invoice:"+orderID, gin.H{"order_id": orderID})
	_ = s.queue.PublishJob(ctx, queue.JobSettleSellerWallets, "settle_wallets:"+paymentID, gin.H{"payment_id": paymentID})
}

func uuidText(id pgtype.UUID) string {
	if u, ok := pgxutil.ToUUID(id); ok {
		return u.String()
	}
	return ""
}

type capturePODRequest struct {
	Note *string `json:"note"`
}

func (s *Service) HandleCapturePOD(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	paymentID, ok := uuidParam(c, "payment_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment id.")
		return
	}
	var req capturePODRequest
	_ = c.ShouldBindJSON(&req)

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not capture POD.")
		return
	}
	defer tx.Rollback(ctx)
	pq := s.pay.WithTx(tx)
	sq := s.sys.WithTx(tx)

	payment, err := pq.GetPaymentByID(ctx, paymentID)
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Payment not found.")
		return
	}
	if payment.Provider != "pod" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_provider", "Payment is not POD.")
		return
	}
	if payment.Status == "paid" {
		apiresponse.OK(c, payment)
		return
	}

	adapter, err := s.provider("pod")
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_provider", "Unsupported payment provider.")
		return
	}
	actorIsSeller := hasRole(c, "seller") && !hasRole(c, "admin", "ops_payments")
	sellerID := pgtype.UUID{}
	if actorIsSeller {
		if sid, ok := s.sellerIDForUser(c); ok {
			sellerID = sid
		} else {
			apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
			return
		}
	}
	if err := adapter.ValidateCapture(ctx, s, payment.OrderID, sellerID, actorIsSeller); err != nil {
		switch err {
		case errPODNotDelivered:
			apiresponse.Error(c, http.StatusBadRequest, "pod_not_delivered", "POD can only be captured after delivery.")
		default:
			apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "POD capture is not allowed.")
		}
		return
	}

	if _, err := pq.CreatePODCapture(ctx, paymentstore.CreatePODCaptureParams{
		PaymentID:   paymentID,
		CapturedBy:  pgtype.UUID{Bytes: actor, Valid: true},
		AmountCents: payment.AmountCents,
		Currency:    text(payment.Currency),
		Note:        text(ptrString(req.Note)),
	}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not capture POD.")
		return
	}

	updated, err := pq.UpdatePaymentStatus(ctx, paymentstore.UpdatePaymentStatusParams{ID: paymentID, Status: "paid"})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not capture POD.")
		return
	}
	if _, err := s.orders.TransitionOrder(ctx, payment.OrderID, "paid", "payment captured: pod", pgtype.UUID{Bytes: actor, Valid: true}); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "order_transition_failed", "Could not transition order to paid.")
		return
	}

	_, _ = sq.CreateAuditLog(ctx, systemstore.CreateAuditLogParams{
		ActorUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "pod_captured",
		EntityType:  text("payment"),
		EntityID:    paymentID,
		OldValues:   jsonObject(map[string]any{"status": payment.Status}),
		NewValues:   jsonObject(map[string]any{"status": "paid"}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})
	_, _ = sq.CreateAdminActivityLog(ctx, systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "pod_captured",
		EntityType:  text("payment"),
		EntityID:    paymentID,
		Metadata:    jsonObject(map[string]any{"note": ptrString(req.Note)}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not capture POD.")
		return
	}
	apiresponse.OK(c, updated)
}

type reviewRefundRequest struct {
	Status string  `json:"status" binding:"required"` // approved or rejected
	Note   *string `json:"note"`
}

func (s *Service) HandleReviewRefund(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	refundID, ok := uuidParam(c, "refund_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid refund id.")
		return
	}
	var req reviewRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}
	status := strings.TrimSpace(req.Status)
	if status != "approved" && status != "rejected" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_status", "Status must be approved or rejected.")
		return
	}

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not review refund.")
		return
	}
	defer tx.Rollback(ctx)
	pq := s.pay.WithTx(tx)
	sq := s.sys.WithTx(tx)

	refund, err := pq.ReviewRefund(ctx, paymentstore.ReviewRefundParams{ID: refundID, Status: status, ReviewedBy: pgtype.UUID{Bytes: actor, Valid: true}})
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Refund not found or already reviewed.")
		return
	}

	if status == "approved" {
		if refund.PaymentID.Valid {
			payment, err := pq.GetPaymentByID(ctx, refund.PaymentID)
			if err != nil {
				apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load payment for refund.")
				return
			}
			if refund.AmountCents > payment.AmountCents {
				apiresponse.Error(c, http.StatusBadRequest, "invalid_amount", "Refund amount exceeds payment amount.")
				return
			}
		}
	}

	_, _ = sq.CreateAuditLog(ctx, systemstore.CreateAuditLogParams{
		ActorUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "refund_reviewed",
		EntityType:  text("refund"),
		EntityID:    refundID,
		OldValues:   jsonObject(map[string]any{"note": ptrString(req.Note)}),
		NewValues:   jsonObject(map[string]any{"status": status}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})
	_, _ = sq.CreateAdminActivityLog(ctx, systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "refund_reviewed",
		EntityType:  text("refund"),
		EntityID:    refundID,
		Metadata:    jsonObject(map[string]any{"status": status, "note": ptrString(req.Note)}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not review refund.")
		return
	}
	apiresponse.OK(c, refund)
}

func (s *Service) HandleProcessRefund(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Not authenticated.")
		return
	}
	refundID, ok := uuidParam(c, "refund_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid refund id.")
		return
	}

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not process refund.")
		return
	}
	defer tx.Rollback(ctx)
	pq := s.pay.WithTx(tx)
	sq := s.sys.WithTx(tx)

	refund, err := pq.MarkRefundProcessed(ctx, refundID)
	if err != nil {
		notFoundOrInternal(c, err, "not_found", "Refund not found or not approved.")
		return
	}

	if _, err := s.orders.TransitionOrder(ctx, refund.OrderID, "refunded", "refund processed", pgtype.UUID{Bytes: actor, Valid: true}); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "order_transition_failed", "Could not transition order to refunded.")
		return
	}
	if refund.PaymentID.Valid {
		_, _ = pq.UpdatePaymentStatus(ctx, paymentstore.UpdatePaymentStatusParams{ID: refund.PaymentID, Status: "refunded"})
	}

	_, _ = sq.CreateAuditLog(ctx, systemstore.CreateAuditLogParams{
		ActorUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "refund_processed",
		EntityType:  text("refund"),
		EntityID:    refundID,
		OldValues:   jsonObject(map[string]any{}),
		NewValues:   jsonObject(map[string]any{"status": "refunded"}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})
	_, _ = sq.CreateAdminActivityLog(ctx, systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgtype.UUID{Bytes: actor, Valid: true},
		Action:      "refund_processed",
		EntityType:  text("refund"),
		EntityID:    refundID,
		Metadata:    jsonObject(map[string]any{"order_id": refund.OrderID}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not process refund.")
		return
	}
	apiresponse.OK(c, refund)
}

func (s *Service) HandleExpirePendingPayments(c *gin.Context) {
	apiresponse.Error(c, http.StatusNotImplemented, "not_implemented", "Pending payment expiration is not implemented yet.")
}

func (s *Service) waveInstructions(ctx context.Context) string {
	setting, err := s.sys.GetPlatformSetting(ctx, "wave_manual_instructions")
	if err == nil {
		raw := strings.TrimSpace(string(setting.Value))
		if raw != "" {
			var decoded string
			if jsonErr := json.Unmarshal(setting.Value, &decoded); jsonErr == nil && strings.TrimSpace(decoded) != "" {
				return strings.TrimSpace(decoded)
			}
			return raw
		}
	}
	return "Send the exact amount via Wave and include your order reference. Then wait for approval."
}

func ptrString(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

func valueOrDefault(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return strings.TrimSpace(fallback)
}

func (s *Service) HandleWebhook(c *gin.Context) {
	provider := strings.TrimSpace(c.Param("provider"))
	if provider == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Missing provider.")
		return
	}
	eventID := strings.TrimSpace(c.GetHeader("X-Event-Id"))
	if eventID == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Missing X-Event-Id header.")
		return
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 2<<20))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not read body.")
		return
	}
	if err := s.verifyWebhookSignature(provider, c.GetHeader("X-Signature"), body); err != nil {
		apiresponse.Error(c, http.StatusUnauthorized, "invalid_signature", "Webhook signature verification failed.")
		return
	}
	row, err := s.pay.InsertPaymentWebhookEvent(c.Request.Context(), paymentstore.InsertPaymentWebhookEventParams{
		Provider:  provider,
		EventID:   eventID,
		EventType: text(strings.TrimSpace(c.GetHeader("X-Event-Type"))),
		Payload:   body,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not store webhook event.")
		return
	}
	apiresponse.OK(c, row)
}

func (s *Service) verifyWebhookSignature(provider string, signatureHeader string, body []byte) error {
	secret := strings.TrimSpace(s.cfg.Payments.WebhookSecret)
	if secret == "" {
		// Allow unsigned webhooks only when secret is not configured yet.
		return nil
	}
	signature := strings.TrimSpace(signatureHeader)
	if signature == "" {
		return errors.New("missing_signature")
	}
	signature = strings.TrimPrefix(signature, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret+":"+provider))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("bad_signature")
	}
	return nil
}

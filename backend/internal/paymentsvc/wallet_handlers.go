package paymentsvc

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	paymentstore "github.com/bitik/backend/internal/store/payments"
	systemstore "github.com/bitik/backend/internal/store/system"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) sellerIDForUser(c *gin.Context) (pgtype.UUID, bool) {
	userID, ok := currentUserID(c)
	if !ok {
		return pgtype.UUID{}, false
	}
	seller, err := s.sellers.GetSellerByUserID(c.Request.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		return pgtype.UUID{}, false
	}
	return seller.ID, true
}

func (s *Service) HandleGetWallet(c *gin.Context) {
	sellerID, ok := s.sellerIDForUser(c)
	if !ok {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
		return
	}
	wallet, err := s.pay.GetOrCreateSellerWallet(c.Request.Context(), paymentstore.GetOrCreateSellerWalletParams{SellerID: sellerID})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load wallet.")
		return
	}
	apiresponse.OK(c, wallet)
}

func (s *Service) HandleListWalletTransactions(c *gin.Context) {
	sellerID, ok := s.sellerIDForUser(c)
	if !ok {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
		return
	}
	wallet, err := s.pay.GetOrCreateSellerWallet(c.Request.Context(), paymentstore.GetOrCreateSellerWalletParams{SellerID: sellerID})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load wallet.")
		return
	}
	limit := int32(50)
	offset := int32(0)
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = int32(n)
		}
	}
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	items, err := s.pay.ListWalletTransactions(c.Request.Context(), paymentstore.ListWalletTransactionsParams{
		SellerWalletID: wallet.ID,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list transactions.")
		return
	}
	apiresponse.OK(c, items)
}

type requestPayoutRequest struct {
	AmountCents int64  `json:"amount_cents" binding:"required"`
	Currency    string `json:"currency"`
}

func (s *Service) HandleRequestPayout(c *gin.Context) {
	sellerID, ok := s.sellerIDForUser(c)
	if !ok {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
		return
	}
	var req requestPayoutRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.AmountCents <= 0 {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payout request.")
		return
	}

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not request payout.")
		return
	}
	defer tx.Rollback(ctx)
	pq := s.pay.WithTx(tx)

	_, err = pq.GetOrCreateSellerWallet(ctx, paymentstore.GetOrCreateSellerWalletParams{SellerID: sellerID})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not request payout.")
		return
	}
	wallet, err := pq.LockSellerWallet(ctx, sellerID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not request payout.")
		return
	}
	if req.AmountCents > wallet.BalanceCents {
		apiresponse.Error(c, http.StatusBadRequest, "insufficient_funds", "Insufficient wallet balance.")
		return
	}

	payout, err := pq.CreatePayoutRequest(ctx, paymentstore.CreatePayoutRequestParams{
		SellerID:         sellerID,
		AmountCents:      req.AmountCents,
		Currency:         text(valueOrDefault(req.Currency, wallet.Currency)),
		Status:           text("pending"),
		Provider:         text("manual"),
		ProviderPayoutID: text(""),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not request payout.")
		return
	}

	before := wallet.BalanceCents
	after := wallet.BalanceCents - req.AmountCents
	if _, err := pq.CreateWalletTransaction(ctx, paymentstore.CreateWalletTransactionParams{
		SellerWalletID:     wallet.ID,
		Type:               "debit",
		AmountCents:        req.AmountCents,
		BalanceBeforeCents: before,
		BalanceAfterCents:  after,
		ReferenceType:      text("payout"),
		ReferenceID:        payout.ID,
		Description:        text("payout requested"),
	}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not request payout.")
		return
	}
	if _, err := pq.UpdateWalletBalances(ctx, paymentstore.UpdateWalletBalancesParams{
		ID:                  wallet.ID,
		BalanceCents:        after,
		PendingBalanceCents: wallet.PendingBalanceCents,
	}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not request payout.")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not request payout.")
		return
	}
	apiresponse.OK(c, payout)
}

func (s *Service) HandleListPayouts(c *gin.Context) {
	sellerID, ok := s.sellerIDForUser(c)
	if !ok {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Seller not found.")
		return
	}
	limit := int32(50)
	offset := int32(0)
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = int32(n)
		}
	}
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	items, err := s.pay.ListPayoutsForSeller(c.Request.Context(), paymentstore.ListPayoutsForSellerParams{
		SellerID: sellerID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list payouts.")
		return
	}
	apiresponse.OK(c, items)
}

type updatePayoutStatusRequest struct {
	Status           string  `json:"status" binding:"required"` // processed|failed|cancelled
	Provider         *string `json:"provider"`
	ProviderPayoutID *string `json:"provider_payout_id"`
}

func (s *Service) HandleUpdatePayoutStatus(c *gin.Context) {
	actor, _ := currentUserID(c)
	payoutID, ok := uuidParam(c, "payout_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payout id.")
		return
	}
	var req updatePayoutStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid request.")
		return
	}
	status := strings.TrimSpace(req.Status)
	if status != "processed" && status != "failed" && status != "cancelled" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_status", "Invalid payout status.")
		return
	}

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update payout.")
		return
	}
	defer tx.Rollback(ctx)
	q := s.pay.WithTx(tx)

	payout, err := q.LockPayoutByID(ctx, payoutID)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Payout not found.")
		return
	}
	current := strings.TrimSpace(payout.Status)
	if current == "processed" || current == "failed" || current == "cancelled" {
		apiresponse.OK(c, payout)
		return
	}

	updated, err := q.UpdatePayoutStatus(ctx, paymentstore.UpdatePayoutStatusParams{
		ID:               payoutID,
		Status:           status,
		Provider:         text(ptrString(req.Provider)),
		ProviderPayoutID: text(ptrString(req.ProviderPayoutID)),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update payout.")
		return
	}

	if status == "failed" || status == "cancelled" {
		_, err := q.GetOrCreateSellerWallet(ctx, paymentstore.GetOrCreateSellerWalletParams{SellerID: payout.SellerID})
		if err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reverse payout.")
			return
		}
		wallet, err := q.LockSellerWallet(ctx, payout.SellerID)
		if err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reverse payout.")
			return
		}
		before := wallet.BalanceCents
		after := wallet.BalanceCents + payout.AmountCents
		if _, err := q.CreateWalletTransaction(ctx, paymentstore.CreateWalletTransactionParams{
			SellerWalletID:     wallet.ID,
			Type:               "credit",
			AmountCents:        payout.AmountCents,
			BalanceBeforeCents: before,
			BalanceAfterCents:  after,
			ReferenceType:      text("payout_reversal"),
			ReferenceID:        payoutID,
			Description:        text("payout reversed"),
		}); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reverse payout.")
			return
		}
		if _, err := q.UpdateWalletBalances(ctx, paymentstore.UpdateWalletBalancesParams{
			ID:                  wallet.ID,
			BalanceCents:        after,
			PendingBalanceCents: wallet.PendingBalanceCents,
		}); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reverse payout.")
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update payout.")
		return
	}
	_, _ = s.sys.CreateAdminActivityLog(ctx, systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgtype.UUID{Bytes: actor, Valid: actor != uuid.Nil},
		Action:      "payout_status_updated",
		EntityType:  text("payout"),
		EntityID:    payoutID,
		Metadata:    jsonObject(map[string]any{"status": status}),
		IpAddress:   ipAddrPtr(c),
		UserAgent:   text(c.Request.UserAgent()),
	})
	apiresponse.OK(c, updated)
}

func (s *Service) HandleSettleSellerWallets(c *gin.Context) {
	const batch = int32(100)
	credits, err := s.pay.ListUnsettledSellerOrderCredits(c.Request.Context(), batch)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load settlement batch.")
		return
	}
	settled := 0
	skipped := 0
	for _, credit := range credits {
		if err := s.applySettlementCredit(c.Request.Context(), credit.SellerID, credit.OrderID, credit.AmountCents); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				skipped++
				continue
			}
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Settlement failed.")
			return
		}
		settled++
	}
	apiresponse.OK(c, gin.H{"settled": settled, "skipped": skipped})
}

func (s *Service) applySettlementCredit(ctx context.Context, sellerID, orderID pgtype.UUID, amount int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := s.pay.WithTx(tx)

	wallet, err := q.GetOrCreateSellerWallet(ctx, paymentstore.GetOrCreateSellerWalletParams{SellerID: sellerID})
	if err != nil {
		return err
	}
	locked, err := q.LockSellerWallet(ctx, sellerID)
	if err != nil {
		return err
	}
	beforePending := locked.PendingBalanceCents
	afterPending := locked.PendingBalanceCents + amount
	if _, err := q.CreateWalletTransaction(ctx, paymentstore.CreateWalletTransactionParams{
		SellerWalletID:     wallet.ID,
		Type:               "credit",
		AmountCents:        amount,
		BalanceBeforeCents: beforePending,
		BalanceAfterCents:  afterPending,
		ReferenceType:      text("order"),
		ReferenceID:        orderID,
		Description:        text("order settlement (pending)"),
	}); err != nil {
		return err
	}
	if _, err := q.UpdateWalletBalances(ctx, paymentstore.UpdateWalletBalancesParams{
		ID:                  wallet.ID,
		BalanceCents:        locked.BalanceCents,
		PendingBalanceCents: afterPending,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) HandleReleaseWalletHolds(c *gin.Context) {
	olderDays := int32(7)
	if v := strings.TrimSpace(c.Query("older_than_days")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 365 {
			olderDays = int32(n)
		}
	}
	batch := int32(100)
	items, err := s.pay.ListOrderCreditsPendingHoldRelease(c.Request.Context(), paymentstore.ListOrderCreditsPendingHoldReleaseParams{
		OlderThanDays: olderDays,
		Limit:         batch,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load holds.")
		return
	}
	released := 0
	skipped := 0
	for _, item := range items {
		if err := s.releaseHold(c.Request.Context(), item.SellerID, item.OrderID, item.AmountCents); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				skipped++
				continue
			}
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Hold release failed.")
			return
		}
		released++
	}
	apiresponse.OK(c, gin.H{"released": released, "skipped": skipped})
}

func (s *Service) releaseHold(ctx context.Context, sellerID, orderID pgtype.UUID, amount int64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := s.pay.WithTx(tx)

	wallet, err := q.GetOrCreateSellerWallet(ctx, paymentstore.GetOrCreateSellerWalletParams{SellerID: sellerID})
	if err != nil {
		return err
	}
	locked, err := q.LockSellerWallet(ctx, sellerID)
	if err != nil {
		return err
	}
	if locked.PendingBalanceCents < amount {
		return errors.New("pending balance underflow")
	}

	beforeAvail := locked.BalanceCents
	afterAvail := locked.BalanceCents + amount

	if _, err := q.CreateWalletTransaction(ctx, paymentstore.CreateWalletTransactionParams{
		SellerWalletID:     wallet.ID,
		Type:               "credit",
		AmountCents:        amount,
		BalanceBeforeCents: beforeAvail,
		BalanceAfterCents:  afterAvail,
		ReferenceType:      text("hold_release"),
		ReferenceID:        orderID,
		Description:        text("hold released to available"),
	}); err != nil {
		return err
	}
	if _, err := q.UpdateWalletBalances(ctx, paymentstore.UpdateWalletBalancesParams{
		ID:                  wallet.ID,
		BalanceCents:        afterAvail,
		PendingBalanceCents: locked.PendingBalanceCents - amount,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

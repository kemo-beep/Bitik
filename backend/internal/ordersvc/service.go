package ordersvc

import (
	"context"
	"encoding/json"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/notify"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/queue"
	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	defaultPage             = int32(1)
	defaultPerPage          = int32(25)
	maxPerPage              = int32(100)
	defaultCheckoutTTL      = 30 * time.Minute
	defaultReservationBatch = int32(100)
	defaultShippingCents    = int64(0)
	defaultTaxCents         = int64(0)
	defaultOrderCurrency    = "USD"
)

type Service struct {
	cfg     config.Config
	log     *zap.Logger
	pool    *pgxpool.Pool
	queries *orderstore.Queries
	pub     notify.Publisher
	queue   *queue.Producer
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool, pub notify.Publisher) *Service {
	return &Service{cfg: cfg, log: logger, pool: pool, queries: orderstore.New(pool), pub: pub}
}

func (s *Service) SetQueueProducer(p *queue.Producer) {
	s.queue = p
}

func (s *Service) TransitionOrder(ctx context.Context, orderID pgtype.UUID, next, note string, actor pgtype.UUID) (orderstore.Order, error) {
	order, err := s.queries.GetOrderByID(ctx, orderID)
	if err != nil {
		return orderstore.Order{}, err
	}
	return s.transitionOrder(ctx, order, next, note, actor)
}

func (s *Service) transitionOrderWithQueries(ctx context.Context, q *orderstore.Queries, order orderstore.Order, next, note string, actor pgtype.UUID) (orderstore.Order, error) {
	if !canTransition(statusString(order.Status), next) {
		return orderstore.Order{}, errInvalidTransition
	}

	updated, err := q.UpdateOrderStatus(ctx, orderstore.UpdateOrderStatusParams{ID: order.ID, Status: next})
	if err != nil {
		return orderstore.Order{}, err
	}
	if _, err := q.InsertOrderStatusHistory(ctx, orderstore.InsertOrderStatusHistoryParams{OrderID: order.ID, OldStatus: order.Status, NewStatus: next, Note: text(note), CreatedBy: actor}); err != nil {
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

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup, auth *authsvc.Service) {
	if auth == nil {
		return
	}
	internal := v1.Group("/internal/jobs", middleware.RequireInternalAPI(s.cfg))
	internal.POST("/release-expired-inventory", s.HandleReleaseExpiredInventory)
	internal.POST("/expire-checkout", s.HandleExpireCheckout)
	internal.POST("/cancel-unpaid-orders", s.HandleCancelUnpaidOrders)
	internal.POST("/generate-invoices", s.HandleGenerateInvoices)

	protected := v1.Group("", middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())

	buyer := protected.Group("/buyer")
	buyer.GET("/addresses", s.HandleListAddresses)
	buyer.POST("/addresses", s.HandleCreateAddress)
	buyer.GET("/addresses/:address_id", s.HandleGetAddress)
	buyer.PATCH("/addresses/:address_id", s.HandleUpdateAddress)
	buyer.DELETE("/addresses/:address_id", s.HandleDeleteAddress)
	buyer.POST("/addresses/:address_id/set-default", s.HandleSetDefaultAddress)

	buyer.GET("/cart", s.HandleGetCart)
	buyer.POST("/cart/items", s.HandleAddCartItem)
	buyer.PATCH("/cart/items/:cart_item_id", s.HandleUpdateCartItem)
	buyer.DELETE("/cart/items/:cart_item_id", s.HandleDeleteCartItem)
	buyer.DELETE("/cart", s.HandleClearCart)
	buyer.POST("/cart/merge", s.HandleMergeCart)
	buyer.POST("/cart/select-items", s.HandleSelectCartItems)
	buyer.POST("/cart/apply-voucher", s.HandleApplyCartVoucher)
	buyer.DELETE("/cart/voucher/:voucher_id", s.HandleRemoveCartVoucher)

	buyer.POST("/checkout/sessions", s.HandleCreateCheckoutSession)
	buyer.GET("/checkout/sessions/:checkout_session_id", s.HandleGetCheckoutSession)
	buyer.PATCH("/checkout/sessions/:checkout_session_id/address", s.HandleUpdateCheckoutAddress)
	buyer.PATCH("/checkout/sessions/:checkout_session_id/shipping", s.HandleUpdateCheckoutShipping)
	buyer.PATCH("/checkout/sessions/:checkout_session_id/payment-method", s.HandleUpdateCheckoutPaymentMethod)
	buyer.POST("/checkout/sessions/:checkout_session_id/apply-voucher", s.HandleApplyCheckoutVoucher)
	buyer.DELETE("/checkout/sessions/:checkout_session_id/voucher/:voucher_id", s.HandleRemoveCheckoutVoucher)
	buyer.POST("/checkout/sessions/:checkout_session_id/validate", s.HandleValidateCheckout)
	buyer.POST("/checkout/sessions/:checkout_session_id/place-order", s.HandlePlaceOrder)

	buyer.GET("/orders", s.HandleBuyerListOrders)
	buyer.GET("/orders/:order_id", s.HandleBuyerGetOrder)
	buyer.GET("/orders/:order_id/items", s.HandleBuyerOrderItems)
	buyer.GET("/orders/:order_id/status-history", s.HandleOrderStatusHistory)
	buyer.POST("/orders/:order_id/cancel", s.HandleBuyerCancelOrder)
	buyer.POST("/orders/:order_id/confirm-received", s.HandleBuyerConfirmReceived)
	buyer.POST("/orders/:order_id/request-refund", s.HandleBuyerRequestRefund)
	buyer.POST("/orders/:order_id/request-return", s.HandleBuyerRequestReturn)
	buyer.POST("/orders/:order_id/dispute", s.HandleBuyerDispute)
	buyer.GET("/orders/:order_id/invoice", s.HandleBuyerInvoice)
	buyer.GET("/orders/:order_id/tracking", s.HandleBuyerTracking)

	seller := protected.Group("/seller", s.requireRole("seller", "admin"), s.requireSeller())
	seller.GET("/orders", s.HandleSellerListOrders)
	seller.GET("/orders/:order_id", s.HandleSellerGetOrder)
	seller.GET("/orders/:order_id/items", s.HandleSellerOrderItems)
	seller.POST("/orders/:order_id/accept", s.HandleSellerAcceptOrder)
	seller.POST("/orders/:order_id/reject", s.HandleSellerRejectOrder)
	seller.POST("/orders/:order_id/mark-processing", s.HandleSellerMarkProcessing)
	seller.POST("/orders/:order_id/pack", s.HandleSellerPackOrder)
	seller.POST("/orders/:order_id/ship", s.HandleSellerShipOrder)
	seller.POST("/orders/:order_id/cancel", s.HandleSellerCancelOrder)
	seller.POST("/orders/:order_id/refund", s.HandleSellerRefundOrder)
	seller.POST("/orders/:order_id/return/approve", s.HandleSellerApproveReturn)
	seller.POST("/orders/:order_id/return/reject", s.HandleSellerRejectReturn)
	seller.POST("/orders/:order_id/return/received", s.HandleSellerMarkReturnReceived)

	admin := protected.Group("/admin", s.requireRole("admin"))
	admin.GET("/orders", s.HandleAdminListOrders)
	admin.GET("/orders/:order_id", s.HandleAdminGetOrder)
	admin.PATCH("/orders/:order_id", s.HandleAdminUpdateOrder)
	admin.POST("/orders/:order_id/cancel", s.HandleAdminCancelOrder)
	admin.POST("/orders/:order_id/refund", s.HandleAdminRefundOrder)
	admin.GET("/orders/:order_id/status-history", s.HandleAdminStatusHistory)
	admin.GET("/orders/:order_id/payments", s.HandleAdminPayments)
	admin.GET("/orders/:order_id/shipments", s.HandleAdminShipments)

	// Internal jobs are registered above with internal token (+ optional CIDR) protection.
}

type pageParams struct {
	Page    int32
	PerPage int32
	Limit   int32
	Offset  int32
}

func pagination(c *gin.Context) pageParams {
	page := parsePositiveInt32(c.DefaultQuery("page", "1"), defaultPage)
	perPage := parsePositiveInt32(c.DefaultQuery("per_page", c.DefaultQuery("limit", strconv.Itoa(int(defaultPerPage)))), defaultPerPage)
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return pageParams{Page: page, PerPage: perPage, Limit: perPage, Offset: (page - 1) * perPage}
}

func parsePositiveInt32(raw string, fallback int32) int32 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 32)
	if err != nil || v < 1 {
		return fallback
	}
	return int32(v)
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(middleware.AuthUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := raw.(uuid.UUID)
	return id, ok
}

func (s *Service) requireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !hasRole(c, roles...) {
			apiresponse.Error(c, http.StatusForbidden, "forbidden", "Required role is missing.")
			c.Abort()
			return
		}
		c.Next()
	}
}

func hasRole(c *gin.Context, allowed ...string) bool {
	raw, _ := c.Get(middleware.AuthRolesKey)
	roles, _ := raw.([]string)
	for _, role := range roles {
		for _, want := range allowed {
			if role == want {
				return true
			}
		}
	}
	return false
}

func (s *Service) requireSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := currentUserID(c)
		if !ok {
			apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
			c.Abort()
			return
		}
		seller, err := s.queries.GetSellerByUserID(c.Request.Context(), pgxutil.UUID(uid))
		if err != nil {
			if err == pgx.ErrNoRows {
				apiresponse.Error(c, http.StatusForbidden, "seller_required", "Seller account is required.")
				c.Abort()
				return
			}
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load seller.")
			c.Abort()
			return
		}
		if statusString(seller.Status) != "active" && !hasRole(c, "admin") {
			apiresponse.Error(c, http.StatusForbidden, "seller_inactive", "Seller account is not active.")
			c.Abort()
			return
		}
		c.Set("seller", seller)
		c.Next()
	}
}

func sellerFromContext(c *gin.Context) (orderstore.Seller, bool) {
	raw, ok := c.Get("seller")
	if !ok {
		return orderstore.Seller{}, false
	}
	seller, ok := raw.(orderstore.Seller)
	return seller, ok
}

func uuidParam(c *gin.Context, name string) (pgtype.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func parseUUID(raw string) (pgtype.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func text(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	return pgtype.Text{String: value, Valid: value != ""}
}

func jsonBytes(value any) []byte {
	if value == nil {
		return nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return b
}

func jsonObject(value any) []byte {
	b := jsonBytes(value)
	if len(b) == 0 {
		return []byte(`{}`)
	}
	return b
}

func uuidString(id pgtype.UUID) string {
	if v, ok := pgxutil.ToUUID(id); ok {
		return v.String()
	}
	return ""
}

func nullableUUID(id pgtype.UUID) any {
	if !id.Valid {
		return nil
	}
	return uuidString(id)
}

func textValue(t pgtype.Text) any {
	if !t.Valid {
		return nil
	}
	return t.String
}

func statusString(v any) string {
	if v == nil {
		return ""
	}
	return strings.Trim(v.(string), `"`)
}

func numericString(n pgtype.Numeric) string {
	if !n.Valid || n.Int == nil {
		return "0"
	}
	r := new(big.Rat).SetInt(n.Int)
	if n.Exp > 0 {
		r.Mul(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)))
	} else if n.Exp < 0 {
		r.Quo(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)))
	}
	return r.FloatString(2)
}

func rawJSON(data []byte) any {
	if len(data) == 0 {
		return nil
	}
	return json.RawMessage(data)
}

func notFoundOrInternal(c *gin.Context, err error, code, message string) {
	if err == pgx.ErrNoRows {
		apiresponse.Error(c, http.StatusNotFound, code, message)
		return
	}
	apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load resource.")
}

func calculateDiscount(v orderstore.Voucher, subtotal, shipping int64) int64 {
	if subtotal < v.MinOrderCents {
		return 0
	}
	var discount int64
	switch v.DiscountType {
	case "fixed":
		discount = v.DiscountValue
	case "percentage":
		discount = int64(math.Floor(float64(subtotal) * float64(v.DiscountValue) / 100.0))
	case "free_shipping":
		discount = shipping
	}
	if v.MaxDiscountCents.Valid && discount > v.MaxDiscountCents.Int64 {
		discount = v.MaxDiscountCents.Int64
	}
	if discount > subtotal+shipping {
		discount = subtotal + shipping
	}
	if discount < 0 {
		return 0
	}
	return discount
}

func voucherScopeMatchesCart(v orderstore.Voucher, items []orderstore.ListCartItemsDetailedRow) bool {
	if !v.SellerID.Valid {
		return true
	}
	// Prefer selected items. If none selected, fall back to all items.
	selectedCount := 0
	seen := map[[16]byte]struct{}{}
	for _, it := range items {
		if it.Selected {
			selectedCount++
		}
	}
	for _, it := range items {
		if selectedCount > 0 && !it.Selected {
			continue
		}
		seen[it.SellerID.Bytes] = struct{}{}
		if len(seen) > 1 {
			return false
		}
	}
	if len(seen) == 0 {
		return false
	}
	for sellerBytes := range seen {
		return sellerBytes == v.SellerID.Bytes
	}
	return false
}

func voucherScopeMatchesCheckout(v orderstore.Voucher, items []orderstore.CheckoutSessionItem) bool {
	if !v.SellerID.Valid {
		return true
	}
	seen := map[[16]byte]struct{}{}
	for _, it := range items {
		seen[it.SellerID.Bytes] = struct{}{}
		if len(seen) > 1 {
			return false
		}
	}
	if len(seen) == 0 {
		return false
	}
	for sellerBytes := range seen {
		return sellerBytes == v.SellerID.Bytes
	}
	return false
}

func orderNumber() string {
	return "B" + time.Now().UTC().Format("20060102150405") + strings.ToUpper(uuid.NewString()[:8])
}

func invoiceNumber(order orderstore.Order) string {
	return "INV-" + order.OrderNumber
}

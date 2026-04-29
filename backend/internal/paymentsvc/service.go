package paymentsvc

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/ordersvc"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/queue"
	orderstore "github.com/bitik/backend/internal/store/orders"
	paymentstore "github.com/bitik/backend/internal/store/payments"
	sellerstore "github.com/bitik/backend/internal/store/sellers"
	systemstore "github.com/bitik/backend/internal/store/system"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	cfg     config.Config
	log     *zap.Logger
	pool    *pgxpool.Pool
	pay     *paymentstore.Queries
	sys     *systemstore.Queries
	ord     *orderstore.Queries
	sellers *sellerstore.Queries
	orders  *ordersvc.Service
	queue   *queue.Producer
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool) *Service {
	return &Service{
		cfg:     cfg,
		log:     logger,
		pool:    pool,
		pay:     paymentstore.New(pool),
		sys:     systemstore.New(pool),
		ord:     orderstore.New(pool),
		sellers: sellerstore.New(pool),
		orders:  ordersvc.NewService(cfg, logger, pool, nil),
	}
}

func (s *Service) SetQueueProducer(p *queue.Producer) {
	s.queue = p
}

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup, auth *authsvc.Service) {
	if auth == nil {
		return
	}
	// Provider webhooks are unauthenticated by design. Store raw events for later processing.
	v1.POST("/webhooks/:provider", s.HandleWebhook)

	internal := v1.Group("/internal/jobs", middleware.RequireInternalAPI(s.cfg))
	internal.POST("/expire-pending-payments", s.HandleExpirePendingPayments)
	internal.POST("/settle-seller-wallets", s.HandleSettleSellerWallets)
	internal.POST("/release-wallet-holds", s.HandleReleaseWalletHolds)

	protected := v1.Group("", middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())

	buyer := protected.Group("/buyer")
	buyer.GET("/payment-methods", s.HandleListPaymentMethods)
	buyer.POST("/payment-methods", s.HandleCreatePaymentMethod)
	buyer.DELETE("/payment-methods/:method_id", s.HandleDeletePaymentMethod)
	buyer.POST("/payment-methods/:method_id/set-default", s.HandleSetDefaultPaymentMethod)

	buyer.POST("/payments/create-intent", s.HandleCreateIntent)
	buyer.POST("/payments/confirm", s.HandleConfirmPayment)
	buyer.GET("/payments/:payment_id", s.HandleGetPayment)
	buyer.POST("/payments/:payment_id/retry", s.HandleRetryPayment)
	buyer.POST("/payments/:payment_id/cancel", s.HandleCancelPayment)

	admin := protected.Group("/admin", s.requireRole("admin", "ops_payments"))
	admin.GET("/payments/wave/pending", s.HandleListPendingWave)
	admin.POST("/payments/:payment_id/wave/approve", s.HandleApproveWave)
	admin.POST("/payments/:payment_id/wave/reject", s.HandleRejectWave)

	admin.POST("/payments/:payment_id/pod/capture", s.HandleCapturePOD)
	admin.POST("/refunds/:refund_id/review", s.HandleReviewRefund)
	admin.POST("/refunds/:refund_id/process", s.HandleProcessRefund)
	admin.POST("/payouts/:payout_id/status", s.HandleUpdatePayoutStatus)
	admin.GET("/payments/webhooks", s.HandleAdminListWebhookEvents)
	admin.GET("/payments/webhooks/:event_id", s.HandleAdminGetWebhookEvent)
	admin.POST("/payments/webhooks/:event_id/reprocess", s.HandleAdminReprocessWebhookEvent)

	seller := protected.Group("/seller", s.requireRole("seller", "admin", "shipping"))
	seller.POST("/payments/:payment_id/pod/capture", s.HandleCapturePOD)
	seller.GET("/wallet", s.HandleGetWallet)
	seller.GET("/wallet/transactions", s.HandleListWalletTransactions)
	seller.GET("/bank-accounts", s.HandleListSellerBankAccounts)
	seller.POST("/bank-accounts", s.HandleCreateSellerBankAccount)
	seller.PATCH("/bank-accounts/:bank_account_id", s.HandleUpdateSellerBankAccount)
	seller.DELETE("/bank-accounts/:bank_account_id", s.HandleDeleteSellerBankAccount)
	seller.POST("/bank-accounts/:bank_account_id/set-default", s.HandleSetDefaultSellerBankAccount)
	seller.POST("/payouts/request", s.HandleRequestPayout)
	seller.GET("/payouts", s.HandleListPayouts)

	// Internal jobs are registered above with internal token (+ optional CIDR) protection.
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

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(middleware.AuthUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := raw.(uuid.UUID)
	return id, ok
}

func uuidParam(c *gin.Context, name string) (pgtype.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func text(v string) pgtype.Text {
	v = strings.TrimSpace(v)
	return pgtype.Text{String: v, Valid: v != ""}
}

func jsonObject(v any) []byte {
	if v == nil {
		return []byte(`{}`)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	return b
}

func ipAddrPtr(c *gin.Context) *netip.Addr {
	ipStr := strings.TrimSpace(c.ClientIP())
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil
	}
	return &ip
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

func idempotencyKey(c *gin.Context) string {
	raw, ok := c.Get(middleware.IdempotencyContextKey)
	if !ok {
		return strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	}
	if s, ok := raw.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(c.GetHeader("Idempotency-Key"))
}

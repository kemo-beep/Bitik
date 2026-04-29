package promotionsvc

import (
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	orderstore "github.com/bitik/backend/internal/store/orders"
	promostore "github.com/bitik/backend/internal/store/promotions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	cfg    config.Config
	log    *zap.Logger
	promo  *promostore.Queries
	orders *orderstore.Queries
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool) *Service {
	return &Service{
		cfg:    cfg,
		log:    logger,
		promo:  promostore.New(pool),
		orders: orderstore.New(pool),
	}
}

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup, auth *authsvc.Service) {
	if auth == nil {
		return
	}
	protected := v1.Group("", middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())

	buyer := protected.Group("/buyer", s.requireRole("buyer", "admin"))
	buyer.GET("/vouchers", s.HandleBuyerListVouchers)
	buyer.POST("/vouchers/validate", s.HandleBuyerValidateVoucher)

	seller := protected.Group("/seller", s.requireRole("seller", "admin"), s.requireSeller())
	seller.GET("/vouchers", s.HandleSellerListVouchers)
	seller.POST("/vouchers", s.HandleSellerCreateVoucher)
	seller.PATCH("/vouchers/:voucher_id", s.HandleSellerUpdateVoucher)
	seller.DELETE("/vouchers/:voucher_id", s.HandleSellerDeleteVoucher)

	admin := protected.Group("/admin", s.requireRole("admin"))
	admin.GET("/vouchers", s.HandleAdminListVouchers)
	admin.POST("/vouchers", s.HandleAdminCreateVoucher)
	admin.PATCH("/vouchers/:voucher_id", s.HandleAdminUpdateVoucher)
	admin.DELETE("/vouchers/:voucher_id", s.HandleAdminDeleteVoucher)
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
		raw, _ := c.Get(middleware.AuthRolesKey)
		got, _ := raw.([]string)
		for _, r := range got {
			for _, want := range roles {
				if r == want {
					c.Next()
					return
				}
			}
		}
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Required role is missing.")
		c.Abort()
	}
}

func (s *Service) requireSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := currentUserID(c)
		if !ok {
			apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
			c.Abort()
			return
		}
		seller, err := s.orders.GetSellerByUserID(c.Request.Context(), pgxutil.UUID(uid))
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
		c.Set("seller", seller)
		c.Next()
	}
}

func sellerFromContext(c *gin.Context) (orderstore.Seller, bool) {
	raw, ok := c.Get("seller")
	if !ok {
		return orderstore.Seller{}, false
	}
	s, ok := raw.(orderstore.Seller)
	return s, ok
}


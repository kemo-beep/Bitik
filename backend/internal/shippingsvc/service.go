package shippingsvc

import (
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	orderstore "github.com/bitik/backend/internal/store/orders"
	shippingstore "github.com/bitik/backend/internal/store/shipping"
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
	shipQ   *shippingstore.Queries
	orderQ  *orderstore.Queries
	systemQ *systemstore.Queries
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool) *Service {
	return &Service{
		cfg:     cfg,
		log:     logger,
		pool:    pool,
		shipQ:   shippingstore.New(pool),
		orderQ:  orderstore.New(pool),
		systemQ: systemstore.New(pool),
	}
}

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup, auth *authsvc.Service) {
	if auth == nil {
		return
	}
	internal := v1.Group("/internal/jobs", middleware.RequireInternalAPI(s.cfg))
	internal.POST("/update-shipment-tracking", s.HandleUpdateShipmentTracking)

	protected := v1.Group("", middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())

	buyer := protected.Group("/buyer")
	buyer.GET("/orders/:order_id/shipments", s.HandleBuyerOrderShipments)

	seller := protected.Group("/seller", s.requireRole("seller", "admin"), s.requireSeller())
	seller.GET("/orders/:order_id/shipments", s.HandleSellerOrderShipments)
	seller.PATCH("/shipments/:shipment_id", s.HandleSellerPatchShipment)
	seller.POST("/shipments/:shipment_id/mark-packed", s.HandleSellerMarkPacked)
	seller.POST("/shipments/:shipment_id/mark-shipped", s.HandleSellerMarkShipped)
	seller.POST("/shipments/:shipment_id/mark-in-transit", s.HandleSellerMarkInTransit)
	seller.POST("/shipments/:shipment_id/mark-delivered", s.HandleSellerMarkDelivered)
	seller.POST("/shipments/:shipment_id/labels", s.HandleSellerCreateLabel)

	admin := protected.Group("/admin", s.requireRole("admin"))
	admin.GET("/shipping/providers", s.HandleAdminListProviders)
	admin.POST("/shipping/providers", s.HandleAdminCreateProvider)
	admin.PATCH("/shipping/providers/:provider_id", s.HandleAdminUpdateProvider)
	admin.GET("/shipping/shipments", s.HandleAdminListShipments)
	admin.GET("/shipping/shipments/:shipment_id", s.HandleAdminGetShipment)
	admin.GET("/shipping/shipments/:shipment_id/tracking", s.HandleAdminShipmentTracking)
	admin.POST("/shipping/shipments/:shipment_id/status", s.HandleAdminUpdateShipmentStatus)

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

func (s *Service) requireSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := currentUserID(c)
		if !ok {
			apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
			c.Abort()
			return
		}
		seller, err := s.orderQ.GetSellerByUserID(c.Request.Context(), pgxutil.UUID(uid))
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
	seller, ok := raw.(orderstore.Seller)
	return seller, ok
}

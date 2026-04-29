package reviewsvc

import (
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/notify"
	"github.com/bitik/backend/internal/pgxutil"
	orderstore "github.com/bitik/backend/internal/store/orders"
	reviewstore "github.com/bitik/backend/internal/store/reviews"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	cfg    config.Config
	log    *zap.Logger
	reviews *reviewstore.Queries
	orders  *orderstore.Queries
	pub     notify.Publisher
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool, pub notify.Publisher) *Service {
	return &Service{
		cfg:     cfg,
		log:     logger,
		reviews: reviewstore.New(pool),
		orders:  orderstore.New(pool),
		pub:     pub,
	}
}

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup, auth *authsvc.Service) {
	if auth == nil {
		return
	}
	protected := v1.Group("", middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())

	buyer := protected.Group("/buyer", s.requireRole("buyer", "admin"))
	buyer.POST("/reviews", s.HandleBuyerCreateReview)
	buyer.PATCH("/reviews/:review_id", s.HandleBuyerUpdateReview)
	buyer.DELETE("/reviews/:review_id", s.HandleBuyerDeleteReview)
	buyer.GET("/reviews/:review_id/images", s.HandleListReviewImages)
	buyer.POST("/reviews/:review_id/images", s.HandleAddReviewImage)
	buyer.DELETE("/reviews/:review_id/images/:image_id", s.HandleDeleteReviewImage)
	buyer.POST("/reviews/:review_id/vote", s.HandleVoteReview)
	buyer.POST("/reviews/:review_id/report", s.HandleReportReview)

	seller := protected.Group("/seller", s.requireRole("seller", "admin"), s.requireSeller())
	seller.POST("/reviews/:review_id/reply", s.HandleSellerReply)
	seller.POST("/reviews/:review_id/report", s.HandleSellerReportReview)

	admin := protected.Group("/admin", s.requireRole("admin"))
	admin.PATCH("/reviews/:review_id/hide", s.HandleAdminHideReview)
	admin.DELETE("/reviews/:review_id", s.HandleAdminDeleteReview)
	admin.GET("/reviews/reports", s.HandleAdminListOpenReports)
	admin.PATCH("/reviews/reports/:report_id/resolve", s.HandleAdminResolveReport)
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


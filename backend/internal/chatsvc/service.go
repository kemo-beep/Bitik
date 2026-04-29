package chatsvc

import (
	"encoding/json"
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/notify"
	"github.com/bitik/backend/internal/pgxutil"
	chatstore "github.com/bitik/backend/internal/store/chat"
	notifystore "github.com/bitik/backend/internal/store/notifications"
	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	cfg    config.Config
	log    *zap.Logger
	chat   *chatstore.Queries
	notifs *notifystore.Queries
	orders *orderstore.Queries
	pub    notify.Publisher
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool, pub notify.Publisher) *Service {
	return &Service{
		cfg:    cfg,
		log:    logger,
		chat:   chatstore.New(pool),
		notifs: notifystore.New(pool),
		orders: orderstore.New(pool),
		pub:    pub,
	}
}

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup, auth *authsvc.Service) {
	if auth == nil {
		return
	}
	protected := v1.Group("", middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())

	chat := protected.Group("/chat", s.requireRole("buyer", "seller", "admin"))
	chat.GET("/conversations", s.HandleListConversations)
	chat.POST("/conversations", s.HandleCreateConversation)
	chat.GET("/conversations/:conversation_id/messages", s.HandleListMessages)
	chat.POST("/conversations/:conversation_id/messages", s.HandleSendMessage)
	chat.PATCH("/conversations/:conversation_id/read", s.HandleMarkRead)
	chat.DELETE("/conversations/:conversation_id/messages/:message_id", s.HandleDeleteMessage)
	chat.DELETE("/conversations/:conversation_id", s.HandleDeleteConversation)
}

func jsonObject(v any) []byte {
	if v == nil {
		return []byte("{}")
	}
	b, err := json.Marshal(v)
	if err != nil || len(b) == 0 {
		return []byte("{}")
	}
	return b
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

func (s *Service) sellerIDForUser(c *gin.Context, uid uuid.UUID) (uuid.UUID, bool) {
	seller, err := s.orders.GetSellerByUserID(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		return uuid.Nil, false
	}
	if v, ok := pgxutil.ToUUID(seller.ID); ok {
		return v, true
	}
	return uuid.Nil, false
}

func (s *Service) isSeller(c *gin.Context) bool {
	raw, _ := c.Get(middleware.AuthRolesKey)
	got, _ := raw.([]string)
	for _, r := range got {
		if r == "seller" {
			return true
		}
	}
	return false
}

var _ = pgx.ErrNoRows

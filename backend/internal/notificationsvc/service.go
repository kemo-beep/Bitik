package notificationsvc

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/notify"
	"github.com/bitik/backend/internal/pgxutil"
	notifystore "github.com/bitik/backend/internal/store/notifications"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	cfg     config.Config
	log     *zap.Logger
	queries *notifystore.Queries
	pub     notify.Publisher
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool, pub notify.Publisher) *Service {
	return &Service{
		cfg:     cfg,
		log:     logger,
		queries: notifystore.New(pool),
		pub:     pub,
	}
}

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup, auth *authsvc.Service) {
	if auth == nil {
		return
	}
	protected := v1.Group("", middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())
	buyer := protected.Group("/buyer")
	seller := protected.Group("/seller")
	admin := protected.Group("/admin")

	buyer.GET("/notifications", s.HandleList)
	buyer.GET("/notifications/unread-count", s.HandleUnreadCount)
	buyer.PATCH("/notifications/:notification_id/read", s.HandleMarkRead)
	buyer.PATCH("/notifications/read-all", s.HandleReadAll)
	buyer.DELETE("/notifications/:notification_id", s.HandleDelete)

	buyer.GET("/notifications/preferences", s.HandleGetPreferences)
	buyer.PUT("/notifications/preferences", s.HandleUpsertPreferences)

	buyer.GET("/push-tokens", s.HandleListPushTokens)
	buyer.POST("/push-tokens", s.HandleCreatePushToken)
	buyer.DELETE("/push-tokens", s.HandleDeletePushToken)

	seller.GET("/notifications", s.HandleList)
	seller.GET("/notifications/unread-count", s.HandleUnreadCount)
	seller.PATCH("/notifications/:notification_id/read", s.HandleMarkRead)
	seller.PATCH("/notifications/read-all", s.HandleReadAll)
	seller.DELETE("/notifications/:notification_id", s.HandleDelete)
	seller.GET("/notifications/preferences", s.HandleGetPreferences)
	seller.PUT("/notifications/preferences", s.HandleUpsertPreferences)
	seller.GET("/push-tokens", s.HandleListPushTokens)
	seller.POST("/push-tokens", s.HandleCreatePushToken)
	seller.DELETE("/push-tokens", s.HandleDeletePushToken)

	admin.GET("/notifications", s.HandleList)
	admin.GET("/notifications/unread-count", s.HandleUnreadCount)
	admin.PATCH("/notifications/:notification_id/read", s.HandleMarkRead)
	admin.PATCH("/notifications/read-all", s.HandleReadAll)
	admin.DELETE("/notifications/:notification_id", s.HandleDelete)
	admin.GET("/notifications/preferences", s.HandleGetPreferences)
	admin.PUT("/notifications/preferences", s.HandleUpsertPreferences)
	admin.GET("/push-tokens", s.HandleListPushTokens)
	admin.POST("/push-tokens", s.HandleCreatePushToken)
	admin.DELETE("/push-tokens", s.HandleDeletePushToken)
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(middleware.AuthUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := raw.(uuid.UUID)
	return id, ok
}

type pageParams struct {
	Page  int32
	Limit int32
	Offset int32
}

func parsePage(c *gin.Context) pageParams {
	page := int32(1)
	limit := int32(20)
	if v, err := strconv.ParseInt(strings.TrimSpace(c.DefaultQuery("page", "1")), 10, 32); err == nil && v > 0 {
		page = int32(v)
	}
	if v, err := strconv.ParseInt(strings.TrimSpace(c.DefaultQuery("limit", "20")), 10, 32); err == nil && v > 0 {
		limit = int32(v)
	}
	if limit > 100 {
		limit = 100
	}
	return pageParams{Page: page, Limit: limit, Offset: (page - 1) * limit}
}

func pageMeta(p pageParams, total int64) map[string]any {
	hasNext := int64(p.Page)*int64(p.Limit) < total
	return map[string]any{"page": p.Page, "limit": p.Limit, "total": total, "has_next": hasNext}
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func (s *Service) HandleList(c *gin.Context) {
	uid, _ := currentUserID(c)
	p := parsePage(c)
	total, err := s.queries.CountNotifications(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load notifications.")
		return
	}
	rows, err := s.queries.ListNotifications(c.Request.Context(), notifystore.ListNotificationsParams{UserID: pgxutil.UUID(uid), Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load notifications.")
		return
	}
	items := make([]gin.H, 0, len(rows))
	for _, n := range rows {
		items = append(items, notificationJSON(n))
	}
	apiresponse.Respond(c, http.StatusOK, gin.H{"items": items}, map[string]any{"pagination": pageMeta(p, total)})
}

func (s *Service) HandleUnreadCount(c *gin.Context) {
	uid, _ := currentUserID(c)
	count, err := s.queries.CountUnreadNotifications(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load unread count.")
		return
	}
	apiresponse.OK(c, gin.H{"unread": count})
}

func (s *Service) HandleMarkRead(c *gin.Context) {
	uid, _ := currentUserID(c)
	id, ok := parseUUIDParam(c, "notification_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_notification_id", "Invalid notification id.")
		return
	}
	updated, err := s.queries.MarkNotificationRead(c.Request.Context(), notifystore.MarkNotificationReadParams{ID: pgxutil.UUID(id), UserID: pgxutil.UUID(uid)})
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "notification_not_found", "Notification not found.")
		return
	}
	apiresponse.OK(c, notificationJSON(updated))
}

func (s *Service) HandleReadAll(c *gin.Context) {
	uid, _ := currentUserID(c)
	if err := s.queries.MarkAllNotificationsRead(c.Request.Context(), pgxutil.UUID(uid)); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not mark all read.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleDelete(c *gin.Context) {
	uid, _ := currentUserID(c)
	id, ok := parseUUIDParam(c, "notification_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_notification_id", "Invalid notification id.")
		return
	}
	if err := s.queries.DeleteNotification(c.Request.Context(), notifystore.DeleteNotificationParams{ID: pgxutil.UUID(id), UserID: pgxutil.UUID(uid)}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete notification.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleGetPreferences(c *gin.Context) {
	uid, _ := currentUserID(c)
	row, err := s.queries.GetNotificationPreferences(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		// Treat missing row as defaults.
		apiresponse.OK(c, gin.H{
			"email_enabled":     true,
			"sms_enabled":       true,
			"push_enabled":      true,
			"marketing_enabled": false,
			"quiet_hours":       map[string]any{},
		})
		return
	}
	apiresponse.OK(c, preferencesJSON(row))
}

func (s *Service) HandleUpsertPreferences(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		EmailEnabled     *bool          `json:"email_enabled"`
		SmsEnabled       *bool          `json:"sms_enabled"`
		PushEnabled      *bool          `json:"push_enabled"`
		MarketingEnabled *bool          `json:"marketing_enabled"`
		QuietHours       map[string]any `json:"quiet_hours"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid preferences request.")
		return
	}
	// If nil, keep current defaults.
	email := valueOr(req.EmailEnabled, true)
	sms := valueOr(req.SmsEnabled, true)
	push := valueOr(req.PushEnabled, true)
	marketing := valueOr(req.MarketingEnabled, false)
	row, err := s.queries.UpsertNotificationPreferences(c.Request.Context(), notifystore.UpsertNotificationPreferencesParams{
		UserID:           pgxutil.UUID(uid),
		EmailEnabled:     email,
		SmsEnabled:       sms,
		PushEnabled:      push,
		MarketingEnabled: marketing,
		QuietHours:       jsonObject(req.QuietHours),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "preferences_update_failed", "Could not update preferences.")
		return
	}
	apiresponse.OK(c, preferencesJSON(row))
}

func (s *Service) HandleListPushTokens(c *gin.Context) {
	uid, _ := currentUserID(c)
	p := parsePage(c)
	rows, err := s.queries.ListPushTokens(c.Request.Context(), notifystore.ListPushTokensParams{UserID: pgxutil.UUID(uid), Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load push tokens.")
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, t := range rows {
		out = append(out, gin.H{"id": uuidString(t.ID), "token": t.Token, "platform": t.Platform, "created_at": t.CreatedAt.Time})
	}
	apiresponse.OK(c, gin.H{"items": out})
}

func (s *Service) HandleCreatePushToken(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		Token    string `json:"token" binding:"required"`
		Platform string `json:"platform" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Token) == "" || strings.TrimSpace(req.Platform) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid push token request.")
		return
	}
	row, err := s.queries.CreatePushToken(c.Request.Context(), notifystore.CreatePushTokenParams{UserID: pgxutil.UUID(uid), Token: req.Token, Platform: req.Platform})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "push_token_create_failed", "Could not register push token.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, gin.H{"id": uuidString(row.ID), "token": row.Token, "platform": row.Platform, "created_at": row.CreatedAt.Time}, nil)
}

func (s *Service) HandleDeletePushToken(c *gin.Context) {
	uid, _ := currentUserID(c)
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "token", Message: "is required (query param)"}})
		return
	}
	if err := s.queries.DeletePushToken(c.Request.Context(), notifystore.DeletePushTokenParams{UserID: pgxutil.UUID(uid), Token: token}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete push token.")
		return
	}
	c.Status(http.StatusNoContent)
}

// JSON helpers

func notificationJSON(n notifystore.Notification) gin.H {
	return gin.H{
		"id":         uuidString(n.ID),
		"type":       n.Type,
		"title":      n.Title,
		"body":       textValue(n.Body),
		"data":       n.Data,
		"read_at":    timestamptzValue(n.ReadAt),
		"created_at": n.CreatedAt.Time,
	}
}

func preferencesJSON(p notifystore.NotificationPreference) gin.H {
	return gin.H{
		"email_enabled":     p.EmailEnabled,
		"sms_enabled":       p.SmsEnabled,
		"push_enabled":      p.PushEnabled,
		"marketing_enabled": p.MarketingEnabled,
		"quiet_hours":       p.QuietHours,
		"updated_at":        p.UpdatedAt.Time,
	}
}

func uuidString(id pgtype.UUID) string {
	if v, ok := pgxutil.ToUUID(id); ok {
		return v.String()
	}
	return ""
}

func textValue(t pgtype.Text) any {
	if !t.Valid {
		return nil
	}
	return t.String
}

func timestamptzValue(t pgtype.Timestamptz) any {
	if !t.Valid {
		return nil
	}
	return t.Time
}

func valueOr(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func jsonObject(v map[string]any) []byte {
	if v == nil {
		return []byte("{}")
	}
	b, _ := json.Marshal(v)
	if len(b) == 0 {
		return []byte("{}")
	}
	return b
}


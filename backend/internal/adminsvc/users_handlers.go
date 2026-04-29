package adminsvc

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"sort"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	rbacstore "github.com/bitik/backend/internal/store/rbac"
	systemstore "github.com/bitik/backend/internal/store/system"
	userstore "github.com/bitik/backend/internal/store/users"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleAdminListUsers(c *gin.Context) {
	page := parsePositiveInt32(c.DefaultQuery("page", "1"), 1)
	perPage := parsePositiveInt32(c.DefaultQuery("per_page", c.DefaultQuery("limit", "25")), 25)
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	q := strings.TrimSpace(c.Query("q"))
	status := strings.TrimSpace(c.Query("status"))

	total, err := s.users.AdminCountUsers(c.Request.Context(), userstore.AdminCountUsersParams{
		Status: text(status),
		Q:      text(q),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list users.")
		return
	}
	rows, err := s.users.AdminListUsers(c.Request.Context(), userstore.AdminListUsersParams{
		Status: text(status),
		Q:      text(q),
		Offset: offset,
		Limit:  perPage,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list users.")
		return
	}

	items := make([]gin.H, 0, len(rows))
	for _, u := range rows {
		items = append(items, adminUserJSON(u))
	}
	apiresponse.OK(c, gin.H{
		"items": items,
		"pagination": gin.H{
			"page":     page,
			"per_page": perPage,
			"total":    total,
		},
	})
}

func (s *Service) HandleAdminGetUser(c *gin.Context) {
	userID, ok := uuidParam(c, "user_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_user_id", "Invalid user id.")
		return
	}
	u, err := s.users.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "User not found.")
		return
	}
	roles, _ := s.rbac.ListRoleNamesForUser(c.Request.Context(), u.ID)
	apiresponse.OK(c, gin.H{"user": adminUserJSON(u), "roles": roles})
}

func (s *Service) HandleAdminUpdateUserStatus(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	targetID, ok := uuidParam(c, "user_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_user_id", "Invalid user id.")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Status) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid status update request.")
		return
	}
	newStatus := strings.TrimSpace(req.Status)

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update user.")
		return
	}
	defer tx.Rollback(ctx)

	uq := s.users.WithTx(tx)
	sq := systemstore.New(s.pool).WithTx(tx)

	before, err := uq.GetUserByID(ctx, targetID)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "User not found.")
		return
	}
	after, err := uq.AdminUpdateUserStatus(ctx, userstore.AdminUpdateUserStatusParams{
		Status: newStatus,
		ID:     targetID,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update user.")
		return
	}

	oldVals := map[string]any{"status": statusString(before.Status)}
	newVals := map[string]any{"status": statusString(after.Status)}

	_, _ = sq.CreateAdminActivityLog(ctx, systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "user_status_updated",
		EntityType:  text("user"),
		EntityID:    after.ID,
		Metadata:    jsonObject(map[string]any{"old": oldVals, "new": newVals}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})
	_, _ = sq.CreateAuditLog(ctx, systemstore.CreateAuditLogParams{
		ActorUserID: pgxutil.UUID(actor),
		Action:      "user.status_update.admin",
		EntityType:  text("user"),
		EntityID:    after.ID,
		OldValues:   jsonObject(oldVals),
		NewValues:   jsonObject(newVals),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update user.")
		return
	}
	apiresponse.OK(c, adminUserJSON(after))
}

func (s *Service) HandleAdminGetUserRoles(c *gin.Context) {
	userID, ok := uuidParam(c, "user_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_user_id", "Invalid user id.")
		return
	}
	roles, err := s.rbac.ListRoleNamesForUser(c.Request.Context(), userID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load roles.")
		return
	}
	apiresponse.OK(c, gin.H{"roles": roles})
}

func (s *Service) HandleAdminReplaceUserRoles(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	userID, ok := uuidParam(c, "user_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_user_id", "Invalid user id.")
		return
	}
	var req struct {
		Roles []string `json:"roles"`
	}
	_ = c.ShouldBindJSON(&req)
	roles := make([]string, 0, len(req.Roles))
	seen := map[string]struct{}{}
	for _, r := range req.Roles {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		roles = append(roles, r)
	}
	sort.Strings(roles)

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update roles.")
		return
	}
	defer tx.Rollback(ctx)

	rq := s.rbac.WithTx(tx)
	sq := systemstore.New(s.pool).WithTx(tx)

	before, _ := rq.ListRoleNamesForUser(ctx, userID)
	if err := rq.ClearUserRoles(ctx, userID); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update roles.")
		return
	}
	for _, r := range roles {
		if err := rq.AssignUserRoleByRoleName(ctx, rbacstore.AssignUserRoleByRoleNameParams{UserID: userID, RoleName: r}); err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_role", "Role assignment failed.")
			return
		}
	}
	after, _ := rq.ListRoleNamesForUser(ctx, userID)

	_, _ = sq.CreateAdminActivityLog(ctx, systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "user_roles_replaced",
		EntityType:  text("user"),
		EntityID:    userID,
		Metadata:    jsonObject(map[string]any{"old": before, "new": after}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})
	_, _ = sq.CreateAuditLog(ctx, systemstore.CreateAuditLogParams{
		ActorUserID: pgxutil.UUID(actor),
		Action:      "user.roles_replace.admin",
		EntityType:  text("user"),
		EntityID:    userID,
		OldValues:   jsonObject(map[string]any{"roles": before}),
		NewValues:   jsonObject(map[string]any{"roles": after}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update roles.")
		return
	}
	apiresponse.OK(c, gin.H{"roles": after})
}

func adminUserJSON(u userstore.User) gin.H {
	return gin.H{
		"id":             uuidString(u.ID),
		"email":          textValue(u.Email),
		"phone":          textValue(u.Phone),
		"status":         statusString(u.Status),
		"email_verified": u.EmailVerified,
		"phone_verified": u.PhoneVerified,
		"last_login_at":  timestamptzValue(u.LastLoginAt),
		"created_at":     timestamptzValue(u.CreatedAt),
		"updated_at":     timestamptzValue(u.UpdatedAt),
	}
}

func parsePositiveInt32(raw string, def int32) int32 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 32)
	if err != nil || v <= 0 {
		return def
	}
	return int32(v)
}

func uuidString(id pgtype.UUID) string {
	u, ok := pgxutil.ToUUID(id)
	if !ok {
		return ""
	}
	return u.String()
}

func statusString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return ""
}

func textValue(v pgtype.Text) string {
	if !v.Valid {
		return ""
	}
	return v.String
}

func timestamptzValue(v pgtype.Timestamptz) any {
	if !v.Valid {
		return nil
	}
	return v.Time
}

func jsonObject(v any) []byte {
	if v == nil {
		return []byte(`{}`)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{}`)
	}
	if len(b) == 0 {
		return []byte(`{}`)
	}
	return b
}

func ipAddrPtr(raw string) *netip.Addr {
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

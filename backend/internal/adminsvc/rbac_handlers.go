package adminsvc

import (
	"net/http"
	"sort"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	rbacstore "github.com/bitik/backend/internal/store/rbac"
	systemstore "github.com/bitik/backend/internal/store/system"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleAdminListRoles(c *gin.Context) {
	rows, err := s.rbac.AdminListRoles(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list roles.")
		return
	}
	apiresponse.OK(c, gin.H{"items": rows})
}

func (s *Service) HandleAdminCreateRole(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid role request.")
		return
	}
	created, err := s.rbac.AdminCreateRole(c.Request.Context(), rbacstore.AdminCreateRoleParams{
		Name:        strings.TrimSpace(req.Name),
		Description: optText(req.Description),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not create role.")
		return
	}
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "rbac_role_created",
		EntityType:  text("role"),
		EntityID:    created.ID,
		Metadata:    jsonObject(map[string]any{"name": created.Name}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})
	apiresponse.OK(c, created)
}

func (s *Service) HandleAdminUpdateRole(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	roleID, ok := uuidParam(c, "role_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_role_id", "Invalid role id.")
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	_ = c.ShouldBindJSON(&req)
	updated, err := s.rbac.AdminUpdateRole(c.Request.Context(), rbacstore.AdminUpdateRoleParams{
		ID:          roleID,
		Name:        optText(req.Name),
		Description: optText(req.Description),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update role.")
		return
	}
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "rbac_role_updated",
		EntityType:  text("role"),
		EntityID:    updated.ID,
		Metadata:    jsonObject(map[string]any{"name": updated.Name}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})
	apiresponse.OK(c, updated)
}

func (s *Service) HandleAdminDeleteRole(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	roleID, ok := uuidParam(c, "role_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_role_id", "Invalid role id.")
		return
	}
	if err := s.rbac.AdminDeleteRole(c.Request.Context(), roleID); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not delete role.")
		return
	}
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "rbac_role_deleted",
		EntityType:  text("role"),
		EntityID:    roleID,
		Metadata:    jsonObject(map[string]any{}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleAdminListPermissions(c *gin.Context) {
	rows, err := s.rbac.AdminListPermissions(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list permissions.")
		return
	}
	apiresponse.OK(c, gin.H{"items": rows})
}

func (s *Service) HandleAdminCreatePermission(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	var req struct {
		Key         string  `json:"key" binding:"required"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Key) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid permission request.")
		return
	}
	created, err := s.rbac.AdminCreatePermission(c.Request.Context(), rbacstore.AdminCreatePermissionParams{
		Key:         strings.TrimSpace(req.Key),
		Description: optText(req.Description),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not create permission.")
		return
	}
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "rbac_permission_created",
		EntityType:  text("permission"),
		EntityID:    created.ID,
		Metadata:    jsonObject(map[string]any{"key": created.Key}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})
	apiresponse.OK(c, created)
}

func (s *Service) HandleAdminUpdatePermission(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	permID, ok := uuidParam(c, "permission_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_permission_id", "Invalid permission id.")
		return
	}
	var req struct {
		Key         *string `json:"key"`
		Description *string `json:"description"`
	}
	_ = c.ShouldBindJSON(&req)
	updated, err := s.rbac.AdminUpdatePermission(c.Request.Context(), rbacstore.AdminUpdatePermissionParams{
		ID:          permID,
		Key:         optText(req.Key),
		Description: optText(req.Description),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update permission.")
		return
	}
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "rbac_permission_updated",
		EntityType:  text("permission"),
		EntityID:    updated.ID,
		Metadata:    jsonObject(map[string]any{"key": updated.Key}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})
	apiresponse.OK(c, updated)
}

func (s *Service) HandleAdminDeletePermission(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	permID, ok := uuidParam(c, "permission_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_permission_id", "Invalid permission id.")
		return
	}
	if err := s.rbac.AdminDeletePermission(c.Request.Context(), permID); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not delete permission.")
		return
	}
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "rbac_permission_deleted",
		EntityType:  text("permission"),
		EntityID:    permID,
		Metadata:    jsonObject(map[string]any{}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleAdminListRolePermissions(c *gin.Context) {
	roleID, ok := uuidParam(c, "role_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_role_id", "Invalid role id.")
		return
	}
	rows, err := s.rbac.AdminListPermissionsForRole(c.Request.Context(), roleID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list role permissions.")
		return
	}
	apiresponse.OK(c, gin.H{"items": rows})
}

func (s *Service) HandleAdminReplaceRolePermissions(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	roleID, ok := uuidParam(c, "role_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_role_id", "Invalid role id.")
		return
	}
	var req struct {
		PermissionIDs []string `json:"permission_ids"`
	}
	_ = c.ShouldBindJSON(&req)

	ids := make([]uuid.UUID, 0, len(req.PermissionIDs))
	seen := map[uuid.UUID]struct{}{}
	for _, raw := range req.PermissionIDs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_permission_id", "Invalid permission id.")
			return
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })

	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update role permissions.")
		return
	}
	defer tx.Rollback(ctx)

	rq := s.rbac.WithTx(tx)
	sq := systemstore.New(s.pool).WithTx(tx)

	if err := rq.AdminClearPermissionsForRole(ctx, roleID); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update role permissions.")
		return
	}
	for _, id := range ids {
		if err := rq.AdminAddPermissionToRole(ctx, rbacstore.AdminAddPermissionToRoleParams{RoleID: roleID, PermissionID: pgxutil.UUID(id)}); err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update role permissions.")
			return
		}
	}
	after, _ := rq.AdminListPermissionsForRole(ctx, roleID)

	_, _ = sq.CreateAdminActivityLog(ctx, systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      "rbac_role_permissions_replaced",
		EntityType:  text("role"),
		EntityID:    roleID,
		Metadata:    jsonObject(map[string]any{"permission_count": len(after)}),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update role permissions.")
		return
	}
	apiresponse.OK(c, gin.H{"items": after})
}

func optText(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return text(*v)
}

package adminsvc

import (
	"context"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	cmsstore "github.com/bitik/backend/internal/store/cms"
	rbacstore "github.com/bitik/backend/internal/store/rbac"
	systemstore "github.com/bitik/backend/internal/store/system"
	userstore "github.com/bitik/backend/internal/store/users"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	cfg            config.Config
	log            *zap.Logger
	pool           *pgxpool.Pool
	cms            *cmsstore.Queries
	users          *userstore.Queries
	rbac           *rbacstore.Queries
	systemQ        *systemstore.Queries
	readinessCheck func(context.Context) (map[string]string, bool)
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool, readinessCheck func(context.Context) (map[string]string, bool)) *Service {
	return &Service{
		cfg:            cfg,
		log:            logger,
		pool:           pool,
		cms:            cmsstore.New(pool),
		users:          userstore.New(pool),
		rbac:           rbacstore.New(pool),
		systemQ:        systemstore.New(pool),
		readinessCheck: readinessCheck,
	}
}

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup, auth *authsvc.Service) {
	if auth == nil || auth.Casbin == nil {
		return
	}

	protected := v1.Group("", middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())
	protected.POST("/analytics/events", s.HandleIngestAnalyticsEvent)
	admin := protected.Group("/admin", middleware.RequireCasbinHTTP(auth.Casbin))

	admin.GET("/health", s.HandleAdminHealth)

	// Users
	admin.GET("/users", s.HandleAdminListUsers)
	admin.GET("/users/:user_id", s.HandleAdminGetUser)
	admin.PATCH("/users/:user_id/status", s.HandleAdminUpdateUserStatus)
	admin.GET("/users/:user_id/roles", s.HandleAdminGetUserRoles)
	admin.PUT("/users/:user_id/roles", s.HandleAdminReplaceUserRoles)

	// Catalog admin
	admin.GET("/categories", s.HandleAdminListCategories)
	admin.POST("/categories", s.HandleAdminCreateCategory)
	admin.PATCH("/categories/:category_id", s.HandleAdminUpdateCategory)
	admin.DELETE("/categories/:category_id", s.HandleAdminDeleteCategory)
	admin.POST("/categories/reorder", s.HandleAdminReorderCategories)
	admin.GET("/brands", s.HandleAdminListBrands)
	admin.POST("/brands", s.HandleAdminCreateBrand)
	admin.PATCH("/brands/:brand_id", s.HandleAdminUpdateBrand)
	admin.DELETE("/brands/:brand_id", s.HandleAdminDeleteBrand)
	admin.GET("/products", s.HandleAdminListProducts)
	admin.GET("/products/:product_id", s.HandleAdminGetProduct)

	// RBAC
	admin.GET("/rbac/roles", s.HandleAdminListRoles)
	admin.POST("/rbac/roles", s.HandleAdminCreateRole)
	admin.PATCH("/rbac/roles/:role_id", s.HandleAdminUpdateRole)
	admin.DELETE("/rbac/roles/:role_id", s.HandleAdminDeleteRole)

	admin.GET("/rbac/permissions", s.HandleAdminListPermissions)
	admin.POST("/rbac/permissions", s.HandleAdminCreatePermission)
	admin.PATCH("/rbac/permissions/:permission_id", s.HandleAdminUpdatePermission)
	admin.DELETE("/rbac/permissions/:permission_id", s.HandleAdminDeletePermission)

	admin.GET("/rbac/roles/:role_id/permissions", s.HandleAdminListRolePermissions)
	admin.PUT("/rbac/roles/:role_id/permissions", s.HandleAdminReplaceRolePermissions)

	// CMS
	admin.GET("/cms/pages", s.HandleAdminListPages)
	admin.POST("/cms/pages", s.HandleAdminCreatePage)
	admin.GET("/cms/pages/:page_id", s.HandleAdminGetPage)
	admin.PATCH("/cms/pages/:page_id", s.HandleAdminUpdatePage)
	admin.DELETE("/cms/pages/:page_id", s.HandleAdminDeletePage)

	admin.GET("/cms/banners", s.HandleAdminListBanners)
	admin.POST("/cms/banners", s.HandleAdminCreateBanner)
	admin.GET("/cms/banners/:banner_id", s.HandleAdminGetBanner)
	admin.PATCH("/cms/banners/:banner_id", s.HandleAdminUpdateBanner)
	admin.DELETE("/cms/banners/:banner_id", s.HandleAdminDeleteBanner)

	admin.GET("/cms/faqs", s.HandleAdminListFaqs)
	admin.POST("/cms/faqs", s.HandleAdminCreateFaq)
	admin.GET("/cms/faqs/:faq_id", s.HandleAdminGetFaq)
	admin.PATCH("/cms/faqs/:faq_id", s.HandleAdminUpdateFaq)
	admin.DELETE("/cms/faqs/:faq_id", s.HandleAdminDeleteFaq)

	admin.GET("/cms/announcements", s.HandleAdminListAnnouncements)
	admin.POST("/cms/announcements", s.HandleAdminCreateAnnouncement)
	admin.GET("/cms/announcements/:announcement_id", s.HandleAdminGetAnnouncement)
	admin.PATCH("/cms/announcements/:announcement_id", s.HandleAdminUpdateAnnouncement)
	admin.DELETE("/cms/announcements/:announcement_id", s.HandleAdminDeleteAnnouncement)

	// Moderation
	admin.GET("/moderation/reports", s.HandleAdminListModerationReports)
	admin.GET("/moderation/reports/:report_id", s.HandleAdminGetModerationReport)
	admin.PATCH("/moderation/reports/:report_id/status", s.HandleAdminUpdateModerationReportStatus)
	admin.GET("/moderation/cases", s.HandleAdminListModerationCases)
	admin.GET("/moderation/cases/:case_id", s.HandleAdminGetModerationCase)
	admin.POST("/moderation/cases", s.HandleAdminCreateModerationCase)
	admin.PATCH("/moderation/cases/:case_id", s.HandleAdminUpdateModerationCase)

	// Settings and feature flags
	admin.GET("/settings/platform", s.HandleAdminListPlatformSettings)
	admin.PUT("/settings/platform/:key", s.HandleAdminUpsertPlatformSetting)
	admin.GET("/settings/feature-flags", s.HandleAdminListFeatureFlags)
	admin.PUT("/settings/feature-flags/:key", s.HandleAdminUpsertFeatureFlag)

	// Logs and analytics dashboards
	admin.GET("/logs/audit", s.HandleAdminListAuditLogs)
	admin.GET("/logs/activity", s.HandleAdminListAdminActivityLogs)
	admin.GET("/dashboard/overview", s.HandleAdminDashboardOverview)
	admin.GET("/dashboard/charts/events", s.HandleAdminDashboardEventChart)

	// Analytics event ingestion
	admin.POST("/analytics/events", s.HandleIngestAnalyticsEvent)

	internal := v1.Group("/internal/jobs", middleware.RequireInternalAPI(s.cfg))
	internal.POST("/process-analytics-events", s.HandleProcessAnalyticsEvents)
}

func (s *Service) HandleAdminHealth(c *gin.Context) {
	if s.readinessCheck == nil {
		apiresponse.OK(c, gin.H{"checks": gin.H{}, "ready": true})
		return
	}
	checks, ready := s.readinessCheck(c.Request.Context())
	apiresponse.OK(c, gin.H{"checks": checks, "ready": ready})
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, ok := c.Get(middleware.AuthUserIDKey)
	if !ok {
		return uuid.Nil, false
	}
	switch v := raw.(type) {
	case uuid.UUID:
		return v, true
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil, false
		}
		return id, true
	default:
		return uuid.Nil, false
	}
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

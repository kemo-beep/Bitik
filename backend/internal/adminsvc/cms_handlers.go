package adminsvc

import (
	"net/http"
	"strings"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	cmsstore "github.com/bitik/backend/internal/store/cms"
	systemstore "github.com/bitik/backend/internal/store/system"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleAdminListPages(c *gin.Context) {
	p := listParams(c)
	items, err := s.cms.AdminListPages(c.Request.Context(), cmsstore.AdminListPagesParams{
		Status: text(strings.TrimSpace(c.Query("status"))),
		Q:      text(strings.TrimSpace(c.Query("q"))),
		Limit:  p.Limit,
		Offset: p.Offset,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list pages.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminGetPage(c *gin.Context) {
	id, ok := uuidParam(c, "page_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_page_id", "Invalid page id.")
		return
	}
	row, err := s.cms.AdminGetPageByID(c.Request.Context(), id)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Page not found.")
		return
	}
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminCreatePage(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	var req struct {
		Slug        string     `json:"slug" binding:"required"`
		Title       string     `json:"title" binding:"required"`
		Body        string     `json:"body" binding:"required"`
		Status      *string    `json:"status"`
		PublishedAt *time.Time `json:"published_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid page request.")
		return
	}
	row, err := s.cms.AdminCreatePage(c.Request.Context(), cmsstore.AdminCreatePageParams{
		Slug:        strings.TrimSpace(req.Slug),
		Title:       strings.TrimSpace(req.Title),
		Body:        req.Body,
		Status:      enumValue(req.Status),
		PublishedAt: tsValue(req.PublishedAt),
		ActorUserID: pgxutil.UUID(actor),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not create page.")
		return
	}
	s.logAdminAction(c, actor, "cms_page_created", "cms_page", row.ID, map[string]any{"slug": row.Slug})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminUpdatePage(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "page_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_page_id", "Invalid page id.")
		return
	}
	var req struct {
		Slug        *string    `json:"slug"`
		Title       *string    `json:"title"`
		Body        *string    `json:"body"`
		Status      *string    `json:"status"`
		PublishedAt *time.Time `json:"published_at"`
	}
	_ = c.ShouldBindJSON(&req)
	row, err := s.cms.AdminUpdatePage(c.Request.Context(), cmsstore.AdminUpdatePageParams{
		ID:          id,
		Slug:        optText(req.Slug),
		Title:       optText(req.Title),
		Body:        optText(req.Body),
		Status:      enumValue(req.Status),
		PublishedAt: tsValue(req.PublishedAt),
		ActorUserID: pgxutil.UUID(actor),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update page.")
		return
	}
	s.logAdminAction(c, actor, "cms_page_updated", "cms_page", row.ID, map[string]any{"slug": row.Slug})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminDeletePage(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "page_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_page_id", "Invalid page id.")
		return
	}
	if err := s.cms.AdminDeletePage(c.Request.Context(), id); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not delete page.")
		return
	}
	s.logAdminAction(c, actor, "cms_page_deleted", "cms_page", id, map[string]any{})
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleAdminListBanners(c *gin.Context) {
	p := listParams(c)
	items, err := s.cms.AdminListBanners(c.Request.Context(), cmsstore.AdminListBannersParams{
		Status:    enumValue(strPtr(c.Query("status"))),
		Placement: text(strings.TrimSpace(c.Query("placement"))),
		Limit:     p.Limit,
		Offset:    p.Offset,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list banners.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminGetBanner(c *gin.Context) {
	id, ok := uuidParam(c, "banner_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_banner_id", "Invalid banner id.")
		return
	}
	row, err := s.cms.AdminGetBannerByID(c.Request.Context(), id)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Banner not found.")
		return
	}
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminCreateBanner(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	var req struct {
		Title     string     `json:"title" binding:"required"`
		ImageURL  string     `json:"image_url" binding:"required"`
		LinkURL   *string    `json:"link_url"`
		Placement *string    `json:"placement"`
		SortOrder *int32     `json:"sort_order"`
		Status    *string    `json:"status"`
		StartsAt  *time.Time `json:"starts_at"`
		EndsAt    *time.Time `json:"ends_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid banner request.")
		return
	}
	row, err := s.cms.AdminCreateBanner(c.Request.Context(), cmsstore.AdminCreateBannerParams{
		Title:     strings.TrimSpace(req.Title),
		ImageUrl:  strings.TrimSpace(req.ImageURL),
		LinkUrl:   optText(req.LinkURL),
		Placement: optText(req.Placement),
		SortOrder: int32Value(req.SortOrder),
		Status:    enumValue(req.Status),
		StartsAt:  tsValue(req.StartsAt),
		EndsAt:    tsValue(req.EndsAt),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not create banner.")
		return
	}
	s.logAdminAction(c, actor, "cms_banner_created", "cms_banner", row.ID, map[string]any{"placement": row.Placement})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminUpdateBanner(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "banner_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_banner_id", "Invalid banner id.")
		return
	}
	var req struct {
		Title     *string    `json:"title"`
		ImageURL  *string    `json:"image_url"`
		LinkURL   *string    `json:"link_url"`
		Placement *string    `json:"placement"`
		SortOrder *int32     `json:"sort_order"`
		Status    *string    `json:"status"`
		StartsAt  *time.Time `json:"starts_at"`
		EndsAt    *time.Time `json:"ends_at"`
	}
	_ = c.ShouldBindJSON(&req)
	row, err := s.cms.AdminUpdateBanner(c.Request.Context(), cmsstore.AdminUpdateBannerParams{
		ID:        id,
		Title:     optText(req.Title),
		ImageUrl:  optText(req.ImageURL),
		LinkUrl:   optText(req.LinkURL),
		Placement: optText(req.Placement),
		SortOrder: int32Value(req.SortOrder),
		Status:    enumValue(req.Status),
		StartsAt:  tsValue(req.StartsAt),
		EndsAt:    tsValue(req.EndsAt),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update banner.")
		return
	}
	s.logAdminAction(c, actor, "cms_banner_updated", "cms_banner", row.ID, map[string]any{"placement": row.Placement})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminDeleteBanner(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "banner_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_banner_id", "Invalid banner id.")
		return
	}
	if err := s.cms.AdminDeleteBanner(c.Request.Context(), id); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not delete banner.")
		return
	}
	s.logAdminAction(c, actor, "cms_banner_deleted", "cms_banner", id, map[string]any{})
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleAdminListFaqs(c *gin.Context) {
	p := listParams(c)
	items, err := s.cms.AdminListFaqs(c.Request.Context(), cmsstore.AdminListFaqsParams{
		Status:   enumValue(strPtr(c.Query("status"))),
		Category: text(strings.TrimSpace(c.Query("category"))),
		Limit:    p.Limit,
		Offset:   p.Offset,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list FAQs.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminGetFaq(c *gin.Context) {
	id, ok := uuidParam(c, "faq_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_faq_id", "Invalid faq id.")
		return
	}
	row, err := s.cms.AdminGetFaqByID(c.Request.Context(), id)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "FAQ not found.")
		return
	}
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminCreateFaq(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	var req struct {
		Question  string  `json:"question" binding:"required"`
		Answer    string  `json:"answer" binding:"required"`
		Category  *string `json:"category"`
		SortOrder *int32  `json:"sort_order"`
		Status    *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid faq request.")
		return
	}
	row, err := s.cms.AdminCreateFaq(c.Request.Context(), cmsstore.AdminCreateFaqParams{
		Question:  strings.TrimSpace(req.Question),
		Answer:    req.Answer,
		Category:  optText(req.Category),
		SortOrder: int32Value(req.SortOrder),
		Status:    enumValue(req.Status),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not create faq.")
		return
	}
	s.logAdminAction(c, actor, "cms_faq_created", "cms_faq", row.ID, map[string]any{"category": textValue(row.Category)})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminUpdateFaq(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "faq_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_faq_id", "Invalid faq id.")
		return
	}
	var req struct {
		Question  *string `json:"question"`
		Answer    *string `json:"answer"`
		Category  *string `json:"category"`
		SortOrder *int32  `json:"sort_order"`
		Status    *string `json:"status"`
	}
	_ = c.ShouldBindJSON(&req)
	row, err := s.cms.AdminUpdateFaq(c.Request.Context(), cmsstore.AdminUpdateFaqParams{
		ID:        id,
		Question:  optText(req.Question),
		Answer:    optText(req.Answer),
		Category:  optText(req.Category),
		SortOrder: int32Value(req.SortOrder),
		Status:    enumValue(req.Status),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update faq.")
		return
	}
	s.logAdminAction(c, actor, "cms_faq_updated", "cms_faq", row.ID, map[string]any{"category": textValue(row.Category)})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminDeleteFaq(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "faq_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_faq_id", "Invalid faq id.")
		return
	}
	if err := s.cms.AdminDeleteFaq(c.Request.Context(), id); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not delete faq.")
		return
	}
	s.logAdminAction(c, actor, "cms_faq_deleted", "cms_faq", id, map[string]any{})
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleAdminListAnnouncements(c *gin.Context) {
	p := listParams(c)
	items, err := s.cms.AdminListAnnouncements(c.Request.Context(), cmsstore.AdminListAnnouncementsParams{
		Status:   enumValue(strPtr(c.Query("status"))),
		Audience: text(strings.TrimSpace(c.Query("audience"))),
		Limit:    p.Limit,
		Offset:   p.Offset,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list announcements.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminGetAnnouncement(c *gin.Context) {
	id, ok := uuidParam(c, "announcement_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_announcement_id", "Invalid announcement id.")
		return
	}
	row, err := s.cms.AdminGetAnnouncementByID(c.Request.Context(), id)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Announcement not found.")
		return
	}
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminCreateAnnouncement(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	var req struct {
		Title    string     `json:"title" binding:"required"`
		Body     string     `json:"body" binding:"required"`
		Audience *string    `json:"audience"`
		Status   *string    `json:"status"`
		StartsAt *time.Time `json:"starts_at"`
		EndsAt   *time.Time `json:"ends_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid announcement request.")
		return
	}
	row, err := s.cms.AdminCreateAnnouncement(c.Request.Context(), cmsstore.AdminCreateAnnouncementParams{
		Title:    strings.TrimSpace(req.Title),
		Body:     req.Body,
		Audience: optText(req.Audience),
		Status:   enumValue(req.Status),
		StartsAt: tsValue(req.StartsAt),
		EndsAt:   tsValue(req.EndsAt),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not create announcement.")
		return
	}
	s.logAdminAction(c, actor, "cms_announcement_created", "cms_announcement", row.ID, map[string]any{"audience": row.Audience})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminUpdateAnnouncement(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "announcement_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_announcement_id", "Invalid announcement id.")
		return
	}
	var req struct {
		Title    *string    `json:"title"`
		Body     *string    `json:"body"`
		Audience *string    `json:"audience"`
		Status   *string    `json:"status"`
		StartsAt *time.Time `json:"starts_at"`
		EndsAt   *time.Time `json:"ends_at"`
	}
	_ = c.ShouldBindJSON(&req)
	row, err := s.cms.AdminUpdateAnnouncement(c.Request.Context(), cmsstore.AdminUpdateAnnouncementParams{
		ID:       id,
		Title:    optText(req.Title),
		Body:     optText(req.Body),
		Audience: optText(req.Audience),
		Status:   enumValue(req.Status),
		StartsAt: tsValue(req.StartsAt),
		EndsAt:   tsValue(req.EndsAt),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update announcement.")
		return
	}
	s.logAdminAction(c, actor, "cms_announcement_updated", "cms_announcement", row.ID, map[string]any{"audience": row.Audience})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminDeleteAnnouncement(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "announcement_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_announcement_id", "Invalid announcement id.")
		return
	}
	if err := s.cms.AdminDeleteAnnouncement(c.Request.Context(), id); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not delete announcement.")
		return
	}
	s.logAdminAction(c, actor, "cms_announcement_deleted", "cms_announcement", id, map[string]any{})
	c.Status(http.StatusNoContent)
}

type listPage struct {
	Limit  int32
	Offset int32
}

func listParams(c *gin.Context) listPage {
	page := parsePositiveInt32(c.DefaultQuery("page", "1"), 1)
	perPage := parsePositiveInt32(c.DefaultQuery("per_page", c.DefaultQuery("limit", "25")), 25)
	if perPage > 100 {
		perPage = 100
	}
	return listPage{Limit: perPage, Offset: (page - 1) * perPage}
}

func enumValue(v *string) any {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return s
}

func int32Value(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func tsValue(v *time.Time) pgtype.Timestamptz {
	if v == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: v.UTC(), Valid: true}
}

func strPtr(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func (s *Service) logAdminAction(c *gin.Context, actor uuid.UUID, action, entityType string, entityID pgtype.UUID, metadata map[string]any) {
	_, _ = s.systemQ.CreateAdminActivityLog(c.Request.Context(), systemstore.CreateAdminActivityLogParams{
		AdminUserID: pgxutil.UUID(actor),
		Action:      action,
		EntityType:  text(entityType),
		EntityID:    entityID,
		Metadata:    jsonObject(metadata),
		IpAddress:   ipAddrPtr(c.ClientIP()),
		UserAgent:   text(c.Request.UserAgent()),
	})
}

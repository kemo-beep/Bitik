package adminsvc

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	systemstore "github.com/bitik/backend/internal/store/system"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleAdminListModerationReports(c *gin.Context) {
	p := listParams(c)
	status := strings.TrimSpace(c.Query("status"))
	targetType := strings.TrimSpace(c.Query("target_type"))

	total, err := s.systemQ.AdminCountModerationReports(c.Request.Context(), struct {
		Status     interface{} `db:"status" json:"status"`
		TargetType pgtype.Text `db:"target_type" json:"target_type"`
	}{Status: enumAny(status), TargetType: text(targetType)})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list moderation reports.")
		return
	}
	items, err := s.systemQ.AdminListModerationReports(c.Request.Context(), systemstore.AdminListModerationReportsParams{
		Status:     enumAny(status),
		TargetType: text(targetType),
		Offset:     p.Offset,
		Limit:      p.Limit,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list moderation reports.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items, "pagination": gin.H{"page": (p.Offset / p.Limit) + 1, "per_page": p.Limit, "total": total}})
}

func (s *Service) HandleAdminGetModerationReport(c *gin.Context) {
	id, ok := uuidParam(c, "report_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_report_id", "Invalid report id.")
		return
	}
	row, err := s.systemQ.AdminGetModerationReportByID(c.Request.Context(), id)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Moderation report not found.")
		return
	}
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminUpdateModerationReportStatus(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "report_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_report_id", "Invalid report id.")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Status) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid report status request.")
		return
	}
	updated, err := s.systemQ.AdminUpdateModerationReportStatus(c.Request.Context(), systemstore.AdminUpdateModerationReportStatusParams{
		Status: strings.TrimSpace(req.Status),
		ID:     id,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update report status.")
		return
	}
	s.logAdminAction(c, actor, "moderation_report_status_updated", "moderation_report", updated.ID, map[string]any{"status": updated.Status})
	apiresponse.OK(c, updated)
}

func (s *Service) HandleAdminListModerationCases(c *gin.Context) {
	p := listParams(c)
	status := strings.TrimSpace(c.Query("status"))
	total, err := s.systemQ.AdminCountModerationCases(c.Request.Context(), enumAny(status))
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list moderation cases.")
		return
	}
	items, err := s.systemQ.AdminListModerationCases(c.Request.Context(), systemstore.AdminListModerationCasesParams{
		Status: enumAny(status),
		Offset: p.Offset,
		Limit:  p.Limit,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list moderation cases.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items, "pagination": gin.H{"page": (p.Offset / p.Limit) + 1, "per_page": p.Limit, "total": total}})
}

func (s *Service) HandleAdminGetModerationCase(c *gin.Context) {
	id, ok := uuidParam(c, "case_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_case_id", "Invalid case id.")
		return
	}
	row, err := s.systemQ.AdminGetModerationCaseByID(c.Request.Context(), id)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Moderation case not found.")
		return
	}
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminCreateModerationCase(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	var req struct {
		ReportID   string  `json:"report_id" binding:"required"`
		AssignedTo *string `json:"assigned_to"`
		Status     *string `json:"status"`
		Resolution *string `json:"resolution"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid moderation case request.")
		return
	}
	reportID, err := uuid.Parse(strings.TrimSpace(req.ReportID))
	if err != nil {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "report_id", Message: "must be a valid uuid"}})
		return
	}
	assignedTo, ok := parseOptionalUUIDStrict(req.AssignedTo)
	if !ok {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "assigned_to", Message: "must be a valid uuid"}})
		return
	}
	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create moderation case.")
		return
	}
	defer tx.Rollback(ctx)

	q := s.systemQ.WithTx(tx)
	row, err := q.AdminCreateModerationCase(ctx, systemstore.AdminCreateModerationCaseParams{
		ReportID:   pgxutil.UUID(reportID),
		AssignedTo: assignedTo,
		Status:     enumAnyPtr(req.Status),
		Resolution: optText(req.Resolution),
		ResolvedAt: pgtype.Timestamptz{},
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not create moderation case.")
		return
	}
	if mapped := reportStatusFromCaseStatus(statusString(row.Status)); mapped != "" {
		if _, err := q.AdminUpdateModerationReportStatus(ctx, systemstore.AdminUpdateModerationReportStatusParams{
			Status: mapped,
			ID:     row.ReportID,
		}); err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not sync moderation report status.")
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create moderation case.")
		return
	}
	s.logAdminAction(c, actor, "moderation_case_created", "moderation_case", row.ID, map[string]any{"report_id": req.ReportID})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminUpdateModerationCase(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	id, ok := uuidParam(c, "case_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_case_id", "Invalid case id.")
		return
	}
	var req struct {
		AssignedTo *string    `json:"assigned_to"`
		Status     *string    `json:"status"`
		Resolution *string    `json:"resolution"`
		ResolvedAt *time.Time `json:"resolved_at"`
	}
	_ = c.ShouldBindJSON(&req)
	assignedTo, valid := parseOptionalUUIDStrict(req.AssignedTo)
	if !valid {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "assigned_to", Message: "must be a valid uuid"}})
		return
	}
	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update moderation case.")
		return
	}
	defer tx.Rollback(ctx)
	q := s.systemQ.WithTx(tx)
	row, err := q.AdminUpdateModerationCase(ctx, systemstore.AdminUpdateModerationCaseParams{
		AssignedTo: assignedTo,
		Status:     enumAnyPtr(req.Status),
		Resolution: optText(req.Resolution),
		ResolvedAt: tsValue(req.ResolvedAt),
		ID:         id,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update moderation case.")
		return
	}
	if mapped := reportStatusFromCaseStatus(statusString(row.Status)); mapped != "" {
		if _, err := q.AdminUpdateModerationReportStatus(ctx, systemstore.AdminUpdateModerationReportStatusParams{
			Status: mapped,
			ID:     row.ReportID,
		}); err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not sync moderation report status.")
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update moderation case.")
		return
	}
	s.logAdminAction(c, actor, "moderation_case_updated", "moderation_case", row.ID, map[string]any{"status": row.Status})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminListPlatformSettings(c *gin.Context) {
	items, err := s.systemQ.ListPlatformSettings(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list platform settings.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminUpsertPlatformSetting(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_key", "Invalid platform setting key.")
		return
	}
	var req struct {
		Value       any     `json:"value" binding:"required"`
		Description *string `json:"description"`
		IsPublic    bool    `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid platform setting request.")
		return
	}
	row, err := s.systemQ.UpsertPlatformSetting(c.Request.Context(), systemstore.UpsertPlatformSettingParams{
		Key:         key,
		Value:       jsonObject(req.Value),
		Description: optText(req.Description),
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update platform setting.")
		return
	}
	s.logAdminAction(c, actor, "platform_setting_upserted", "platform_setting", pgtype.UUID{}, map[string]any{"key": row.Key, "is_public": row.IsPublic})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminListFeatureFlags(c *gin.Context) {
	items, err := s.systemQ.ListFeatureFlags(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list feature flags.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminUpsertFeatureFlag(c *gin.Context) {
	actor, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing auth context.")
		return
	}
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_key", "Invalid feature flag key.")
		return
	}
	var req struct {
		Description *string        `json:"description"`
		Enabled     bool           `json:"enabled"`
		Rules       map[string]any `json:"rules"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid feature flag request.")
		return
	}
	row, err := s.systemQ.UpsertFeatureFlag(c.Request.Context(), systemstore.UpsertFeatureFlagParams{
		Key:         key,
		Description: optText(req.Description),
		Enabled:     req.Enabled,
		Rules:       jsonObject(req.Rules),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Could not update feature flag.")
		return
	}
	s.logAdminAction(c, actor, "feature_flag_upserted", "feature_flag", pgtype.UUID{}, map[string]any{"key": row.Key, "enabled": row.Enabled})
	apiresponse.OK(c, row)
}

func (s *Service) HandleAdminListAuditLogs(c *gin.Context) {
	p := listParams(c)
	actorID, ok := parseOptionalUUIDStrict(strPtr(c.Query("actor_user_id")))
	if !ok {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "actor_user_id", Message: "must be a valid uuid"}})
		return
	}
	fromTime := parseOptionalTime(c.Query("from"))
	toTime := parseOptionalTime(c.Query("to"))

	total, err := s.systemQ.AdminCountAuditLogs(c.Request.Context(), systemstore.AdminCountAuditLogsParams{
		ActorUserID: actorID,
		Action:      text(c.Query("action")),
		EntityType:  text(c.Query("entity_type")),
		FromTime:    fromTime,
		ToTime:      toTime,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list audit logs.")
		return
	}
	items, err := s.systemQ.AdminListAuditLogs(c.Request.Context(), systemstore.AdminListAuditLogsParams{
		ActorUserID: actorID,
		Action:      text(c.Query("action")),
		EntityType:  text(c.Query("entity_type")),
		FromTime:    fromTime,
		ToTime:      toTime,
		Offset:      p.Offset,
		Limit:       p.Limit,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list audit logs.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items, "pagination": gin.H{"page": (p.Offset / p.Limit) + 1, "per_page": p.Limit, "total": total}})
}

func (s *Service) HandleAdminListAdminActivityLogs(c *gin.Context) {
	p := listParams(c)
	adminID, ok := parseOptionalUUIDStrict(strPtr(c.Query("admin_user_id")))
	if !ok {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "admin_user_id", Message: "must be a valid uuid"}})
		return
	}
	fromTime := parseOptionalTime(c.Query("from"))
	toTime := parseOptionalTime(c.Query("to"))

	total, err := s.systemQ.AdminCountAdminActivityLogs(c.Request.Context(), systemstore.AdminCountAdminActivityLogsParams{
		AdminUserID: adminID,
		Action:      text(c.Query("action")),
		EntityType:  text(c.Query("entity_type")),
		FromTime:    fromTime,
		ToTime:      toTime,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list admin activity logs.")
		return
	}
	items, err := s.systemQ.AdminListAdminActivityLogs(c.Request.Context(), systemstore.AdminListAdminActivityLogsParams{
		AdminUserID: adminID,
		Action:      text(c.Query("action")),
		EntityType:  text(c.Query("entity_type")),
		FromTime:    fromTime,
		ToTime:      toTime,
		Offset:      p.Offset,
		Limit:       p.Limit,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list admin activity logs.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items, "pagination": gin.H{"page": (p.Offset / p.Limit) + 1, "per_page": p.Limit, "total": total}})
}

func (s *Service) HandleIngestAnalyticsEvent(c *gin.Context) {
	var req struct {
		EventName  string         `json:"event_name" binding:"required"`
		EntityType *string        `json:"entity_type"`
		EntityID   *string        `json:"entity_id"`
		Metadata   map[string]any `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.EventName) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid analytics event payload.")
		return
	}
	var userID pgtype.UUID
	if uid, ok := currentUserID(c); ok {
		userID = pgxutil.UUID(uid)
	}
	var entityID pgtype.UUID
	if req.EntityID != nil {
		var ok bool
		entityID, ok = parseOptionalUUIDStrict(req.EntityID)
		if !ok {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "entity_id", Message: "must be a valid uuid"}})
			return
		}
	}
	ev, err := s.systemQ.CreateEventLog(c.Request.Context(), systemstore.CreateEventLogParams{
		UserID:     userID,
		EventName:  strings.TrimSpace(req.EventName),
		EntityType: optText(req.EntityType),
		EntityID:   entityID,
		Metadata:   jsonObject(req.Metadata),
		IpAddress:  ipAddrPtr(c.ClientIP()),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not ingest analytics event.")
		return
	}
	_ = s.systemQ.EnqueueAnalyticsEvent(c.Request.Context(), ev.ID)
	c.Status(http.StatusAccepted)
}

func (s *Service) HandleProcessAnalyticsEvents(c *gin.Context) {
	limit := parsePositiveInt32(c.DefaultQuery("limit", "100"), 100)
	if limit > 1000 {
		limit = 1000
	}
	ctx := c.Request.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not start analytics processing.")
		return
	}
	defer tx.Rollback(ctx)

	q := s.systemQ.WithTx(tx)
	rows, err := q.DequeueAnalyticsEvents(ctx, limit)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not dequeue analytics events.")
		return
	}

	processed := int32(0)
	failed := int32(0)
	for _, row := range rows {
		eventLog, err := q.GetEventLogByID(ctx, row.EventLogID)
		if err != nil {
			_ = q.MarkAnalyticsEventFailed(ctx, systemstore.MarkAnalyticsEventFailedParams{
				LastError: text("event_not_found"),
				RetryAfter: pgtype.Interval{
					Microseconds: int64((time.Minute).Microseconds()),
					Valid:        true,
				},
				ID: row.ID,
			})
			failed++
			continue
		}
		agg, err := q.AggregateEventLogDay(ctx, systemstore.AggregateEventLogDayParams{
			EventName: eventLog.EventName,
			Day:       pgtype.Timestamptz{Time: eventLog.CreatedAt.Time.UTC(), Valid: true},
		})
		if err != nil {
			_ = markAnalyticsFailed(q, ctx, row.ID, "aggregate_failed")
			failed++
			continue
		}
		if err := q.UpsertAdminMetricDaily(ctx, systemstore.UpsertAdminMetricDailyParams{
			MetricDate:  pgtype.Date{Time: eventLog.CreatedAt.Time.UTC(), Valid: true},
			EventName:   eventLog.EventName,
			TotalCount:  agg.TotalCount,
			UniqueUsers: agg.UniqueUsers,
		}); err != nil {
			_ = markAnalyticsFailed(q, ctx, row.ID, "rollup_upsert_failed")
			failed++
			continue
		}
		if err := q.MarkAnalyticsEventProcessed(ctx, row.ID); err != nil {
			_ = markAnalyticsFailed(q, ctx, row.ID, "mark_processed_failed")
			failed++
			continue
		}
		processed++
	}

	if err := tx.Commit(ctx); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not finish analytics processing.")
		return
	}
	apiresponse.OK(c, gin.H{"dequeued": len(rows), "processed": processed, "failed": failed})
}

func (s *Service) HandleAdminDashboardOverview(c *gin.Context) {
	overview, err := s.systemQ.AdminDashboardOverview(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load dashboard overview.")
		return
	}
	apiresponse.OK(c, overview)
}

func (s *Service) HandleAdminDashboardEventChart(c *gin.Context) {
	days := parsePositiveInt32(c.DefaultQuery("days", "14"), 14)
	if days > 90 {
		days = 90
	}
	to := time.Now().UTC().Truncate(24 * time.Hour)
	from := to.AddDate(0, 0, -int(days)+1)
	items, err := s.systemQ.ListAdminMetricsDaily(c.Request.Context(), systemstore.ListAdminMetricsDailyParams{
		FromDate: pgtype.Date{Time: from, Valid: true},
		ToDate:   pgtype.Date{Time: to, Valid: true},
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load dashboard charts.")
		return
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func enumAny(v string) any {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return v
}

func enumAnyPtr(v *string) any {
	if v == nil {
		return nil
	}
	return enumAny(*v)
}

func parseOptionalUUIDStrict(v *string) (pgtype.UUID, bool) {
	if v == nil {
		return pgtype.UUID{}, true
	}
	id, err := uuid.Parse(strings.TrimSpace(*v))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func parseOptionalTime(raw string) pgtype.Timestamptz {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Timestamptz{}
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func reportStatusFromCaseStatus(caseStatus string) string {
	switch strings.TrimSpace(strings.ToLower(caseStatus)) {
	case "under_review":
		return "under_review"
	case "resolved":
		return "resolved"
	case "dismissed":
		return "dismissed"
	default:
		return ""
	}
}

func markAnalyticsFailed(q *systemstore.Queries, ctx context.Context, queueID pgtype.UUID, reason string) error {
	return q.MarkAnalyticsEventFailed(ctx, systemstore.MarkAnalyticsEventFailedParams{
		LastError: text(reason),
		RetryAfter: pgtype.Interval{
			Microseconds: int64((time.Minute).Microseconds()),
			Valid:        true,
		},
		ID: queueID,
	})
}

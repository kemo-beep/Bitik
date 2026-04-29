package reviewsvc

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/notify"
	"github.com/bitik/backend/internal/pgxutil"
	orderstore "github.com/bitik/backend/internal/store/orders"
	reviewstore "github.com/bitik/backend/internal/store/reviews"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleBuyerCreateReview(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		ProductID   string `json:"product_id" binding:"required"`
		OrderItemID string `json:"order_item_id"`
		Rating      int32  `json:"rating" binding:"required"`
		Title       string `json:"title"`
		Body        string `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid review request.")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "rating", Message: "must be between 1 and 5"}})
		return
	}
	pid, err := uuid.Parse(req.ProductID)
	if err != nil {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "product_id", Message: "must be a valid uuid"}})
		return
	}
	var orderItemID pgtype.UUID
	verified := false
	if strings.TrimSpace(req.OrderItemID) != "" {
		oid, err := uuid.Parse(req.OrderItemID)
		if err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "order_item_id", Message: "must be a valid uuid"}})
			return
		}
		orderItemID = pgxutil.UUID(oid)
		row, err := s.reviews.GetOrderItemForReviewVerification(c.Request.Context(), reviewstore.GetOrderItemForReviewVerificationParams{
			ID:     orderItemID,
			UserID: pgxutil.UUID(uid),
		})
		if err == nil && pgxutil.UUID(pid) == row.ProductID {
			verified = true
		}
	}
	created, err := s.reviews.CreateReview(c.Request.Context(), reviewstore.CreateReviewParams{
		ProductID:          pgxutil.UUID(pid),
		OrderItemID:        orderItemID,
		UserID:             pgxutil.UUID(uid),
		Rating:             req.Rating,
		Title:              text(req.Title),
		Body:               text(req.Body),
		IsVerifiedPurchase: verified,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "review_create_failed", "Could not create review.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, reviewJSON(created), nil)
}

func (s *Service) HandleBuyerUpdateReview(c *gin.Context) {
	uid, _ := currentUserID(c)
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	var req struct {
		Rating *int32  `json:"rating"`
		Title  *string `json:"title"`
		Body   *string `json:"body"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid review update.")
		return
	}
	if req.Rating != nil && (*req.Rating < 1 || *req.Rating > 5) {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "rating", Message: "must be between 1 and 5"}})
		return
	}
	updated, err := s.reviews.UpdateReviewForUser(c.Request.Context(), reviewstore.UpdateReviewForUserParams{
		ID:     rid,
		UserID: pgxutil.UUID(uid),
		Rating: int4(req.Rating),
		Title:  textPtr(req.Title),
		Body:   textPtr(req.Body),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "review_not_found", "Review not found.")
		return
	}
	apiresponse.OK(c, reviewJSON(updated))
}

func (s *Service) HandleBuyerDeleteReview(c *gin.Context) {
	uid, _ := currentUserID(c)
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	if err := s.reviews.SoftDeleteReviewForUser(c.Request.Context(), reviewstore.SoftDeleteReviewForUserParams{ID: rid, UserID: pgxutil.UUID(uid)}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete review.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleListReviewImages(c *gin.Context) {
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	rows, err := s.reviews.ListReviewImages(c.Request.Context(), rid)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load images.")
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{"id": uuidString(r.ID), "url": r.Url, "sort_order": r.SortOrder})
	}
	apiresponse.OK(c, gin.H{"items": out})
}

func (s *Service) HandleAddReviewImage(c *gin.Context) {
	uid, _ := currentUserID(c)
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	if _, err := s.reviews.GetReviewForUser(c.Request.Context(), reviewstore.GetReviewForUserParams{ID: rid, UserID: pgxutil.UUID(uid)}); err != nil {
		apiresponse.Error(c, http.StatusNotFound, "review_not_found", "Review not found.")
		return
	}
	var req struct {
		URL       string `json:"url" binding:"required"`
		SortOrder int32  `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid image request.")
		return
	}
	img, err := s.reviews.AddReviewImage(c.Request.Context(), reviewstore.AddReviewImageParams{ReviewID: rid, Url: req.URL, SortOrder: req.SortOrder})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "image_add_failed", "Could not add image.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, gin.H{"id": uuidString(img.ID), "url": img.Url, "sort_order": img.SortOrder}, nil)
}

func (s *Service) HandleDeleteReviewImage(c *gin.Context) {
	uid, _ := currentUserID(c)
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	iid, ok := uuidParam(c, "image_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_image_id", "Invalid image id.")
		return
	}
	if _, err := s.reviews.GetReviewForUser(c.Request.Context(), reviewstore.GetReviewForUserParams{ID: rid, UserID: pgxutil.UUID(uid)}); err != nil {
		apiresponse.Error(c, http.StatusNotFound, "review_not_found", "Review not found.")
		return
	}
	if err := s.reviews.DeleteReviewImage(c.Request.Context(), reviewstore.DeleteReviewImageParams{ID: iid, ReviewID: rid}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete image.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleVoteReview(c *gin.Context) {
	uid, _ := currentUserID(c)
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	var req struct {
		Vote int32 `json:"vote" binding:"required"` // -1, 0, 1
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid vote request.")
		return
	}
	if req.Vote == 0 {
		_ = s.reviews.DeleteReviewVote(c.Request.Context(), reviewstore.DeleteReviewVoteParams{ReviewID: rid, UserID: pgxutil.UUID(uid)})
		c.Status(http.StatusNoContent)
		return
	}
	if req.Vote != -1 && req.Vote != 1 {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "vote", Message: "must be -1, 0, or 1"}})
		return
	}
	v, err := s.reviews.UpsertReviewVote(c.Request.Context(), reviewstore.UpsertReviewVoteParams{ReviewID: rid, UserID: pgxutil.UUID(uid), Vote: req.Vote})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "vote_failed", "Could not vote.")
		return
	}
	apiresponse.OK(c, gin.H{"review_id": uuidString(v.ReviewID), "vote": v.Vote})
}

func (s *Service) HandleReportReview(c *gin.Context) {
	uid, _ := currentUserID(c)
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid report request.")
		return
	}
	report, err := s.reviews.CreateReviewReport(c.Request.Context(), reviewstore.CreateReviewReportParams{
		ReviewID:        rid,
		ReporterUserID:  pgxutil.UUID(uid),
		Reason:          req.Reason,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "report_failed", "Could not report review.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, gin.H{"id": uuidString(report.ID), "status": report.Status}, nil)
}

func (s *Service) HandleSellerReply(c *gin.Context) {
	seller, ok := sellerFromContext(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing seller context.")
		return
	}
	uid, _ := currentUserID(c)
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	var req struct {
		Body string `json:"body" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Body) == "" {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid reply request.")
		return
	}
	updated, err := s.reviews.SetSellerReply(c.Request.Context(), reviewstore.SetSellerReplyParams{
		ID:           rid,
		SellerReply:  text(req.Body),
		SellerReplyBy: pgxutil.UUID(uid),
		SellerID:     seller.ID,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "review_not_found", "Review not found.")
		return
	}
	if s.pub != nil {
		if v, ok := pgxutil.ToUUID(updated.UserID); ok {
			s.pub.Publish(c.Request.Context(), notify.Event{
				Type:   notify.EventNotificationCreated,
				UserID: v.String(),
				Data: map[string]any{
					"type":      "review_reply",
					"review_id": uuidString(updated.ID),
					"product_id": uuidString(updated.ProductID),
				},
			})
		}
	}
	apiresponse.OK(c, reviewJSON(updated))
}

func (s *Service) HandleSellerReportReview(c *gin.Context) {
	// Same as buyer report for now.
	s.HandleReportReview(c)
}

func (s *Service) HandleAdminHideReview(c *gin.Context) {
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	var req struct {
		Hidden bool `json:"hidden"`
	}
	_ = c.ShouldBindJSON(&req)
	updated, err := s.reviews.HideReviewAdmin(c.Request.Context(), reviewstore.HideReviewAdminParams{ID: rid, IsHidden: req.Hidden})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "hide_failed", "Could not update review visibility.")
		return
	}
	apiresponse.OK(c, reviewJSON(updated))
}

func (s *Service) HandleAdminDeleteReview(c *gin.Context) {
	rid, ok := uuidParam(c, "review_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_review_id", "Invalid review id.")
		return
	}
	if err := s.reviews.SoftDeleteReviewAdmin(c.Request.Context(), rid); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete review.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleAdminListOpenReports(c *gin.Context) {
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
	offset := (page - 1) * limit
	rows, err := s.reviews.ListOpenReviewReports(c.Request.Context(), reviewstore.ListOpenReviewReportsParams{Limit: limit, Offset: offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load reports.")
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":         uuidString(r.ID),
			"review_id":  uuidString(r.ReviewID),
			"reason":     r.Reason,
			"status":     r.Status,
			"created_at": r.CreatedAt.Time,
		})
	}
	apiresponse.OK(c, gin.H{"items": out})
}

func (s *Service) HandleAdminResolveReport(c *gin.Context) {
	uid, _ := currentUserID(c)
	rid, ok := uuidParam(c, "report_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_report_id", "Invalid report id.")
		return
	}
	updated, err := s.reviews.ResolveReviewReport(c.Request.Context(), reviewstore.ResolveReviewReportParams{ID: rid, ResolvedBy: pgxutil.UUID(uid)})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "resolve_failed", "Could not resolve report.")
		return
	}
	apiresponse.OK(c, gin.H{"id": uuidString(updated.ID), "status": updated.Status})
}

// helpers

func uuidParam(c *gin.Context, name string) (pgtype.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func uuidString(id pgtype.UUID) string {
	if v, ok := pgxutil.ToUUID(id); ok {
		return v.String()
	}
	return ""
}

func text(raw string) pgtype.Text {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: raw, Valid: true}
}

func textPtr(raw *string) pgtype.Text {
	if raw == nil {
		return pgtype.Text{}
	}
	return text(*raw)
}

func int4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func reviewJSON(r reviewstore.ProductReview) gin.H {
	return gin.H{
		"id":                   uuidString(r.ID),
		"product_id":           uuidString(r.ProductID),
		"user_id":              uuidString(r.UserID),
		"order_item_id":        nullableUUID(r.OrderItemID),
		"rating":               r.Rating,
		"title":                textValue(r.Title),
		"body":                 textValue(r.Body),
		"is_verified_purchase": r.IsVerifiedPurchase,
		"is_hidden":            r.IsHidden,
		"seller_reply":         textValue(r.SellerReply),
		"seller_reply_at":      timeValue(r.SellerReplyAt),
		"created_at":           r.CreatedAt.Time,
		"updated_at":           r.UpdatedAt.Time,
	}
}

func nullableUUID(id pgtype.UUID) any {
	if !id.Valid {
		return nil
	}
	return uuidString(id)
}

func textValue(t pgtype.Text) any {
	if !t.Valid {
		return nil
	}
	return t.String
}

func timeValue(t pgtype.Timestamptz) any {
	if !t.Valid {
		return nil
	}
	return t.Time
}

var _ = orderstore.Seller{} // keep import tidy if needed


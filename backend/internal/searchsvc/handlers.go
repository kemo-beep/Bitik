package searchsvc

import (
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	platformsearch "github.com/bitik/backend/internal/platform/search"
	catalogstore "github.com/bitik/backend/internal/store/catalog"
	searchstore "github.com/bitik/backend/internal/store/search"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultPage  = int32(1)
	defaultLimit = int32(20)
	maxLimit     = int32(100)
)

type pageParams struct {
	Page  int32
	Limit int32
	From  int
}

func parsePage(c *gin.Context) pageParams {
	page := parsePositiveInt32(c.DefaultQuery("page", "1"), defaultPage)
	limit := parsePositiveInt32(c.DefaultQuery("limit", c.DefaultQuery("per_page", "20")), defaultLimit)
	if limit > maxLimit {
		limit = maxLimit
	}
	return pageParams{Page: page, Limit: limit, From: int((page - 1) * limit)}
}

func parsePositiveInt32(raw string, fallback int32) int32 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 32)
	if err != nil || v < 1 {
		return fallback
	}
	return int32(v)
}

func pageMeta(p pageParams, total int64) map[string]any {
	hasNext := int64(p.Page)*int64(p.Limit) < total
	return map[string]any{
		"page":     p.Page,
		"limit":    p.Limit,
		"total":    total,
		"has_next": hasNext,
	}
}

func (s *Service) HandleSearch(c *gin.Context) {
	p := parsePage(c)
	q := strings.TrimSpace(c.Query("q"))
	sort := strings.TrimSpace(c.DefaultQuery("sort", "popular"))

	var categoryID, brandID, sellerID string
	if raw := strings.TrimSpace(c.Query("category_id")); raw != "" {
		if _, err := uuid.Parse(raw); err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "category_id", Message: "must be a valid uuid"}})
			return
		}
		categoryID = raw
	}
	if raw := strings.TrimSpace(c.Query("brand_id")); raw != "" {
		if _, err := uuid.Parse(raw); err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "brand_id", Message: "must be a valid uuid"}})
			return
		}
		brandID = raw
	}
	if raw := strings.TrimSpace(c.Query("seller_id")); raw != "" {
		if _, err := uuid.Parse(raw); err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "seller_id", Message: "must be a valid uuid"}})
			return
		}
		sellerID = raw
	}

	var minPrice, maxPrice *int64
	if raw := strings.TrimSpace(c.Query("min_price_cents")); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v < 0 {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "min_price_cents", Message: "must be a non-negative integer"}})
			return
		}
		minPrice = &v
	}
	if raw := strings.TrimSpace(c.Query("max_price_cents")); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v < 0 {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "max_price_cents", Message: "must be a non-negative integer"}})
			return
		}
		maxPrice = &v
	}
	if minPrice != nil && maxPrice != nil && *minPrice > *maxPrice {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "min_price_cents", Message: "must be <= max_price_cents"}})
		return
	}

	searchBackend := "opensearch"
	var items []gin.H
	var total int64

	if s.os != nil {
		resp, err := s.os.SearchProducts(c.Request.Context(), platformsearch.SearchRequest{
			Query:         q,
			CategoryID:    categoryID,
			BrandID:       brandID,
			SellerID:      sellerID,
			MinPriceCents: minPrice,
			MaxPriceCents: maxPrice,
			Sort:          sort,
			From:          p.From,
			Size:          int(p.Limit),
		})
		if err == nil {
			total = resp.Total
			items = make([]gin.H, 0, len(resp.Hits))
			for _, h := range resp.Hits {
				items = append(items, gin.H{
					"id":              h.ID,
					"seller_id":       h.SellerID,
					"category_id":     nullableString(h.CategoryID),
					"brand_id":        nullableString(h.BrandID),
					"name":            h.Name,
					"slug":            h.Slug,
					"min_price_cents": h.MinPriceCents,
					"max_price_cents": h.MaxPriceCents,
					"currency":        h.Currency,
					"total_sold":      h.TotalSold,
					"rating":          h.Rating,
					"review_count":    h.ReviewCount,
					"image_url":       nullableString(h.ImageURL),
				})
			}
		} else {
			searchBackend = "postgres"
		}
	} else {
		searchBackend = "postgres"
	}

	if searchBackend == "postgres" {
		params, ok := s.postgresProductParams(c, p, sort)
		if !ok {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "filters", Message: "invalid filters"}})
			return
		}
		totalCount, err := s.catalog.CountPublicProducts(c.Request.Context(), catalogstore.CountPublicProductsParams{
			CategoryID:    params.CategoryID,
			BrandID:       params.BrandID,
			SellerID:      params.SellerID,
			MinPriceCents: params.MinPriceCents,
			MaxPriceCents: params.MaxPriceCents,
			Query:         params.Query,
		})
		if err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not search products.")
			return
		}
		rows, err := s.catalog.ListProducts(c.Request.Context(), params)
		if err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not search products.")
			return
		}
		total = totalCount
		items = make([]gin.H, 0, len(rows))
		for _, r := range rows {
			items = append(items, productRowJSON(r))
		}
	}

	// Track recent query best-effort (don’t block response).
	s.trackRecentBestEffort(c, q, gin.H{
		"category_id":     categoryID,
		"brand_id":        brandID,
		"seller_id":       sellerID,
		"sort":            sort,
		"min_price_cents": minPrice,
		"max_price_cents": maxPrice,
	})

	apiresponse.Respond(c, http.StatusOK, gin.H{"items": items}, map[string]any{
		"pagination":     pageMeta(p, total),
		"search_backend": searchBackend,
	})
}

func (s *Service) HandleSuggestions(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		apiresponse.OK(c, gin.H{"items": []any{}})
		return
	}
	// Minimal v1: return top matching product names.
	items := make([]any, 0, 10)
	if s.os != nil {
		resp, err := s.os.SearchProducts(c.Request.Context(), platformsearch.SearchRequest{
			Query: q,
			From:  0,
			Size:  10,
		})
		if err == nil {
			seen := map[string]struct{}{}
			for _, h := range resp.Hits {
				name := strings.TrimSpace(h.Name)
				if name == "" {
					continue
				}
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				items = append(items, name)
				if len(items) >= 10 {
					break
				}
			}
		}
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleTrending(c *gin.Context) {
	window := strings.TrimSpace(c.DefaultQuery("window", "24h"))
	limit := parsePositiveInt32(c.DefaultQuery("limit", "10"), 10)
	if limit > 50 {
		limit = 50
	}
	interval := parseInterval(window)
	rows, err := s.search.ListTrendingSearchQueries(c.Request.Context(), searchstore.ListTrendingSearchQueriesParams{
		Window: interval,
		Limit:  limit,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load trending searches.")
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{"query": r.Query, "clicks": r.Clicks})
	}
	apiresponse.OK(c, gin.H{"items": out})
}

func (s *Service) HandleRecent(c *gin.Context) {
	limit := parsePositiveInt32(c.DefaultQuery("limit", "10"), 10)
	if limit > 50 {
		limit = 50
	}
	if uid, ok := currentUserID(c); ok {
		rows, err := s.search.ListRecentSearchQueriesForUser(c.Request.Context(), searchstore.ListRecentSearchQueriesForUserParams{
			UserID: pgxutil.UUID(uid),
			Limit:  limit,
		})
		if err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load recent searches.")
			return
		}
		apiresponse.OK(c, gin.H{"items": mapRecent(rows)})
		return
	}
	sid := sessionID(c)
	if sid == "" {
		apiresponse.OK(c, gin.H{"items": []any{}})
		return
	}
	rows, err := s.search.ListRecentSearchQueriesForSession(c.Request.Context(), searchstore.ListRecentSearchQueriesForSessionParams{
		SessionID: text(sid),
		Limit:     limit,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load recent searches.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapRecent(rows)})
}

func (s *Service) HandleClearRecent(c *gin.Context) {
	if uid, ok := currentUserID(c); ok {
		if err := s.search.ClearRecentSearchQueriesForUser(c.Request.Context(), pgxutil.UUID(uid)); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not clear recent searches.")
			return
		}
		c.Status(http.StatusNoContent)
		return
	}
	sid := sessionID(c)
	if sid == "" {
		c.Status(http.StatusNoContent)
		return
	}
	if err := s.search.ClearRecentSearchQueriesForSession(c.Request.Context(), text(sid)); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not clear recent searches.")
		return
	}
	c.Status(http.StatusNoContent)
}

func text(v string) pgtype.Text {
	v = strings.TrimSpace(v)
	return pgtype.Text{String: v, Valid: v != ""}
}

func parseInterval(raw string) pgtype.Interval {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Interval{}
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		// default 24h
		d = 24 * time.Hour
	}
	return pgtype.Interval{Microseconds: d.Microseconds(), Valid: true}
}

func (s *Service) HandleClick(c *gin.Context) {
	var req struct {
		Query     string         `json:"query" binding:"required"`
		ProductID string         `json:"product_id"`
		Position  *int32         `json:"position"`
		Metadata  map[string]any `json:"metadata"`
		SessionID string         `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid click request.")
		return
	}
	var productID pgtype.UUID
	if strings.TrimSpace(req.ProductID) != "" {
		id, err := uuid.Parse(req.ProductID)
		if err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "product_id", Message: "must be a valid uuid"}})
			return
		}
		productID = pgxutil.UUID(id)
	}
	var position pgtype.Int4
	if req.Position != nil && *req.Position > 0 {
		position = pgtype.Int4{Int32: *req.Position, Valid: true}
	}
	meta := jsonObject(req.Metadata)
	sid := strings.TrimSpace(req.SessionID)
	if sid == "" {
		sid = sessionID(c)
	}
	var uid pgtype.UUID
	if user, ok := currentUserID(c); ok {
		uid = pgxutil.UUID(user)
	}
	_, err := s.catalog.TrackSearchClick(c.Request.Context(), catalogstore.TrackSearchClickParams{
		UserID:    uid,
		SessionID: textParam(sid),
		Query:     req.Query,
		ProductID: productID,
		Position:  position,
		Metadata:  meta,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not track click.")
		return
	}
	c.Status(http.StatusNoContent)
}

// Internal jobs

func (s *Service) HandleIndexProduct(c *gin.Context) {
	if s.os == nil {
		apiresponse.Error(c, http.StatusServiceUnavailable, "search_unavailable", "Search backend is unavailable.")
		return
	}
	var req struct {
		ProductID string `json:"product_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid index request.")
		return
	}
	pid, err := uuid.Parse(req.ProductID)
	if err != nil {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "product_id", Message: "must be a valid uuid"}})
		return
	}
	if err := s.indexProductByID(c, pid); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "index_failed", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleReindexProducts(c *gin.Context) {
	if s.os == nil {
		apiresponse.Error(c, http.StatusServiceUnavailable, "search_unavailable", "Search backend is unavailable.")
		return
	}
	var req struct {
		Limit           int32  `json:"limit"`
		CursorID        string `json:"cursor_id"`
		CursorUpdatedAt string `json:"cursor_updated_at"`
	}
	_ = c.ShouldBindJSON(&req)
	limit := req.Limit
	if limit < 1 {
		limit = 500
	}
	if limit > 5000 {
		limit = 5000
	}

	var cursorID pgtype.UUID
	if strings.TrimSpace(req.CursorID) != "" {
		id, err := uuid.Parse(req.CursorID)
		if err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "cursor_id", Message: "must be a valid uuid"}})
			return
		}
		cursorID = pgxutil.UUID(id)
	}
	var cursorUpdatedAt pgtype.Timestamptz
	if strings.TrimSpace(req.CursorUpdatedAt) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(req.CursorUpdatedAt))
		if err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "cursor_updated_at", Message: "must be RFC3339 timestamp"}})
			return
		}
		cursorUpdatedAt = pgtype.Timestamptz{Time: t, Valid: true}
	}

	rows, err := s.search.ListIndexableProducts(c.Request.Context(), searchstore.ListIndexableProductsParams{
		Limit:           limit,
		CursorUpdatedAt: cursorUpdatedAt,
		CursorID:        cursorID,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list products for reindex.")
		return
	}

	indexed := int32(0)
	var nextCursorID string
	var nextCursorUpdatedAt string
	for _, r := range rows {
		_ = s.indexProductRow(c, r)
		indexed++
		nextCursorID = uuidString(r.ID)
		nextCursorUpdatedAt = r.UpdatedAt.Time.UTC().Format(time.RFC3339)
	}
	apiresponse.OK(c, gin.H{
		"indexed": indexed,
		"next_cursor": gin.H{
			"id":         nextCursorID,
			"updated_at": nextCursorUpdatedAt,
		},
	})
}

// helpers

func nullableString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func sessionID(c *gin.Context) string {
	if v := strings.TrimSpace(c.GetHeader("X-Device-Id")); v != "" {
		return v
	}
	return strings.TrimSpace(c.GetHeader("X-Request-ID"))
}

func (s *Service) trackRecentBestEffort(c *gin.Context, q string, filters map[string]any) {
	q = strings.TrimSpace(q)
	if q == "" {
		return
	}
	sid := sessionID(c)
	var uid pgtype.UUID
	if user, ok := currentUserID(c); ok {
		uid = pgxutil.UUID(user)
	}
	_, _ = s.catalog.TrackSearchQuery(c.Request.Context(), catalogstore.TrackSearchQueryParams{
		UserID:    uid,
		SessionID: textParam(sid),
		Query:     q,
		Filters:   jsonObject(filters),
	})
}

func mapRecent(rows []searchstore.SearchRecentQuery) []gin.H {
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":         uuidString(r.ID),
			"query":      r.Query,
			"filters":    r.Filters,
			"created_at": r.CreatedAt.Time,
		})
	}
	return out
}

func textParam(raw string) pgtype.Text {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: raw, Valid: true}
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

func (s *Service) postgresProductParams(c *gin.Context, p pageParams, sort string) (catalogstore.ListProductsParams, bool) {
	switch sort {
	case "newest", "latest", "price_asc", "price_desc", "popular", "rating":
	default:
		sort = "newest"
	}
	categoryID, ok := parseOptionalUUIDStrict(c.Query("category_id"))
	if !ok {
		return catalogstore.ListProductsParams{}, false
	}
	brandID, ok := parseOptionalUUIDStrict(c.Query("brand_id"))
	if !ok {
		return catalogstore.ListProductsParams{}, false
	}
	sellerID, ok := parseOptionalUUIDStrict(c.Query("seller_id"))
	if !ok {
		return catalogstore.ListProductsParams{}, false
	}
	minPrice, ok := parseOptionalInt64(c.Query("min_price_cents"))
	if !ok {
		return catalogstore.ListProductsParams{}, false
	}
	maxPrice, ok := parseOptionalInt64(c.Query("max_price_cents"))
	if !ok {
		return catalogstore.ListProductsParams{}, false
	}
	if minPrice.Valid && maxPrice.Valid && minPrice.Int64 > maxPrice.Int64 {
		return catalogstore.ListProductsParams{}, false
	}
	return catalogstore.ListProductsParams{
		CategoryID:    categoryID,
		BrandID:       brandID,
		SellerID:      sellerID,
		MinPriceCents: minPrice,
		MaxPriceCents: maxPrice,
		Query:         textParam(c.Query("q")),
		Sort:          sort,
		Limit:         p.Limit,
		Offset:        int32(p.From),
	}, true
}

func parseOptionalUUIDStrict(raw string) (pgtype.UUID, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.UUID{}, true
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func parseOptionalInt64(raw string) (pgtype.Int8, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Int8{}, true
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v < 0 {
		return pgtype.Int8{}, false
	}
	return pgtype.Int8{Int64: v, Valid: true}, true
}

func productRowJSON(p catalogstore.ListProductsRow) gin.H {
	return gin.H{
		"id":              uuidString(p.ID),
		"seller_id":       uuidString(p.SellerID),
		"category_id":     nullableUUID(p.CategoryID),
		"brand_id":        nullableUUID(p.BrandID),
		"name":            p.Name,
		"slug":            p.Slug,
		"description":     textValue(p.Description),
		"min_price_cents": p.MinPriceCents,
		"max_price_cents": p.MaxPriceCents,
		"currency":        p.Currency,
		"total_sold":      p.TotalSold,
		"rating":          numericString(p.Rating),
		"review_count":    p.ReviewCount,
		"published_at":    timeValue(p.PublishedAt),
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

func (s *Service) indexProductByID(c *gin.Context, id uuid.UUID) error {
	row, err := s.catalog.GetProductByID(c.Request.Context(), pgxutil.UUID(id))
	if err != nil {
		return err
	}
	images, _ := s.catalog.ListProductImages(c.Request.Context(), row.ID)
	imageURL := ""
	for _, img := range images {
		if img.IsPrimary {
			imageURL = img.Url
			break
		}
	}
	doc := platformsearch.ProductDocument{
		ID:            uuidString(row.ID),
		SellerID:      uuidString(row.SellerID),
		CategoryID:    uuidString(row.CategoryID),
		BrandID:       uuidString(row.BrandID),
		Name:          row.Name,
		Slug:          row.Slug,
		Description:   safeText(row.Description),
		MinPriceCents: row.MinPriceCents,
		MaxPriceCents: row.MaxPriceCents,
		Currency:      row.Currency,
		TotalSold:     row.TotalSold,
		Rating:        numericToFloat(row.Rating),
		ReviewCount:   row.ReviewCount,
		PublishedAt:   timePtr(row.PublishedAt),
		UpdatedAt:     row.UpdatedAt.Time,
		ImageURL:      imageURL,
	}
	return s.os.IndexProduct(c.Request.Context(), doc)
}

func (s *Service) indexProductRow(c *gin.Context, r searchstore.ListIndexableProductsRow) error {
	images, _ := s.catalog.ListProductImages(c.Request.Context(), r.ID)
	imageURL := ""
	for _, img := range images {
		if img.IsPrimary {
			imageURL = img.Url
			break
		}
	}
	doc := platformsearch.ProductDocument{
		ID:            uuidString(r.ID),
		SellerID:      uuidString(r.SellerID),
		CategoryID:    uuidString(r.CategoryID),
		BrandID:       uuidString(r.BrandID),
		Name:          r.Name,
		Slug:          r.Slug,
		Description:   safeText(r.Description),
		MinPriceCents: r.MinPriceCents,
		MaxPriceCents: r.MaxPriceCents,
		Currency:      r.Currency,
		TotalSold:     r.TotalSold,
		Rating:        numericToFloat(r.Rating),
		ReviewCount:   r.ReviewCount,
		PublishedAt:   timePtr(r.PublishedAt),
		UpdatedAt:     r.UpdatedAt.Time,
		ImageURL:      imageURL,
	}
	return s.os.IndexProduct(c.Request.Context(), doc)
}

func timePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	v := t.Time
	return &v
}

func uuidString(id pgtype.UUID) string {
	if v, ok := pgxutil.ToUUID(id); ok {
		return v.String()
	}
	return ""
}

func safeText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func numericString(n pgtype.Numeric) string {
	if !n.Valid || n.Int == nil {
		return "0"
	}
	r := new(big.Rat).SetInt(n.Int)
	if n.Exp > 0 {
		r.Mul(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)))
	} else if n.Exp < 0 {
		r.Quo(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)))
	}
	f, _ := r.Float64()
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid || n.Int == nil {
		return 0
	}
	r := new(big.Rat).SetInt(n.Int)
	if n.Exp > 0 {
		r.Mul(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n.Exp)), nil)))
	} else if n.Exp < 0 {
		r.Quo(r, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-n.Exp)), nil)))
	}
	f, _ := r.Float64()
	return f
}

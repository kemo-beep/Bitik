package catalogsvc

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/config"
	"github.com/bitik/backend/internal/pgxutil"
	catalogstore "github.com/bitik/backend/internal/store/catalog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	defaultPage    = 1
	defaultPerPage = 24
	maxPerPage     = 100
)

type Service struct {
	cfg     config.Config
	log     *zap.Logger
	queries *catalogstore.Queries
}

func NewService(cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool) *Service {
	return &Service{
		cfg:     cfg,
		log:     logger,
		queries: catalogstore.New(pool),
	}
}

type pagination struct {
	Page    int32
	PerPage int32
	Limit   int32
	Offset  int32
}

func parsePagination(c *gin.Context) pagination {
	page := parseInt32(c.DefaultQuery("page", strconv.Itoa(defaultPage)), defaultPage)
	perPage := parseInt32(c.DefaultQuery("per_page", strconv.Itoa(defaultPerPage)), defaultPerPage)
	if page < 1 {
		page = defaultPage
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return pagination{
		Page:    page,
		PerPage: perPage,
		Limit:   perPage,
		Offset:  (page - 1) * perPage,
	}
}

func parseInt32(raw string, fallback int32) int32 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 32)
	if err != nil {
		return fallback
	}
	return int32(v)
}

func parseInt64Ptr(raw string) pgtype.Int8 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Int8{}
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: v, Valid: true}
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

func parseUUIDParam(c *gin.Context, name string) (pgtype.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}

func parseOptionalUUID(raw string) pgtype.UUID {
	if raw == "" {
		return pgtype.UUID{}
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgxutil.UUID(id)
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

func textParam(raw string) pgtype.Text {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: raw, Valid: true}
}

func productParams(c *gin.Context, p pagination) (catalogstore.ListProductsParams, bool) {
	sort := strings.TrimSpace(c.DefaultQuery("sort", "newest"))
	switch sort {
	case "newest", "price_asc", "price_desc", "popular", "rating":
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
		Offset:        p.Offset,
	}, true
}

func countParams(params catalogstore.ListProductsParams) catalogstore.CountPublicProductsParams {
	return catalogstore.CountPublicProductsParams{
		CategoryID:    params.CategoryID,
		BrandID:       params.BrandID,
		SellerID:      params.SellerID,
		MinPriceCents: params.MinPriceCents,
		MaxPriceCents: params.MaxPriceCents,
		Query:         params.Query,
	}
}

func pageMeta(p pagination, total int64) map[string]any {
	pages := int64(0)
	if p.PerPage > 0 {
		pages = (total + int64(p.PerPage) - 1) / int64(p.PerPage)
	}
	return map[string]any{
		"page":        p.Page,
		"per_page":    p.PerPage,
		"total":       total,
		"total_pages": pages,
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
	return r.FloatString(2)
}

func categoryJSON(c catalogstore.Category) gin.H {
	return gin.H{
		"id":         uuidString(c.ID),
		"parent_id":  nullableUUID(c.ParentID),
		"name":       c.Name,
		"slug":       c.Slug,
		"image_url":  textValue(c.ImageUrl),
		"sort_order": c.SortOrder,
	}
}

func brandJSON(b catalogstore.Brand) gin.H {
	return gin.H{
		"id":       uuidString(b.ID),
		"name":     b.Name,
		"slug":     b.Slug,
		"logo_url": textValue(b.LogoUrl),
	}
}

func sellerJSON(s catalogstore.Seller) gin.H {
	return sellerJSONFromFields(s.ID, s.ShopName, s.Slug, s.Description, s.LogoUrl, s.BannerUrl, s.Rating, s.TotalSales)
}

func sellerByIDJSON(s catalogstore.GetPublicSellerByIDRow) gin.H {
	return sellerJSONFromFields(s.ID, s.ShopName, s.Slug, s.Description, s.LogoUrl, s.BannerUrl, s.Rating, s.TotalSales)
}

func sellerBySlugJSON(s catalogstore.GetPublicSellerBySlugRow) gin.H {
	return sellerJSONFromFields(s.ID, s.ShopName, s.Slug, s.Description, s.LogoUrl, s.BannerUrl, s.Rating, s.TotalSales)
}

func sellerJSONFromFields(id pgtype.UUID, shopName, slug string, description, logoURL, bannerURL pgtype.Text, rating pgtype.Numeric, totalSales int64) gin.H {
	return gin.H{
		"id":          uuidString(id),
		"shop_name":   shopName,
		"slug":        slug,
		"description": textValue(description),
		"logo_url":    textValue(logoURL),
		"banner_url":  textValue(bannerURL),
		"rating":      numericString(rating),
		"total_sales": totalSales,
	}
}

func nullableUUID(id pgtype.UUID) any {
	if !id.Valid {
		return nil
	}
	return uuidString(id)
}

type productLike interface {
	productID() pgtype.UUID
	productSellerID() pgtype.UUID
	productCategoryID() pgtype.UUID
	productBrandID() pgtype.UUID
	productName() string
	productSlug() string
	productDescription() pgtype.Text
	productMinPrice() int64
	productMaxPrice() int64
	productCurrency() string
	productTotalSold() int64
	productRating() pgtype.Numeric
	productReviewCount() int64
}

func productJSONFromFields(
	id, sellerID, categoryID, brandID pgtype.UUID,
	name, slug string,
	description pgtype.Text,
	minPrice, maxPrice int64,
	currency string,
	totalSold int64,
	rating pgtype.Numeric,
	reviewCount int64,
	primaryImageURL pgtype.Text,
) gin.H {
	return gin.H{
		"id":                uuidString(id),
		"seller_id":         uuidString(sellerID),
		"category_id":       nullableUUID(categoryID),
		"brand_id":          nullableUUID(brandID),
		"name":              name,
		"slug":              slug,
		"description":       textValue(description),
		"min_price_cents":   minPrice,
		"max_price_cents":   maxPrice,
		"currency":          currency,
		"total_sold":        totalSold,
		"rating":            numericString(rating),
		"review_count":      reviewCount,
		"primary_image_url": textValue(primaryImageURL),
	}
}

func productRowJSON(p catalogstore.ListProductsRow) gin.H {
	return productJSONFromFields(p.ID, p.SellerID, p.CategoryID, p.BrandID, p.Name, p.Slug, p.Description, p.MinPriceCents, p.MaxPriceCents, p.Currency, p.TotalSold, p.Rating, p.ReviewCount, p.PrimaryImageUrl)
}

func productByIDJSON(p catalogstore.GetProductByIDRow) gin.H {
	return productJSONFromFields(p.ID, p.SellerID, p.CategoryID, p.BrandID, p.Name, p.Slug, p.Description, p.MinPriceCents, p.MaxPriceCents, p.Currency, p.TotalSold, p.Rating, p.ReviewCount, pgtype.Text{})
}

func productBySlugJSON(p catalogstore.GetProductBySlugRow) gin.H {
	return productJSONFromFields(p.ID, p.SellerID, p.CategoryID, p.BrandID, p.Name, p.Slug, p.Description, p.MinPriceCents, p.MaxPriceCents, p.Currency, p.TotalSold, p.Rating, p.ReviewCount, pgtype.Text{})
}

func relatedProductJSON(p catalogstore.ListRelatedProductsRow) gin.H {
	return productJSONFromFields(p.ID, p.SellerID, p.CategoryID, p.BrandID, p.Name, p.Slug, p.Description, p.MinPriceCents, p.MaxPriceCents, p.Currency, p.TotalSold, p.Rating, p.ReviewCount, p.PrimaryImageUrl)
}

func imageJSON(img catalogstore.ProductImage) gin.H {
	return gin.H{
		"id":         uuidString(img.ID),
		"url":        img.Url,
		"alt_text":   textValue(img.AltText),
		"sort_order": img.SortOrder,
		"is_primary": img.IsPrimary,
	}
}

func variantJSON(v catalogstore.ProductVariant) gin.H {
	return gin.H{
		"id":                     uuidString(v.ID),
		"sku":                    v.Sku,
		"name":                   textValue(v.Name),
		"price_cents":            v.PriceCents,
		"compare_at_price_cents": int8Value(v.CompareAtPriceCents),
		"currency":               v.Currency,
		"weight_grams":           int4Value(v.WeightGrams),
	}
}

func reviewJSONFromListProduct(r catalogstore.ListProductReviewsRow) gin.H {
	return gin.H{
		"id":                   uuidString(r.ID),
		"product_id":           uuidString(r.ProductID),
		"user_id":              uuidString(r.UserID),
		"rating":               r.Rating,
		"title":                textValue(r.Title),
		"body":                 textValue(r.Body),
		"is_verified_purchase": r.IsVerifiedPurchase,
		"seller_reply":         textValue(r.SellerReply),
		"seller_reply_at":      timestamptzValue(r.SellerReplyAt),
		"created_at":           r.CreatedAt.Time,
	}
}

func reviewJSONFromListSeller(r catalogstore.ListSellerReviewsRow) gin.H {
	return gin.H{
		"id":                   uuidString(r.ID),
		"product_id":           uuidString(r.ProductID),
		"user_id":              uuidString(r.UserID),
		"rating":               r.Rating,
		"title":                textValue(r.Title),
		"body":                 textValue(r.Body),
		"is_verified_purchase": r.IsVerifiedPurchase,
		"seller_reply":         textValue(r.SellerReply),
		"seller_reply_at":      timestamptzValue(r.SellerReplyAt),
		"created_at":           r.CreatedAt.Time,
	}
}

func timestamptzValue(t pgtype.Timestamptz) any {
	if !t.Valid {
		return nil
	}
	return t.Time
}

func bannerJSON(b catalogstore.CmsBanner) gin.H {
	return gin.H{
		"id":         uuidString(b.ID),
		"title":      b.Title,
		"image_url":  b.ImageUrl,
		"link_url":   textValue(b.LinkUrl),
		"placement":  b.Placement,
		"sort_order": b.SortOrder,
	}
}

func int8Value(v pgtype.Int8) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}

func int4Value(v pgtype.Int4) any {
	if !v.Valid {
		return nil
	}
	return v.Int32
}

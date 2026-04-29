package sellersvc

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/queue"
	sellerstore "github.com/bitik/backend/internal/store/sellers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleListProducts(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	p := pagination(c)
	total, err := s.queries.CountSellerProducts(c.Request.Context(), seller.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not count products.")
		return
	}
	products, err := s.queries.ListSellerProducts(c.Request.Context(), sellerstore.ListSellerProductsParams{SellerID: seller.ID, Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load products.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(products, productFromList), "pagination": pageMeta(p, total)})
}

func (s *Service) HandleCreateProduct(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	var req struct {
		CategoryID    string `json:"category_id"`
		BrandID       string `json:"brand_id"`
		Name          string `json:"name" binding:"required"`
		Slug          string `json:"slug"`
		Description   string `json:"description"`
		MinPriceCents int64  `json:"min_price_cents" binding:"required"`
		MaxPriceCents int64  `json:"max_price_cents" binding:"required"`
		Currency      string `json:"currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.MinPriceCents < 0 || req.MaxPriceCents < req.MinPriceCents {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid product request.")
		return
	}
	categoryID, ok := optionalUUID(req.CategoryID)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_category_id", "Invalid category id.")
		return
	}
	brandID, ok := optionalUUID(req.BrandID)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_brand_id", "Invalid brand id.")
		return
	}
	slug := req.Slug
	if slug == "" {
		slug = toSlug(req.Name)
	}
	created, err := s.queries.CreateSellerProduct(c.Request.Context(), sellerstore.CreateSellerProductParams{
		SellerID:      seller.ID,
		CategoryID:    categoryID,
		BrandID:       brandID,
		Name:          strings.TrimSpace(req.Name),
		Slug:          idempotentSlug(slug),
		Description:   text(req.Description),
		Status:        "draft",
		MinPriceCents: req.MinPriceCents,
		MaxPriceCents: req.MaxPriceCents,
		Currency:      req.Currency,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusConflict, "product_create_failed", "Could not create product.")
		return
	}
	s.enqueueProductIndex(c.Request.Context(), uuidString(created.ID))
	apiresponse.Respond(c, http.StatusCreated, productFromCreate(created), nil)
}

func (s *Service) HandleGetProduct(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	product, err := s.queries.GetSellerProduct(c.Request.Context(), sellerstore.GetSellerProductParams{ID: productID, SellerID: seller.ID})
	if err != nil {
		writeNotFound(c, err, "product_not_found", "Product not found.")
		return
	}
	images, err := s.queries.ListProductImagesForSeller(c.Request.Context(), sellerstore.ListProductImagesForSellerParams{ProductID: productID, SellerID: seller.ID})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load product images.")
		return
	}
	variants, err := s.queries.ListProductVariantsForSeller(c.Request.Context(), sellerstore.ListProductVariantsForSellerParams{ProductID: productID, SellerID: seller.ID})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load product variants.")
		return
	}
	payload := productFromGet(product)
	payload["images"] = mapSlice(images, imageJSON)
	payload["variants"] = mapSlice(variants, variantJSON)
	apiresponse.OK(c, payload)
}

func (s *Service) HandleUpdateProduct(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	var req struct {
		CategoryID    string `json:"category_id"`
		BrandID       string `json:"brand_id"`
		Name          string `json:"name"`
		Slug          string `json:"slug"`
		Description   string `json:"description"`
		MinPriceCents *int64 `json:"min_price_cents"`
		MaxPriceCents *int64 `json:"max_price_cents"`
		Currency      string `json:"currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid product update.")
		return
	}
	categoryID, ok := optionalUUID(req.CategoryID)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_category_id", "Invalid category id.")
		return
	}
	brandID, ok := optionalUUID(req.BrandID)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_brand_id", "Invalid brand id.")
		return
	}
	updated, err := s.queries.UpdateSellerProduct(c.Request.Context(), sellerstore.UpdateSellerProductParams{
		ID:            productID,
		SellerID:      seller.ID,
		CategoryID:    categoryID,
		BrandID:       brandID,
		Name:          text(req.Name),
		Slug:          text(req.Slug),
		Description:   text(req.Description),
		MinPriceCents: int8Ptr(req.MinPriceCents),
		MaxPriceCents: int8Ptr(req.MaxPriceCents),
		Currency:      text(req.Currency),
	})
	if err != nil {
		writeNotFound(c, err, "product_not_found", "Product not found or cannot be updated.")
		return
	}
	s.enqueueProductIndex(c.Request.Context(), uuidString(updated.ID))
	apiresponse.OK(c, updated)
}

func (s *Service) enqueueProductIndex(ctx context.Context, productID string) {
	if s.queue == nil || strings.TrimSpace(productID) == "" {
		return
	}
	_ = s.queue.PublishJob(ctx, queue.JobIndexProduct, "index_product:"+productID, gin.H{"product_id": productID})
}

func (s *Service) HandleDeleteProduct(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	if err := s.queries.SoftDeleteSellerProduct(c.Request.Context(), sellerstore.SoftDeleteSellerProductParams{ID: productID, SellerID: seller.ID}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete product.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandlePublishProduct(c *gin.Context) {
	s.productTransition(c, "publish")
}

func (s *Service) HandleUnpublishProduct(c *gin.Context) {
	s.productTransition(c, "unpublish")
}

func (s *Service) productTransition(c *gin.Context, action string) {
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	var data any
	var err error
	if action == "publish" {
		data, err = s.queries.PublishSellerProduct(c.Request.Context(), sellerstore.PublishSellerProductParams{ID: productID, SellerID: seller.ID})
	} else {
		data, err = s.queries.UnpublishSellerProduct(c.Request.Context(), sellerstore.UnpublishSellerProductParams{ID: productID, SellerID: seller.ID})
	}
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_transition", "Product status transition is not allowed.")
		return
	}
	apiresponse.OK(c, data)
}

func (s *Service) HandleDuplicateProduct(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
		Slug string `json:"slug"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid duplicate request.")
		return
	}
	slug := req.Slug
	if slug == "" {
		slug = toSlug(req.Name)
	}
	created, err := s.queries.DuplicateSellerProduct(c.Request.Context(), sellerstore.DuplicateSellerProductParams{SourceProductID: productID, SellerID: seller.ID, Name: req.Name, Slug: slug})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "duplicate_failed", "Could not duplicate product.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, created, nil)
}

func (s *Service) HandleCreateProductImage(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	if _, err := s.queries.GetSellerProduct(c.Request.Context(), sellerstore.GetSellerProductParams{ID: productID, SellerID: seller.ID}); err != nil {
		writeNotFound(c, err, "product_not_found", "Product not found.")
		return
	}
	var req struct {
		URL       string `json:"url" binding:"required"`
		AltText   string `json:"alt_text"`
		SortOrder int32  `json:"sort_order"`
		IsPrimary bool   `json:"is_primary"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid image request.")
		return
	}
	img, err := s.queries.CreateProductImage(c.Request.Context(), sellerstore.CreateProductImageParams{ProductID: productID, Url: req.URL, AltText: text(req.AltText), SortOrder: req.SortOrder, Column5: req.IsPrimary})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "image_create_failed", "Could not create product image.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, imageJSON(img), nil)
}

func (s *Service) HandleListProductImages(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	images, err := s.queries.ListProductImagesForSeller(c.Request.Context(), sellerstore.ListProductImagesForSellerParams{ProductID: productID, SellerID: seller.ID})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load images.")
		return
	}
	apiresponse.OK(c, mapSlice(images, imageJSON))
}

func (s *Service) HandleUpdateProductImage(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	imageID, ok := uuidParam(c, "image_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_image_id", "Invalid image id.")
		return
	}
	var req struct {
		SortOrder int32 `json:"sort_order"`
		IsPrimary bool  `json:"is_primary"`
	}
	_ = c.ShouldBindJSON(&req)
	img, err := s.queries.UpdateProductImageOrder(c.Request.Context(), sellerstore.UpdateProductImageOrderParams{ID: imageID, SellerID: seller.ID, SortOrder: req.SortOrder, IsPrimary: req.IsPrimary})
	if err != nil {
		writeNotFound(c, err, "image_not_found", "Image not found.")
		return
	}
	apiresponse.OK(c, imageJSON(img))
}

func (s *Service) HandleDeleteProductImage(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	imageID, ok := uuidParam(c, "image_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_image_id", "Invalid image id.")
		return
	}
	if err := s.queries.DeleteProductImage(c.Request.Context(), sellerstore.DeleteProductImageParams{ID: imageID, SellerID: seller.ID}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete image.")
		return
	}
	c.Status(http.StatusNoContent)
}

func writeBadOrNotFound(c *gin.Context, err error, code, message string) {
	if errors.Is(err, pgx.ErrNoRows) {
		apiresponse.Error(c, http.StatusNotFound, code, message)
		return
	}
	apiresponse.Error(c, http.StatusBadRequest, code, message)
}

func parseUUIDString(raw string) pgtype.UUID {
	id, err := uuid.Parse(raw)
	if err != nil {
		return pgtype.UUID{}
	}
	return pgxutil.UUID(id)
}

package catalogsvc

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	catalogstore "github.com/bitik/backend/internal/store/catalog"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (s *Service) RegisterRoutes(rg *gin.RouterGroup) {
	public := rg.Group("/public")
	public.GET("/home", s.HandleHome)
	public.GET("/home/sections", s.HandleHomeSections)
	public.GET("/banners", s.HandleListBanners)
	public.GET("/categories", s.HandleListCategories)
	public.GET("/categories/:category_id", s.HandleGetCategory)
	public.GET("/categories/:category_id/products", s.HandleCategoryProducts)
	public.GET("/brands", s.HandleListBrands)
	public.GET("/brands/:brand_id", s.HandleGetBrand)
	public.GET("/brands/:brand_id/products", s.HandleBrandProducts)
	public.GET("/products", s.HandleListProducts)
	public.GET("/products/slug/:slug", s.HandleGetProductBySlug)
	public.GET("/products/:product_id", s.HandleGetProduct)
	public.GET("/products/:product_id/variants", s.HandleProductVariants)
	public.GET("/products/:product_id/reviews", s.HandleProductReviews)
	public.GET("/products/:product_id/related", s.HandleRelatedProducts)
	public.GET("/sellers/slug/:slug", s.HandleGetSellerBySlug)
	public.GET("/sellers/:seller_id", s.HandleGetSeller)
	public.GET("/sellers/:seller_id/products", s.HandleSellerProducts)
	public.GET("/sellers/:seller_id/reviews", s.HandleSellerReviews)
}

func (s *Service) HandleHome(c *gin.Context) {
	banners, err := s.queries.ListActiveBanners(c.Request.Context(), catalogstore.ListActiveBannersParams{
		Placement: textParam("home"),
		Limit:     10,
		Offset:    0,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load home banners.")
		return
	}
	sections, err := s.queries.ListHomeSections(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load home sections.")
		return
	}
	products, err := s.queries.ListProducts(c.Request.Context(), catalogstore.ListProductsParams{
		Sort:  "popular",
		Limit: 12,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load home products.")
		return
	}
	apiresponse.OK(c, gin.H{
		"banners":  mapSlice(banners, bannerJSON),
		"sections": homeSectionsJSON(sections),
		"products": mapSlice(products, productRowJSON),
	})
}

func (s *Service) HandleHomeSections(c *gin.Context) {
	sections, err := s.queries.ListHomeSections(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load home sections.")
		return
	}
	apiresponse.OK(c, homeSectionsJSON(sections))
}

func (s *Service) HandleListBanners(c *gin.Context) {
	p := parsePagination(c)
	banners, err := s.queries.ListActiveBanners(c.Request.Context(), catalogstore.ListActiveBannersParams{
		Placement: textParam(c.DefaultQuery("placement", "home")),
		Limit:     p.Limit,
		Offset:    p.Offset,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load banners.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(banners, bannerJSON)})
}

func (s *Service) HandleListCategories(c *gin.Context) {
	categories, err := s.queries.ListCategories(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load categories.")
		return
	}
	apiresponse.OK(c, mapSlice(categories, categoryJSON))
}

func (s *Service) HandleGetCategory(c *gin.Context) {
	id, ok := parseUUIDParam(c, "category_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_category_id", "Invalid category id.")
		return
	}
	category, err := s.queries.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		writeNotFoundOrError(c, err, "category_not_found", "Category not found.")
		return
	}
	apiresponse.OK(c, categoryJSON(category))
}

func (s *Service) HandleCategoryProducts(c *gin.Context) {
	id, ok := parseUUIDParam(c, "category_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_category_id", "Invalid category id.")
		return
	}
	s.listProducts(c, func(params *catalogstore.ListProductsParams) { params.CategoryID = id })
}

func (s *Service) HandleListBrands(c *gin.Context) {
	p := parsePagination(c)
	brands, err := s.queries.ListBrands(c.Request.Context(), catalogstore.ListBrandsParams{Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load brands.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(brands, brandJSON), "pagination": pageMeta(p, int64(len(brands)))})
}

func (s *Service) HandleGetBrand(c *gin.Context) {
	id, ok := parseUUIDParam(c, "brand_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_brand_id", "Invalid brand id.")
		return
	}
	brand, err := s.queries.GetBrandByID(c.Request.Context(), id)
	if err != nil {
		writeNotFoundOrError(c, err, "brand_not_found", "Brand not found.")
		return
	}
	apiresponse.OK(c, brandJSON(brand))
}

func (s *Service) HandleBrandProducts(c *gin.Context) {
	id, ok := parseUUIDParam(c, "brand_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_brand_id", "Invalid brand id.")
		return
	}
	s.listProducts(c, func(params *catalogstore.ListProductsParams) { params.BrandID = id })
}

func (s *Service) HandleListProducts(c *gin.Context) {
	s.listProducts(c, nil)
}

func (s *Service) HandleGetProduct(c *gin.Context) {
	id, ok := parseUUIDParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	product, err := s.queries.GetProductByID(c.Request.Context(), id)
	if err != nil {
		writeNotFoundOrError(c, err, "product_not_found", "Product not found.")
		return
	}
	images, err := s.queries.ListProductImages(c.Request.Context(), id)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load product images.")
		return
	}
	variants, err := s.queries.ListProductVariants(c.Request.Context(), id)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load product variants.")
		return
	}
	payload := productByIDJSON(product)
	payload["images"] = mapSlice(images, imageJSON)
	payload["variants"] = mapSlice(variants, variantJSON)
	apiresponse.OK(c, payload)
}

func (s *Service) HandleGetProductBySlug(c *gin.Context) {
	product, err := s.queries.GetProductBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeNotFoundOrError(c, err, "product_not_found", "Product not found.")
		return
	}
	images, err := s.queries.ListProductImages(c.Request.Context(), product.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load product images.")
		return
	}
	variants, err := s.queries.ListProductVariants(c.Request.Context(), product.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load product variants.")
		return
	}
	payload := productBySlugJSON(product)
	payload["images"] = mapSlice(images, imageJSON)
	payload["variants"] = mapSlice(variants, variantJSON)
	apiresponse.OK(c, payload)
}

func (s *Service) HandleProductVariants(c *gin.Context) {
	id, ok := parseUUIDParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	if _, err := s.queries.GetProductByID(c.Request.Context(), id); err != nil {
		writeNotFoundOrError(c, err, "product_not_found", "Product not found.")
		return
	}
	variants, err := s.queries.ListProductVariants(c.Request.Context(), id)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load variants.")
		return
	}
	apiresponse.OK(c, mapSlice(variants, variantJSON))
}

func (s *Service) HandleProductReviews(c *gin.Context) {
	id, ok := parseUUIDParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	p := parsePagination(c)
	reviews, err := s.queries.ListProductReviews(c.Request.Context(), catalogstore.ListProductReviewsParams{ProductID: id, Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load reviews.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(reviews, reviewJSONFromListProduct)})
}

func (s *Service) HandleRelatedProducts(c *gin.Context) {
	id, ok := parseUUIDParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	products, err := s.queries.ListRelatedProducts(c.Request.Context(), catalogstore.ListRelatedProductsParams{ProductID: id, Limit: 12})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load related products.")
		return
	}
	apiresponse.OK(c, mapSlice(products, relatedProductJSON))
}

func (s *Service) HandleGetSeller(c *gin.Context) {
	id, ok := parseUUIDParam(c, "seller_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_seller_id", "Invalid seller id.")
		return
	}
	seller, err := s.queries.GetPublicSellerByID(c.Request.Context(), id)
	if err != nil {
		writeNotFoundOrError(c, err, "seller_not_found", "Seller not found.")
		return
	}
	apiresponse.OK(c, sellerByIDJSON(seller))
}

func (s *Service) HandleSellerProducts(c *gin.Context) {
	id, ok := parseUUIDParam(c, "seller_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_seller_id", "Invalid seller id.")
		return
	}
	s.listProducts(c, func(params *catalogstore.ListProductsParams) { params.SellerID = id })
}

func (s *Service) HandleSellerReviews(c *gin.Context) {
	id, ok := parseUUIDParam(c, "seller_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_seller_id", "Invalid seller id.")
		return
	}
	p := parsePagination(c)
	reviews, err := s.queries.ListSellerReviews(c.Request.Context(), catalogstore.ListSellerReviewsParams{SellerID: id, Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load seller reviews.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(reviews, reviewJSONFromListSeller)})
}

func (s *Service) HandleGetSellerBySlug(c *gin.Context) {
	seller, err := s.queries.GetPublicSellerBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeNotFoundOrError(c, err, "seller_not_found", "Seller not found.")
		return
	}
	apiresponse.OK(c, sellerBySlugJSON(seller))
}

func (s *Service) listProducts(c *gin.Context, mutate func(*catalogstore.ListProductsParams)) {
	p := parsePagination(c)
	params, ok := productParams(c, p)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_filters", "Invalid product filters.")
		return
	}
	if mutate != nil {
		mutate(&params)
	}
	total, err := s.queries.CountPublicProducts(c.Request.Context(), countParams(params))
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not count products.")
		return
	}
	products, err := s.queries.ListProducts(c.Request.Context(), params)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load products.")
		return
	}
	apiresponse.OK(c, gin.H{
		"items":      mapSlice(products, productRowJSON),
		"pagination": pageMeta(p, total),
	})
}

func homeSectionsJSON(sections []catalogstore.PlatformSetting) []gin.H {
	out := make([]gin.H, 0, len(sections))
	for _, section := range sections {
		var value any = json.RawMessage(section.Value)
		out = append(out, gin.H{
			"key":         section.Key,
			"value":       value,
			"description": textValue(section.Description),
		})
	}
	return out
}

func writeNotFoundOrError(c *gin.Context, err error, code, message string) {
	if errors.Is(err, pgx.ErrNoRows) {
		apiresponse.Error(c, http.StatusNotFound, code, message)
		return
	}
	apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load resource.")
}

func mapSlice[T any](items []T, mapper func(T) gin.H) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, mapper(item))
	}
	return out
}

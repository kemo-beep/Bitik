package sellersvc

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/authsvc"
	"github.com/bitik/backend/internal/middleware"
	"github.com/bitik/backend/internal/pgxutil"
	sellerstore "github.com/bitik/backend/internal/store/sellers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Service) RegisterRoutes(v1 *gin.RouterGroup, auth *authsvc.Service) {
	if auth == nil {
		return
	}
	protected := v1.Group("", middleware.RequireBearerJWT(s.cfg), auth.RequireActiveUser())

	seller := protected.Group("/seller")
	seller.POST("/apply", s.HandleApply)
	seller.GET("/application", s.HandleGetApplication)
	seller.PATCH("/application", s.HandlePatchApplication)
	seller.POST("/documents", s.HandleCreateDocument)
	seller.DELETE("/documents/:document_id", s.HandleDeleteDocument)

	center := seller.Group("", s.requireRole("seller", "admin"), s.requireSeller())
	center.GET("/me", s.HandleGetSeller)
	center.PATCH("/me", s.HandlePatchSeller)
	center.PATCH("/me/profile", s.HandlePatchSeller)
	center.PATCH("/me/settings", s.HandlePatchSettings)
	center.PATCH("/me/media", s.HandlePatchMedia)
	center.GET("/dashboard", s.HandleDashboard)
	center.GET("/products", s.HandleListProducts)
	center.POST("/products", s.HandleCreateProduct)
	center.GET("/products/:product_id", s.HandleGetProduct)
	center.PATCH("/products/:product_id", s.HandleUpdateProduct)
	center.DELETE("/products/:product_id", s.HandleDeleteProduct)
	center.POST("/products/:product_id/publish", s.HandlePublishProduct)
	center.POST("/products/:product_id/unpublish", s.HandleUnpublishProduct)
	center.POST("/products/:product_id/duplicate", s.HandleDuplicateProduct)
	center.POST("/products/:product_id/images", s.HandleCreateProductImage)
	center.GET("/products/:product_id/images", s.HandleListProductImages)
	center.PATCH("/products/:product_id/images/:image_id", s.HandleUpdateProductImage)
	center.DELETE("/products/:product_id/images/:image_id", s.HandleDeleteProductImage)
	center.GET("/products/:product_id/variants", s.HandleListVariants)
	center.POST("/products/:product_id/variants", s.HandleCreateVariant)
	center.PATCH("/variants/:variant_id", s.HandleUpdateVariant)
	center.DELETE("/variants/:variant_id", s.HandleDeleteVariant)
	center.GET("/products/:product_id/options", s.HandleListOptions)
	center.POST("/products/:product_id/options", s.HandleCreateOption)
	center.POST("/options/:option_id/values", s.HandleCreateOptionValue)
	center.DELETE("/options/:option_id", s.HandleDeleteOption)
	center.GET("/inventory", s.HandleListInventory)
	center.GET("/inventory/low-stock", s.HandleLowStock)
	center.GET("/inventory/:inventory_id", s.HandleGetInventory)
	center.PATCH("/inventory/:inventory_id", s.HandleUpdateInventory)
	center.POST("/inventory/:inventory_id/adjust", s.HandleAdjustInventory)
	center.GET("/inventory/:inventory_id/movements", s.HandleInventoryMovements)
	center.POST("/inventory/bulk-update", s.HandleBulkUpdateInventory)

	admin := protected.Group("/admin", s.requireRole("admin"))
	admin.GET("/seller-applications", s.HandleAdminListApplications)
	admin.POST("/seller-applications/:application_id/review", s.HandleAdminReviewApplication)
	admin.PATCH("/sellers/:seller_id/status", s.HandleAdminSellerStatus)
	admin.PATCH("/products/:product_id/moderation", s.HandleAdminModerateProduct)
}

func (s *Service) requireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !hasRole(c, roles...) {
			apiresponse.Error(c, http.StatusForbidden, "forbidden", "Required role is missing.")
			c.Abort()
			return
		}
		c.Next()
	}
}

func (s *Service) requireSeller() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := currentUserID(c)
		if !ok {
			apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
			c.Abort()
			return
		}
		seller, err := s.queries.GetSellerByUserID(c.Request.Context(), pgxutil.UUID(uid))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				apiresponse.Error(c, http.StatusForbidden, "seller_required", "Active seller account is required.")
				c.Abort()
				return
			}
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load seller.")
			c.Abort()
			return
		}
		if statusString(seller.Status) != "active" && !hasRole(c, "admin") {
			apiresponse.Error(c, http.StatusForbidden, "seller_inactive", "Seller account is not active.")
			c.Abort()
			return
		}
		c.Set("seller", seller)
		c.Set("seller_id", seller.ID)
		c.Next()
	}
}

func sellerFromContext(c *gin.Context) (sellerstore.Seller, bool) {
	raw, ok := c.Get("seller")
	if !ok {
		return sellerstore.Seller{}, false
	}
	seller, ok := raw.(sellerstore.Seller)
	return seller, ok
}

func (s *Service) HandleApply(c *gin.Context) {
	uid, ok := currentUserID(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing user context.")
		return
	}
	var req struct {
		ShopName     string         `json:"shop_name" binding:"required"`
		Slug         string         `json:"slug"`
		BusinessType string         `json:"business_type"`
		Country      string         `json:"country"`
		Currency     string         `json:"currency"`
		Metadata     map[string]any `json:"metadata"`
		Submit       bool           `json:"submit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid seller application request.")
		return
	}
	slug := req.Slug
	if slug == "" {
		slug = toSlug(req.ShopName)
	}
	app, err := s.queries.CreateSellerApplication(c.Request.Context(), sellerstore.CreateSellerApplicationParams{
		UserID:       pgxutil.UUID(uid),
		ShopName:     strings.TrimSpace(req.ShopName),
		Slug:         slug,
		BusinessType: text(req.BusinessType),
		Country:      text(req.Country),
		Column6:      req.Currency,
		Column7:      jsonOrEmpty(req.Metadata),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusConflict, "application_conflict", "Could not create seller application.")
		return
	}
	if req.Submit {
		app, err = s.queries.SubmitSellerApplication(c.Request.Context(), sellerstore.SubmitSellerApplicationParams{ID: app.ID, UserID: pgxutil.UUID(uid)})
		if err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "application_not_submittable", "Seller application cannot be submitted.")
			return
		}
	}
	apiresponse.Respond(c, http.StatusCreated, applicationJSON(app), nil)
}

func (s *Service) HandleGetApplication(c *gin.Context) {
	uid, _ := currentUserID(c)
	app, err := s.queries.GetSellerApplicationByUserID(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		writeNotFound(c, err, "application_not_found", "Seller application not found.")
		return
	}
	docs, _ := s.queries.ListSellerDocuments(c.Request.Context(), sellerstore.ListSellerDocumentsParams{SellerApplicationID: app.ID})
	payload := applicationJSON(app)
	payload["documents"] = mapSlice(docs, documentJSON)
	apiresponse.OK(c, payload)
}

func (s *Service) HandlePatchApplication(c *gin.Context) {
	uid, _ := currentUserID(c)
	app, err := s.queries.GetSellerApplicationByUserID(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		writeNotFound(c, err, "application_not_found", "Seller application not found.")
		return
	}
	var req struct {
		ShopName     string         `json:"shop_name"`
		Slug         string         `json:"slug"`
		BusinessType string         `json:"business_type"`
		Country      string         `json:"country"`
		Currency     string         `json:"currency"`
		Metadata     map[string]any `json:"metadata"`
		Submit       bool           `json:"submit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid seller application update.")
		return
	}
	updated, err := s.queries.UpdateSellerApplication(c.Request.Context(), sellerstore.UpdateSellerApplicationParams{
		ID:           app.ID,
		UserID:       pgxutil.UUID(uid),
		ShopName:     text(req.ShopName),
		Slug:         text(req.Slug),
		BusinessType: text(req.BusinessType),
		Country:      text(req.Country),
		Currency:     text(req.Currency),
		Metadata:     jsonOptional(req.Metadata),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "application_not_editable", "Seller application cannot be updated.")
		return
	}
	if req.Submit {
		updated, err = s.queries.SubmitSellerApplication(c.Request.Context(), sellerstore.SubmitSellerApplicationParams{ID: updated.ID, UserID: pgxutil.UUID(uid)})
		if err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "application_not_submittable", "Seller application cannot be submitted.")
			return
		}
	}
	apiresponse.OK(c, applicationJSON(updated))
}

func (s *Service) HandleCreateDocument(c *gin.Context) {
	uid, _ := currentUserID(c)
	app, appErr := s.queries.GetSellerApplicationByUserID(c.Request.Context(), pgxutil.UUID(uid))
	seller, sellerErr := s.queries.GetSellerByUserID(c.Request.Context(), pgxutil.UUID(uid))
	if appErr != nil && sellerErr != nil {
		apiresponse.Error(c, http.StatusBadRequest, "seller_context_missing", "Seller application or account is required.")
		return
	}
	var req struct {
		DocumentType string         `json:"document_type" binding:"required"`
		FileURL      string         `json:"file_url" binding:"required"`
		Metadata     map[string]any `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid seller document request.")
		return
	}
	doc := sellerstore.CreateSellerDocumentParams{DocumentType: req.DocumentType, FileUrl: req.FileURL, Metadata: jsonOrEmpty(req.Metadata)}
	if appErr == nil {
		doc.SellerApplicationID = app.ID
	}
	if sellerErr == nil {
		doc.SellerID = seller.ID
	}
	created, err := s.queries.CreateSellerDocument(c.Request.Context(), doc)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create seller document.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, documentJSON(created), nil)
}

func (s *Service) HandleDeleteDocument(c *gin.Context) {
	uid, _ := currentUserID(c)
	docID, ok := uuidParam(c, "document_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_document_id", "Invalid document id.")
		return
	}
	if err := s.queries.DeleteSellerDocument(c.Request.Context(), sellerstore.DeleteSellerDocumentParams{ID: docID, UserID: pgxutil.UUID(uid)}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete seller document.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleGetSeller(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	apiresponse.OK(c, sellerJSON(seller))
}

func (s *Service) HandlePatchSeller(c *gin.Context) {
	uid, _ := currentUserID(c)
	seller, _ := sellerFromContext(c)
	var req struct {
		ShopName    string `json:"shop_name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid seller profile request.")
		return
	}
	updated, err := s.queries.UpdateSellerProfile(c.Request.Context(), sellerstore.UpdateSellerProfileParams{ID: seller.ID, UserID: pgxutil.UUID(uid), ShopName: text(req.ShopName), Slug: text(req.Slug), Description: text(req.Description)})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "seller_update_failed", "Could not update seller.")
		return
	}
	apiresponse.OK(c, sellerJSON(updated))
}

func (s *Service) HandlePatchSettings(c *gin.Context) {
	uid, _ := currentUserID(c)
	seller, _ := sellerFromContext(c)
	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid seller settings request.")
		return
	}
	updated, err := s.queries.UpdateSellerSettings(c.Request.Context(), sellerstore.UpdateSellerSettingsParams{ID: seller.ID, UserID: pgxutil.UUID(uid), Settings: jsonOrEmpty(req)})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "settings_update_failed", "Could not update seller settings.")
		return
	}
	apiresponse.OK(c, sellerJSON(updated))
}

func (s *Service) HandlePatchMedia(c *gin.Context) {
	uid, _ := currentUserID(c)
	seller, _ := sellerFromContext(c)
	var req struct {
		LogoURL   string `json:"logo_url"`
		BannerURL string `json:"banner_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid seller media request.")
		return
	}
	updated, err := s.queries.UpdateSellerMedia(c.Request.Context(), sellerstore.UpdateSellerMediaParams{ID: seller.ID, UserID: pgxutil.UUID(uid), LogoUrl: text(req.LogoURL), BannerUrl: text(req.BannerURL)})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "media_update_failed", "Could not update seller media.")
		return
	}
	apiresponse.OK(c, sellerJSON(updated))
}

func (s *Service) HandleDashboard(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	stats, err := s.queries.SellerDashboardStats(c.Request.Context(), seller.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load dashboard stats.")
		return
	}
	chart, err := s.queries.SellerSalesChart(c.Request.Context(), sellerstore.SellerSalesChartParams{SellerID: seller.ID, Days: 30})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load sales chart.")
		return
	}
	top, err := s.queries.SellerTopProducts(c.Request.Context(), sellerstore.SellerTopProductsParams{SellerID: seller.ID, Limit: 5})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load top products.")
		return
	}
	orders, err := s.queries.SellerRecentOrders(c.Request.Context(), sellerstore.SellerRecentOrdersParams{SellerID: seller.ID, Limit: 10})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load recent orders.")
		return
	}
	low, err := s.queries.ListLowStockForSeller(c.Request.Context(), sellerstore.ListLowStockForSellerParams{SellerID: seller.ID, Limit: 10})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load low stock.")
		return
	}
	apiresponse.OK(c, gin.H{"stats": stats, "sales_chart": chart, "top_products": top, "recent_orders": orders, "low_stock": mapSlice(low, inventoryJSON)})
}

func writeNotFound(c *gin.Context, err error, code, message string) {
	if errors.Is(err, pgx.ErrNoRows) {
		apiresponse.Error(c, http.StatusNotFound, code, message)
		return
	}
	apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load resource.")
}

func idempotentSlug(base string) string {
	if base == "" {
		return fmt.Sprintf("seller-%s", uuid.NewString()[:8])
	}
	return base
}

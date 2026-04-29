package adminsvc

import (
	"net/http"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func strPtrValue(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func (s *Service) HandleAdminListCategories(c *gin.Context) {
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, parent_id, name, slug, image_url, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE deleted_at IS NULL
		ORDER BY sort_order ASC, created_at ASC
	`)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list categories.")
		return
	}
	defer rows.Close()
	items := make([]gin.H, 0)
	for rows.Next() {
		var (
			id, parentID              uuid.UUID
			name, slug                string
			imageURL                  *string
			sortOrder                 int32
			isActive                  bool
			createdAt, updatedAt      any
		)
		if err := rows.Scan(&id, &parentID, &name, &slug, &imageURL, &sortOrder, &isActive, &createdAt, &updatedAt); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not parse categories.")
			return
		}
		items = append(items, gin.H{
			"id": id, "parent_id": parentID, "name": name, "slug": slug, "image_url": imageURL,
			"sort_order": sortOrder, "is_active": isActive, "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminCreateCategory(c *gin.Context) {
	var req struct {
		ParentID  *string `json:"parent_id"`
		Name      string  `json:"name" binding:"required"`
		Slug      string  `json:"slug" binding:"required"`
		ImageURL  *string `json:"image_url"`
		SortOrder *int32  `json:"sort_order"`
		IsActive  *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid category payload.")
		return
	}
	var parentID any
	if req.ParentID != nil && strings.TrimSpace(*req.ParentID) != "" {
		id, err := uuid.Parse(strings.TrimSpace(*req.ParentID))
		if err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid parent category id.")
			return
		}
		parentID = id
	}
	sortOrder := int32(0)
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	var id uuid.UUID
	if err := s.pool.QueryRow(c.Request.Context(), `
		INSERT INTO categories (parent_id, name, slug, image_url, sort_order, is_active)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6)
		RETURNING id
	`, parentID, strings.TrimSpace(req.Name), strings.TrimSpace(req.Slug), strPtrValue(req.ImageURL), sortOrder, isActive).Scan(&id); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create category.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, gin.H{"id": id}, nil)
}

func (s *Service) HandleAdminUpdateCategory(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("category_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid category id.")
		return
	}
	var req struct {
		Name      *string `json:"name"`
		Slug      *string `json:"slug"`
		ImageURL  *string `json:"image_url"`
		SortOrder *int32  `json:"sort_order"`
		IsActive  *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid category payload.")
		return
	}
	cmd, err := s.pool.Exec(c.Request.Context(), `
		UPDATE categories
		SET
		  name = COALESCE(NULLIF($2, ''), name),
		  slug = COALESCE(NULLIF($3, ''), slug),
		  image_url = COALESCE(NULLIF($4, ''), image_url),
		  sort_order = COALESCE($5, sort_order),
		  is_active = COALESCE($6, is_active)
		WHERE id = $1 AND deleted_at IS NULL
	`, categoryID, strPtrValue(req.Name), strPtrValue(req.Slug), strPtrValue(req.ImageURL), req.SortOrder, req.IsActive)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update category.")
		return
	}
	if cmd.RowsAffected() == 0 {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Category not found.")
		return
	}
	apiresponse.OK(c, gin.H{"id": categoryID, "updated": true})
}

func (s *Service) HandleAdminDeleteCategory(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("category_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid category id.")
		return
	}
	cmd, err := s.pool.Exec(c.Request.Context(), `UPDATE categories SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, categoryID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete category.")
		return
	}
	if cmd.RowsAffected() == 0 {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Category not found.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleAdminReorderCategories(c *gin.Context) {
	var req struct {
		Items []struct {
			ID        string `json:"id"`
			SortOrder int32  `json:"sort_order"`
		} `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid reorder payload.")
		return
	}
	tx, err := s.pool.Begin(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reorder categories.")
		return
	}
	defer tx.Rollback(c.Request.Context())
	for _, it := range req.Items {
		id, err := uuid.Parse(strings.TrimSpace(it.ID))
		if err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid category id in reorder.")
			return
		}
		if _, err := tx.Exec(c.Request.Context(), `UPDATE categories SET sort_order = $2 WHERE id = $1 AND deleted_at IS NULL`, id, it.SortOrder); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reorder categories.")
			return
		}
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not reorder categories.")
		return
	}
	apiresponse.OK(c, gin.H{"updated": len(req.Items)})
}

func (s *Service) HandleAdminListBrands(c *gin.Context) {
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, name, slug, logo_url, is_active, created_at, updated_at
		FROM brands
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list brands.")
		return
	}
	defer rows.Close()
	items := make([]gin.H, 0)
	for rows.Next() {
		var id uuid.UUID
		var name, slug string
		var logoURL *string
		var isActive bool
		var createdAt, updatedAt any
		if err := rows.Scan(&id, &name, &slug, &logoURL, &isActive, &createdAt, &updatedAt); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not parse brands.")
			return
		}
		items = append(items, gin.H{"id": id, "name": name, "slug": slug, "logo_url": logoURL, "is_active": isActive, "created_at": createdAt, "updated_at": updatedAt})
	}
	apiresponse.OK(c, gin.H{"items": items})
}

func (s *Service) HandleAdminCreateBrand(c *gin.Context) {
	var req struct {
		Name     string  `json:"name" binding:"required"`
		Slug     string  `json:"slug" binding:"required"`
		LogoURL  *string `json:"logo_url"`
		IsActive *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid brand payload.")
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	var id uuid.UUID
	if err := s.pool.QueryRow(c.Request.Context(), `
		INSERT INTO brands (name, slug, logo_url, is_active)
		VALUES ($1, $2, NULLIF($3, ''), $4)
		RETURNING id
	`, strings.TrimSpace(req.Name), strings.TrimSpace(req.Slug), strPtrValue(req.LogoURL), isActive).Scan(&id); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not create brand.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, gin.H{"id": id}, nil)
}

func (s *Service) HandleAdminUpdateBrand(c *gin.Context) {
	brandID, err := uuid.Parse(c.Param("brand_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid brand id.")
		return
	}
	var req struct {
		Name     *string `json:"name"`
		Slug     *string `json:"slug"`
		LogoURL  *string `json:"logo_url"`
		IsActive *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid brand payload.")
		return
	}
	cmd, err := s.pool.Exec(c.Request.Context(), `
		UPDATE brands
		SET
		  name = COALESCE(NULLIF($2, ''), name),
		  slug = COALESCE(NULLIF($3, ''), slug),
		  logo_url = COALESCE(NULLIF($4, ''), logo_url),
		  is_active = COALESCE($5, is_active)
		WHERE id = $1 AND deleted_at IS NULL
	`, brandID, strPtrValue(req.Name), strPtrValue(req.Slug), strPtrValue(req.LogoURL), req.IsActive)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update brand.")
		return
	}
	if cmd.RowsAffected() == 0 {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Brand not found.")
		return
	}
	apiresponse.OK(c, gin.H{"id": brandID, "updated": true})
}

func (s *Service) HandleAdminDeleteBrand(c *gin.Context) {
	brandID, err := uuid.Parse(c.Param("brand_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid brand id.")
		return
	}
	cmd, err := s.pool.Exec(c.Request.Context(), `UPDATE brands SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, brandID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete brand.")
		return
	}
	if cmd.RowsAffected() == 0 {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Brand not found.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleAdminListProducts(c *gin.Context) {
	page := parsePositiveInt32(c.DefaultQuery("page", "1"), 1)
	perPage := parsePositiveInt32(c.DefaultQuery("per_page", "25"), 25)
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage
	rows, err := s.pool.Query(c.Request.Context(), `
		SELECT id, seller_id, category_id, brand_id, name, slug, status, moderation_status, created_at, updated_at
		FROM products
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not list products.")
		return
	}
	defer rows.Close()
	items := make([]gin.H, 0)
	for rows.Next() {
		var id, sellerID, categoryID, brandID uuid.UUID
		var name, slug string
		var status, moderationStatus any
		var createdAt, updatedAt any
		if err := rows.Scan(&id, &sellerID, &categoryID, &brandID, &name, &slug, &status, &moderationStatus, &createdAt, &updatedAt); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not parse products.")
			return
		}
		items = append(items, gin.H{
			"id": id, "seller_id": sellerID, "category_id": categoryID, "brand_id": brandID, "name": name, "slug": slug,
			"status": statusString(status), "moderation_status": statusString(moderationStatus), "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	apiresponse.OK(c, gin.H{"items": items, "pagination": gin.H{"page": page, "per_page": perPage}})
}

func (s *Service) HandleAdminGetProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid product id.")
		return
	}
	var (
		id, sellerID, categoryID, brandID uuid.UUID
		name, slug                        string
		description                       *string
		status, moderationStatus          any
		createdAt, updatedAt              any
	)
	err = s.pool.QueryRow(c.Request.Context(), `
		SELECT id, seller_id, category_id, brand_id, name, slug, description, status, moderation_status, created_at, updated_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`, productID).Scan(&id, &sellerID, &categoryID, &brandID, &name, &slug, &description, &status, &moderationStatus, &createdAt, &updatedAt)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Product not found.")
		return
	}
	apiresponse.OK(c, gin.H{
		"id": id, "seller_id": sellerID, "category_id": categoryID, "brand_id": brandID, "name": name, "slug": slug,
		"description": description, "status": statusString(status), "moderation_status": statusString(moderationStatus), "created_at": createdAt, "updated_at": updatedAt,
	})
}


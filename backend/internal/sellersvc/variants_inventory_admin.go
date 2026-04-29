package sellersvc

import (
	"context"
	"errors"
	"net/http"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	sellerstore "github.com/bitik/backend/internal/store/sellers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleListVariants(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	variants, err := s.queries.ListProductVariantsForSeller(c.Request.Context(), sellerstore.ListProductVariantsForSellerParams{ProductID: productID, SellerID: seller.ID})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load variants.")
		return
	}
	apiresponse.OK(c, mapSlice(variants, variantJSON))
}

func (s *Service) HandleCreateVariant(c *gin.Context) {
	uid, _ := currentUserID(c)
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	var req struct {
		SKU                 string `json:"sku" binding:"required"`
		Name                string `json:"name"`
		PriceCents          int64  `json:"price_cents" binding:"required"`
		CompareAtPriceCents *int64 `json:"compare_at_price_cents"`
		Currency            string `json:"currency"`
		WeightGrams         *int32 `json:"weight_grams"`
		IsActive            *bool  `json:"is_active"`
		QuantityAvailable   int32  `json:"quantity_available"`
		LowStockThreshold   int32  `json:"low_stock_threshold"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PriceCents < 0 {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid variant request.")
		return
	}
	if req.QuantityAvailable < 0 {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_inventory", "Initial quantity cannot be negative.")
		return
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	variant, item, movement, err := s.createVariantTx(c.Request.Context(), seller.ID, productID, pgxutil.UUID(uid), req.SKU, req.Name, req.PriceCents, req.CompareAtPriceCents, req.Currency, req.WeightGrams, active, req.QuantityAvailable, req.LowStockThreshold)
	if err != nil {
		apiresponse.Error(c, http.StatusConflict, "variant_create_failed", "Could not create variant.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, gin.H{"variant": variantJSON(variant), "inventory": inventoryJSON(item), "movement": movementPtrJSON(movement)}, nil)
}

func (s *Service) createVariantTx(ctx context.Context, sellerID, productID, actorID pgtype.UUID, sku, name string, priceCents int64, compareAt *int64, currency string, weight *int32, active bool, quantityAvailable int32, lowStockThreshold int32) (sellerstore.ProductVariant, sellerstore.InventoryItem, *sellerstore.InventoryMovement, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return sellerstore.ProductVariant{}, sellerstore.InventoryItem{}, nil, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	if _, err := q.GetSellerProduct(ctx, sellerstore.GetSellerProductParams{ID: productID, SellerID: sellerID}); err != nil {
		return sellerstore.ProductVariant{}, sellerstore.InventoryItem{}, nil, err
	}
	variant, err := q.CreateProductVariant(ctx, sellerstore.CreateProductVariantParams{ProductID: productID, Sku: sku, Name: text(name), PriceCents: priceCents, CompareAtPriceCents: int8Ptr(compareAt), Column6: currency, WeightGrams: int4Ptr(weight), Column8: active})
	if err != nil {
		return sellerstore.ProductVariant{}, sellerstore.InventoryItem{}, nil, err
	}
	threshold := lowStockThreshold
	if threshold == 0 {
		threshold = 5
	}
	item, err := q.UpsertInventoryItem(ctx, sellerstore.UpsertInventoryItemParams{ProductID: productID, VariantID: variant.ID, QuantityAvailable: quantityAvailable, Column4: threshold})
	if err != nil {
		return sellerstore.ProductVariant{}, sellerstore.InventoryItem{}, nil, err
	}
	var movement *sellerstore.InventoryMovement
	if quantityAvailable > 0 {
		created, err := q.CreateInventoryMovement(ctx, sellerstore.CreateInventoryMovementParams{InventoryItemID: item.ID, MovementType: "stock_in", Quantity: quantityAvailable, Reason: text("initial stock"), ReferenceType: text("variant"), ReferenceID: variant.ID, ActorUserID: actorID, BeforeAvailable: pgtype.Int4{Int32: 0, Valid: true}, AfterAvailable: pgtype.Int4{Int32: item.QuantityAvailable, Valid: true}, BeforeReserved: pgtype.Int4{Int32: 0, Valid: true}, AfterReserved: pgtype.Int4{Int32: item.QuantityReserved, Valid: true}})
		if err != nil {
			return sellerstore.ProductVariant{}, sellerstore.InventoryItem{}, nil, err
		}
		movement = &created
	}
	if err := tx.Commit(ctx); err != nil {
		return sellerstore.ProductVariant{}, sellerstore.InventoryItem{}, nil, err
	}
	return variant, item, movement, nil
}

func (s *Service) HandleUpdateVariant(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	variantID, ok := uuidParam(c, "variant_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_variant_id", "Invalid variant id.")
		return
	}
	var req struct {
		SKU                 string `json:"sku"`
		Name                string `json:"name"`
		PriceCents          *int64 `json:"price_cents"`
		CompareAtPriceCents *int64 `json:"compare_at_price_cents"`
		Currency            string `json:"currency"`
		WeightGrams         *int32 `json:"weight_grams"`
		IsActive            *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid variant update.")
		return
	}
	variant, err := s.queries.UpdateProductVariant(c.Request.Context(), sellerstore.UpdateProductVariantParams{ID: variantID, SellerID: seller.ID, Sku: text(req.SKU), Name: text(req.Name), PriceCents: int8Ptr(req.PriceCents), CompareAtPriceCents: int8Ptr(req.CompareAtPriceCents), Currency: text(req.Currency), WeightGrams: int4Ptr(req.WeightGrams), IsActive: boolPtr(req.IsActive)})
	if err != nil {
		writeNotFound(c, err, "variant_not_found", "Variant not found.")
		return
	}
	apiresponse.OK(c, variantJSON(variant))
}

func (s *Service) HandleDeleteVariant(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	variantID, ok := uuidParam(c, "variant_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_variant_id", "Invalid variant id.")
		return
	}
	if err := s.queries.DeleteProductVariant(c.Request.Context(), sellerstore.DeleteProductVariantParams{ID: variantID, SellerID: seller.ID}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete variant.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleListOptions(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	options, err := s.queries.ListVariantOptionsForSeller(c.Request.Context(), sellerstore.ListVariantOptionsForSellerParams{ProductID: productID, SellerID: seller.ID})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load options.")
		return
	}
	out := make([]gin.H, 0, len(options))
	for _, option := range options {
		values, _ := s.queries.ListVariantOptionValues(c.Request.Context(), option.ID)
		item := optionJSON(option)
		item["values"] = mapSlice(values, optionValueJSON)
		out = append(out, item)
	}
	apiresponse.OK(c, out)
}

func (s *Service) HandleCreateOption(c *gin.Context) {
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
		Name      string `json:"name" binding:"required"`
		SortOrder int32  `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid option request.")
		return
	}
	option, err := s.queries.CreateVariantOption(c.Request.Context(), sellerstore.CreateVariantOptionParams{ProductID: productID, Name: req.Name, Column3: req.SortOrder})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "option_create_failed", "Could not create option.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, optionJSON(option), nil)
}

func (s *Service) HandleCreateOptionValue(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	optionID, ok := uuidParam(c, "option_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_option_id", "Invalid option id.")
		return
	}
	var req struct {
		Value     string `json:"value" binding:"required"`
		SortOrder int32  `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid option value request.")
		return
	}
	value, err := s.queries.CreateVariantOptionValue(c.Request.Context(), sellerstore.CreateVariantOptionValueParams{OptionID: optionID, SellerID: seller.ID, Value: req.Value, SortOrder: req.SortOrder})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "option_value_create_failed", "Could not create option value.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, optionValueJSON(value), nil)
}

func (s *Service) HandleDeleteOption(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	optionID, ok := uuidParam(c, "option_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_option_id", "Invalid option id.")
		return
	}
	if err := s.queries.DeleteVariantOption(c.Request.Context(), sellerstore.DeleteVariantOptionParams{ID: optionID, SellerID: seller.ID}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete option.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleListInventory(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	p := pagination(c)
	items, err := s.queries.ListInventoryForSeller(c.Request.Context(), sellerstore.ListInventoryForSellerParams{SellerID: seller.ID, Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load inventory.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(items, inventoryJSON), "pagination": gin.H{"page": p.Page, "per_page": p.PerPage}})
}

func (s *Service) HandleLowStock(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	p := pagination(c)
	items, err := s.queries.ListLowStockForSeller(c.Request.Context(), sellerstore.ListLowStockForSellerParams{SellerID: seller.ID, Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load low stock.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(items, inventoryJSON)})
}

func (s *Service) HandleGetInventory(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	itemID, ok := uuidParam(c, "inventory_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_inventory_id", "Invalid inventory id.")
		return
	}
	item, err := s.queries.GetInventoryItemForSeller(c.Request.Context(), sellerstore.GetInventoryItemForSellerParams{ID: itemID, SellerID: seller.ID})
	if err != nil {
		writeNotFound(c, err, "inventory_not_found", "Inventory item not found.")
		return
	}
	apiresponse.OK(c, inventoryJSON(item))
}

func (s *Service) HandleUpdateInventory(c *gin.Context) {
	s.inventoryAdjust(c, false)
}

func (s *Service) HandleAdjustInventory(c *gin.Context) {
	s.inventoryAdjust(c, true)
}

func (s *Service) inventoryAdjust(c *gin.Context, deltaMode bool) {
	uid, _ := currentUserID(c)
	seller, _ := sellerFromContext(c)
	itemID, ok := uuidParam(c, "inventory_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_inventory_id", "Invalid inventory id.")
		return
	}
	var req struct {
		QuantityAvailable *int32 `json:"quantity_available"`
		QuantityReserved  *int32 `json:"quantity_reserved"`
		LowStockThreshold *int32 `json:"low_stock_threshold"`
		DeltaAvailable    int32  `json:"delta_available"`
		DeltaReserved     int32  `json:"delta_reserved"`
		Reason            string `json:"reason"`
		ReferenceType     string `json:"reference_type"`
		ReferenceID       string `json:"reference_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid inventory request.")
		return
	}
	item, movement, err := s.adjustInventoryTx(c.Request.Context(), seller.ID, itemID, pgxutil.UUID(uid), req.QuantityAvailable, req.QuantityReserved, req.LowStockThreshold, req.DeltaAvailable, req.DeltaReserved, deltaMode, req.Reason, req.ReferenceType, parseUUIDString(req.ReferenceID))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "inventory_adjust_failed", err.Error())
		return
	}
	apiresponse.OK(c, gin.H{"inventory": inventoryJSON(item), "movement": movementPtrJSON(movement)})
}

func (s *Service) adjustInventoryTx(ctx context.Context, sellerID, itemID, actorID pgtype.UUID, available, reserved, threshold *int32, deltaAvailable, deltaReserved int32, deltaMode bool, reason, refType string, refID pgtype.UUID) (sellerstore.InventoryItem, *sellerstore.InventoryMovement, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return sellerstore.InventoryItem{}, nil, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	updated, movement, err := adjustInventoryWithQueries(ctx, q, sellerID, itemID, actorID, available, reserved, threshold, deltaAvailable, deltaReserved, deltaMode, reason, refType, refID)
	if err != nil {
		return sellerstore.InventoryItem{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return sellerstore.InventoryItem{}, nil, err
	}
	return updated, movement, nil
}

func adjustInventoryWithQueries(ctx context.Context, q *sellerstore.Queries, sellerID, itemID, actorID pgtype.UUID, available, reserved, threshold *int32, deltaAvailable, deltaReserved int32, deltaMode bool, reason, refType string, refID pgtype.UUID) (sellerstore.InventoryItem, *sellerstore.InventoryMovement, error) {
	current, err := q.LockInventoryItemForSeller(ctx, sellerstore.LockInventoryItemForSellerParams{ID: itemID, SellerID: sellerID})
	if err != nil {
		return sellerstore.InventoryItem{}, nil, err
	}
	nextAvailable := current.QuantityAvailable
	nextReserved := current.QuantityReserved
	if deltaMode {
		nextAvailable += deltaAvailable
		nextReserved += deltaReserved
	} else {
		if available != nil {
			nextAvailable = *available
		}
		if reserved != nil {
			nextReserved = *reserved
		}
	}
	nextThreshold := current.LowStockThreshold
	if threshold != nil {
		nextThreshold = *threshold
	}
	if err := validateInventoryState(nextAvailable, nextReserved, nextThreshold); err != nil {
		return sellerstore.InventoryItem{}, nil, err
	}
	updated, err := q.UpdateInventoryQuantities(ctx, sellerstore.UpdateInventoryQuantitiesParams{ID: itemID, QuantityAvailable: nextAvailable, QuantityReserved: nextReserved, LowStockThreshold: nextThreshold})
	if err != nil {
		return sellerstore.InventoryItem{}, nil, err
	}
	moveQty := movementQuantity(current.QuantityAvailable, updated.QuantityAvailable, current.QuantityReserved, updated.QuantityReserved)
	if moveQty == 0 {
		return updated, nil, nil
	}
	movementType := "adjustment"
	if deltaMode && deltaReserved > 0 && deltaAvailable == 0 {
		movementType = "reserve"
	} else if deltaMode && deltaReserved < 0 && deltaAvailable == 0 {
		movementType = "release"
	} else if updated.QuantityAvailable > current.QuantityAvailable {
		movementType = "stock_in"
	} else if updated.QuantityAvailable < current.QuantityAvailable {
		movementType = "stock_out"
	}
	movement, err := q.CreateInventoryMovement(ctx, sellerstore.CreateInventoryMovementParams{InventoryItemID: itemID, MovementType: movementType, Quantity: moveQty, Reason: text(reason), ReferenceType: text(refType), ReferenceID: refID, ActorUserID: actorID, BeforeAvailable: pgtype.Int4{Int32: current.QuantityAvailable, Valid: true}, AfterAvailable: pgtype.Int4{Int32: updated.QuantityAvailable, Valid: true}, BeforeReserved: pgtype.Int4{Int32: current.QuantityReserved, Valid: true}, AfterReserved: pgtype.Int4{Int32: updated.QuantityReserved, Valid: true}})
	if err != nil {
		return sellerstore.InventoryItem{}, nil, err
	}
	return updated, &movement, nil
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func movementQuantity(beforeAvailable, afterAvailable, beforeReserved, afterReserved int32) int32 {
	return abs32(afterAvailable-beforeAvailable) + abs32(afterReserved-beforeReserved)
}

func validateInventoryState(available, reserved, threshold int32) error {
	if available < 0 || reserved < 0 {
		return errors.New("inventory quantities cannot be negative")
	}
	if reserved > available {
		return errors.New("reserved quantity cannot exceed available quantity")
	}
	if threshold < 0 {
		return errors.New("low stock threshold cannot be negative")
	}
	return nil
}

func (s *Service) HandleInventoryMovements(c *gin.Context) {
	seller, _ := sellerFromContext(c)
	itemID, ok := uuidParam(c, "inventory_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_inventory_id", "Invalid inventory id.")
		return
	}
	p := pagination(c)
	movements, err := s.queries.ListInventoryMovementsForSeller(c.Request.Context(), sellerstore.ListInventoryMovementsForSellerParams{InventoryItemID: itemID, SellerID: seller.ID, Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load inventory movements.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(movements, movementJSON)})
}

func (s *Service) HandleBulkUpdateInventory(c *gin.Context) {
	uid, _ := currentUserID(c)
	seller, _ := sellerFromContext(c)
	var req struct {
		Items []struct {
			InventoryID       string `json:"inventory_id"`
			QuantityAvailable *int32 `json:"quantity_available"`
			QuantityReserved  *int32 `json:"quantity_reserved"`
			LowStockThreshold *int32 `json:"low_stock_threshold"`
			Reason            string `json:"reason"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid bulk inventory request.")
		return
	}
	out := make([]gin.H, 0, len(req.Items))
	tx, err := s.pool.Begin(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not start bulk inventory update.")
		return
	}
	defer tx.Rollback(c.Request.Context())
	q := s.queries.WithTx(tx)
	for _, item := range req.Items {
		id, err := uuid.Parse(item.InventoryID)
		if err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_inventory_id", "Invalid inventory id in bulk update.")
			return
		}
		updated, movement, err := adjustInventoryWithQueries(c.Request.Context(), q, seller.ID, pgxutil.UUID(id), pgxutil.UUID(uid), item.QuantityAvailable, item.QuantityReserved, item.LowStockThreshold, 0, 0, false, item.Reason, "bulk_update", pgtype.UUID{})
		if err != nil {
			apiresponse.Error(c, http.StatusBadRequest, "inventory_update_failed", err.Error())
			return
		}
		out = append(out, gin.H{"inventory": inventoryJSON(updated), "movement": movementPtrJSON(movement)})
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not commit bulk inventory update.")
		return
	}
	apiresponse.OK(c, gin.H{"updated": out, "count": len(out)})
}

func (s *Service) HandleAdminListApplications(c *gin.Context) {
	p := pagination(c)
	status := text(c.Query("status"))
	apps, err := s.queries.ListSellerApplicationsForAdmin(c.Request.Context(), sellerstore.ListSellerApplicationsForAdminParams{Status: status, Limit: p.Limit, Offset: p.Offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load applications.")
		return
	}
	apiresponse.OK(c, gin.H{"items": mapSlice(apps, applicationJSON)})
}

func (s *Service) HandleAdminReviewApplication(c *gin.Context) {
	adminID, _ := currentUserID(c)
	appID, ok := uuidParam(c, "application_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_application_id", "Invalid application id.")
		return
	}
	var req struct {
		Status          string `json:"status" binding:"required"`
		RejectionReason string `json:"rejection_reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != "approved" && req.Status != "rejected" && req.Status != "in_review") {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid review request.")
		return
	}
	app, seller, err := s.reviewApplicationTx(c.Request.Context(), appID, pgxutil.UUID(adminID), req.Status, req.RejectionReason)
	if err != nil {
		writeNotFound(c, err, "application_not_found", "Application not found or cannot be reviewed.")
		return
	}
	payload := applicationJSON(app)
	if seller != nil {
		payload["seller"] = sellerJSON(*seller)
	}
	apiresponse.OK(c, payload)
}

func (s *Service) reviewApplicationTx(ctx context.Context, appID, adminID pgtype.UUID, status, rejectionReason string) (sellerstore.SellerApplication, *sellerstore.Seller, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return sellerstore.SellerApplication{}, nil, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	app, err := q.ReviewSellerApplication(ctx, sellerstore.ReviewSellerApplicationParams{ID: appID, Status: status, ReviewedBy: adminID, RejectionReason: text(rejectionReason)})
	if err != nil {
		return sellerstore.SellerApplication{}, nil, err
	}
	var seller *sellerstore.Seller
	if status == "approved" {
		created, err := q.CreateSellerFromApplication(ctx, app.ID)
		if err != nil {
			return sellerstore.SellerApplication{}, nil, err
		}
		if err := q.AssignSellerRole(ctx, app.UserID); err != nil {
			return sellerstore.SellerApplication{}, nil, err
		}
		seller = &created
	}
	if err := tx.Commit(ctx); err != nil {
		return sellerstore.SellerApplication{}, nil, err
	}
	return app, seller, nil
}

func (s *Service) HandleAdminSellerStatus(c *gin.Context) {
	sellerID, ok := uuidParam(c, "seller_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_seller_id", "Invalid seller id.")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != "active" && req.Status != "suspended" && req.Status != "rejected" && req.Status != "pending") {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid seller status.")
		return
	}
	seller, err := s.queries.SuspendSellerForAdmin(c.Request.Context(), sellerstore.SuspendSellerForAdminParams{ID: sellerID, Status: req.Status})
	if err != nil {
		writeNotFound(c, err, "seller_not_found", "Seller not found.")
		return
	}
	apiresponse.OK(c, sellerJSON(seller))
}

func (s *Service) HandleAdminModerateProduct(c *gin.Context) {
	adminID, _ := currentUserID(c)
	productID, ok := uuidParam(c, "product_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_product_id", "Invalid product id.")
		return
	}
	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != "approved" && req.Status != "rejected" && req.Status != "pending") {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid product moderation request.")
		return
	}
	product, err := s.queries.ModerateProductForAdmin(c.Request.Context(), sellerstore.ModerateProductForAdminParams{ID: productID, ModerationStatus: req.Status, ModerationReason: text(req.Reason), ModeratedBy: pgxutil.UUID(adminID)})
	if err != nil {
		writeNotFound(c, err, "product_not_found", "Product not found.")
		return
	}
	apiresponse.OK(c, product)
}

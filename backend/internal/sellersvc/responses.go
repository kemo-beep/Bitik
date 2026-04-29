package sellersvc

import (
	"encoding/json"

	sellerstore "github.com/bitik/backend/internal/store/sellers"
	"github.com/gin-gonic/gin"
)

func applicationJSON(a sellerstore.SellerApplication) gin.H {
	return gin.H{
		"id":               uuidString(a.ID),
		"user_id":          uuidString(a.UserID),
		"shop_name":        a.ShopName,
		"slug":             a.Slug,
		"business_type":    textValue(a.BusinessType),
		"country":          textValue(a.Country),
		"currency":         a.Currency,
		"status":           statusString(a.Status),
		"submitted_at":     a.SubmittedAt.Time,
		"reviewed_by":      nullableUUID(a.ReviewedBy),
		"reviewed_at":      a.ReviewedAt.Time,
		"rejection_reason": textValue(a.RejectionReason),
		"metadata":         json.RawMessage(a.Metadata),
		"created_at":       a.CreatedAt.Time,
		"updated_at":       a.UpdatedAt.Time,
	}
}

func sellerJSON(s sellerstore.Seller) gin.H {
	return gin.H{
		"id":             uuidString(s.ID),
		"user_id":        uuidString(s.UserID),
		"application_id": nullableUUID(s.ApplicationID),
		"shop_name":      s.ShopName,
		"slug":           s.Slug,
		"description":    textValue(s.Description),
		"logo_url":       textValue(s.LogoUrl),
		"banner_url":     textValue(s.BannerUrl),
		"status":         statusString(s.Status),
		"rating":         numericString(s.Rating),
		"total_sales":    s.TotalSales,
		"settings":       json.RawMessage(s.Settings),
		"created_at":     s.CreatedAt.Time,
		"updated_at":     s.UpdatedAt.Time,
	}
}

func documentJSON(d sellerstore.SellerDocument) gin.H {
	return gin.H{
		"id":                    uuidString(d.ID),
		"seller_application_id": nullableUUID(d.SellerApplicationID),
		"seller_id":             nullableUUID(d.SellerID),
		"document_type":         d.DocumentType,
		"file_url":              d.FileUrl,
		"status":                d.Status,
		"metadata":              json.RawMessage(d.Metadata),
		"created_at":            d.CreatedAt.Time,
	}
}

func productJSON(
	id, sellerID, categoryID, brandID string,
	name, slug string,
	description any,
	status any,
	minPrice, maxPrice int64,
	currency string,
	totalSold int64,
	rating string,
	reviewCount int64,
	moderationStatus string,
	moderationReason any,
) gin.H {
	return gin.H{
		"id":                id,
		"seller_id":         sellerID,
		"category_id":       categoryID,
		"brand_id":          brandID,
		"name":              name,
		"slug":              slug,
		"description":       description,
		"status":            statusString(status),
		"min_price_cents":   minPrice,
		"max_price_cents":   maxPrice,
		"currency":          currency,
		"total_sold":        totalSold,
		"rating":            rating,
		"review_count":      reviewCount,
		"moderation_status": moderationStatus,
		"moderation_reason": moderationReason,
	}
}

func productFromCreate(p sellerstore.CreateSellerProductRow) gin.H {
	return productJSON(uuidString(p.ID), uuidString(p.SellerID), uuidString(p.CategoryID), uuidString(p.BrandID), p.Name, p.Slug, textValue(p.Description), p.Status, p.MinPriceCents, p.MaxPriceCents, p.Currency, p.TotalSold, numericString(p.Rating), p.ReviewCount, p.ModerationStatus, textValue(p.ModerationReason))
}

func productFromList(p sellerstore.ListSellerProductsRow) gin.H {
	return productJSON(uuidString(p.ID), uuidString(p.SellerID), uuidString(p.CategoryID), uuidString(p.BrandID), p.Name, p.Slug, textValue(p.Description), p.Status, p.MinPriceCents, p.MaxPriceCents, p.Currency, p.TotalSold, numericString(p.Rating), p.ReviewCount, p.ModerationStatus, textValue(p.ModerationReason))
}

func productFromGet(p sellerstore.GetSellerProductRow) gin.H {
	return productJSON(uuidString(p.ID), uuidString(p.SellerID), uuidString(p.CategoryID), uuidString(p.BrandID), p.Name, p.Slug, textValue(p.Description), p.Status, p.MinPriceCents, p.MaxPriceCents, p.Currency, p.TotalSold, numericString(p.Rating), p.ReviewCount, p.ModerationStatus, textValue(p.ModerationReason))
}

func imageJSON(img sellerstore.ProductImage) gin.H {
	return gin.H{"id": uuidString(img.ID), "product_id": uuidString(img.ProductID), "url": img.Url, "alt_text": textValue(img.AltText), "sort_order": img.SortOrder, "is_primary": img.IsPrimary}
}

func variantJSON(v sellerstore.ProductVariant) gin.H {
	return gin.H{"id": uuidString(v.ID), "product_id": uuidString(v.ProductID), "sku": v.Sku, "name": textValue(v.Name), "price_cents": v.PriceCents, "compare_at_price_cents": int8Value(v.CompareAtPriceCents), "currency": v.Currency, "weight_grams": int4Value(v.WeightGrams), "is_active": v.IsActive}
}

func optionJSON(o sellerstore.VariantOption) gin.H {
	return gin.H{"id": uuidString(o.ID), "product_id": uuidString(o.ProductID), "name": o.Name, "sort_order": o.SortOrder}
}

func optionValueJSON(v sellerstore.VariantOptionValue) gin.H {
	return gin.H{"id": uuidString(v.ID), "option_id": uuidString(v.OptionID), "value": v.Value, "sort_order": v.SortOrder}
}

func inventoryJSON(i sellerstore.InventoryItem) gin.H {
	return gin.H{"id": uuidString(i.ID), "product_id": uuidString(i.ProductID), "variant_id": uuidString(i.VariantID), "quantity_available": i.QuantityAvailable, "quantity_reserved": i.QuantityReserved, "low_stock_threshold": i.LowStockThreshold, "created_at": i.CreatedAt.Time, "updated_at": i.UpdatedAt.Time}
}

func movementJSON(m sellerstore.InventoryMovement) gin.H {
	return gin.H{"id": uuidString(m.ID), "inventory_item_id": uuidString(m.InventoryItemID), "movement_type": statusString(m.MovementType), "quantity": m.Quantity, "reason": textValue(m.Reason), "reference_type": textValue(m.ReferenceType), "reference_id": nullableUUID(m.ReferenceID), "actor_user_id": nullableUUID(m.ActorUserID), "before_available": int4Value(m.BeforeAvailable), "after_available": int4Value(m.AfterAvailable), "before_reserved": int4Value(m.BeforeReserved), "after_reserved": int4Value(m.AfterReserved), "created_at": m.CreatedAt.Time}
}

func movementPtrJSON(m *sellerstore.InventoryMovement) any {
	if m == nil {
		return nil
	}
	return movementJSON(*m)
}

func mapSlice[T any](items []T, mapper func(T) gin.H) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, mapper(item))
	}
	return out
}

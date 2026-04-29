package promotionsvc

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	orderstore "github.com/bitik/backend/internal/store/orders"
	promostore "github.com/bitik/backend/internal/store/promotions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func parsePage(c *gin.Context) (page int32, limit int32, offset int32) {
	page = 1
	limit = 20
	if v, err := strconv.ParseInt(strings.TrimSpace(c.DefaultQuery("page", "1")), 10, 32); err == nil && v > 0 {
		page = int32(v)
	}
	if v, err := strconv.ParseInt(strings.TrimSpace(c.DefaultQuery("limit", "20")), 10, 32); err == nil && v > 0 {
		limit = int32(v)
	}
	if limit > 100 {
		limit = 100
	}
	offset = (page - 1) * limit
	return
}

func pageMeta(page, limit int32, total int64) map[string]any {
	hasNext := int64(page)*int64(limit) < total
	return map[string]any{"page": page, "limit": limit, "total": total, "has_next": hasNext}
}

func (s *Service) HandleBuyerListVouchers(c *gin.Context) {
	page, limit, offset := parsePage(c)
	var sellerID pgtype.UUID
	if raw := strings.TrimSpace(c.Query("seller_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "seller_id", Message: "must be a valid uuid"}})
			return
		}
		sellerID = pgxutil.UUID(id)
	}
	total, err := s.promo.CountActiveVouchers(c.Request.Context(), sellerID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load vouchers.")
		return
	}
	rows, err := s.promo.ListActiveVouchers(c.Request.Context(), promostore.ListActiveVouchersParams{
		SellerID: sellerID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load vouchers.")
		return
	}
	apiresponse.Respond(c, http.StatusOK, gin.H{"items": mapSlice(rows, voucherJSONPromo)}, map[string]any{"pagination": pageMeta(page, limit, total)})
}

func (s *Service) HandleBuyerValidateVoucher(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid voucher request.")
		return
	}
	voucher, err := s.orders.GetActiveVoucherByCode(c.Request.Context(), req.Code)
	if err != nil {
		apiresponse.Error(c, http.StatusNotFound, "voucher_not_found", "Voucher not found.")
		return
	}

	cart, err := s.orders.GetOrCreateCart(c.Request.Context(), orderstore.GetOrCreateCartParams{UserID: pgxutil.UUID(uid)})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	items, err := s.orders.ListCartItemsDetailed(c.Request.Context(), cart.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	subtotal := int64(0)
	for _, item := range items {
		if !item.Selected {
			continue
		}
		subtotal += int64(item.Quantity) * item.PriceCents
	}
	scopeOK := voucherScopeMatchesCart(voucher, items)
	discount := int64(0)
	if scopeOK {
		discount = calculateDiscount(voucher, subtotal, 0)
	}
	apiresponse.OK(c, gin.H{
		"valid":        scopeOK && discount > 0,
		"scope_ok":     scopeOK,
		"discount_cents": discount,
		"subtotal_cents": subtotal,
		"voucher":      voucherJSONOrder(voucher),
	})
}

func (s *Service) HandleSellerListVouchers(c *gin.Context) {
	seller, ok := sellerFromContext(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing seller context.")
		return
	}
	page, limit, offset := parsePage(c)
	total, err := s.promo.CountSellerVouchers(c.Request.Context(), seller.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load vouchers.")
		return
	}
	rows, err := s.promo.ListSellerVouchers(c.Request.Context(), promostore.ListSellerVouchersParams{SellerID: seller.ID, Limit: limit, Offset: offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load vouchers.")
		return
	}
	apiresponse.Respond(c, http.StatusOK, gin.H{"items": mapSlice(rows, voucherJSONPromo)}, map[string]any{"pagination": pageMeta(page, limit, total)})
}

func (s *Service) HandleSellerCreateVoucher(c *gin.Context) {
	seller, ok := sellerFromContext(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing seller context.")
		return
	}
	var req createVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid voucher request.")
		return
	}
	if err := req.Validate(); err != nil {
		apiresponse.ValidationError(c, err)
		return
	}
	row, err := s.promo.CreateVoucher(c.Request.Context(), promostore.CreateVoucherParams{
		Code:            req.Code,
		Title:           req.Title,
		Description:     text(req.Description),
		DiscountType:    req.DiscountType,
		DiscountValue:   req.DiscountValue,
		MinOrderCents:   req.MinOrderCents,
		MaxDiscountCents: int8(req.MaxDiscountCents),
		UsageLimit:      int4(req.UsageLimit),
		StartsAt:        pgtype.Timestamptz{Time: req.StartsAt, Valid: true},
		EndsAt:          pgtype.Timestamptz{Time: req.EndsAt, Valid: true},
		IsActive:        req.IsActive,
		SellerID:        seller.ID,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "voucher_create_failed", "Could not create voucher.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, voucherJSONPromo(row), nil)
}

func (s *Service) HandleSellerUpdateVoucher(c *gin.Context) {
	seller, ok := sellerFromContext(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing seller context.")
		return
	}
	id, ok := parseUUIDParam(c, "voucher_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_voucher_id", "Invalid voucher id.")
		return
	}
	var req updateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid voucher update.")
		return
	}
	if err := req.Validate(); err != nil {
		apiresponse.ValidationError(c, err)
		return
	}
	row, err := s.promo.UpdateVoucher(c.Request.Context(), promostore.UpdateVoucherParams{
		ID:             id,
		Title:          text(req.Title),
		Description:    text(req.Description),
		DiscountType:   text(req.DiscountType),
		DiscountValue:  int8(req.DiscountValue),
		MinOrderCents:  int8(req.MinOrderCents),
		MaxDiscountCents: int8(req.MaxDiscountCents),
		UsageLimit:     int4(req.UsageLimit),
		StartsAt:       timestamptz(req.StartsAt),
		EndsAt:         timestamptz(req.EndsAt),
		IsActive:       boolPtr(req.IsActive),
	})
	if err != nil || (row.SellerID.Valid && row.SellerID != seller.ID) {
		apiresponse.Error(c, http.StatusBadRequest, "voucher_update_failed", "Could not update voucher.")
		return
	}
	apiresponse.OK(c, voucherJSONPromo(row))
}

func (s *Service) HandleSellerDeleteVoucher(c *gin.Context) {
	seller, ok := sellerFromContext(c)
	if !ok {
		apiresponse.Error(c, http.StatusUnauthorized, "unauthorized", "Missing seller context.")
		return
	}
	id, ok := parseUUIDParam(c, "voucher_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_voucher_id", "Invalid voucher id.")
		return
	}
	// Enforce seller ownership by checking first.
	row, err := s.orders.GetVoucherByID(c.Request.Context(), id)
	if err != nil || !row.SellerID.Valid || row.SellerID != seller.ID {
		apiresponse.Error(c, http.StatusNotFound, "voucher_not_found", "Voucher not found.")
		return
	}
	if err := s.promo.DeleteVoucher(c.Request.Context(), id); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete voucher.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleAdminListVouchers(c *gin.Context) {
	page, limit, offset := parsePage(c)
	total, err := s.promo.CountAllVouchersAdmin(c.Request.Context())
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load vouchers.")
		return
	}
	rows, err := s.promo.ListAllVouchersAdmin(c.Request.Context(), promostore.ListAllVouchersAdminParams{Limit: limit, Offset: offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load vouchers.")
		return
	}
	apiresponse.Respond(c, http.StatusOK, gin.H{"items": mapSlice(rows, voucherJSONPromo)}, map[string]any{"pagination": pageMeta(page, limit, total)})
}

func (s *Service) HandleAdminCreateVoucher(c *gin.Context) {
	var req createVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid voucher request.")
		return
	}
	if err := req.Validate(); err != nil {
		apiresponse.ValidationError(c, err)
		return
	}
	var sellerID pgtype.UUID
	if strings.TrimSpace(req.SellerID) != "" {
		id, err := uuid.Parse(req.SellerID)
		if err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "seller_id", Message: "must be a valid uuid"}})
			return
		}
		sellerID = pgxutil.UUID(id)
	}
	row, err := s.promo.CreateVoucher(c.Request.Context(), promostore.CreateVoucherParams{
		Code:            req.Code,
		Title:           req.Title,
		Description:     text(req.Description),
		DiscountType:    req.DiscountType,
		DiscountValue:   req.DiscountValue,
		MinOrderCents:   req.MinOrderCents,
		MaxDiscountCents: int8(req.MaxDiscountCents),
		UsageLimit:      int4(req.UsageLimit),
		StartsAt:        pgtype.Timestamptz{Time: req.StartsAt, Valid: true},
		EndsAt:          pgtype.Timestamptz{Time: req.EndsAt, Valid: true},
		IsActive:        req.IsActive,
		SellerID:        sellerID,
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "voucher_create_failed", "Could not create voucher.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, voucherJSONPromo(row), nil)
}

func (s *Service) HandleAdminUpdateVoucher(c *gin.Context) {
	id, ok := parseUUIDParam(c, "voucher_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_voucher_id", "Invalid voucher id.")
		return
	}
	var req updateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid voucher update.")
		return
	}
	if err := req.Validate(); err != nil {
		apiresponse.ValidationError(c, err)
		return
	}
	row, err := s.promo.UpdateVoucher(c.Request.Context(), promostore.UpdateVoucherParams{
		ID:             id,
		Title:          text(req.Title),
		Description:    text(req.Description),
		DiscountType:   text(req.DiscountType),
		DiscountValue:  int8(req.DiscountValue),
		MinOrderCents:  int8(req.MinOrderCents),
		MaxDiscountCents: int8(req.MaxDiscountCents),
		UsageLimit:     int4(req.UsageLimit),
		StartsAt:       timestamptz(req.StartsAt),
		EndsAt:         timestamptz(req.EndsAt),
		IsActive:       boolPtr(req.IsActive),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "voucher_update_failed", "Could not update voucher.")
		return
	}
	apiresponse.OK(c, voucherJSONPromo(row))
}

func (s *Service) HandleAdminDeleteVoucher(c *gin.Context) {
	id, ok := parseUUIDParam(c, "voucher_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_voucher_id", "Invalid voucher id.")
		return
	}
	if err := s.promo.DeleteVoucher(c.Request.Context(), id); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete voucher.")
		return
	}
	c.Status(http.StatusNoContent)
}

// helpers / request models

type createVoucherRequest struct {
	Code           string    `json:"code" binding:"required"`
	Title          string    `json:"title" binding:"required"`
	Description    string    `json:"description"`
	DiscountType   string    `json:"discount_type" binding:"required"`
	DiscountValue  int64     `json:"discount_value" binding:"required"`
	MinOrderCents  int64     `json:"min_order_cents"`
	MaxDiscountCents *int64  `json:"max_discount_cents"`
	UsageLimit     *int32    `json:"usage_limit"`
	StartsAt       time.Time `json:"starts_at" binding:"required"`
	EndsAt         time.Time `json:"ends_at" binding:"required"`
	IsActive       bool      `json:"is_active"`
	SellerID       string    `json:"seller_id"` // admin-only
}

func (r createVoucherRequest) Validate() []apiresponse.FieldError {
	var errs []apiresponse.FieldError
	if strings.TrimSpace(r.Code) == "" {
		errs = append(errs, apiresponse.FieldError{Field: "code", Message: "is required"})
	}
	dt := strings.TrimSpace(r.DiscountType)
	switch dt {
	case "fixed", "percentage", "free_shipping":
	default:
		errs = append(errs, apiresponse.FieldError{Field: "discount_type", Message: "must be fixed|percentage|free_shipping"})
	}
	if r.DiscountValue < 0 {
		errs = append(errs, apiresponse.FieldError{Field: "discount_value", Message: "must be >= 0"})
	}
	if r.MinOrderCents < 0 {
		errs = append(errs, apiresponse.FieldError{Field: "min_order_cents", Message: "must be >= 0"})
	}
	if r.EndsAt.Before(r.StartsAt) || r.EndsAt.Equal(r.StartsAt) {
		errs = append(errs, apiresponse.FieldError{Field: "ends_at", Message: "must be after starts_at"})
	}
	return errs
}

type updateVoucherRequest struct {
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	DiscountType    string     `json:"discount_type"`
	DiscountValue   *int64     `json:"discount_value"`
	MinOrderCents   *int64     `json:"min_order_cents"`
	MaxDiscountCents *int64    `json:"max_discount_cents"`
	UsageLimit      *int32     `json:"usage_limit"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	IsActive        *bool      `json:"is_active"`
}

func (r updateVoucherRequest) Validate() []apiresponse.FieldError {
	var errs []apiresponse.FieldError
	if r.DiscountType != "" {
		switch strings.TrimSpace(r.DiscountType) {
		case "fixed", "percentage", "free_shipping":
		default:
			errs = append(errs, apiresponse.FieldError{Field: "discount_type", Message: "must be fixed|percentage|free_shipping"})
		}
	}
	if r.DiscountValue != nil && *r.DiscountValue < 0 {
		errs = append(errs, apiresponse.FieldError{Field: "discount_value", Message: "must be >= 0"})
	}
	if r.MinOrderCents != nil && *r.MinOrderCents < 0 {
		errs = append(errs, apiresponse.FieldError{Field: "min_order_cents", Message: "must be >= 0"})
	}
	if r.StartsAt != nil && r.EndsAt != nil && (!r.EndsAt.After(*r.StartsAt)) {
		errs = append(errs, apiresponse.FieldError{Field: "ends_at", Message: "must be after starts_at"})
	}
	return errs
}

func voucherJSONOrder(x orderstore.Voucher) gin.H {
	return voucherJSONFromFields(
		x.ID, x.Code, x.Title, x.Description, x.DiscountType, x.DiscountValue,
		x.MinOrderCents, x.MaxDiscountCents, x.UsageLimit, x.UsageCount,
		x.StartsAt, x.EndsAt, x.IsActive, x.CreatedAt, x.SellerID,
	)
}

func voucherJSONPromo(x promostore.Voucher) gin.H {
	return voucherJSONFromFields(
		x.ID, x.Code, x.Title, x.Description, x.DiscountType, x.DiscountValue,
		x.MinOrderCents, x.MaxDiscountCents, x.UsageLimit, x.UsageCount,
		x.StartsAt, x.EndsAt, x.IsActive, x.CreatedAt, x.SellerID,
	)
}

func voucherJSONFromFields(
	id pgtype.UUID,
	code, title string,
	description pgtype.Text,
	discountType string,
	discountValue int64,
	minOrderCents int64,
	maxDiscountCents pgtype.Int8,
	usageLimit pgtype.Int4,
	usageCount int32,
	startsAt, endsAt pgtype.Timestamptz,
	isActive bool,
	createdAt pgtype.Timestamptz,
	sellerID pgtype.UUID,
) gin.H {
	return gin.H{
		"id":                uuidString(id),
		"code":              code,
		"title":             title,
		"description":       textValue(description),
		"discount_type":     discountType,
		"discount_value":    discountValue,
		"min_order_cents":   minOrderCents,
		"max_discount_cents": int8Value(maxDiscountCents),
		"usage_limit":       int4Value(usageLimit),
		"usage_count":       usageCount,
		"starts_at":         timeValue(startsAt),
		"ends_at":           timeValue(endsAt),
		"is_active":         isActive,
		"created_at":        timeValue(createdAt),
		"seller_id":         nullableUUID(sellerID),
	}
}

func uuidString(id pgtype.UUID) string {
	if v, ok := pgxutil.ToUUID(id); ok {
		return v.String()
	}
	return ""
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

func int4Value(v pgtype.Int4) any {
	if !v.Valid {
		return nil
	}
	return v.Int32
}

func int8Value(v pgtype.Int8) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}

func text(raw string) pgtype.Text {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: raw, Valid: true}
}

func int4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func int8(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

func timestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func boolPtr(v *bool) pgtype.Bool {
	if v == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *v, Valid: true}
}

func mapSlice[T any](items []T, mapper func(T) gin.H) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, mapper(item))
	}
	return out
}

func voucherScopeMatchesCart(v orderstore.Voucher, items []orderstore.ListCartItemsDetailedRow) bool {
	if !v.SellerID.Valid {
		return true
	}
	seen := map[[16]byte]struct{}{}
	for _, it := range items {
		if !it.Selected {
			continue
		}
		seen[it.SellerID.Bytes] = struct{}{}
		if len(seen) > 1 {
			return false
		}
	}
	if len(seen) == 0 {
		return false
	}
	for sellerBytes := range seen {
		return sellerBytes == v.SellerID.Bytes
	}
	return false
}

func calculateDiscount(v orderstore.Voucher, subtotal, shipping int64) int64 {
	if subtotal < v.MinOrderCents {
		return 0
	}
	discount := int64(0)
	switch v.DiscountType {
	case "fixed":
		discount = v.DiscountValue
	case "percentage":
		discount = int64(float64(subtotal) * float64(v.DiscountValue) / 100.0)
	case "free_shipping":
		discount = shipping
	}
	if v.MaxDiscountCents.Valid && discount > v.MaxDiscountCents.Int64 {
		discount = v.MaxDiscountCents.Int64
	}
	if discount > subtotal+shipping {
		discount = subtotal + shipping
	}
	if discount < 0 {
		return 0
	}
	return discount
}

func parseUUIDParam(c *gin.Context, name string) (pgtype.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		return pgtype.UUID{}, false
	}
	return pgxutil.UUID(id), true
}


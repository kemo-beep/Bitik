package ordersvc

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/pgxutil"
	"github.com/bitik/backend/internal/platform/queue"
	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleListAddresses(c *gin.Context) {
	uid, _ := currentUserID(c)
	items, err := s.queries.ListUserAddresses(c.Request.Context(), pgxutil.UUID(uid))
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load addresses.")
		return
	}
	apiresponse.OK(c, mapSlice(items, addressJSON))
}

func (s *Service) HandleCreateAddress(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		FullName     string `json:"full_name" binding:"required"`
		Phone        string `json:"phone" binding:"required"`
		Country      string `json:"country" binding:"required"`
		State        string `json:"state"`
		City         string `json:"city"`
		District     string `json:"district"`
		PostalCode   string `json:"postal_code"`
		AddressLine1 string `json:"address_line1" binding:"required"`
		AddressLine2 string `json:"address_line2"`
		IsDefault    bool   `json:"is_default"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid address request.")
		return
	}
	if req.IsDefault {
		_ = s.queries.ClearDefaultAddress(c.Request.Context(), pgxutil.UUID(uid))
	}
	address, err := s.queries.CreateUserAddress(c.Request.Context(), orderstore.CreateUserAddressParams{UserID: pgxutil.UUID(uid), FullName: req.FullName, Phone: req.Phone, Country: req.Country, State: text(req.State), City: text(req.City), District: text(req.District), PostalCode: text(req.PostalCode), AddressLine1: req.AddressLine1, AddressLine2: text(req.AddressLine2), IsDefault: req.IsDefault})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "address_create_failed", "Could not create address.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, addressJSON(address), nil)
}

func (s *Service) HandleGetAddress(c *gin.Context) {
	uid, _ := currentUserID(c)
	id, ok := uuidParam(c, "address_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_address_id", "Invalid address id.")
		return
	}
	address, err := s.queries.GetUserAddress(c.Request.Context(), orderstore.GetUserAddressParams{ID: id, UserID: pgxutil.UUID(uid)})
	if err != nil {
		notFoundOrInternal(c, err, "address_not_found", "Address not found.")
		return
	}
	apiresponse.OK(c, addressJSON(address))
}

func (s *Service) HandleUpdateAddress(c *gin.Context) {
	uid, _ := currentUserID(c)
	id, ok := uuidParam(c, "address_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_address_id", "Invalid address id.")
		return
	}
	var req struct {
		FullName     string `json:"full_name"`
		Phone        string `json:"phone"`
		Country      string `json:"country"`
		State        string `json:"state"`
		City         string `json:"city"`
		District     string `json:"district"`
		PostalCode   string `json:"postal_code"`
		AddressLine1 string `json:"address_line1"`
		AddressLine2 string `json:"address_line2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid address update.")
		return
	}
	address, err := s.queries.UpdateUserAddress(c.Request.Context(), orderstore.UpdateUserAddressParams{ID: id, UserID: pgxutil.UUID(uid), FullName: text(req.FullName), Phone: text(req.Phone), Country: text(req.Country), State: text(req.State), City: text(req.City), District: text(req.District), PostalCode: text(req.PostalCode), AddressLine1: text(req.AddressLine1), AddressLine2: text(req.AddressLine2)})
	if err != nil {
		notFoundOrInternal(c, err, "address_not_found", "Address not found.")
		return
	}
	apiresponse.OK(c, addressJSON(address))
}

func (s *Service) HandleDeleteAddress(c *gin.Context) {
	uid, _ := currentUserID(c)
	id, ok := uuidParam(c, "address_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_address_id", "Invalid address id.")
		return
	}
	if err := s.queries.DeleteUserAddress(c.Request.Context(), orderstore.DeleteUserAddressParams{ID: id, UserID: pgxutil.UUID(uid)}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete address.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleSetDefaultAddress(c *gin.Context) {
	uid, _ := currentUserID(c)
	id, ok := uuidParam(c, "address_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_address_id", "Invalid address id.")
		return
	}
	if err := s.queries.ClearDefaultAddress(c.Request.Context(), pgxutil.UUID(uid)); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update default address.")
		return
	}
	address, err := s.queries.SetDefaultAddress(c.Request.Context(), orderstore.SetDefaultAddressParams{ID: id, UserID: pgxutil.UUID(uid)})
	if err != nil {
		notFoundOrInternal(c, err, "address_not_found", "Address not found.")
		return
	}
	apiresponse.OK(c, addressJSON(address))
}

func (s *Service) getOrCreateCart(ctx context.Context, uid uuid.UUID) (orderstore.Cart, error) {
	return s.queries.GetOrCreateCart(ctx, orderstore.GetOrCreateCartParams{UserID: pgxutil.UUID(uid)})
}

func (s *Service) cartPayload(ctx context.Context, cart orderstore.Cart) (gin.H, error) {
	items, err := s.queries.ListCartItemsDetailed(ctx, cart.ID)
	if err != nil {
		return nil, err
	}
	return cartJSON(cart, items), nil
}

func (s *Service) HandleGetCart(c *gin.Context) {
	uid, _ := currentUserID(c)
	cart, err := s.getOrCreateCart(c.Request.Context(), uid)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	payload, err := s.cartPayload(c.Request.Context(), cart)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart items.")
		return
	}
	apiresponse.OK(c, payload)
}

func (s *Service) HandleAddCartItem(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		VariantID string `json:"variant_id" binding:"required"`
		Quantity  int32  `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Quantity < 1 {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid cart item request.")
		return
	}
	variantID, ok := parseUUID(req.VariantID)
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_variant_id", "Invalid variant id.")
		return
	}
	variant, err := s.queries.GetActiveVariantForCart(c.Request.Context(), variantID)
	if err != nil {
		notFoundOrInternal(c, err, "variant_not_found", "Variant not found.")
		return
	}
	if variant.PurchasableQuantity < req.Quantity {
		apiresponse.Error(c, http.StatusBadRequest, "insufficient_inventory", "Not enough inventory is available.")
		return
	}
	cart, err := s.getOrCreateCart(c.Request.Context(), uid)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	item, err := s.queries.UpsertCartItem(c.Request.Context(), orderstore.UpsertCartItemParams{CartID: cart.ID, ProductID: variant.ProductID, VariantID: variant.VariantID, Quantity: req.Quantity})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "cart_add_failed", "Could not add cart item.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, cartItemJSON(item), nil)
}

func (s *Service) HandleUpdateCartItem(c *gin.Context) {
	uid, _ := currentUserID(c)
	itemID, ok := uuidParam(c, "cart_item_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_cart_item_id", "Invalid cart item id.")
		return
	}
	var req struct {
		Quantity int32 `json:"quantity" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Quantity < 1 {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid cart item update.")
		return
	}
	item, err := s.queries.UpdateCartItemQuantity(c.Request.Context(), orderstore.UpdateCartItemQuantityParams{ID: itemID, UserID: pgxutil.UUID(uid), Quantity: req.Quantity})
	if err != nil {
		notFoundOrInternal(c, err, "cart_item_not_found", "Cart item not found.")
		return
	}
	apiresponse.OK(c, cartItemJSON(item))
}

func (s *Service) HandleDeleteCartItem(c *gin.Context) {
	uid, _ := currentUserID(c)
	itemID, ok := uuidParam(c, "cart_item_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_cart_item_id", "Invalid cart item id.")
		return
	}
	if err := s.queries.DeleteCartItem(c.Request.Context(), orderstore.DeleteCartItemParams{ID: itemID, UserID: pgxutil.UUID(uid)}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete cart item.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleClearCart(c *gin.Context) {
	uid, _ := currentUserID(c)
	cart, err := s.getOrCreateCart(c.Request.Context(), uid)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	if err := s.queries.ClearCart(c.Request.Context(), cart.ID); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not clear cart.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleMergeCart(c *gin.Context) {
	uid, _ := currentUserID(c)
	cart, err := s.getOrCreateCart(c.Request.Context(), uid)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	payload, err := s.cartPayload(c.Request.Context(), cart)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart items.")
		return
	}
	apiresponse.OK(c, payload)
}

func (s *Service) HandleSelectCartItems(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		CartItemIDs []string `json:"cart_item_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid cart selection.")
		return
	}
	cart, err := s.getOrCreateCart(c.Request.Context(), uid)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	ids := make([]pgtype.UUID, 0, len(req.CartItemIDs))
	for _, raw := range req.CartItemIDs {
		id, ok := parseUUID(raw)
		if !ok {
			apiresponse.Error(c, http.StatusBadRequest, "invalid_cart_item_id", "Invalid cart item id.")
			return
		}
		ids = append(ids, id)
	}
	if err := s.queries.SetAllCartItemsSelected(c.Request.Context(), orderstore.SetAllCartItemsSelectedParams{CartID: cart.ID, Selected: false}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update cart selection.")
		return
	}
	if len(ids) > 0 {
		if err := s.queries.SelectCartItems(c.Request.Context(), orderstore.SelectCartItemsParams{CartID: cart.ID, Column2: ids}); err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not update cart selection.")
			return
		}
	}
	payload, err := s.cartPayload(c.Request.Context(), cart)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	apiresponse.OK(c, payload)
}

func (s *Service) HandleApplyCartVoucher(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid voucher request.")
		return
	}
	voucher, err := s.queries.GetActiveVoucherByCode(c.Request.Context(), req.Code)
	if err != nil {
		notFoundOrInternal(c, err, "voucher_not_found", "Voucher not found.")
		return
	}
	cart, err := s.getOrCreateCart(c.Request.Context(), uid)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	items, err := s.queries.ListCartItemsDetailed(c.Request.Context(), cart.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart items.")
		return
	}
	if !voucherScopeMatchesCart(voucher, items) {
		apiresponse.Error(c, http.StatusBadRequest, "voucher_scope_invalid", "Voucher is not applicable to this cart.")
		return
	}
	cart, err = s.queries.ApplyCartVoucher(c.Request.Context(), orderstore.ApplyCartVoucherParams{ID: cart.ID, UserID: pgxutil.UUID(uid), VoucherID: voucher.ID})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "voucher_apply_failed", "Could not apply voucher.")
		return
	}
	payload, _ := s.cartPayload(c.Request.Context(), cart)
	payload["voucher"] = voucher.Code
	apiresponse.OK(c, payload)
}

func (s *Service) HandleRemoveCartVoucher(c *gin.Context) {
	uid, _ := currentUserID(c)
	voucherID, ok := uuidParam(c, "voucher_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_voucher_id", "Invalid voucher id.")
		return
	}
	cart, err := s.getOrCreateCart(c.Request.Context(), uid)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load cart.")
		return
	}
	cart, err = s.queries.RemoveCartVoucher(c.Request.Context(), orderstore.RemoveCartVoucherParams{ID: cart.ID, UserID: pgxutil.UUID(uid), VoucherID: voucherID})
	if err != nil {
		notFoundOrInternal(c, err, "voucher_not_found", "Voucher not found on cart.")
		return
	}
	payload, _ := s.cartPayload(c.Request.Context(), cart)
	apiresponse.OK(c, payload)
}

func (s *Service) createCheckoutSessionTx(ctx context.Context, uid uuid.UUID) (orderstore.CheckoutSession, []orderstore.CheckoutSessionItem, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return orderstore.CheckoutSession{}, nil, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	cart, err := q.GetOrCreateCart(ctx, orderstore.GetOrCreateCartParams{UserID: pgxutil.UUID(uid)})
	if err != nil {
		return orderstore.CheckoutSession{}, nil, err
	}
	cartItems, err := q.ListCartItemsDetailed(ctx, cart.ID)
	if err != nil {
		return orderstore.CheckoutSession{}, nil, err
	}
	subtotal := int64(0)
	currency := defaultOrderCurrency
	selected := make([]orderstore.ListCartItemsDetailedRow, 0)
	for _, item := range cartItems {
		if !item.Selected {
			continue
		}
		if item.PurchasableQuantity < item.Quantity {
			return orderstore.CheckoutSession{}, nil, errors.New("selected item is no longer available")
		}
		selected = append(selected, item)
		subtotal += int64(item.Quantity) * item.PriceCents
		currency = item.Currency
	}
	if len(selected) == 0 {
		return orderstore.CheckoutSession{}, nil, errors.New("cart has no selected items")
	}
	discount := int64(0)
	if cart.VoucherID.Valid {
		if voucher, err := q.GetVoucherByID(ctx, cart.VoucherID); err == nil {
			discount = calculateDiscount(voucher, subtotal, defaultShippingCents)
		}
	}
	session, err := q.CreateCheckoutSession(ctx, orderstore.CreateCheckoutSessionParams{UserID: pgxutil.UUID(uid), CartID: cart.ID, Currency: text(currency), SubtotalCents: subtotal, DiscountCents: discount, ShippingCents: defaultShippingCents, TaxCents: defaultTaxCents, TotalCents: subtotal - discount + defaultShippingCents + defaultTaxCents, ExpiresAt: pgtype.Timestamptz{Time: time.Now().UTC().Add(defaultCheckoutTTL), Valid: true}, VoucherID: cart.VoucherID})
	if err != nil {
		return orderstore.CheckoutSession{}, nil, err
	}
	out := make([]orderstore.CheckoutSessionItem, 0, len(selected))
	for _, item := range selected {
		inv, err := q.LockInventoryByVariant(ctx, item.VariantID)
		if err != nil {
			return orderstore.CheckoutSession{}, nil, err
		}
		beforeReserved := inv.QuantityReserved
		updated, err := q.UpdateInventoryReserved(ctx, orderstore.UpdateInventoryReservedParams{ID: inv.ID, DeltaReserved: item.Quantity})
		if err != nil {
			return orderstore.CheckoutSession{}, nil, errors.New("insufficient inventory")
		}
		meta := jsonObject(gin.H{"product_name": item.ProductName, "variant_name": textValue(item.VariantName), "sku": item.Sku, "image_url": item.ImageUrl, "product_slug": item.ProductSlug})
		created, err := q.CreateCheckoutSessionItem(ctx, orderstore.CreateCheckoutSessionItemParams{CheckoutSessionID: session.ID, SellerID: item.SellerID, ProductID: item.ProductID, VariantID: item.VariantID, Quantity: item.Quantity, UnitPriceCents: item.PriceCents, TotalPriceCents: item.PriceCents * int64(item.Quantity), Currency: item.Currency, Metadata: meta})
		if err != nil {
			return orderstore.CheckoutSession{}, nil, err
		}
		out = append(out, created)
		res, err := q.CreateInventoryReservation(ctx, orderstore.CreateInventoryReservationParams{InventoryItemID: inv.ID, UserID: pgxutil.UUID(uid), CheckoutSessionID: session.ID, Quantity: item.Quantity, ExpiresAt: session.ExpiresAt})
		if err != nil {
			return orderstore.CheckoutSession{}, nil, err
		}
		_, err = q.CreateInventoryMovement(ctx, orderstore.CreateInventoryMovementParams{InventoryItemID: inv.ID, MovementType: "reserve", Quantity: item.Quantity, Reason: text("checkout reservation"), ReferenceType: text("checkout_session"), ReferenceID: session.ID, ActorUserID: pgxutil.UUID(uid), BeforeAvailable: pgtype.Int4{Int32: inv.QuantityAvailable, Valid: true}, AfterAvailable: pgtype.Int4{Int32: updated.QuantityAvailable, Valid: true}, BeforeReserved: pgtype.Int4{Int32: beforeReserved, Valid: true}, AfterReserved: pgtype.Int4{Int32: updated.QuantityReserved, Valid: true}})
		if err != nil {
			return orderstore.CheckoutSession{}, nil, err
		}
		_ = res
	}
	if err := tx.Commit(ctx); err != nil {
		return orderstore.CheckoutSession{}, nil, err
	}
	return session, out, nil
}

func (s *Service) HandleCreateCheckoutSession(c *gin.Context) {
	uid, _ := currentUserID(c)
	session, items, err := s.createCheckoutSessionTx(c.Request.Context(), uid)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "checkout_create_failed", err.Error())
		return
	}
	apiresponse.Respond(c, http.StatusCreated, checkoutJSON(session, items), nil)
}

func (s *Service) HandleGetCheckoutSession(c *gin.Context) {
	uid, _ := currentUserID(c)
	sessionID, ok := uuidParam(c, "checkout_session_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_checkout_session_id", "Invalid checkout session id.")
		return
	}
	session, err := s.queries.GetCheckoutSessionForUser(c.Request.Context(), orderstore.GetCheckoutSessionForUserParams{ID: sessionID, UserID: pgxutil.UUID(uid)})
	if err != nil {
		notFoundOrInternal(c, err, "checkout_not_found", "Checkout session not found.")
		return
	}
	items, err := s.queries.ListCheckoutSessionItems(c.Request.Context(), session.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load checkout items.")
		return
	}
	apiresponse.OK(c, checkoutJSON(session, items))
}

func (s *Service) HandleUpdateCheckoutAddress(c *gin.Context) {
	uid, _ := currentUserID(c)
	sessionID, ok := uuidParam(c, "checkout_session_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_checkout_session_id", "Invalid checkout session id.")
		return
	}
	var req struct {
		ShippingAddress map[string]any `json:"shipping_address"`
		BillingAddress  map[string]any `json:"billing_address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid checkout address request.")
		return
	}
	session, err := s.queries.UpdateCheckoutAddress(c.Request.Context(), orderstore.UpdateCheckoutAddressParams{ID: sessionID, UserID: pgxutil.UUID(uid), ShippingAddress: jsonBytes(req.ShippingAddress), BillingAddress: jsonBytes(req.BillingAddress)})
	if err != nil {
		notFoundOrInternal(c, err, "checkout_not_found", "Checkout session not found.")
		return
	}
	items, _ := s.queries.ListCheckoutSessionItems(c.Request.Context(), session.ID)
	apiresponse.OK(c, checkoutJSON(session, items))
}

func (s *Service) HandleUpdateCheckoutShipping(c *gin.Context) {
	uid, _ := currentUserID(c)
	sessionID, ok := uuidParam(c, "checkout_session_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_checkout_session_id", "Invalid checkout session id.")
		return
	}
	var req struct {
		ShippingCents int64          `json:"shipping_cents"`
		Option        map[string]any `json:"option"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ShippingCents < 0 {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid shipping request.")
		return
	}
	session, err := s.queries.UpdateCheckoutShipping(c.Request.Context(), orderstore.UpdateCheckoutShippingParams{ID: sessionID, UserID: pgxutil.UUID(uid), ShippingCents: req.ShippingCents, SelectedShippingOption: jsonObject(req.Option)})
	if err != nil {
		notFoundOrInternal(c, err, "checkout_not_found", "Checkout session not found.")
		return
	}
	items, _ := s.queries.ListCheckoutSessionItems(c.Request.Context(), session.ID)
	apiresponse.OK(c, checkoutJSON(session, items))
}

func (s *Service) HandleUpdateCheckoutPaymentMethod(c *gin.Context) {
	uid, _ := currentUserID(c)
	sessionID, ok := uuidParam(c, "checkout_session_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_checkout_session_id", "Invalid checkout session id.")
		return
	}
	var req struct {
		PaymentMethod string `json:"payment_method" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid payment method request.")
		return
	}
	session, err := s.queries.UpdateCheckoutPaymentMethod(c.Request.Context(), orderstore.UpdateCheckoutPaymentMethodParams{ID: sessionID, UserID: pgxutil.UUID(uid), PaymentMethod: text(req.PaymentMethod)})
	if err != nil {
		notFoundOrInternal(c, err, "checkout_not_found", "Checkout session not found.")
		return
	}
	items, _ := s.queries.ListCheckoutSessionItems(c.Request.Context(), session.ID)
	apiresponse.OK(c, checkoutJSON(session, items))
}

func (s *Service) HandleApplyCheckoutVoucher(c *gin.Context) {
	uid, _ := currentUserID(c)
	sessionID, ok := uuidParam(c, "checkout_session_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_checkout_session_id", "Invalid checkout session id.")
		return
	}
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid voucher request.")
		return
	}
	voucher, err := s.queries.GetActiveVoucherByCode(c.Request.Context(), req.Code)
	if err != nil {
		notFoundOrInternal(c, err, "voucher_not_found", "Voucher not found.")
		return
	}
	session, err := s.queries.GetCheckoutSessionForUser(c.Request.Context(), orderstore.GetCheckoutSessionForUserParams{ID: sessionID, UserID: pgxutil.UUID(uid)})
	if err != nil {
		notFoundOrInternal(c, err, "checkout_not_found", "Checkout session not found.")
		return
	}
	items, err := s.queries.ListCheckoutSessionItems(c.Request.Context(), session.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load checkout items.")
		return
	}
	if !voucherScopeMatchesCheckout(voucher, items) {
		apiresponse.Error(c, http.StatusBadRequest, "voucher_scope_invalid", "Voucher is not applicable to this checkout.")
		return
	}
	discount := calculateDiscount(voucher, session.SubtotalCents, session.ShippingCents)
	session, err = s.queries.ApplyCheckoutVoucher(c.Request.Context(), orderstore.ApplyCheckoutVoucherParams{ID: sessionID, UserID: pgxutil.UUID(uid), VoucherID: voucher.ID, DiscountCents: discount})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "voucher_apply_failed", "Could not apply voucher.")
		return
	}
	apiresponse.OK(c, checkoutJSON(session, items))
}

func (s *Service) HandleRemoveCheckoutVoucher(c *gin.Context) {
	uid, _ := currentUserID(c)
	sessionID, ok := uuidParam(c, "checkout_session_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_checkout_session_id", "Invalid checkout session id.")
		return
	}
	voucherID, ok := uuidParam(c, "voucher_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_voucher_id", "Invalid voucher id.")
		return
	}
	session, err := s.queries.RemoveCheckoutVoucher(c.Request.Context(), orderstore.RemoveCheckoutVoucherParams{ID: sessionID, UserID: pgxutil.UUID(uid), VoucherID: voucherID})
	if err != nil {
		notFoundOrInternal(c, err, "voucher_not_found", "Voucher not found on checkout.")
		return
	}
	items, _ := s.queries.ListCheckoutSessionItems(c.Request.Context(), session.ID)
	apiresponse.OK(c, checkoutJSON(session, items))
}

func (s *Service) HandleValidateCheckout(c *gin.Context) {
	uid, _ := currentUserID(c)
	sessionID, ok := uuidParam(c, "checkout_session_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_checkout_session_id", "Invalid checkout session id.")
		return
	}
	session, err := s.queries.GetCheckoutSessionForUser(c.Request.Context(), orderstore.GetCheckoutSessionForUserParams{ID: sessionID, UserID: pgxutil.UUID(uid)})
	if err != nil {
		notFoundOrInternal(c, err, "checkout_not_found", "Checkout session not found.")
		return
	}
	items, err := s.queries.ListCheckoutSessionItems(c.Request.Context(), session.ID)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not validate checkout.")
		return
	}
	apiresponse.OK(c, gin.H{"valid": session.Status == "open" && session.ExpiresAt.Time.After(time.Now()) && len(items) > 0, "checkout": checkoutJSON(session, items)})
}

func (s *Service) HandlePlaceOrder(c *gin.Context) {
	uid, _ := currentUserID(c)
	sessionID, ok := uuidParam(c, "checkout_session_id")
	if !ok {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_checkout_session_id", "Invalid checkout session id.")
		return
	}
	order, items, shipments, err := s.placeOrderTx(c.Request.Context(), pgxutil.UUID(uid), sessionID)
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "place_order_failed", err.Error())
		return
	}
	s.enqueueOrderLifecycleJobs(c.Request.Context(), order)
	apiresponse.Respond(c, http.StatusCreated, orderDetailJSON(order, items, shipments), nil)
}

func (s *Service) enqueueOrderLifecycleJobs(ctx context.Context, order orderstore.Order) {
	if s.queue == nil {
		return
	}
	orderID := uuidString(order.ID)
	_ = s.queue.PublishJob(ctx, queue.JobPaymentConfirmationTimeout, "payment_timeout:"+orderID, gin.H{"order_id": orderID})
	_ = s.queue.PublishJob(ctx, queue.JobCancelUnpaidOrders, "cancel_unpaid:"+orderID, gin.H{"order_id": orderID})
	_ = s.queue.PublishJob(ctx, queue.JobReleaseExpiredInventory, "release_inventory:"+orderID, gin.H{"order_id": orderID})
}

func (s *Service) placeOrderTx(ctx context.Context, userID, sessionID pgtype.UUID) (orderstore.Order, []orderstore.OrderItem, []orderstore.Shipment, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return orderstore.Order{}, nil, nil, err
	}
	defer tx.Rollback(ctx)
	q := s.queries.WithTx(tx)
	session, err := q.LockCheckoutSessionForUser(ctx, orderstore.LockCheckoutSessionForUserParams{ID: sessionID, UserID: userID})
	if err != nil {
		return orderstore.Order{}, nil, nil, err
	}
	if session.Status != "open" || session.ExpiresAt.Time.Before(time.Now()) {
		return orderstore.Order{}, nil, nil, errors.New("checkout session is not placeable")
	}
	if len(session.ShippingAddress) == 0 {
		return orderstore.Order{}, nil, nil, errors.New("shipping address is required")
	}
	checkoutItems, err := q.ListCheckoutSessionItems(ctx, session.ID)
	if err != nil || len(checkoutItems) == 0 {
		return orderstore.Order{}, nil, nil, errors.New("checkout has no items")
	}
	order, err := q.CreateOrder(ctx, orderstore.CreateOrderParams{OrderNumber: orderNumber(), UserID: userID, CheckoutSessionID: session.ID, Status: "pending_payment", SubtotalCents: session.SubtotalCents, DiscountCents: session.DiscountCents, ShippingCents: session.ShippingCents, TaxCents: session.TaxCents, TotalCents: session.TotalCents, Currency: text(session.Currency), ShippingAddress: session.ShippingAddress, BillingAddress: session.BillingAddress})
	if err != nil {
		return orderstore.Order{}, nil, nil, err
	}
	orderItems := make([]orderstore.OrderItem, 0, len(checkoutItems))
	for _, checkoutItem := range checkoutItems {
		item, err := q.CreateOrderItemFromCheckoutItem(ctx, orderstore.CreateOrderItemFromCheckoutItemParams{OrderID: order.ID, CheckoutItemID: checkoutItem.ID})
		if err != nil {
			return orderstore.Order{}, nil, nil, err
		}
		orderItems = append(orderItems, item)
	}
	reservations, err := q.ListReservationsForCheckout(ctx, session.ID)
	if err != nil {
		return orderstore.Order{}, nil, nil, err
	}
	for _, res := range reservations {
		if _, err := q.AttachReservationToOrder(ctx, orderstore.AttachReservationToOrderParams{ID: res.ID, OrderID: order.ID}); err != nil {
			return orderstore.Order{}, nil, nil, err
		}
	}
	if session.VoucherID.Valid && session.DiscountCents > 0 {
		if err := q.IncrementVoucherUsage(ctx, session.VoucherID); err != nil {
			return orderstore.Order{}, nil, nil, err
		}
		if _, err := q.CreateVoucherRedemption(ctx, orderstore.CreateVoucherRedemptionParams{VoucherID: session.VoucherID, UserID: userID, OrderID: order.ID, DiscountCents: session.DiscountCents}); err != nil {
			return orderstore.Order{}, nil, nil, err
		}
	}
	if _, err := q.CompleteCheckoutSession(ctx, orderstore.CompleteCheckoutSessionParams{ID: session.ID, UserID: userID}); err != nil {
		return orderstore.Order{}, nil, nil, err
	}
	shipments, err := q.CreateShipmentsForOrderSellers(ctx, order.ID)
	if err != nil {
		return orderstore.Order{}, nil, nil, err
	}
	if _, err := q.InsertOrderStatusHistory(ctx, orderstore.InsertOrderStatusHistoryParams{OrderID: order.ID, NewStatus: "pending_payment", Note: text("Order placed"), CreatedBy: userID}); err != nil {
		return orderstore.Order{}, nil, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return orderstore.Order{}, nil, nil, err
	}
	return order, orderItems, shipments, nil
}

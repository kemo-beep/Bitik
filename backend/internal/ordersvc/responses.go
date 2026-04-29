package ordersvc

import (
	orderstore "github.com/bitik/backend/internal/store/orders"
	"github.com/gin-gonic/gin"
)

func addressJSON(a orderstore.UserAddress) gin.H {
	return gin.H{"id": uuidString(a.ID), "user_id": uuidString(a.UserID), "full_name": a.FullName, "phone": a.Phone, "country": a.Country, "state": textValue(a.State), "city": textValue(a.City), "district": textValue(a.District), "postal_code": textValue(a.PostalCode), "address_line1": a.AddressLine1, "address_line2": textValue(a.AddressLine2), "is_default": a.IsDefault, "created_at": a.CreatedAt.Time, "updated_at": a.UpdatedAt.Time}
}

func cartJSON(cart orderstore.Cart, items []orderstore.ListCartItemsDetailedRow) gin.H {
	subtotal := int64(0)
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		if item.Selected {
			subtotal += int64(item.Quantity) * item.PriceCents
		}
		out = append(out, cartItemDetailedJSON(item))
	}
	return gin.H{"id": uuidString(cart.ID), "user_id": uuidString(cart.UserID), "session_id": textValue(cart.SessionID), "voucher_id": nullableUUID(cart.VoucherID), "subtotal_cents": subtotal, "items": out, "created_at": cart.CreatedAt.Time, "updated_at": cart.UpdatedAt.Time}
}

func cartItemJSON(item orderstore.CartItem) gin.H {
	return gin.H{"id": uuidString(item.ID), "cart_id": uuidString(item.CartID), "product_id": uuidString(item.ProductID), "variant_id": uuidString(item.VariantID), "quantity": item.Quantity, "selected": item.Selected, "created_at": item.CreatedAt.Time, "updated_at": item.UpdatedAt.Time}
}

func cartItemDetailedJSON(item orderstore.ListCartItemsDetailedRow) gin.H {
	return gin.H{"id": uuidString(item.ID), "cart_id": uuidString(item.CartID), "product_id": uuidString(item.ProductID), "variant_id": uuidString(item.VariantID), "seller_id": uuidString(item.SellerID), "product_name": item.ProductName, "product_slug": item.ProductSlug, "sku": item.Sku, "variant_name": textValue(item.VariantName), "image_url": item.ImageUrl, "quantity": item.Quantity, "selected": item.Selected, "unit_price_cents": item.PriceCents, "total_price_cents": item.PriceCents * int64(item.Quantity), "currency": item.Currency, "purchasable_quantity": item.PurchasableQuantity}
}

func checkoutJSON(session orderstore.CheckoutSession, items []orderstore.CheckoutSessionItem) gin.H {
	return gin.H{"id": uuidString(session.ID), "user_id": uuidString(session.UserID), "cart_id": nullableUUID(session.CartID), "status": session.Status, "currency": session.Currency, "subtotal_cents": session.SubtotalCents, "discount_cents": session.DiscountCents, "shipping_cents": session.ShippingCents, "tax_cents": session.TaxCents, "total_cents": session.TotalCents, "shipping_address": rawJSON(session.ShippingAddress), "billing_address": rawJSON(session.BillingAddress), "payment_method": textValue(session.PaymentMethod), "selected_shipping_option": rawJSON(session.SelectedShippingOption), "voucher_id": nullableUUID(session.VoucherID), "expires_at": session.ExpiresAt.Time, "completed_at": session.CompletedAt.Time, "items": mapSlice(items, checkoutItemJSON), "created_at": session.CreatedAt.Time, "updated_at": session.UpdatedAt.Time}
}

func checkoutItemJSON(item orderstore.CheckoutSessionItem) gin.H {
	return gin.H{"id": uuidString(item.ID), "checkout_session_id": uuidString(item.CheckoutSessionID), "seller_id": uuidString(item.SellerID), "product_id": uuidString(item.ProductID), "variant_id": uuidString(item.VariantID), "quantity": item.Quantity, "unit_price_cents": item.UnitPriceCents, "total_price_cents": item.TotalPriceCents, "currency": item.Currency, "metadata": rawJSON(item.Metadata), "created_at": item.CreatedAt.Time}
}

func orderJSON(order orderstore.Order) gin.H {
	return gin.H{"id": uuidString(order.ID), "order_number": order.OrderNumber, "user_id": uuidString(order.UserID), "checkout_session_id": nullableUUID(order.CheckoutSessionID), "status": statusString(order.Status), "subtotal_cents": order.SubtotalCents, "discount_cents": order.DiscountCents, "shipping_cents": order.ShippingCents, "tax_cents": order.TaxCents, "total_cents": order.TotalCents, "currency": order.Currency, "shipping_address": rawJSON(order.ShippingAddress), "billing_address": rawJSON(order.BillingAddress), "placed_at": order.PlacedAt.Time, "paid_at": order.PaidAt.Time, "cancelled_at": order.CancelledAt.Time, "completed_at": order.CompletedAt.Time, "created_at": order.CreatedAt.Time, "updated_at": order.UpdatedAt.Time}
}

func orderItemJSON(item orderstore.OrderItem) gin.H {
	return gin.H{"id": uuidString(item.ID), "order_id": uuidString(item.OrderID), "seller_id": uuidString(item.SellerID), "product_id": uuidString(item.ProductID), "variant_id": uuidString(item.VariantID), "product_name": item.ProductName, "variant_name": textValue(item.VariantName), "sku": textValue(item.Sku), "image_url": textValue(item.ImageUrl), "quantity": item.Quantity, "unit_price_cents": item.UnitPriceCents, "total_price_cents": item.TotalPriceCents, "currency": item.Currency, "status": statusString(item.Status), "created_at": item.CreatedAt.Time}
}

func orderDetailJSON(order orderstore.Order, items []orderstore.OrderItem, shipments []orderstore.Shipment) gin.H {
	payload := orderJSON(order)
	payload["items"] = mapSlice(items, orderItemJSON)
	payload["shipments"] = mapSlice(shipments, shipmentJSON)
	return payload
}

func historyJSON(h orderstore.OrderStatusHistory) gin.H {
	return gin.H{"id": uuidString(h.ID), "order_id": uuidString(h.OrderID), "old_status": statusString(h.OldStatus), "new_status": statusString(h.NewStatus), "note": textValue(h.Note), "created_by": nullableUUID(h.CreatedBy), "created_at": h.CreatedAt.Time}
}

func shipmentJSON(s orderstore.Shipment) gin.H {
	return gin.H{"id": uuidString(s.ID), "order_id": uuidString(s.OrderID), "seller_id": uuidString(s.SellerID), "provider_id": nullableUUID(s.ProviderID), "tracking_number": textValue(s.TrackingNumber), "label_url": textValue(s.LabelUrl), "provider_metadata": rawJSON(s.ProviderMetadata), "status": statusString(s.Status), "shipped_at": s.ShippedAt.Time, "delivered_at": s.DeliveredAt.Time, "created_at": s.CreatedAt.Time, "updated_at": s.UpdatedAt.Time}
}

func trackingEventJSON(e orderstore.ShipmentTrackingEvent) gin.H {
	return gin.H{"id": uuidString(e.ID), "shipment_id": uuidString(e.ShipmentID), "status": e.Status, "location": textValue(e.Location), "message": textValue(e.Message), "event_time": e.EventTime.Time, "created_at": e.CreatedAt.Time}
}

func refundJSON(r orderstore.Refund) gin.H {
	return gin.H{"id": uuidString(r.ID), "order_id": uuidString(r.OrderID), "payment_id": nullableUUID(r.PaymentID), "requested_by": nullableUUID(r.RequestedBy), "status": statusString(r.Status), "reason": textValue(r.Reason), "amount_cents": r.AmountCents, "currency": r.Currency, "metadata": rawJSON(r.Metadata), "created_at": r.CreatedAt.Time, "updated_at": r.UpdatedAt.Time}
}

func returnJSON(r orderstore.ReturnRequest) gin.H {
	return gin.H{"id": uuidString(r.ID), "order_id": uuidString(r.OrderID), "order_item_id": nullableUUID(r.OrderItemID), "requested_by": nullableUUID(r.RequestedBy), "status": statusString(r.Status), "reason": textValue(r.Reason), "quantity": r.Quantity, "metadata": rawJSON(r.Metadata), "created_at": r.CreatedAt.Time, "updated_at": r.UpdatedAt.Time}
}

func invoiceJSON(i orderstore.OrderInvoice) gin.H {
	return gin.H{"id": uuidString(i.ID), "order_id": uuidString(i.OrderID), "invoice_number": i.InvoiceNumber, "status": i.Status, "payload": rawJSON(i.Payload), "generated_at": i.GeneratedAt.Time, "created_at": i.CreatedAt.Time}
}

func paymentJSON(p orderstore.Payment) gin.H {
	return gin.H{"id": uuidString(p.ID), "order_id": uuidString(p.OrderID), "payment_method_id": nullableUUID(p.PaymentMethodID), "provider": p.Provider, "provider_payment_id": textValue(p.ProviderPaymentID), "status": statusString(p.Status), "amount_cents": p.AmountCents, "currency": p.Currency, "metadata": rawJSON(p.Metadata), "paid_at": p.PaidAt.Time, "failed_at": p.FailedAt.Time, "created_at": p.CreatedAt.Time, "updated_at": p.UpdatedAt.Time}
}

func mapSlice[T any](items []T, mapper func(T) gin.H) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, mapper(item))
	}
	return out
}

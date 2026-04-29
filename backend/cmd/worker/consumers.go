package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/bitik/backend/internal/platform/queue"
	workerstore "github.com/bitik/backend/internal/store/worker"
	"github.com/gin-gonic/gin"
)

func (w *WorkerApp) registerConsumers(ctx context.Context) error {
	guard := queue.NewGuard(workerstore.New(w.pool))
	consumer := queue.NewConsumer(w.broker).WithGuard(guard)
	handlers := map[queue.JobType]queue.HandlerFunc{
		queue.JobExpireCheckout:          w.handleExpireCheckout,
		queue.JobCancelUnpaidOrders:      w.handleCancelUnpaidOrders,
		queue.JobReleaseExpiredInventory: w.handleReleaseExpiredInventory,
		queue.JobGenerateInvoice:         w.handleGenerateInvoice,
		queue.JobUpdateShipmentTracking:  w.handleUpdateShipmentTracking,
		queue.JobSettleSellerWallets:     w.handleSettleSellerWallets,
		queue.JobPaymentConfirmationTimeout: func(ctx context.Context, evt queue.Envelope) error {
			return w.handleExpirePendingPayments(ctx, evt)
		},
		queue.JobWaveStaleOrderTimeout: func(ctx context.Context, evt queue.Envelope) error {
			return w.handleCancelUnpaidOrders(ctx, evt)
		},
		queue.JobIndexProduct:        w.handleIndexProduct,
		queue.JobReindexProductsFull: w.handleReindexProductsFull,
		queue.JobSendEmail:           w.handleSendEmail,
		queue.JobSendSMSOTP:          w.handleSendSMSOTP,
		queue.JobSendPush:            w.handleSendPush,
		queue.JobProcessImage:        w.handleProcessImage,
		queue.JobProcessPayouts:      w.handleProcessPayouts,
		queue.JobGenerateReports:     w.handleGenerateReports,
		queue.JobNotificationFanout:  w.handleNotificationFanout,
	}
	for jobType, h := range handlers {
		if err := consumer.Consume(ctx, jobType, h); err != nil {
			return fmt.Errorf("consume %s: %w", jobType, err)
		}
	}
	return nil
}

func (w *WorkerApp) handleExpireCheckout(_ context.Context, evt queue.Envelope) error {
	return invokeJSONHandler("/api/v1/internal/jobs/expire-checkout", evt.Payload, w.orders.HandleExpireCheckout)
}

func (w *WorkerApp) handleCancelUnpaidOrders(_ context.Context, evt queue.Envelope) error {
	return invokeJSONHandler("/api/v1/internal/jobs/cancel-unpaid-orders", evt.Payload, w.orders.HandleCancelUnpaidOrders)
}

func (w *WorkerApp) handleReleaseExpiredInventory(_ context.Context, evt queue.Envelope) error {
	return invokeJSONHandler("/api/v1/internal/jobs/release-expired-inventory", evt.Payload, w.orders.HandleReleaseExpiredInventory)
}

func (w *WorkerApp) handleGenerateInvoice(_ context.Context, evt queue.Envelope) error {
	return invokeJSONHandler("/api/v1/internal/jobs/generate-invoices", evt.Payload, w.orders.HandleGenerateInvoices)
}

func (w *WorkerApp) handleUpdateShipmentTracking(_ context.Context, evt queue.Envelope) error {
	return invokeJSONHandler("/api/v1/internal/jobs/update-shipment-tracking", evt.Payload, w.shipping.HandleUpdateShipmentTracking)
}

func (w *WorkerApp) handleSettleSellerWallets(_ context.Context, evt queue.Envelope) error {
	return invokeJSONHandler("/api/v1/internal/jobs/settle-seller-wallets", evt.Payload, w.payments.HandleSettleSellerWallets)
}

func (w *WorkerApp) handleExpirePendingPayments(_ context.Context, evt queue.Envelope) error {
	return invokeJSONHandler("/api/v1/internal/jobs/expire-pending-payments", evt.Payload, w.payments.HandleExpirePendingPayments)
}

func (w *WorkerApp) handleIndexProduct(_ context.Context, evt queue.Envelope) error {
	return invokeJSONHandler("/api/v1/internal/jobs/index-product", evt.Payload, w.search.HandleIndexProduct)
}

func (w *WorkerApp) handleReindexProductsFull(_ context.Context, evt queue.Envelope) error {
	return invokeJSONHandler("/api/v1/internal/jobs/reindex-products", evt.Payload, w.search.HandleReindexProducts)
}

func (w *WorkerApp) handleSendEmail(ctx context.Context, evt queue.Envelope) error {
	return w.adapter.Send(ctx, "send_email", evt.Payload)
}

func (w *WorkerApp) handleSendSMSOTP(ctx context.Context, evt queue.Envelope) error {
	return w.adapter.Send(ctx, "send_sms_otp", evt.Payload)
}

func (w *WorkerApp) handleSendPush(ctx context.Context, evt queue.Envelope) error {
	return w.adapter.Send(ctx, "send_push", evt.Payload)
}

func (w *WorkerApp) handleProcessImage(ctx context.Context, evt queue.Envelope) error {
	return w.adapter.Send(ctx, "process_image", evt.Payload)
}

func (w *WorkerApp) handleProcessPayouts(ctx context.Context, evt queue.Envelope) error {
	return w.adapter.Send(ctx, "process_payouts", evt.Payload)
}

func (w *WorkerApp) handleGenerateReports(ctx context.Context, evt queue.Envelope) error {
	return w.adapter.Send(ctx, "generate_reports", evt.Payload)
}

func (w *WorkerApp) handleNotificationFanout(ctx context.Context, evt queue.Envelope) error {
	return w.adapter.Send(ctx, "notification_fanout", evt.Payload)
}

func invokeJSONHandler(path string, payload json.RawMessage, handler func(*gin.Context)) error {
	body := payload
	if len(body) == 0 {
		body = []byte("{}")
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", path, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	if rec.Code >= 300 {
		return fmt.Errorf("handler status %d body=%s", rec.Code, rec.Body.String())
	}
	return nil
}

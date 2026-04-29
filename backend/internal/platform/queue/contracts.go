package queue

import "time"

type JobType string

const (
	JobSendEmail                  JobType = "worker.send_email"
	JobSendSMSOTP                 JobType = "worker.send_sms_otp"
	JobSendPush                   JobType = "worker.send_push"
	JobPaymentConfirmationTimeout JobType = "worker.payment_confirmation_timeout"
	JobWaveStaleOrderTimeout      JobType = "worker.wave_stale_order_timeout"
	JobGenerateInvoice            JobType = "worker.generate_invoice"
	JobProcessImage               JobType = "worker.process_image"
	JobIndexProduct               JobType = "worker.index_product"
	JobReindexProductsFull        JobType = "worker.reindex_products_full"
	JobExpireCheckout             JobType = "worker.expire_checkout"
	JobCancelUnpaidOrders         JobType = "worker.cancel_unpaid_orders"
	JobReleaseExpiredInventory    JobType = "worker.release_expired_inventory"
	JobUpdateShipmentTracking     JobType = "worker.update_shipment_tracking"
	JobSettleSellerWallets        JobType = "worker.settle_seller_wallets"
	JobProcessPayouts             JobType = "worker.process_payouts"
	JobGenerateReports            JobType = "worker.generate_reports"
	JobNotificationFanout         JobType = "worker.notification_fanout"
)

const (
	ExchangeJobs      = "bitik.jobs"
	ExchangeJobsRetry = "bitik.jobs.retry"
	ExchangeJobsDLQ   = "bitik.jobs.dlq"
	MaxAttempts       = 8
)

type JobDefinition struct {
	QueueName      string
	RetryQueueName string
	DLQName        string
	RoutingKey     string
	RetryBackoff   []time.Duration
	Prefetch       int
	Concurrency    int
}

func jobDefinition(queueName, routingKey string) JobDefinition {
	return JobDefinition{
		QueueName:      queueName,
		RetryQueueName: queueName + ".retry",
		DLQName:        queueName + ".dlq",
		RoutingKey:     routingKey,
		RetryBackoff: []time.Duration{
			30 * time.Second,
			2 * time.Minute,
			10 * time.Minute,
			30 * time.Minute,
			1 * time.Hour,
			3 * time.Hour,
			6 * time.Hour,
			12 * time.Hour,
		},
		Prefetch:    20,
		Concurrency: 4,
	}
}

var JobDefinitions = map[JobType]JobDefinition{
	JobSendEmail:                  jobDefinition("jobs.send_email.v1", "notify.email.send"),
	JobSendSMSOTP:                 jobDefinition("jobs.send_sms_otp.v1", "notify.sms.otp.send"),
	JobSendPush:                   jobDefinition("jobs.send_push.v1", "notify.push.send"),
	JobPaymentConfirmationTimeout: jobDefinition("jobs.payment_confirmation_timeout.v1", "payments.timeout.confirmation"),
	JobWaveStaleOrderTimeout:      jobDefinition("jobs.wave_stale_order_timeout.v1", "payments.wave.stale_order_timeout"),
	JobGenerateInvoice:            jobDefinition("jobs.generate_invoice.v1", "orders.invoice.generate"),
	JobProcessImage:               jobDefinition("jobs.process_image.v1", "media.image.process"),
	JobIndexProduct:               jobDefinition("jobs.index_product.v1", "search.product.index"),
	JobReindexProductsFull:        jobDefinition("jobs.reindex_products_full.v1", "search.product.reindex_full"),
	JobExpireCheckout:             jobDefinition("jobs.expire_checkout.v1", "orders.checkout.expire"),
	JobCancelUnpaidOrders:         jobDefinition("jobs.cancel_unpaid_orders.v1", "orders.cancel_unpaid"),
	JobReleaseExpiredInventory:    jobDefinition("jobs.release_expired_inventory.v1", "orders.inventory.release_expired"),
	JobUpdateShipmentTracking:     jobDefinition("jobs.update_shipment_tracking.v1", "shipping.tracking.refresh"),
	JobSettleSellerWallets:        jobDefinition("jobs.settle_seller_wallets.v1", "wallets.settle_seller"),
	JobProcessPayouts:             jobDefinition("jobs.process_payouts.v1", "wallets.payout.process"),
	JobGenerateReports:            jobDefinition("jobs.generate_reports.v1", "reports.generate"),
	JobNotificationFanout:         jobDefinition("jobs.notification_fanout.v1", "notify.fanout"),
}

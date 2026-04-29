package queue

import "testing"

func TestJobDefinitionsContainRequiredContracts(t *testing.T) {
	required := []JobType{
		JobSendEmail,
		JobSendSMSOTP,
		JobSendPush,
		JobPaymentConfirmationTimeout,
		JobWaveStaleOrderTimeout,
		JobGenerateInvoice,
		JobProcessImage,
		JobIndexProduct,
		JobReindexProductsFull,
		JobExpireCheckout,
		JobCancelUnpaidOrders,
		JobReleaseExpiredInventory,
		JobUpdateShipmentTracking,
		JobSettleSellerWallets,
		JobProcessPayouts,
		JobGenerateReports,
		JobNotificationFanout,
	}
	for _, jt := range required {
		def, ok := JobDefinitions[jt]
		if !ok {
			t.Fatalf("missing definition for %s", jt)
		}
		if def.QueueName == "" || def.RoutingKey == "" {
			t.Fatalf("invalid definition for %s", jt)
		}
	}
}

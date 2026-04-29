package shippingsvc

import (
	"net/http"
	"time"

	"github.com/bitik/backend/internal/apiresponse"
	shippingstore "github.com/bitik/backend/internal/store/shipping"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) HandleUpdateShipmentTracking(c *gin.Context) {
	// v1 worker: ensure delivered shipments have a delivered tracking event.
	ctx := c.Request.Context()
	limit := int32(200)
	shipments, err := s.shipQ.ListDeliveredShipmentsWithoutTrackingEvents(ctx, limit)
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not scan shipments.")
		return
	}
	created := 0
	for _, sh := range shipments {
		ts := time.Now().UTC()
		if sh.DeliveredAt.Valid {
			ts = sh.DeliveredAt.Time.UTC()
		}
		_, err := s.shipQ.CreateShipmentTrackingEvent(ctx, shippingstore.CreateShipmentTrackingEventParams{
			ShipmentID: sh.ID,
			Status:     "delivered",
			EventTime:  pgtype.Timestamptz{Time: ts, Valid: true},
		})
		if err == nil {
			created++
		}
	}
	apiresponse.OK(c, gin.H{"scanned": len(shipments), "created": created})
}

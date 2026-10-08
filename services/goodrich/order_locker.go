package goodrich

import (
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	_ "github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/WilliamsStudentsOnline/wso-go/services/goodrich/lease"
	"github.com/gin-gonic/gin"
)

type OrderLease struct {
	ID     string    `json:"id"`
	Expiry time.Time `json:"expiry"`
}

// GetOrderLease godoc
// @Summary Get order lease
// @Description gets a lease to be able to order
// @ID getOrderLease
// @Tags goodrich
// @Accept  json
// @Produce  json
// @Success 200 {object} goodrich.OrderLease
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /goodrich/order-lease [get]
func (t *Controller) GetOrderLease(c *gin.Context) {
	date := time.Now().Format(DateFormat)

	// Ensure goodrich is open today
	foundValidDate := false
	for _, openDay := range t.cfg.GoodrichOpenDays {
		if date == openDay {
			foundValidDate = true
			break
		}
	}
	if !foundValidDate {
		t.RespondAPIError(c, lib.ErrorGoodrichDateClosed)
		return
	}

	// Get slots left. if none, reject
	tsOpenCount, err := t.countOpenTimeSlots()
	if err != nil {
		t.RespondError(c, err)
		return
	}

	if tsOpenCount <= 0 {
		t.RespondAPIError(c, lib.ErrorGoodrichNoTimesAvailable)
		return
	}

	// Get slots left. If there are more leases than slots left, do not give a lease.
	// Instead, error in the same way as not having a lease and have the client either
	// wait for a lease to open up without taking a slot, or fail when there are no
	// slots left.
	if tsOpenCount <= t.orderLessor.CountUsedLeases() {
		t.RespondAPIError(c, lib.ErrorGoodrichNoLeasesAvailable)
		return
	}

	// Request a lease. If we get one, send it!
	ol, err := t.orderLessor.RequestLease()
	if err != nil {
		if err == lease.ErrLessorTooFull {
			t.RespondAPIError(c, lib.ErrorGoodrichNoLeasesAvailable)
		} else {
			t.RespondError(c, err)
		}
		return
	}

	t.RespondOK(c, OrderLease{
		ID:     ol.ID.String(),
		Expiry: ol.Expiry,
	})
}

func (t *Controller) countOpenTimeSlots() (int, error) {
	date := time.Now().Format(DateFormat)

	// Get timeslots that exist today
	validTimeSlots := t.generateTimeSlotsAfter(time.Now())

	// Get timeslot availability
	tsAvailabilityMap, err := t.generateSlotAvailabilityMap(date)
	if err != nil {
		return 0, err
	}

	openTimeSlots := 0

	for _, slot := range validTimeSlots {
		closedSpots := tsAvailabilityMap[slot.String()]
		if t.cfg.GoodrichSlotSpotSize-closedSpots > 0 {
			openTimeSlots += t.cfg.GoodrichSlotSpotSize - closedSpots
		}
	}

	return openTimeSlots, nil
}

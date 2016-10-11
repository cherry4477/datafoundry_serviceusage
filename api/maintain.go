package api

import (
	"fmt"
	"time"

	"github.com/asiainfoLDP/datafoundry_serviceusage/openshift"
)

//======================================================
// 
//======================================================

func StartMaintaining() {
	Logger.Infof("Maintaining started ...")

	// todo:
	// find all consuming orders which deadline < now()+7_days.
}

//======================================================
// 
//======================================================

func renewOrder(accountId, orderId string, plan *Plan) error {
	// make a payment

	err := makePayment(openshift.AdminToken(), accountId, plan.Price)
	if err != nil {
		return err
	} 

	// todo: create a consuming report // need? maybe payment moodule has recoreded it.

	// change order status => consuming // will do in renew
	
	var extendedDuration time.Duration
	switch plan.Cycle {
	default:
		return fmt.Errorf("unknown plan cycle: %s", plan.Cycle)
	case PLanCircle_Month:
		_ = extendedDuration
	}

	// usage.RenewOrder(db *sql.DB, orderId string, extendedDuration time.Duration)

	return nil
}
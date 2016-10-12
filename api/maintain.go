package api

import (
	"fmt"
	"time"

	"github.com/asiainfoLDP/datafoundry_serviceusage/openshift"

	"github.com/asiainfoLDP/datafoundry_serviceusage/usage"
)

//======================================================
// 
//======================================================

func StartMaintaining() {
	Logger.Infof("Maintaining service started.")

	// todo:
	// find all consuming orders which deadline < now()+7_days.
}

//======================================================
// 
//======================================================

func renewOrder(accountId, orderId string, plan *Plan, reason string) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("db not inited")
	}

	// ...

	err := makePayment(openshift.AdminToken(), accountId, plan.Price, reason)
	if err != nil {
		err2 := usage.IncreaseOrderRenewalFails(db, orderId)
		if err2 != nil {
			Logger.Errorf("IncreaseOrderRenewalFails error: %s", err2.Error())
		}

		return err
	}

	now := time.Now()

	// todo: create a consuming report // need? maybe payment moodule has recoreded it.

	// change order status => consuming // will do in renew
	
	var extendedDuration time.Duration

	switch plan.Cycle {
	default:
		return fmt.Errorf("unknown plan cycle: %s", plan.Cycle)
	case PLanCircle_Month:
		extendedDuration = usage.DeadlineExtendedDuration_Month
	}

	order, err := usage.RenewOrder(db, orderId, extendedDuration)
	if err != nil {
		// todo: retry

		Logger.Errorf("RenewOrder error: %s", err.Error())
		return err
	}

	err = usage.CreateConsumeHistory(db, order, now, plan.Price)
	if err != nil {
		// todo: retry

		Logger.Errorf("CreateConsumeHistory error: %s", err.Error())
		return err
	}

	return nil
}
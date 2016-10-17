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

	timerRenewOrders := time.After(time.Minute)
	for {
		select {
		case <-timerRenewOrders:
			timerRenewOrders = TryToRenewConsumingOrders()
		}
	}
}

func TryToRenewConsumingOrders() <- chan time.Time {

	// todo:
	// find all consuming orders which deadline < now()+7_days.

	return time.After(time.Hour)
}

//======================================================
// 
//======================================================	

func OrderRenewReason(orderId string, renewTimes int) string {
	return fmt.Sprintf("order:%s:%d", orderId, renewTimes) // DON'T change
}

func renewOrder(accountId string, order *usage.PurchaseOrder, plan *Plan, lastConsume *usage.ConsumeHistory) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("db not inited")
	}

	// calculate the payment money

	var moeny float32 = 0.0

	if lastConsume == nil {
		moeny = plan.Price
	} else {

	}

	// ...

	renewReason := OrderRenewReason(order.Order_id, order.Last_consume_id + 1)

	err := makePayment(openshift.AdminToken(), accountId, moeny, renewReason)
	if err != nil {
		err2 := usage.IncreaseOrderRenewalFails(db, order.Order_id)
		if err2 != nil {
			Logger.Warningf("IncreaseOrderRenewalFails error: %s", err2.Error())
		}

		return err
	}

	//now := time.Now()

	// ...


	return nil
}

func renewOrder_old(accountId, orderId string, plan *Plan, renewReason string) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("db not inited")
	}

	// ...

	err := makePayment(openshift.AdminToken(), accountId, plan.Price, renewReason)
	if err != nil {
		err2 := usage.IncreaseOrderRenewalFails(db, orderId)
		if err2 != nil {
			Logger.Warningf("IncreaseOrderRenewalFails error: %s", err2.Error())
		}

		return err
	}

	now := time.Now()

	// change order status => consuming // will do in following renew calling
	
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

		Logger.Warningf("RenewOrder error: %s", err.Error())
		return err
	}

	err = usage.CreateConsumeHistory(db, order, now, plan.Price, plan.Id)
	if err != nil {
		// todo: retry

		Logger.Warningf("CreateConsumeHistory error: %s", err.Error())
		return err
	}

	return nil
}
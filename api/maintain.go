package api

import (
	"fmt"
	"time"
	"math"
	"database/sql"

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

	// ConsumeExtraInfo_RenewOrder

	// if order is expired, change quota to 0

	return time.After(time.Hour)
}

//======================================================
// 
//======================================================	

func OrderRenewReason(orderId string, renewTimes int) string {
	return fmt.Sprintf("order:%s:%d", orderId, renewTimes) // DON'T change
}

// the return bool means insufficient balance or not
func renewOrder(drytry bool, db *sql.DB, accountId string, order *usage.PurchaseOrder, plan *Plan, oldOrder *usage.PurchaseOrder) (float32, error, bool) {

	var err error
	var lastConsume *usage.ConsumeHistory
	if oldOrder != nil {

		// get last payment, so we can check how much money is remaining to switch order

		lastConsume, err = usage.RetrieveConsumeHistory(db, oldOrder.Id, oldOrder.Order_id, oldOrder.Last_consume_id)
		if err != nil {
			return 0.0, fmt.Errorf("Failed to switch plan: " + err.Error()), false
		}

		if lastConsume == nil {
			return 0.0, fmt.Errorf("Failed to switch plan: last payment not found"), false
		}
	}

	// if old order exists, end it.
	// its remaining money will apply on the new order.

	var consumExtraInfo int
	var paymentMoney float32


	if lastConsume == nil {

		paymentMoney = plan.Price
		consumExtraInfo = usage.ConsumeExtraInfo_NewOrder
	} else {

		var remaingMoney float32
		now := time.Now()

		if now.Before(lastConsume.Consume_time) { // impossible

			return 0.0, fmt.Errorf("last consume time is after now"), false

		} else if now.After(lastConsume.Deadline_time) {

			remaingMoney = 0.0

			// try to end last order 
			err := usage.EndOrder(db, oldOrder, now, lastConsume, 0.0)
			if err != nil {
				return 0.0, fmt.Errorf("end old order (%s) error: %s", lastConsume.Order_id, err.Error()), false
			}

			// create new 

			paymentMoney = plan.Price
			consumExtraInfo = usage.ConsumeExtraInfo_NewOrder

		} else {

			// by current design, plan.price must be larger than or equal 
			// the remaining charging of the last payment.
			const DayDuration = float64(time.Hour * 24)

			remainingDays := math.Floor(float64(lastConsume.Deadline_time.Sub(now)) / DayDuration)
			allDays := math.Floor(0.5 + float64(lastConsume.Deadline_time.Sub(lastConsume.Consume_time)) / DayDuration)
			ratio := float32(remainingDays) / float32(allDays)
			remaingMoney = ratio * lastConsume.Money
			remaingMoney = 0.01 * float32(math.Floor(float64(remaingMoney) * 100.0))

			if remaingMoney > plan.Price {

				// todo: now, withdraw is not supported
				return 0.0, fmt.Errorf("old order (%s) has too much remaining spending", lastConsume.Order_id), false
			}

			// ...

			paymentMoney = plan.Price - remaingMoney
			consumExtraInfo = usage.ConsumeExtraInfo_SwitchOrder
		}

		// try to end last order 
		err := usage.EndOrder(db, oldOrder, now, lastConsume, remaingMoney)
		if err != nil {
			return 0.0, fmt.Errorf("end old order (%s) error: %s", lastConsume.Order_id, err.Error()), false
		}
	}

	if drytry {
		return paymentMoney, nil, false
	}

	// ...
	
	var extendedDuration time.Duration

	switch plan.Cycle {
	default:
		return paymentMoney, fmt.Errorf("unknown plan cycle: %s", plan.Cycle), false
	case PLanCircle_Month:
		extendedDuration = usage.DeadlineExtendedDuration_Month
	}

	// ...

	if paymentMoney > 0.0 {
		paymentReason := OrderRenewReason(order.Order_id, order.Last_consume_id + 1)

		err, insufficientBalance := makePayment(openshift.AdminToken(), accountId, paymentMoney, paymentReason, order.Region)
		if err != nil && insufficientBalance {
			err2 := usage.IncreaseOrderRenewalFails(db, order.Id)
			if err2 != nil {
				Logger.Warningf("IncreaseOrderRenewalFails error: %s", err2.Error())
			}

			return 0.0, err, insufficientBalance
		}

		order.Last_consume_id = order.Last_consume_id + 1
	}

	// ...

	order, err = usage.RenewOrder(db, order.Id, extendedDuration)
	if err != nil {
		// todo: retry

		Logger.Warningf("RenewOrder error: %s", err.Error())
		return paymentMoney, err, false
	}

	now := time.Now()

	// err = usage.CreateConsumeHistory(db, order, now, paymentMoney, plan.Id, consumExtraInfo)
	err = usage.CreateConsumeHistory(db, order, now, plan.Price, plan.Id, consumExtraInfo)
	if err != nil {
		// todo: retry

		Logger.Warningf("CreateConsumeHistory error: %s", err.Error())
		return paymentMoney, err, false
	}

	// modify quota

	go func() {
		switch consumExtraInfo {
		case usage.ConsumeExtraInfo_NewOrder, usage.ConsumeExtraInfo_SwitchOrder:
			err := changeDfProjectQuota(order.Creator, accountId, plan)
			if err != nil {
				// todo: retry
				
				Logger.Warningf("changeDfProjectQuota (%s, %s, %s) error: %s", 
					order.Creator, accountId, plan.Plan_id, err.Error())
			}
		}
	}()

	// ...

	return paymentMoney, nil, false
}

/*
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
*/

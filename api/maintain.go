package api

import (
	"fmt"
	"time"
	"math"
	"database/sql"

	//"github.com/asiainfoLDP/datafoundry_serviceusage/openshift"

	"github.com/asiainfoLDP/datafoundry_serviceusage/usage"
)

//======================================================
// 
//======================================================

func StartMaintaining() {
	Logger.Infof("Maintaining service started.")

	timerRenewConsumingOrders := time.After(time.Minute)
	
	for {
		select {
		case <-timerRenewConsumingOrders:
			timerRenewConsumingOrders = TryToRenewConsumingOrders()
		}
	}
}

// find all consuming orders need to renew (orders which will expire in RenewMargin),
// 1. if ordre has bee expired for EndOrderMargin, 
//    > change cpu/memory quotas to zero.
//    > delete volume ...
//    > delete bsi ...
// 2. try to renew them, 
// 3. if balance is insufficient, send wanring emails.
func TryToRenewConsumingOrders() (tm <- chan time.Time) {
	dur := 5 * time.Minute // default duration to invoke this function again.
	defer func () {
		tm = time.After(dur)
	}()

	// ...

	db := getDB()
	if db == nil {
		Logger.Error("TryToRenewConsumingOrders error: db not inited")
		return
	}

	// ...

	const EndOrderMargin = 7 * 24 * time.Hour

	// return if the order is ended.
	onInsufficientBalance := func(order *usage.PurchaseOrder, plan *Plan) bool {
		if time.Now().Before(order.Deadline_time.Add(EndOrderMargin)) {

			t, err := usage.GetOrderLastWarningMessageTime(db, order.Id)
			if err != nil {
				Logger.Errorf("onInsufficientBalance GetOrderLastWarningMessageTime error: %s\n", err.Error())
				return false
			}

			now := time.Now()
			if now.Sub(t) < 24 * time.Hour {
				// send most one email per day
				return false
			}

			// todo: need a return error
			SendBalanceInsufficientEmail(order, plan)
			
			usage.SetOrderLastWarningMessageTime(db, order.Id, now)

			return false
		}

		err := endOrder(db, order)
		if err != nil {
			Logger.Error(err.Error())
				
			return false
		}

		SendEndOrderEmail_BalanceInsufficient(order, plan)

		return true
	}

	// ...

	const RenewMargin = 7 * 24 * time.Hour

	// todo: use cursor instead
	numAll, orders, err := usage.QueryConsumingOrdersToRenew(db, RenewMargin, 32)
	_ = numAll
	if err != nil {
		Logger.Warningf("TryToRenewConsumingOrders at %s error: %s", time.Now().Format("2006-01-02 15:04:05.999999"), err.Error())
	} else {
		Logger.Warningf("TryToRenewConsumingOrders started at %s", time.Now().Format("2006-01-02 15:04:05.999999"))

		for _, order := range orders {
			// 
			plan, err := getPlanByID(order.Plan_id, order.Region)
			if err != nil {
				Logger.Errorf("TryToRenewConsumingOrders getPlanByID (%s) error: %s", order.Plan_id, err.Error())
				continue
			}

			_, err, errReason := createOrder(false, db, nil, order, plan, nil)
			if err != nil {
				Logger.Errorf("TryToRenewConsumingOrders createOrder (%d) error: %s", order.Id, err.Error())

				if errReason == ErrorCodeInsufficentBalance {
					onInsufficientBalance(order, plan)
				}
				
				continue
			}

			Logger.Infof("TryToRenewConsumingOrders createOrder (%d) succeeded.", order.Id)

			SendRenewOrderEmail(order, plan)
		}
	}

	// ...

	// todo: it is hard to calculate a proper duration to call this function again.
	//       so used the default duration set at the beginning of this function now.

	//if int(numAll) > len(orders) {
	//	dur = 10 * time.Second
	//	Logger.Debugf("TryToRenewConsumingOrders len(orders) == maxCount")
	//} else {
	//	d, err := usage.GetDurationToRenewNextConsumingOrder(db, RenewMargin)
	//	if err != nil {
	//		Logger.Warningf("TryToRenewConsumingOrders GetDurationToRenewNextConsumingOrder error: %s", err.Error())
	//	} else {
	//		dur = d
	//	}
	//}

	return
}

//======================================================
// 
//======================================================	

func OrderRenewReason(orderId string, renewTimes int) string {
	return fmt.Sprintf("order:%s:%d", orderId, renewTimes) // DON'T change
}

// createParams == nil means for renew.
// the last return result means the exact error reason id. If its value <= 0, it will be ignored.
func createOrder(drytry bool, db *sql.DB, createParams *OrderCreationParams, order *usage.PurchaseOrder, plan *Plan, oldOrder *usage.PurchaseOrder) (float32, error, int) {

	var err error
	var lastConsume *usage.ConsumeHistory
	if oldOrder != nil {

		// get last payment, so we can check how much money is remaining to switch order

		lastConsume, err = usage.RetrieveConsumeHistory(db, oldOrder.Id, oldOrder.Order_id, oldOrder.Last_consume_id)
		if err != nil {
			return 0.0, fmt.Errorf("Failed to switch plan: " + err.Error()), -1
		}

		if lastConsume == nil {
			return 0.0, fmt.Errorf("Failed to switch plan: last payment not found"), -1
		}
	}

	// if old order exists, end it.
	// its remaining money will apply on the new order.
	
	var remaingMoney float32

	var consumExtraInfo int
	var paymentMoney float32

	if lastConsume == nil {

		paymentMoney = plan.Price
		if createParams == nil {
			consumExtraInfo = usage.ConsumeExtraInfo_RenewOrder
		} else {
			consumExtraInfo = usage.ConsumeExtraInfo_NewOrder
		}
	} else {

		now := time.Now()

		if now.Before(lastConsume.Consume_time) { // impossible

			return 0.0, fmt.Errorf("last consume time is after now"), -1

		} else if now.After(lastConsume.Deadline_time) {

			remaingMoney = 0.0

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
				return 0.0, fmt.Errorf("old order (%s) has too much remaining spending", lastConsume.Order_id), -1
			}

			// ...

			paymentMoney = plan.Price - remaingMoney
			consumExtraInfo = usage.ConsumeExtraInfo_SwitchOrder
		}

		// try to end last order 
		// (now moved after payment)
		//if ! drytry {
		//	err := usage.EndOrder(db, oldOrder, now, /*lastConsume,*/ remaingMoney)
		//	if err != nil {
		//		return 0.0, fmt.Errorf("end old order (%s) error: %s", oldOrder.Order_id, err.Error()), -1
		//	}
		//}
	}

	if drytry {
		return paymentMoney, nil, -1
	}

	// ...
	
	var extendedDuration time.Duration

	switch plan.Cycle {
	default:
		return paymentMoney, fmt.Errorf("unknown plan cycle: %s", plan.Cycle), -1
	case PLanCircle_Month:
		extendedDuration = usage.DeadlineExtendedDuration_Month
	}

	// ...

	if paymentMoney > 0.0 {
		paymentReason := OrderRenewReason(order.Order_id, order.Last_consume_id + 1)

		//err, insufficientBalance := makePayment(openshift.AdminToken(), order.Region, accountId, paymentMoney, paymentReason)
		err, insufficientBalance := makePayment(order.Region, order.Account_id, paymentMoney, paymentReason)
		if err != nil {
			err2 := usage.IncreaseOrderRenewalFails(db, order.Id)
			if err2 != nil {
				Logger.Warningf("IncreaseOrderRenewalFails error: %s", err2.Error())
			}

			if insufficientBalance {
				return 0.0, err, ErrorCodeInsufficentBalance
			} else {
				return 0.0, err, -1
			}
		}

		// todo: looks this line is useless now. 
		// for usage.RenewOrder will do it. 
		//order.Last_consume_id = order.Last_consume_id + 1
	}

	// ...

	order, err = usage.RenewOrder(db, order.Id, order.Last_consume_id, extendedDuration,
					plan.Price, plan.Id, consumExtraInfo)
	if err != nil {
		// todo: retry

		Logger.Warningf("RenewOrder error: %s", err.Error())
		return paymentMoney, err, -1
	}

	// 
	if lastConsume != nil {
		err := usage.EndOrder(db, oldOrder, time.Now(), /*lastConsume,*/ remaingMoney)
		if err != nil {
			return 0.0, fmt.Errorf("end old order (%s) error: %s", oldOrder.Order_id, err.Error()), -1
		}
	}

	// CreateConsumeHistory has been merged into RenewOrder above
	// ... 
	//go func() {
	//
	//	// err = usage.CreateConsumeHistory(db, order, now, paymentMoney, plan.Id, consumExtraInfo)
	//	err = usage.CreateConsumeHistory(db, order, time.Now(), plan.Price, plan.Id, consumExtraInfo)
	//	if err != nil {
	//		// todo: retry
	//
	//		Logger.Warningf("CreateConsumeHistory error: %s", err.Error())
	//		//return paymentMoney, err, false
	//	}
	//}()

	// ... 

	//go 
	finalErr, specialErrReason := func() (error, int) {
		if createParams == nil { // 
			return nil, 0
		}

		switch plan.Plan_type {
		default:
			Logger.Warningf("createOrder, unknown plan type: %s", plan.Plan_type)
			
			return nil, 0
		case PLanType_Quotas:

			switch consumExtraInfo {
			case usage.ConsumeExtraInfo_NewOrder, usage.ConsumeExtraInfo_SwitchOrder:
				err := changeDfProjectQuotaWithPlan(order.Creator, order.Region, order.Account_id, plan)
				if err != nil {
					// todo: retry
					
					Logger.Warningf("changeDfProjectQuotaWithPlan (%s, %s, %s, %s) error: %s", 
						order.Creator, order.Region, order.Account_id, plan.Plan_id, err.Error())
					
					return err, ErrorCodeChargedButFailedToCreateResource
				}
			}

		case PLanType_Volume:

			err := createPersistentVolume(order.Creator, createParams.ResName, order.Region, order.Account_id, plan)
			if err != nil {
				// todo: retry
				
				Logger.Warningf("createPersistentVolume (%s, %s, %s, %s) error: %s", 
					order.Creator, order.Region, order.Account_id, plan.Plan_id, err.Error())
				
				return err, ErrorCodeChargedButFailedToCreateResource
			}

		case PLanType_BSI:

			err := createBSI(order.Creator, createParams.ResName, order.Region, order.Account_id, plan)
			if err != nil {
				// todo: retry
				
				Logger.Warningf("createBSI (%s, %s, %s, %s) error: %s", 
					order.Creator, order.Region, order.Account_id, plan.Plan_id, err.Error())

				return err, ErrorCodeChargedButFailedToCreateResource
			}

			Logger.Infof("createBSI (%s, %s, %s, %s) succeeded", 
					order.Creator, order.Region, order.Account_id, plan.Plan_id)
		}

		return nil, 0
	}()

	// ...

	return paymentMoney, finalErr, specialErrReason
}



func endOrder(db *sql.DB, order *usage.PurchaseOrder) error {

	// destroy ordered resources

	switch order.Plan_type {

	case PLanType_Quotas:
		// zero quotas
		err := changeDfProjectQuota(order.Creator, order.Region, order.Account_id, 0, 0)
		if err != nil {
			// todo: retry
			
			return fmt.Errorf("onInsufficientBalance changeDfProjectQuota (%s, %s, %s, 0, 0) error: %s", 
				order.Creator, order.Region, order.Account_id, err.Error())
		}

	case PLanType_Volume:
		// delete volume
		err := destroyPersistentVolume(order.Resource_name, order.Region, order.Account_id)
		if err != nil {
			// todo: retry
			
			return fmt.Errorf("onInsufficientBalance destroyPersistentVolume (%s, %s, %s) error: %s", 
				order.Resource_name, order.Region, order.Account_id, err.Error())
		}

	case PLanType_BSI:
		// destroy bsi
		err := destroyBSI(order.Resource_name, order.Region, order.Account_id)
		if err != nil {
			// todo: retry
			
			return fmt.Errorf("onInsufficientBalance destroyBSI (%s, %s, %s) error: %s", 
				order.Resource_name, order.Region, order.Account_id, err.Error())
		}
	}

	// end order

	//lastConsume, err := usage.RetrieveConsumeHistory(db, order.Id, order.Order_id, order.Last_consume_id)
	//if err != nil {
	//	Logger.Errorf("TryToRenewConsumingOrders onInsufficientBalance RetrieveConsumeHistory (%s) error: %s", order.Id, err.Error())
	//	return false
	//}
	//
	//if lastConsume == nil {
	//	Logger.Errorf("TryToRenewConsumingOrders onInsufficientBalance RetrieveConsumeHistory (%s): lastConsume == nil", order.Id)
	//	return false
	//}

	err := usage.EndOrder(db, order, time.Now(), /*lastConsume,*/ 0.0)
	if err != nil {
		return fmt.Errorf("onInsufficientBalance EndOrder (%s, %s, %s) error: %s", 
			order.Resource_name, order.Region, order.Account_id, err.Error())
	}

	return nil
}
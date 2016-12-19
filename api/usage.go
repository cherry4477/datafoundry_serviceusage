package api

import (
	"net/http"
	"time"
	"crypto/rand"
	"fmt"
	//"strings"
	mathrand "math/rand"
	//neturl "net/url"
	"encoding/base64"

	"github.com/julienschmidt/httprouter"

	"github.com/asiainfoLDP/datahub_commons/common"

	"github.com/asiainfoLDP/datafoundry_serviceusage/usage"
	"github.com/asiainfoLDP/datafoundry_serviceusage/openshift"
)

//==================================================================
//
//==================================================================

func init() {
	mathrand.Seed(time.Now().UnixNano())
}

func genUUID() string {
	bs := make([]byte, 16)
	_, err := rand.Read(bs)
	if err != nil {
		Logger.Warning("genUUID error: ", err.Error())

		//mathrand.Read(bs)
		n := time.Now().UnixNano()
		for i := uint(0); i < 8; i ++ {
			bs[i] = byte((n >> i) & 0xff)
		}

		n = mathrand.Int63()
		for i := uint(0); i < 8; i ++ {
			bs[i+8] = byte((n >> i) & 0xff)
		}
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", bs[0:4], bs[4:6], bs[6:8], bs[8:10], bs[10:])
}

func genOrderID(accountId, planType string) string {
	switch planType {
	case PLanType_Quotas:
		return fmt.Sprintf("%s-%s", accountId, planType)
		// most one order allowed for such plan types
	}

	bs := make([]byte, 12)
	_, err := rand.Read(bs)
	if err != nil {
		Logger.Warning("genOrderID rand error: ", err.Error())

		//mathrand.Read(bs)
		n := time.Now().UnixNano()
		for i := uint(0); i < 8; i ++ {
			bs[i] = byte((n >> i) & 0xff)
		}

		n = int64(mathrand.Int31())
		for i := uint(0); i < 4; i ++ {
			bs[i+4] = byte((n >> i) & 0xff)
		}
	}

	return string(base64.RawURLEncoding.EncodeToString(bs))
}

//func buildOrderID(accountId, planType string) string {
//	return fmt.Sprintf("%s_%s", accountId, planType)
//}

//==================================================================
//
//==================================================================

/*
func validateOrderMode(modeName string) (int, *Error) {
	switch modeName {
	case "prepay":
		return usage.OrderMode_Prepay, nil
	case "postpay":
		return usage.OrderMode_Postpay, nil
	}
	
	return -1, newInvalidParameterError("invalid mode parameter")
}
*/



func validateDfResName(planType, resName string, checkLength bool) *Error {
	if shouldResNameBeBlank(planType) && len(resName) > 0 {
		return newInvalidParameterError("resource_name should be blank: " + resName)
	}

	if openshift.IsDNS1123Label(resName) {
		return nil
	}

	if checkLength {
		n := len(resName)
		if n < 4 || n > 30 {
			return newInvalidParameterError("invalid df res name length")
		}
	}
	
	return newInvalidParameterError(fmt.Sprintf("%s is not a legal df res name", resName))
}

func validateOrderAutoGenID(idstr string) (int64, *Error) {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	id, e := _mustIntParam("id", idstr)
	return id, e
}

func validateOrderID(orderId string) (string, *Error) {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	orderId, e := _mustStringParam("orderId", orderId, 50, StringParamType_UrlWord)
	return orderId, e
}

func validateAccountID(accountId string) (string, *Error) {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	accountId, e := _mustStringParam("namespace", accountId, 50, StringParamType_UrlWord)
	return accountId, e
}

func validateUsername(accountId string) (string, *Error) {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	accountId, e := _mustStringParam("username", accountId, 50, StringParamType_UrlWord)
	return accountId, e
}

func validatePlanID(planId string) (string, *Error) {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	planId, e := _mustStringParam("plan_id", planId, 50, StringParamType_UrlWord)
	return planId, e
}

func validateRegion(region string) (string, *Error) {
	region, e := _mustStringParam("region", region, 50, StringParamType_UrlWord)
	return region, e
}

func validatePlanType(planType string) (string, *Error) {
	planType, e := _mustStringParam("plan-type", planType, 32, StringParamType_UrlWord)
	return planType, e
}

const (
	OrderStatusLabel_Pending   = "pending"
	OrderStatusLabel_Consuming = "consuming"
	//OrderStatusLabel_Ending    = "ending"
	OrderStatusLabel_Ended     = "ended"

	OrderStatusLabel_RenewalFailed = "renewfailed" // fake label
)

func validateOrderStatus(statusName string) (int, *Error) {
	var status = -1

	switch statusName {
	default:
		return -1, newInvalidParameterError("invalid status parameter")
	case OrderStatusLabel_Pending:
		status = usage.OrderStatus_Pending
	case OrderStatusLabel_Consuming, OrderStatusLabel_RenewalFailed:
		status = usage.OrderStatus_Consuming
	//case OrderStatusLabel_Ending:
	//	status = usage.OrderStatus_Ending
	case OrderStatusLabel_Ended:
		status = usage.OrderStatus_Ended
	}

	return status, nil
}

func orderStatusToLabel(status int) string {
	switch status {
	case usage.OrderStatus_Pending:
		return OrderStatusLabel_Pending
	case usage.OrderStatus_Consuming:
		return OrderStatusLabel_Consuming
	//case usage.OrderStatus_Ending:
	//	return OrderStatusLabel_Ending
	case usage.OrderStatus_Ended:
		return OrderStatusLabel_Ended
	}

	return ""
}

// ...

func validateAuth(region, token string) (string, *Error) {
	if token == "" {
		return "", GetError(ErrorCodeAuthFailed)
	}

	username, err := getDFUserame(region, token)
	if err != nil {
		return "", GetError2(ErrorCodeAuthFailed, err.Error())
	}

	return username, nil
}

func canManagePurchaseOrders(username string) bool {
	return username == "admin"
}

//==================================================================
// 
//==================================================================


type OrderCreationParams struct {
	ResName string  `json:"resource_name,omitempty"`
		// may be VolumeName, BsiName, etc.
}

type OrderCreation struct {
	AccountID string    `json:"namespace,omitempty"`
	PlanID    string    `json:"plan_id,omitempty"`
	//Creator   string    `json:"creator,omitempty"`

	Params OrderCreationParams `json:"parameters,omitempty"`
}

func CreateOrder(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	//if !canManagePurchaseOrders(username) {
	//	JsonResult(w, http.StatusForbidden, GetError(ErrorCodePermissionDenied), nil)
	//	return
	//}

	// ...

	orderCreation := &OrderCreation{}
	err := common.ParseRequestJsonInto(r, orderCreation)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeParseJsonFailed, err.Error()), nil)
		return
	}

	// ...

	planRegion := r.FormValue("region")
	planRegion, e := validateRegion(planRegion)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	planId, e := validatePlanID(orderCreation.PlanID)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	plan, err := getPlanByID(planId, planRegion)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGetPlan, err.Error()), nil)
		return
	}

	// assert planId == plan.Plan_id 
	planType := plan.Plan_type
	//planRegion := plan.Region

	// todo: switch 
	//>>
	if plan.Region != planRegion {
		Logger.Warning("region not match",planRegion, plan.Region,plan.Id)
		plan.Region = planRegion
	}
	//<<
	//if plan.Region != planRegion {
	//	JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGetPlan, "plan regions not match"), nil)
	//	return
	//}

	if ! isValidPlanType(planType) {
		JsonResult(w, http.StatusBadRequest, newInvalidParameterError("plan type is invalid"), nil)
		return
	}

	// ...

	if orderCreation.Params.ResName != "" {
		e = validateDfResName(planType, orderCreation.Params.ResName, true)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	// auth

	username, e := validateAuth(planRegion, r.Header.Get("Authorization"))
	if e != nil {
		Logger.Warning("validateAuth err", r.Header.Get("Authorization"), ",", planRegion, plan.Region,plan.Id, e)

		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	accountId := orderCreation.AccountID
	if accountId == "" {
		accountId = username
	} else {
		accountId, e = validateAccountID(accountId)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	// check if user can manipulate project or not
	if accountId != username {
		_, err = getDfProject(planRegion, username, r.Header.Get("Authorization"), accountId)
		if err != nil {
			JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodePermissionDenied, err.Error()), nil)
			return
		}
	}

	// if user is admin, ... (canceled)
	/*
	if orderCreation.Creator != "" {
		orderCreation.Creator, e = validateUsername(orderCreation.AccountID)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}

		// todo: validate auth username must be admin
	}
	*/
	creator := username

	// ...

	orderId := genOrderID(accountId, planType)

	// check if there is an old order

	oldOrder, err := usage.RetrieveOrderByID(db, orderId, plan.Region, usage.OrderStatus_Consuming)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGetOrder, err.Error()), nil)
		return
	}
	if oldOrder != nil {
		if oldOrder.Plan_id == planId {
			JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeCreateOrder, "Plan not changed"), nil)
			return
		}
		//if oldOrder.Region != plan.Region {
		//	JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeCreateOrder, fmt.Sprintf("Region not switchable (%s -> %s)", oldOrder.Region, plan.Region)), nil)
		//	return
		//}
	}

	// create new order (in pending status)

	now := time.Now()
	startTime := now
	endTime := now
	deadlineTime := now

	order := &usage.PurchaseOrder {
		Order_id: orderId,
		Account_id: accountId,

		Plan_id : planId,
		Plan_type: planType,
		Region: planRegion,

		Start_time: startTime,
		End_time: endTime,
		Deadline_time: deadlineTime,

		Status: usage.OrderStatus_Pending,

		Creator: creator,

		Resource_name: orderCreation.Params.ResName,
	}

	// drytry=1: don't create order actually
	drytry := r.FormValue("drytry") == "1"

	if !drytry {
		order.Id, err = usage.CreateOrder(db, order)
		if err != nil {
			JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeCreateOrder, err.Error()), nil)
			return
		}
	}

	// make the payment

	paymentMoney, err, specialErrCode := createOrder(drytry, db, &orderCreation.Params, order, plan, oldOrder)
	if err != nil {
		var errCode uint = ErrorCodeRenewOrder
		if specialErrCode > 0 {
			errCode = uint(specialErrCode)
		}
		JsonResult(w, http.StatusBadRequest, GetError2(errCode, err.Error()), nil)
		return
	}

	// ...

	if ! drytry {
		SendCreateOrderEmail(order, plan)
	}

	// ...

	result := struct {
		Money string               `json:"money,omitempty"`
		Order *usage.PurchaseOrder `json:"order,omitempty"`
	}{
		fmt.Sprintf("%.2f", paymentMoney),
		order,
	}

	JsonResult(w, http.StatusOK, nil, result)
}

type OrderModification struct {
	Action    string  `json:"action,omitempty"`
	AccountID string  `json:"namespace,omitempty"`
}

func ModifyOrder(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	// ...

	//orderId, e := validateOrderID(params.ByName("order_id"))
	//if e != nil {
	//	JsonResult(w, http.StatusBadRequest, e, nil)
	//	return
	//}
	id, e := validateOrderAutoGenID(params.ByName("id"))
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	orderMod := &OrderModification{}
	err := common.ParseRequestJsonInto(r, orderMod)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeParseJsonFailed, err.Error()), nil)
		return
	}

	// only orders in consuming status can be modified now.
	// there should be most one consuming order for a orderId.
	//oldOrder, err := usage.RetrieveOrderByID(db, orderId, usage.OrderStatus_Consuming)
	oldOrder, err := usage.RetrieveOrderByStatusAndAutoGenID(db, id, usage.OrderStatus_Consuming)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGetOrder, err.Error()), nil)
		return
	}
	if oldOrder == nil {
		JsonResult(w, http.StatusBadRequest, GetError(ErrorCodeOrderNotFound), nil)
		return
	}

	// auth

	username, e := validateAuth(oldOrder.Region, r.Header.Get("Authorization"))
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	//if !canManagePurchaseOrders(username) {
	//	JsonResult(w, http.StatusUnauthorized, GetError(ErrorCodePermissionDenied), nil)
	//	return
	//}

	accountId := orderMod.AccountID
	if accountId == "" {
		accountId = username
	} else {
		accountId, e = validateAccountID(accountId)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	if oldOrder.Account_id != accountId {
		JsonResult(w, http.StatusBadRequest, GetError(ErrorCodePermissionDenied), nil)
		return
	}

	// check if user can manipulate project or not
	if accountId != username {
		_, err = getDfProject(oldOrder.Region, username, r.Header.Get("Authorization"), accountId)
		if err != nil {
			JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodePermissionDenied, err.Error()), nil)
			return
		}
	}

	switch orderMod.Action {
	default: 

		JsonResult(w, http.StatusBadRequest, newInvalidParameterError(fmt.Sprintf("unknown action: %s", orderMod.Action)), nil)
		return

	case "cancel":

		err = endOrder(db, oldOrder)
		if err != nil {
			JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeCancelOrder, err.Error()), nil)
			return
		}

		go func() {
			plan, err := getPlanByID(oldOrder.Plan_id, oldOrder.Region)
			if err != nil {
				Logger.Errorf("cancel order (%d) getPlanByID (%s) error: %s", oldOrder.Id, oldOrder.Plan_id, err.Error())
				return
			}

			SendEndOrderEmail_CancelledManually(oldOrder, plan)
		}()
	}

	JsonResult(w, http.StatusOK, nil, nil)
}

func GetAccountOrder(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	// ...

	//orderId, e := validateOrderID(params.ByName("order_id"))
	//if e != nil {
	//	JsonResult(w, http.StatusBadRequest, e, nil)
	//	return
	//}
	id, e := validateOrderAutoGenID(params.ByName("id"))
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	//order, err := usage.RetrieveOrderByID(db, orderId, usage.OrderStatus_Consuming)
	// pending orders will not be returned
	order, err := usage.RetrieveOrderByStatusAndAutoGenID(db, id, usage.OrderStatus_Consuming)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGetOrder, err.Error()), nil)
		return
	}
	
	if order != nil {
		// auth

		username, e := validateAuth(order.Region, r.Header.Get("Authorization"))
		if e != nil {
			JsonResult(w, http.StatusUnauthorized, e, nil)
			return
		}

		accountId := r.FormValue("namespace")
		if accountId == "" {
			accountId = username
		} else {
			accountId, e = validateAccountID(accountId)
			if e != nil {
				JsonResult(w, http.StatusBadRequest, e, nil)
				return
			}
		}

		if order.Account_id != accountId {
			JsonResult(w, http.StatusBadRequest, GetError(ErrorCodePermissionDenied), nil)
			return
		}
		
		if accountId != username {
			// check if user can manipulate project or not
			_, err := getDfProject(order.Region, username, r.Header.Get("Authorization"), accountId)
			if err != nil {
				JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodePermissionDenied, err.Error()), nil)
				return
			}
		}
	}

	// ...

	result := struct {
		Order *usage.PurchaseOrder `json:"order,omitempty"`
	}{
		order,
	}

	JsonResult(w, http.StatusOK, nil, result)
}

func QueryAccountOrders(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	// ...
	/*
	region := r.FormValue("region")
	if region != "" {
		region, e = validateRegion(region)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}
	*/

	region, e := validateRegion(r.FormValue("region"))
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	resName := r.FormValue("resource_name")
	if resName != "" {
		e = validateDfResName("", resName, false)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	// ...

	planType := r.FormValue("type")
	if planType != "" {
		planType, e = validatePlanType(r.FormValue("type"))
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	// auth

	username, e := validateAuth(region, r.Header.Get("Authorization"))
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	accountId := r.FormValue("namespace")
	if accountId == "" {
		accountId = username
	} else {
		accountId, e = validateAccountID(accountId)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	// check if user can manipulate project or not
	if accountId != username {
		_, err := getDfProject(region, username, r.Header.Get("Authorization"), accountId)
		if err != nil {
			JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodePermissionDenied, err.Error()), nil)
			return
		}
	}

	// ...

	status, statusLabel := -1, r.FormValue("status")
	if statusLabel == "" {
		// status = usage.OrderStatus_Consuming 
		// zongsan: blank means consuming
		// cancelled
	} else {
		status, e = validateOrderStatus(statusLabel)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	renewalFailedOnly := statusLabel == OrderStatusLabel_RenewalFailed

	// ...
	
	offset, size := optionalOffsetAndSize(r, 30, 1, 100)
	//orderBy := usage.ValidateOrderBy(r.FormValue("orderby"))
	//sortOrder := usage.ValidateSortOrder(r.FormValue("sortorder"), false)

	count, orders, err := usage.QueryOrders(db, accountId, region, resName, planType, status, renewalFailedOnly, offset, size)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeQueryOrders, err.Error()), nil)
		return
	}

	for _, o := range orders {
		o.StatusLabel = orderStatusToLabel(o.Status)
	}

	// ...

	type NamedOrder struct {
		Order *usage.PurchaseOrder `json:"order,omitempty"`
	}
	result := make([]NamedOrder, len(orders))
	for i, o := range orders {
		result[i].Order = o
	}

	JsonResult(w, http.StatusOK, nil, newQueryListResult(count, result))
}

//==================================================================
// reports
//==================================================================

/*
type GroupedReports struct {
	Time    string                   `json:"time,omitempty"`
	Reports []*usage.ConsumingReport `json:"reports,omitempty"`
}

func groupReports(reports []*usage.ConsumingReport, timeStep int) []*GroupedReports {
	// todo

	return []*GroupedReports {
		{
			Time: reports[0].Time_tag,
			Reports: reports,
		},
	}
}
*/

func QueryAccountConsumingReports(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	// ...
	/*
	region := r.FormValue("region")
	if region != "" {
		region, e = validateRegion(region)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}
	*/

	region, e := validateRegion(r.FormValue("region"))
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	// auth

	username, e := validateAuth(region, r.Header.Get("Authorization"))
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	accountId := r.FormValue("namespace")
	if accountId == "" {
		accountId = username
	} else {
		accountId, e = validateAccountID(accountId)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	// check if user can manipulate project or not
	if accountId != username {
		_, err := getDfProject(region, username, r.Header.Get("Authorization"), accountId)
		if err != nil {
			JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodePermissionDenied, err.Error()), nil)
			return
		}
	}

	// ...

	orderId := r.FormValue("order")
	if orderId != "" {
		orderId, e = validateOrderID(orderId)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}
	}

	// ...

	offset, size := optionalOffsetAndSize(r, 30, 1, 100)
	//orderBy := usage.ValidateOrderBy(r.FormValue("orderby"))
	//sortOrder := usage.ValidateSortOrder(r.FormValue("sortorder"), false)

	count, consumings, err := usage.QueryConsumeHistories(db, accountId, orderId, region, offset, size)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeQueryConsumings, err.Error()), nil)
		return
	}

	JsonResult(w, http.StatusOK, nil, newQueryListResult(count, consumings))
}

/*
type ConsumingSpeed struct {
	Money    float64 `json:"money,omitempty"`
	Duration int     `json:"duration,omitempty"`
}

func GetAccountConsumingSpeed(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	
	speed := &ConsumingSpeed{
		Money: 56.78,
		Duration: usage.ReportStep_Day,
	}

	JsonResult(w, http.StatusOK, nil, speed)
	
	return
	/////////////////////////////////////////////////////////////////////////////////
}
*/










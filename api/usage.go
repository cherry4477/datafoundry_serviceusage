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

func genOrderID() string {
	bs := make([]byte, 12)
	_, err := rand.Read(bs)
	if err != nil {
		Logger.Warning("genUUID error: ", err.Error())

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

func validateOrderID(orderId string) (string, *Error) {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	orderId, e := _mustStringParam("id", orderId, 50, StringParamType_UrlWord)
	return orderId, e
}

func validateAccountID(accountId string) (string, *Error) {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	accountId, e := _mustStringParam("accountId", accountId, 50, StringParamType_UrlWord)
	return accountId, e
}

func validateUsername(accountId string) (string, *Error) {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	accountId, e := _mustStringParam("username", accountId, 50, StringParamType_UrlWord)
	return accountId, e
}

func validatePlanID(planId string) (string, *Error) {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	planId, e := _mustStringParam("planId", planId, 50, StringParamType_UrlWord)
	return planId, e
}

func validateOrderStatus(statusName string) (int, *Error) {
	var status = -1

	switch statusName {
	default:
		return -1, newInvalidParameterError("invalid status parameter")
	case "":
		status = -1
	case "pending":
		status = usage.OrderStatus_Pending
	case "consuming":
		status = usage.OrderStatus_Consuming
	case "ending":
		status = usage.OrderStatus_Ending
	case "ended":
		status = usage.OrderStatus_Ended
	}

	return status, nil
}

/*
func validateOrderName(name string, musBeNotBlank bool) (string, *Error) {
	if musBeNotBlank || name != "" {
		// most 20 Chinese chars
		name_param, e := _mustStringParam("name", name, 60, StringParamType_General)
		if e != nil {
			return "", e
		}
		name = name_param
	}

	return name, nil
}

func validateOrderVersion(version string, musBeNotBlank bool) (string, *Error) {
	if musBeNotBlank || version != "" {
		version_param, e := _mustStringParam("version", version, 32, StringParamType_General)
		if e != nil {
			return "", e
		}
		version = version_param
	}

	return version, nil
}

func validateOrderProvider(provider string, musBeNotBlank bool) (string, *Error) {
	if musBeNotBlank || provider != "" {
		// most 20 Chinese chars
		provider_param, e := _mustStringParam("provider", provider, 60, StringParamType_General)
		if e != nil {
			return "", e
		}
		provider = provider_param
	}

	return provider, nil
}

func validateOrderCategory(category string, musBeNotBlank bool) (string, *Error) {
	if musBeNotBlank || category != "" {
		// most 10 Chinese chars
		category_param, e := _mustStringParam("category", category, 32, StringParamType_General)
		if e != nil {
			return "", e
		}
		category = category_param
	}

	return category, nil
}

func validateOrderDescription(description string, musBeNotBlank bool) (string, *Error) {
	if musBeNotBlank || description != "" {
		// most about 666 Chinese chars
		description_param, e := _mustStringParam("description", description, 2000, StringParamType_General)
		if e != nil {
			return "", e
		}
		description = description_param
	}

	return description, nil
}

func validateUrl(url string, musBeNotBlank bool, paramName string) (string, *Error) {
	url = strings.TrimSpace(url)

	if len(url) > 200 {
		return "", newInvalidParameterError(fmt.Sprintf("%s is too long", paramName))
	}

	if url == "" {
		if musBeNotBlank {
			return "", newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
		}

		_, err := neturl.Parse(url)
		if err != nil {
			return "", newInvalidParameterError(err.Error())
		}
	}

	return url, nil
}
*/

// ...

func validateAuth(token string) (string, *Error) {
	if token == "" {
		return "", GetError(ErrorCodeAuthFailed)
	}

	username, err := getDFUserame(token)
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

type OrderCreation struct {
	AccountID string    `json:"project,omitempty"`
	PlanID    string    `json:"planId,omitempty"`
	Creator   string    `json:"creator,omitempty"`
}

func CreateOrder(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	JsonResult(w, http.StatusOK, nil, "98DED98A-F7A1-EDF2-3DF7-B799333D2FD3")
	
	return

	// the real implementation

	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	// auth

	username, e := validateAuth(r.Header.Get("Authorization"))
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	if !canManagePurchaseOrders(username) {
		JsonResult(w, http.StatusUnauthorized, GetError(ErrorCodePermissionDenied), nil)
		return
	}

	// ...

	orderCreation := &OrderCreation{}
	err := common.ParseRequestJsonInto(r, orderCreation)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeParseJsonFailed, err.Error()), nil)
		return
	}

	//orderMode, e := validateOrderMode(orderCreation.Mode)
	//if e != nil {
	//	JsonResult(w, http.StatusBadRequest, e, nil)
	//	return
	//}

	accountId, e := validateAccountID(orderCreation.AccountID)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	// todo: validate auth username accountId relation

	if orderCreation.Creator != "" {
		orderCreation.Creator, e = validateUsername(orderCreation.AccountID)
		if e != nil {
			JsonResult(w, http.StatusBadRequest, e, nil)
			return
		}

		// todo: validate auth username must be admin
	}

	planId, e := validatePlanID(orderCreation.PlanID)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	// todo: remote get plan

	planId = planId 
	planType := "unknown"
	planRegion := "unknown"

	// ...
	now := time.Now()
	startTime := now
	endTime := now
	nextConsumeTime := now

	order := &usage.PurchaseOrder{
		Order_id: genUUID(),
		Account_id: accountId,

		Plan_id : planId,
		Plan_type: planType,
		Region: planRegion,

		Start_time: startTime,
		End_time: endTime,
		Next_consume_time: nextConsumeTime,

		Status: usage.OrderStatus_Pending,
	}

	err = usage.CreateOrder(db, order)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeCreateOrder, err.Error()), nil)
		return
	}

	// todo: remote make payment
	// todo: change order status => consuming
	// todo: create a consuming report

	JsonResult(w, http.StatusOK, nil, order.Order_id)
}

type OrderModification struct {
	Action      string  `json:"action,omitempty"`
	AccountID string    `json:"project,omitempty"`
}

func ModifyOrder(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	JsonResult(w, http.StatusOK, nil, nil)
	
	return


	// the real implementation

	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	// auth

	username, e := validateAuth(r.Header.Get("Authorization"))
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	if !canManagePurchaseOrders(username) {
		JsonResult(w, http.StatusUnauthorized, GetError(ErrorCodePermissionDenied), nil)
		return
	}

	// ...

	orderId, e := validateOrderID(params.ByName("id"))
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

	accountId, e := validateAccountID(orderMod.AccountID)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	// todo: validate username accountId relation

	switch orderMod.Action {
	default: 

		JsonResult(w, http.StatusBadRequest, newInvalidParameterError(fmt.Sprintf("unknown action: %s", orderMod.Action)), nil)
		return

	case "cancel":

		// todo: different plan types may need different handlings

		err = usage.EndOrder(db, orderId, accountId)
		if err != nil {
			JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeModifyOrder, err.Error()), nil)
			return
		}

	}

	JsonResult(w, http.StatusOK, nil, nil)
}

func GetAccountOrder(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	order := &usage.PurchaseOrder {
		Order_id: "98DED98A-F7A1-EDF2-3DF7-B799333D2FD3",
		Account_id: "88DED98A-F7A1-EDF2-3DF7-B799333D2FD3",
		Region: "bj",
		Quantities: 1,
		Plan_id: "89DED98A-F7A1-EDF2-3DF7-A799333D2FD3",
		Start_time: time.Date(2016, time.May, 10, 23, 0, 0, 0, time.UTC),
		EndTime: nil,
		Status: usage.OrderStatus_Consuming,
	}

	JsonResult(w, http.StatusOK, nil, order)
	
	return


	// the real implementation

	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	// auth

	username, e := validateAuth(r.Header.Get("Authorization"))
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	accountId, e := validateAccountID(r.FormValue("project"))
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	// todo: check if username has the permission to view orders of accountId.
	_, _ = username, accountId

	// ...

	orderId, e := validateOrderID(params.ByName("id"))
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	order, err := usage.RetrieveOrderByID(db, orderId)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGetOrder, err.Error()), nil)
		return
	}

	JsonResult(w, http.StatusOK, nil, order)
}

func QueryAccountOrders(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	orders := []*usage.PurchaseOrder {
		{
			Order_id: "98DED98A-F7A1-EDF2-3DF7-B799333D2FD3",
			Account_id: "88DED98A-F7A1-EDF2-3DF7-B799333D2FD3",
			Region: "bj",
			Quantities: 1,
			Plan_id: "89DED98A-F7A1-EDF2-3DF7-A799333D2FD3",
			Start_time: time.Date(2016, time.May, 10, 23, 0, 0, 0, time.UTC),
			EndTime: nil,
			Status: usage.OrderStatus_Consuming,
		},
		{
			Order_id: "98DED98A-F7A1-EDF2-3DF7-B799333D2FD5",
			Account_id: "88DED98A-F7A1-EDF2-3DF7-B799333D2FD3",
			Region: "bj",
			Quantities: 1,
			Plan_id: "89DED98A-F7A1-EDF2-3DF7-A799333D2FD3",
			Start_time: time.Date(2016, time.May, 10, 23, 0, 0, 0, time.UTC),
			EndTime: nil,
			Status: usage.OrderStatus_Consuming,
		},
	}

	JsonResult(w, http.StatusOK, nil, newQueryListResult(1000, orders))
	
	return


	// the real implementation

	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	// auth

	username, e := validateAuth(r.Header.Get("Authorization"))
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}

	accountId, e := validateAccountID(r.FormValue("project"))
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	// todo: check if username has the permission to view orders of accountId.
	_, _ = username, accountId

	// ...

	status, e := validateOrderStatus(r.FormValue("status"))
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}
	
	offset, size := optionalOffsetAndSize(r, 30, 1, 100)
	//orderBy := usage.ValidateOrderBy(r.FormValue("orderby"))
	//sortOrder := usage.ValidateSortOrder(r.FormValue("sortorder"), false)

	count, orders, err := usage.QueryOrders(db, accountId, status, offset, size)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeQueryOrders, err.Error()), nil)
		return
	}

	JsonResult(w, http.StatusOK, nil, newQueryListResult(count, orders))
}

//==================================================================
//
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
	reports := []*usage.ConsumingReport {
		{
			Order_id: "98DED98A-F7A1-EDF2-3DF7-B799333D2FD5",
			Consume_time: time.Date(2016, time.May, 10, 24, 0, 0, 0, time.UTC),
			Money: 1.23,
			Plan_id: "89DED98A-F7A1-EDF2-3DF7-9799333D2FD3",
		},
		{
			Order_id: "98DED98A-F7A1-EDF2-3DF7-B799333D2FD5",
			Consume_time: time.Date(2016, time.May, 10, 25, 0, 0, 0, time.UTC),
			Money: 1.23,
			Plan_id: "89DED98A-F7A1-EDF2-3DF7-9799333D2FD3",
		},
	}

	JsonResult(w, http.StatusOK, nil, newQueryListResult(1000, reports))
	
	return
	/////////////////////////////////////////////////////////////////////////////////
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










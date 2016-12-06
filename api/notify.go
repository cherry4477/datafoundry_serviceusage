package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	//"strings"
	//"time"
	"github.com/asiainfoLDP/datahub_commons/common"
	//stat "github.com/asiainfoLDP/datafoundry_serviceusage/statistics"
	//"github.com/asiainfoLDP/datahub_commons/log"

	"github.com/asiainfoLDP/datafoundry_serviceusage/usage"
	//"github.com/asiainfoLDP/datahub_commons/log"
)

//======================================================
//
//======================================================

// order is ended by project owner self

type MessageOrEmail struct {
	Reason string               `json:reason,omitempty`
	Order  *usage.PurchaseOrder `json:order,omitempty`
	Plan   *Plan                `json:plan,omitempty`
}

func SendEmailOrMessage(order *usage.PurchaseOrder, plan *Plan, reason string) {
	if Debug {
		return
	}
	oc := osAdminClients[order.Region]
	if oc == nil {
		Logger.Errorf("SendEmailOrMessage %s getadmin token \n", reason)
		return
	}
	message := MessageOrEmail{
		Reason: reason,
		Order:  order,
		Plan:   plan,
	}
	url := fmt.Sprintf("%s/lapi/inbox?type=orderevent", SendMessageService)
	data, err := json.Marshal(message)
	if err != nil {
		Logger.Errorf("SendEmailOrMessage %s Marshal error: %s\n", reason, err.Error())
		return
	}
	response, response_data, err := common.RemoteCallWithJsonBody("POST", url, oc.BearerToken(), "", data)
	if err != nil {
		Logger.Errorf("SendEmailOrMessage %s error: %s", reason, err.Error())
		return
	}
	if response.StatusCode != http.StatusOK {
		Logger.Errorf("SendEmailOrMessage %s remote (%s) status code: %d. data=%s", reason, url, response.StatusCode, string(response_data))
		return
	}
		Logger.Error("SendEmailOrMessage is success")
}
func SendCreateOrderEmail(order *usage.PurchaseOrder, plan *Plan) {
	SendEmailOrMessage(order, plan, "order_created")
}

// warning balance insufficient
func SendBalanceInsufficientEmail(order *usage.PurchaseOrder, plan *Plan) {
	SendEmailOrMessage(order, plan, "order_renew_failed")
}

// order is ended for insufficient balance
func SendEndOrderEmail_BalanceInsufficient(order *usage.PurchaseOrder, plan *Plan) {
	SendEmailOrMessage(order, plan, "order_closed")
}

// order is cancelled by project owner self
func SendEndOrderEmail_CancelledManually(order *usage.PurchaseOrder, plan *Plan) {
	SendEmailOrMessage(order, plan, "order_cancled")
}

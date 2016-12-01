package api

import (
	"net/http"
	"fmt"
	"encoding/json"
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

type MessageOrEmail struct{
	Reason string                `json:reason,omitempty`
	Order  *usage.PurchaseOrder  `json:order,omitempty`
	Plan   *Plan                 `json:plan,omitempty`
}

func SendCreateOrderEmail(order *usage.PurchaseOrder, plan *Plan) {
		if Debug {
			return
		}
		oc := osAdminClients[order.Region]
		if oc == nil {
			Logger.Errorf("SendCreateOrderEmail getadmin token \n")
			return
		}
	    message := MessageOrEmail{
	 		Reason: "order_created",
			Order: order,
			Plan: plan,
	    }
		url := fmt.Sprintf("%s/lapi/inbox?type=orderevent", SendMessageService)
		data, err := json.Marshal(message)
		if err != nil {
			Logger.Errorf("SendCreateOrderEmail Marshal error: %s\n", err.Error())
			return 
		}
		response, _, err := common.RemoteCallWithJsonBody("POST", url, oc.BearerToken(), "", data)
		if err != nil {
			Logger.Errorf("SendCreateOrderEmail error: %s", err.Error())
			return
	    }
		if response.StatusCode != http.StatusOK {
			Logger.Errorf("SendCreateOrderEmail remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
			return 
	    }
}

// warning balance insufficient
func SendBalanceInsufficientEmail(order *usage.PurchaseOrder, plan *Plan) {
		if Debug {
			return
		}
		oc := osAdminClients[order.Region]
		if oc == nil {
			Logger.Errorf("SendBalanceInsufficientEmail getadmin token \n")
			return
		}
	    message := MessageOrEmail{
	 		Reason: "order_field",
			Order: order,
			Plan: plan,
	    }
		url := fmt.Sprintf("%s/lapi/inbox?type=orderevent", SendMessageService)
		data, err := json.Marshal(message)
		if err != nil {
			Logger.Errorf("SendBalanceInsufficientEmail Marshal error: %s\n", err.Error())
			return 
		}
		response, _, err := common.RemoteCallWithJsonBody("POST", url, oc.BearerToken(), "", data)
		if err != nil {
			Logger.Errorf("SendBalanceInsufficientEmail error: %s", err.Error())
			return
	    }
		if response.StatusCode != http.StatusOK {
			Logger.Errorf("SendBalanceInsufficientEmail remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
			return 
	    }
}

// order is ended for insufficient balance
func SendEndOrderEmail_BalanceInsufficient(order *usage.PurchaseOrder, plan *Plan) {
	if Debug {
			return
		}
		oc := osAdminClients[order.Region]
		if oc == nil {
			Logger.Errorf("SendEndOrderEmail_BalanceInsufficient getadmin token \n")
			return
		}
	    message := MessageOrEmail{
	 		Reason: "order will fall due",
	    }
		url := fmt.Sprintf("%s/lapi/inbox?type=orderevent", SendMessageService)
		data, err := json.Marshal(message)
		if err != nil {
			Logger.Errorf("SendEndOrderEmail_BalanceInsufficient Marshal error: %s\n", err.Error())
			return 
		}
		response, _, err := common.RemoteCallWithJsonBody("POST", url, oc.BearerToken(), "", data)
		if err != nil {
			Logger.Errorf("SendEndOrderEmail_BalanceInsufficient error: %s", err.Error())
			return
	    }
		if response.StatusCode != http.StatusOK {
			Logger.Errorf("SendEndOrderEmail_BalanceInsufficient remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
			return 
	    }
}

// order is cancelled by project owner self
func SendEndOrderEmail_CancelledManually(order *usage.PurchaseOrder, plan *Plan) {
	if Debug {
			return
		}
		oc := osAdminClients[order.Region]
		if oc == nil {
			Logger.Errorf("SendEndOrderEmail_CancelledManually getadmin token \n")
			return
		}
	    message := MessageOrEmail{
	 		Reason: "Successful cancellation of the order",
			Order: order,
			Plan: plan,
	    }
		url := fmt.Sprintf("%s/lapi/inbox?type=orderevent", SendMessageService)
		data, err := json.Marshal(message)
		if err != nil {
			Logger.Errorf("SendEndOrderEmail_CancelledManually Marshal error: %s\n", err.Error())
			return 
		}
		response, _, err := common.RemoteCallWithJsonBody("POST", url, oc.BearerToken(), "", data)
		if err != nil {
			Logger.Errorf("SendEndOrderEmail_CancelledManually error: %s", err.Error())
			return
	    }
		if response.StatusCode != http.StatusOK {
			Logger.Errorf("SendEndOrderEmail_CancelledManually remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
			return 
	    }
}
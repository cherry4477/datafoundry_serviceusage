package api

import (
	//"fmt"
	//"time"

	//stat "github.com/asiainfoLDP/datafoundry_serviceusage/statistics"
	//"github.com/asiainfoLDP/datahub_commons/log"

	"github.com/asiainfoLDP/datafoundry_serviceusage/usage"
)

//======================================================
// 
//======================================================

// order is ended by project owner self
func SendCreateOrderEmail(order *usage.PurchaseOrder, plan *Plan) {

}

// warning balance insufficient
func SendBalanceInsufficientEmail(order *usage.PurchaseOrder, plan *Plan) {

}

// order is ended for insufficient balance
func SendEndOrderEmail_BalanceInsufficient(order *usage.PurchaseOrder, plan *Plan) {

}

// order is cancelled by project owner self
func SendEndOrderEmail_CancelledManually(order *usage.PurchaseOrder, plan *Plan) {

}
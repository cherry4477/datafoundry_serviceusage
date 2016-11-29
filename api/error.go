package api

import (
	"fmt"
)

type Error struct {
	code    uint
	message string
}

var (
	Errors            [NumErrors]*Error
	ErrorNone         *Error
	ErrorUnkown       *Error
	ErrorJsonBuilding *Error
)

const (
	ErrorCodeAny = -1 // this is for testing only
	
	ErrorCodeNone = 0

	ErrorCodeUnkown                = 3300
	ErrorCodeJsonBuilding          = 3301
	ErrorCodeParseJsonFailed       = 3302
	ErrorCodeUrlNotSupported       = 3303
	ErrorCodeDbNotInitlized        = 3304
	ErrorCodeAuthFailed            = 3305
	ErrorCodePermissionDenied      = 3306
	ErrorCodeInvalidParameters     = 3307
	ErrorCodeCreateOrder           = 3308
	ErrorCodeDeleteOrder           = 3309
	ErrorCodeModifyOrder           = 3310
	ErrorCodeGetOrder              = 3311
	ErrorCodeQueryOrders           = 3312
	ErrorCodeRenewOrder            = 3313
	ErrorCodeQueryConsumings       = 3314
	ErrorCodeGetPlan               = 3315
	ErrorCodeInsufficentBalance    = 3316 // DON'T CHAGNE	

	ErrorCodeChargedButFailedToCreateResource = 3317 // DON'T CHAGNE	



	NumErrors = 3500 // about 32k memroy wasted
)

func init() {
	initError(ErrorCodeNone, "OK")
	initError(ErrorCodeUnkown, "unknown error")
	initError(ErrorCodeJsonBuilding, "json building error")
	initError(ErrorCodeParseJsonFailed, "parse json failed")

	initError(ErrorCodeUrlNotSupported, "unsupported url")
	initError(ErrorCodeDbNotInitlized, "db is not inited")
	initError(ErrorCodeAuthFailed, "auth failed")
	initError(ErrorCodePermissionDenied, "permission denied")
	initError(ErrorCodeInvalidParameters, "invalid parameters")

	initError(ErrorCodeCreateOrder, "failed to create order")
	initError(ErrorCodeDeleteOrder, "failed to delete order")
	initError(ErrorCodeModifyOrder, "failed to modify order")
	initError(ErrorCodeGetOrder, "failed to retrieve order")
	initError(ErrorCodeQueryOrders, "failed to query orders")
	initError(ErrorCodeRenewOrder, "failed to renew order")

	initError(ErrorCodeQueryConsumings, "failed to consuming history")

	initError(ErrorCodeGetPlan, "failed to get plan")

	initError(ErrorCodeInsufficentBalance, "insufficient balance")
	initError(ErrorCodeChargedButFailedToCreateResource, "charged but failed to create resource")

	ErrorNone = GetError(ErrorCodeNone)
	ErrorUnkown = GetError(ErrorCodeUnkown)
	ErrorJsonBuilding = GetError(ErrorCodeJsonBuilding)
}

func initError(code uint, message string) {
	if code < NumErrors {
		Errors[code] = newError(code, message)
	}
}

func GetError(code uint) *Error {
	if code > NumErrors {
		return Errors[ErrorCodeUnkown]
	}

	return Errors[code]
}

func GetError2(code uint, message string) *Error {
	e := GetError(code)
	if e == nil {
		return newError(code, message)
	} else {
		return newError(code, fmt.Sprintf("%s (%s)", e.message, message))
	}
}

func newError(code uint, message string) *Error {
	return &Error{code: code, message: message}
}

func newUnknownError(message string) *Error {
	return &Error{
		code:    ErrorCodeUnkown,
		message: message,
	}
}

func newInvalidParameterError(paramName string) *Error {
	return &Error{
		code:    ErrorCodeInvalidParameters,
		message: fmt.Sprintf("%s: %s", GetError(ErrorCodeInvalidParameters).message, paramName),
	}
}

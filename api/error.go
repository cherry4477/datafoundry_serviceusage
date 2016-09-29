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

	initError(ErrorCodeCreateOrder, "failed to create app")
	initError(ErrorCodeDeleteOrder, "failed to delete app")
	initError(ErrorCodeModifyOrder, "failed to modify app")
	initError(ErrorCodeGetOrder, "failed to retrieve app")
	initError(ErrorCodeQueryOrders, "failed to query apps")

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

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
	//"net"
	"sync"

	"github.com/julienschmidt/httprouter"
	//"github.com/miekg/dns"

	_ "github.com/go-sql-driver/mysql"

	"github.com/asiainfoLDP/datahub_commons/common"
	"github.com/asiainfoLDP/datahub_commons/log"

	"github.com/asiainfoLDP/datafoundry_serviceusage/usage"
)

//======================================================
// 
//======================================================

const (
	Platform_Local  = "local"
	Platform_DataOS = "dataos"
)

var Platform = Platform_DataOS

var Port int
var Debug = false
var Logger = log.DefaultLogger()

func Init(router *httprouter.Router) bool {
	Platform = os.Getenv("CLOUD_PLATFORM")
	if Platform == "" {
		Platform = Platform_DataOS
	}

	if initDB() == false {
		return false
	}

	initGateWay()
	initRouter(router)

	return true
}

func initRouter(router *httprouter.Router) {
	router.POST("/usageapi/v1/orders", TimeoutHandle(60000*time.Millisecond, CreateOrder))
	//router.PUT("/usageapi/v1/orders/:order_id", TimeoutHandle(5000*time.Millisecond, ModifyOrder))
	//router.GET("/usageapi/v1/orders/:order_id", TimeoutHandle(4600*time.Millisecond, GetAccountOrder))
	router.PUT("/usageapi/v1/orders/:id", TimeoutHandle(60000*time.Millisecond, ModifyOrder))
			// timeout for the above api should be larger than 15 seconds, for delete bsi need some time to be done.
	router.GET("/usageapi/v1/orders/:id", TimeoutHandle(60000*time.Millisecond, GetAccountOrder))
	router.GET("/usageapi/v1/orders", TimeoutHandle(60000*time.Millisecond, QueryAccountOrders))
	router.GET("/usageapi/v1/usages", TimeoutHandle(60000*time.Millisecond, QueryAccountConsumingReports))
	//router.GET("/usageapi/v1/speed", TimeoutHandle(500*time.Millisecond, GetAccountConsumingSpeed))
}

//=============================
//
//=============================

func MysqlAddrPort() (string, string) {
	//switch Platform {
	//case Platform_DataOS:
	return os.Getenv(os.Getenv("ENV_NAME_MYSQL_ADDR")), os.Getenv(os.Getenv("ENV_NAME_MYSQL_PORT"))
	//case Platform_Local:
	//	return os.Getenv("MYSQL_PORT_3306_TCP_ADDR"), os.Getenv("MYSQL_PORT_3306_TCP_PORT")
	//}
	//
	//return "", ""
}

func MysqlDatabaseUsernamePassword() (string, string, string) {
	//switch Platform {
	//case Platform_DataOS:
	return os.Getenv(os.Getenv("ENV_NAME_MYSQL_DATABASE")),
		os.Getenv(os.Getenv("ENV_NAME_MYSQL_USER")),
		os.Getenv(os.Getenv("ENV_NAME_MYSQL_PASSWORD"))
	//}
	//
	//return os.Getenv("MYSQL_ENV_MYSQL_DATABASE"),
	//	os.Getenv("MYSQL_ENV_MYSQL_USER"),
	//	os.Getenv("MYSQL_ENV_MYSQL_PASSWORD")
}

var (
	dbInstance *sql.DB
	dbMutex sync.Mutex
)

func DB() *sql.DB {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	return dbInstance
}

func getDB() *sql.DB {
	if usage.IsServing() {
		return DB()
	} else {
		return nil
	}
}

func setDB(db *sql.DB) {
	dbMutex.Lock()
	dbInstance = db
	dbMutex.Unlock()
}

func initDB() bool {
	// return true // temp, mysqlnocase.servicebroker.dataos.io is not available now.

	for i := 0; i < 3; i++ {
		connectDB()
		
		if DB() == nil {
			select {
			case <-time.After(time.Second * 10):
				continue
			}
		} else {
			break
		}
	}

	if DB() == nil {
		return false
	}

	// upgradeDB()
	err := usage.TryToUpgradeDatabase(DB(), "datafoundry:serviceusage", os.Getenv("MYSQL_CONFIG_DONT_UPGRADE_TABLES") != "yes") // don't change the name
	if err != nil {
		Logger.Errorf("TryToUpgradeDatabase error: %s", err.Error())
		return false
	}

	// ...

	go StartMaintaining()

	go updateDB()

	return true
}

func updateDB() {
	var err error
	ticker := time.Tick(5 * time.Second)
	for range ticker {
		db := getDB()
		if db == nil {
			connectDB()
		} else if err = db.Ping(); err != nil {
			Logger.Errorf("db ping error: %s\n", err)
			db.Close()
			// setDB(nil) // draw snake feet
			connectDB()
		}
	}
}

func connectDB() {
	DB_ADDR, DB_PORT := MysqlAddrPort()
	DB_DATABASE, DB_USER, DB_PASSWORD := MysqlDatabaseUsernamePassword()

	DB_URL := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true`, DB_USER, DB_PASSWORD, DB_ADDR, DB_PORT, DB_DATABASE)

	Logger.Info("connect to ", DB_URL)
	db, err := sql.Open("mysql", DB_URL) // ! here, err is always nil, db is never nil.
	if err == nil {
		err = db.Ping()
	}

	if err != nil {
		Logger.Errorf("error: %s\n", err)
		if db != nil {
			db.Close()
		}
		return
	}
	
	setDB(db)
}

//======================================================
// errors
//======================================================

const (
	StringParamType_General        = 0
	StringParamType_UrlWord        = 1
	StringParamType_UnicodeUrlWord = 2
	StringParamType_Email          = 3
)

//======================================================
//
//======================================================

var Json_ErrorBuildingJson []byte

func getJsonBuildingErrorJson() []byte {
	if Json_ErrorBuildingJson == nil {
		Json_ErrorBuildingJson = []byte(fmt.Sprintf(`{"code": %d, "msg": %s}`, ErrorJsonBuilding.code, ErrorJsonBuilding.message))
	}

	return Json_ErrorBuildingJson
}

type Result struct {
	Code uint        `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// if data only has one item, then the item key will be ignored.
func JsonResult(w http.ResponseWriter, statusCode int, e *Error, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if e == nil {
		e = ErrorNone
	}
	result := Result{Code: e.code, Msg: e.message, Data: data}
	jsondata, err := json.MarshalIndent(&result, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(getJsonBuildingErrorJson()))
	} else {
		w.WriteHeader(statusCode)
		w.Write(jsondata)
	}
}

type QueryListResult struct {
	Total   int64       `json:"total"`
	Results interface{} `json:"results"`
}

func newQueryListResult(count int64, results interface{}) *QueryListResult {
	return &QueryListResult{Total: count, Results: results}
}

//======================================================
//
//======================================================

func mustBoolParam(params httprouter.Params, paramName string) (bool, *Error) {
	bool_str := params.ByName(paramName)
	if bool_str == "" {
		return false, newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
	}

	b, err := strconv.ParseBool(bool_str)
	if err != nil {
		return false, newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, bool_str))
	}

	return b, nil
}

func mustBoolParamInQuery(r *http.Request, paramName string) (bool, *Error) {
	bool_str := r.Form.Get(paramName)
	if bool_str == "" {
		return false, newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
	}

	b, err := strconv.ParseBool(bool_str)
	if err != nil {
		return false, newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, bool_str))
	}

	return b, nil
}

func optionalBoolParamInQuery(r *http.Request, paramName string, defaultValue bool) bool {
	bool_str := r.Form.Get(paramName)
	if bool_str == "" {
		return defaultValue
	}

	b, err := strconv.ParseBool(bool_str)
	if err != nil {
		return defaultValue
	}

	return b
}

func _mustIntParam(paramName string, int_str string) (int64, *Error) {
	if int_str == "" {
		return 0, newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
	}

	i, err := strconv.ParseInt(int_str, 10, 64)
	if err != nil {
		return 0, newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, int_str))
	}

	return i, nil
}

func mustIntParamInQuery(r *http.Request, paramName string) (int64, *Error) {
	return _mustIntParam(paramName, r.Form.Get(paramName))
}

func mustIntParamInPath(params httprouter.Params, paramName string) (int64, *Error) {
	return _mustIntParam(paramName, params.ByName(paramName))
}

func mustIntParamInMap(m map[string]interface{}, paramName string) (int64, *Error) {
	v, ok := m[paramName]
	if ok {
		i, ok := v.(float64)
		if ok {
			return int64(i), nil
		}

		return 0, newInvalidParameterError(fmt.Sprintf("param %s is not int", paramName))
	}

	return 0, newInvalidParameterError(fmt.Sprintf("param %s is not found", paramName))
}

func _optionalIntParam(intStr string, defaultInt int64) int64 {
	if intStr == "" {
		return defaultInt
	}

	i, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return defaultInt
	} else {
		return i
	}
}

func optionalIntParamInQuery(r *http.Request, paramName string, defaultInt int64) int64 {
	return _optionalIntParam(r.Form.Get(paramName), defaultInt)
}

func optionalIntParamInMap(m map[string]interface{}, paramName string, defaultValue int64) int64 {
	v, ok := m[paramName]
	if ok {
		i, ok := v.(float64)
		if ok {
			return int64(i)
		}
	}

	return defaultValue
}

func mustFloatParam(params httprouter.Params, paramName string) (float64, *Error) {
	float_str := params.ByName(paramName)
	if float_str == "" {
		return 0.0, newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
	}

	f, err := strconv.ParseFloat(float_str, 64)
	if err != nil {
		return 0.0, newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, float_str))
	}

	return f, nil
}

func _mustStringParam(paramName string, str string, maxLength int, paramType int) (string, *Error) {
	if str == "" {
		return "", newInvalidParameterError(fmt.Sprintf("%s can't be blank", paramName))
	}
	if len(str) > maxLength {
		return "", newInvalidParameterError(fmt.Sprintf("%s is too long", paramName))
	}

	if paramType == StringParamType_UrlWord {
		str2, ok := common.ValidateUrlWord(str)
		if !ok {
			return "", newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, str))
		}
		str = str2
	} else if paramType == StringParamType_UnicodeUrlWord {
		str2, ok := common.ValidateUnicodeUrlWord(str)
		if !ok {
			return "", newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, str))
		}
		str = str2
	} else if paramType == StringParamType_Email {
		str2, ok := common.ValidateEmail(str)
		if !ok {
			return "", newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, str))
		}
		str = str2
	} else { // if paramType == StringParamType_General
		str2, ok := common.ValidateGeneralWord(str)
		if !ok {
			return "", newInvalidParameterError(fmt.Sprintf("%s=%s", paramName, str))
		}
		str = str2
	}

	return str, nil
}

func mustStringParamInPath(params httprouter.Params, paramName string, maxLength int, paramType int) (string, *Error) {
	return _mustStringParam(paramName, params.ByName(paramName), maxLength, paramType)
}

func mustStringParamInQuery(r *http.Request, paramName string, maxLength int, paramType int) (string, *Error) {
	return _mustStringParam(paramName, r.Form.Get(paramName), maxLength, paramType)
}

func mustStringParamInMap(m map[string]interface{}, paramName string, maxLength int, paramType int) (string, *Error) {
	v, ok := m[paramName]
	if ok {
		str, ok := v.(string)
		if ok {
			return _mustStringParam(paramName, str, maxLength, paramType)
		}

		return "", newInvalidParameterError(fmt.Sprintf("param %s is not string", paramName))
	}

	return "", newInvalidParameterError(fmt.Sprintf("param %s is not found", paramName))
}

//======================================================
//
//======================================================

//func mustCurrentUserName(r *http.Request) (string, *Error) {
//	username, _, ok := r.BasicAuth()
//	if !ok {
//		return "", GetError(ErrorCodeAuthFailed)
//	}
//
//	return username, nil
//}

func mustCurrentUserName(r *http.Request) (string, *Error) {
	username := r.Header.Get("User")
	if username == "" {
		return "", GetError(ErrorCodeAuthFailed)
	}

	// needed?
	//username, ok = common.ValidateEmail(username)
	//if !ok {
	//	return "", newInvalidParameterError(fmt.Sprintf("user (%s) is not valid", username))
	//}

	return username, nil
}

func getCurrentUserName(r *http.Request) string {
	return r.Header.Get("User")
}

func optionalOffsetAndSize(r *http.Request, defaultSize int64, minSize int64, maxSize int64) (int64, int) {
	page := optionalIntParamInQuery(r, "page", 0)
	if page < 1 {
		page = 1
	}
	page -= 1

	if minSize < 1 {
		minSize = 1
	}
	if maxSize < 1 {
		maxSize = 1
	}
	if minSize > maxSize {
		minSize, maxSize = maxSize, minSize
	}

	size := optionalIntParamInQuery(r, "size", defaultSize)
	if size < minSize {
		size = minSize
	} else if size > maxSize {
		size = maxSize
	}

	return page * size, int(size)
}

func mustOffsetAndSize(r *http.Request, defaultSize, minSize, maxSize int) (offset int64, size int, e *Error) {
	if minSize < 1 {
		minSize = 1
	}
	if maxSize < 1 {
		maxSize = 1
	}
	if minSize > maxSize {
		minSize, maxSize = maxSize, minSize
	}

	page_size := int64(defaultSize)
	if r.Form.Get("size") != "" {
		page_size, e = mustIntParamInQuery(r, "size")
		if e != nil {
			return
		}
	}

	size = int(page_size)
	if size < minSize {
		size = minSize
	} else if size > maxSize {
		size = maxSize
	}

	// ...

	page := int64(0)
	if r.Form.Get("page") != "" {
		page, e = mustIntParamInQuery(r, "page")
		if e != nil {
			return
		}
		if page < 1 {
			page = 1
		}
		page -= 1
	}

	offset = page * page_size

	return
}

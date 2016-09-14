package api

import (
	"net/http"
	"time"
	"crypto/rand"
	"fmt"
	"strings"
	mathrand "math/rand"
	neturl "net/url"
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

func validateAppInfo(app *usage.SaasApp) *Error {
	var e *Error

	app.Name, e = validateAppName(app.Name, true)
	if e != nil {
		return e
	}

	app.Version, e = validateAppVersion(app.Version, true)
	if e != nil {
		return e
	}

	app.Provider, e = validateAppProvider(app.Provider, true)
	if e != nil {
		return e
	}

	app.Category, e = validateAppCategory(app.Category, true)
	if e != nil {
		return e
	}

	app.Description, e = validateAppDescription(app.Description, true)
	if e != nil {
		return e
	}

	app.Url, e = validateUrl(app.Url, true, "url")
	if e != nil {
		return e
	}

	app.Icon_url, e = validateUrl(app.Icon_url, true, "iconUrl")
	if e != nil {
		return e
	}

	return nil
}

func validateAppID(appId string) *Error {
	// GetError2(ErrorCodeInvalidParameters, err.Error())
	_, e := _mustStringParam("appid", appId, 50, StringParamType_UrlWord)
	return e
}

func validateAppName(name string, musBeNotBlank bool) (string, *Error) {
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

func validateAppVersion(version string, musBeNotBlank bool) (string, *Error) {
	if musBeNotBlank || version != "" {
		version_param, e := _mustStringParam("version", version, 32, StringParamType_General)
		if e != nil {
			return "", e
		}
		version = version_param
	}

	return version, nil
}

func validateAppProvider(provider string, musBeNotBlank bool) (string, *Error) {
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

func validateAppCategory(category string, musBeNotBlank bool) (string, *Error) {
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

func validateAppDescription(description string, musBeNotBlank bool) (string, *Error) {
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

func canEditSaasApps(username string) bool {
	return username == "admin"
}

//==================================================================
//
//==================================================================

func CreateApp(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	if !canEditSaasApps(username) {
		JsonResult(w, http.StatusUnauthorized, GetError(ErrorCodePermissionDenied), nil)
		return
	}

	// ...

	app := &usage.SaasApp{}
	err := common.ParseRequestJsonInto(r, app)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeParseJsonFailed, err.Error()), nil)
		return
	}

	e = validateAppInfo(app)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}
	
	app.App_id = genUUID()
	// followings will be ignored
	//app.Create_time = time.Now()
	//app.Hotness = 0

	err = usage.CreateApp(db, app)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeCreateApp, err.Error()), nil)
		return
	}

	JsonResult(w, http.StatusOK, nil, app.App_id)
}

func DeleteApp(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	if !canEditSaasApps(username) {
		JsonResult(w, http.StatusUnauthorized, GetError(ErrorCodePermissionDenied), nil)
		return
	}

	// ...

	appId := params.ByName("id")

	e = validateAppID(appId)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	err := usage.DeleteApp(db, appId)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeDeleteApp, err.Error()), nil)
		return
	}

	JsonResult(w, http.StatusOK, nil, nil)
}

func ModifyApp(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
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

	if !canEditSaasApps(username) {
		JsonResult(w, http.StatusUnauthorized, GetError(ErrorCodePermissionDenied), nil)
		return
	}

	// ...

	appId := params.ByName("id")

	e = validateAppID(appId)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	app := &usage.SaasApp{}
	err := common.ParseRequestJsonInto(r, app)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeParseJsonFailed, err.Error()), nil)
		return
	}

	e = validateAppInfo(app)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	err = usage.ModifyApp(db, app)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeModifyApp, err.Error()), nil)
		return
	}


	JsonResult(w, http.StatusOK, nil, nil)
}

func RetrieveApp(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	//JsonResult(w, http.StatusOK, nil, appNewRelic)
	//return
	//
	// todo: auth
	
	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	// ...

	appId := params.ByName("id")

	e := validateAppID(appId)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	app, err := usage.RetrieveAppByID(db, appId)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGetApp, err.Error()), nil)
		return
	}

	JsonResult(w, http.StatusOK, nil, app)
}

func QueryAppList(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	//apps := []*usage.SaasApp{
	//	&appNewRelic,
	//}
	//
	//JsonResult(w, http.StatusOK, nil, newQueryListResult(int64(len(apps)), apps))
	//return

	// todo: auth
	
	// ...

	db := getDB()
	if db == nil {
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	r.ParseForm()

	provider, e := validateAppProvider(r.Form.Get("provider"), false)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}

	category, e := validateAppCategory(r.Form.Get("category"), false)
	if e != nil {
		JsonResult(w, http.StatusBadRequest, e, nil)
		return
	}
	
	offset, size := optionalOffsetAndSize(r, 30, 1, 100)
	orderBy := usage.ValidateOrderBy(r.Form.Get("orderby"))
	sortOrder := usage.ValidateSortOrder(r.Form.Get("sortorder"), false)

	count, apps, err := usage.QueryApps(db, provider, category, orderBy, sortOrder, offset, size)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeQueryApps, err.Error()), nil)
		return
	}

	JsonResult(w, http.StatusOK, nil, newQueryListResult(count, apps))
}










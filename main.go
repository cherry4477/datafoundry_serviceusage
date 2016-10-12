package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/asiainfoLDP/datafoundry_serviceusage/api"
	"github.com/asiainfoLDP/datahub_commons/httputil"
	"github.com/asiainfoLDP/datahub_commons/log"
)

var debug = flag.Bool("debug", false, "is debug mode?")
var port = flag.Int("port", 3000, "server port")

func init() {
	flag.Parse()
	api.Debug = *debug

	log_level := log.LevelString2Int(os.Getenv("LOG_LEVEL"))

	if log_level >= 0 {
		log.SetDefaultLoggerLevel(log_level)
	} else if *debug {
		log.SetDefaultLoggerLevel(log.LevelDebug)
	} else {
		log.SetDefaultLoggerLevel(log.LevelInfo)
	}

	log_name := fmt.Sprintf("%s-%s", os.Getenv("SERVICE_NAME"), os.Getenv("HOSTNAME"))
	log.SetDefaultLoggerName(log_name)
}

//=======================================================
//
//=======================================================

type Service struct {
	httpPort int
}

func newService(httpPort int) *Service {
	service := &Service{
		httpPort: httpPort,
	}

	return service
}

//=======================================================
//
//=======================================================

func handler_Index(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	api.JsonResult(w, http.StatusNotFound, api.GetError(api.ErrorCodeUrlNotSupported), nil)
}

func httpNotFound(w http.ResponseWriter, r *http.Request) {
	api.JsonResult(w, http.StatusNotFound, api.GetError(api.ErrorCodeUrlNotSupported), nil)
}

type HttpHandler struct {
	handler http.HandlerFunc
}

func (httpHandler *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if httpHandler.handler != nil {
		httpHandler.handler(w, r)
	}
}

//======================================================
//
//======================================================

func NewRouter() *httprouter.Router {
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	router.NotFound = &HttpHandler{httpNotFound}
	router.MethodNotAllowed = &HttpHandler{httpNotFound}
	//router.Handler ("GET", "/static", http.StripPrefix ("/static/", http.FileServer (http.Dir ("public"))))

	return router
}

//======================================================
//
//======================================================

func main() {
	router := NewRouter()

	// market

	if api.Init(router) == false {
		log.DefaultLogger().Fatal("failed to initdb")
	}

	// ...

	service := newService(*port)
	address := fmt.Sprintf(":%d", service.httpPort)
	log.DefaultLogger().Infof("Listening http at: %s\n", address)
	log.DefaultLogger().Fatal(http.ListenAndServe(address, httputil.TimeoutHandler(router, 10000*time.Millisecond, ""))) // will block here
}

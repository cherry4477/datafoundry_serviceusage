package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	//"fmt"
	//"log"
	//"io/ioutil"
	"bytes"
	"bufio"
	
	"github.com/julienschmidt/httprouter"
	
	//stars "github.com/asiainfoLDP/datahub_stars"
	//"github.com/asiainfoLDP/datahub_stars/api"
)

var router *httprouter.Router
func init() {
	router = NewRouter()
}

//==================================================================
//
//==================================================================

type HttpRequestCase struct {
	name         string
	requestInput string
	
	expectedStatusCode int
}

func newHttpRequestCase(caseName string, rInput string, expectedStatusCode int) *HttpRequestCase {
	a_case := &HttpRequestCase{
			name: caseName,
			requestInput: rInput,
			expectedStatusCode: expectedStatusCode,
		}
	
	return a_case
}

func _testCases(t *testing.T, cases []*HttpRequestCase) {
	for _, cs := range cases {
		r, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer([]byte(cs.requestInput))))
		if err != nil {
			t.Errorf("[%s] error: %s", cs.name, err.Error())
			continue
		}
		
		handler, params, _ := router.Lookup(r.Method, r.URL.EscapedPath())
		if handler == nil {
			t.Errorf("[%s] handler == nil", cs.name)
			continue
		}
		//if params == nil {
		//	t.Errorf("[%s] params == nil", cs.name)
		//	continue
		//}
		w := httptest.NewRecorder()
		handler(w, r, params)
		
		if w.Code != cs.expectedStatusCode {
			t.Errorf("[%s] w.Code (%d) != %d. \n======= response.Body ====== \n%s", cs.name, w.Code, cs.expectedStatusCode, string(w.Body.Bytes()))
		}
	}
}

//==================================================================
//
//==================================================================

// the ordr of cases are important!!!
var All_Cases = []*HttpRequestCase{
	
// ChangeStarStatus

newHttpRequestCase("1",
`GET / HTTP/1.1
Accept: application/json
User: zhang3@aaa.com

`, http.StatusNotFound),
}

func TestMain(t *testing.T) {
	_testCases(t, All_Cases)
}

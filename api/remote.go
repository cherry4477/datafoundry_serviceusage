package api

import (
	"os"
	"encoding/json"
	"net/http"
	"fmt"
	"time"
	"strings"

	"github.com/asiainfoLDP/datahub_commons/common"

	"github.com/asiainfoLDP/datafoundry_serviceusage/openshift"
	projectapi "github.com/openshift/origin/project/api/v1"
)

//======================================================
// remote end point
//======================================================

var (
	DataFoundryHost string

	PaymentService  string
	PlanService     string
	RechargeSercice string
)

func BuildServiceUrlPrefixFromEnv(name string, isHttps bool, addrEnv string, portEnv string) string {
	addr := os.Getenv(addrEnv)
	if addr == "" {
		Logger.Fatalf("%s env should not be null", addrEnv)
	}
	if portEnv != "" {
		port := os.Getenv(portEnv)
		if port != "" {
			addr += ":" + port
		}
	}

	prefix := ""
	if isHttps {
		prefix = fmt.Sprintf("https://%s", addr)
	} else {
		prefix = fmt.Sprintf("http://%s", addr)
	}

	Logger.Infof("%s = %s", name, prefix)
	
	return prefix
}


func initGateWay() {
	DataFoundryHost = BuildServiceUrlPrefixFromEnv("DataFoundryHost", true, "DATAFOUNDRY_HOST_ADDR", "")
	openshift.Init(DataFoundryHost, os.Getenv("DATAFOUNDRY_ADMIN_USER"), os.Getenv("DATAFOUNDRY_ADMIN_PASS"))

	PaymentService = BuildServiceUrlPrefixFromEnv("PaymentService", false, "DATAFOUNDRYPAYMENT_SERVICE_HOST", "DATAFOUNDRYPAYMENT_SERVICE_PORT")
	PlanService = BuildServiceUrlPrefixFromEnv("PlanService", false, "DATAFOUNDRYPLAN_SERVICE_HOST", "DATAFOUNDRYPLAN_SERVICE_PORT")
	RechargeSercice = BuildServiceUrlPrefixFromEnv("ChargeSercice", false, "DATAFOUNDRYRECHARGE_SERVICE_HOST", "DATAFOUNDRYRECHARGE_SERVICE_PORT")
}

//================================================================
// 
//================================================================

type ObjectMeta struct {
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
}

type User struct {
	ObjectMeta `json:"metadata,omitempty"`

	// FullName is the full name of user
	FullName string `json:"fullName,omitempty"`

	// Identities are the identities associated with this user
	Identities []string `json:"identities"`

	// Groups are the groups that this user is a member of
	Groups []string `json:"groups"`
}

/*
func authDF(token string) (*User, error) {
	url := fmt.Sprintf("%s/oapi/v1/users/~", DataFoundryHost)

	response, data, err := common.RemoteCall("GET", url, token, "")
	if err != nil {
		Logger.Debugf("authDF error: ", err.Error())
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		Logger.Debugf("remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return nil, fmt.Errorf("remote (%s) status code: %d.", url, response.StatusCode)
	}

	user := new(User)
	err = json.Unmarshal(data, user)
	if err != nil {
		Logger.Debugf("authDF Unmarshal error: %s. Data: %s\n", err.Error(), string(data))
		return nil, err
	}

	return user, nil
}
*/

func authDF(userToken string) (*User, error) {
	u := &User{}
	osRest := openshift.NewOpenshiftREST(openshift.NewOpenshiftClient(userToken))
	uri := "/users/~"
	osRest.OGet(uri, u)
	if osRest.Err != nil {
		Logger.Infof("authDF, uri(%s) error: %s", uri, osRest.Err)
		return nil, osRest.Err
	}

	return u, nil
}

func dfUser(user *User) string {
	return user.Name
}

func getDFUserame(token string) (string, error) {
	//Logger.Info("token = ", token)
	//if Debug {
	//	return "liuxu", nil
	//}

	user, err := authDF(token)
	if err != nil {
		return "", err
	}
	return dfUser(user), nil
}

//=======================================================================
// 
//=======================================================================

func getDFProject(usernameForLog, userToken, project string) (*projectapi.Project, error) {
	p := &projectapi.Project{}
	osRest := openshift.NewOpenshiftREST(openshift.NewOpenshiftClient(userToken))
	uri := "/projects/"+project
	osRest.OGet(uri, p)
	if osRest.Err != nil {
		Logger.Infof("user (%s) get df project (%s), uri(%s) error: %s", usernameForLog, project, uri, osRest.Err)
		return nil, osRest.Err
	}

	return p, nil
}

//=======================================================================
// 
//=======================================================================

// !!! plan types should NOT contains "_", see genOrderID for details,
const PLanType_Quota = "c"

const PLanCircle_Month = "m"

type Plan struct {
	Id              int64     `json:"id,omitempty"`
	Plan_id         string    `json:"plan_id,omitempty"`
	Plan_name       string    `json:"plan_name,omitempty"`
	Plan_type       string    `json:"plan_type,omitempty"`
	Plan_level      int       `json:"plan_level,omitempty"`
	Specification1  string    `json:"specification1,omitempty"`
	Specification2  string    `json:"specification2,omitempty"`
	Price           float32   `json:"price,omitempty"`
	Cycle           string    `json:"cycle,omitempty"`
	Region          string    `json:"region,omitempty"`
	Region_describe string    `json:"region_describe,omitempty"`
	Create_time     time.Time `json:"creation_time,omitempty"`
	Status          string    `json:"status,omitempty"`
}

// todo: add historyId to retrieve history info?
func getPlanByID(planId string) (*Plan, error) {
	if Debug {
		return &Plan{
			ID: 123,
			Plan_id: planId,
			Plan_name: "plan1",
			Plan_type: PLanType_Quota,
			Price: 12.3,
			Cycle: PLanCircle_Month,
			Region: "bj",
			Create_time: time.Date(2015, time.November, 10, 23, 0, 0, 0, time.UTC),
			Status: "",
		}, nil
	}

	url := fmt.Sprintf("%s/charge/v1/plans/%s", PlanService, planId)
	
	response, data, err := common.RemoteCall("GET", url, "", "")
	if err != nil {
		Logger.Infof("getPlan error: ", err.Error())
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		Logger.Infof("getPlan remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return nil, fmt.Errorf("getPlan remote (%s) status code: %d.", url, response.StatusCode)
	}

	plan := new(Plan)
	result := Result{Data: plan}
	err = json.Unmarshal(data, &result)
	if err != nil {
		Logger.Infof("getPlan Unmarshal error: %s. Data: %s\n", err.Error(), string(data))
		return nil, err
	}

	// ...

	plan.Cycle = strings.ToLower(plan.Cycle)
	plan.Plan_type = strings.ToLower(plan.Plan_type)

	// ...

	return plan, nil
}

//=======================================================================
// 
//=======================================================================

func makePayment(adminToken, accountId string, money float32, reason string) error {
	if Debug {
		return nil
	}

	body := fmt.Sprintf(
		`{"namespace":"%s","amount":%.3f,"reason":"%s"}`, 
		accountId, money, reason,
		)
	url := fmt.Sprintf("%s/charge/v1/recharge?type=deduction", RechargeSercice)
	
	response, data, err := common.RemoteCallWithJsonBody("POST", url, adminToken, "", []byte(body))
	if err != nil {
		Logger.Infof("makePayment error: ", err.Error())
		return err
	}

	if response.StatusCode != http.StatusOK {
		Logger.Infof("makePayment remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return fmt.Errorf("makePayment remote (%s) status code: %d.", url, response.StatusCode)
	}
	
	return nil
}

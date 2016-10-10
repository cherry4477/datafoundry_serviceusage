package api

import (
	"os"
	"encoding/json"
	"net/http"
	"fmt"
	"time"

	"github.com/asiainfoLDP/datahub_commons/common"

	"github.com/asiainfoLDP/datafoundry_serviceusage/openshift"
)

//======================================================
// remote end point
//======================================================

var (
	DataFoundryHost string

	PlanService   string
	ChargeSercice string
)

func BuildServiceUrlPrefixFromEnv(name string, addrEnv string) string {
	addr := os.Getenv(addrEnv)
	if addr == "" {
		Logger.Fatalf("%s env should not be null", addrEnv)
	}

	prefix := fmt.Sprintf("http://%s", addr)

	Logger.Infof("%s = %s", name, prefix)
	
	return prefix
}


func initGateWay() {
	DataFoundryHost = BuildServiceUrlPrefixFromEnv("DataFoundryHost", "DATAFOUNDRY_HOST_ADDR")
	openshift.Init(DataFoundryHost, os.Getenv("DATAFOUNDRY_ADMIN_USER"), os.Getenv("DATAFOUNDRY_ADMIN_PASS"))


	PlanService = BuildServiceUrlPrefixFromEnv("PlanService", "PLAN_SERVICE_API_SERVER")
	ChargeSercice = BuildServiceUrlPrefixFromEnv("ChargeSercice", "CHARGE_SERVICE_API_SERVER")
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

func authDF(token string) (*User, error) {
	url := fmt.Sprintf("%s/oapi/v1/users/~", DataFoundryHost)

	response, data, err := common.RemoteCall("GET", url, token, "")
	if err != nil {
		Logger.Debugf("authDF error: ", err.Error())
		return nil, err
	}

	// todo: use return code and msg instead
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

func dfUser(user *User) string {
	return user.Name
}

func getDFUserame(token string) (string, error) {
	//Logger.Info("token = ", token)

	user, err := authDF(token)
	if err != nil {
		return "", err
	}
	return dfUser(user), nil
}

//=======================================================================
// 
//=======================================================================

// todo: check user permission on project

//=======================================================================
// 
//=======================================================================

const PLanType_Quota = "C"

const PLanCircle_Month = "M"

type Plan struct {
	Plan_id        string    `json:"plan_id,omitempty"`
	Plan_name      string    `json:"plan_name,omitempty"`
	Plan_type      string    `json:"plan_type,omitempty"`
	Specification1 string    `json:"specification1,omitempty"`
	Specification2 string    `json:"specification2,omitempty"`
	Price          float32   `json:"price,omitempty"`
	Cycle          string    `json:"cycle,omitempty"`
	Region         string    `json:"region,omitempty"`
	Create_time    time.Time `json:"creation_time,omitempty"`
	Status         string    `json:"status,omitempty"`
}

// todo: retrieve plan

func getPlanByID(planId string) (*Plan, error) {
	if Debug {
		return &Plan{
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
		Logger.Debugf("getPlan error: ", err.Error())
		return nil, err
	}

	// todo: use return code and msg instead
	if response.StatusCode != http.StatusOK {
		Logger.Debugf("remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return nil, fmt.Errorf("remote (%s) status code: %d.", url, response.StatusCode)
	}

	plan := new(Plan)
	err = json.Unmarshal(data, plan)
	if err != nil {
		Logger.Debugf("authDF Unmarshal error: %s. Data: %s\n", err.Error(), string(data))
		return nil, err
	}

	return plan, nil
}

// todo: check if user can manage project (make payment)

//=======================================================================
// 
//=======================================================================

// todo: send consume money request

func makePayment(accountId string, money float64) error {
	if Debug {
		return nil
	}

	return fmt.Errorf("not implemented")
}

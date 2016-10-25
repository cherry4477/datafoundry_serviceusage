package api

import (
	"os"
	"encoding/json"
	"net/http"
	"fmt"
	"time"
	"strings"
	"errors"
	"strconv"

	"github.com/asiainfoLDP/datahub_commons/common"


	"github.com/asiainfoLDP/datafoundry_serviceusage/openshift"
	userapi "github.com/openshift/origin/pkg/user/api/v1"
	projectapi "github.com/openshift/origin/pkg/project/api/v1"
	kapi "k8s.io/kubernetes/pkg/api/v1"
	kapiresource "k8s.io/kubernetes/pkg/api/resource"
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

	PaymentService = BuildServiceUrlPrefixFromEnv("PaymentService", false, os.Getenv("ENV_NAME_DATAFOUNDRYPAYMENT_SERVICE_HOST"), os.Getenv("ENV_NAME_DATAFOUNDRYPAYMENT_SERVICE_PORT"))
	PlanService = BuildServiceUrlPrefixFromEnv("PlanService", false, os.Getenv("ENV_NAME_DATAFOUNDRYPLAN_SERVICE_HOST"), os.Getenv("ENV_NAME_DATAFOUNDRYPLAN_SERVICE_PORT"))
	RechargeSercice = BuildServiceUrlPrefixFromEnv("ChargeSercice", false, os.Getenv("ENV_NAME_DATAFOUNDRYRECHARGE_SERVICE_HOST"), os.Getenv("ENV_NAME_DATAFOUNDRYRECHARGE_SERVICE_PORT"))
}

//================================================================
// 
//================================================================

func authDF(userToken string) (*userapi.User, error) {
	if Debug {
		return &userapi.User{
			ObjectMeta: kapi.ObjectMeta {
				Name: "local",
			},
		}, nil
	}

	u := &userapi.User{}
	osRest := openshift.NewOpenshiftREST(openshift.NewOpenshiftClient(userToken))
	uri := "/users/~"
	osRest.OGet(uri, u)
	if osRest.Err != nil {
		Logger.Infof("authDF, uri(%s) error: %s", uri, osRest.Err)
		return nil, osRest.Err
	}

	return u, nil
}

func dfUser(user *userapi.User) string {
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

func getDfProject(usernameForLog, userToken, project string) (*projectapi.Project, error) {
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

//================================================================
// 
//================================================================

// The following identify resource constants for Kubernetes object types
const (
	// Pods, number
	ResourcePods kapi.ResourceName = "pods"
	// Services, number
	ResourceServices kapi.ResourceName = "services"
	// ReplicationControllers, number
	ResourceReplicationControllers kapi.ResourceName = "replicationcontrollers"
	// ResourceQuotas, number
	ResourceQuotas kapi.ResourceName = "resourcequotas"
	// ResourceSecrets, number
	ResourceSecrets kapi.ResourceName = "secrets"
	// ResourceConfigMaps, number
	ResourceConfigMaps kapi.ResourceName = "configmaps"
	// ResourcePersistentVolumeClaims, number
	ResourcePersistentVolumeClaims kapi.ResourceName = "persistentvolumeclaims"
	// ResourceServicesNodePorts, number
	ResourceServicesNodePorts kapi.ResourceName = "services.nodeports"
	// ResourceServicesLoadBalancers, number
	ResourceServicesLoadBalancers kapi.ResourceName = "services.loadbalancers"
	// CPU request, in cores. (500m = .5 cores)
	ResourceRequestsCPU kapi.ResourceName = "requests.cpu"
	// Memory request, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceRequestsMemory kapi.ResourceName = "requests.memory"
	// CPU limit, in cores. (500m = .5 cores)
	ResourceLimitsCPU kapi.ResourceName = "limits.cpu"
	// Memory limit, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceLimitsMemory kapi.ResourceName = "limits.memory"
)

const ProjectQuotaName = "quota"

func changeDfProjectQuota(usernameForLog, project string, plan *Plan) error {

	// ...

	cpus, mems, err := plan.ParsePlanQuotas()
	if err != nil {
		return err
	}

	const Gi = int64(1) << 30
	cpuQuantity := *kapiresource.NewQuantity(int64(cpus), kapiresource.DecimalExponent)
	memQuantity := *kapiresource.NewQuantity(int64(mems)*Gi, kapiresource.BinarySI)

	quota := kapi.ResourceQuota {}
	quota.Kind = "ResourceQuota"
	quota.APIVersion = "v1"
	quota.Name = ProjectQuotaName
	quota.Spec.Hard = kapi.ResourceList {
		ResourceLimitsCPU:      cpuQuantity,
		ResourceLimitsMemory:   memQuantity,
		ResourceRequestsCPU:    cpuQuantity,
		ResourceRequestsMemory: memQuantity,
	}
	
	// ...
	
	uri := "/namespaces/" + project + "/resourcequotas"

	osRest := openshift.NewOpenshiftREST(nil)

	// delete all quotas

	osRest.KDelete(uri, nil)
	if osRest.Err != nil {
		Logger.Warningf("delete quota (%s) error: %s", uri, osRest.Err)

		return osRest.Err
	}

	// create new one

	osRest.KPost(uri, &quota, nil)
	if osRest.Err != nil {
		Logger.Warningf("create quota (%s) error: %s", uri, osRest.Err)

		return osRest.Err
	}

	return nil
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

func (plan *Plan) ParsePlanQuotas() (int, int, error) {
	//"specification1": "16 CPU Cores",
	//"specification2": "32 GB RAM"，

	if plan.Plan_type != PLanType_Quota {
		return 0, 0, errors.New("not a quota plan")
	}
	
	var index int

	index = strings.Index(plan.Specification1, " CPU Cores")
	if index < 0 {
		return 0, 0, errors.New("invalid cpu format")
	}
	cpus, err := strconv.Atoi(plan.Specification1[:index])
	if err != nil || cpus < 0 {
		return 0, 0, errors.New("invalid cpu format.")
	}
	
	index = strings.Index(plan.Specification2, " GB RAM")
	if index < 0 {
		return 0, 0, errors.New("invalid memory format")
	}
	mems, err := strconv.Atoi(plan.Specification2[:index])
	if err != nil || mems < 0 {
		return 0, 0, errors.New("invalid memory format.")
	}

	return cpus, mems, nil
}

// todo: add historyId to retrieve history info?
func getPlanByID(planId string) (*Plan, error) {
	if Debug {
		return &Plan{
			Id: 123,
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
		Logger.Infof("getPlan error: %s", err.Error())
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

const ErrorCodeUpdateBalance = 1309

// the return bool means insufficient balance or not
func makePayment(adminToken, accountId string, money float32, reason, region string) (error, bool) {
	if Debug {
		return nil, false
	}

	body := fmt.Sprintf(
		`{"namespace":"%s","amount":%.3f,"reason":"%s", "region":"%s"}`, 
		accountId, money, reason, region, 
		)
	url := fmt.Sprintf("%s/charge/v1/recharge?type=deduction", RechargeSercice)
	
	response, data, err := common.RemoteCallWithJsonBody("POST", url, adminToken, "", []byte(body))
	if err != nil {
		Logger.Infof("makePayment error: %s", err.Error())
		return err, false
	}

	if response.StatusCode != http.StatusOK {
		insufficentData := false
		r := &Result{}
		if json.Unmarshal(data, &r) == nil {
			insufficentData = (r.Code == ErrorCodeUpdateBalance)
		}

		Logger.Infof("makePayment remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return fmt.Errorf("makePayment remote (%s) status code: %d.", url, response.StatusCode), insufficentData
	}
	
	return nil, false
}

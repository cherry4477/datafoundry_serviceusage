package api

import (
	"os"
	"encoding/json"
	"net/http"
	"fmt"
	"time"
	"strings"
	//"errors"
	"strconv"

	"github.com/asiainfoLDP/datahub_commons/common"


	"github.com/asiainfoLDP/datafoundry_serviceusage/openshift"
	userapi "github.com/openshift/origin/pkg/user/api/v1"
	projectapi "github.com/openshift/origin/pkg/project/api/v1"
	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api/v1"
	kapi "k8s.io/kubernetes/pkg/api/v1"
	kapiresource "k8s.io/kubernetes/pkg/api/resource"
)

//======================================================
// remote end point
//======================================================

const (
	DfRegion_CnNorth01 = "cn-north-1"
	DfRegion_CnNorth02 = "cn-north-2"

	NumDfRegions = 2

	DfRegion_Default = DfRegion_CnNorth01
)

var (
	//DataFoundryHost string
	osAdminClients map[string]*openshift.OpenshiftClient // region -> client
	VolumeServices map[string]string                     // region -> service

	PaymentService  string
	PlanService     string
	RechargeSercice string
)

func BuildDataFoundryClient(infoEnv string, durPhase time.Duration) (*openshift.OpenshiftClient, string) {
	info := os.Getenv(infoEnv)
	params := strings.Split(strings.TrimSpace(info), " ")
	if len(params) != 4 {
		Logger.Fatal("BuildDataFoundryClient, len(params) is not correct: ", len(params))
	}

	return openshift.CreateOpenshiftClient(infoEnv, params[0], params[1], params[2], durPhase),
		params[3]
}

func BuildServiceUrlPrefixFromEnv(name string, isHttps bool, address string, port string) string {
	if address == "" {
		Logger.Fatalf("%s: address should not be null", name)
	}
	if port != "" {
		address += ":" + port
	}

	prefix := ""
	if isHttps {
		prefix = fmt.Sprintf("https://%s", address)
	} else {
		prefix = fmt.Sprintf("http://%s", address)
	}

	Logger.Infof("%s = %s", name, prefix)
	
	return prefix
}

func initGateWay() {
	//DataFoundryHost = BuildServiceUrlPrefixFromEnv("DataFoundryHost", true, os.Getenv("DATAFOUNDRY_HOST_ADDR"), "")
	//openshift.Init(DataFoundryHost, os.Getenv("DATAFOUNDRY_ADMIN_USER"), os.Getenv("DATAFOUNDRY_ADMIN_PASS"))
	var durPhase time.Duration
	phaseStep := time.Hour / NumDfRegions

	// ...

	osAdminClients = make(map[string]*openshift.OpenshiftClient, NumDfRegions)
	VolumeServices = make(map[string]string, NumDfRegions)

	osAdminClients[DfRegion_CnNorth01], VolumeServices[DfRegion_CnNorth01] = 
		BuildDataFoundryClient("DATAFOUNDRY_INFO_CN_NORTH_1", durPhase)
	durPhase += phaseStep
	osAdminClients[DfRegion_CnNorth02], VolumeServices[DfRegion_CnNorth02] = 
		BuildDataFoundryClient("DATAFOUNDRY_INFO_CN_NORTH_2", durPhase)
	durPhase += phaseStep

	// ...

	// ...

	PaymentService = BuildServiceUrlPrefixFromEnv("PaymentService", false, os.Getenv(os.Getenv("ENV_NAME_DATAFOUNDRYPAYMENT_SERVICE_HOST")), os.Getenv(os.Getenv("ENV_NAME_DATAFOUNDRYPAYMENT_SERVICE_PORT")))
	PlanService = BuildServiceUrlPrefixFromEnv("PlanService", false, os.Getenv(os.Getenv("ENV_NAME_DATAFOUNDRYPLAN_SERVICE_HOST")), os.Getenv(os.Getenv("ENV_NAME_DATAFOUNDRYPLAN_SERVICE_PORT")))
	RechargeSercice = BuildServiceUrlPrefixFromEnv("ChargeSercice", false, os.Getenv(os.Getenv("ENV_NAME_DATAFOUNDRYRECHARGE_SERVICE_HOST")), os.Getenv(os.Getenv("ENV_NAME_DATAFOUNDRYRECHARGE_SERVICE_PORT")))
}

//================================================================
// 
//================================================================

func authDF(region, userToken string) (*userapi.User, error) {
	if Debug {
		return &userapi.User{
			ObjectMeta: kapi.ObjectMeta {
				Name: "local",
			},
		}, nil
	}

	u := &userapi.User{}
	//osRest := openshift.NewOpenshiftREST(openshift.NewOpenshiftClient(userToken))
	oc := osAdminClients[region]
	if oc == nil {
		return nil, fmt.Errorf("user noud found @ region (%s).")
	}
	oc = oc.NewOpenshiftClient(userToken)
	osRest := openshift.NewOpenshiftREST(oc)

	uri := "/users/~"
	osRest.OGet(uri, u)
	if osRest.Err != nil {
		Logger.Infof("authDF, region(%s), uri(%s) error: %s", region, uri, osRest.Err)
		//Logger.Infof("authDF, region(%s), token(%s), uri(%s) error: %s", region, userToken, uri, osRest.Err)
		return nil, osRest.Err
	}

	return u, nil
}

func dfUser(user *userapi.User) string {
	return user.Name
}

func getDFUserame(region, token string) (string, error) {
	//Logger.Info("token = ", token)
	//if Debug {
	//	return "liuxu", nil
	//}

	user, err := authDF(region, token)
	if err != nil {
		return "", err
	}
	return dfUser(user), nil
}

//=======================================================================
// 
//=======================================================================

func getDfProject(region, usernameForLog, userToken, project string) (*projectapi.Project, error) {
	p := &projectapi.Project{}
	
	//osRest := openshift.NewOpenshiftREST(openshift.NewOpenshiftClient(userToken))
	oc := osAdminClients[region]
	if oc == nil {
		return nil, fmt.Errorf("open shift client not found for region: %s", region)
	}
	oc = oc.NewOpenshiftClient(userToken)
	osRest := openshift.NewOpenshiftREST(oc)

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

const ProjectCpuMemoryQuotaName = "standard-quota"
const ProjectCpuMemoryLimitsName = "standard-limits"



func changeDfProjectQuotaWithPlan(usernameForLog, region, project string, plan *Plan) error {

	cpus, mems, err := plan.ParsePlanQuotas()
	if err != nil {
		return err
	}

	return changeDfProjectQuota(usernameForLog, region, project, cpus, mems)
}

func changeDfProjectQuota(usernameForLog, region, project string, cpus, mems int) error {

	oc := osAdminClients[region]
	if oc == nil {
		return fmt.Errorf("changeDfProjectQuota: open shift client not found for region: %s", region)
	}

	// ...

	const Mi = int64(1) << 20
	const Gi = int64(1) << 30

	cpuQuantity := *kapiresource.NewQuantity(int64(cpus), kapiresource.DecimalSI)
	memQuantity := *kapiresource.NewQuantity(int64(mems)*Gi, kapiresource.BinarySI)

	cpuQuantity_PodMin := *kapiresource.NewMilliQuantity(100, kapiresource.DecimalSI)
	memQuantity_PodMin := *kapiresource.NewQuantity(6*Mi, kapiresource.BinarySI)

	cpuQuantity_ContainerMin := *kapiresource.NewMilliQuantity(100, kapiresource.DecimalSI)
	memQuantity_ContainerMin := *kapiresource.NewQuantity(4*Mi, kapiresource.BinarySI)

	cpuQuantity_ContainerDefault := *kapiresource.NewMilliQuantity(100, kapiresource.DecimalSI)
	memQuantity_ContainerDefault := *kapiresource.NewQuantity(500*Mi, kapiresource.BinarySI)

	namespaceUri := "/namespaces/" + project

	// the new implementation: check existance, create on not found, update on found. 

	// set quotas
	{
		uri := namespaceUri + "/resourcequotas"
		fullUri := namespaceUri + "/resourcequotas/" + ProjectCpuMemoryQuotaName

		quota := kapi.ResourceQuota {}
		quota.Kind = "ResourceQuota"
		quota.APIVersion = "v1"
		quota.Name = ProjectCpuMemoryQuotaName
		quota.Spec.Hard = kapi.ResourceList {
			ResourceLimitsCPU:      cpuQuantity,
			ResourceLimitsMemory:   memQuantity,
			ResourceRequestsCPU:    cpuQuantity,
			ResourceRequestsMemory: memQuantity,
		}

		osRest := openshift.NewOpenshiftREST(oc)

		oldQuota := kapi.ResourceQuota {}
		osRest.KGet(fullUri, &oldQuota)
		if osRest.Err != nil {
			if osRest.StatusCode != 404 {
				Logger.Warningf("get quota (%s) error: %s", fullUri, osRest.Err)

				return osRest.Err
			}

			// create new
			osRest = openshift.NewOpenshiftREST(oc)
			osRest.KPost(uri, &quota, nil) 
			if osRest.Err != nil {
				Logger.Warningf("create quota (%s) error: %s", uri, osRest.Err)

				return osRest.Err
			}
		} else {
			// todo: if old and new are equal, do nothing

			// update quota
			osRest = openshift.NewOpenshiftREST(oc)
			osRest.KPut(fullUri, &quota, nil)
			if osRest.Err != nil {
				Logger.Warningf("update quota (%s) error: %s", fullUri, osRest.Err)

				return osRest.Err
			}
		}
	}

	// set limit
	{
		uri := namespaceUri + "/limitranges"
		fullUri := namespaceUri + "/limitranges/" + ProjectCpuMemoryLimitsName

		limit := kapi.LimitRange {}
		limit.Kind = "LimitRange"
		limit.APIVersion = "v1"
		limit.Name = ProjectCpuMemoryLimitsName
		limit.Spec.Limits = []kapi.LimitRangeItem {
			{
				Type: kapi.LimitTypePod,
				Max: kapi.ResourceList{
						kapi.ResourceCPU:    cpuQuantity,
						kapi.ResourceMemory: memQuantity,
					},
				Min: kapi.ResourceList{
						kapi.ResourceCPU:    cpuQuantity_PodMin,
						kapi.ResourceMemory: memQuantity_PodMin,
					},
			},
			{
				Type: kapi.LimitTypeContainer,
				Max: kapi.ResourceList{
						kapi.ResourceCPU:    cpuQuantity,
						kapi.ResourceMemory: memQuantity,
					},
				Min: kapi.ResourceList{
						kapi.ResourceCPU:    cpuQuantity_ContainerMin,
						kapi.ResourceMemory: memQuantity_ContainerMin,
					},
				Default: kapi.ResourceList{
						kapi.ResourceCPU:    cpuQuantity_ContainerDefault,
						kapi.ResourceMemory: memQuantity_ContainerDefault,
					},
				DefaultRequest: kapi.ResourceList{
						kapi.ResourceCPU:    cpuQuantity_ContainerDefault,
						kapi.ResourceMemory: memQuantity_ContainerDefault,
					},
			},
		}

		osRest := openshift.NewOpenshiftREST(oc)

		oldLimit := kapi.LimitRange {}
		osRest.KGet(fullUri, &oldLimit)
		if osRest.Err != nil {
			if osRest.StatusCode != 404 {
				Logger.Warningf("get limit (%s) error: %s", fullUri, osRest.Err)

				return osRest.Err
			}

			// create new
			osRest = openshift.NewOpenshiftREST(oc)
			osRest.KPost(uri, &limit, nil) 
			if osRest.Err != nil {
				Logger.Warningf("create limit (%s) error: %s", uri, osRest.Err)

				return osRest.Err
			}
		} else {
			// todo: if old and new are equal, do nothing

			// update limit
			osRest = openshift.NewOpenshiftREST(oc)
			osRest.KPut(fullUri, &limit, nil)
			if osRest.Err != nil {
				Logger.Warningf("update limit (%s) error: %s", fullUri, osRest.Err)

				return osRest.Err
			}
		}
	}

	// remove all quotas other than ProjectCpuMemoryQuotaName
	go func() {
		uri := namespaceUri + "/resourcequotas"

		quotaList := struct{
			Items []kapi.ResourceQuota `json:"items,omitempty"`
		}{
			[]kapi.ResourceQuota {},
		}

		osRest := openshift.NewOpenshiftREST(oc)

		osRest.KGet(uri, &quotaList)
		if osRest.Err != nil {
			Logger.Warningf("list quotas (%s) error: %s", uri, osRest.Err)
		} else {
			for _, quota := range quotaList.Items {
				if quota.Name == "" || quota.Name == ProjectCpuMemoryQuotaName {
					continue
				}
				
				fullUrl := uri + "/" + quota.Name
				osRest = openshift.NewOpenshiftREST(oc)
				osRest.KDelete(uri, nil)
				if osRest.Err != nil {
					Logger.Warningf("delete quota (%s) error: %s", fullUrl, osRest.Err)
				}
			}
		}
	}()

	// remove all limits other than ProjectCpuMemoryLimitsName
	go func() {
		uri := namespaceUri + "/limitranges"

		limitList := struct{
			Items []kapi.LimitRange `json:"items,omitempty"`
		}{
			[]kapi.LimitRange {},
		}

		osRest := openshift.NewOpenshiftREST(oc)

		osRest.KGet(uri, &limitList)
		if osRest.Err != nil {
			Logger.Warningf("list limits (%s) error: %s", uri, osRest.Err)
		} else {
			for _, limit := range limitList.Items {
				if limit.Name == "" || limit.Name == ProjectCpuMemoryLimitsName{
					continue
				}
				
				fullUrl := uri + "/" + limit.Name
				osRest = openshift.NewOpenshiftREST(oc)
				osRest.KDelete(uri, nil)
				if osRest.Err != nil {
					Logger.Warningf("delete limit (%s) error: %s", fullUrl, osRest.Err)
				}
			}
		}
	}()

	return nil
}

//=======================================================================
// 
//=======================================================================

const PLanCircle_Month = "m"

const MaxPlanTypeLength = 16
// !!! plan types should NOT contains "_", see genOrderID for details,
const PLanType_Quotas = "resources"
const PLanType_Volume = "volume"
const PLanType_BSI    = "bsi"

func isValidPlanType(planType string) bool {
	if len(planType) > MaxPlanTypeLength {
		return false
	}

	switch planType {
	case PLanType_Quotas:
		return true
	case PLanType_Volume:
		return true
	case PLanType_BSI:
		return true
	}

	return false
}

// for quotas plans, specification1 stores the cpu cores, specification2 stores the memory size
func (plan *Plan) ParsePlanQuotas() (int, int, error) {
	//"specification1": "16 CPU Cores",
	//"specification2": "32 GB RAM"，

	if plan.Plan_type != PLanType_Quotas {
		return 0, 0, fmt.Errorf("not a quota plan: %s", plan.Plan_type)
	}
	
	var index int

	index = strings.Index(plan.Specification1, " CPU Core")
	if index < 0 {
		return 0, 0, fmt.Errorf("invalid cpu format: %s", plan.Specification1)
	}
	cpus, err := strconv.Atoi(plan.Specification1[:index])
	if err != nil || cpus < 0 {
		return 0, 0, fmt.Errorf("invalid cpu format: %s", plan.Specification1)
	}
	
	index = strings.Index(plan.Specification2, " GB RAM")
	if index < 0 {
		return 0, 0, fmt.Errorf("invalid memory format: %s", plan.Specification2)
	}
	mems, err := strconv.Atoi(plan.Specification2[:index])
	if err != nil || mems < 0 {
		return 0, 0, fmt.Errorf("invalid memory format: %s", plan.Specification2)
	}

	return cpus, mems, nil
}

// for volume plans, specification1 stores the disk size
func (plan *Plan) ParsePlanVolume() (int, error) {
	//"specification1": "16 GB",

	if plan.Plan_type != PLanType_Volume {
		return 0, fmt.Errorf("not a volume plan: %s", plan.Plan_type)
	}
	
	var index int

	index = strings.Index(plan.Specification1, " GB")
	if index < 0 {
		return 0, fmt.Errorf("invalid volume format: %s", plan.Specification1)
	}
	vols, err := strconv.Atoi(plan.Specification1[:index])
	if err != nil || vols < 1 { // vols should be times of 10, here not check it carefully.
		return 0, fmt.Errorf("invalid volume format: %s", plan.Specification1)
	}

	return vols, nil
}

// todo:
// for bsi plans, 
func (plan *Plan) ParsePlanBSI() (string, string, error) {
	//"specification1": "16 GB",

	if plan.Plan_type != PLanType_BSI {
		return "", "", fmt.Errorf("not a bsi plan: %s", plan.Plan_type)
	}

	serviceName := plan.Specification1
	if serviceName == "" {
		return "", "", fmt.Errorf("service name is blank")
	}

	//planUUID := plan.Specification2
	planUUID := plan.Plan_id
	if planUUID == "" {
		return "", "", fmt.Errorf("plan uuid is blank")
	}

	return serviceName, planUUID, nil
}

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
			Id: 123,
			Plan_id: planId,
			Plan_name: "plan1",
			Plan_type: PLanType_Quotas,
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
//func makePayment(adminToken, accountId string, money float32, reason, region string) (error, bool) {
func makePayment(region, accountId string, money float32, reason string) (error, bool) {
	if Debug {
		return nil, false
	}

	oc := osAdminClients[region]
	if oc == nil {
		return fmt.Errorf("makePayment: open shift client not found for region: %s", region), false
	}
	adminToken := oc.BearerToken()

	body := fmt.Sprintf(
		`{"namespace":"%s","amount":%.3f,"reason":"%s", "region":"%s"}`, 
		accountId, money, reason, region, 
		)
	url := fmt.Sprintf("%s/charge/v1/recharge?type=deduction&region=%s", RechargeSercice, region)
	
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

//=======================================================================
// 
//=======================================================================

type VolumnCreateOptions struct {
	Name string     `json:"name,omitempty"`
	Size int        `json:"size,omitempty"`
	kapi.ObjectMeta `json:"metadata,omitempty"`
}

func createPersistentVolume(usernameForLog, volumeName, region, project string, plan *Plan) error {
	if Debug {
		return nil
	}

	// ...

	sizeGB, err := plan.ParsePlanVolume()
	if err != nil {
		return err
	}

	// ...

	oc := osAdminClients[region]
	if oc == nil {
		return fmt.Errorf("createPersistentVolume: open shift client not found for region: %s", region)
	}
	adminToken := oc.BearerToken()

	var vco = &VolumnCreateOptions{
		volumeName,
		sizeGB,
		kapi.ObjectMeta {
			Annotations: map[string]string {
				"dadafoundry.io/create-by": usernameForLog,
			},
		},
	}
	body, err := json.Marshal(vco)
	if err != nil {
		Logger.Infof("createPersistentVolume Marshal error: %s", err.Error())
		return err
	}

	volumeService := VolumeServices[region]
	if volumeService == "" {
		return fmt.Errorf("createPersistentVolume: volumeService not found for region: %s", region)
	}

	url := fmt.Sprintf("http://%s/lapi/v1/namespaces/%s/volumes", volumeService, project)

	response, data, err := common.RemoteCallWithJsonBody("POST", url, adminToken, "", []byte(body))
	if err != nil {
		Logger.Infof("createPersistentVolume error: %s", err.Error())
		return err
	}

	if response.StatusCode != http.StatusOK {
		Logger.Infof("createPersistentVolume remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return fmt.Errorf("createPersistentVolume remote (%s) status code: %d.", url, response.StatusCode)
	}

	return nil
}

func destroyPersistentVolume(volumeName, region, project string) error {
	if Debug {
		return nil
	}

	// ...

	oc := osAdminClients[region]
	if oc == nil {
		return fmt.Errorf("destroyPersistentVolume: open shift client not found for region: %s", region)
	}
	adminToken := oc.BearerToken()

	volumeService := VolumeServices[region]
	if volumeService == "" {
		return fmt.Errorf("destroyPersistentVolume: volumeService not found for region: %s", region)
	}

	url := fmt.Sprintf("http://%s/lapi/v1/namespaces/%s/volumes/%s", volumeService, project, volumeName)

	response, data, err := common.RemoteCallWithJsonBody("DELETE", url, adminToken, "", nil)
	if err != nil {
		Logger.Infof("destroyPersistentVolume error: %s", err.Error())
		return err
	}

	if response.StatusCode != http.StatusOK {
		Logger.Infof("destroyPersistentVolume remote (%s) status code: %d. data=%s", url, response.StatusCode, string(data))
		return fmt.Errorf("destroyPersistentVolume remote (%s) status code: %d.", url, response.StatusCode)
	}

	return nil
}

//=======================================================================
// 
//=======================================================================

func createBSI(usernameForLog, bsiName, region, project string, plan *Plan) error {
	if Debug {
		return nil
	}

	// ...

	serviceName, servicePlanUUID, err := plan.ParsePlanBSI()
	if err != nil {
		return err
	}

	// ...

	inputBSI := backingserviceinstanceapi.BackingServiceInstance {}
	inputBSI.Name = bsiName
	inputBSI.Spec.BackingServiceName =serviceName
	inputBSI.Spec.BackingServicePlanGuid = servicePlanUUID

	oc := osAdminClients[region]
	if oc == nil {
		return fmt.Errorf("createBSI: open shift client not found for region: %s", region)
	}
	uri := "/namespaces/"+project+"/backingserviceinstances"
	osRest := openshift.NewOpenshiftREST(oc)
	osRest.OPost(uri, &inputBSI, nil)
	if osRest.Err != nil {
		Logger.Infof("createBSI, region(%s), uri(%s) error: %s", region, uri, osRest.Err)
		return osRest.Err
	}

	return nil
}

func destroyBSI(bsiName, region, project string) error {
	if Debug {
		return nil
	}

	// todo: unbind all and deprovision

	return nil
}




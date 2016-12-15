package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

// BackingServiceInstance describe a BackingServiceInstance
type BackingServiceInstance struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the behavior of the Namespace.
	Spec BackingServiceInstanceSpec `json:"spec,omitempty" description:"spec defines the behavior of the BackingServiceInstance"`

	// Status describes the current status of a Namespace
	Status BackingServiceInstanceStatus `json:"status,omitempty" description:"status describes the current status of a Project; read-only"`
}

// BackingServiceInstanceList describe a list of BackingServiceInstance
type BackingServiceInstanceList struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	unversioned.ListMeta `json:"metadata,omitempty"`

	// Items is a list of routes
	Items []BackingServiceInstance `json:"items" description:"list of BackingServiceInstances"`
}

/*
type BackingServiceInstanceSpec struct {
	Config                 map[string]string `json:"config, omitempty"`
	InstanceID             string            `json:"instance_id, omitempty"`
	DashboardUrl           string            `json:"dashboard_url, omitempty"`
	BackingServiceName     string            `json:"backingservice_name, omitempty"`
	BackingServiceID       string            `json:"backingservice_id, omitempty"`
	BackingServicePlanGuid string            `json:"backingservice_plan_guid, omitempty"`
	Parameters             map[string]string `json:"parameters, omitempty"`
	Binding                bool              `json:"binding, omitempty"`
	BindUuid               string            `json:"bind_uuid, omitempty"`
	BindDeploymentConfig   map[string]string `json:"bind_deploymentconfig, omitempty"`
	Credential             map[string]string `json:"credential, omitempty"`
	Tags                   []string          `json:"tags, omitempty"`
}
*/

// BackingServiceInstanceSpec describes the attributes on a BackingServiceInstance
type BackingServiceInstanceSpec struct {
	// description of an instance.
	InstanceProvisioning `json:"provisioning, omitempty"`
	// description of an user-provided-service
	UserProvidedService `json:"userprovidedservice, omitempty"`
	// bindings of an instance
	Binding []InstanceBinding `json:"binding, omitempty"`
	// binding number of an instance
	Bound int `json:"bound, omitempty"`
	// id of an instance
	InstanceID string `json:"instance_id, omitempty"`
	// tags of an instance
	Tags []string `json:"tags, omitempty"`
}

/*
type InstanceProvisioning struct {
	DashboardUrl           string            `json:"dashboard_url, omitempty"`
	BackingService         string            `json:"backingservice, omitempty"`
	BackingServiceName     string            `json:"backingservice_name, omitempty"`
	BackingServiceID       string            `json:"backingservice_id, omitempty"`
	BackingServicePlanGuid string            `json:"backingservice_plan_guid, omitempty"`
	BackingServicePlanName string            `json:"backingservice_plan_name, omitempty"`
	Parameters             map[string]string `json:"parameters, omitempty"`
}
*/

// UserProvidedService describe an user-provided-service
type UserProvidedService struct{
	Credentials map[string]string `json:"credentials, omitempty"`
}

// InstanceProvisioning describe an InstanceProvisioning detail
type InstanceProvisioning struct {
	// dashboard url of an instance
	DashboardUrl string `json:"dashboard_url, omitempty"`
	// bs name of an instance
	BackingServiceName string `json:"backingservice_name, omitempty"`
	// bs id of an instance
	BackingServiceSpecID string `json:"backingservice_spec_id, omitempty"`
	// bs plan id of an instance
	BackingServicePlanGuid string `json:"backingservice_plan_guid, omitempty"`
	// bs plan name of an instance
	BackingServicePlanName string `json:"backingservice_plan_name, omitempty"`
	// parameters of an instance
	Parameters map[string]string `json:"parameters, omitempty"`
}

// InstanceBinding describe an instance binding.
type InstanceBinding struct {
	// bound time of an instance binding
	BoundTime *unversioned.Time `json:"bound_time,omitempty"`
	// bind uid of an instance binding
	BindUuid string `json:"bind_uuid, omitempty"`
	// deploymentconfig of an binding.
	BindDeploymentConfig string `json:"bind_deploymentconfig, omitempty"`
	// credentials of an instance binding
	Credentials map[string]string `json:"credentials, omitempty"`
}

// BackingServiceInstanceStatus describe the status of a BackingServiceInstance
type BackingServiceInstanceStatus struct {
	// phase is the current lifecycle phase of the instance
	Phase BackingServiceInstancePhase `json:"phase, omitempty"`
	// action is the action of the instance
	Action BackingServiceInstanceAction `json:"action, omitempty"`
	//last operation  of a instance provisioning
	LastOperation *LastOperation `json:"last_operation, omitempty"`
}

// LastOperation describe last operation of an instance provisioning
type LastOperation struct {
	// state of last operation
	State string `json:"state"`
	// description of last operation
	Description string `json:"description"`
	// async_poll_interval_seconds of a last operation
	AsyncPollIntervalSeconds int `json:"async_poll_interval_seconds, omitempty"`
}

type BackingServiceInstancePhase string
type BackingServiceInstanceAction string

const (
	BackingServiceInstancePhaseProvisioning BackingServiceInstancePhase = "Provisioning"
	BackingServiceInstancePhaseUnbound      BackingServiceInstancePhase = "Unbound"
	BackingServiceInstancePhaseBound        BackingServiceInstancePhase = "Bound"
	BackingServiceInstancePhaseDeleted      BackingServiceInstancePhase = "Deleted"

	BackingServiceInstanceActionToBind   BackingServiceInstanceAction = "_ToBind"
	BackingServiceInstanceActionToUnbind BackingServiceInstanceAction = "_ToUnbind"
	BackingServiceInstanceActionToDelete BackingServiceInstanceAction = "_ToDelete"

	BindDeploymentConfigBinding   string = "binding"
	BindDeploymentConfigUnbinding string = "unbinding"
	BindDeploymentConfigBound     string = "bound"

	UPS string = "USER-PROVIDED-SERVICE"
)

//=====================================================
//
//=====================================================

const BindKind_DeploymentConfig = "DeploymentConfig"

//type BindingRequest struct {
//	unversioned.TypeMeta
//	kapi.ObjectMeta
//
//	// the dc
//	DeploymentConfigName string `json:"deployment_name, omitempty"`
//}

// BindingRequestOptions describe a BindingRequestOptions.
type BindingRequestOptions struct {
	unversioned.TypeMeta `json:",inline"`
	// Standard object's metadata.
	kapi.ObjectMeta `json:"metadata,omitempty"`
	// bind kind is bindking of an instance binding
	BindKind string `json:"bindKind, omitempty"`
	// bindResourceVersion is bindResourceVersion of an instance binding.
	BindResourceVersion string `json:"bindResourceVersion, omitempty"`
	// resourceName of an instance binding
	ResourceName string `json:"resourceName, omitempty"`

}

func NewBindingRequestOptions(kind, version, name string) *BindingRequestOptions {
	return &BindingRequestOptions{
		BindKind:            kind,
		BindResourceVersion: version,
		ResourceName:        name,
	}
}

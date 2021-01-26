package model

type BindingRequest struct {
	Context string
	//*
	ServiceId string
	//*
	PlanId string
	AppGuid string
	BindResource BindResource
	Parameters interface{}
}

type BindingAcceptedDeleted struct {
	Operation string
}

type BindingCreated struct {
	Metadata BindingMetadata
	Credentials interface{}
	SyslogDrainUrl string
	RouteServiceUrl string
	VolumeMounts []VolumeMount
	Endpoints []Endpoint
}

type BindingOperationPollResponse struct {
	//*
	//Valid values are in progress, succeeded, and failed
	State string

	Description string
}

//request
type BindingRotation struct {
	PredecessorBindingId string
}

//response
type FetchingBindingResponse struct {
	Metadata []BindingMetadata
	Credentials interface{}
	SyslogDrainUrl string
	RouteServiceUrl string
	VolumeMounts []VolumeMount
	Parameters interface{}
	Endpoints []Endpoint
}

type BindResource struct {
	AppGuid string
	Route string
}

type BindingMetadata struct {
	ExpiresAt string
	RenewBefore string
}

type VolumeMount struct {
	Driver string
	ContainerDir string
	Mode string
	DeviceType string
	Device Device
}

type Device struct {
	VolumeId string
	MountConfig interface{}
}

type Endpoint struct {
	//*
	Host string
	//*
	Ports []string
	Protocol string
}
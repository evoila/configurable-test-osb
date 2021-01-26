package model

type CreateBindingRequest struct {
	Context string `json:"context"`
	//*
	ServiceId string `json:"service_id" binding:"required"`
	//*
	PlanId       string        `json:"plan_id" binding:"required"`
	AppGuid      string        `json:"app_guid"`
	BindResource *BindResource `json:"bind_resource"`
	Parameters   interface{}   `json:"parameters"`
}

type CreateRotateBindingAcceptedResponse struct {
	Operation string `json:"operation,omitempty"`
}

type CreateRotateBindingOkCreatedResponse struct {
	Metadata        *BindingMetadata `json:"metadata,omitempty"`
	Credentials     interface{}      `json:"credentials,omitempty"`
	SyslogDrainUrl  string           `json:"syslog_drain_url,omitempty"`
	RouteServiceUrl string           `json:"route_service_url,omitempty"`
	VolumeMounts    []*VolumeMount   `json:"volume_mounts,omitempty"`
	Endpoints       []*Endpoint      `json:"endpoints,omitempty"`
}

type BindingOperationPollResponse struct {
	//*
	//Valid values are in progress, succeeded, and failed
	State string `json:"state" binding:"required"`

	Description string `json:"description,omitempty"`
}

//request
type RotateBindingRequest struct {
	PredecessorBindingId string `json:"predecessor_binding_id" binding:"required"`
}

//response
type FetchingBindingResponse struct {
	Metadata        *BindingMetadata `json:"metadata,omitempty"`
	Credentials     interface{}      `json:"credentials,omitempty"`
	SyslogDrainUrl  string           `json:"syslog_drain_url,omitempty"`
	RouteServiceUrl string           `json:"route_service_url,omitempty"`
	VolumeMounts    []*VolumeMount   `json:"volume_mounts,omitempty"`
	Parameters      interface{}      `json:"parameters,omitempty"`
	Endpoints       []Endpoint       `json:"endpoints,omitempty"`
}

type BindResource struct {
	AppGuid string `json:"app_guid"`
	Route   string `json:"route"`
}

type BindingMetadata struct {
	ExpiresAt   string `json:"expires_at,omitempty"`
	RenewBefore string `json:"renew_before,omitempty"`
}

type VolumeMount struct {
	Driver       string  `json:"driver" binding:"required"`
	ContainerDir string  `json:"container_dir" binding:"required"`
	Mode         string  `json:"mode" binding:"required"`
	DeviceType   string  `json:"device_type" binding:"required"`
	Device       *Device `json:"device" binding:"required"`
}

type Device struct {
	VolumeId    string      `json:"volume_id" binding:"required"`
	MountConfig interface{} `json:"mount_config,omitempty"`
}

type Endpoint struct {
	//*
	Host string `json:"host" binding:"required"`
	//*
	Ports    []string `json:"ports" binding:"required"`
	Protocol string   `json:"protocol,omitempty"`
}

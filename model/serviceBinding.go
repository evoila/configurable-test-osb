package model

type CreateBindingRequest struct {
	Context *interface{} `json:"context"`
	//*
	ServiceID *string `json:"service_id" binding:"required"`
	//*
	PlanID       *string       `json:"plan_id" binding:"required"`
	AppGUID      *string       `json:"app_guid"`
	BindResource *BindResource `json:"bind_resource"`
	Parameters   *interface{}  `json:"parameters"`
}

/*
NOT NEEDED???!
the operation field could be in CreateRotateBindingResponse and just be omitted, if empty (and if 202 should
be returned, the other fields can be left empty, which will then be omitted)
*/
type CreateRotateBindingAcceptedResponse struct {
	Operation string `json:"operation,omitempty"`
}

//changed name, because this contains also the accepted response
type CreateRotateBindingResponse struct {
	Metadata        *BindingMetadata `json:"metadata,omitempty"`
	Credentials     *interface{}     `json:"credentials,omitempty"`
	SyslogDrainUrl  *string          `json:"syslog_drain_url,omitempty"`
	RouteServiceUrl *string          `json:"route_service_url,omitempty"`
	VolumeMounts    []*VolumeMount   `json:"volume_mounts,omitempty"`
	Endpoints       []*Endpoint      `json:"endpoints,omitempty"`
	Operation       *string          `json:"operation,omitempty"`
}

type BindingOperationPollResponse struct {
	//Valid values are in progress, succeeded, and failed
	//REQUIRED
	State string `json:"state"`

	Description string `json:"description,omitempty"`
}

type RotateBindingRequest struct {
	PredecessorBindingId *string      `json:"predecessor_binding_id" binding:"required"`
	Parameters           *interface{} `json:"parameters"`
}

type CreateRotateFetchBindingResponse struct {
	Metadata        *BindingMetadata `json:"metadata,omitempty"`
	Credentials     *interface{}     `json:"credentials,omitempty"`
	SyslogDrainUrl  *string          `json:"syslog_drain_url,omitempty"`
	RouteServiceUrl *string          `json:"route_service_url,omitempty"`
	VolumeMounts    *[]VolumeMount   `json:"volume_mounts,omitempty"`
	Parameters      *interface{}     `json:"parameters,omitempty"`
	Endpoints       *[]Endpoint      `json:"endpoints,omitempty"`
	Operation       *string          `json:"operation,omitempty"`
}

type BindResource struct {
	AppGuid string `json:"app_guid"`
	Route   string `json:"route"`
}

type BindingMetadata struct {
	ExpiresAt   *string `json:"expires_at,omitempty"`
	RenewBefore *string `json:"renew_before,omitempty"`
}

type VolumeMount struct {
	//REQUIRED
	Driver string `json:"driver"`
	//REQUIRED
	ContainerDir string `json:"container_dir"`
	//REQUIRED
	Mode string `json:"mode"`
	//REQUIRED
	DeviceType string `json:"device_type"`
	//REQUIRED
	Device *Device `json:"device"`
}

type Device struct {
	//REQUIRED
	VolumeId string `json:"volume_id"`

	MountConfig *interface{} `json:"mount_config,omitempty"`
}

type Endpoint struct {
	//REQUIRED
	Host string `json:"Host"`
	//REQUIRED
	Ports    []string `json:"ports"`
	Protocol *string  `json:"protocol,omitempty"`
}

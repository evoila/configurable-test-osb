package model

type CreateBindingRequest struct {
	Context      *interface{}  `json:"context"`
	ServiceID    *string       `json:"service_id" binding:"required"`
	PlanID       *string       `json:"plan_id" binding:"required"`
	AppGUID      *string       `json:"app_guid"`
	BindResource *BindResource `json:"bind_resource"`
	Parameters   *interface{}  `json:"parameters"`
}

type CreateRotateBindingAcceptedResponse struct {
	Operation string `json:"operation,omitempty"`
}

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
	State       string `json:"state"`
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
	Driver       string  `json:"driver"`
	ContainerDir string  `json:"container_dir"`
	Mode         string  `json:"mode"`
	DeviceType   string  `json:"device_type"`
	Device       *Device `json:"device"`
}

type Device struct {
	VolumeId    string       `json:"volume_id"`
	MountConfig *interface{} `json:"mount_config,omitempty"`
}

type Endpoint struct {
	Host     string   `json:"Host"`
	Ports    []string `json:"ports"`
	Protocol *string  `json:"protocol,omitempty"`
}

type DeleteRequest struct {
	Parameters *interface{} `json:"parameters"`
}

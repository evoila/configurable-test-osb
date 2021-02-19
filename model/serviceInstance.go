package model

//request (only request????)
type ProvideServiceInstanceRequest struct {
	ServiceID        string          `json:"service_id" binding:"required"`
	PlanID           string          `json:"plan_id" binding:"required"`
	Context          interface{}     `json:"context"`
	OrganizationGUID string          `json:"organization_guid" binding:"required"`
	SpaceGUID        string          `json:"space_guid" binding:"required"`
	Parameters       *interface{}    `json:"parameters"`
	MaintenanceInfo  MaintenanceInfo `json:"maintenance_info"`

	//RequestSettings RequestSettings `json:"request_settings"`
}

//Provision and Update have the same response form
type ProvideUpdateServiceInstanceResponse struct {
	DashboardUrl *string                  `json:"dashboard_url,omitempty"`
	Operation    *string                  `json:"operation,omitempty"`
	Metadata     *ServiceInstanceMetadata `json:"metadata,omitempty"`
}

func NewProvideServiceInstanceResponse(dashboardUrl *string, operation *string,
	metadata *ServiceInstanceMetadata, settings *Settings) *ProvideUpdateServiceInstanceResponse {
	response := ProvideUpdateServiceInstanceResponse{}
	if settings.ProvisionSettings.ShowDashboardURL {
		response.DashboardUrl = dashboardUrl
	}
	if settings.ProvisionSettings.ShowOperation {
		response.Operation = operation
	}
	if settings.ProvisionSettings.ShowMetadata {
		response.Metadata = metadata
	}
	return &response
}

func NewUpdateServiceInstanceResponse() {

}

type FetchingServiceInstanceResponse struct {
	ServiceId       string                   `json:"service_id,omitempty"`
	PlanId          string                   `json:"plan_id,omitempty"`
	DashboardUrl    *string                  `json:"dashboard_url,omitempty"`
	Parameters      *interface{}             `json:"parameters,omitempty"`
	MaintenanceInfo *MaintenanceInfo         `json:"maintenance_info,omitempty"`
	Metadata        *ServiceInstanceMetadata `json:"metadata,omitempty"`
}

type UpdateServiceInstanceRequest struct {
	Context         *interface{}     `json:"context"`
	ServiceId       *string          `json:"service_id" binding:"required"`
	PlanId          *string          `json:"plan_id"`
	Parameters      *interface{}     `json:"parameters"`
	PreviousValues  *PreviousValues  `json:"previous_values"`
	MaintenanceInfo *MaintenanceInfo `json:"maintenance_info"`
}

type PreviousValues struct {
	ServiceId       *string          `json:"service_id"`
	PlanId          *string          `json:"plan_id"`
	OrganizationId  *string          `json:"organization_id"`
	SpaceID         *string          `json:"space_id"`
	MaintenanceInfo *MaintenanceInfo `json:"maintenance_info"`
}

//SAME AS PROVISION RESPONSE
/*
type UpdateInstanceResponse struct {
	DashboardUrl string
	LastOperationID string
	Metadata ServiceInstanceMetadata
}
*/

type InstanceOperationPollResponse struct {
	//*
	//does binding:"required" count in both directions?
	//NO, BUT I CAN CHECK BEFORE CONVERTING TO JSON
	//REQUIRED
	State string `json:"state"`

	Description      string `json:"description,omitempty"`
	InstanceUsable   bool   `json:"instance_usable,omitempty"`
	UpdateRepeatable bool   `json:"update_repeatable,omitempty"`
}

type ServiceInstanceMetadata struct {
	Labels     interface{} `json:"labels,omitempty"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type DeleteServiceResponse struct {
	Operation string `json:"lastOperation,omitempty"`
}

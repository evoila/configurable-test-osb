package model

//request (only request????)
type Service struct {
	//NEEDED
	ServiceId string
	//NEEDER
	PlanId string
	Context interface{}
	//NEEDED
	OrganizationGuid string
	//NEEDER
	SpaceGuid string
	Parameters interface{}
	MaintenanceInfo MaintenanceInfo
}

type ProvisionResponse struct {
	DashboardUrl string
	Operation string
	Metadata ServiceInstanceMetadata
}

type FetchingServiceResponse struct {
	ServiceId string
	PlanId string
	DashboardUrl string
	Parameters interface{}
	MaintenanceInfo MaintenanceInfo
	Metadata ServiceInstanceMetadata
}

type UpdateInstanceRequest struct {
	Context interface{}
	PlanId string
	Parameters interface{}
	PreviousValues PreviousValues
	MaintenanceInfo MaintenanceInfo
}

type PreviousValues struct {
	ServiceId string
	PlanId string
	OrganizationId string
	SpaceId string
	MaintenanceInfo MaintenanceInfo
}

//SAME AS PROVISION RESPONSE
type UpdateInstanceResponse struct {
	DashboardUrl string
	Operation string
	Metadata ServiceInstanceMetadata
}

type InstanceOperationPollResponse struct {
	//*
	State string
	Description string
	InstanceUsable bool
	UpdateRepeatable bool
}

type ServiceInstanceMetadata struct {
	Labels interface{}
	Attributes interface{}
}

type ServiceInstanceDeleted struct {
	Operation string
}
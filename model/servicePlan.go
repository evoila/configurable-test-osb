package model

type ServicePlan struct {
	//*
	/*
		An identifier used to correlate this Service Plan in future requests to the Service Broker. This MUST be globally
		unique such that Platforms (and their users) MUST be able to assume that seeing the same value
		(no matter what Service Broker uses it) will always refer to this Service Plan and for the same Service Offering.
		MUST be a non-empty string. Using a GUID is RECOMMENDED.
	*/
	Id string

	//*
	/*
		The name of the Service Plan. MUST be unique within the Service Offering. MUST be a non-empty string.
		Using a CLI-friendly name is RECOMMENDED.
	*/
	Name string

	//*
	/*
		A short description of the Service Plan. MUST be a non-empty string.
	*/
	Description string

	Metadata interface{}
	Free bool
	Bindable bool
	BindingRotatable bool
	PlanUpdateable bool
	Schemas Schemas
	MaximumPollingDuration int
	MaintenanceInfo MaintenanceInfo
}

type Schemas struct {
	ServiceInstance ServiceInstanceSchema
	ServiceBinding ServiceBindingSchema
}

type ServiceInstanceSchema struct {
	Create InputParametersSchema
	Update InputParametersSchema
}

type ServiceBindingSchema struct {
	Create InputParametersSchema
}

type InputParametersSchema struct {
	//Parameters JSON schema object???
}

type MaintenanceInfo struct {
	//*
	/*
		This MUST be a string conforming to a semantic version 2.0. The Platform MAY use this field to determine
		whether a maintenance update is available for a Service Instance.
	*/
	Version string

	//*
	/*
		This SHOULD be a string describing the impact of the maintenance update, for example, important version changes,
		configuration changes, default value changes, etc. The Platform MAY present this information to the user before
		they trigger the maintenance update.
	*/
	Description string
}

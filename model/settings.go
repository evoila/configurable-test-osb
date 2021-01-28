package model

type CatalogSettings struct {
	//These are the settings for service offerings and the fields it uses

	/*
		The tag binding:"required" has no effect for decoding json files to structs but is here in case an endpoint
		will be created for providing catalogSettings in a request body
	*/
	//amount > 0
	Amount           int      `json:"amount" binding:"required"`
	TagsMin          int      `json:"tags_min" binding:"required"`
	TagsMax          int      `json:"tags_max" binding:"required"`
	Requires         []string `json:"requires" binding:"required"`
	RequiresMin      int      `json:"requires_min" binding:"required"`
	OfferingBindable string   `json:"offering_bindable" binding:"required"`

	//NEW
	InstancesRetrievableExists string `json:"instance_retrievable_exists" binding:"required"`
	//
	InstancesRetrievable string `json:"instances_retrievable" binding:"required"`
	//NEW
	BindingsRetrievableExists string `json:"bindings_retrievable_exists" binding:"required"`
	//
	BindingsRetrievable string `json:"bindings_retrievable" binding:"required"`
	//NEW
	AllowContextUpdatesExists string `json:"allow_context_updates_exists" binding:"required"`
	//
	AllowContextUpdates string `json:"allow_context_updates" binding:"required"`
	OfferingMetadata    string `json:"offering_metadata" binding:"required"`
	DashboardClient     string `json:"dashboard_client" binding:"required"`
	//NEW
	OfferingPlanUpdateableExists string `json:"offering_plan_updateable_exists" binding:"required"`
	//
	OfferingPlanUpdateable string `json:"offering_plan_updateable" binding:"required"`
	//PlansMin > 0
	PlansMin int `json:"plans_min" binding:"required"`
	//PlansMax >= PlansMin
	PlansMax     int    `json:"plans_max" binding:"required"`
	PlanMetadata string `json:"plan_metadata" binding:"required"`
	//NEW
	FreeExists string `json:"free_exists" binding:"required"`
	//
	Free string `json:"free" binding:"required"`
	//NEW
	PlanBindableExists string `json:"plan_bindable_exists" binding:"required"`
	//
	PlanBindable string `json:"plan_bindable" binding:"required"`
	//NEW
	BindingRotatableExists string `json:"binding_rotatable_exists" binding:"required"`
	//
	BindingRotatable string `json:"binding_rotatable" binding:"required"`
	//NEW
	PlanUpdateableExists string `json:"plan_updateable_exists" binding:"required"`
	//
	PlanUpdateable             string `json:"plan_updateable" binding:"required"`
	Schemas                    string `json:"schemas" binding:"required"`
	ServiceInstanceSchema      string `json:"service_instance_schema" binding:"required"`
	ServiceBindingSchema       string `json:"service_binding_schema" binding:"required"`
	MaxPollingDurationMin      int    `json:"max_polling_duration_min" binding:"required"`
	MaxPollingDurationMax      int    `json:"max_polling_duration_max" binding:"required"`
	MaintenanceInfo            string `json:"maintenance_info" binding:"required"`
	MaintenanceInfoVersion     string `json:"maintenance_info_version" binding:"required"`
	MaintenanceInfoDescription string `json:"maintenance_info_description" binding:"required"`
}

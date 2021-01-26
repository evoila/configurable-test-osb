package model

type ServiceOffering struct {
	//NEEDED
	Name string `json:"name" binding:"required"` //gin gedoens

	//NEEDED
	Id string `json:"id" binding:"required"`

	//NEEDED
	Description string `json:"description" binding:"required"`

	Tags     []string `json:"tags,omitempty"`
	Requires []string `json:"requires,omitempty"`

	//NEEDED
	Bindable bool `json:"bindable" binding:"required"`

	InstancesRetrievable bool        `json:"instances_retrievable,omitempty"`
	BindingsRetrievable  bool        `json:"bindings_retrievable,omitempty"`
	AllowContextUpdates  bool        `json:"allow_context_updates,omitempty"`
	Metadata             interface{} `json:"metadata,omitempty"`

	//A Cloud Foundry extension described in Catalog Extensions. Contains the data necessary to activate
	//the Dashboard SSO feature for this service.
	//currently in responseFields.go
	DashboardClient *DashboardClient `json:"dashboard_client,omitempty"`

	//misspelling kept by osbapi
	PlanUpdateable bool `json:"plan_updateable,omitempty"`
	//*
	Plans []ServicePlan `json:"plans" binding:"required"`
}

type DashboardClient struct {
	//*The id of the OAuth client that the dashboard will use. If present, MUST be a non-empty string.
	Id string `json:"id,omitempty"`

	//*A secret for the dashboard client. If present, MUST be a non-empty string.
	Secret string `json:"secret,omitempty"`

	//A URI for the service dashboard. Validated by the OAuth token server when the dashboard requests a token.
	RedirectUri string `json:"redirect_uri,omitempty"`
}

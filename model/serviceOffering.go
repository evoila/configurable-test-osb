package model

type ServiceOffering struct {
	//REQUIRED
	Name string `json:"name"`
	//REQUIRED
	Id string `json:"id"`
	//REQUIRED
	Description string `json:"description"`

	Tags     []string `json:"tags,omitempty"`
	Requires []string `json:"requires,omitempty"`

	//REQUIRED
	Bindable bool `json:"bindable"`

	InstancesRetrievable bool        `json:"instances_retrievable,omitempty"`
	BindingsRetrievable  bool        `json:"bindings_retrievable,omitempty"`
	AllowContextUpdates  bool        `json:"allow_context_updates,omitempty"`
	Metadata             interface{} `json:"metadata,omitempty"`

	//A Cloud Foundry extension described in Catalog Extensions. Contains the data necessary to activate
	//the Dashboard SSO feature for this service.
	DashboardClient *DashboardClient `json:"dashboard_client,omitempty"`

	//misspelling kept by osbapi
	PlanUpdateable bool `json:"plan_updateable,omitempty"`

	//REQUIRED
	Plans []ServicePlan `json:"plans"`
}

type DashboardClient struct {
	//*The id of the OAuth client that the dashboard will use. If present, MUST be a non-empty string.
	Id string `json:"id,omitempty"`

	//*A secret for the dashboard client. If present, MUST be a non-empty string.
	Secret string `json:"secret,omitempty"`

	//A URI for the service dashboard. Validated by the OAuth token server when the dashboard requests a token.
	RedirectUri string `json:"redirect_uri,omitempty"`
}

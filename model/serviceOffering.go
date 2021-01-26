package model

type ServiceOffering struct {
	//NEEDED
	Name string //gin gedoens

	//NEEDED
	Id string

	//NEEDED
	Description string

	Tags []string
	Requires []string

	//NEEDED
	Bindable bool

	InstancesRetrievable bool
	BindingsRetrievable bool
	AllowContextUpdates bool
	Metadata interface{}

	//A Cloud Foundry extension described in Catalog Extensions. Contains the data necessary to activate
	//the Dashboard SSO feature for this service.
	//currently in responseFields.go
	DashboardClient DashboardClient

	//misspelling kept by osbapi
	PlanUpdateable bool
	//*
	Plans []ServicePlan
}

type DashboardClient struct {
	//*The id of the OAuth client that the dashboard will use. If present, MUST be a non-empty string.
	Id string

	//*A secret for the dashboard client. If present, MUST be a non-empty string.
	Secret string

	//A URI for the service dashboard. Validated by the OAuth token server when the dashboard requests a token.
	RedirectUri string
}




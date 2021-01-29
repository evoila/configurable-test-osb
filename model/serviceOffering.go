package model

import "github.com/MaxFuhrich/serviceBrokerDummy/generator"

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
	Bindable *bool `json:"bindable"`

	InstancesRetrievable *bool       `json:"instances_retrievable,omitempty"`
	BindingsRetrievable  *bool       `json:"bindings_retrievable,omitempty"`
	AllowContextUpdates  *bool       `json:"allow_context_updates,omitempty"`
	Metadata             interface{} `json:"metadata,omitempty"`

	//A Cloud Foundry extension described in Catalog Extensions. Contains the data necessary to activate
	//the Dashboard SSO feature for this service.
	DashboardClient *DashboardClient `json:"dashboard_client,omitempty"`

	//misspelling kept by osbapi
	PlanUpdateable *bool `json:"plan_updateable,omitempty"`

	//REQUIRED
	Plans []ServicePlan `json:"plans"`
}

type DashboardClient struct {
	//*The id of the OAuth client that the dashboard will use. If present, MUST be a non-empty string.
	Id string `json:"id"`

	//*A secret for the dashboard client. If present, MUST be a non-empty string.
	Secret string `json:"secret"`

	//A URI for the service dashboard. Validated by the OAuth token server when the dashboard requests a token.
	RedirectUri *string `json:"redirect_uri,omitempty"`
}

func newServiceOffering(catalogSettings *CatalogSettings, tags []string) *ServiceOffering {
	offering := ServiceOffering{
		//MUST BE UNIQUE
		Name:        generator.RandomString(5),
		Id:          generator.RandomString(8) + "-XXXX-XXXX-XXXX-" + generator.RandomString(12),
		Description: generator.RandomString(6),
		Tags:        generator.SelectRandomTags(tags, catalogSettings.TagsMin, catalogSettings.TagsMax),
		Requires:    generator.RandomRequires(catalogSettings.Requires, catalogSettings.RequiresMin),
		Bindable:    generator.ReturnBoolean(catalogSettings.OfferingBindable),
		InstancesRetrievable: generator.ReturnFieldByBoolean(generator.ReturnBoolean(catalogSettings.InstancesRetrievableExists),
			catalogSettings.InstancesRetrievable), //returnBoolean(catalogSettings.InstancesRetrievable),
		BindingsRetrievable: generator.ReturnFieldByBoolean(generator.ReturnBoolean(catalogSettings.BindingsRetrievableExists),
			catalogSettings.BindingsRetrievable), //returnBoolean(catalogSettings.BindingsRetrievable),
		AllowContextUpdates: generator.ReturnFieldByBoolean(generator.ReturnBoolean(catalogSettings.AllowContextUpdatesExists),
			catalogSettings.AllowContextUpdates), //AllowContextUpdates: returnBoolean(catalogSettings.AllowContextUpdates),
		Metadata:        generator.MetadataByBool(generator.ReturnBoolean(catalogSettings.OfferingMetadata)),                                           //Metadata: metadataByBool(returnBoolean(catalogSettings.OfferingMetadata )),
		DashboardClient: generator.ReturnDashboardClient(catalogSettings),                                                                              //DashboardClient:
		PlanUpdateable:  generator.ReturnFieldByBoolean(generator.ReturnBoolean(catalogSettings.PlanUpdateableExists), catalogSettings.PlanUpdateable), //PlanUpdateable: returnBoolean(catalogSettings.PlanUpdateable),
		//Plans:
	}
	return &offering
}

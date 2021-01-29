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

func newServiceOffering(catalogSettings *CatalogSettings, tags []string) *ServiceOffering {
	offering := ServiceOffering{
		//MUST BE UNIQUE
		Name:        RandomString(5),
		Id:          RandomString(8) + "-XXXX-XXXX-XXXX-" + RandomString(12),
		Description: RandomString(6),
		Tags:        SelectRandomTags(tags, catalogSettings.TagsMin, catalogSettings.TagsMax),
		Requires:    RandomRequires(catalogSettings.Requires, catalogSettings.RequiresMin),
		Bindable:    ReturnBoolean(catalogSettings.OfferingBindable),
		InstancesRetrievable: ReturnFieldByBoolean(ReturnBoolean(catalogSettings.InstancesRetrievableExists),
			catalogSettings.InstancesRetrievable), //returnBoolean(catalogSettings.InstancesRetrievable),
		BindingsRetrievable: ReturnFieldByBoolean(ReturnBoolean(catalogSettings.BindingsRetrievableExists),
			catalogSettings.BindingsRetrievable), //returnBoolean(catalogSettings.BindingsRetrievable),
		AllowContextUpdates: ReturnFieldByBoolean(ReturnBoolean(catalogSettings.AllowContextUpdatesExists),
			catalogSettings.AllowContextUpdates), //AllowContextUpdates: returnBoolean(catalogSettings.AllowContextUpdates),
		Metadata:        MetadataByBool(ReturnBoolean(catalogSettings.OfferingMetadata)),                                           //Metadata: metadataByBool(returnBoolean(catalogSettings.OfferingMetadata )),
		DashboardClient: NewDashboardClient(catalogSettings),                                                                       //DashboardClient:
		PlanUpdateable:  ReturnFieldByBoolean(ReturnBoolean(catalogSettings.PlanUpdateableExists), catalogSettings.PlanUpdateable), //PlanUpdateable: returnBoolean(catalogSettings.PlanUpdateable),
		//Plans:
	}
	return &offering
}

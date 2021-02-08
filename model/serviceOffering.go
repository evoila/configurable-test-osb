package model

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/generator"
	"math/rand"
)

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

func newServiceOffering(catalogSettings *CatalogSettings, catalog *Catalog, tags []string) *ServiceOffering {
	offering := ServiceOffering{
		//MUST BE UNIQUE
		Name:        catalog.createUniqueName(5), //generator.RandomString(5),
		Id:          catalog.createUniqueId(),    //generator.RandomString(8) + "-XXXX-XXXX-XXXX-" + generator.RandomString(12),
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
		DashboardClient: NewDashboardClient(catalogSettings),                                                                                           //DashboardClient:
		PlanUpdateable:  generator.ReturnFieldByBoolean(generator.ReturnBoolean(catalogSettings.PlanUpdateableExists), catalogSettings.PlanUpdateable), //PlanUpdateable: returnBoolean(catalogSettings.PlanUpdateable),
		Plans:           makePlans(catalogSettings, catalog),
	}
	return &offering
}

func (catalog *Catalog) createUniqueName(n int) string {
	name := generator.RandomString(n)
	for catalog.nameExists(name) {
		name = generator.RandomString(n)
	}
	return name
}

func (catalog *Catalog) createUniqueId() string {
	id := generator.RandomString(8) + "-XXXX-XXXX-XXXX-" + generator.RandomString(12)
	for _, exists := catalog.GetServiceOfferingById(id); exists == true; {
		id = generator.RandomString(8) + "-XXXX-XXXX-XXXX-" + generator.RandomString(12)
	}
	return id
}

func makePlans(catalogSettings *CatalogSettings, catalog *Catalog) []ServicePlan {
	var servicePlans []ServicePlan
	numberOfPlans := rand.Intn(catalogSettings.PlansMax-catalogSettings.PlansMin+1) + catalogSettings.PlansMin
	for i := 0; i < numberOfPlans; i++ {
		//pointer ok?????
		servicePlans = append(servicePlans, *newServicePlan(catalogSettings, catalog))
	}
	return servicePlans
}

func (serviceOffering *ServiceOffering) GetPlanByID(planID string) (*ServicePlan, bool) {
	for _, plan := range serviceOffering.Plans {
		if planID == plan.ID {
			return &plan, true
		}
	}
	return nil, false
}

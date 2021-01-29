package generator

import "github.com/MaxFuhrich/serviceBrokerDummy/model"

func returnServiceOffering(catalogSettings *model.CatalogSettings, tags []string) *model.ServiceOffering {
	offering := model.ServiceOffering{
		//MUST BE UNIQUE
		Name:        RandomString(5),
		Id:          RandomString(8) + "-XXXX-XXXX-XXXX-" + RandomString(12),
		Description: RandomString(6),
		Tags:        SelectRandomTags(tags, catalogSettings.TagsMin, catalogSettings.TagsMax),
		Requires:    RandomRequires(catalogSettings.Requires, catalogSettings.RequiresMin),
		Bindable:    ReturnBoolean(catalogSettings.OfferingBindable),
		InstancesRetrievable: ReturnFieldByBoolean(ReturnBoolean(catalogSettings.InstancesRetrievableExists),
			catalogSettings.InstancesRetrievable), //ReturnBoolean(catalogSettings.InstancesRetrievable),
		BindingsRetrievable: ReturnFieldByBoolean(ReturnBoolean(catalogSettings.BindingsRetrievableExists),
			catalogSettings.BindingsRetrievable), //ReturnBoolean(catalogSettings.BindingsRetrievable),
		AllowContextUpdates: ReturnFieldByBoolean(ReturnBoolean(catalogSettings.AllowContextUpdatesExists),
			catalogSettings.AllowContextUpdates), //AllowContextUpdates: ReturnBoolean(catalogSettings.AllowContextUpdates),
		Metadata:        MetadataByBool(ReturnBoolean(catalogSettings.OfferingMetadata)),                                           //Metadata: MetadataByBool(ReturnBoolean(catalogSettings.OfferingMetadata )),
		DashboardClient: ReturnDashboardClient(catalogSettings),                                                                    //DashboardClient:
		PlanUpdateable:  ReturnFieldByBoolean(ReturnBoolean(catalogSettings.PlanUpdateableExists), catalogSettings.PlanUpdateable), //PlanUpdateable: ReturnBoolean(catalogSettings.PlanUpdateable),
		//Plans:
	}
	return &offering
}

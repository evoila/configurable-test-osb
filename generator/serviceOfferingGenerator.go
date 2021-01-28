package generator

import "github.com/MaxFuhrich/serviceBrokerDummy/model"

func returnServiceOffering(catalogSettings *model.CatalogSettings, tags []string) *model.ServiceOffering {
	offering := model.ServiceOffering{
		//MUST BE UNIQUE
		Name:        randomString(5),
		Id:          randomString(8) + "-XXXX-XXXX-XXXX-" + randomString(12),
		Description: randomString(6),
		Tags:        selectRandomTags(tags, catalogSettings.TagsMin, catalogSettings.TagsMax),
		Requires:    randomRequires(catalogSettings.Requires, catalogSettings.RequiresMin),
		Bindable:    returnBoolean(catalogSettings.OfferingBindable),
		InstancesRetrievable: returnFieldByBoolean(returnBoolean(catalogSettings.InstancesRetrievableExists),
			catalogSettings.InstancesRetrievable), //returnBoolean(catalogSettings.InstancesRetrievable),
		BindingsRetrievable: returnFieldByBoolean(returnBoolean(catalogSettings.BindingsRetrievableExists),
			catalogSettings.BindingsRetrievable), //returnBoolean(catalogSettings.BindingsRetrievable),
		AllowContextUpdates: returnFieldByBoolean(returnBoolean(catalogSettings.AllowContextUpdatesExists),
			catalogSettings.AllowContextUpdates), //AllowContextUpdates: returnBoolean(catalogSettings.AllowContextUpdates),
		Metadata:        metadataByBool(returnBoolean(catalogSettings.OfferingMetadata)),                                           //Metadata: metadataByBool(returnBoolean(catalogSettings.OfferingMetadata )),
		DashboardClient: returnDashboardClient(catalogSettings),                                                                    //DashboardClient:
		PlanUpdateable:  returnFieldByBoolean(returnBoolean(catalogSettings.PlanUpdateableExists), catalogSettings.PlanUpdateable), //PlanUpdateable: returnBoolean(catalogSettings.PlanUpdateable),
		//Plans:
	}
	return &offering
}

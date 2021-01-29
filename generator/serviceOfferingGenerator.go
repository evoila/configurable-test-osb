package generator

/*func returnServiceOffering(catalogSettings *model.CatalogSettings, tags []string) *model.ServiceOffering {
	offering := model.ServiceOffering{
		//MUST BE UNIQUE
		Name:        model.RandomString(5),
		Id:          model.RandomString(8) + "-XXXX-XXXX-XXXX-" + model.RandomString(12),
		Description: model.RandomString(6),
		Tags:        model.SelectRandomTags(tags, catalogSettings.TagsMin, catalogSettings.TagsMax),
		Requires:    model.RandomRequires(catalogSettings.Requires, catalogSettings.RequiresMin),
		Bindable:    model.ReturnBoolean(catalogSettings.OfferingBindable),
		InstancesRetrievable: model.ReturnFieldByBoolean(model.ReturnBoolean(catalogSettings.InstancesRetrievableExists),
			catalogSettings.InstancesRetrievable), //ReturnBoolean(catalogSettings.InstancesRetrievable),
		BindingsRetrievable: model.ReturnFieldByBoolean(model.ReturnBoolean(catalogSettings.BindingsRetrievableExists),
			catalogSettings.BindingsRetrievable), //ReturnBoolean(catalogSettings.BindingsRetrievable),
		AllowContextUpdates: model.ReturnFieldByBoolean(model.ReturnBoolean(catalogSettings.AllowContextUpdatesExists),
			catalogSettings.AllowContextUpdates),                                                                                               //AllowContextUpdates: ReturnBoolean(catalogSettings.AllowContextUpdates),
		Metadata:        model.MetadataByBool(model.ReturnBoolean(catalogSettings.OfferingMetadata)),                                           //Metadata: MetadataByBool(ReturnBoolean(catalogSettings.OfferingMetadata )),
		DashboardClient: ReturnDashboardClient(catalogSettings),                                                                                //DashboardClient:
		PlanUpdateable:  model.ReturnFieldByBoolean(model.ReturnBoolean(catalogSettings.PlanUpdateableExists), catalogSettings.PlanUpdateable), //PlanUpdateable: ReturnBoolean(catalogSettings.PlanUpdateable),
		//Plans:
	}
	return &offering
}


*/
//

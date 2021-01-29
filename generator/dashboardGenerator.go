package generator

import "github.com/MaxFuhrich/serviceBrokerDummy/model"

func ReturnDashboardClient(catalogSettings *model.CatalogSettings) *model.DashboardClient {
	var dashBoardClient model.DashboardClient
	if *ReturnBoolean(catalogSettings.DashboardClient) {
		dashBoardClient = model.DashboardClient{
			Id:          RandomString(8) + "-XXXX-XXXX-XXXX-" + RandomString(12),
			Secret:      RandomString(8) + "-XXXX-XXXX-XXXX-" + RandomString(12),
			RedirectUri: RandomUriByFrequency(catalogSettings.DashboardRedirectUri, 5),
		}
	}
	return &dashBoardClient
}

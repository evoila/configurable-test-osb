package generator

import "github.com/MaxFuhrich/serviceBrokerDummy/model"

func returnDashboardClient(catalogSettings *model.CatalogSettings) *model.DashboardClient {
	var dashBoardClient model.DashboardClient
	if *returnBoolean(catalogSettings.DashboardClient) {
		dashBoardClient = model.DashboardClient{
			Id:          randomString(8) + "-XXXX-XXXX-XXXX-" + randomString(12),
			Secret:      randomString(8) + "-XXXX-XXXX-XXXX-" + randomString(12),
			RedirectUri: randomUriByFrequency(catalogSettings.DashboardRedirectUri, 5),
		}
	}
	return &dashBoardClient
}

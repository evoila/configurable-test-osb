package model

import "github.com/MaxFuhrich/serviceBrokerDummy/generator"

type DashboardClient struct {
	Id          string  `json:"id"`
	Secret      string  `json:"secret"`
	RedirectUri *string `json:"redirect_uri,omitempty"`
}

func NewDashboardClient(catalogSettings *CatalogSettings) *DashboardClient {
	var dashBoardClient DashboardClient
	if *generator.ReturnBoolean(catalogSettings.DashboardClient) {
		dashBoardClient = DashboardClient{
			Id:          generator.RandomString(8) + "-XXXX-XXXX-XXXX-" + generator.RandomString(12),
			Secret:      generator.RandomString(8) + "-XXXX-XXXX-XXXX-" + generator.RandomString(12),
			RedirectUri: generator.RandomUriByFrequency(catalogSettings.DashboardRedirectUri, 5),
		}
	}
	return &dashBoardClient
}

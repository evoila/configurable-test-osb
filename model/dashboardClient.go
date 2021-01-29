package model

import "github.com/MaxFuhrich/serviceBrokerDummy/generator"

type DashboardClient struct {
	//*The id of the OAuth client that the dashboard will use. If present, MUST be a non-empty string.
	Id string `json:"id"`

	//*A secret for the dashboard client. If present, MUST be a non-empty string.
	Secret string `json:"secret"`

	//A URI for the service dashboard. Validated by the OAuth token server when the dashboard requests a token.
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

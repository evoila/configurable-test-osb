package model

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
	if *ReturnBoolean(catalogSettings.DashboardClient) {
		dashBoardClient = DashboardClient{
			Id:          RandomString(8) + "-XXXX-XXXX-XXXX-" + RandomString(12),
			Secret:      RandomString(8) + "-XXXX-XXXX-XXXX-" + RandomString(12),
			RedirectUri: RandomUriByFrequency(catalogSettings.DashboardRedirectUri, 5),
		}
	}
	return &dashBoardClient
}

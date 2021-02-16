package model

type Settings struct {
	GeneralSettings              GeneralSettings              `json:"general_settings" binding:"required"`
	HeaderSettings               HeaderSettings               `json:"header_settings" binding:"required"`
	ProvisionSettings            ProvisionSettings            `json:"provision_settings" binding:"required"`
	FetchServiceInstanceSettings FetchServiceInstanceSettings `json:"fetch_service_instance_settings" binding:"required"`
}

type HeaderSettings struct {
	BrokerVersion         string `json:"broker_version" binding:"required"`
	RejectWrongAPIVersion bool   `json:"reject_wrong_api_version" binding:"required"`
	RejectEmptyAPIVersion bool   `json:"reject_empty_api_version" binding:"required"`
	//NO ITS NOT UP TO THE BROKER TO DECIDE. THE PLATFORM MAY USE THIS SO ITS UP TO THE PLATFORM CHANGE?!
	OriginIDRequired              bool `json:"origin_id_required" binding:"required"`
	OriginIDValMustMatchProfile   bool `json:"origin_id_val_must_match_profile" binding:"required"`
	RequestIDRequired             bool `json:"request_id_required" binding:"required"`
	LogRequestID                  bool `json:"log_request_id" binding:"required"`
	RequestIDInResponse           bool `json:"request_id_in_response" binding:"required"`
	EtagIfModifiedSinceInResponse bool `json:"etag_if_modified_since_in_response" binding:"required"`
}

type GeneralSettings struct {
}

type ProvisionSettings struct {
	StatusCodeOK     bool `json:"status_code_ok" binding:"required"`
	Async            bool `json:"async" binding:"required"`
	DashboardURL     bool `json:"dashboard_url" binding:"required"`
	Operation        bool `json:"operation" binding:"required"`
	Metadata         bool `json:"metadata" binding:"required"`
	SecondsToFinish  int  `json:"seconds_to_finish" binding:"required"`
	ShowDashboardURL bool `json:"show_dashboard_url" binding:"required"`
	ShowOperation    bool `json:"show_operation" binding:"required"`
	ShowMetadata     bool `json:"show_metadata" binding:"required"`
}

type FetchServiceInstanceSettings struct {
	OfferingIDMustMatch bool `json:"offering_id_must_match" binding:"required"`
	PlanIDMustMatch     bool `json:"plan_id_must_match" binding:"required"`
	ShowServiceID       bool `json:"show_service_id" binding:"required"`
	ShowPlanID          bool `json:"show_plan_id" binding:"required"`
	ShowDashboardURL    bool `json:"show_dashboard_url" binding:"required"`
	ShowParameters      bool `json:"show_parameters" binding:"required"`
	ShowMaintenanceInfo bool `json:"show_maintenance_info" binding:"required"`
	ShowMetadata        bool `json:"show_metadata" binding:"required"`
}

package model

type Settings struct {
	GeneralSettings               GeneralSettings               `json:"general_settings" binding:"required"`
	HeaderSettings                HeaderSettings                `json:"header_settings" binding:"required"`
	ProvisionSettings             ProvisionSettings             `json:"provision_settings" binding:"required"`
	FetchServiceInstanceSettings  FetchServiceInstanceSettings  `json:"fetch_service_instance_settings" binding:"required"`
	PollInstanceOperationSettings PollInstanceOperationSettings `json:"poll_instance_operation_settings" binding:"required"`
	BindingSettings               BindingSettings               `json:"binding_settings" binding:"required"`
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
	StatusCodeOK                 bool `json:"status_code_ok" binding:"required"`
	Async                        bool `json:"async" binding:"required"`
	DashboardURL                 bool `json:"dashboard_url" binding:"required"`
	ReturnOperationIfAsync       bool `json:"return_operation_if_async" binding:"required"`
	Metadata                     bool `json:"metadata" binding:"required"`
	SecondsToFinish              int  `json:"seconds_to_finish" binding:"required"`
	ShowDashboardURL             bool `json:"show_dashboard_url" binding:"required"`
	ShowOperation                bool `json:"show_operation" binding:"required"`
	ShowMetadata                 bool `json:"show_metadata" binding:"required"`
	AllowDeprovisionWithBindings bool `json:"allow_deprovision_with_bindings" binding:"required"`
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

type PollInstanceOperationSettings struct {
	DescriptionInResponse                  bool `json:"description_in_response" binding:"required"`
	RetryPollInstanceOperationAfterSeconds int  `json:"retry_poll_instance_operation_after_seconds" binding:"required"`
}

type BindingSettings struct {
	ReturnBindingInformationOnce          bool                       `json:"return_binding_information_once" binding:"required"`
	ReturnOperationIfAsync                bool                       `json:"return_operation_if_async" binding:"required"`
	BindingMetadataSettings               BindingMetadataSettings    `json:"binding_metadata_settings" binding:"required"`
	ReturnCredentials                     bool                       `json:"return_credentials" binding:"required"`
	ReturnSyslogDrainURL                  bool                       `json:"return_syslog_drain_url" binding:"required"`
	ReturnRouteServiceURL                 bool                       `json:"return_route_service_url" binding:"required"`
	BindingVolumeMountSettings            BindingVolumeMountSettings `json:"binding_volume_mount_settings" binding:"required"`
	BindingEndpointSettings               BindingEndpointSettings    `json:"binding_endpoint_settings" binding:"required"`
	ReturnParameters                      bool                       `json:"return_parameters" binding:"required"`
	StatusCodeOK                          bool                       `json:"status_code_ok" binding:"required"`
	ReturnDescriptionLastOperation        bool                       `json:"return_description_last_operation" binding"required"`
	RetryPollBindingOperationAfterSeconds int                        `json:"retry_poll_binding_operation_after_seconds" binding:"required"`
}

type BindingMetadataSettings struct {
	ReturnMetadata    bool `json:"return_metadata" binding:"required"`
	ReturnExpiresAt   bool `json:"return_expires_at" binding:"required"`
	ReturnRenewBefore bool `json:"return_renew_before" binding:"required"`
}

type BindingVolumeMountSettings struct {
	ReturnVolumeMounts bool `json:"return_volume_mounts" binding:"required"`
	ReturnMountConfig  bool `json:"return_mount_config" binding:"required"`
}

type BindingEndpointSettings struct {
	ReturnEndpoints bool `json:"return_endpoints" binding:"required"`
	//???
	//the next two fields could be grouped by using *string and not requiring a binding
	//nil -> don't return protocol; value set -> return protocol
	//this is currently not done because of consistency with the rest of the settings
	ReturnProtocol bool   `json:"return_protocol" binding:"required"`
	ProtocolValue  string `json:"protocol_value" binding:"required"`
}

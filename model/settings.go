package model

type Settings struct {
	HeaderSettings HeaderSettings `json:"header_settings"`
}

type HeaderSettings struct {
	BrokerVersion                 string `json:"broker_version" binding:"required"`
	RejectWrongAPIVersion         bool   `json:"reject_wrong_api_version" binding:"required"`
	RejectEmptyAPIVersion         bool   `json:"reject_empty_api_version" binding:"required"`
	OriginIDValMustMatchProfile   bool   `json:"origin_id_val_must_match_profile"`
	LogRequestID                  bool   `json:"log_request_id"`
	RequestIDInResponse           bool   `json:"request_id_in_response"`
	EtagIfModifiedSinceInResponse bool   `json:"etag_if_modified_since_in_response"`
}

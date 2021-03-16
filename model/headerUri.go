package model

//REQUEST AND RESPONSE???
type Header struct {
	//omitempty has no effect, probably because `header...` and not `json`is used
	//SETTABLE
	APIVersionHeader *string `header:"X-Broker-API-Version"`

	OriginID *string `header:"X-Broker-API-Originating-Identity"`

	RequestID *string `header:"X-Broker-API-Request-Identity"`

	ETag *string `header:"ETag"`
	//ifmodifiedsince time in format Mon, 02 Jan 2006 15:04:05 MST -> RFC 1123
	IfModifiedSince *string `header:"If-Modified-Since"`
}

type UriProperties struct {
	ServiceId         string `form:"service_id" binding:"omitempty"`
	PlanId            string `form:"plan_id" binding:"omitempty"`
	InstanceId        string `form:"instance_id" binding:"omitempty"`
	BindingId         string `form:"binding_id" binding:"omitempty"`
	Operation         string `form:"operation" binding:"omitempty"`
	AcceptsIncomplete bool   `form:"accepts_incomplete" binding:"omitempty"`
}

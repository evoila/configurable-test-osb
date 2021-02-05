package model

/*
NOTE
Service Brokers and Platforms MAY support the ETag and If-Modified-Since HTTP headers to enable caching of the catalog.
(See RFC 7232 for details.)
Etag ressourcen identifizieren

*/

//STRUCTS FOR REQUEST FIELDS HERE

//REQUEST AND RESPONSE???
type Header struct {
	//omitempty has no effect, probably because `header...` and not `json`is used
	//SETTABLE
	APIVersionHeader *string `header:"X-Broker-API-Version"`

	/*
		-For any OSBAPI request that is the result of an action taken by a Platform's user,
		there MUST be an associated X-Broker-API-Originating-Identity header on that HTTP request.
		-Any OSBAPI request that is not associated with an action from a Platform's user, such as the
		Platform refetching the catalog, MAY exclude the header from that HTTP request.
		-If present on a request, the X-Broker-API-Originating-Identity header MUST contain the
		identify information for the Platform's user that took the action to cause the request to be sent.
		-Value Base64 encoded
		-Form X-Broker-API-Originating-Identity: Platform value(base64encoded)
		-ex X-Broker-API-Originating-Identity: kubernetes 1234
	*/
	OriginID *string `header:"X-Broker-API-Originating-Identity"`

	/*
		-Wenn vorhanden
		-For any OSBAPI request, there MUST be an associated X-Broker-API-Request-Identity header on the HTTP request.
		-The Service Broker MAY include this value in log messages generated as a result of the request.
		-The Service Broker SHOULD include this header in the response to the request.
		-value MUST be a non-empty string indicating the identity of the request being sent. The specific value MAY
		be unique for each request sent to the broker. Using a GUID is RECOMMENDED.
	*/
	RequestID *string `header:"X-Broker-API-Request-Identity"`

	//for "get catalog" responses
	ETag *string `header:"ETag"`
	//ifmodifiedsince time in format Mon, 02 Jan 2006 15:04:05 MST -> RFC 1123
	IfModifiedSince *string `header:"If-Modified-Since"`
}

type UriProperties struct {
	//Validators for binding
	//Values of type uppercase and lowercase letters, decimal digits, hyphen, period, underscore and tilde
	ServiceId  string `form:"service_id" binding:"omitempty"`
	PlanId     string `form:"plan_id" binding:"omitempty"`
	InstanceId string `form:"instance_id" binding:"omitempty"` //binding:"-"`
	BindingId  string `form:"binding_id" binding:"omitempty"`
	Operation  string `form:"operation" binding:"omitempty"`
	//Check later if place is still appropriate:
	AcceptsIncomplete bool `form:"accepts_incomplete" binding:"omitempty"`
}

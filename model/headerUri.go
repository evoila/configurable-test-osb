package model

//STRUCTS FOR REQUEST FIELDS HERE

type RequestHeader struct {
	APIVersionHeader string `header:"X-Broker-API-Version"`

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
	UserId string `header:"X-Broker-API-Originating-Identity"`

	/*
	-Wenn vorhanden
	-For any OSBAPI request, there MUST be an associated X-Broker-API-Request-Identity header on the HTTP request.
	-The Service Broker MAY include this value in log messages generated as a result of the request.
	-The Service Broker SHOULD include this header in the response to the request.
	-value MUST be a non-empty string indicating the identity of the request being sent. The specific value MAY
	be unique for each request sent to the broker. Using a GUID is RECOMMENDED.
	 */
	RequestId string `header:"X-Broker-API-Request-Identity "`
}

type UriParams struct {
	//Validators fuer binding
	//Values of type uppercase and lowercase letters, decimal digits, hyphen, period, underscore and tilde
	InstanceId string `uri:"instance_id" binding:"-"`
	BindingId	string `uri:"binding_id" binding:"-"`
	//Following ones needed? YES
	ServiceId string
	PlanId string
	Operation string
}
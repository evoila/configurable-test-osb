package model

//STRUCTS FOR RESPONSE FIELDS HERE

//Alternative: const, see errors.go
/* errors.go IN USE
var ErrorCodes = map[string]string{
	//Expected Action: The query parameter accepts_incomplete=true MUST be included the request.
	"AsyncRequired": "This request requires client support for asynchronous service operations.",

	//Expected Action: Clients MUST wait until pending requests have completed for the specified resources.
	"ConcurrencyError": "The Service Broker does not support concurrent requests that mutate the same resource.",

	//Expected Action: The app_guid MUST be included in the request body.
	"RequiresApp": "The request body is missing the app_guid field.",

	//Expected Action: The Platform SHOULD fetch the latest version of the Service Broker's Catalog.
	"MaintenanceInfoConflict": "The maintenance_info.version field provided in the request does not match the " +
		"maintenance_info.version field provided in the Service Broker's Catalog.",
}
 */














//CATALOG EXTENSIONS



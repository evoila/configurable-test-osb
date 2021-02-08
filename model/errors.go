package model

//Expected Action: The query parameter accepts_incomplete=true MUST be included the request.
const AsyncRequired = "This request requires client support for asynchronous service operations."

//Expected Action: Clients MUST wait until pending requests have completed for the specified resources.
const ConcurrencyError = "The Service Broker does not support concurrent requests that mutate the same resource."

//Expected Action: The app_guid MUST be included in the request body.
const RequiresApp = "The request body is missing the app_guid field."

//Expected Action: The Platform SHOULD fetch the latest version of the Service Broker's Catalog.
const MaintenanceInfoConflict = "The maintenance_info.version field provided in the request does not match the " +
	"maintenance_info.version field provided in the Service Broker's Catalog."

type ServiceBrokerError struct {
	Error            string `json:"error,omitempty"`
	Description      string `json:"description,omitempty"`
	InstanceUsable   string `json:"instance_usable,omitempty"`
	UpdateRepeatable string `json:"update_repeatable,omitempty"`
}

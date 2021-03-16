package model

const AsyncRequired = "This request requires client support for asynchronous service operations."

const ConcurrencyError = "The Service Broker does not support concurrent requests that mutate the same resource."

const RequiresApp = "The request body is missing the app_guid field."

const MaintenanceInfoConflict = "The maintenance_info.version field provided in the request does not match the " +
	"maintenance_info.version field provided in the Service Broker's Catalog."

type ServiceBrokerError struct {
	Error            string `json:"error,omitempty"`
	Description      string `json:"description,omitempty"`
	InstanceUsable   *bool  `json:"instance_usable,omitempty"`
	UpdateRepeatable *bool  `json:"update_repeatable,omitempty"`
}

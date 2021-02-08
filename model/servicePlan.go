package model

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/generator"
	"math/rand"
)

type ServicePlan struct {
	//*
	/*
		An identifier used to correlate this Service Plan in future requests to the Service Broker. This MUST be globally
		unique such that Platforms (and their users) MUST be able to assume that seeing the same value
		(no matter what Service Broker uses it) will always refer to this Service Plan and for the same Service Offering.
		MUST be a non-empty string. Using a GUID is RECOMMENDED.
	*/
	//REQUIRED
	ID string `json:"id"`

	//*
	/*
		The name of the Service Plan. MUST be unique within the Service Offering. MUST be a non-empty string.
		Using a CLI-friendly name is RECOMMENDED.
	*/
	//REQUIRED
	Name string `json:"name"`

	//*
	/*
		A short description of the Service Plan. MUST be a non-empty string.
	*/
	//REQUIRED
	Description string `json:"description"`

	Metadata               interface{}      `json:"metadata,omitempty"`
	Free                   *bool            `json:"free,omitempty"`
	Bindable               *bool            `json:"bindable,omitempty"`
	BindingRotatable       *bool            `json:"binding_rotatable,omitempty"`
	PlanUpdateable         *bool            `json:"plan_updateable,omitempty"`
	Schemas                *Schemas         `json:"schemas,omitempty"`
	MaximumPollingDuration int              `json:"maximum_polling_duration,omitempty"`
	MaintenanceInfo        *MaintenanceInfo `json:"maintenance_info,omitempty"`
}

type Schemas struct {
	ServiceInstance *ServiceInstanceSchema `json:"service_instance,omitempty"`
	ServiceBinding  *ServiceBindingSchema  `json:"service_binding,omitempty"`
}

type ServiceInstanceSchema struct {
	Create *InputParametersSchema `json:"create,omitempty"`
	Update *InputParametersSchema `json:"update,omitempty"`
}

type ServiceBindingSchema struct {
	Create *InputParametersSchema `json:"create,omitempty"`
}

type InputParametersSchema struct {
	//Parameters JSON schema object???
	Parameters interface{} `json:"parameters,omitempty"`
}

type MaintenanceInfo struct {
	//*
	/*
		This MUST be a string conforming to a semantic version 2.0. The Platform MAY use this field to determine
		whether a maintenance update is available for a Service Instance.
	*/
	Version *string `json:"version,omitempty"`

	//*
	/*
		This SHOULD be a string describing the impact of the maintenance update, for example, important version changes,
		configuration changes, default value changes, etc. The Platform MAY present this information to the user before
		they trigger the maintenance update.
	*/
	Description string `json:"description,omitempty"`
}

func newServicePlan(catalogSettings *CatalogSettings, catalog *Catalog) *ServicePlan {
	servicePlan := ServicePlan{
		ID:          catalog.createUniqueId(),
		Name:        catalog.createUniqueName(5),
		Description: generator.RandomString(8),
		Metadata:    generator.MetadataByBool(generator.ReturnBoolean(catalogSettings.PlanMetadata)),
		Free: generator.ReturnFieldByBoolean(generator.ReturnBoolean(catalogSettings.FreeExists),
			catalogSettings.Free),
		Bindable: generator.ReturnFieldByBoolean(generator.ReturnBoolean(catalogSettings.PlanBindableExists),
			catalogSettings.PlanBindable),
		BindingRotatable: generator.ReturnFieldByBoolean(generator.ReturnBoolean(catalogSettings.BindingRotatableExists),
			catalogSettings.BindingRotatable),
		PlanUpdateable: generator.ReturnFieldByBoolean(generator.ReturnBoolean(catalogSettings.PlanUpdateableExists),
			catalogSettings.PlanUpdateable),
		Schemas: newSchema(catalogSettings),
		MaximumPollingDuration: rand.Intn(catalogSettings.MaxPollingDurationMax-catalogSettings.MaxPollingDurationMin+1) +
			catalogSettings.MaxPollingDurationMin,
		MaintenanceInfo: nil,
	}
	return &servicePlan
}

//TO DO

func newSchema(catalogSettings *CatalogSettings) *Schemas {
	if *generator.ReturnBoolean(catalogSettings.Schemas) {
		schemas := Schemas{
			ServiceInstance: newServiceInstanceSchema(catalogSettings),
			ServiceBinding:  newServiceBindingSchema(catalogSettings),
		}
		return &schemas
	}
	return nil
}

func newServiceInstanceSchema(catalogSettings *CatalogSettings) *ServiceInstanceSchema {
	if *generator.ReturnBoolean(catalogSettings.ServiceInstanceSchema) {
		serviceInstanceSchema := ServiceInstanceSchema{
			Create: nil,
			Update: nil,
		}
		return &serviceInstanceSchema
	}
	return nil
}

func newServiceBindingSchema(catalogSettings *CatalogSettings) *ServiceBindingSchema {
	if *generator.ReturnBoolean(catalogSettings.ServiceBindingSchema) {
		serviceBindingSchema := ServiceBindingSchema{Create: nil}
		return &serviceBindingSchema
	}
	return nil
}

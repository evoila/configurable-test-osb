package model

import (
	"github.com/MaxFuhrich/configurable-test-osb/generator"
	"math/rand"
)

type ServicePlan struct {
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	Description            string           `json:"description"`
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
	Parameters interface{} `json:"parameters,omitempty"`
}

type MaintenanceInfo struct {
	Version     *string `json:"version,omitempty"`
	Description string  `json:"description,omitempty"`
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

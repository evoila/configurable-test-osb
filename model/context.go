package model

import (
	"encoding/json"
	"log"
)

type Context struct {
	Platform *string `json:"platform"`
}

type CloudFoundryContext struct {
	OrganizationGUID *string `json:"organization_guid"`
	SpaceGUID        *string `json:"space_guid"`
	//2.15
	OrganizationName *string `json:"organization_name"`
	SpaceName        *string `json:"space_name"`
	InstanceName     *string `json:"instance_name"`
	//2.16
	OrganizationAnnotations *interface{} `json:"organization_annotations"`
	SpaceAnnotations        *interface{} `json:"space_annotations"`
	InstanceAnnotations     *interface{} `json:"instance_annotations"`
}

type KubernetesContext struct {
	Namespace *string `json:"namespace"`
	ClusterID *string `json:"cluster_id"`
	//2.15
	InstanceName *string `json:"instance_name"`
	//2.16
	NamespaceAnnotations *interface{} `json:"namespace_annotations"`
	InstanceAnnotations  *interface{} `json:"instance_annotations"`
}

func CorrectContext(context *interface{}, brokerVersion *string, platform *string, binding bool) *ServiceBrokerError {
	b, err := json.Marshal(context)
	if err != nil {
		log.Println("Could not marshal context object: ", err)
		return &ServiceBrokerError{
			Error:       "MarshalError",
			Description: "Could not marshal context object",
		}
	}
	var contextPlatform Context
	err = json.Unmarshal(b, &contextPlatform)
	if err != nil {
		log.Println("Error while unmarshalling context object to context struct: ", err)
		return &ServiceBrokerError{
			Error:       "UnmarshalError",
			Description: "Could not unmarshal context object to context struct",
		}
	}
	if contextPlatform.Platform == nil {
		return &ServiceBrokerError{
			Error:       "PlatformMissing",
			Description: "Platform value must be provided if context is passed",
		}
	}
	if platform != nil && *platform != "" && *platform != *contextPlatform.Platform {
		return &ServiceBrokerError{
			Error:       "PlatformMismatch",
			Description: "Platform provided in X-Broker-API-Originating-Identity does not match platform in context",
		}
	}
	if *contextPlatform.Platform == "cloudfoundry" {
		return correctCloudfoundryContext(&b, brokerVersion, binding)
	}
	if *contextPlatform.Platform == "kubernetes" {
		return correctKubernetesContext(&b, brokerVersion, binding)
	}
	return nil
}

func correctKubernetesContext(contextJson *[]byte, brokerVersion *string, binding bool) *ServiceBrokerError {
	var kubernetesContext KubernetesContext
	err := json.Unmarshal(*contextJson, &kubernetesContext)
	if err != nil {
		return &ServiceBrokerError{
			Error:       "UnmarshalError",
			Description: "Could not unmarshal context bytes to KubernetesContext struct",
		}
	}
	if kubernetesContext.Namespace == nil || *kubernetesContext.Namespace == "" {
		return &ServiceBrokerError{
			Error:       "MissingValue",
			Description: "namespace must be provided and a non-empty string",
		}
	}
	if kubernetesContext.ClusterID == nil || *kubernetesContext.ClusterID == "" {
		return &ServiceBrokerError{
			Error:       "MissingValue",
			Description: "cluster_id must be provided and a non-empty string",
		}
	}
	if *brokerVersion > "2.14" {
		if !binding {
			if kubernetesContext.InstanceName == nil || *kubernetesContext.InstanceName == "" {
				return &ServiceBrokerError{
					Error:       "MissingValue",
					Description: "instance_name must be provided and a non-empty string",
				}
			}
		}
	}
	return nil
}

func correctCloudfoundryContext(contextJson *[]byte, brokerVersion *string, binding bool) *ServiceBrokerError {
	var cloudFoundryContext CloudFoundryContext
	err := json.Unmarshal(*contextJson, &cloudFoundryContext)
	if err != nil {
		return &ServiceBrokerError{
			Error:       "UnmarshalError",
			Description: "Could not unmarshal context bytes to CloudFoundryContext struct",
		}
	}
	if cloudFoundryContext.OrganizationGUID == nil || *cloudFoundryContext.OrganizationGUID == "" {
		return &ServiceBrokerError{
			Error:       "MissingValue",
			Description: "organization_guid must be provided and a non-empty string",
		}
	}
	if cloudFoundryContext.SpaceGUID == nil || *cloudFoundryContext.SpaceGUID == "" {
		return &ServiceBrokerError{
			Error:       "MissingValue",
			Description: "space_guid must be provided and a non-empty string",
		}
	}
	if *brokerVersion > "2.14" {
		if cloudFoundryContext.OrganizationName == nil || *cloudFoundryContext.OrganizationName == "" {
			return &ServiceBrokerError{
				Error:       "MissingValue",
				Description: "organization_name must be provided and a non-empty string",
			}
		}
		if cloudFoundryContext.SpaceName == nil || *cloudFoundryContext.SpaceName == "" {
			return &ServiceBrokerError{
				Error:       "MissingValue",
				Description: "space_name must be provided and a non-empty string",
			}
		}
		if !binding {
			if cloudFoundryContext.InstanceName == nil || *cloudFoundryContext.InstanceName == "" {
				return &ServiceBrokerError{
					Error:       "MissingValue",
					Description: "instance_name must be provided and a non-empty string",
				}
			}
		}
	}
	return nil
}

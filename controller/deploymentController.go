package controller

import (
	"github.com/evoila/configurable-test-osb/model"
	"github.com/evoila/configurable-test-osb/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

type DeploymentController struct {
	settings          *model.Settings
	deploymentService *service.DeploymentService
	platform          *string
}

//NewDeploymentController is the constructor for the struct DeploymentController which ensures, that the
//DeploymentController has access to the settings and the DeploymentService
func NewDeploymentController(deploymentService *service.DeploymentService, settings *model.Settings, platform *string) *DeploymentController {
	return &DeploymentController{
		settings:          settings,
		deploymentService: deploymentService,
		platform:          platform,
	}
}

//deploymentController.Provision is the handler for the "PUT /v2/service_instances/:instance_id" endpoint
//The request is bound here, checked if required parameters are empty and. checked if async is required. The request
//will then be passed to deploymentService.ProvideService(provisionRequest *model.ProvideServiceInstanceRequest,
//instanceID *string) which deploys the service and returns a response, which is used by deploymentController.Provision
func (deploymentController *DeploymentController) Provision(context *gin.Context) {
	instanceID := context.Param("instance_id")
	if instanceID == "" {
		context.JSON(412, gin.H{
			"message": "error while parsing url parameter \"instance_id\"",
			"error":   "invalid value, value must not be \"\" and unique",
		})
		return
	}
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(412, gin.H{
			"message": "error while parsing query parameter \"accepts_incomplete\"",
			"error":   "invalid value, value must be either \"true\", \"false\" or omitted which defaults to \"false\"",
		})
		return
	}
	var provisionRequest model.ProvideServiceInstanceRequest
	if err := context.ShouldBindJSON(&provisionRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while binding request body to struct",
			"error":   err.Error(),
		})
		return
	}
	if provisionRequest.Context != nil {
		err := model.CorrectContext(&provisionRequest.Context, &deploymentController.settings.HeaderSettings.BrokerVersion, deploymentController.platform, false)
		if err != nil {
			context.JSON(412, err)
			return
		}
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(provisionRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: model.AsyncRequired,
		})
		return
	}
	if provisionRequest.OrganizationGUID == nil || *provisionRequest.OrganizationGUID == "" {
		context.JSON(412, &model.ServiceBrokerError{
			Error:       "EmptyOrganizationGUID",
			Description: "organization_guid must be a non-empty string",
		})
		return
	}
	if provisionRequest.SpaceGUID == nil || *provisionRequest.SpaceGUID == "" {
		context.JSON(412, &model.ServiceBrokerError{
			Error:       "EmptySpaceGUID",
			Description: "space_guid must be a non-empty string",
		})
		return
	}
	statusCode, response, err := deploymentController.deploymentService.ProvideService(&provisionRequest, &instanceID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	if deploymentController.settings.HeaderSettings.BrokerVersion < "2.16" {

	}
	context.JSON(statusCode, response)
}

//deploymentController.FetchServiceInstance is the handler for the "GET /v2/service_instances/:instance_id" endpoint
//The request is bound here, checked if required parameters are empty and checked if async is required. The request
//will then be passed to deploymentService.FetchServiceInstance(instanceID *string, serviceID *string, planID *string)
//which deploys the service and returns a response, which is used by deploymentController.FetchServiceInstance
func (deploymentController *DeploymentController) FetchServiceInstance(context *gin.Context) {
	instanceID := context.Param("instance_id")
	var serviceID *string
	value, exists := context.GetQuery("service_id")
	if exists {
		if value == "" {
			context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
				Error:       "MalformedRequest",
				Description: "Query parameter \"service_id\" must not be an empty string (but can be omitted)",
			})
		} else {
			serviceID = &value
		}
	}
	var planID *string
	value, exists = context.GetQuery("plan_id")
	if exists {
		if value == "" {
			context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
				Error:       "MalformedRequest",
				Description: "Query parameter \"plan_id\" must not be an empty string (but can be omitted)",
			})
		} else {
			planID = &value
		}
	}
	statusCode, response, err := deploymentController.deploymentService.FetchServiceInstance(&instanceID, serviceID, planID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

//deploymentController.UpdateServiceInstance is the handler for the "PATCH /v2/service_instances/:instance_id" endpoint
//The request is bound here, checked if required parameters are empty and checked if async is required. The request
//will then be passed to deploymentService.UpdateServiceInstance(updateRequest *model.UpdateServiceInstanceRequest, instanceID *string)
//which updates the service and returns a response.
func (deploymentController *DeploymentController) UpdateServiceInstance(context *gin.Context) {
	instanceID := context.Param("instance_id")
	var updateRequest model.UpdateServiceInstanceRequest
	if err := context.ShouldBindJSON(&updateRequest); err != nil {
		context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
			Error:       "InvalidData",
			Description: "Error while binding request body to struct",
		})
		return
	}
	if deploymentController.settings.HeaderSettings.BrokerVersion < "2.15" && updateRequest.Context != nil {
		log.Println("broker version < 2.15, setting context in update request to nil")
		updateRequest.Context = nil
	}
	if updateRequest.Context != nil {
		err := model.CorrectContext(&updateRequest.Context, &deploymentController.settings.HeaderSettings.BrokerVersion, deploymentController.platform, false)
		if err != nil {
			context.JSON(400, err)
			return
		}
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(updateRequest.Parameters)
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
			Error:       "InvalidData",
			Description: "Invalid value for accepts_incomplete, value must be either \"true\", \"false\" or omitted which defaults to \"false\"",
		})
		return
	}
	if *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: model.AsyncRequired,
		})
		return
	}
	var header model.Header
	_ = context.ShouldBindHeader(&header)
	statusCode, response, err := deploymentController.deploymentService.UpdateServiceInstance(&updateRequest, &instanceID)
	if err != nil {
		if deploymentController.settings.HeaderSettings.BrokerVersion < "2.16" {
			err.InstanceUsable = nil
			err.UpdateRepeatable = nil
		}
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

//deploymentController.PollOperationState is the handler for the
//"GET /v2/service_instances/:instance_id/last_operation" endpoint.
//The request parameters are bound here which will then be passed to
//deploymentService.PollOperationState(instanceID *string, serviceID *string, planID *string, operationName *string)
//which polls the (last) operation of the service instance.
func (deploymentController *DeploymentController) PollOperationState(context *gin.Context) {
	instanceID := context.Param("instance_id")
	var serviceID *string
	valueServiceID, exists := context.GetQuery("service_id")
	if exists {
		if valueServiceID == "" {
			context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
				Error:       "MalformedRequest",
				Description: "Query parameter \"service_id\" must not be an empty string (but can be omitted)",
			})
		} else {
			serviceID = &valueServiceID
		}
	}
	var planID *string
	valuePlanID, exists := context.GetQuery("plan_id")
	if exists {
		if valuePlanID == "" {
			context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
				Error:       "MalformedRequest",
				Description: "Query parameter \"plan_id\" must not be an empty string (but can be omitted)",
			})
		} else {
			planID = &valuePlanID
		}
	}
	var operation *string
	valueOperation, exists := context.GetQuery("operation")
	if exists {
		if valueOperation == "" {
			context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
				Error:       "MalformedRequest",
				Description: "Query parameter \"operation\" must not be an empty string (but can be omitted)",
			})
		} else {
			operation = &valueOperation
		}
	}
	statusCode, response, err := deploymentController.deploymentService.PollOperationState(&instanceID, serviceID, planID, operation)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	if deploymentController.settings.HeaderSettings.BrokerVersion > "2.14" && response.State == model.PROGRESSING &&
		deploymentController.settings.PollInstanceOperationSettings.RetryPollInstanceOperationAfterSeconds > 0 {
		retryAfter := time.Second * time.Duration(deploymentController.settings.PollInstanceOperationSettings.RetryPollInstanceOperationAfterSeconds)
		context.Header("Retry-After", retryAfter.String())
	}
	context.JSON(statusCode, response)
}

//deploymentController.Delete is the handler for the "DELETE /v2/service_instances/:instance_id" endpoint.
//The request parameters are bound here which will then be passed to
//deploymentService.Delete(deleteRequest *model.DeleteRequest, instanceID *string, serviceID *string, planID *string)
//which removes the binding.
func (deploymentController *DeploymentController) Delete(context *gin.Context) {
	instanceID := context.Param("instance_id")
	serviceOfferingID, exists := context.GetQuery("service_id")
	if !exists {
		context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
			Error:       "MalformedRequest",
			Description: "service_id must be included as query parameter",
		})
		return
	}
	servicePlanID, exists := context.GetQuery("plan_id")
	if !exists {
		context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
			Error:       "MalformedRequest",
			Description: "plan_id must be included as query parameter",
		})
		return
	}
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing query parameter \"accepts_incomplete\"",
			"error":   "invalid value, value must be either \"true\", \"false\" or omitted which defaults to \"false\"",
		})
		return
	}
	var deleteRequest model.DeleteRequest
	_ = context.ShouldBindJSON(&deleteRequest)
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(deleteRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: model.AsyncRequired,
		})
		return
	}
	statusCode, response, err := deploymentController.deploymentService.Delete(&deleteRequest, &instanceID, &serviceOfferingID, &servicePlanID)
	if err != nil {
		if deploymentController.settings.HeaderSettings.BrokerVersion < "2.16" {
			err.InstanceUsable = nil
			err.UpdateRepeatable = nil
		}
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

//BONUS
func (deploymentController *DeploymentController) CurrentServiceInstances(context *gin.Context) {
	resp := struct {
		Instances *map[string]*model.ServiceDeployment `json:"instances"`
	}{}
	resp.Instances = deploymentController.deploymentService.CurrentServiceInstances()
	context.JSON(200, resp)
}

package controller

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"log"
	"net/http"
	"time"
)

type DeploymentController struct {
	settings          *model.Settings
	deploymentService *service.DeploymentService
}

//NewDeploymentController is the constructor for the struct DeploymentController which ensures, that the
//DeploymentController has access to the settings and the DeploymentService
func NewDeploymentController(deploymentService *service.DeploymentService, settings *model.Settings) *DeploymentController {
	return &DeploymentController{
		settings:          settings,
		deploymentService: deploymentService,
	}
}

//deploymentController.Provision is the handler for the "PUT /v2/service_instances/:instance_id" endpoint
//The request is bound here, checked if required parameters are empty and. checked if async is required. The request
//will then be passed to deployment.ProvideService(provisionRequest *model.ProvideServiceInstanceRequest,
//instanceID *string) which deploys the service and returns a response, which is used by deploymentController.Provision
func (deploymentController *DeploymentController) Provision(context *gin.Context) {
	instanceID := context.Param("instance_id")
	if instanceID == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing url parameter \"instance_id\"",
			"error":   "invalid value, value must not be \"\" and unique",
		})
		return
	}
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing query parameter \"accepts_incomplete\"",
			"error":   "invalid value, value must be either \"true\", \"false\" or omitted which defaults to \"false\"",
			//fmt.Println(acceptsIncomplete)
			//fmt.Println(instanceID)
			//statuscode must be returned by ProvideService too
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
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(provisionRequest.Parameters)

	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: "This Broker requires client support for asynchronous service operations.",
		})
		return
	}

	if provisionRequest.OrganizationGUID == "" {
		context.JSON(http.StatusBadRequest, &model.ServiceBrokerError{
			Error:       "EmptyOrganizationGUID",
			Description: "organization_guid must be a non-empty string",
		})
		return
	}
	if provisionRequest.SpaceGUID == "" {
		context.JSON(http.StatusBadRequest, &model.ServiceBrokerError{
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
	context.JSON(statusCode, response)
}

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

func (deploymentController *DeploymentController) UpdateServiceInstance(context *gin.Context) {
	//what happens if empty???
	instanceID := context.Param("instance_id")
	var updateRequest model.UpdateServiceInstanceRequest
	if err := context.ShouldBindJSON(&updateRequest); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while binding request body to struct",
			"error":   err.Error(),
		})
		return
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(updateRequest.Parameters)
	//GROUP IN ITS OWN FUNCTION???
	acceptsIncomplete := context.DefaultQuery("accepts_incomplete", "false")
	if acceptsIncomplete != "false" && acceptsIncomplete != "true" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing query parameter \"accepts_incomplete\"",
			"error":   "invalid value, value must be either \"true\", \"false\" or omitted which defaults to \"false\"",
		})
		return
	}
	//checking for nil not needed because asyncEndpoint gets default value = false
	//if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
	//fmt.Printf("accepts incomplete :%s\n", acceptsIncomplete)
	//fmt.Printf("asyncendpoint: %v\n", *requestSettings.AsyncEndpoint)
	if *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		//fmt.Println("accepts incomplete false")
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: "This Broker requires client support for asynchronous service operations.",
		})
		return
	}
	var header model.Header
	//error not assigned because this should already be checked by middleware
	_ = context.ShouldBindHeader(&header)
	statusCode, response, err := deploymentController.deploymentService.UpdateServiceInstance(&updateRequest, &instanceID, header.RequestID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	//fmt.Println("error was nil, response is:")
	//fmt.Println(response)
	context.JSON(statusCode, response)
}

func (deploymentController *DeploymentController) PollOperationState(context *gin.Context) {
	instanceID := context.Param("instance_id")
	var serviceID *string
	valueServiceID, exists := context.GetQuery("service_id")
	if exists {
		log.Printf("service_id valueServiceID: %v\n", valueServiceID)
		if valueServiceID == "" {
			context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
				Error:       "MalformedRequest",
				Description: "Query parameter \"service_id\" must not be an empty string (but can be omitted)",
			})
		} else {
			serviceID = &valueServiceID
			log.Printf("service_id assigned valueServiceID: %v\n", *serviceID)
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
	if response.State == model.PROGRESSING && deploymentController.settings.PollInstanceOperationSettings.RetryPollInstanceOperationAfterSeconds > 0 {
		retryAfter := time.Second * time.Duration(deploymentController.settings.PollInstanceOperationSettings.RetryPollInstanceOperationAfterSeconds)
		context.Header("Retry-After", retryAfter.String())
	}
	context.JSON(statusCode, response)
}

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
	_ = context.ShouldBindBodyWith(&deleteRequest, binding.JSON)
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(deleteRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: "This Broker requires client support for asynchronous service operations.",
		})
		return
	}
	statusCode, response, err := deploymentController.deploymentService.Delete(&deleteRequest, &instanceID, &serviceOfferingID, &servicePlanID)
	if err != nil {
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

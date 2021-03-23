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

type BindingController struct {
	settings       *model.Settings
	bindingService *service.BindingService
	platform       *string
}

func NewBindingController(bindingService *service.BindingService, settings *model.Settings, platform *string) *BindingController {
	return &BindingController{
		settings:       settings,
		bindingService: bindingService,
		platform:       platform,
	}
}

//bindingController.CreateBinding is the handler for the "PUT /v2/service_instances/:instance_id/service_bindings/:binding_id" endpoint
//The request is bound here, checked if required parameters are empty and checked if async is required.
//If the request can't be bound to the corresponding struct, the context will be passed to
//bindingController.rotateBinding(context *gin.Context) which has the same endpoint.
//The request will then be passed to bindingService.CreateBinding(bindingRequest *model.CreateBindingRequest, instanceID *string, bindingID *string)
//which creates a binding and returns a response, which is also used by deploymentController.FetchBinding.
func (bindingController *BindingController) CreateBinding(context *gin.Context) {
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")
	if bindingID == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing url parameter \"binding_id\"",
			"error":   "invalid value, value must not be \"\" and unique",
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
	var bindingRequest model.CreateBindingRequest
	//ShouldBindBodyWith instead of ShouldBindJSON used because ShouldBindBodyWith does not consume the JSON body
	if err := context.ShouldBindBodyWith(&bindingRequest, binding.JSON); err != nil {
		if bindingController.settings.HeaderSettings.BrokerVersion > "2.16" {
			//checking, if the request was a binding request and if so, rotate the bindings
			bindingController.rotateBinding(context)
			return
		}
		context.JSON(400, model.ServiceBrokerError{
			Error:       "MalformedRequest",
			Description: "Request has invalid fields",
		})
		return
	}
	if bindingRequest.Context != nil {
		err := model.CorrectContext(&bindingRequest.Context, &bindingController.settings.HeaderSettings.BrokerVersion, bindingController.platform, true)
		if err != nil {
			context.JSON(400, err)
			return
		}
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(bindingRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: model.AsyncRequired,
		})
		return
	}
	if bindingController.settings.BindingSettings.AppGUIDRequired && bindingRequest.AppGUID == nil {
		context.JSON(422, model.ServiceBrokerError{
			Error:       "RequiresApp",
			Description: model.RequiresApp,
		})
		return
	}
	if bindingRequest.AppGUID != nil && *bindingRequest.AppGUID == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid app_guid in request, value must not be \"\"",
			"error":   "InvalidValue",
		})
		return
	}
	statusCode, response, err := bindingController.bindingService.CreateBinding(&bindingRequest, &instanceID, &bindingID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

//bindingController.rotateBinding is the handler for the "PUT /v2/service_instances/:instance_id/service_bindings/:binding_id"
//endpoint in case it is a request to rotate a binding. If the request body can't be bound in CreateBinding, this function will be called
//The request is bound here, checked if required parameters are empty and checked if async is required.
//The request will then be passed to bindingService.RotateBinding(rotateBindingRequest *model.RotateBindingRequest, instanceID *string, bindingID *string)
//which creates a binding from an existing one and returns a response, which is also used by deploymentController.FetchBinding.
func (bindingController *BindingController) rotateBinding(context *gin.Context) {
	log.Println("rotateBinding called")
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")
	if bindingID == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "error while parsing url parameter \"binding_id\"",
			"error":   "invalid value, value must not be \"\" and unique",
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
	var rotateBindingRequest model.RotateBindingRequest
	if err := context.ShouldBindBodyWith(&rotateBindingRequest, binding.JSON); err != nil {
		context.JSON(http.StatusBadRequest, model.ServiceBrokerError{
			Error:       "InvalidRequest",
			Description: "Error while binding request body to struct, the body must either be the one specified for creating bindings or the one for rotating bindings",
		})
		return
	}
	var requestSettings *model.RequestSettings
	requestSettings, _ = model.GetRequestSettings(rotateBindingRequest.Parameters)
	if requestSettings.AsyncEndpoint != nil && *requestSettings.AsyncEndpoint && acceptsIncomplete == "false" {
		context.JSON(422, &model.ServiceBrokerError{
			Error:       "AsyncRequired",
			Description: model.AsyncRequired,
		})
		return
	}
	statusCode, response, err := bindingController.bindingService.RotateBinding(&rotateBindingRequest, &instanceID, &bindingID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

//bindingController.FetchBinding is the handler for the "GET /v2/service_instances/:instance_id/service_bindings/:binding_id" endpoint.
//The request parameters are bound here which will then be passed to
//bindingService.FetchBinding(instanceID *string, bindingID *string, serviceID *string, planID *string)
//which fetches the requested binding.
func (bindingController *BindingController) FetchBinding(context *gin.Context) {
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")
	serviceID := context.Query("service_id")
	planID := context.Query("plan_id")
	statusCode, response, err := bindingController.bindingService.FetchBinding(&instanceID, &bindingID, &serviceID, &planID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

//bindingController.PollOperationState is the handler for the
//"GET /v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation" endpoint.
//The request parameters are bound here which will then be passed to
//bindingService.PollOperationState(instanceID *string, bindingID *string, serviceID *string, planID *string, operationName *string)
//which polls the (last) operation of the binding.
func (bindingController *BindingController) PollOperationState(context *gin.Context) {
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")
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
	statusCode, response, err := bindingController.bindingService.PollOperationState(&instanceID, &bindingID, serviceID, planID, operation)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	if bindingController.settings.HeaderSettings.BrokerVersion > "2.14" && response.State == model.PROGRESSING && bindingController.settings.BindingSettings.RetryPollBindingOperationAfterSeconds > 0 {
		retryAfter := time.Second * time.Duration(bindingController.settings.BindingSettings.RetryPollBindingOperationAfterSeconds)
		context.Header("Retry-After", retryAfter.String())
	}
	context.JSON(statusCode, response)
}

//bindingController.Unbind is the handler for the
//"DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id" endpoint.
//The request parameters are bound here which will then be passed to
//bindingService.Unbind(deleteRequest *model.DeleteRequest, instanceID *string, bindingID *string, serviceID *string, planID *string)
//which removes the binding.
func (bindingController *BindingController) Unbind(context *gin.Context) {
	instanceID := context.Param("instance_id")
	bindingID := context.Param("binding_id")

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
			Description: model.AsyncRequired,
		})
		return
	}
	statusCode, response, err := bindingController.bindingService.Unbind(&deleteRequest, &instanceID, &bindingID, &serviceOfferingID, &servicePlanID)
	if err != nil {
		context.JSON(statusCode, err)
		return
	}
	context.JSON(statusCode, response)
}

//BONUS
func (bindingController *BindingController) CurrentBindings(context *gin.Context) {
	resp := struct {
		Bindings *map[string]*model.ServiceBinding `json:"service_bindings"`
	}{}
	resp.Bindings = bindingController.bindingService.CurrentBindings()
	//this won't show the fields inside a service instance, since its fields are not public (and not annotated)
	//only the names (with empty values) will be shown
	context.JSON(200, resp)
}

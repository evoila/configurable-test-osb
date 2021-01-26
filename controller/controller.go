package controller

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/generator"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Controller struct {
	catalog model.Catalog
}

func New() *Controller {
	var controller Controller
	catalogJson, err := os.Open("catalog/catalog.json")
	if err != nil {
		log.Println("Error while opening catalog file! error: " + err.Error())
	} else {
		byteVal, err := ioutil.ReadAll(catalogJson)
		if err != nil {
			log.Println("Error reading from catalog file! error: " + err.Error())
		} else {
			err = json.Unmarshal(byteVal, &controller.catalog)
			if err != nil {
				log.Println("Error unmarshalling the catalog file to the catalog struct! error: " + err.Error())
			}
		}
	}
	return &controller
}

//FUNCTIONS FOR HANDLING ENDPOINT REQUESTS GO HERE

func (controller *Controller) Hello(context *gin.Context) {
	context.JSON(http.StatusOK, controller.catalog)
}
func (controller *Controller) TestBind(context *gin.Context) {
	var offering model.ServiceOffering
	if err := context.ShouldBindJSON(&offering); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"Couldn't bind service offering! error": err.Error()})
		log.Println("Couldn't bind service offering! error: " + err.Error())
		return
	}
	context.JSON(http.StatusOK, offering)
}

/*func GetCatalog(context *gin.Context) {

}
/*
//GET /v2/service_instances/:instance_id/last_operation
func LastOpServiceInstance(context *gin.Context)  {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}

	//VORERST
	context.String(http.StatusOK, uriParams.InstanceId)
}
//GET /v2/service_instances/:instance_id/service_bindings/:binding_id/last_operation
func LastOpServiceBinding(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}

	//Vorerst

	fmt.Println(uriParams.InstanceId)
	fmt.Println("Help")
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//PUT /v2/service_instances/:instance_id
func ProvideService(context *gin.Context)  {

	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}

	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId)
}

//GET /v2/service_instances/:instance_id
func FetchServiceInstance(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId)
}

//PATCH /v2/service_instances/:instance_id
func UpdateServiceInstance(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId)
}

//PUT /v2/service_instances/:instance_id/service_bindings/:binding_id
func CreateServiceBinding(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//PUT /v2/service_instances/:instance_id/service_bindings/:binding_id
func RotateServiceBinding(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//GET /v2/service_instances/:instance_id/service_bindings/:binding_id
func FetchServiceBinding(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//DELETE /v2/service_instances/:instance_id/service_bindings/:binding_id
func Unbind(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId + " BindingID: " + uriParams.BindingId)
}

//DELETE /v2/service_instances/:instance_id
func Deprovide(context *gin.Context) {
	var uriParams model.UriParams
	if err := context.ShouldBindUri(&uriParams); err != nil {
		//Appropriate error response needed
		context.JSON(http.StatusBadRequest, gin.H{"message": err})
	}
	context.String(http.StatusOK, "InstanceID: " + uriParams.InstanceId)
}
*/

func (controller *Controller) GenerateCatalog(context *gin.Context) {
	//Generate new catalog according to settings
	newCatalog, err := generator.GenerateCatalog()
	if err != nil {
		log.Println("Unable to generate new catalog! error: " + err.Error())
	} else {
		controller.catalog = *newCatalog
	}
	context.String(http.StatusOK, "New catalog generated and loaded!")
}

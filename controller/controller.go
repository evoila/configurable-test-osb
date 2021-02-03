package controller

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Controller struct {
}

func New() *Controller {
	var controller Controller

	return &controller
}

//FUNCTIONS FOR HANDLING ENDPOINT REQUESTS GO HERE

func (controller *Controller) Hello(context *gin.Context) {
	context.JSON(http.StatusOK, "HI!")
	//fmt.Println("HI")
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

/*
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

/*func (controller *Controller) PrintCatalog()  {
	s, _ := json.MarshalIndent(controller.catalog, "", "\t");
	fmt.Print(string(s))
}

*/

//should header struct be returned?
func bindAndCheckHeader(context *gin.Context, settings *model.Settings) (*model.Header, error) {
	//is the bound header NEEDED by caller of this function?
	var header model.Header
	err := context.ShouldBindHeader(&header)
	if err != nil {
		//only return or already response here?
	}
	/*
		if no checks are done this is not needed
		userId := strings.Split(header.UserId, " ")

		for _, val := range userId {
			fmt.Println(val)
		}

	*/

	s, _ := json.MarshalIndent(header, "", "\t")
	log.Println(string(s))
	log.Println("Header settings:")
	s, _ = json.MarshalIndent(settings, "", "\t")
	log.Println(string(s))
	return &header, nil
}

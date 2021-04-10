package controller

import (
	"github.com/evoila/configurable-test-osb/model"
	"github.com/evoila/configurable-test-osb/service"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type CatalogController struct {
	catalogService *service.CatalogService
	settings       *model.Settings
}

func NewCatalogController(catalogService *service.CatalogService, settings *model.Settings) CatalogController {
	return CatalogController{
		catalogService: catalogService,
		settings:       settings,
	}
}

//GetCatalog is the handler for the endpoint "GET /v2/catalog".
//catalogService.GetCatalog() is called, which returns the catalog.
func (catalogController *CatalogController) GetCatalog(context *gin.Context) {
	context.JSON(http.StatusOK, catalogController.catalogService.GetCatalog())
}

//GenerateCatalog is the handler for the custom endpoint "GET /v2/catalog/generate".
//model.NewCatalog(catalogSettings *CatalogSettings) is called, which generates a new randomized catalog.
//The new catalog will be put in the response body.
//The new catalog does NOT replace the one used by the service broker.
func (catalogController *CatalogController) GenerateCatalog(context *gin.Context) {
	settings, err := model.NewCatalogSettings()
	if err != nil {
		context.JSON(500, err)
		log.Println("There has been an error while creating the catalog generator settings!", err.Error())
		log.Printf("In order to work a valid catalog generator settings file has to be provided. \nThe path to " +
			"the catalog generator settings file can either be set \n- by setting the environment variable " +
			"CATALOG_GENERATOR_FILE_PATH=\"PathToSettings\"+\"Filename\".json\n" +
			"- by putting a catalogSettings.json file in the directory from where this program is executed or in a subdirectory named \"config\".")
		return
	}
	catalog := model.NewCatalog(settings)

	context.JSON(http.StatusOK, catalog)
}

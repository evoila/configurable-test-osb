package controller

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
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

func (catalogController *CatalogController) GetCatalog(context *gin.Context) {
	context.JSON(http.StatusOK, catalogController.catalogService.GetCatalog())
}

func (catalogController *CatalogController) GenerateCatalog(context *gin.Context) {
	s, _ := os.Open("config/catalogSettings.json")
	byteVal, _ := ioutil.ReadAll(s)
	var settings model.CatalogSettings
	_ = json.Unmarshal(byteVal, &settings)
	catalog, _ := model.NewCatalog(&settings)
	context.JSON(http.StatusOK, catalog)
}

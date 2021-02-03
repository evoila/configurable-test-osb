package service

import (
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type CatalogService struct {
	catalog         *model.Catalog
	catalogSettings *model.CatalogSettings
}

func NewCatalogService(catalog *model.Catalog) CatalogService {
	return CatalogService{catalog: catalog}
}

func (catalogService *CatalogService) GetCatalog() *model.Catalog {
	return catalogService.catalog
}

func (catalogService *CatalogService) GenerateCatalog(context *gin.Context) {
	//Generate new catalog according to settings
	catalogService.catalogSettings, _ = model.NewCatalogSettings()
	catalog, err := model.NewCatalog(catalogService.catalogSettings) //generator.GenerateCatalog()
	//newCatalog, err := generator.GenerateCatalog()
	if err != nil {
		log.Println("Unable to load settings! error: " + err.Error())
	} else {
		catalogService.catalog = catalog
		catalogService.logCatalog()
	}
	if context != nil {
		context.JSON(http.StatusOK, catalog)
	}

}

func (catalogService *CatalogService) logCatalog() {
	s, _ := json.MarshalIndent(catalogService.GetCatalog(), "", "\t")
	log.Print(string(s))
}

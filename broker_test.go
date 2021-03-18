package main

import (
	"github.com/MaxFuhrich/serviceBrokerDummy/controller"
	"github.com/MaxFuhrich/serviceBrokerDummy/server"
	"github.com/MaxFuhrich/serviceBrokerDummy/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCatalog(t *testing.T) {
	catalog, err := server.MakeCatalog()
	if err != nil {
		t.Errorf("Catalog could not be created!")
	}
	settings, err := server.MakeSettings()
	if err != nil {
		t.Errorf("Settings could not be created!")
	}
	catalogService := service.NewCatalogService(catalog)
	catalogController := controller.NewCatalogController(&catalogService, settings)
	router := gin.Default()
	router.GET("/v2/catalog", catalogController.GetCatalog)
	w := performRequest(router, "GET", "/v2/catalog")
	catalogJson, err := os.Open("config/catalog.json")
	if err != nil {
		t.Errorf("Could not open config/catalog.json")
	}
	byteVal, err := ioutil.ReadAll(catalogJson)
	require.JSONEq(t, w.Body.String(), string(byteVal))
}

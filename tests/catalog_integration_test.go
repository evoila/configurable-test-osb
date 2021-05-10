package tests

import (
	"github.com/evoila/configurable-test-osb/controller"
	"github.com/evoila/configurable-test-osb/server"
	"github.com/evoila/configurable-test-osb/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

//Tests, if catalog is valid
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
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/v2/catalog", catalogController.GetCatalog)
	w := performRequest(router, "GET", "/v2/catalog", nil)
	currentPath, _ := os.Getwd()
	directories := strings.Split(currentPath, string(os.PathSeparator))
	directories = directories[:len(directories)-1]
	var target string
	target = directories[0] + string(os.PathSeparator)
	directories = directories[1:]
	var temp string
	temp = filepath.Join(append(directories, temp)...)
	target = filepath.Join(target, temp, "config", "catalog.json")
	catalogJson, err := os.Open(target)
	if err != nil {
		t.Errorf("Could not open config/catalog.json")
	}
	byteVal, err := ioutil.ReadAll(catalogJson)
	require.JSONEq(t, w.Body.String(), string(byteVal))
}

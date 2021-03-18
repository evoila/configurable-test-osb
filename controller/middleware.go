package controller

import (
	"encoding/base64"
	"encoding/json"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/model/profiles"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

type Middleware struct {
	settings *model.Settings
	platform *string
}

func NewMiddleware(settings *model.Settings, platform *string) Middleware {
	return Middleware{
		settings: settings,
		platform: platform,
	}
}

func (middleware *Middleware) BindAndCheckHeader(context *gin.Context) {
	var header model.Header
	err := context.ShouldBindHeader(&header)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	if header.APIVersionHeader == nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "The header \"X-Broker-API-Version\" is required but missing")
		return
	}
	if middleware.settings.HeaderSettings.RejectWrongAPIVersion {
		if middleware.settings.HeaderSettings.BrokerVersion != *header.APIVersionHeader {
			context.AbortWithStatusJSON(http.StatusPreconditionFailed, "Header \"X-Broker-API-Version\" is uses the wrong version")
			return
		}
	}
	if middleware.settings.HeaderSettings.OriginIDRequired || header.OriginID != nil {
		if header.OriginID == nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, "The header \"X-Broker-API-Originating-Identity\" is required but missing")
			return
		}
		separator := regexp.MustCompile(` `)
		split := separator.Split(*header.OriginID, 2)
		if len(split) != 2 {
			context.AbortWithStatusJSON(http.StatusBadRequest, "Header X-Broker-API-Originating-Identity has "+
				"malformed format! Format must be \"platform value\"")
			return
		}
		*middleware.platform = split[0]
		if middleware.settings.HeaderSettings.OriginIDValMustMatchProfile {
			if split[0] == "cloudfoundry" {
				decoded, err := base64.StdEncoding.DecodeString(split[1])
				if err != nil {
					context.AbortWithStatusJSON(http.StatusBadRequest, "Value in header "+
						"X-Broker-API-Originating-Identity could not be decoded: "+err.Error())
					return
				}
				var cf profiles.CloudFoundryOriginatingIdentityHeader
				err = json.Unmarshal(decoded, &cf)
				if err != nil {
					context.AbortWithStatusJSON(http.StatusBadRequest, "Unable to unmarshal value from header "+
						"X-Broker-API-Originating-Identity: "+err.Error())
					return
				}
			} else if split[0] == "kubernetes" {
				decoded, err := base64.StdEncoding.DecodeString(split[1])
				if err != nil {
					context.AbortWithStatusJSON(http.StatusBadRequest, "Value in header "+
						"X-Broker-API-Originating-Identity could not be decoded: "+err.Error())
					return
				}
				var k8 profiles.KubernetesOriginatingIdentityHeader
				err = json.Unmarshal(decoded, &k8)
				if err != nil {
					context.AbortWithStatusJSON(http.StatusBadRequest, "Unable to unmarshal value from header "+
						"X-Broker-API-Originating-Identity: "+err.Error())
					return
				}
			}
		}
	} else {
		*middleware.platform = ""
	}
	if middleware.settings.HeaderSettings.BrokerVersion > "2.14" && middleware.settings.HeaderSettings.RequestIDRequired {
		if header.RequestID == nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, "The header \"X-Broker-API-Request-Identity\" is required but missing")
		}
		context.Header("X-Broker-API-Request-Identity", *header.RequestID)
	}
}

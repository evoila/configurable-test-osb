package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/MaxFuhrich/serviceBrokerDummy/model"
	"github.com/MaxFuhrich/serviceBrokerDummy/model/profiles"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

type Middleware struct {
	//Only needed if the header should be bound once and other functions/handlers are supposed to use the pointer
	//header *model.Header
	settings *model.Settings
}

func NewMiddleware(settings *model.Settings) Middleware {
	return Middleware{settings: settings}
}

func (middleware *Middleware) BindAndCheckHeader(context *gin.Context) {
	//is the bound header NEEDED by caller of this function? YES
	var header model.Header
	err := context.ShouldBindHeader(&header)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return

	}
	if middleware.settings.HeaderSettings.RejectEmptyAPIVersion {
		if header.APIVersionHeader == nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, "the header \"X-Broker-API-Version\" is required but missing")
			return
		}
	}
	if middleware.settings.HeaderSettings.RejectWrongAPIVersion {
		if middleware.settings.HeaderSettings.BrokerVersion != *header.APIVersionHeader {
			context.AbortWithStatusJSON(http.StatusPreconditionFailed, "header \"X-Broker-API-Version\" is uses the wrong version")
			return
		}
	}
	if middleware.settings.HeaderSettings.OriginIDRequired {
		if header.OriginID == nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, "the header \"X-Broker-API-Originating-Identity\" is required but missing")
			return
		}
		separator := regexp.MustCompile(` `)
		split := separator.Split(*header.OriginID, 2)
		fmt.Println("GEORG")
		if len(split) != 2 {
			context.AbortWithStatusJSON(http.StatusBadRequest, "header X-Broker-API-Originating-Identity has "+
				"malformed format! format must be \"platform value\"")
			return
		}
		if middleware.settings.HeaderSettings.OriginIDValMustMatchProfile {
			if split[0] == "cloudfoundry" {
				decoded, err := base64.StdEncoding.DecodeString(split[1])
				if err != nil {
					context.AbortWithStatusJSON(http.StatusBadRequest, "value in header "+
						"X-Broker-API-Originating-Identity could not be decoded: "+err.Error())
					return
				}
				var cf profiles.CloudFoundryOriginatingIdentityHeader
				err = json.Unmarshal(decoded, &cf)
				if err != nil {
					context.AbortWithStatusJSON(http.StatusBadRequest, "unable to unmarshal value from header "+
						"X-Broker-API-Originating-Identity: "+err.Error())
					return
				}
			} else if split[0] == "kubernetes" {
				decoded, err := base64.StdEncoding.DecodeString(split[1])
				if err != nil {
					context.AbortWithStatusJSON(http.StatusBadRequest, "value in header "+
						"X-Broker-API-Originating-Identity could not be decoded: "+err.Error())
					return
				}
				var k8 profiles.KubernetesOriginatingIdentityHeader
				err = json.Unmarshal(decoded, &k8)
				if err != nil {
					context.AbortWithStatusJSON(http.StatusBadRequest, "unable to unmarshal value from header "+
						"X-Broker-API-Originating-Identity: "+err.Error())
					return
				}
				/*
					s, _ := json.MarshalIndent(k8, "", "\t")
					log.Println(string(s))
				*/
			}
		}
		//fmt.Println(split[0])
		//fmt.Println(split[1])
	}
	if middleware.settings.HeaderSettings.RequestIDRequired {
		if header.RequestID == nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, "the header \"X-Broker-API-Request-Identity\" is required but missing")
		}
	}

	/*
		TO DO
		"origin_id_val_must_match_profile": true,
		"log_request_id": true,
		"request_id_in_response": true,
		"etag_if_modified_since_in_response": false
	*/

	/*s, _ := json.MarshalIndent(header, "", "\t")
	log.Println(string(s))
	log.Println("Header Middleware.settings:")
	s, _ = json.MarshalIndent(middleware.settings, "", "\t")
	log.Println(string(s))

	*/
	//return &header, nil
}

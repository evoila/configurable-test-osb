package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type RequestSettings struct {
	AsyncEndpoint             *bool `json:"async_endpoint"`
	SecondsToComplete         *int  `json:"seconds_to_complete"`
	FailAtOperation           *bool `json:"fail_at_operation"`
	InstanceUsableAfterFail   *bool `json:"instance_usable_after_fail,omitempty"`
	UpdateRepeatableAfterFail *bool `json:"update_repeatable_after_fail,omitempty"`
}

func GetRequestSettings(params *interface{}) (*RequestSettings, error) {
	var requestSettings RequestSettings
	if params != nil {
		//fmt.Println("params is not nil")
		paramMap := (*params).(map[string]interface{})
		settingsInterface := paramMap["config_broker_settings"]
		if settingsInterface == nil {
			return nil, errors.New("config_broker_settings not found")
		}
		settingsMap := settingsInterface.(map[string]interface{})
		jsonBody, err := json.Marshal(settingsMap)
		if err != nil {
			fmt.Println("error: " + err.Error())
		}
		if err = json.Unmarshal(jsonBody, &requestSettings); err != nil {
			// do error check
			fmt.Println(err)
			return nil, err
		}
	}
	//fmt.Println("right after checking params")
	if requestSettings.AsyncEndpoint == nil {
		fmt.Println("settings asyncendpoint to false")
		val := false
		requestSettings.AsyncEndpoint = &val
	} else {
		fmt.Println("asyncendpoint:")
		fmt.Println(requestSettings.AsyncEndpoint)
	}
	if requestSettings.FailAtOperation == nil {
		val := false
		requestSettings.FailAtOperation = &val
	}
	/*if requestSettings.InstanceUsableAfterFail == nil {
		val := true
		requestSettings.InstanceUsableAfterFail = &val
	}
	if requestSettings.UpdateRepeatableAfterFail == nil {
		val := true
		requestSettings.UpdateRepeatableAfterFail = &val
	}*/
	if requestSettings.SecondsToComplete == nil {
		val := 0
		requestSettings.SecondsToComplete = &val
	}
	log.Println("now returning requestSettings...")
	return &requestSettings, nil
}

package model

import (
	"encoding/json"
	"errors"
)

type RequestSettings struct {
	AsyncEndpoint             *bool `json:"async_endpoint"`
	SecondsToComplete         *int  `json:"seconds_to_complete"`
	FailAtOperation           *bool `json:"fail_at_operation"`
	InstanceUsableAfterFail   *bool `json:"instance_usable_after_fail,omitempty"`
	UpdateRepeatableAfterFail *bool `json:"update_repeatable_after_fail,omitempty"`
}

//GetRequestSettings tries to get RequestSettings from the field "config_broker_settings" in params in the request body.
//If values in the request are missing, they will be replaced by default values.
//The requestSettings are responsible for the behaviour of and endpoint call (for example, if async)
//Returns *RequestSettings, error
func GetRequestSettings(params *interface{}) (*RequestSettings, error) {
	var requestSettings RequestSettings
	if params != nil {
		paramMap := (*params).(map[string]interface{})
		settingsInterface := paramMap["config_broker_settings"]
		if settingsInterface == nil {
			return nil, errors.New("config_broker_settings not found")
		}
		settingsMap := settingsInterface.(map[string]interface{})
		jsonBody, err := json.Marshal(settingsMap)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(jsonBody, &requestSettings); err != nil {
			return nil, err
		}
	}
	if requestSettings.AsyncEndpoint == nil {
		val := false
		requestSettings.AsyncEndpoint = &val
	}
	if requestSettings.FailAtOperation == nil {
		val := false
		requestSettings.FailAtOperation = &val
	}
	if requestSettings.SecondsToComplete == nil {
		val := 0
		requestSettings.SecondsToComplete = &val
	}
	return &requestSettings, nil
}

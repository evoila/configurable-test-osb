package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type RequestSettings struct {
	AsyncEndpoint     *bool `json:"async_endpoint"`
	SecondsToComplete *int  `json:"seconds_to_complete"`
}

func GetRequestSettings(params interface{}) (*RequestSettings, error) {
	var requestSettings RequestSettings
	//settings := reflect.ValueOf(params).FieldByName("config_broker_settings")
	paramMap := params.(map[string]interface{})
	settingsInterface := paramMap["config_broker_settings"]
	if settingsInterface == nil {
		return nil, errors.New("config_broker_settings not found")
	}
	settingsMap := settingsInterface.(map[string]interface{})
	jsonbody, err := json.Marshal(settingsMap)
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
	if err = json.Unmarshal(jsonbody, &requestSettings); err != nil {
		// do error check
		fmt.Println(err)
		return nil, err
	}

	/*if !settings.IsValid() {
		return nil, errors.New("No field \"")
	}*/
	fmt.Println("checking secondstocomplete")
	if requestSettings.SecondsToComplete == nil {
		val := 0
		requestSettings.SecondsToComplete = &val
	}
	log.Println("now returning requestSettings...")
	return &requestSettings, nil
}

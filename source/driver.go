/* JSON notes:
After much wrestling with trying to parse arbitrary complex JSON using json.Unmarshal() and then access stuff in the resulting
interface map and an aborted attempt to use a JSON parser that seemed good at first but proved flaky by producing different results
from the same simple code depending on whether it was running in a microservice or in a small function, we hit on the scheme used
below.  We convert JSON received from a Bravia to a struct using this tool:
https://mholt.github.io/json-to-go/
Then we unmarshal into that struct and access the elements via the struct type.  This seems to work well, and even better new
elements in the JSON appear to be ignored by json.Unmarshal(). */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Dartmouth-OpenAV/microservice-framework/framework"
)

var apiKey = "1234"

func getAudioMute(socketKey string, output string) (string, error) {
	function := "getVolume"
	framework.Log(function + " - called for: " + socketKey)

	volume, mute, err := getVolumeAndMute(socketKey, output)
	volume = volume // appease golang parser
	return mute, err
}

func setAudioMute(socketKey string, output string, mute string) (string, error) {
	function := "setAudioMute"
	var err error
	var bodyStr string

	framework.Log(function + " - called for: " + socketKey)

	mute = strings.Trim(mute, `"`) // trim quotes off JSON body input string

	if mute == "toggle" {
		framework.Log("AAAAAA calling getAudioMute()")
		currentMute, err := getAudioMute(socketKey, output)
		framework.Log("AAAAAA back from calling getAudioMute()")
		if err != nil {
			errMsg := fmt.Sprintf(function+" - vaev234 error getting current mute state for toggle: "+currentMute+" and error %v", err)
			framework.Log(errMsg)
			return errMsg, errors.New("POST error")
		}
		if currentMute == `"true"` {
			mute = "false"
		} else if currentMute == `"false"` {
			mute = "true"
		} else {
			errMsg := function + " - ohunjk32 got illegal mute value from getAudioMute"
			framework.Log(errMsg)
			return errMsg, errors.New("mute retrieve error")
		}
	}

	maxRetries := 2
	for maxRetries > 0 {
		framework.Log("AAAAA doing post to set mute: " + mute)
		bodyStr, err = framework.DoPost(socketKey, "sony/audio",
			`{"method":"setAudioMute", "id":601, "params":[{"status":`+mute+`}], "version":"1.0"}`)
		if err != nil || strings.Contains(bodyStr, "Display Is Turned off") { // Something went wrong - perhaps try again
			maxRetries--
			if maxRetries != 0 {
				if strings.Contains(bodyStr, "Display Is Turned off") {
					framework.Log("retrying because display is off")
				} else {
					framework.Log("POST failed, got: " + bodyStr + " retrying")
				}
			}
			time.Sleep(time.Second)
		} else { // Succeeded
			maxRetries = 0
		}
	}
	if err != nil { // still have an error - report this
		errMsg := fmt.Sprintf(function+" - vasd234 error doing post, got response: "+bodyStr+" and error %v", err)
		return errMsg, errors.New("POST error")
	}
	framework.Log(function + " - 5i76fg got: " + bodyStr + " back from post to: " + socketKey)

	// The Bravia REST API doesn't return any confirmation of the state, so since we didn't get an error, we're done
	return `"ok"`, nil
}

func getVolume(socketKey string, name string) (string, error) {
	function := "getVolume"
	framework.Log(function + " - called for: " + socketKey)

	volume, mute, err := getVolumeAndMute(socketKey, name)
	mute = mute
	return volume, err
}

func getVolumeAndMute(socketKey string, name string) (string, string, error) {
	function := "getVolumeAndMute"
	framework.Log(function + " - called for: " + socketKey)
	var err error
	var bodyStr string

	bodyStr, err = framework.DoPost(socketKey, "sony/audio",
		`{"method":"getVolumeInformation", "id":33, "params":[], "version":"1.0"}`)
	if err != nil {
		errMsg := fmt.Sprintf(function+" - iohnlk879 error doing post, got response: "+bodyStr+" and error %v", err)
		return errMsg, errMsg, errors.New("POST error")
	}
	framework.Log(function + " - kpbj76 got: " + bodyStr + " back from post to: " + socketKey)

	volume, mute, err := parseVolumeAndMute(bodyStr)
	if err != nil {
		return volume, mute, err
	}

	framework.Log("volume is: " + volume + " mute is: " + mute)

	return `"` + volume + `"`, `"` + mute + `"`, nil
}

func parseVolumeAndMute(jsonStr string) (string, string, error) {
	function := "parseVolumeAndMute"

	// Responses look like:
	//    {"result":[[{"target":"speaker","volume":35,"mute":false,"maxVolume":100,"minVolume":0}]],"id":33}
	// or if power is off:
	//    {"error":[40005,"Display Is Turned off"],"id":103}

	framework.Log(function + " - asdasnw34 jsonStr is: " + jsonStr)

	type errorStruct struct {
		Error []interface{} `json:"error"`
		ID    int           `json:"id"`
	}

	type volumeAndMuteStruct struct {
		Result [][]struct {
			Target    string `json:"target"`
			Volume    int    `json:"volume"`
			Mute      bool   `json:"mute"`
			MaxVolume int    `json:"maxVolume"`
			MinVolume int    `json:"minVolume"`
		} `json:"result"`
		ID int `json:"id"`
	}

	if strings.Contains(jsonStr, "error") {
		resp := &errorStruct{}
		err := json.Unmarshal([]byte(jsonStr), resp)
		errorNum := -999.0
		errorStr := "noErrorStringFound"
		if err == nil {
			errorNum = resp.Error[0].(float64)
			errorStr = resp.Error[1].(string)
		}
		if errorNum == 40005.0 { // Not really an error condition that the device is off: return "unknown"
			framework.Log(function + " - returning 0 volume and unknown mute because power is off")
			return "0", "unknown", nil
		} else { // Got an actual error
			errMsg := fmt.Sprintf(function+" - lihil89 Sony returned error string: %v number: %.f: ", errorStr, errorNum)
			framework.Log(errMsg)
			return errMsg, errMsg, errors.New("Sony error")
		}
	}

	resp2 := &volumeAndMuteStruct{}
	volume := -999
	mute := false
	err := json.Unmarshal([]byte(jsonStr), resp2)
	if err != nil {
		errMsg := function + " - ########### q32rasd error unmarshaling: " + jsonStr
		framework.Log(errMsg)
	} else {
		volume = resp2.Result[0][0].Volume
		mute = resp2.Result[0][0].Mute
		fmt.Printf("########### unmarshaled volume: %v mute: %v\n", volume, mute)
	}

	volumeStr := fmt.Sprintf("%v", volume)
	var muteStr string
	if mute {
		muteStr = "true"
	} else if mute == false {
		muteStr = "false"
	} else {
		muteStr = "unknown"
	}

	framework.Log(function + " - returning volume: " + volumeStr + " mute: " + muteStr)
	return volumeStr, muteStr, nil
}

func setVolume(socketKey string, output string, volume string) (string, error) {
	function := "setVolume"
	var err error
	var bodyStr string

	framework.Log(function + " - called for: " + socketKey)

	volume = strings.Trim(volume, `"`) // trim quotes off JSON body input string
	if volume == "up" {
		volume = "+5"
	} else if volume == "down" {
		volume = "-5"
	}
	// if neither "up" or "down", then pass through the numeric value
	maxRetries := 2
	for maxRetries > 0 {
		framework.Log("AAAAA doing post to set volume: " + volume)
		bodyStr, err = framework.DoPost(socketKey, "sony/audio",
			`{"method":"setAudioVolume", "id":98, "params":[{"volume": "`+volume+`", "ui":"on", "target":""}], "version":"1.0"}`)

		if err != nil || strings.Contains(bodyStr, "Display Is Turned off") { // Something went wrong - perhaps try again
			maxRetries--
			if maxRetries != 0 {
				if strings.Contains(bodyStr, "Display Is Turned off") {
					framework.Log("retrying because display is off")
				} else {
					framework.Log("POST failed, got: " + bodyStr + " retrying")
				}
				time.Sleep(time.Second)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	if err != nil { // still have an error - report this
		errMsg := fmt.Sprintf(function+" - qf3svd error doing post, got response: "+bodyStr+" and error %v", err)
		return errMsg, errors.New("POST error")
	}
	framework.Log(function + " - 4hnfs got: " + bodyStr + " back from post to: " + socketKey)

	// The Bravia REST API doesn't return any confirmation of the state, so since we didn't get an error, we're done
	return `"ok"`, nil
}

func getVideoRoute(socketKey string, output string) (string, error) {
	function := "getVideoRoute"
	var err error
	var bodyStr string

	framework.Log(function + " - called for: " + socketKey)
	bodyStr, err = framework.DoPost(socketKey, "sony/avContent",
		`{"method":"getPlayingContentInfo", "id":103, "params":[], "version":"1.0"}`)
	if err != nil {
		errMsg := fmt.Sprintf(function+" - 8j;alknf8 error doing post, got response: "+bodyStr+" and error %v", err)
		return errMsg, errors.New("POST error")
	}
	framework.Log(function + " - lkkjr32 got: " + bodyStr + " back from post to: " + socketKey)

	input, err := parseVideoRoute(bodyStr)
	if err != nil {
		return input, err
	}

	framework.Log("input is: " + input)

	return `"` + input + `"`, nil
}

func parseVideoRoute(jsonStr string) (string, error) {
	function := "parseVideoRoute"

	// Responses look like:
	//    {"result":[{"uri":"extInput:hdmi?port=2","source":"extInput:hdmi","title":"HDMI 2"}],"id":103}
	// or if power is off:
	//    {"error":[40005,"Display Is Turned off"],"id":103}

	framework.Log(function + " - asdasnw34 jsonStr is: " + jsonStr)

	type errorStruct struct {
		Error []interface{} `json:"error"`
		ID    int           `json:"id"`
	}

	type videoRouteStruct struct {
		Result []struct {
			URI    string `json:"uri"`
			Source string `json:"source"`
			Title  string `json:"title"`
		} `json:"result"`
		ID int `json:"id"`
	}

	if strings.Contains(jsonStr, "error") {
		resp := &errorStruct{}
		err := json.Unmarshal([]byte(jsonStr), resp)
		errorNum := -999.0
		errorStr := "noErrorStringFound"
		if err == nil {
			errorNum = resp.Error[0].(float64)
			errorStr = resp.Error[1].(string)
		}
		if errorNum == 40005.0 { // Not really an error condition that the device is off: return "unknown"
			framework.Log(function + " - returning unknown videoroute because power is off")
			return "unknown", nil
		} else { // Got an actual error
			errMsg := fmt.Sprintf(function+" - w4gsdb Sony returned error string: %v number: %.f: ", errorStr, errorNum)
			framework.Log(errMsg)
			return errMsg, errors.New("Sony error")
		}
	}

	resp2 := &videoRouteStruct{}
	uri := "notFound"
	err := json.Unmarshal([]byte(jsonStr), resp2)
	if err != nil {
		errMsg := function + " - ########### 23qrgar error unmarshaling: " + jsonStr
		framework.Log(errMsg)
	} else {
		uri = resp2.Result[0].URI
		framework.Log("########### unmarshaled uri: " + uri)
	}

	input := strings.Split(uri, "=")[1] // should only be one "="

	framework.Log(function + " - returning input: " + input)
	return input, nil
}

func setVideoRoute(socketKey string, output string, input string) (string, error) {
	function := "setVideoRoute"
	var err error
	var bodyStr string

	framework.Log(function + " - called for: " + socketKey)

	input = strings.Trim(input, `"`) // trim quotes off JSON body input string
	maxRetries := 15                 // increasing this to see if it helps in the huddle spaces setting the correct input on power up
	for maxRetries > 0 {
		framework.Log("AAAAA doing post to set video route: " + input)
		bodyStr, err = framework.DoPost(socketKey, "sony/avContent",
			`{"method":"setPlayContent", "id":101, "params":[{"uri": "extInput:hdmi?port=`+input+`"}], "version":"1.0"}`)

		if err != nil || strings.Contains(bodyStr, "Display Is Turned off") { // Something went wrong - perhaps try again
			maxRetries--
			if maxRetries != 0 {
				if strings.Contains(bodyStr, "Display Is Turned off") {
					framework.Log("retrying because display is off")
				} else {
					framework.Log("POST failed, got: " + bodyStr + " retrying")
				}
				time.Sleep(time.Second)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	if err != nil { // still have an error - report this
		errMsg := fmt.Sprintf(function+" - warvd354 error doing post, got response: "+bodyStr+" and error %v", err)
		return errMsg, errors.New("POST error")
	}
	framework.Log(function + " - tn463df got: " + bodyStr + " back from post to: " + socketKey)

	// The Bravia REST API doesn't return any confirmation of the state, so since we didn't get an error, we're done
	return `"ok"`, nil
}

func getPower(socketKey string) (string, error) {
	function := "getPower"
	var err error
	var bodyStr string

	framework.Log(function + " - called for: " + socketKey)
	bodyStr, err = framework.DoPost(socketKey, "sony/system",
		`{"method":"getPowerStatus", "id":50, "params":[], "version":"1.0"}`)
	if err != nil {
		errMsg := fmt.Sprintf(function+" - 43q2sdfa error doing post, got response: "+bodyStr+" and error %v", err)
		return errMsg, errors.New("POST error")
	}
	framework.Log(function + " - lnk3;lk got: " + bodyStr + " back from post to: " + socketKey)

	status, err := parsePowerStatus(bodyStr)
	if err != nil {
		return status, err
	}

	framework.Log("status is: " + status)
	if status == "active" {
		status = "on"
	} else {
		status = "off"
	}

	return `"` + status + `"`, nil
}

func parsePowerStatus(jsonStr string) (string, error) {
	function := "parsePowerStatus"
	statusStr := "unknown"

	// Responses look like:
	// {"result":[{"status":"standby"}],"id":50}

	framework.Log(function + " - asdasnw34 jsonStr is: " + jsonStr)

	type powerStruct struct {
		Result []struct {
			Status string `json:"status"`
		} `json:"result"`
		ID int `json:"id"`
	}

	resp := &powerStruct{}
	err := json.Unmarshal([]byte(jsonStr), resp)
	if err != nil {
		errMsg := function + " - ########### acsd423 error unmarshaling: " + jsonStr
		framework.Log(errMsg)
	} else {
		statusStr = resp.Result[0].Status
		framework.Log("########### unmarshaled status: " + statusStr)
	}
	return statusStr, nil
}

func setPower(socketKey string, value string) (string, error) {
	function := "setPower"
	var err error
	var bodyStr string
	var powerState string

	framework.Log(function + " - setting: " + socketKey + " to: " + value)
	if value == `"on"` {
		powerState = "true"
	} else if value == `"off"` {
		powerState = "false"
	} else {
		errMsg := function + " - 789hlfsad invalid power value: " + value
		return errMsg, errors.New("invalid power value")
	}

	postBody := `{"method":"setPowerStatus", "id":55, "params":[{"status": ` + powerState + `}], "version":"1.0"}`
	maxRetries := 2
	for maxRetries > 0 {
		framework.Log("AAAAA doing post to set power: " + powerState)
		bodyStr, err = framework.DoPost(socketKey, "sony/system", postBody)

		if err != nil || strings.Contains(bodyStr, "Display Is Turned off") { // Something went wrong - perhaps try again
			maxRetries--
			if maxRetries != 0 {
				if strings.Contains(bodyStr, "Display Is Turned off") {
					framework.Log("retrying because display is off")
				} else {
					framework.Log("POST failed, got: " + bodyStr + " retrying")
				}
				time.Sleep(time.Second)
			}
		} else { // Succeeded
			maxRetries = 0
		}
	}

	if err != nil { // still have an error - report this
		errMsg := fmt.Sprintf(function+" - 543efad error doing post, got response: "+bodyStr+" and error %v", err)
		return errMsg, errors.New("POST error")
	}
	framework.Log(function + " - lnnlkh787 got: " + bodyStr + " back from post to: " + socketKey)

	// The Bravia REST API doesn't return any confirmation of the state, so since we didn't get an error, we're done
	return `"ok"`, nil
}

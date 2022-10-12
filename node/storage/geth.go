package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"node/constants"
	"node/random"
)

type gethRequest struct {
	Params  interface{} `json:"params"`
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      int         `json:"id"`
}

type gethResponse struct {
	Result  interface{}     `json:"result"`
	Jsonrpc string          `json:"jsonrpc"`
	Error   gethErrorStruct `json:"error"`
	ID      int             `json:"id"`
}

type gethErrorStruct struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func makeGethRequestString(method string, params interface{}) (string, error) {
	response, err := makeGethRequestInterface(method, params)
	if err != nil {
		return "", err
	}
	result := response.Result

	// Needed since geth returns different types in result field
	switch reflect.TypeOf(result) {
	case nil:
		// nil
		return "", nil
	case reflect.TypeOf(""):
		// string
		return result.(string), nil //nolint: forcetypeassert
	case reflect.TypeOf(true):
		// bool
		if !result.(bool) { //nolint: forcetypeassert
			return "", fmt.Errorf("geth returned false for method '%s'", method)
		}

		return "", nil
	default:
		// other
		return "", fmt.Errorf("geth returned unexpected type %s for method '%s'", reflect.TypeOf(result), method)
	}
}

func makeGethRequestInterface(method string, params interface{}) (gethResponse, error) {
	requestID := random.PositiveIntFromRange(0, 2048)
	requestData := gethRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      requestID,
	}

	requestBytes, err := json.Marshal(requestData)
	if err != nil {
		return gethResponse{}, fmt.Errorf("could not parse gethRequest struct: %w", err)
	}

	// Build request
	request, err := http.NewRequest(http.MethodPost, constants.GethAddress, bytes.NewReader(requestBytes))
	if err != nil {
		return gethResponse{}, fmt.Errorf("could not create request: %w", err)
	}
	request.Header.Add("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return gethResponse{}, fmt.Errorf("could not make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return gethResponse{}, fmt.Errorf("could not read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return gethResponse{}, fmt.Errorf("status code of '%d' indicates failure. Body: %s", resp.StatusCode, string(body))
	}

	var response gethResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	err = decoder.Decode(&response)
	if err != nil {
		return gethResponse{}, fmt.Errorf("error '%w' when decoding response body '%s'", err, resp.Body)
	}

	if !reflect.DeepEqual(response.Error, gethErrorStruct{}) {
		return gethResponse{}, fmt.Errorf("geth returned an error with code: %d and message: %s", response.Error.Code, response.Error.Message)
	}

	if response.ID != requestID {
		return gethResponse{}, fmt.Errorf("geth returned unexpected ID '%d'; expected: '%d'", response.ID, requestID)
	}

	return response, nil
}

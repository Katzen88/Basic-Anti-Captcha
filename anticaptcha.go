package anticaptcha

import (
	"bytes"
	//"encoding/json"
	//"errors"
	"net/http"
	//"net/url"
	"time"
	"fmt"
	"strings"

	"io/ioutil"
)

var (
	checkInterval = 2 * time.Second
)
type Client struct {
	APIKey string
}

// Method to check the result of a given task, returns the json returned from the api
func (c *Client) getTaskResult(taskID string) (string, error) {
	//baseURL := "https://api.anti-captcha.com"
	client := &http.Client{}

	// Mount the data to be sent
	bData := fmt.Sprintf("{ \"clientKey\":\"%s\", \"taskId\": %v}", c.APIKey, taskID)
	var jsonStr = []byte(bData)
	// Make the request
	resultPreResp, err := http.NewRequest("POST","https://api.anti-captcha.com/getTaskResult", bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println(err)
	}
	resultPreResp.Header.Set("Accept", "application/json")
	resultPreResp.Header.Set("Content-Type", "application/json")
	resultResp, err := client.Do(resultPreResp)
	if err != nil {
		fmt.Println(err)
	}
	defer resultResp.Body.Close()

	// Decode response
	respBody, err := ioutil.ReadAll(resultResp.Body)
	if err != nil {
		fmt.Println(err)
	}
	dSTR := fmt.Sprint(string(respBody))
	fmt.Println(dSTR)
	return dSTR, nil
}


// Method to create the task to process the image captcha, returns the task_id
func (c *Client) createTaskImage(imgString string) (string, error) {
	parsedIMGb := strings.Split(imgString,"base64,")[1]
	// Mount the data to be sent
	client := &http.Client{}
	postBytes := fmt.Sprintf("{ \"clientKey\":\"%s\", \"task\": { \"type\":\"ImageToTextTask\", \"body\":\"%s\", \"phrase\":false, \"case\":false, \"numeric\":0, \"math\":false, \"minLength\":0, \"maxLength\":0 } }", c.APIKey, parsedIMGb)
	// Make the request
	preResp, err := http.NewRequest("POST", "https://api.anti-captcha.com/createTask", strings.NewReader(postBytes))
	if err != nil {
		return "0", err
	}
	preResp.Header.Set("Accept", "application/json")
	preResp.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(preResp)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// Decode response
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(respBody))
	parsedTaskID := strings.Split(string(respBody),"taskId\":")[1]
	taskID := strings.ReplaceAll(parsedTaskID, "}", "")
	fmt.Println(taskID)
	// TODO treat api errors and handle them properly
	return taskID, nil
}

// SendImage Method to encapsulate the processing of the image captcha
// Given a base64 string from the image, it sends to the api and waits until
// the processing is complete to return the evaluated key
func (c *Client) SendImage(imgString string) (string, error) {
	//fmt.Println(imgString)
	// Create the task on anti-captcha api and get the task_id

	taskID, err := c.createTaskImage(imgString)
	fmt.Println("cow talk")
	if err != nil {
		return "", err
	}
	fmt.Println("cow talk", taskID)

	// Check if the result is ready, if not loop until it is
	response, err := c.getTaskResult(taskID)
	if err != nil {
		return "", err
	}
	var respBodySTRk string
	for {
		respBodySTR := fmt.Sprintf(response)
		if strings.Contains(respBodySTR, "status\":\"processing\""){
			time.Sleep(checkInterval)
			response, err = c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
		} else {
			respBodySTRk = fmt.Sprintf(respBodySTR)
			break
		}
	}
	return respBodySTRk, nil
}

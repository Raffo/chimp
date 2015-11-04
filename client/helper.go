package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var _ = log.Print

func (bc *Client) makeRequest(method string, url string, entity interface{}) (*http.Request, *http.Response, error) {
	req, err := bc.buildRequest(method, url, entity)
	if err != nil {
		return req, nil, err
	}
	res, err := http.DefaultClient.Do(req)
	return req, res, err
}

func (bc *Client) buildRequest(method string, url string, entity interface{}) (*http.Request, error) {
	body, err := encodeEntity(entity)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return req, err
	}

	req.Header.Set("content-type", "application/json")
	if bc.AccessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bc.AccessToken))
	}

	return req, err
}

func encodeEntity(entity interface{}) (io.Reader, error) {
	if entity == nil {
		return nil, nil
	} else {
		b, err := json.Marshal(entity)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(b), nil
	}
}

type InfoMessage struct {
	Info []Artifact `json:"data"`
}

type ListMessage struct {
	List []string `json:"data"`
}

//object at go level
type ChimpInfoResponse struct {
	Title  string
	Detail string
	Data   []Artifact
}

type ChimpListResponse struct {
	Title  string
	Detail string
	Data   []string
}

func unmarshalResponse(r *http.Response, data interface{}) error {
	respBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return errors.New("Cannot read response from body")
	}
	err = json.Unmarshal(respBody, data)
	if err != nil {
		fmt.Printf("%T\n%s\n%#v\n", err, err, err)
		return errors.New("Cannot unmarshal json data")
	} else {
		return nil
	}
}

func checkStatusOK(status int) bool {
	if status >= 500 {
		return false
	} else {
		return true
	}
}

func checkAuthOK(status int) bool {
	if status == 401 || status == 403 {
		return false
	} else {
		return true
	}
}

func handleAuthNOK(status int) {
	switch status {
	case 401:
		fmt.Println("Unauthorized. Please check the provided token.")
	case 403:
		fmt.Println("You are not authorized to perform this action.")
	}
}

func handleStatusNOK(status int) {
	switch status {
	case 500:
		fmt.Println("Internal error.")
	default:
		fmt.Println("Generic error.")
	}
}

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NewControllerClient create the new controller client
func NewControllerClient(address string) (*ControllerClient, error) {
	address = strings.TrimPrefix(address, "tcp://")

	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}

	if !strings.HasSuffix(address, "/v1") {
		address += "/v1"
	}

	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(u.Host, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid address %s, must have a port in it", address)
	}

	timeout := time.Duration(2 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	return &ControllerClient{
		Host:       parts[0],
		Address:    address,
		httpClient: client,
	}, nil
}

func (c *ControllerClient) Post(path string, req, resp interface{}) error {
	return c.Do("POST", path, req, resp)
}

func (c *ControllerClient) Do(method, path string, req, resp interface{}) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	bodyType := "application/json"
	url := path
	if !strings.HasPrefix(url, "http") {
		url = c.Address + path

	}

	httpReq, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", bodyType)

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 300 {
		content, _ := ioutil.ReadAll(httpResp.Body)
		return fmt.Errorf("Bad response: %d %s: %s", httpResp.StatusCode, httpResp.Status, content)
	}

	if resp == nil {
		return nil
	}
	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func GetVolume(path string) (*Volumes, error) {
	var volume VolumeCollection
	var c ControllerClient

	err := c.Get(path+"/volumes", &volume)
	if err != nil {
		return nil, err
	}

	if len(volume.Data) == 0 {
		return nil, errors.New("No volume found")
	}

	return &volume.Data[0], nil
}
func (c *ControllerClient) Get(path string, obj interface{}) error {
	resp, err := http.Get(path)

	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(obj)
}

func (c *ControllerClient) ListReplicas(path string) ([]Replica, error) {
	var resp ReplicaCollection

	err := c.Get(path+"/replicas", &resp)

	return resp.Data, err
}

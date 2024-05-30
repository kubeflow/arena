// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
)

var defaultProxiedMlflowClient *MlflowClient
var once sync.Once

const (
	RESOURCE_DOES_NOT_EXIST_ERROR = "RESOURCE_DOES_NOT_EXIST"
	RESOURCE_ALREADY_EXISTS_ERROR = "RESOURCE_ALREADY_EXISTS"
)

type MlflowError struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

func (e *MlflowError) IsResourceDoesNotExistError() bool {
	return e.ErrorCode == "RESOURCE_DOES_NOT_EXIST"
}

func (e *MlflowError) IsResourceAlreadyExistsError() bool {
	return e.ErrorCode == "RESOURCE_ALREADY_EXISTS"
}

func (e *MlflowError) Error() string {
	bytes, err := json.Marshal(e)
	if err != nil {
		return ""
	}
	return string(bytes)
}

type MlflowClient struct {
	RestyClient *resty.Client
}

func NewMlflowClient(trackingUri, username, password string) *MlflowClient {
	restyClient := resty.New().
		SetBaseURL(trackingUri).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetDisableWarn(true)

	if username != "" && password != "" {
		restyClient.SetBasicAuth(username, password)
	}

	return &MlflowClient{
		RestyClient: restyClient,
	}
}

// Create a MLflow client proxied by Kubernetes api server
func NewProxiedMlflowClient(configr *config.ArenaConfiger, service *corev1.Service, username string, password string) *MlflowClient {
	once.Do(func() {
		namespace := service.ObjectMeta.Namespace
		name := service.ObjectMeta.Name
		port := 5000
		if len(service.Spec.Ports) > 0 {
			port = int(service.Spec.Ports[0].Port)
		}
		restClient := configr.GetClientSet().CoreV1().RESTClient().(*rest.RESTClient)
		baseUrl := restClient.Get().
			Resource("services").
			Namespace(namespace).
			Name(fmt.Sprintf("%s:%d", name, port)).
			SubResource("proxy").
			URL().
			String()

		restyClient := resty.New().
			SetTransport(restClient.Client.Transport).
			SetBaseURL(baseUrl).
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetDisableWarn(true)

		if username != "" && password != "" {
			// api-server proxy will strip out authorization header, see https://github.com/kubernetes/kubernetes/issues/38775
			log.Warn("proxied mlflow client currently does not support basic authentication")
		}

		defaultProxiedMlflowClient = &MlflowClient{
			RestyClient: restyClient,
		}
	})
	return defaultProxiedMlflowClient
}

func (c *MlflowClient) CheckHealth() (bool, error) {
	resp, err := c.RestyClient.
		R().
		Get("/health")
	if err != nil {
		return false, fmt.Errorf("failed to check whether mlflow tracking server is healthy: %v", err)
	}
	if resp.IsError() {
		return false, fmt.Errorf("failed to check whether mlflow tracking server is healthy: %v", resp.Status())
	}
	status := string(resp.Body())
	if status != "OK" {
		return false, nil
	}
	return true, nil
}

// For detailed information about MLflow REST API, see https://mlflow.org/docs/latest/rest-api.html
func (c *MlflowClient) CreateRegisteredModel(name string, tags []*types.RegisteredModelTag, description string) (*types.RegisteredModel, error) {
	req := &struct {
		Name        string                      `json:"name"`
		Tags        []*types.RegisteredModelTag `json:"tags"`
		Description string                      `json:"description"`
	}{
		Name:        name,
		Tags:        tags,
		Description: description,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqBody := string(reqBytes)

	res := &struct {
		RegisteredModel *types.RegisteredModel `json:"registered_model"`
	}{}

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Post("api/2.0/mlflow/registered-models/create")
	if err != nil {
		return nil, fmt.Errorf("failed to create registered model %v: %v", name, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return nil, fmt.Errorf("failed to create registered model \"%v\": %v", name, strings.ToUpper(resp.Status()))
		}
		return nil, fmt.Errorf("failed to create registered model %v: %v: %v", name, strings.ToUpper(resp.Status()), mlflowError)
	}
	return res.RegisteredModel, nil
}

func (c *MlflowClient) GetRegisteredModel(name string) (*types.RegisteredModel, error) {
	res := &struct {
		RegisteredModel *types.RegisteredModel `json:"registered_model"`
	}{}

	resp, err := c.RestyClient.
		R().
		SetQueryParam("name", name).
		SetResult(res).
		Get("api/2.0/mlflow/registered-models/get")
	if err != nil {
		return nil, fmt.Errorf("failed to get registered model %v: %v", name, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return nil, fmt.Errorf("failed to get registered model \"%v\": %v", name, strings.ToUpper(resp.Status()))
		}
		return nil, fmt.Errorf("failed to get registered model \"%v\": %v: %v", name, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("get registered model \"%v\" successfully", name)
	return res.RegisteredModel, nil
}

func (c *MlflowClient) RenameRegisteredModel(name, newName string) (*types.RegisteredModel, error) {
	req := &struct {
		Name    string `json:"name"`
		NewName string `json:"new_name"`
	}{
		Name:    name,
		NewName: newName,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqBody := string(reqBytes)

	res := &struct {
		RegisteredModel types.RegisteredModel `json:"registered_model"`
	}{}

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Post("api/2.0/mlflow/registered-models/rename")
	if err != nil {
		return nil, fmt.Errorf("failed to rename registered model %s to %s: %v", name, newName, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return nil, fmt.Errorf("failed to rename registered model \"%s\" to \"%s\": %v", name, newName, strings.ToUpper(resp.Status()))
		}
		return nil, fmt.Errorf("failed to rename registered model \"%s\" to \"%s\": %v: %v", name, newName, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("rename registered model \"%s\" to \"%s\" successfully", name, newName)
	return &res.RegisteredModel, nil
}

func (c *MlflowClient) UpdateRegisteredModel(name string, description string) (*types.RegisteredModel, error) {
	req := &struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{
		Name:        name,
		Description: description,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqBody := string(reqBytes)

	res := &struct {
		RegisteredModel *types.RegisteredModel `json:"registered_model"`
	}{}

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Patch("api/2.0/mlflow/registered-models/update")
	if err != nil {
		return nil, fmt.Errorf("failed to update registered model \"%s\": %v", name, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return nil, fmt.Errorf("failed to update registered model \"%s\": %v", name, strings.ToUpper(resp.Status()))
		}
		return nil, fmt.Errorf("failed to update registered model \"%s\": %v: %v", name, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("update registered model \"%s\" successfully", name)
	return res.RegisteredModel, nil
}

func (c *MlflowClient) DeleteRegisteredModel(name string) error {
	req := &struct {
		Name string `json:"name"`
	}{
		Name: name,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqBody := string(reqBytes)

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/registered-models/delete")
	if err != nil {
		return fmt.Errorf("failed to delete registered model %v: %v", name, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return fmt.Errorf("failed to delete registered model \"%s\": %v", name, strings.ToUpper(resp.Status()))
		}
		return fmt.Errorf("failed to delete registered model \"%s\": %v: %v", name, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("delete registered model \"%s\" successfully", name)
	return nil
}

func (c *MlflowClient) GetLatestModelVersions(name string, stages []string) ([]*types.ModelVersion, error) {
	req := &struct {
		Name   string   `json:"name"`
		Stages []string `json:"stages"`
	}{
		Name:   name,
		Stages: stages,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqBody := string(reqBytes)

	res := &struct {
		ModelVersions []*types.ModelVersion `json:"model_versions"`
	}{}

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Post("2.0/mlflow/registered-models/get-latest-versions")
	if err != nil {
		return nil, fmt.Errorf("failed to get latest versions of model \"%s\": %v", name, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return nil, fmt.Errorf("failed to get latest versions of model \"%s\": %v", name, strings.ToUpper(resp.Status()))
		}
		return nil, fmt.Errorf("failed to get latest versions of model \"%s\": %v: %v", name, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("get latest versions of model \"%s\" successfully", name)
	return res.ModelVersions, nil
}

func (c *MlflowClient) CreateModelVersion(name, source, runId string, tags []*types.ModelVersionTag, runLink, description string) (*types.ModelVersion, error) {
	if source == "" {
		return nil, errors.New("model version source must be specified when registering a model version")
	}
	modelVersion := types.ModelVersion{
		Name:        name,
		Source:      source,
		Tags:        tags,
		Description: description,
	}
	reqBytes, err := json.Marshal(modelVersion)
	if err != nil {
		return nil, err
	}
	reqBody := string(reqBytes)

	res := &struct {
		ModelVersion *types.ModelVersion `json:"model_version"`
	}{}
	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Post("api/2.0/mlflow/model-versions/create")
	if err != nil {
		return nil, fmt.Errorf("failed to create registered model %v: %v", name, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return nil, fmt.Errorf("failed to create model version of model \"%s\": %v", name, strings.ToUpper(resp.Status()))
		}
		return nil, fmt.Errorf("failed to create model version of model \"%s\": %v: %v", name, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("create model version of model \"%s\" successfully", name)
	return res.ModelVersion, nil
}

func (c *MlflowClient) GetModelVersion(name, version string) (*types.ModelVersion, error) {
	res := &struct {
		ModelVersion *types.ModelVersion `json:"model_version"`
	}{}

	resp, err := c.RestyClient.
		R().
		SetQueryParam("name", name).
		SetQueryParam("version", version).
		SetResult(res).
		Get("api/2.0/mlflow/model-versions/get")
	if err != nil {
		return nil, fmt.Errorf("failed to get model version %v/%v: %v", name, version, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return nil, fmt.Errorf("failed to get model version \"%s/%s\": %v", name, version, strings.ToUpper(resp.Status()))
		}
		return nil, fmt.Errorf("failed to get model version \"%s/%s\": %v: %v", name, version, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("get model version \"%s/%s\" successfully", name, version)
	return res.ModelVersion, nil
}

func (c *MlflowClient) UpdateModelVersion(name string, version string, description string) (*types.ModelVersion, error) {
	req := &struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
	}{
		Name:        name,
		Version:     version,
		Description: description,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqBody := string(reqBytes)

	res := &struct {
		ModelVersion *types.ModelVersion `json:"model_version"`
	}{}

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Patch("api/2.0/mlflow/model-versions/update")
	if err != nil {
		return nil, fmt.Errorf("failed to update model version %v/%v: %v", name, version, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return nil, fmt.Errorf("failed to update model version \"%s/%s\": %v", name, version, strings.ToUpper(resp.Status()))
		}
		return nil, fmt.Errorf("failed to update model version \"%s/%s\": %v: %v", name, version, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("update model version \"%s/%s\" successfully", name, version)
	return res.ModelVersion, nil
}

func (c *MlflowClient) DeleteModelVersion(name string, version string) error {
	req := types.ModelVersion{
		Name:    name,
		Version: version,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqBody := string(reqBytes)

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/model-versions/delete")
	if err != nil {
		return fmt.Errorf("failed to delete model version \"%s/%s\": %v", name, version, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return fmt.Errorf("failed to delete model version \"%s/%s\": %s", name, version, strings.ToUpper(resp.Status()))
		}
		return fmt.Errorf("failed to delete model version \"%s/%s\": %s: %v", name, version, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("delete model version \"%s/%s\" successfully", name, version)
	return nil
}

func (c *MlflowClient) SearchModelVersions(filter string, maxResults int, orderBy []string) ([]*types.ModelVersion, error) {
	parsedUrl, err := url.Parse("api/2.0/mlflow/model-versions/search")
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	if filter != "" {
		params.Add("filter", filter)
	}
	if maxResults > 0 {
		params.Add("max_results", strconv.Itoa(maxResults))
	}
	for _, clause := range orderBy {
		params.Add("order_by", clause)
	}

	modelVersions := []*types.ModelVersion{}
	nextPageToken := ""
	for {
		if nextPageToken != "" {
			params.Set("page_token", nextPageToken)
		}
		parsedUrl.RawQuery = params.Encode()
		url := parsedUrl.String()

		res := &struct {
			ModelVersions []*types.ModelVersion `json:"model_versions"`
			NextPageToken string                `json:"next_page_token"`
		}{}
		resp, err := c.RestyClient.
			R().
			SetResult(res).
			Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to search model versions: %v", err)
		}
		if resp.IsError() {
			mlflowError := &MlflowError{}
			if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
				return nil, fmt.Errorf("failed to search model versions: %v", strings.ToUpper(resp.Status()))
			}
			return nil, fmt.Errorf("failed to search model versions: %v: %v", strings.ToUpper(resp.Status()), mlflowError)
		}
		modelVersions = append(modelVersions, res.ModelVersions...)
		nextPageToken = res.NextPageToken
		if nextPageToken == "" {
			break
		}
	}
	log.Debugf("search model versions successfully")
	return modelVersions, nil
}

func (c *MlflowClient) GetDownloadUri(name, version string) (string, error) {
	req := &struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}{
		Name:    name,
		Version: version,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	reqBody := string(reqBytes)

	res := &struct {
		ArtifactUri string `json:"artifact_uri"`
	}{}

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Get("api/2.0/mlflow/model-versions/get-download-uri")
	if err != nil {
		return "", fmt.Errorf("failed to get artifacts download uri of \"%v/%v\": %v", name, version, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return "", fmt.Errorf("failed to get artifacts download uri of \"%v/%v\": %v", name, version, strings.ToUpper(resp.Status()))
		}
		return "", fmt.Errorf("failed to get artifacts download uri of \"%v/%v\": %v: %v", name, version, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("get artifacts download uri of \"%v/%v\" successfully", name, version)
	return res.ArtifactUri, nil
}

func (c *MlflowClient) SearchRegisteredModels(filter string, maxResults int, orderBy []string) ([]*types.RegisteredModel, error) {
	parsedUrl, err := url.Parse("api/2.0/mlflow/registered-models/search")
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	if filter != "" {
		params.Add("filter", filter)
	}
	if maxResults > 0 {
		params.Add("max_results", strconv.Itoa(maxResults))
	}
	for _, clause := range orderBy {
		params.Add("order_by", clause)
	}

	registeredModels := []*types.RegisteredModel{}
	nextPageToken := ""
	for {
		if nextPageToken != "" {
			params.Set("page_token", nextPageToken)
		}
		parsedUrl.RawQuery = params.Encode()
		url := parsedUrl.String()

		res := &struct {
			RegisteredModels []*types.RegisteredModel `json:"registered_models"`
			NextPageToken    string                   `json:"next_page_token"`
		}{}
		resp, err := c.RestyClient.
			R().
			SetResult(res).
			Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to search registered models: %v", err)
		}
		if resp.IsError() {
			mlflowError := &MlflowError{}
			if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
				return nil, fmt.Errorf("failed to search registered models: %v", strings.ToUpper(resp.Status()))
			}
			return nil, fmt.Errorf("failed to search registered models: %v: %v", strings.ToUpper(resp.Status()), mlflowError)
		}
		registeredModels = append(registeredModels, res.RegisteredModels...)
		nextPageToken = res.NextPageToken
		if nextPageToken == "" {
			break
		}
	}
	log.Debugf("search registered models successfully")
	return registeredModels, nil
}

func (c *MlflowClient) SetRegisteredModelTag(name, key, value string) error {
	req := &struct {
		Name  string `json:"name"`
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		Name:  name,
		Key:   key,
		Value: value,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqBody := string(reqBytes)

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		Post("api/2.0/mlflow/registered-models/set-tag")
	if err != nil {
		return fmt.Errorf("failed to set registered model tag of \"%v\": %v", name, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return fmt.Errorf("failed to set registered model tag of \"%v\": %v", name, strings.ToUpper(resp.Status()))
		}
		return fmt.Errorf("failed to set registered model tag of \"%v\": %v:  %v", name, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("set registered model tag of \"%v\" successfully", name)
	return nil
}

func (c *MlflowClient) SetModelVersionTag(name, version, key, value string) error {
	req := &struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Key     string `json:"key"`
		Value   string `json:"value"`
	}{
		Name:    name,
		Version: version,
		Key:     key,
		Value:   value,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqBody := string(reqBytes)

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		Post("api/2.0/mlflow/model-versions/set-tag")
	if err != nil {
		return fmt.Errorf("failed to set model version tag: %v", err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return fmt.Errorf("failed to set model version tag of \"%v/%v\": %v", name, version, strings.ToUpper(resp.Status()))
		}
		return fmt.Errorf("failed to set model version tag of \"%v/%v\": %v:  %v", name, version, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("set model version tag of \"%v/%v\" successfully", name, version)
	return nil
}

func (c *MlflowClient) DeleteRegisteredModelTag(name, key string) error {
	req := &struct {
		Name string `json:"name"`
		Key  string `json:"key"`
	}{
		Name: name,
		Key:  key,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqBody := string(reqBytes)

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/registered-models/delete-tag")
	if err != nil {
		return fmt.Errorf("failed to delete registered model tag: %v", err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return fmt.Errorf("failed to delete registered model tag of \"%v\": %v", name, strings.ToUpper(resp.Status()))
		}
		return fmt.Errorf("failed to delete registered model tag of \"%v\": %v:  %v", name, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("delete registered model tag of \"%v\" successfully", name)
	return nil
}

func (c *MlflowClient) DeleteModelVersionTag(name, version, key string) error {
	req := &struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Key     string `json:"key"`
	}{
		Name:    name,
		Version: version,
		Key:     key,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqBody := string(reqBytes)

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/model-versions/delete-tag")
	if err != nil {
		return fmt.Errorf("failed to delete model version tag: %v", err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return fmt.Errorf("failed to set model version tag of \"%v/%v\": %v", name, version, strings.ToUpper(resp.Status()))
		}
		return fmt.Errorf("failed to set model version tag of \"%v/%v\": %v:  %v", name, version, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("set model version tag of \"%v/%v\" successfully", name, version)
	return nil
}

func (c *MlflowClient) DeleteRegisteredModelAlias(name, alias string) error {
	req := &struct {
		Name  string `json:"name"`
		Alias string `json:"alias"`
	}{
		Name:  name,
		Alias: alias,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	reqBody := string(reqBytes)

	resp, err := c.RestyClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/registered-models/alias")
	if err != nil {
		return fmt.Errorf("failed to delete registered model alias \"%v@%v\": %v", name, alias, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return fmt.Errorf("failed to delete registered model alias \"%v@%v\": %v", name, alias, strings.ToUpper(resp.Status()))
		}
		return fmt.Errorf("failed to delete registered model alias \"%v@%v\": %v:  %v", name, alias, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("delete registered model alias \"%v@%v\" successfully", name, alias)
	return nil
}

func (c *MlflowClient) GetModelVersionByAlias(name, alias string) (*types.ModelVersion, error) {
	res := &struct {
		ModelVersion *types.ModelVersion `json:"model_version"`
	}{}

	resp, err := c.RestyClient.
		R().
		SetQueryParam("name", name).
		SetQueryParam("alias", alias).
		SetResult(res).
		Get("api/2.0/mlflow/registered-models/alias")
	if err != nil {
		return nil, fmt.Errorf("failed to get model version by alias \"%v@%v\": %v", name, alias, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return nil, fmt.Errorf("failed to get model version by alias \"%v@%v\": %v", name, alias, strings.ToUpper(resp.Status()))
		}
		return nil, fmt.Errorf("failed to get model version by alias \"%v@%v\": %v:  %v", name, alias, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("get model version by alias \"%v@%v\" successfully", name, alias)
	return res.ModelVersion, nil
}

func (c *MlflowClient) SetRegisteredModelAlias(name, version, alias string) error {
	req := struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Alias   string `json:"alias"`
	}{
		Name:    name,
		Version: version,
		Alias:   alias,
	}

	resp, err := c.RestyClient.
		R().
		SetBody(req).
		Post("api/2.0/mlflow/registered-models/alias")
	if err != nil {
		return fmt.Errorf("failed to set registered model alias \"%v\" to version \"%v\" of model \"%v\": %v", alias, version, name, err)
	}
	if resp.IsError() {
		mlflowError := &MlflowError{}
		if err := json.Unmarshal(resp.Body(), mlflowError); err != nil {
			return fmt.Errorf("failed to set registered model alias \"%v\" to version \"%v\" of model \"%v\": %v", alias, version, name, strings.ToUpper(resp.Status()))
		}
		return fmt.Errorf("failed to set registered model alias \"%v\" to version \"%v\" of model \"%v\": %v: %v", alias, version, name, strings.ToUpper(resp.Status()), mlflowError)
	}
	log.Debugf("set registered model alias \"%v\" to version \"%v\" of model \"%v\" successfully", alias, version, name)
	return nil
}

func (c *MlflowClient) CreateRegisteredModelAndModelVersion(
	name string,
	description string,
	tags []*types.RegisteredModelTag,
	version string,
	versionDescription string,
	versionTags []*types.ModelVersionTag,
	source string,
) (*types.RegisteredModel, *types.ModelVersion, error) {
	// Create a registered model if not exists
	var registeredModel *types.RegisteredModel
	var modelVersion *types.ModelVersion
	registeredModel, err := c.GetRegisteredModel(name)
	if err != nil {
		if !strings.Contains(err.Error(), RESOURCE_DOES_NOT_EXIST_ERROR) {
			return nil, nil, err
		}
	}
	if registeredModel == nil {
		registeredModel, err := c.CreateRegisteredModel(name, tags, description)
		// TODO: fix mlflow bug that deletes a registered model does not delete associated permissions when basic authentication is enabled
		if err != nil && !strings.Contains(err.Error(), RESOURCE_ALREADY_EXISTS_ERROR) {
			return registeredModel, nil, err
		}
		log.Infof("registered model \"%s\" created\n", name)
	}

	// Create a model version
	if version != "auto" {
		return registeredModel, modelVersion, fmt.Errorf("model version currently only supports `auto`")
	}
	modelVersion, err = c.CreateModelVersion(name, source, "", versionTags, "", versionDescription)
	if err != nil {
		return registeredModel, modelVersion, err
	}
	log.Infof("model version %s for \"%s\" created\n", modelVersion.Version, modelVersion.Name)
	return registeredModel, modelVersion, err
}

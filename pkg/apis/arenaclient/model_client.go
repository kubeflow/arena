package arenaclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
)

type ModelClient struct {
	namespace string
	configr   *config.ArenaConfiger

	MlflowClient
}

func NewModelClient(namespace string, configer *config.ArenaConfiger) (*ModelClient, error) {
	// Construct a MLflow client
	mlflowTrackingUri := os.Getenv("MLFLOW_TRACKING_URI")
	if mlflowTrackingUri == "" {
		return nil, errors.New("environment variable \"MLFLOW_TRACKING_URI\" must be specified")
	}
	mlflowTrackingUsername := os.Getenv("MLFLOW_TRACKING_USERNAME")
	mlflowTrackingPassword := os.Getenv("MLFLOW_TRACKING_PASSWORD")
	mlflowClient := NewMlflowClient(mlflowTrackingUri, mlflowTrackingUsername, mlflowTrackingPassword)
	modelClient := &ModelClient{
		namespace:    namespace,
		configr:      configer,
		MlflowClient: *mlflowClient,
	}
	return modelClient, nil
}

type MlflowClient struct {
	RestClient *resty.Client
}

func NewMlflowClient(trackingUri, username, password string) *MlflowClient {
	restClient := resty.New().
		SetBaseURL(trackingUri).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetDisableWarn(true)

	if username != "" && password != "" {
		restClient.SetBasicAuth(username, password)
	}

	return &MlflowClient{
		RestClient: restClient,
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Post("api/2.0/mlflow/registered-models/create")
	if err != nil {
		return nil, fmt.Errorf("failed to create registered model %v: %v", name, err)
	}
	if resp.IsError() {
		respBody := &struct {
			ErrorCode string `json:"error_code"`
			Message   string `json:"message"`
		}{}
		if err := json.Unmarshal(resp.Body(), respBody); err != nil {
			return nil, err
		}
		if resp.Status() == "401 UNAUTHORIZED" {
			return nil, fmt.Errorf("failed to create registered model \"%v\": environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", name, resp.Status())
		}
		if respBody.ErrorCode == "RESOURCE_ALREADY_EXISTS" {
			return nil, fmt.Errorf("failed to create registered model \"%v\": resource already exists", name)
		}
		return nil, fmt.Errorf("failed to create registered model \"%v\": %v", name, resp.Status())
	}
	return res.RegisteredModel, nil
}

func (c *MlflowClient) GetRegisteredModel(name string) (*types.RegisteredModel, error) {
	res := &struct {
		RegisteredModel *types.RegisteredModel `json:"registered_model"`
	}{}

	resp, err := c.RestClient.
		R().
		SetQueryParam("name", name).
		SetResult(res).
		Get("api/2.0/mlflow/registered-models/get")
	if err != nil {
		return nil, fmt.Errorf("failed to get registered model %v: %v", name, err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return nil, fmt.Errorf("failed to get registered model: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return nil, fmt.Errorf("failed to get registered model %v: %v", name, resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Post("api/2.0/mlflow/registered-models/rename")
	if err != nil {
		return nil, fmt.Errorf("failed to rename registered model %s to %s: %v", name, newName, err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return nil, fmt.Errorf("failed to rename registered model: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return nil, fmt.Errorf("failed to rename registered model %s to %s: %v", name, newName, resp.Status())
	}
	log.Infof("rename registered model \"%s\" to \"%s\"\n", name, newName)
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Patch("api/2.0/mlflow/registered-models/update")
	if err != nil {
		if resp.Status() == "401 UNAUTHORIZED" {
			return nil, fmt.Errorf("failed to update registered model: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return nil, fmt.Errorf("failed to update registered model %v: %v", name, err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to update registered model %v: %v", name, resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/registered-models/delete")
	if err != nil {
		return fmt.Errorf("failed to delete registered model %v: %v", name, err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return fmt.Errorf("failed to delete registered model: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return fmt.Errorf("failed to delete registered model %v: %v", name, resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Post("2.0/mlflow/registered-models/get-latest-versions")
	if err != nil {
		return nil, fmt.Errorf("failed to get latest versions of model %v: %v", name, err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return nil, fmt.Errorf("failed to get latest versions of model %v: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", name, resp.Status())
		}
		return nil, fmt.Errorf("failed to get latest versions of model %v: %v", name, resp.Status())
	}
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
	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Post("api/2.0/mlflow/model-versions/create")
	if err != nil {
		return nil, fmt.Errorf("failed to create registered model %v: %v", name, err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return nil, fmt.Errorf("failed to create registered model %v: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", name, resp.Status())
		}
		return nil, fmt.Errorf("failed to create registered model %v: %v", name, resp.Status())
	}
	return res.ModelVersion, nil
}

func (c *MlflowClient) GetModelVersion(name, version string) (*types.ModelVersion, error) {
	res := &struct {
		ModelVersion *types.ModelVersion `json:"model_version"`
	}{}

	resp, err := c.RestClient.
		R().
		SetQueryParam("name", name).
		SetQueryParam("version", version).
		SetResult(res).
		Get("api/2.0/mlflow/model-versions/get")
	if err != nil {
		return nil, fmt.Errorf("failed to get model version %v/%v: %v", name, version, err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return nil, fmt.Errorf("failed to get model version %v/%v: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", name, version, resp.Status())
		}
		return nil, fmt.Errorf("failed to get model version %v/%v: %v", name, version, resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Patch("api/2.0/mlflow/model-versions/update")
	if err != nil {
		return nil, fmt.Errorf("failed to update model version %v/%v: %v", name, version, err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return nil, fmt.Errorf("failed to update model version %v/%v: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", name, version, resp.Status())
		}
		return nil, fmt.Errorf("failed to update model version %v/%v: %v", name, version, resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/model-versions/delete")
	if err != nil {
		return fmt.Errorf("failed to delete model version %v/%v: %v", name, version, err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete model version %v/%v: %v", name, version, resp.Status())
	}
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
		resp, err := c.RestClient.
			R().
			SetResult(res).
			Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to search model version: %v", err)
		}
		if resp.IsError() {
			if resp.Status() == "401 UNAUTHORIZED" {
				return nil, fmt.Errorf("failed to search model version: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
			}
			return nil, fmt.Errorf("failed to search model version: %v", resp.Status())
		}
		modelVersions = append(modelVersions, res.ModelVersions...)
		nextPageToken = res.NextPageToken
		if nextPageToken == "" {
			break
		}
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		SetResult(res).
		Get("api/2.0/mlflow/model-versions/get-download-uri")
	if err != nil {
		return "", fmt.Errorf("failed to get artifacts download uri: %v", err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return "", fmt.Errorf("failed to get artifacts download uri: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return "", fmt.Errorf("failed to get artifacts download uri : %v", resp.Status())
	}
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
		resp, err := c.RestClient.
			R().
			SetResult(res).
			Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to search registered models: %v", err)
		}
		if resp.IsError() {
			if resp.Status() == "401 UNAUTHORIZED" {
				return nil, fmt.Errorf("failed to search registered models: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
			}
			return nil, fmt.Errorf("failed to search registered models: %v", resp.Status())
		}
		registeredModels = append(registeredModels, res.RegisteredModels...)
		nextPageToken = res.NextPageToken
		if nextPageToken == "" {
			break
		}
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		Post("api/2.0/mlflow/registered-models/set-tag")
	if err != nil {
		return fmt.Errorf("failed to set registered model tag: %v", err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return fmt.Errorf("failed to set registered model tag: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return fmt.Errorf("failed to set registered model tag: %v", resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		Post("api/2.0/mlflow/model-versions/set-tag")
	if err != nil {
		return fmt.Errorf("failed to set model version tag: %v", err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return fmt.Errorf("failed to set model version tag: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return fmt.Errorf("failed to set model version tag: %v", resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/registered-models/delete-tag")
	if err != nil {
		return fmt.Errorf("failed to delete registered model tag: %v", err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return fmt.Errorf("failed to delete registered model tag: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return fmt.Errorf("failed to delete registered model tag: %v", resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/model-versions/delete-tag")
	if err != nil {
		return fmt.Errorf("failed to delete model version tag: %v", err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return fmt.Errorf("failed to delete model version tag: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return fmt.Errorf("failed to delete model version tag: %v", resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(reqBody).
		Delete("api/2.0/mlflow/registered-models/alias")
	if err != nil {
		return fmt.Errorf("failed to delete registered model alias: %v", err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return fmt.Errorf("failed to delete registered model alias: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return fmt.Errorf("failed to delete registered model alias: %v", resp.Status())
	}
	return nil
}

func (c *MlflowClient) GetModelVersionByAlias(name, alias string) (*types.ModelVersion, error) {
	res := &struct {
		ModelVersion *types.ModelVersion `json:"model_version"`
	}{}

	resp, err := c.RestClient.
		R().
		SetQueryParam("name", name).
		SetQueryParam("alias", alias).
		SetResult(res).
		Get("api/2.0/mlflow/registered-models/alias")
	if err != nil {
		return nil, fmt.Errorf("failed to get model version by alias: %v", err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return nil, fmt.Errorf("failed to get model version by alias: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return nil, fmt.Errorf("failed to get model version by alias: %v", resp.Status())
	}
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

	resp, err := c.RestClient.
		R().
		SetBody(req).
		Post("api/2.0/mlflow/registered-models/alias")
	if err != nil {
		return fmt.Errorf("failed to set registered model alias: %v", err)
	}
	if resp.IsError() {
		if resp.Status() == "401 UNAUTHORIZED" {
			return fmt.Errorf("failed to set registered model alias: environment variables \"MLFLOW_TRACKING_USERNAME\" and \"MLFLOW_TRACKING_PASSWORD\" must be set properly: %v", resp.Status())
		}
		return fmt.Errorf("failed to set registered model alias: %v", resp.Status())
	}
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

	var exists bool
	registeredModel, err := c.GetRegisteredModel(name)
	if err != nil {
		return nil, nil, err
	}
	if registeredModel != nil {
		exists = true
	}
	if !exists {
		registeredModel, err := c.CreateRegisteredModel(name, tags, description)
		if err != nil && !strings.Contains(err.Error(), "resource already exists") {
			return registeredModel, nil, err
		}
		log.Infof("registered model \"%s\" created\n", name)
	}

	// Create a model version
	if version != "auto" {
		return registeredModel, modelVersion, fmt.Errorf("version currently only supports `auto`")
	}
	modelVersion, err = c.CreateModelVersion(name, source, "", versionTags, "", versionDescription)
	if err != nil {
		return registeredModel, modelVersion, err
	}
	log.Infof("model version %s for \"%s\" created\n", modelVersion.Version, modelVersion.Name)
	return registeredModel, modelVersion, err
}

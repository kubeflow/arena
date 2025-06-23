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

package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/kubeflow/arena/pkg/apis/types"
)

const (
	// tf-operator added labels for pods and servers.
	labelGroupName         = "group-name"
	labelGroupNameV1alpha2 = "group_name"

	// pytorchjob
	labelPyTorchGroupName = "group-name"

	// etjob
	etLabelGroupName = "group-name"

	// deepspeedjob
	deepspeedGroupName = "group-name"

	// training-operator
	OperatorNameLabel = "training.kubeflow.org/operator-name"
)

// GetTrainingJobTypes returns the supported training job types
func GetTrainingJobTypes() []types.TrainingJobType {
	trainingTypes := []types.TrainingJobType{}
	for trainingType := range types.TrainingTypeMap {
		trainingTypes = append(trainingTypes, trainingType)
	}
	return trainingTypes
}

func GetSupportTrainingJobTypesInfo() string {
	trainingTypes := []string{}
	for _, typeInfo := range types.TrainingTypeMap {
		item := fmt.Sprintf("%v(%v)", typeInfo.Shorthand, typeInfo.Alias)
		trainingTypes = append(trainingTypes, item)
	}
	return strings.Join(trainingTypes, ",")
}

// TransferTrainingJobType returns the training job type
func TransferTrainingJobType(jobType string) types.TrainingJobType {
	if jobType == "" {
		return types.AllTrainingJob
	}
	for trainingType, typeInfo := range types.TrainingTypeMap {
		if strings.EqualFold(string(typeInfo.Name), jobType) {
			return trainingType
		}
		if strings.EqualFold(typeInfo.Alias, jobType) {
			return trainingType
		}
		if strings.EqualFold(typeInfo.Shorthand, jobType) {
			return trainingType
		}
	}
	return types.UnknownTrainingJob
}

func GetSupportedNodeTypes() []string {
	items := []string{}
	for _, typeInfo := range types.NodeTypeSlice {
		items = append(items, fmt.Sprintf("%v(%v)", typeInfo.Shorthand, typeInfo.Alias))
	}
	return items
}

func TransferNodeType(nodeType string) types.NodeType {
	if nodeType == "" {
		return types.AllKnownNode
	}
	for _, typeInfo := range types.NodeTypeSlice {
		if strings.EqualFold(typeInfo.Alias, nodeType) {
			return typeInfo.Name
		}
		if string(typeInfo.Name) == nodeType {
			return typeInfo.Name
		}
		if typeInfo.Shorthand == nodeType {
			return typeInfo.Name
		}
	}
	return types.UnknownNode
}

func GetServingJobTypes() []types.ServingJobType {
	servingTypes := []types.ServingJobType{}
	for servingType := range types.ServingTypeMap {
		servingTypes = append(servingTypes, servingType)
	}
	return servingTypes
}

func GetSupportServingJobTypesInfo() string {
	servingTypes := []string{}
	for _, typeInfo := range types.ServingTypeMap {
		item := fmt.Sprintf("%v(%v)", typeInfo.Shorthand, typeInfo.Alias)
		servingTypes = append(servingTypes, item)
	}
	return strings.Join(servingTypes, ",")
}

func TransferServingJobType(jobType string) types.ServingJobType {
	if jobType == "" {
		return types.AllServingJob
	}
	for servingType, typeInfo := range types.ServingTypeMap {
		if strings.EqualFold(string(typeInfo.Name), jobType) {
			return servingType
		}
		if strings.EqualFold(typeInfo.Alias, jobType) {
			return servingType
		}
		if strings.EqualFold(typeInfo.Shorthand, jobType) {
			return servingType
		}
	}
	return types.UnknownServingJob
}

func GetLogLevel() []types.LogLevel {
	return []types.LogLevel{
		types.LogDebug,
		types.LogError,
		types.LogInfo,
		types.LogWarning,
	}
}
func TransferLogLevel(loglevel string) types.LogLevel {
	for _, knownLogLevel := range GetLogLevel() {
		if types.LogLevel(loglevel) == knownLogLevel {
			return knownLogLevel
		}
	}
	return types.LogUnknown
}
func GetFormatStyle() []types.FormatStyle {
	return []types.FormatStyle{
		types.JsonFormat,
		types.WideFormat,
		types.YamlFormat,
	}
}

func TransferPrintFormat(format string) types.FormatStyle {
	for _, knownFormat := range GetFormatStyle() {
		if types.FormatStyle(format) == knownFormat {
			return knownFormat
		}
	}
	return types.UnknownFormat
}

// print the help information
func PrintErrorMessage(message string) {
	if strings.Contains(message, "please use '--type' or '--version' to filter.") {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprint(w, message)
		w.Flush()
		return
	}
	if strings.Contains(message, "please use '-i' or '--instance' to filter.") {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprint(w, message)
		w.Flush()
		return
	}
	log.Errorf("%v", message)
}

func DefineNodeStatus(node *corev1.Node) string {
	conditionMap := make(map[corev1.NodeConditionType]*corev1.NodeCondition)
	NodeAllConditions := []corev1.NodeConditionType{corev1.NodeReady}
	for i := range node.Status.Conditions {
		cond := node.Status.Conditions[i]
		conditionMap[cond.Type] = &cond
	}
	var status []string
	for _, validCondition := range NodeAllConditions {
		if condition, ok := conditionMap[validCondition]; ok {
			if condition.Status == corev1.ConditionTrue {
				status = append(status, string(condition.Type))
			} else {
				status = append(status, "Not"+string(condition.Type))
			}
		}
	}
	if len(status) == 0 {
		status = append(status, "Unknown")
	}
	if node.Spec.Unschedulable {
		status = append(status, "SchedulingDisabled")
	}
	return strings.Join(status, ",")
}

func CheckFileExist(filename string) bool {
	var exist = true
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func DataUnitTransfer(from string, to string, value float64) float64 {
	knownUnits := []string{"bytes", "KiB", "MiB", "GiB", "TiB"}
	fromPosition := -1
	toPosition := -1
	for index, unit := range knownUnits {
		if unit == from {
			fromPosition = index
		}
		if unit == to {
			toPosition = index
		}
	}
	if fromPosition == -1 || toPosition == -1 {
		return value
	}
	return value * math.Pow(1024, float64(fromPosition-toPosition))
}

func ParseK8sObjectsFromYamlFile(filename string) ([]types.K8sObject, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	objects := []types.K8sObject{}
	for _, content := range strings.Split(string(data), "---\n") {
		object := types.K8sObject{}
		err := json.Unmarshal([]byte(content), &object)
		if err != nil {
			log.Debugf("failed to parse k8s object from yaml file %v,reason: %v", filename, err)
			return nil, err
		}
		objects = append(objects, object)
	}
	return objects, nil
}

func GetSupportModelJobTypesInfo() string {
	var modelJobTypes []string
	for _, typeInfo := range types.ModelTypeMap {
		item := fmt.Sprintf("%v(%v)", typeInfo.Shorthand, typeInfo.Alias)
		modelJobTypes = append(modelJobTypes, item)
	}
	return strings.Join(modelJobTypes, ",")
}

func TransferModelJobType(jobType string) types.ModelJobType {
	if jobType == "" {
		return types.AllModelJob
	}
	for modelJobType, typeInfo := range types.ModelTypeMap {
		if strings.EqualFold(string(typeInfo.Name), jobType) {
			return modelJobType
		}
		if strings.EqualFold(typeInfo.Alias, jobType) {
			return modelJobType
		}
		if strings.EqualFold(typeInfo.Shorthand, jobType) {
			return modelJobType
		}
	}
	return types.UnknownModelJob
}

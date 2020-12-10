package utils

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

const (
	// tf-operator added labels for pods and servers.
	labelGroupName         = "group-name"
	labelGroupNameV1alpha2 = "group_name"

	// pytorchjob
	labelPyTorchGroupName = "group-name"

	// etjob
	etLabelGroupName = "group-name"
)

// GetTrainingJobTypes returns the supported training job types
func GetTrainingJobTypes() []types.TrainingJobType {
	trainingTypes := []types.TrainingJobType{}
	for trainingType, _ := range types.TrainingTypeMap {
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
		if string(typeInfo.Name) == jobType {
			return trainingType
		}
		if strings.ToLower(typeInfo.Alias) == strings.ToLower(jobType) {
			return trainingType
		}
		if typeInfo.Shorthand == jobType {
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
		if strings.ToLower(typeInfo.Alias) == strings.ToLower(nodeType) {
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
		if string(typeInfo.Name) == jobType {
			return servingType
		}
		if strings.ToLower(typeInfo.Alias) == strings.ToLower(jobType) {
			return servingType
		}
		if typeInfo.Shorthand == jobType {
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

// print the help infomation
func PrintErrorMessage(message string) {
	if strings.Contains(message, "please use '--type' or '--version' to filter.") {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, message)
		w.Flush()
		return
	}
	if strings.Contains(message, "please use '-i' or '--instance' to filter.") {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, message)
		w.Flush()
		return
	}
	log.Errorf("%v", message)
}

func DefineNodeStatus(node *v1.Node) string {
	conditionMap := make(map[v1.NodeConditionType]*v1.NodeCondition)
	NodeAllConditions := []v1.NodeConditionType{v1.NodeReady}
	for i := range node.Status.Conditions {
		cond := node.Status.Conditions[i]
		conditionMap[cond.Type] = &cond
	}
	var status []string
	for _, validCondition := range NodeAllConditions {
		if condition, ok := conditionMap[validCondition]; ok {
			if condition.Status == v1.ConditionTrue {
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

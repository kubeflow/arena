// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package serving

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// job manager manages all jobs
type ServingJobManager interface {
	types.JobPrinter
	types.JobManager
}

// this is use to
type ServingJobFilterArgs struct {
	Namespace string
	Type      string
	Version   string
	Name      string
}

type manager struct {
	jobs []ServingJob
}

func NewServingJobManager(client *kubernetes.Clientset, ns string) (ServingJobManager, error) {
	// get all deployments with label "serviceName"
	jobs := []ServingJob{}
	deployments, err := client.AppsV1().Deployments(ns).List(metav1.ListOptions{
		LabelSelector: "serviceName",
	})
	if err != nil {
		return nil, fmt.Errorf("Failed due to %v", err)
	}
	// get all pods with label "serviceName"
	podListObject, err := client.CoreV1().Pods(ns).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: "serviceName",
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to get pods by label serviceName=,reason=%s", err.Error())
	}
	// get all service with label "serviceName"
	serviceList, err := client.CoreV1().Services(ns).List(metav1.ListOptions{
		LabelSelector: "servingName",
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to list services due to %v", err)
	}
	pods := podListObject.Items
	svcs := serviceList.Items
	for _, deploy := range deployments.Items {
		jobs = append(jobs, NewServingJob(deploy, pods, svcs))
	}
	return &manager{jobs: jobs}, nil
}

func (m *manager) GetAllJobs() []interface{} {
	jobs := []interface{}{}
	for _, job := range m.jobs {
		jobs = append(jobs, interface{}(job))
	}
	return jobs
}

func (m *manager) GetTargetJobs(args interface{}) ([]interface{}, error) {
	filter, ok := args.(ServingJobFilterArgs)
	if !ok {
		return nil, fmt.Errorf("type error: function want args with type ServingJobFilterArgs,but get %v", reflect.TypeOf(args))
	}
	jobs := []interface{}{}
	for _, job := range m.jobs {
		if !job.IsMatchedTargetType(filter.Type) {
			continue
		}
		if !job.IsMatchedTargetVersion(filter.Version) {
			continue
		}
		if !job.IsMatchedTargetNamespace(filter.Namespace) {
			continue
		}
		if !job.IsMatchedTargetName(filter.Name) {
			continue
		}
		jobs = append(jobs, interface{}(job))
	}
	return jobs, nil
}

// TODO: return manager information with json format
func (m *manager) GetJsonFormatString() (string, error) {
	return "", nil
}

// TODO: return mamanger information with yaml format
func (m *manager) GetYamlFormatString() (string, error) {
	return "", nil
}

// TODO: return manager information with wide format
func (m *manager) GetWideFormatString() (string, error) {
	return "", nil
}

// implement interface JobPrinter
func (m *manager) GetHelpInfo(objs ...interface{}) (string, error) {
	if len(objs) < 1 {
		return "", fmt.Errorf("you should give args for function GetHelpInfo")
	}
	obj := objs[0]
	jobs := obj.([]ServingJob)
	if len(jobs) == 0 {
		return "", types.ErrNotFoundJobs
	}
	header := fmt.Sprintf("There is %d jobs have been found:", len(jobs))
	tableHeader := "NAME\tTYPE\tVERSION"
	printLines := []string{tableHeader}
	footer := fmt.Sprintf("please use \"--type\" or \"--version\" to filter.")
	for _, job := range jobs {
		typ := job.GetType().(types.ServingType)
		line := fmt.Sprintf("%s\t%s\t%s",
			job.GetName(),
			string(typ),
			job.GetVersion(),
		)
		printLines = append(printLines, line)
	}
	helpInfo := fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(printLines, "\n"), footer)
	return helpInfo, nil
}

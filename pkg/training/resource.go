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

package training

import v1 "k8s.io/api/core/v1"

type Resource struct {
	Name         string
	Uid          string
	ResourceType ResourceType
}
type ResourceType string

const ResourceTypePod = ResourceType("Pod")
const ResourceTypeStatefulSet = ResourceType("StatefulSet")
const ResourceTypeJob = ResourceType("Job")

func podResources(pods []*v1.Pod) []Resource {
	resources := []Resource{}
	for _, pod := range pods {
		resources = append(resources, Resource{
			Name:         pod.Name,
			Uid:          string(pod.UID),
			ResourceType: ResourceTypePod,
		})
	}
	return resources
}

type BasicJobInfo struct {
	name      string
	resources []Resource
}

func (j *BasicJobInfo) Resources() []Resource {
	return j.resources
}

package types

import "k8s.io/api/core/v1"

type Resource struct {
	Name         string
	Uid          string
	ResourceType ResourceType
}
type ResourceType string

const ResourceTypePod = ResourceType("Pod")
const ResourceTypeStatefulSet = ResourceType("StatefulSet")
const ResourceTypeReplicaset = ResourceType("ReplicaSet")
const ResourceTypeJob = ResourceType("Job")

func PodResources(pods []v1.Pod) []Resource {
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

func NewBasicJobInfo(name string, resources []Resource) *BasicJobInfo {
	return &BasicJobInfo{
		name:      name,
		resources: resources,
	}
}

func (j *BasicJobInfo) Name() string {
	return j.name
}

func (j *BasicJobInfo) Resources() []Resource {
	return j.resources
}

func (j *BasicJobInfo) Project() string {
	return ""
}

func (j *BasicJobInfo) User() string {
	return ""
}

func (j *BasicJobInfo) Interactive() string {
	return ""
}

func (j *BasicJobInfo) Image() string {
	return ""
}

func (*BasicJobInfo) CreatedByCLI() bool {
	return false
}

func (*BasicJobInfo) ServiceURLs() []string {
	return []string{}
}

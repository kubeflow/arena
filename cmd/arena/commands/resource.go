package commands

import "k8s.io/api/core/v1"

type Resource struct {
	Name         string
	Uid          string
	ResourceType ResourceType
}
type ResourceType string

const ResourceTypePod = ResourceType("Pod")
const ResourceTypeStatefulSet = ResourceType("StatefulSet")
const ResourceTypeJob = ResourceType("Job")

func podResources(pods []v1.Pod) []Resource {
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

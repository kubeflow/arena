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

package types

type NodeType string

const (
	GPUShareNode     NodeType = "GPUShare"
	GPUExclusiveNode NodeType = "GPUExclusive"
	GPUTopologyNode  NodeType = "GPUTopology"
	NormalNode       NodeType = "Normal"
	UnknownNode      NodeType = "unknown"
	AllKnownNode     NodeType = ""
)

type NodeTypeInfo struct {
	Name      NodeType
	Alias     string
	Shorthand string
}

var NodeTypeSlice = []NodeTypeInfo{
	{
		Name:      NormalNode,
		Alias:     "none",
		Shorthand: "n",
	},
	{
		Name:      GPUExclusiveNode,
		Alias:     "exclusive",
		Shorthand: "e",
	},
	{
		Name:      GPUTopologyNode,
		Alias:     "topology",
		Shorthand: "t",
	},
	{
		Name:      GPUShareNode,
		Alias:     "share",
		Shorthand: "s",
	},
}

type CommonNodeInfo struct {
	Name        string   `json:"name"        yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	IP          string   `json:"ip"          yaml:"ip"`
	Status      string   `json:"status"      yaml:"status"`
	Role        string   `json:"role"        yaml:"role"`
	Type        NodeType `json:"type"        yaml:"type"`
}

/*
type CommonGPUNodeInfo struct {
	TotalGPUs           int             `json:"totalGPUs" yaml:"totalGPUs"`
	AllocatedGPUs       int             `json:"allocatedGPUs" yaml:"allocatedGPUs"`
	UnhealthyGPUs       int             `json:"unhealthyGPUs" yaml:"unhealthyGPUs"`
	TotalGPUMemory      float64         `json:"totalGPUMemory" yaml:"totalGPUMemory"`
	AllocatedGPUMemory  float64         `json:"allocatedGPUMemory" yaml:"allocatedGPUMemory"`
	UsedGPUMemory       float64         `json:"usedGPUMemory" yaml:"usedGPUMemory"`
	Devices             []GPUDeviceInfo `json:"devices" yaml:"devices"`
	DutyCycle           float64         `json:"dutyCycle" yaml:"dutyCycle"`
	GPUMetrics          NodeGpuMetric   `json:"gpuMetrics" yaml:"gpuMetrics"`
	GPUMetricsIsEnabled bool            `json:"gpuMetricsIsEnabled" yaml:"gpuMetricsIsEnabled"`
}
*/

type CommonGPUNodeInfo struct {
	TotalGPUs     float64              `json:"totalGPUs"     yaml:"totalGPUs"`
	AllocatedGPUs float64              `json:"allocatedGPUs" yaml:"allocatedGPUs"`
	UnhealthyGPUs float64              `json:"unhealthyGPUs" yaml:"unhealthyGPUs"`
	GPUMetrics    []*AdvancedGpuMetric `json:"gpuMetrics"    yaml:"gpuMetrics"`
}

type GPUDeviceInfo struct {
	ID                 string  `json:"id"                 yaml:"id"`
	TotalGPUMemory     float64 `json:"totalGPUMemory"     yaml:"totalGPUMemory"`
	AllocatedGPUMemory float64 `json:"allocatedGPUMemory" yaml:"allocatedGPUMemory"`
	UsedGPUMemory      float64 `json:"usedGPUMemory"      yaml:"usedGPUMemory"`
	DutyCycle          float64 `json:"dutyCycle"          yaml:"dutyCycle"`
}

type GPUShareNodeInfo struct {
	PodInfos           []GPUSharePodInfo    `json:"instances"          yaml:"instances"`
	TotalGPUMemory     float64              `json:"totalGPUMemory"     yaml:"totalGPUMemory"`
	AllocatedGPUMemory float64              `json:"allocatedGPUMemory" yaml:"allocatedGPUMemory"`
	TotalGPUCore       int64                `json:"totalGPUCore"       yaml:"totalGPUCore"`
	AllocatedGPUCore   int64                `json:"allocatedGPUCore"   yaml:"allocatedGPUCore"`
	Devices            []GPUShareNodeDevice `json:"devices"            yaml:"devices"`
	CommonGPUNodeInfo  `json:",inline"            yaml:",inline"`
	CommonNodeInfo     `json:",inline"            yaml:",inline"`
}

type GPUSharePodInfo struct {
	Name                string         `json:"name"                yaml:"name"`
	Namespace           string         `json:"namespace"           yaml:"namespace"`
	Status              string         `json:"status"              yaml:"status"`
	RequestMemory       int            `json:"requestGPUMemory"    yaml:"requestGPUMemory"`
	RequestCore         int            `json:"requestGPUCore"      yaml:"requestGPUCore"`
	GPUMemoryAllocation map[string]int `json:"gpuMemoryAllocation" yaml:"gpuMemoryAllocation"`
	GPUCoreAllocation   map[string]int `json:"gpuCoreAllocation"   yaml:"gpuCoreAllocation"`
}

type GPUShareNodeDevice struct {
	Id                 string  `json:"id"                 yaml:"id"`
	TotalGPUMemory     float64 `json:"totalGPUMemory"     yaml:"totalGPUMemory"`
	AllocatedGPUMemory float64 `json:"allocatedGPUMemory" yaml:"allocatedGPUMemory"`
	TotalGPUCore       int64   `json:"totalGPUCore"       yaml:"totalGPUCore"`
	AllocatedGPUCore   int64   `json:"allocatedGPUCore"   yaml:"allocatedGPUCore"`
}

type GPUExclusiveNodeInfo struct {
	PodInfos          []GPUExclusivePodInfo `json:"instances" yaml:"instances"`
	CommonNodeInfo    `json:",inline"   yaml:",inline"`
	CommonGPUNodeInfo `json:",inline"   yaml:",inline"`
}

type GPUExclusivePodInfo struct {
	Name       string `json:"name"        yaml:"name"`
	Namespace  string `json:"namespace"   yaml:"namespace"`
	Status     string `json:"status"      yaml:"status"`
	RequestGPU int    `json:"requestGPUs" yaml:"requestGPUs"`
}

type NormalNodeInfo struct {
	CommonNodeInfo `json:",inline" yaml:",inline"`
}

type AllNodeInfo map[string][]interface{}

type GPUTopologyNodeInfo struct {
	PodInfos          []GPUTopologyPodInfo `json:"instances"   yaml:"instances"`
	GPUTopology       GPUTopology          `json:"gpuTopology" yaml:"gpuTopology"`
	CommonGPUNodeInfo `json:",inline"     yaml:",inline"`
	CommonNodeInfo    `json:",inline"     yaml:",inline"`
	Devices           []GPUTopologyNodeDevice `json:"devices"     yaml:"devices"`
}

type GPUTopology struct {
	LinkMatrix      [][]string  `json:"linkMatrix"      yaml:"linkMatrix"`
	BandwidthMatrix [][]float32 `json:"bandwidthMatrix" yaml:"bandwidthMatrix"`
}

type GPUTopologyPodInfo struct {
	Name        string   `json:"name"        yaml:"name"`
	Namespace   string   `json:"namespace"   yaml:"namespace"`
	Status      string   `json:"status"      yaml:"status"`
	RequestGPU  int      `json:"requestGPUs" yaml:"requestGPUs"`
	Allocation  []string `json:"allocation"  yaml:"allocation"`
	VisibleGPUs []string `json:"visibleGPUs" yaml:"visibleGPUs"`
}

type GPUTopologyNodeDevice struct {
	Id      string `json:"id"      yaml:"id"`
	Healthy bool   `json:"healthy" yaml:"healthy"`
	Status  string `json:"status"  yaml:"status"`
}

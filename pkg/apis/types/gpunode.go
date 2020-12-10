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
	Name   string   `json:"name" yaml:"name"`
	IP     string   `json:"ip" yaml:"ip"`
	Status string   `json:"status" yaml:"status"`
	Role   string   `json:"role" yaml:"role"`
	Type   NodeType `json:"type" yaml:"type"`
}

type GPUShareNodeInfo struct {
	GPUCount        int                  `json:"gpuCount" yaml:"gpuCount"`
	TotalGPUMem     int                  `json:"totalGPUMemory" yaml:"totalGPUMemory"`
	UnHealthyGPUMem int                  `json:"unhealthyGPUMemory" yaml:"unhealthyGPUMemory"`
	UsedGPUMem      int                  `json:"usedGPUMemory" yaml:"usedGPUMemory"`
	Pods            []GPUSharePodInfo    `json:"instances" yaml:"instances"`
	Devices         []GPUShareDeviceInfo `json:"devices" yaml:"devices"`
	CommonNodeInfo  `yaml:",inline" json:",inline"`
}

type GPUShareDeviceInfo struct {
	ID          string `json:"id" yaml:"id"`
	IsHealthy   bool   `json:"-" yaml:"-"`
	UsedGPUMem  int    `json:"usedGPUMemory" yaml:"usedGPUMemory"`
	TotalGPUMem int    `json:"totalGPUMemory" yaml:"totalGPUMemory"`
}

type GPUSharePodInfo struct {
	Name          string         `json:"name" yaml:"name"`
	Namespace     string         `json:"namespace" yaml:"namespace"`
	RequestMemory int            `json:"requestGPUMemory" yaml:"requestGPUMemory"`
	Allocation    map[string]int `json:"allocation" yaml:"allocation"`
}

type GPUExclusiveNodeInfo struct {
	TotalGPUs      int                   `json:"totalGPUs" yaml:"totalGPUs"`
	UsedGPUs       int                   `json:"usedGPUs" yaml:"usedGPUs"`
	UnHealthyGPUs  int                   `json:"unhealthyGPUs" yaml:"unhealthyGPUs"`
	PodInfos       []GPUExclusivePodInfo `json:"instances" yaml:"instances"`
	CommonNodeInfo `yaml:",inline" json:",inline"`
}

type GPUExclusivePodInfo struct {
	Name       string `json:"name" yaml:"name"`
	Namespace  string `json:"namespace" yaml:"namespace"`
	RequestGPU int    `json:"requestGPUs" yaml:"requestGPUs"`
}

type NormalNodeInfo struct {
	CommonNodeInfo `yaml:",inline" json:",inline"`
}

type AllNodeInfo map[string][]interface{}

/*
type AllNodeInfo struct {
	GPUShareNodes     []GPUShareNodeInfo     `json:"gpushare_nodes" yaml:"gpushare_nodes"`
	GPUExclusiveNodes []GPUExclusiveNodeInfo `json:"gpu_exclusive_nodes" yaml:"gpu_exclusive_nodes"`
	NormalNodes       []NormalNodeInfo       `json:"normal_nodes" yaml:"normal_nodes"`
	GPUTopologyNodes  []GPUTopologyNodeInfo  `json:"gpu_topology_nodes" yaml:"gpu_topology_nodes"`
}
*/

type GPUTopologyNodeInfo struct {
	TotalGPUs      int                     `json:"totalGPUs" yaml:"totalGPUs"`
	UsedGPUs       int                     `json:"usedGPUs" yaml:"usedGPUs"`
	UnHealthyGPUs  int                     `json:"unhealthyGPUs" yaml:"unhealthyGPUs"`
	PodInfos       []GPUTopologyPodInfo    `json:"instances" yaml:"instances"`
	GPUTopology    GPUTopology             `json:"gpuTopology" yaml:"gpuTopology"`
	Devices        []GPUTopologyDeviceInfo `json:"devices" yaml:"devices"`
	CommonNodeInfo `yaml:",inline" json:",inline"`
}

type GPUTopology struct {
	LinkMatrix      [][]string  `json:"linkMatrix" yaml:"linkMatrix"`
	BandwidthMatrix [][]float32 `json:"bandwidthMatrix" yaml:"bandwidthMatrix"`
}

type GPUTopologyPodInfo struct {
	Name        string   `json:"name" yaml:"name"`
	Namespace   string   `json:"namespace" yaml:"namespace"`
	RequestGPU  int      `json:"requestGPUs" yaml:"requestGPUs"`
	Allocation  []string `json:"allocation" yaml:"allocation"`
	VisibleGPUs []string `json:"visibleGPUs" yaml:"visibleGPUs"`
}

type GPUTopologyDeviceInfo struct {
	Index   string `json:"gpuIndex" yaml:"gpuIndex"`
	Status  string `json:"status" yaml:"status"`
	Healthy bool   `json:"healthy" yaml:"healthy"`
}

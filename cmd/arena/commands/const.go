package commands

const (
	CHART_PKG_LOC = "CHARTREPO"
	// GPUResourceName is the extended name of the GPU resource since v1.8
	// this uses the device plugin mechanism
	NVIDIAGPUResourceName = "nvidia.com/gpu"

	DeprecatedNVIDIAGPUResourceName = "alpha.kubernetes.io/nvidia-gpu"

	masterLabelRole = "node-role.kubernetes.io/master"

	gangSchdName = "kube-batchd"

	// labelNodeRolePrefix is a label prefix for node roles
	// It's copied over to here until it's merged in core: https://github.com/kubernetes/kubernetes/pull/39112
	labelNodeRolePrefix = "node-role.kubernetes.io/"

	// nodeLabelRole specifies the role of a node
	nodeLabelRole = "kubernetes.io/role"
)

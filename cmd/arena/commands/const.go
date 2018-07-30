package commands

const (
	CHART_PKG_LOC = "CHARTREPO"
	// GPUResourceName is the extended name of the GPU resource since v1.8
	// this uses the device plugin mechanism
	NVIDIAGPUResourceName = "nvidia.com/gpu"

	DepricatedNVIDIAGPUResourceName = "alpha.kubernetes.io/nvidia-gpu"

	masterLabelRole = "node-role.kubernetes.io/master"
)

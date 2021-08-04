package types

const (
	CPUResourceName = "cpu"
)

const (
	// defines the nvidia resource name
	NvidiaGPUResourceName = "nvidia.com/gpu"
)

const (
	GPUShareResourceName    = "aliyun.com/gpu-mem"
	GPUShareCountName       = "aliyun.com/gpu-count"
	GPUShareEnvGPUID        = "ALIYUN_COM_GPU_MEM_IDX"
	GPUShareAllocationLabel = "scheduler.framework.gpushare.allocation"
	GPUShareNodeLabels      = "gpushare=true,cgpu=true,ack.node.gpu.schedule=share,ack.node.gpu.schedule=cgpu"
)

const (
	AliyunGPUResourceName      = "aliyun.com/gpu"
	GPUTopologyAllocationLabel = "topology.kubernetes.io/gpu-group"
	GPUTopologyVisibleGPULabel = "topology.kubernetes.io/gpu-visible"
	GPUTopologyNodeLabels      = "ack.node.gpu.schedule=topology"
)

const (
	MultiTenantIsolationLabel = "arena.kubeflow.org/isolate-user"
	UserNameIdLabel           = "arena.kubeflow.org/uid"
	UserNameNameLabel         = "arena.kubeflow.org/username"
)

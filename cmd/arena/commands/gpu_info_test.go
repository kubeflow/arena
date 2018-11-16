package commands

import (
	"testing"
	"github.com/kubeflow/arena/util"
)

func TestGetGpuInfo(t *testing.T) {
	clientset := util.GetClientSetForTest(t)
	if clientset == nil {
		t.Skip("kubeclient not setup")
	}
	gpuMetrics, err := GetGpuInfo(clientset, "prometheus-svc", `avg (nvidia_gpu_duty_cycle{pod_name=~"tensorrt-resnet50-744868bd7d-dl799", container_name!=""}) by (pod_name)`)
	if err != nil {
		t.Errorf("failed to get gpuInfo, err: %++v", err)
	}
	for _, m := range gpuMetrics{
		t.Logf("metric %++v", m)
	}
}

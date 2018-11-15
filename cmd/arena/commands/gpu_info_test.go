package commands

import (
	"testing"
	"github.com/kubeflow/arena/util"
)

func TestGetGpuInfo(t *testing.T) {
	clientset := util.GetClientSetForTest(t)
	gpuMetrics, err := GetGpuInfo(clientset, []string{"tensorrt-resnet50-744868bd7d-dl799"})
	if err != nil {
		t.Errorf("failed to get gpuInfo, err: %++v", err)
	}
	for _, m := range gpuMetrics{
		t.Logf("metric %++v", m)
	}
}

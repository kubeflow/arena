package commands

import (
	"fmt"
	"strings"
	"testing"

	"github.com/kubeflow/arena/pkg/util"
)

func TestQueryMetricByPrometheus(t *testing.T) {
	clientset := util.GetClientSetForTest(t)
	if clientset == nil {
		t.Skip("kubeclient not setup")
	}
	gpuMetrics, _ := QueryMetricByPrometheus(clientset, "prometheus-svc", KUBEFLOW_NAMESPACE, fmt.Sprintf(`{__name__=~"%s"}`, strings.Join(GPU_METRIC_LIST, "|")))

	for _, m := range gpuMetrics {
		t.Logf("metric %++v", m)
		t.Logf("metric name %s, value: %s", m.MetricName, m.Value)
	}
}

package commands

import (
	"context"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/arenacache"
	batchv1alpha1 "github.com/kubeflow/arena/pkg/operators/volcano-operator/apis/batch/v1alpha1"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"strings"
	"testing"
	"time"
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

func TestList(t *testing.T) {
	clientset := util.GetClientSetForTest(t)
	if clientset == nil {
		t.Skip("kubeclient not setup")
	}
	configer, err := config.InitArenaConfiger(types.ArenaClientArgs{})
	if err != nil {
		t.Error(err)
	}
	arenacache.InitCacheClient(configer.GetRestConfig())
	jobs := &batchv1alpha1.JobList{}

	for i := 0; i < 3; i++ {
		err := arenacache.GetCacheClient().List(context.Background(), jobs)
		log.Error(err)
		time.Sleep(10 * time.Second)
	}

}

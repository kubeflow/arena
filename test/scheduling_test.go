package integration

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/task"
)

func TestSchedulingUnimplemented(t *testing.T) {
	yamlPath := filepath.Join(testdataDir(t), "scheduling.yaml")
	taskObj, err := task.LoadFromFile(yamlPath)
	require.NoError(t, err, "LoadFromFile should succeed for scheduling.yaml")

	t.Skip("features not yet implemented: scheduling, affinity, init containers")

	// When features are implemented, assert the following on the CRD:
	//
	// Scheduling:
	//   - Gang scheduling: CRD metadata annotations include gang scheduling labels
	//   - Queue: CRD metadata annotations include queue name "high-priority"
	//   - PriorityClassName: podSpec["priorityClassName"] == "premium"
	//
	// Affinity:
	//   - podSpec["affinity"] contains node affinity rules matching
	//     matchExpressions: [{key: "gpu-type", operator: "In", values: ["A100"]}]
	//   - Policy "binpack" translated to appropriate K8s affinity fields
	//
	// Init containers:
	//   - podSpec["initContainers"] has 1 container named "setup-logs"
	//   - Image is "busybox", command runs "mkdir -p /workspace/logs"
	//
	// Environment variables:
	//   - container["env"] contains NCCL_DEBUG=INFO

	_ = taskObj // suppress unused variable when skip prevents assertions
}

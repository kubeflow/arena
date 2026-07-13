package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyOverrides_Identity(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"name":       "my-job",
		"namespace":  "ml-team",
		"label":      []string{"team=platform", "env=prod"},
		"annotation": []string{"note=experiment"},
	})
	assert.Equal(t, "my-job", task.Name)
	assert.Equal(t, "ml-team", task.Namespace)
	assert.Equal(t, "platform", task.Labels["team"])
	assert.Equal(t, "prod", task.Labels["env"])
	assert.Equal(t, "experiment", task.Annotations["note"])
}

func TestApplyOverrides_Workers(t *testing.T) {
	task := &Task{Worker: &Worker{Replicas: 1}}
	ApplyOverrides(task, map[string]interface{}{"workers": 8})
	assert.Equal(t, 8, task.Worker.Replicas)
}

func TestApplyOverrides_GPUs(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{"gpus": 4})
	require.NotNil(t, task.Worker, "Worker should be auto-created by GPU override")
	assert.Equal(t, "4", task.Worker.Resources["nvidia.com/gpu"])
}

func TestApplyOverrides_GPUType(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{"gpu-type": "A100"})
	assert.Equal(t, "A100", task.Scheduling.NodeSelector["nvidia.com/gpu.product"])
}

func TestApplyOverrides_CPUsAndMemory(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"cpus": "4",
		"mem":  "16Gi",
	})
	require.NotNil(t, task.Worker, "Worker should be auto-created by CPU/memory override")
	assert.Equal(t, "4", task.Worker.Resources["cpu"])
	assert.Equal(t, "16Gi", task.Worker.Resources["memory"])
}

func TestApplyOverrides_SHM(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{"shm": "8Gi"})
	require.Len(t, task.Storages, 1)
	assert.Equal(t, "shm", task.Storages[0].Name)
	assert.Equal(t, "8Gi", task.Storages[0].SHM)
	assert.Equal(t, "/dev/shm", task.Storages[0].MountPath)
}

func TestApplyOverrides_Envs(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"env": []string{"NCCL_DEBUG=INFO", "BATCH_SIZE=32"},
	})
	assert.Equal(t, "INFO", task.Envs["NCCL_DEBUG"].Value)
	assert.Equal(t, "32", task.Envs["BATCH_SIZE"].Value)
}

func TestApplyOverrides_Scheduling(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"priority":            100,
		"priority-class-name": "high",
		"gang":                true,
		"scheduler-name":      "volcano",
		"selector":            []string{"zone=us-west-2a"},
	})
	assert.Equal(t, 100, task.Scheduling.Priority)
	assert.Equal(t, "high", task.Scheduling.PriorityClassName)
	assert.True(t, task.Scheduling.Gang.Enabled)
	assert.Equal(t, "volcano", task.Scheduling.SchedulerName)
	assert.Equal(t, "us-west-2a", task.Scheduling.NodeSelector["zone"])
}

func TestApplyOverrides_Affinity(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"affinity-policy":     "required",
		"affinity-constraint": "node",
	})
	require.NotNil(t, task.Scheduling.Affinity)
	assert.Equal(t, "required", task.Scheduling.Affinity.Policy)
	assert.Equal(t, "node", task.Scheduling.Affinity.Constraint)
}

func TestApplyOverrides_Lifecycle(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"clean-pod-policy":   "Running",
		"active-deadline":    "2h",
		"ttl-after-finished": "7d",
		"backoff-limit":      3,
		"success-policy":     "ChiefWorker",
	})
	assert.Equal(t, "Running", task.Lifecycle.CleanPodPolicy)
	assert.Equal(t, "2h", task.Lifecycle.ActiveDeadline)
	assert.Equal(t, "7d", task.Lifecycle.TTLAfterFinished)
	assert.Equal(t, 3, *task.Lifecycle.BackoffLimit)
	assert.Equal(t, "ChiefWorker", task.Lifecycle.SuccessPolicy)
}

func TestApplyOverrides_Runtime(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"image-pull-policy": "IfNotPresent",
		"image-pull-secret": []string{"reg-secret"},
		"service-account":   "training-sa",
		"restart":           "OnFailure",
		"host-network":      true,
		"host-ipc":          true,
		"host-pid":          false,
	})
	assert.Equal(t, "IfNotPresent", task.ImagePullPolicy)
	assert.Equal(t, []string{"reg-secret"}, task.ImagePullSecrets)
	assert.Equal(t, "training-sa", task.ServiceAccount)
	assert.Equal(t, "OnFailure", task.Restart)
	assert.True(t, task.HostNetwork)
	assert.True(t, task.HostIPC)
	assert.False(t, task.HostPID)
}

func TestApplyOverrides_FrameworkOptions(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"nproc-per-node":     "auto",
		"slots-per-worker":   4,
		"gpu-topology":       true,
		"mounts-on-launcher": true,
	})
	assert.Equal(t, "auto", task.Framework.Options.NprocPerNode)
	assert.Equal(t, 4, task.Framework.Options.SlotsPerWorker)
	assert.True(t, task.Framework.Options.GPUTopology)
	assert.True(t, task.Framework.Options.MountsOnLauncher)
}

func TestApplyOverrides_Logging(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"tensorboard":        true,
		"tensorboard-logdir": "/logs",
		"tensorboard-image":  "custom/tb:latest",
	})
	require.NotNil(t, task.Logging.TensorBoard)
	assert.True(t, task.Logging.TensorBoard.Enabled)
	assert.Equal(t, "/logs", task.Logging.TensorBoard.LogDir)
	assert.Equal(t, "custom/tb:latest", task.Logging.TensorBoard.Image)
}

func TestApplyOverrides_Run(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"run": "python train.py --epochs 10",
	})
	assert.Equal(t, "python train.py --epochs 10", task.Run)
}

func TestApplyOverrides_Data(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"data": []string{"dataset:/data:dataset-pvc"},
	})
	require.Len(t, task.Storages, 1)
	assert.Equal(t, "dataset", task.Storages[0].Name)
	assert.Equal(t, "/data", task.Storages[0].MountPath)
	assert.Equal(t, "dataset-pvc", task.Storages[0].PVC)
}

func TestApplyOverrides_EmptyValuesIgnored(t *testing.T) {
	task := &Task{Name: "keep-me", Image: "keep:latest", Worker: &Worker{Replicas: 3}}
	ApplyOverrides(task, map[string]interface{}{
		"name":    "",
		"image":   "",
		"workers": 0,
	})
	assert.Equal(t, "keep-me", task.Name)
	assert.Equal(t, "keep:latest", task.Image)
	assert.Equal(t, 3, task.Worker.Replicas)
}

func TestApplyOverrides_WrongTypeIgnored(t *testing.T) {
	task := &Task{
		Name:   "original",
		Worker: &Worker{Replicas: 2},
	}

	// Pass wrong type for name (int instead of string)
	ApplyOverrides(task, map[string]interface{}{
		"name":    12345,
		"workers": "not-an-int",
	})

	assert.Equal(t, "original", task.Name)
	assert.Equal(t, 2, task.Worker.Replicas)
}

func TestApplyOverrides_Device(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"device": []string{"amd.com/gpu=1", "intel.com/fpga=2"},
	})
	require.NotNil(t, task.Worker, "Worker should be auto-created by device override")
	assert.Equal(t, "1", task.Worker.Resources["amd.com/gpu"])
	assert.Equal(t, "2", task.Worker.Resources["intel.com/fpga"])
}

func TestApplyOverrides_RoleSections(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"chief":     true,
		"evaluator": true,
		"ps-count":  2,
	})
	require.NotNil(t, task.Chief)
	require.NotNil(t, task.Evaluator)
	require.NotNil(t, task.PS)
	assert.Equal(t, 2, task.PS.Replicas)
}

func TestApplyOverrides_PSCountCreatesSection(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"ps-count": 3,
	})
	require.NotNil(t, task.PS)
	assert.Equal(t, 3, task.PS.Replicas)
}

func TestApplyOverrides_ChiefFalseNoSection(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"chief": false,
	})
	assert.Nil(t, task.Chief)
}

func TestApplyOverrides_EvaluatorFalseNoSection(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"evaluator": false,
	})
	assert.Nil(t, task.Evaluator)
}

func TestApplyOverrides_PSCountZeroNoSection(t *testing.T) {
	task := &Task{}
	ApplyOverrides(task, map[string]interface{}{
		"ps-count": 0,
	})
	assert.Nil(t, task.PS)
}

func TestApplyOverrides_RoleSectionsPreserveExisting(t *testing.T) {
	task := &Task{
		Chief: &RoleConfig{Replicas: 1},
		PS:    &RoleConfig{Replicas: 5},
	}
	ApplyOverrides(task, map[string]interface{}{
		"chief":    true,
		"ps-count": 3,
	})
	// Chief should still be non-nil (not overwritten)
	require.NotNil(t, task.Chief)
	assert.Equal(t, 1, task.Chief.Replicas)
	// PS replicas should be updated
	require.NotNil(t, task.PS)
	assert.Equal(t, 3, task.PS.Replicas)
}

func TestApplyOverrides_Tolerations(t *testing.T) {
	task := &Task{
		Name:  "test",
		Image: "test:latest",
		Run:   "echo hi",
	}

	flags := map[string]interface{}{
		"toleration": []string{"gpu=true:NoSchedule", "dedicated=ml:NoExecute"},
	}

	ApplyOverrides(task, flags)

	assert.Len(t, task.Scheduling.Tolerations, 2)

	// First toleration: gpu=true:NoSchedule
	assert.Equal(t, "gpu", task.Scheduling.Tolerations[0].Key)
	assert.Equal(t, "Equal", task.Scheduling.Tolerations[0].Operator)
	assert.Equal(t, "true", task.Scheduling.Tolerations[0].Value)
	assert.Equal(t, "NoSchedule", task.Scheduling.Tolerations[0].Effect)

	// Second toleration: dedicated=ml:NoExecute
	assert.Equal(t, "dedicated", task.Scheduling.Tolerations[1].Key)
	assert.Equal(t, "Equal", task.Scheduling.Tolerations[1].Operator)
	assert.Equal(t, "ml", task.Scheduling.Tolerations[1].Value)
	assert.Equal(t, "NoExecute", task.Scheduling.Tolerations[1].Effect)
}

func TestApplyOverrides_TolerationKeyOnly(t *testing.T) {
	task := &Task{
		Name:  "test",
		Image: "test:latest",
		Run:   "echo hi",
	}

	flags := map[string]interface{}{
		"toleration": []string{"node.kubernetes.io/not-ready:NoExecute"},
	}

	ApplyOverrides(task, flags)

	assert.Len(t, task.Scheduling.Tolerations, 1)
	assert.Equal(t, "node.kubernetes.io/not-ready", task.Scheduling.Tolerations[0].Key)
	assert.Equal(t, "Exists", task.Scheduling.Tolerations[0].Operator)
	assert.Equal(t, "", task.Scheduling.Tolerations[0].Value)
	assert.Equal(t, "NoExecute", task.Scheduling.Tolerations[0].Effect)
}

func TestApplyOverrides_TolerationInvalid(t *testing.T) {
	task := &Task{
		Name:  "test",
		Image: "test:latest",
		Run:   "echo hi",
	}

	flags := map[string]interface{}{
		"toleration": []string{""},
	}

	ApplyOverrides(task, flags)

	// Empty/invalid tolerations should be skipped
	assert.Empty(t, task.Scheduling.Tolerations)
}

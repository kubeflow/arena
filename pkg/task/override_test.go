package task

import (
	"testing"

	"github.com/kubeflow/arena/pkg/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyOverrides_Identity(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"name":       "my-job",
		"namespace":  "ml-team",
		"label":      []string{"team=platform", "env=prod"},
		"annotation": []string{"note=experiment"},
	})
	require.NoError(t, err)
	assert.Equal(t, "my-job", task.Name)
	assert.Equal(t, "ml-team", task.Namespace)
	assert.Equal(t, "platform", task.Labels["team"])
	assert.Equal(t, "prod", task.Labels["env"])
	assert.Equal(t, "experiment", task.Annotations["note"])
}

func TestApplyOverrides_Workers(t *testing.T) {
	task := &Task{Worker: &Worker{Replicas: 1}}
	err := ApplyOverrides(task, map[string]interface{}{"workers": 8})
	require.NoError(t, err)
	assert.Equal(t, 8, task.Worker.Replicas)
}

func TestApplyOverrides_GPUs(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{"gpus": 4})
	require.NoError(t, err)
	require.NotNil(t, task.Worker, "Worker should be auto-created by GPU override")
	assert.Equal(t, "4", task.Worker.Resources["nvidia.com/gpu"])
}

func TestApplyOverrides_GPUType(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{"gpu-type": "A100"})
	require.NoError(t, err)
	assert.Equal(t, "A100", task.Scheduling.NodeSelector["nvidia.com/gpu.product"])
}

func TestApplyOverrides_CPUAndMemory(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"cpu":    "4",
		"memory": "16Gi",
	})
	require.NoError(t, err)
	require.NotNil(t, task.Worker, "Worker should be auto-created by CPU/memory override")
	assert.Equal(t, "4", task.Worker.Resources["cpu"])
	assert.Equal(t, "16Gi", task.Worker.Resources["memory"])
}

func TestApplyOverrides_SHM(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{"share-memory": "8Gi"})
	require.NoError(t, err)
	require.Len(t, task.Storages, 1)
	assert.Equal(t, "shm", task.Storages[0].Name)
	assert.Equal(t, "8Gi", task.Storages[0].SHM)
	assert.Equal(t, "/dev/shm", task.Storages[0].MountPath)
}

func TestApplyOverrides_Envs(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"env": []string{"NCCL_DEBUG=INFO", "BATCH_SIZE=32"},
	})
	require.NoError(t, err)
	assert.Equal(t, "INFO", task.Envs["NCCL_DEBUG"].Value)
	assert.Equal(t, "32", task.Envs["BATCH_SIZE"].Value)
}

func TestApplyOverrides_Scheduling(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"priority":  "high",
		"gang":      true,
		"scheduler": "volcano",
		"selector":  []string{"zone=us-west-2a"},
	})
	require.NoError(t, err)
	assert.Equal(t, "high", task.Scheduling.PriorityClassName)
	assert.True(t, task.Scheduling.Gang.Enabled)
	assert.Equal(t, "volcano", task.Scheduling.SchedulerName)
	assert.Equal(t, "us-west-2a", task.Scheduling.NodeSelector["zone"])
}

func TestApplyOverrides_QueueSuspendsJob(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{"queue": true})
	require.NoError(t, err)
	require.NotNil(t, task.Lifecycle.Suspend, "v1 --queue should suspend the job")
	assert.True(t, *task.Lifecycle.Suspend)
}

func TestApplyOverrides_Affinity(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"affinity-policy":     "spread",
		"affinity-constraint": "required",
		"affinity-target":     "node",
	})
	require.NoError(t, err)
	require.NotNil(t, task.Scheduling.Affinity)
	assert.Equal(t, "spread", task.Scheduling.Affinity.Policy)
	assert.Equal(t, "required", task.Scheduling.Affinity.Constraint)
	assert.Equal(t, "node", task.Scheduling.Affinity.Target)
}

func TestApplyOverrides_Lifecycle(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"clean-task-policy":  "Running",
		"running-timeout":    "2h",
		"ttl-after-finished": "7d",
		"job-backoff-limit":  3,
		"success-policy":     constants.SuccessPolicyChiefWorkerAlias,
	})
	require.NoError(t, err)
	assert.Equal(t, "Running", task.Lifecycle.CleanPodPolicy)
	assert.Equal(t, "2h", task.Lifecycle.ActiveDeadline)
	assert.Equal(t, "7d", task.Lifecycle.TTLAfterFinished)
	assert.Equal(t, 3, *task.Lifecycle.BackoffLimit)
	assert.Equal(t, constants.SuccessPolicyChiefWorkerAlias, task.Lifecycle.SuccessPolicy)
}

func TestApplyOverrides_Runtime(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"image-pull-policy":  "IfNotPresent",
		"image-pull-secret":  []string{"reg-secret"},
		"service-account":    "training-sa",
		"job-restart-policy": "OnFailure",
		"hostNetwork":        true,
		"hostIPC":            true,
		"hostPID":            false,
	})
	require.NoError(t, err)
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
	err := ApplyOverrides(task, map[string]interface{}{
		"nproc-per-node":     "auto",
		"slots-per-worker":   4,
		"gputopology":        true,
		"mounts-on-launcher": true,
	})
	require.NoError(t, err)
	assert.Equal(t, "auto", task.Framework.Options.NprocPerNode)
	assert.Equal(t, 4, task.Framework.Options.SlotsPerWorker)
	assert.True(t, task.Framework.Options.GPUTopology)
	assert.True(t, task.Framework.Options.MountsOnLauncher)
}

func TestApplyOverrides_Logging(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"tensorboard":       true,
		"logdir":            "/logs",
		"tensorboard-image": "custom/tb:latest",
	})
	require.NoError(t, err)
	require.NotNil(t, task.Logging.TensorBoard)
	assert.True(t, task.Logging.TensorBoard.Enabled)
	assert.Equal(t, "/logs", task.Logging.TensorBoard.LogDir)
	assert.Equal(t, "custom/tb:latest", task.Logging.TensorBoard.Image)
}

func TestApplyOverrides_Run(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"run": "python train.py --epochs 10",
	})
	require.NoError(t, err)
	assert.Equal(t, "python train.py --epochs 10", task.Run)
}

func TestApplyOverrides_Data(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"data": []string{"dataset:/data:dataset-pvc"},
	})
	require.NoError(t, err)
	require.Len(t, task.Storages, 1)
	assert.Equal(t, "dataset", task.Storages[0].Name)
	assert.Equal(t, "/data", task.Storages[0].MountPath)
	assert.Equal(t, "dataset-pvc", task.Storages[0].PVC)
}

func TestApplyOverrides_EmptyValuesIgnored(t *testing.T) {
	task := &Task{Name: "keep-me", Image: "keep:latest", Worker: &Worker{Replicas: 3}}
	err := ApplyOverrides(task, map[string]interface{}{
		"name":    "",
		"image":   "",
		"workers": 0,
	})
	require.NoError(t, err)
	assert.Equal(t, "keep-me", task.Name)
	assert.Equal(t, "keep:latest", task.Image)
	assert.Equal(t, 3, task.Worker.Replicas)
}

func TestApplyOverrides_WrongTypeReturnsError(t *testing.T) {
	task := &Task{
		Name: "original",
	}

	// Pass wrong type for name (int instead of string)
	err := ApplyOverrides(task, map[string]interface{}{
		"name": 12345,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "flag \"name\"")
	assert.Contains(t, err.Error(), "expected string")

	// Pass wrong type for priority (int instead of string)
	err = ApplyOverrides(task, map[string]interface{}{
		"priority": 12345,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "flag \"priority\"")
	assert.Contains(t, err.Error(), "expected string")

	// Pass wrong type for run (int instead of string)
	err = ApplyOverrides(task, map[string]interface{}{
		"run": 12345,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "flag \"run\"")
	assert.Contains(t, err.Error(), "expected string")

	// Pass wrong type for job-backoff-limit (string instead of int)
	err = ApplyOverrides(task, map[string]interface{}{
		"job-backoff-limit": "not-an-int",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "flag \"job-backoff-limit\"")
	assert.Contains(t, err.Error(), "expected int")

	// Name should be unchanged
	assert.Equal(t, "original", task.Name)
}

func TestApplyOverrides_WrongBoolTypeReturnsError(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"hostNetwork": "yes",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected bool")
}

func TestApplyOverrides_Device(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"device": []string{"amd.com/gpu=1", "intel.com/fpga=2"},
	})
	require.NoError(t, err)
	require.NotNil(t, task.Worker, "Worker should be auto-created by device override")
	assert.Equal(t, "1", task.Worker.Resources["amd.com/gpu"])
	assert.Equal(t, "2", task.Worker.Resources["intel.com/fpga"])
}

func TestApplyOverrides_RoleSections(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"chief":     true,
		"evaluator": true,
		"ps":        2,
	})
	require.NoError(t, err)
	require.NotNil(t, task.Chief)
	require.NotNil(t, task.Evaluator)
	require.NotNil(t, task.PS)
	assert.Equal(t, 2, task.PS.Replicas)
}

func TestApplyOverrides_PSCreatesSection(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"ps": 3,
	})
	require.NoError(t, err)
	require.NotNil(t, task.PS)
	assert.Equal(t, 3, task.PS.Replicas)
}

func TestApplyOverrides_ChiefFalseNoSection(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"chief": false,
	})
	require.NoError(t, err)
	assert.Nil(t, task.Chief)
}

func TestApplyOverrides_EvaluatorFalseNoSection(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"evaluator": false,
	})
	require.NoError(t, err)
	assert.Nil(t, task.Evaluator)
}

func TestApplyOverrides_PSZeroNoSection(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{
		"ps": 0,
	})
	require.NoError(t, err)
	assert.Nil(t, task.PS)
}

func TestApplyOverrides_RoleSectionsPreserveExisting(t *testing.T) {
	task := &Task{
		Chief: &RoleConfig{Replicas: 1},
		PS:    &RoleConfig{Replicas: 5},
	}
	err := ApplyOverrides(task, map[string]interface{}{
		"chief": true,
		"ps":    3,
	})
	require.NoError(t, err)
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

	err := ApplyOverrides(task, flags)
	require.NoError(t, err)

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

	err := ApplyOverrides(task, flags)
	require.NoError(t, err)

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

	err := ApplyOverrides(task, flags)
	require.NoError(t, err)

	// Empty/invalid tolerations should be skipped
	assert.Empty(t, task.Scheduling.Tolerations)
}

func TestApplyOverrides_TolerationWithSeconds(t *testing.T) {
	task := &Task{
		Name:  "test",
		Image: "test:latest",
		Run:   "echo hi",
	}

	flags := map[string]interface{}{
		"toleration": []string{"gpu=true:NoExecute:300"},
	}

	err := ApplyOverrides(task, flags)
	require.NoError(t, err)

	require.Len(t, task.Scheduling.Tolerations, 1)
	tol := task.Scheduling.Tolerations[0]
	assert.Equal(t, "gpu", tol.Key)
	assert.Equal(t, "Equal", tol.Operator)
	assert.Equal(t, "true", tol.Value)
	assert.Equal(t, "NoExecute", tol.Effect)
	require.NotNil(t, tol.TolerationSeconds)
	assert.Equal(t, int64(300), *tol.TolerationSeconds)
}

func TestApplyOverrides_TolerationExistsWithSeconds(t *testing.T) {
	task := &Task{
		Name:  "test",
		Image: "test:latest",
		Run:   "echo hi",
	}

	flags := map[string]interface{}{
		"toleration": []string{"node.kubernetes.io/not-ready:NoExecute:60"},
	}

	err := ApplyOverrides(task, flags)
	require.NoError(t, err)

	require.Len(t, task.Scheduling.Tolerations, 1)
	tol := task.Scheduling.Tolerations[0]
	assert.Equal(t, "node.kubernetes.io/not-ready", tol.Key)
	assert.Equal(t, "Exists", tol.Operator)
	assert.Equal(t, "", tol.Value)
	assert.Equal(t, "NoExecute", tol.Effect)
	require.NotNil(t, tol.TolerationSeconds)
	assert.Equal(t, int64(60), *tol.TolerationSeconds)
}

func TestApplyOverrides_TolerationBackwardCompatible(t *testing.T) {
	task := &Task{
		Name:  "test",
		Image: "test:latest",
		Run:   "echo hi",
	}

	flags := map[string]interface{}{
		"toleration": []string{"gpu=true:NoSchedule"},
	}

	err := ApplyOverrides(task, flags)
	require.NoError(t, err)

	require.Len(t, task.Scheduling.Tolerations, 1)
	tol := task.Scheduling.Tolerations[0]
	assert.Equal(t, "gpu", tol.Key)
	assert.Equal(t, "NoSchedule", tol.Effect)
	assert.Nil(t, tol.TolerationSeconds)
}

func TestApplyOverrides_TolerationInvalidSeconds(t *testing.T) {
	task := &Task{
		Name:  "test",
		Image: "test:latest",
		Run:   "echo hi",
	}

	flags := map[string]interface{}{
		"toleration": []string{"gpu=true:NoExecute:abc"},
	}

	err := ApplyOverrides(task, flags)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "toleration seconds must be an integer")
}

func TestApplyOverrides_TolerationNegativeSeconds(t *testing.T) {
	task := &Task{
		Name:  "test",
		Image: "test:latest",
		Run:   "echo hi",
	}

	flags := map[string]interface{}{
		"toleration": []string{"gpu=true:NoExecute:-1"},
	}

	err := ApplyOverrides(task, flags)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-negative")
}

func TestApplyOverrides_TolerationExtraSegments(t *testing.T) {
	task := &Task{
		Name:  "test",
		Image: "test:latest",
		Run:   "echo hi",
	}

	flags := map[string]interface{}{
		"toleration": []string{"gpu=true:NoExecute:300:extra"},
	}

	err := ApplyOverrides(task, flags)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "toleration seconds must be an integer")
}

func TestApplyOverrides_PytorchSingleNodeGpus_RoutesToMaster(t *testing.T) {
	task := &Task{Master: &RoleConfig{}}
	err := ApplyOverrides(task, map[string]interface{}{"gpus": 1})
	require.NoError(t, err)
	assert.Nil(t, task.Worker, "Worker should remain nil for single-node PyTorch")
	require.NotNil(t, task.Master)
	assert.Equal(t, "1", task.Master.Resources["nvidia.com/gpu"])
}

func TestApplyOverrides_PytorchSingleNodeCPUMem_RoutesToMaster(t *testing.T) {
	task := &Task{Master: &RoleConfig{}}
	err := ApplyOverrides(task, map[string]interface{}{
		"cpu":    "4",
		"memory": "16Gi",
	})
	require.NoError(t, err)
	assert.Nil(t, task.Worker, "Worker should remain nil for single-node PyTorch")
	require.NotNil(t, task.Master)
	assert.Equal(t, "4", task.Master.Resources["cpu"])
	assert.Equal(t, "16Gi", task.Master.Resources["memory"])
}

func TestApplyOverrides_PytorchSingleNodeDevice_RoutesToMaster(t *testing.T) {
	task := &Task{Master: &RoleConfig{}}
	err := ApplyOverrides(task, map[string]interface{}{
		"device": []string{"amd.com/gpu=1"},
	})
	require.NoError(t, err)
	assert.Nil(t, task.Worker, "Worker should remain nil for single-node PyTorch")
	require.NotNil(t, task.Master)
	assert.Equal(t, "1", task.Master.Resources["amd.com/gpu"])
}

func TestApplyOverrides_MultiNodeGpus_RoutesToWorker(t *testing.T) {
	task := &Task{Worker: &Worker{Replicas: 2}}
	err := ApplyOverrides(task, map[string]interface{}{"gpus": 2})
	require.NoError(t, err)
	assert.Nil(t, task.Master, "Master should remain nil for multi-node")
	require.NotNil(t, task.Worker)
	assert.Equal(t, 2, task.Worker.Replicas)
	assert.Equal(t, "2", task.Worker.Resources["nvidia.com/gpu"])
}

func TestApplyOverrides_NoWorkerNoMaster_GpusCreatesWorkerWithReplicas1(t *testing.T) {
	task := &Task{}
	err := ApplyOverrides(task, map[string]interface{}{"gpus": 1})
	require.NoError(t, err)
	require.NotNil(t, task.Worker, "Worker should be auto-created")
	assert.Equal(t, 1, task.Worker.Replicas, "auto-created Worker should have Replicas=1")
	assert.Equal(t, "1", task.Worker.Resources["nvidia.com/gpu"])
}

func TestApplyOverrides_PerRoleResources(t *testing.T) {
	tk := &Task{Worker: &Worker{Replicas: 2}}
	err := ApplyOverrides(tk, map[string]interface{}{
		"ps": 1, "chief": true, "evaluator": true,
		"ps-cpu": "2", "ps-memory": "4Gi", "ps-gpus": 1,
		"chief-cpu": "4", "chief-memory": "16Gi",
		"evaluator-cpu": "1", "evaluator-memory": "2Gi",
		"worker-cpu": "8", "worker-memory": "32Gi",
	})
	require.NoError(t, err)
	require.NotNil(t, tk.PS)
	assert.Equal(t, "2", tk.PS.Resources["cpu"])
	assert.Equal(t, "4Gi", tk.PS.Resources["memory"])
	assert.Equal(t, "1", tk.PS.Resources["nvidia.com/gpu"])
	require.NotNil(t, tk.Chief)
	assert.Equal(t, "4", tk.Chief.Resources["cpu"])
	assert.Equal(t, "16Gi", tk.Chief.Resources["memory"])
	require.NotNil(t, tk.Evaluator)
	assert.Equal(t, "1", tk.Evaluator.Resources["cpu"])
	assert.Equal(t, "2Gi", tk.Evaluator.Resources["memory"])
	assert.Equal(t, "8", tk.Worker.Resources["cpu"])
	assert.Equal(t, "32Gi", tk.Worker.Resources["memory"])
}

func TestApplyOverrides_PerRoleBeatsGenericResources(t *testing.T) {
	tk := &Task{Worker: &Worker{Replicas: 2}}
	err := ApplyOverrides(tk, map[string]interface{}{
		"cpu":        "2",
		"worker-cpu": "8",
	})
	require.NoError(t, err)
	assert.Equal(t, "8", tk.Worker.Resources["cpu"], "worker-cpu must override the generic --cpu value")
}

func TestApplyOverrides_WorkerResourcesOnMasterOnlyPyTorch(t *testing.T) {
	tk := &Task{Master: &RoleConfig{}}
	err := ApplyOverrides(tk, map[string]interface{}{
		"worker-cpu":    "4",
		"worker-memory": "8Gi",
	})
	require.NoError(t, err)
	assert.Equal(t, "4", tk.Master.Resources["cpu"])
	assert.Equal(t, "8Gi", tk.Master.Resources["memory"])
}

func TestApplyOverrides_SyncCreatesSharedStorage(t *testing.T) {
	tk := &Task{Worker: &Worker{Replicas: 1}}
	err := ApplyOverrides(tk, map[string]interface{}{
		"sync-mode":   "git",
		"sync-source": "https://github.com/kubeflow/arena.git",
		"sync-image":  "alpine/git:latest",
	})
	require.NoError(t, err)

	require.Len(t, tk.Sync, 1)
	entry := tk.Sync[0]
	assert.Equal(t, "https://github.com/kubeflow/arena.git", entry.Git)
	assert.Equal(t, "alpine/git:latest", entry.Image)
	// v1 semantics: code lands in $workingDir/code (default /root)
	assert.Equal(t, "/root/code", entry.LocalPath)

	var codeSync *Storage
	for i := range tk.Storages {
		if tk.Storages[i].Name == "code-sync" {
			codeSync = &tk.Storages[i]
		}
	}
	require.NotNil(t, codeSync, "sync should create a code-sync storage")
	assert.Equal(t, entry.LocalPath, codeSync.MountPath,
		"code-sync must mount at the sync local_path so the main container sees the files")
	assert.Equal(t, "10Gi", codeSync.Tmp, "code-sync should be a size-limited emptyDir storage")
	require.Len(t, entry.Mounts, 1)
	assert.Equal(t, "code-sync", entry.Mounts[0].Name,
		"sync entry must mount code-sync so the init container writes into the volume")
}

func TestApplyOverrides_SyncModes(t *testing.T) {
	t.Run("rsync", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{
			"sync-mode":   "rsync",
			"sync-source": "10.88.29.56::backup/data.zip",
		})
		require.NoError(t, err)
		require.Len(t, tk.Sync, 1)
		assert.Equal(t, "10.88.29.56::backup/data.zip", tk.Sync[0].Rsync)
	})
	t.Run("hdfs", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{
			"sync-mode":   "hdfs",
			"sync-source": "hdfs://namenode:9000/data",
		})
		require.NoError(t, err)
		require.Len(t, tk.Sync, 1)
		assert.Equal(t, "hdfs://namenode:9000/data", tk.Sync[0].HDFS)
	})
}

func TestApplyOverrides_SyncLocalPathFollowsWorkingDir(t *testing.T) {
	tk := &Task{}
	err := ApplyOverrides(tk, map[string]interface{}{
		"sync-mode":   "git",
		"sync-source": "https://github.com/kubeflow/arena.git",
		"working-dir": "/workspace",
	})
	require.NoError(t, err)
	require.Len(t, tk.Sync, 1)
	assert.Equal(t, "/workspace/code", tk.Sync[0].LocalPath)
}

func TestApplyOverrides_SyncModeValidation(t *testing.T) {
	t.Run("missing source", func(t *testing.T) {
		err := ApplyOverrides(&Task{}, map[string]interface{}{"sync-mode": "git"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--sync-source is required when --sync-mode is set")
	})
	t.Run("invalid mode", func(t *testing.T) {
		err := ApplyOverrides(&Task{}, map[string]interface{}{
			"sync-mode":   "svn",
			"sync-source": "https://example.com/repo",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), `invalid --sync-mode "svn": must be git, rsync, or hdfs`)
	})
}

func TestApplyOverrides_GPUTopology(t *testing.T) {
	t.Run("implies host networking and topology labels", func(t *testing.T) {
		tk := &Task{Worker: &Worker{Replicas: 2}}
		err := ApplyOverrides(tk, map[string]interface{}{"gputopology": true})
		require.NoError(t, err)
		assert.True(t, tk.Framework.Options.GPUTopology)
		assert.True(t, tk.HostNetwork, "gputopology implies host networking (v1 semantics)")
		assert.Equal(t, "true", tk.Labels["gpu-topology"])
		assert.Equal(t, "true", tk.Labels["gpu-topology-replica"])
	})

	t.Run("beats conflicting user label", func(t *testing.T) {
		// User labels are applied before the gputopology branch, which sets
		// the topology labels unconditionally — so gputopology must win over
		// a user-supplied conflicting --label value.
		tk := &Task{Worker: &Worker{Replicas: 2}}
		err := ApplyOverrides(tk, map[string]interface{}{
			"label":       []string{"gpu-topology=false"},
			"gputopology": true,
		})
		require.NoError(t, err)
		assert.Equal(t, "true", tk.Labels["gpu-topology"],
			"gputopology must beat a conflicting user-supplied label")
		assert.Equal(t, "true", tk.Labels["gpu-topology-replica"])
	})
}

func TestApplyOverrides_V1DataSyntax(t *testing.T) {
	t.Run("v1 two-part data mounts the named pvc", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data": []string{"my-pvc:/data"}})
		require.NoError(t, err)
		require.Len(t, tk.Storages, 1)
		s := tk.Storages[0]
		assert.Equal(t, "my-pvc", s.Name)
		assert.Equal(t, "my-pvc", s.PVC)
		assert.Equal(t, "/data", s.MountPath)
	})

	t.Run("v2 three-part data unchanged", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data": []string{"training-data:/data:my-pvc"}})
		require.NoError(t, err)
		require.Len(t, tk.Storages, 1)
		s := tk.Storages[0]
		assert.Equal(t, "training-data", s.Name)
		assert.Equal(t, "my-pvc", s.PVC)
		assert.Equal(t, "/data", s.MountPath)
	})

	t.Run("one-part data errors", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data": []string{"justaname"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "justaname")
	})

	t.Run("relative mount path errors", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data": []string{"my-pvc:data"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})

	t.Run("v2 three-part data with relative mount path errors", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data": []string{"training-data:data:my-pvc"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})
}

func TestApplyOverrides_V1DataDirSyntax(t *testing.T) {
	t.Run("v1 one-part data-dir mounts hostpath at same path", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data-dir": []string{"/data"}})
		require.NoError(t, err)
		require.Len(t, tk.Storages, 1)
		s := tk.Storages[0]
		assert.Equal(t, "training-data-0", s.Name)
		assert.Equal(t, "/data", s.HostPath)
		assert.Equal(t, "/data", s.MountPath)
	})

	t.Run("v1 two-part data-dir splits host and container paths", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data-dir": []string{"/host/data:/container/data"}})
		require.NoError(t, err)
		require.Len(t, tk.Storages, 1)
		s := tk.Storages[0]
		assert.Equal(t, "training-data-0", s.Name)
		assert.Equal(t, "/host/data", s.HostPath)
		assert.Equal(t, "/container/data", s.MountPath)
	})

	t.Run("v2 three-part data-dir unchanged", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data-dir": []string{"d:/c:/host"}})
		require.NoError(t, err)
		require.Len(t, tk.Storages, 1)
		s := tk.Storages[0]
		assert.Equal(t, "d", s.Name)
		assert.Equal(t, "/host", s.HostPath)
		assert.Equal(t, "/c", s.MountPath)
	})

	t.Run("relative host path errors", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data-dir": []string{"data"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})

	t.Run("v2 three-part data-dir with relative mount path errors", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data-dir": []string{"d:c:/host"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})

	t.Run("v2 three-part data-dir with relative host path errors", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"data-dir": []string{"d:/c:host"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})
}

func TestApplyOverrides_V1ConfigFileSyntax(t *testing.T) {
	t.Run("v2 three-part configmap unchanged", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"config-file": []string{"cfg:/etc/app:my-cm"}})
		require.NoError(t, err)
		require.Len(t, tk.Storages, 1)
		s := tk.Storages[0]
		assert.Equal(t, "cfg", s.Name)
		assert.Equal(t, "my-cm", s.ConfigMap)
		assert.Equal(t, "/etc/app", s.MountPath)
	})

	t.Run("v2 three-part config-file with relative mount path errors", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"config-file": []string{"cfg:etc/app:my-cm"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})

	t.Run("v1 two-part host file form errors with configmap guidance", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"config-file": []string{"/host/file.conf:/etc/file.conf"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
		assert.Contains(t, err.Error(), "configmap")
	})

	t.Run("one-part config-file errors", func(t *testing.T) {
		tk := &Task{}
		err := ApplyOverrides(tk, map[string]interface{}{"config-file": []string{"/host/file.conf"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})
}

func TestApplyOverrides_PerRoleLimitFlags(t *testing.T) {
	tk := &Task{Framework: Framework{Name: "tensorflow"}, Worker: &Worker{Replicas: 1}}
	err := ApplyOverrides(tk, map[string]interface{}{
		"ps-cpu":           "1",
		"ps-cpu-limit":     "2",
		"worker-cpu":       "1",
		"worker-cpu-limit": "3",
	})
	require.NoError(t, err)

	require.NotNil(t, tk.PS)
	assert.Equal(t, Resources{"cpu": "1"}, tk.PS.Resources)
	assert.Equal(t, Resources{"cpu": "2"}, tk.PS.Limits)

	assert.Equal(t, Resources{"cpu": "1"}, tk.Worker.Resources)
	assert.Equal(t, Resources{"cpu": "3"}, tk.Worker.Limits)
}

func TestApplyOverrides_LimitFlagAlone(t *testing.T) {
	tk := &Task{Framework: Framework{Name: "tensorflow"}, Worker: &Worker{Replicas: 1}}
	err := ApplyOverrides(tk, map[string]interface{}{"worker-memory-limit": "16Gi"})
	require.NoError(t, err)
	assert.Empty(t, tk.Worker.Resources)
	assert.Equal(t, Resources{"memory": "16Gi"}, tk.Worker.Limits)
}

package provider

import (
	"testing"

	"github.com/kubeflow/arena/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPyTorchBuildCRD(t *testing.T) {
	tk := &task.Task{
		Name:  "pytorch-test",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
			Options: task.FrameworkConfig{
				NprocPerNode: "auto",
			},
		},
		Worker: &task.Worker{
			Replicas: 4,
			Resources: task.Resources{
				"cpu":            "2",
				"memory":         "8Gi",
				"nvidia.com/gpu": "1",
			},
		},
		Envs: map[string]task.EnvValue{
			"NCCL_DEBUG": {Value: "INFO"},
		},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "PyTorchJob", crd.GetKind())
	assert.Equal(t, "pytorch-test", crd.GetName())
	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())

	// Verify replica specs
	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	master := replicaSpecs["Master"].(map[string]interface{})
	assert.Equal(t, int64(1), master["replicas"])

	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(4), worker["replicas"]) // launch path: no N-1

	// Verify resources on worker container
	workerTemplate := worker["template"].(map[string]interface{})
	workerSpec := workerTemplate["spec"].(map[string]interface{})
	containers := workerSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	resources := container["resources"].(map[string]interface{})

	requests := resources["requests"].(map[string]interface{})
	assert.Equal(t, "1", requests["nvidia.com/gpu"])
	assert.Equal(t, "2", requests["cpu"])
	assert.Equal(t, "8Gi", requests["memory"])

	// limits == requests for Guaranteed QoS
	limits := resources["limits"].(map[string]interface{})
	assert.Equal(t, "1", limits["nvidia.com/gpu"])
	assert.Equal(t, "2", limits["cpu"])
	assert.Equal(t, "8Gi", limits["memory"])

	// Verify environment variables
	envVars := container["env"].([]interface{})
	require.Len(t, envVars, 1)
	envVar := envVars[0].(map[string]interface{})
	assert.Equal(t, "NCCL_DEBUG", envVar["name"])
	assert.Equal(t, "INFO", envVar["value"])
}

func TestPyTorchNprocPerNode(t *testing.T) {
	tk := &task.Task{
		Name:  "test",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
			Options: task.FrameworkConfig{
				NprocPerNode: "auto",
			},
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	// Verify nprocPerNode is a top-level spec field
	spec := crd.Object["spec"].(map[string]interface{})
	assert.Equal(t, "auto", spec["nprocPerNode"])
}

func TestPyTorchBuildCRDSingleReplica(t *testing.T) {
	tk := &task.Task{
		Name:  "single-replica",
		Image: "pytorch:1.13",
		Run:   "echo hello",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{Replicas: 1},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	master := replicaSpecs["Master"].(map[string]interface{})
	assert.Equal(t, int64(1), master["replicas"])

	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(1), worker["replicas"]) // launch path: no N-1
}

func TestPyTorchBuildCRDWithRunAndShell(t *testing.T) {
	tk := &task.Task{
		Name:  "with-cmd",
		Image: "pytorch:1.13",
		Run:   "python train.py --epochs 10",
		Shell: "/bin/bash",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
	master := replicaSpecs["Master"].(map[string]interface{})
	template := master["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})
	containers := podSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	assert.Equal(t, []interface{}{"/bin/bash", "-c"}, container["command"])
	assert.Equal(t, []interface{}{"python train.py --epochs 10"}, container["args"])
}

func TestPyTorchBuildCRDWithRunPolicy(t *testing.T) {
	backoff := 3
	tk := &task.Task{
		Name:  "with-policy",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{Replicas: 2},
		Lifecycle: task.Lifecycle{
			CleanPodPolicy:   "Running",
			TTLAfterFinished: "7d",
			BackoffLimit:     &backoff,
		},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	runPolicy := spec["runPolicy"].(map[string]interface{})
	assert.Equal(t, "Running", runPolicy["cleanPodPolicy"])
	assert.Equal(t, int64(7*24*3600), runPolicy["ttlSecondsAfterFinished"])
	assert.Equal(t, int64(3), runPolicy["backoffLimit"])
}

func TestPyTorchBuildCRDInvalidFramework(t *testing.T) {
	tk := &task.Task{
		Name:  "wrong-fw",
		Image: "tensorflow:2.0",
		Run:   "echo hello",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &PyTorchProvider{}
	_, err := provider.BuildCRD(tk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pytorch")
}

func TestPyTorchGetJobType(t *testing.T) {
	provider := &PyTorchProvider{}
	assert.Equal(t, "PyTorchJob", provider.GetJobType())
}

func TestPyTorchProviderImplementsInterface(_ *testing.T) {
	var _ Provider = &PyTorchProvider{}
}

func TestPyTorchBuildCRDWithVolumes(t *testing.T) {
	tk := &task.Task{
		Name:  "with-volumes",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{Replicas: 2},
		Storages: []task.Storage{
			{Name: "data", PVC: "my-pvc", MountPath: "/data"},
			{Name: "shm", SHM: "8Gi", MountPath: "/dev/shm"},
		},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
	master := replicaSpecs["Master"].(map[string]interface{})
	template := master["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})

	// Verify volumes exist
	volumes := podSpec["volumes"].([]interface{})
	require.Len(t, volumes, 2)

	// Verify volume mounts on container
	containers := podSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	mounts := container["volumeMounts"].([]interface{})
	require.Len(t, mounts, 2)
}

func TestPyTorchBuildCRDWithScheduling(t *testing.T) {
	tk := &task.Task{
		Name:  "with-scheduling",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{Replicas: 2},
		Scheduling: task.Scheduling{
			PriorityClassName: "high-priority",
			SchedulerName:     "volcano",
			NodeSelector:      map[string]string{"gpu": "true"},
		},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})
	master := replicaSpecs["Master"].(map[string]interface{})
	template := master["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})

	assert.Equal(t, "high-priority", podSpec["priorityClassName"])
	assert.Equal(t, "volcano", podSpec["schedulerName"])
	ns := podSpec["nodeSelector"].(map[string]interface{})
	assert.Equal(t, "true", ns["gpu"])
}

func TestPyTorchBuildCRDWithMasterOverrides(t *testing.T) {
	tk := &task.Task{
		Name:  "pytorch-master",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{
			Replicas: 4,
			Resources: task.Resources{
				"nvidia.com/gpu": "1",
			},
			Envs: map[string]task.EnvValue{
				"WORKER_VAR": {Value: "from-worker"},
			},
		},
		Envs: map[string]task.EnvValue{
			"NCCL_DEBUG": {Value: "INFO"},
		},
		Master: &task.RoleConfig{
			Resources: task.Resources{
				"nvidia.com/gpu": "2", // master gets more GPUs
			},
			Envs: map[string]task.EnvValue{
				"ROLE": {Value: "master"}, // master-only env
			},
		},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	// Master should have replicas=1 and custom resources
	master := replicaSpecs["Master"].(map[string]interface{})
	assert.Equal(t, int64(1), master["replicas"])

	masterTemplate := master["template"].(map[string]interface{})
	masterPodSpec := masterTemplate["spec"].(map[string]interface{})
	masterContainers := masterPodSpec["containers"].([]interface{})
	masterContainer := masterContainers[0].(map[string]interface{})
	masterRes := masterContainer["resources"].(map[string]interface{})
	masterReqs := masterRes["requests"].(map[string]interface{})
	assert.Equal(t, "2", masterReqs["nvidia.com/gpu"])

	// Master should have global envs + master envs, but NOT worker envs
	masterEnvs := masterContainer["env"].([]interface{})
	envMap := make(map[string]string)
	for _, e := range masterEnvs {
		env := e.(map[string]interface{})
		if val, ok := env["value"]; ok {
			envMap[env["name"].(string)] = val.(string)
		}
	}
	assert.Equal(t, "INFO", envMap["NCCL_DEBUG"], "global env should be present")
	assert.Equal(t, "master", envMap["ROLE"], "master env should be present")
	_, hasWorkerVar := envMap["WORKER_VAR"]
	assert.False(t, hasWorkerVar, "worker env should NOT be inherited by master")

	// Worker should have replicas=4 (not 3) and worker resources
	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(4), worker["replicas"])

	workerTemplate := worker["template"].(map[string]interface{})
	workerPodSpec := workerTemplate["spec"].(map[string]interface{})
	workerContainers := workerPodSpec["containers"].([]interface{})
	workerContainer := workerContainers[0].(map[string]interface{})
	workerRes := workerContainer["resources"].(map[string]interface{})
	workerReqs := workerRes["requests"].(map[string]interface{})
	assert.Equal(t, "1", workerReqs["nvidia.com/gpu"])
}

func TestPyTorchBuildCRDMasterInheritsWorker(t *testing.T) {
	// When Master section is absent, master inherits worker resources
	tk := &task.Task{
		Name:  "pytorch-no-master",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"nvidia.com/gpu": "1",
			},
		},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	// Master should inherit worker resources
	master := replicaSpecs["Master"].(map[string]interface{})
	masterTemplate := master["template"].(map[string]interface{})
	masterPodSpec := masterTemplate["spec"].(map[string]interface{})
	masterContainers := masterPodSpec["containers"].([]interface{})
	masterContainer := masterContainers[0].(map[string]interface{})
	masterRes := masterContainer["resources"].(map[string]interface{})
	masterReqs := masterRes["requests"].(map[string]interface{})
	assert.Equal(t, "1", masterReqs["nvidia.com/gpu"]) // inherited

	// Worker replicas should be 2 (no N-1 in launch path)
	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(2), worker["replicas"])
}

func TestPyTorchBuildCRDMasterOnly(t *testing.T) {
	tk := &task.Task{
		Name:  "pytorch-single",
		Image: "pytorch:2.1",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: nil,
		Master: &task.RoleConfig{
			Resources: task.Resources{
				"nvidia.com/gpu": "1",
				"cpu":            "4",
			},
			Envs: map[string]task.EnvValue{
				"ROLE": {Value: "master"},
			},
		},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	// Only Master should exist
	master := replicaSpecs["Master"].(map[string]interface{})
	assert.Equal(t, int64(1), master["replicas"])

	_, hasWorker := replicaSpecs["Worker"]
	assert.False(t, hasWorker, "Worker should not exist in master-only mode")

	// Master should have its own resources
	masterTemplate := master["template"].(map[string]interface{})
	masterPodSpec := masterTemplate["spec"].(map[string]interface{})
	masterContainers := masterPodSpec["containers"].([]interface{})
	masterContainer := masterContainers[0].(map[string]interface{})
	masterRes := masterContainer["resources"].(map[string]interface{})
	masterReqs := masterRes["requests"].(map[string]interface{})
	assert.Equal(t, "1", masterReqs["nvidia.com/gpu"])
	assert.Equal(t, "4", masterReqs["cpu"])
}

func TestPyTorchPerRoleRunOverride(t *testing.T) {
	tk := &task.Task{
		Name:  "pytorch-override",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{
			Replicas: 2,
			Run:      "python worker_train.py",
		},
		Master: &task.RoleConfig{
			Run: "python master_train.py",
		},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	// Master should use its own run override
	master := replicaSpecs["Master"].(map[string]interface{})
	masterTemplate := master["template"].(map[string]interface{})
	masterSpec := masterTemplate["spec"].(map[string]interface{})
	masterContainers := masterSpec["containers"].([]interface{})
	masterContainer := masterContainers[0].(map[string]interface{})
	assert.Equal(t, []interface{}{"python master_train.py"}, masterContainer["args"])

	// Worker should use its own run override
	worker := replicaSpecs["Worker"].(map[string]interface{})
	workerTemplate := worker["template"].(map[string]interface{})
	workerSpec := workerTemplate["spec"].(map[string]interface{})
	workerContainers := workerSpec["containers"].([]interface{})
	workerContainer := workerContainers[0].(map[string]interface{})
	assert.Equal(t, []interface{}{"python worker_train.py"}, workerContainer["args"])
}

func TestPyTorchMasterInheritsWorkerRun(t *testing.T) {
	tk := &task.Task{
		Name:  "pytorch-inherit",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{
			Replicas: 2,
			Run:      "python worker_train.py",
		},
		// Master not configured — inherits from Worker
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	// Master should inherit Worker's run
	master := replicaSpecs["Master"].(map[string]interface{})
	masterTemplate := master["template"].(map[string]interface{})
	masterSpec := masterTemplate["spec"].(map[string]interface{})
	masterContainers := masterSpec["containers"].([]interface{})
	masterContainer := masterContainers[0].(map[string]interface{})
	assert.Equal(t, []interface{}{"python worker_train.py"}, masterContainer["args"])

	// Worker should use its own run
	worker := replicaSpecs["Worker"].(map[string]interface{})
	workerTemplate := worker["template"].(map[string]interface{})
	workerSpec := workerTemplate["spec"].(map[string]interface{})
	workerContainers := workerSpec["containers"].([]interface{})
	workerContainer := workerContainers[0].(map[string]interface{})
	assert.Equal(t, []interface{}{"python worker_train.py"}, workerContainer["args"])
}

func TestPyTorchMasterOnlyRunOverride(t *testing.T) {
	tk := &task.Task{
		Name:  "pytorch-master-only",
		Image: "pytorch:1.13",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Master: &task.RoleConfig{
			Run: "python master_only.py",
		},
	}

	provider := &PyTorchProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["pytorchReplicaSpecs"].(map[string]interface{})

	master := replicaSpecs["Master"].(map[string]interface{})
	masterTemplate := master["template"].(map[string]interface{})
	masterSpec := masterTemplate["spec"].(map[string]interface{})
	masterContainers := masterSpec["containers"].([]interface{})
	masterContainer := masterContainers[0].(map[string]interface{})
	assert.Equal(t, []interface{}{"python master_only.py"}, masterContainer["args"])
}

func TestPyTorchBuildRBAC(t *testing.T) {
	provider := &PyTorchProvider{}
	resources, err := provider.BuildRBAC(&task.Task{}, metav1.OwnerReference{})
	require.NoError(t, err)
	assert.Nil(t, resources)
}

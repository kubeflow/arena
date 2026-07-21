package provider

import (
	"strings"
	"testing"

	"github.com/kubeflow/arena/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMPIBuildCRD(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-test",
		Image: "openmpi:4.1",
		Run:   "mpirun -np 4 ./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				SlotsPerWorker: 4,
			},
		},
		Worker: &task.Worker{
			Replicas: 4,
			Resources: task.Resources{
				"cpu":    "2",
				"memory": "8Gi",
			},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "MPIJob", crd.GetKind())
	assert.Equal(t, "mpi-test", crd.GetName())
	assert.Equal(t, "kubeflow.org/v2beta1", crd.GetAPIVersion())

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	// Launcher (1 replica)
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	assert.Equal(t, int64(1), launcher["replicas"])

	// Worker (4 replicas)
	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(4), worker["replicas"])

	// Verify slotsPerWorker
	assert.Equal(t, int64(4), spec["slotsPerWorker"])

	// Verify resources on worker container (requests == limits for Guaranteed QoS)
	workerTemplate := worker["template"].(map[string]interface{})
	workerSpec := workerTemplate["spec"].(map[string]interface{})
	containers := workerSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	resources := container["resources"].(map[string]interface{})

	requests := resources["requests"].(map[string]interface{})
	assert.Equal(t, "2", requests["cpu"])
	assert.Equal(t, "8Gi", requests["memory"])

	limits := resources["limits"].(map[string]interface{})
	assert.Equal(t, "2", limits["cpu"])
	assert.Equal(t, "8Gi", limits["memory"])
}

func TestMPIBuildCRDDefaultSlots(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-default-slots",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "kubeflow.org/v2beta1", crd.GetAPIVersion())

	spec := crd.Object["spec"].(map[string]interface{})
	// Default slotsPerWorker should be 1
	assert.Equal(t, int64(1), spec["slotsPerWorker"])
}

func TestMPIBuildCRDWithRunAndShell(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-cmd",
		Image: "openmpi:4.1",
		Run:   "mpirun -np 4 ./train",
		Shell: "/bin/bash",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "kubeflow.org/v2beta1", crd.GetAPIVersion())

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	template := launcher["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})
	containers := podSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	assert.Equal(t, []interface{}{"/bin/bash", "-c"}, container["command"])
	assert.Equal(t, []interface{}{"mpirun -np 4 ./train"}, container["args"])
}

func TestMPIBuildCRDWithEnv(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-env",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{Replicas: 1},
		Envs: map[string]task.EnvValue{
			"OMPI_MCA_btl": {Value: "self,tcp"},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	template := launcher["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})
	containers := podSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	envVars := container["env"].([]interface{})
	require.Len(t, envVars, 1)
	envVar := envVars[0].(map[string]interface{})
	assert.Equal(t, "OMPI_MCA_btl", envVar["name"])
	assert.Equal(t, "self,tcp", envVar["value"])
}

func TestMPIBuildCRDInvalidFramework(t *testing.T) {
	tk := &task.Task{
		Name:  "wrong-fw",
		Image: "tensorflow:2.0",
		Run:   "echo hello",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	_, err := provider.BuildCRD(tk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mpi")
}

func TestMPIGetJobType(t *testing.T) {
	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	assert.Equal(t, "MPIJob", provider.GetJobType())
}

func TestMPIProviderImplementsInterface(_ *testing.T) {
	var _ Provider = &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
}

func TestMPIBuildCRDGPUTopology(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-topo",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				GPUTopology: true,
			},
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	meta := crd.Object["metadata"].(map[string]interface{})
	annotations := meta["annotations"].(map[string]interface{})
	assert.Equal(t, "true", annotations["mpi.kubeflow.org/gpu-topology"])
}

func TestMPIBuildCRDLauncherCPUOnly(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-launcher-cpu",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"nvidia.com/gpu": "2",
				"cpu":            "4",
			},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	// Launcher should NOT have GPU resources (CPU-only by default)
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	launcherTemplate := launcher["template"].(map[string]interface{})
	launcherPodSpec := launcherTemplate["spec"].(map[string]interface{})
	launcherContainers := launcherPodSpec["containers"].([]interface{})
	launcherContainer := launcherContainers[0].(map[string]interface{})

	// Launcher should have no resources (or nil resources)
	if res, ok := launcherContainer["resources"]; ok {
		resMap := res.(map[string]interface{})
		reqs := resMap["requests"].(map[string]interface{})
		_, hasGPU := reqs["nvidia.com/gpu"]
		assert.False(t, hasGPU, "launcher should not have GPU resources by default")
	}

	// Worker SHOULD have GPU resources
	worker := replicaSpecs["Worker"].(map[string]interface{})
	workerTemplate := worker["template"].(map[string]interface{})
	workerPodSpec := workerTemplate["spec"].(map[string]interface{})
	workerContainers := workerPodSpec["containers"].([]interface{})
	workerContainer := workerContainers[0].(map[string]interface{})
	workerRes := workerContainer["resources"].(map[string]interface{})
	workerReqs := workerRes["requests"].(map[string]interface{})
	assert.Equal(t, "2", workerReqs["nvidia.com/gpu"])
}

func TestMPIBuildCRDRunLauncherAsWorker(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-launcher-as-worker",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				RunLauncherAsWorker: true,
			},
		},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"nvidia.com/gpu": "2",
			},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	// Launcher SHOULD have GPU resources when run_launcher_as_worker=true
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	launcherTemplate := launcher["template"].(map[string]interface{})
	launcherPodSpec := launcherTemplate["spec"].(map[string]interface{})
	launcherContainers := launcherPodSpec["containers"].([]interface{})
	launcherContainer := launcherContainers[0].(map[string]interface{})
	launcherRes := launcherContainer["resources"].(map[string]interface{})
	launcherReqs := launcherRes["requests"].(map[string]interface{})
	assert.Equal(t, "2", launcherReqs["nvidia.com/gpu"])
}

func TestMPIBuildCRDMountsOnLauncher(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-mounts",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				MountsOnLauncher: true,
			},
		},
		Worker: &task.Worker{Replicas: 2},
		Storages: []task.Storage{
			{Name: "data", PVC: "my-pvc", MountPath: "/data"},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	// Launcher should have volumes when mounts_on_launcher=true
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	launcherTemplate := launcher["template"].(map[string]interface{})
	launcherPodSpec := launcherTemplate["spec"].(map[string]interface{})
	volumes, ok := launcherPodSpec["volumes"]
	assert.True(t, ok, "launcher should have volumes when mounts_on_launcher=true")
	if ok {
		assert.Len(t, volumes.([]interface{}), 1)
	}
}

func TestMPIBuildCRDSuspendAndManagedBy(t *testing.T) {
	suspendVal := true
	tk := &task.Task{
		Name:  "mpi-suspend",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{Replicas: 2},
		Lifecycle: task.Lifecycle{
			Suspend:   &suspendVal,
			ManagedBy: "kueue.x-k8s.io/multikueue",
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	runPolicy := spec["runPolicy"].(map[string]interface{})
	assert.Equal(t, true, runPolicy["suspend"])
	assert.Equal(t, "kueue.x-k8s.io/multikueue", runPolicy["managedBy"])
}

func TestMPIBuildCRDV2beta1Default(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-v2beta1",
		Image: "openmpi:4.1",
		Run:   "mpirun -np 4 ./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				SlotsPerWorker: 4,
			},
		},
		Worker: &task.Worker{
			Replicas: 4,
			Resources: task.Resources{
				"cpu":    "2",
				"memory": "8Gi",
			},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "kubeflow.org/v2beta1", crd.GetAPIVersion())
	assert.Equal(t, "MPIJob", crd.GetKind())
	assert.Equal(t, "mpi-v2beta1", crd.GetName())

	spec := crd.Object["spec"].(map[string]interface{})
	assert.Equal(t, int64(4), spec["slotsPerWorker"])
	assert.Equal(t, false, spec["runLauncherAsWorker"])
	assert.Equal(t, mpiImplementationOpenMPI, spec["mpiImplementation"])
	assert.Equal(t, launcherCreationPolicyAtStartup, spec["launcherCreationPolicy"])
	assert.Equal(t, mpiDefaultSSHAuthMountPath, spec["sshAuthMountPath"])

	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})
	launcher := replicaSpecs[mpiReplicaTypeLauncher].(map[string]interface{})
	assert.Equal(t, int64(1), launcher["replicas"])
	worker := replicaSpecs[mpiReplicaTypeWorker].(map[string]interface{})
	assert.Equal(t, int64(4), worker["replicas"])
}

func TestMPIBuildCRDV2beta1IntelMPI(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-intel",
		Image: "intelmpi:2021",
		Run:   "mpirun -np 8 ./simulate",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				MPIImplementation:      "Intel",
				LauncherCreationPolicy: "WaitForWorkersReady",
				SSHAuthMountPath:       "/custom/.ssh",
				SlotsPerWorker:         2,
			},
		},
		Worker: &task.Worker{Replicas: 8},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "kubeflow.org/v2beta1", crd.GetAPIVersion())

	spec := crd.Object["spec"].(map[string]interface{})
	assert.Equal(t, mpiImplementationIntel, spec["mpiImplementation"])
	assert.Equal(t, launcherCreationPolicyWaitForWorkersReady, spec["launcherCreationPolicy"])
	assert.Equal(t, "/custom/.ssh", spec["sshAuthMountPath"])
	assert.Equal(t, int64(2), spec["slotsPerWorker"])
}

func TestMPIBuildCRDV2beta1RunLauncherAsWorker(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-v2beta1-launcher",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				RunLauncherAsWorker: true,
			},
		},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"nvidia.com/gpu": "2",
			},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	assert.Equal(t, true, spec["runLauncherAsWorker"])

	// Launcher SHOULD have GPU resources when runLauncherAsWorker=true
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})
	launcher := replicaSpecs[mpiReplicaTypeLauncher].(map[string]interface{})
	launcherTemplate := launcher["template"].(map[string]interface{})
	launcherPodSpec := launcherTemplate["spec"].(map[string]interface{})
	launcherContainers := launcherPodSpec["containers"].([]interface{})
	launcherContainer := launcherContainers[0].(map[string]interface{})
	launcherRes := launcherContainer["resources"].(map[string]interface{})
	launcherReqs := launcherRes["requests"].(map[string]interface{})
	assert.Equal(t, "2", launcherReqs["nvidia.com/gpu"])
}

func TestMPIBuildCRDV2beta1GPUTopology(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-v2beta1-topo",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				GPUTopology: true,
			},
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	meta := crd.Object["metadata"].(map[string]interface{})
	annotations := meta["annotations"].(map[string]interface{})
	assert.Equal(t, "true", annotations["mpi.kubeflow.org/gpu-topology"])
}

func TestMPIBuildCRDV2beta1SuspendManagedBy(t *testing.T) {
	suspendVal := false
	tk := &task.Task{
		Name:  "mpi-v2beta1-policy",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{Replicas: 2},
		Lifecycle: task.Lifecycle{
			Suspend:   &suspendVal,
			ManagedBy: "kueue.x-k8s.io/multikueue",
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	runPolicy := spec["runPolicy"].(map[string]interface{})
	assert.Equal(t, false, runPolicy["suspend"])
	assert.Equal(t, "kueue.x-k8s.io/multikueue", runPolicy["managedBy"])
}

func TestMPIBuildCRDWithLauncherOverrides(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-launcher",
		Image: "openmpi:4.1",
		Run:   "mpirun ./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{
			Replicas: 4,
			Resources: task.Resources{
				"nvidia.com/gpu": "2",
			},
		},
		Launcher: &task.RoleConfig{
			Resources: task.Resources{
				"cpu":    "4",
				"memory": "8Gi",
			},
			Envs: map[string]task.EnvValue{
				"LAUNCHER_MODE": {Value: "coordinator"},
			},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	// Launcher should have custom resources
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	assert.Equal(t, int64(1), launcher["replicas"])

	launcherTemplate := launcher["template"].(map[string]interface{})
	launcherPodSpec := launcherTemplate["spec"].(map[string]interface{})
	launcherContainers := launcherPodSpec["containers"].([]interface{})
	launcherContainer := launcherContainers[0].(map[string]interface{})
	launcherRes := launcherContainer["resources"].(map[string]interface{})
	launcherReqs := launcherRes["requests"].(map[string]interface{})
	assert.Equal(t, "4", launcherReqs["cpu"])
	assert.Equal(t, "8Gi", launcherReqs["memory"])

	// Launcher should have custom env
	launcherEnvs := launcherContainer["env"].([]interface{})
	require.Len(t, launcherEnvs, 1)
	envVar := launcherEnvs[0].(map[string]interface{})
	assert.Equal(t, "LAUNCHER_MODE", envVar["name"])
	assert.Equal(t, "coordinator", envVar["value"])

	// Worker should still have GPU resources (unchanged)
	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(4), worker["replicas"])
	workerTemplate := worker["template"].(map[string]interface{})
	workerPodSpec := workerTemplate["spec"].(map[string]interface{})
	workerContainers := workerPodSpec["containers"].([]interface{})
	workerContainer := workerContainers[0].(map[string]interface{})
	workerRes := workerContainer["resources"].(map[string]interface{})
	workerReqs := workerRes["requests"].(map[string]interface{})
	assert.Equal(t, "2", workerReqs["nvidia.com/gpu"])
}

func TestMPIBuildCRDWithLauncherOverridesNoMerge(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-launcher-no-merge",
		Image: "openmpi:4.1",
		Run:   "mpirun ./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"nvidia.com/gpu": "1",
			},
			Envs: map[string]task.EnvValue{
				"WORKER_VAR": {Value: "from-worker"},
				"SHARED_VAR": {Value: "worker-val"},
			},
		},
		Launcher: &task.RoleConfig{
			Resources: task.Resources{
				"cpu": "2",
			},
			Envs: map[string]task.EnvValue{
				"LAUNCHER_VAR": {Value: "from-launcher"},
				"SHARED_VAR":   {Value: "launcher-val"},
			},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	launcherTemplate := launcher["template"].(map[string]interface{})
	launcherPodSpec := launcherTemplate["spec"].(map[string]interface{})
	launcherContainers := launcherPodSpec["containers"].([]interface{})
	launcherContainer := launcherContainers[0].(map[string]interface{})
	launcherEnvs := launcherContainer["env"].([]interface{})

	envMap := map[string]string{}
	for _, e := range launcherEnvs {
		env := e.(map[string]interface{})
		if val, ok := env["value"]; ok {
			envMap[env["name"].(string)] = val.(string)
		}
	}

	assert.Equal(t, "from-launcher", envMap["LAUNCHER_VAR"])
	assert.Equal(t, "launcher-val", envMap["SHARED_VAR"])
	_, hasWorkerVar := envMap["WORKER_VAR"]
	assert.False(t, hasWorkerVar, "launcher should NOT inherit worker envs")
}

func TestMPIBuildCRDV1(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-v1-test",
		Image: "openmpi:4.1",
		Run:   "mpirun -np 4 ./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				SlotsPerWorker: 4,
			},
		},
		Worker: &task.Worker{
			Replicas: 4,
			Resources: task.Resources{
				"cpu":    "2",
				"memory": "8Gi",
			},
		},
	}

	provider := &MPIProvider{APIVersion: "v1"}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "MPIJob", crd.GetKind())
	assert.Equal(t, "mpi-v1-test", crd.GetName())
	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())

	spec := crd.Object["spec"].(map[string]interface{})

	// v1 shared fields
	assert.Equal(t, int64(4), spec["slotsPerWorker"])

	// v1 must NOT have v2beta1-only fields
	_, hasMPIImpl := spec["mpiImplementation"]
	assert.False(t, hasMPIImpl, "v1 CR should not have mpiImplementation")
	_, hasSSHAuth := spec["sshAuthMountPath"]
	assert.False(t, hasSSHAuth, "v1 CR should not have sshAuthMountPath")
	_, hasLauncherPolicy := spec["launcherCreationPolicy"]
	assert.False(t, hasLauncherPolicy, "v1 CR should not have launcherCreationPolicy")
	_, hasRunLauncher := spec["runLauncherAsWorker"]
	assert.False(t, hasRunLauncher, "v1 CR should not have runLauncherAsWorker")

	// Replica specs should exist
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	assert.Equal(t, int64(1), launcher["replicas"])
	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(4), worker["replicas"])
}

func TestMPIBuildCRDV1WithIntelMPI(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-v1-intel",
		Image: "intelmpi:2021",
		Run:   "mpirun -np 8 ./simulate",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				MPIImplementation: "Intel",
				SlotsPerWorker:    2,
			},
		},
		Worker: &task.Worker{Replicas: 8},
	}

	provider := &MPIProvider{APIVersion: "v1"}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())
	spec := crd.Object["spec"].(map[string]interface{})
	// v1 must NOT have mpiImplementation (v2beta1-only field)
	_, hasMPIImpl := spec["mpiImplementation"]
	assert.False(t, hasMPIImpl, "v1 CR should not have mpiImplementation even when user specifies it")
	assert.Equal(t, int64(2), spec["slotsPerWorker"])
}

func TestMPIBuildCRDV1GPUTopology(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-v1-topo",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				GPUTopology: true,
			},
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &MPIProvider{APIVersion: "v1"}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	meta := crd.Object["metadata"].(map[string]interface{})
	annotations := meta["annotations"].(map[string]interface{})
	assert.Equal(t, "true", annotations["mpi.kubeflow.org/gpu-topology"])
}

func TestMPIBuildCRDUnsupportedVersion(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-bad-version",
		Image: "openmpi:4.1",
		Run:   "./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{Replicas: 1},
	}

	provider := &MPIProvider{APIVersion: "v3"}
	_, err := provider.BuildCRD(tk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported MPI API version")
	assert.Contains(t, err.Error(), "v2beta1")
	assert.Contains(t, err.Error(), "v1")
}

func TestMPIProvider_BuildCRD_EmptyAPIVersion(t *testing.T) {
	p := &MPIProvider{APIVersion: ""}
	tk := &task.Task{
		Name:  "test",
		Image: "test:1",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{Replicas: 1},
	}
	_, err := p.BuildCRD(tk)
	if err == nil {
		t.Error("expected error for empty APIVersion")
	}
	if !strings.Contains(err.Error(), "apiversion is not set") {
		t.Errorf("error = %v, want contain 'apiversion is not set'", err)
	}
}

func TestMPIWorkerNoRunInjection(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-no-worker-run",
		Image: "openmpi:4.1",
		Run:   "mpirun -np 4 ./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	// Worker should NOT have command or args
	worker := replicaSpecs["Worker"].(map[string]interface{})
	workerTemplate := worker["template"].(map[string]interface{})
	workerSpec := workerTemplate["spec"].(map[string]interface{})
	containers := workerSpec["containers"].([]interface{})
	workerContainer := containers[0].(map[string]interface{})

	_, hasCommand := workerContainer["command"]
	assert.False(t, hasCommand, "MPI worker should not have command")
	_, hasArgs := workerContainer["args"]
	assert.False(t, hasArgs, "MPI worker should not have args")

	// Launcher should still have command and args
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	launcherTemplate := launcher["template"].(map[string]interface{})
	launcherSpec := launcherTemplate["spec"].(map[string]interface{})
	launcherContainers := launcherSpec["containers"].([]interface{})
	launcherContainer := launcherContainers[0].(map[string]interface{})

	assert.Equal(t, []interface{}{"/bin/sh", "-c"}, launcherContainer["command"])
	assert.Equal(t, []interface{}{"mpirun -np 4 ./train"}, launcherContainer["args"])
}

func TestMPILauncherRunOverride(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-launcher-override",
		Image: "openmpi:4.1",
		Run:   "mpirun -np 4 ./train",
		Framework: task.Framework{
			Name: "mpi",
		},
		Worker: &task.Worker{Replicas: 2},
		Launcher: &task.RoleConfig{
			Run: "mpirun -np 4 --bind-to none python train.py",
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})

	// Launcher should use its own run override
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	launcherTemplate := launcher["template"].(map[string]interface{})
	launcherSpec := launcherTemplate["spec"].(map[string]interface{})
	containers := launcherSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	assert.Equal(t, []interface{}{"mpirun -np 4 --bind-to none python train.py"}, container["args"])
}

func TestMPIBuildLauncherSpecInjectsDefaultSA(t *testing.T) {
	tk := &task.Task{
		Name:      "mpi-sa-test",
		Image:     "openmpi:4.1",
		Run:       "mpirun ./train",
		Namespace: "default",
		Framework: task.Framework{Name: "mpi"},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"nvidia.com/gpu": "1",
			},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	template := launcher["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})

	assert.Equal(t, "mpi-sa-test-launcher", podSpec["serviceAccountName"])
}

func TestMPIBuildLauncherSpecRespectsUserSA(t *testing.T) {
	tk := &task.Task{
		Name:           "mpi-user-sa",
		Image:          "openmpi:4.1",
		Run:            "mpirun ./train",
		Namespace:      "default",
		ServiceAccount: "my-custom-sa",
		Framework:      task.Framework{Name: "mpi"},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"nvidia.com/gpu": "1",
			},
		},
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV2beta1}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["mpiReplicaSpecs"].(map[string]interface{})
	launcher := replicaSpecs["Launcher"].(map[string]interface{})
	template := launcher["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})

	assert.Equal(t, "my-custom-sa", podSpec["serviceAccountName"])
}

func TestMPIBuildCRD_DoesNotMutateInput(t *testing.T) {
	tk := &task.Task{
		Name:  "mpi-gpu-topo",
		Image: "openmpi:4.1",
		Run:   "mpirun ./train",
		Framework: task.Framework{
			Name: "mpi",
			Options: task.FrameworkConfig{
				GPUTopology: true,
			},
		},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"nvidia.com/gpu": "1",
			},
		},
		Labels: map[string]string{
			"team":     "ml",
			"priority": "high",
		},
	}

	originalHostNetwork := tk.HostNetwork
	originalLabels := make(map[string]string, len(tk.Labels))
	for k, v := range tk.Labels {
		originalLabels[k] = v
	}

	provider := &MPIProvider{APIVersion: MPIAPIVersionV1}
	_, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, originalHostNetwork, tk.HostNetwork, "input Task.HostNetwork must not be mutated")
	assert.Equal(t, originalLabels, tk.Labels, "input Task.Labels must not be mutated")
}

package provider

import (
	"testing"

	"github.com/kubeflow/arena/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTensorFlowBuildCRD(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-test",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{
			Replicas: 3,
			Resources: task.Resources{
				"nvidia.com/gpu": "1",
			},
		},
		Chief: &task.RoleConfig{},
		PS:    &task.RoleConfig{Replicas: 2},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	assert.Equal(t, "TFJob", crd.GetKind())
	assert.Equal(t, "tf-test", crd.GetName())
	assert.Equal(t, "kubeflow.org/v1", crd.GetAPIVersion())

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	// Chief (1 replica)
	chief := replicaSpecs["Chief"].(map[string]interface{})
	assert.Equal(t, int64(1), chief["replicas"])

	// PS (2 replicas from options)
	ps := replicaSpecs["PS"].(map[string]interface{})
	assert.Equal(t, int64(2), ps["replicas"])

	// Worker (3 replicas)
	worker := replicaSpecs["Worker"].(map[string]interface{})
	assert.Equal(t, int64(3), worker["replicas"])

	// Verify resources on worker container (requests == limits for Guaranteed QoS)
	workerTemplate := worker["template"].(map[string]interface{})
	workerSpec := workerTemplate["spec"].(map[string]interface{})
	containers := workerSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})
	resources := container["resources"].(map[string]interface{})

	requests := resources["requests"].(map[string]interface{})
	assert.Equal(t, "1", requests["nvidia.com/gpu"])

	limits := resources["limits"].(map[string]interface{})
	assert.Equal(t, "1", limits["nvidia.com/gpu"])
}

func TestTensorFlowBuildCRDNoPS(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-no-ps",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"cpu":    "2",
				"memory": "8Gi",
			},
		},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	// Worker should exist
	_, ok := replicaSpecs["Worker"]
	assert.True(t, ok)

	// PS should NOT exist when ps_count is 0 or not set
	_, ok = replicaSpecs["PS"]
	assert.False(t, ok)
}

func TestTensorFlowBuildCRDWithRunAndShell(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-cmd",
		Image: "tensorflow:2.15",
		Run:   "python train.py --epochs 10",
		Shell: "/bin/bash",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})
	worker := replicaSpecs["Worker"].(map[string]interface{})
	template := worker["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})
	containers := podSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	assert.Equal(t, []interface{}{"/bin/bash", "-c"}, container["command"])
	assert.Equal(t, []interface{}{"python train.py --epochs 10"}, container["args"])
}

func TestTensorFlowBuildCRDWithEnv(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-env",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{Replicas: 1},
		Envs: map[string]task.EnvValue{
			"TF_DEBUG": {Value: "true"},
		},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})
	worker := replicaSpecs["Worker"].(map[string]interface{})
	template := worker["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})
	containers := podSpec["containers"].([]interface{})
	container := containers[0].(map[string]interface{})

	envVars := container["env"].([]interface{})
	require.Len(t, envVars, 1)
	envVar := envVars[0].(map[string]interface{})
	assert.Equal(t, "TF_DEBUG", envVar["name"])
	assert.Equal(t, "true", envVar["value"])
}

func TestTensorFlowBuildCRDWithoutChief(t *testing.T) {
	p := &TensorFlowProvider{}
	task := &task.Task{
		Name:      "tf-test",
		Image:     "tensorflow:2.15",
		Run:       "python train.py",
		Framework: task.Framework{Name: "tensorflow"},
		Worker:    &task.Worker{Replicas: 2, Resources: task.Resources{"nvidia.com/gpu": "1"}},
	}

	crd, err := p.BuildCRD(task)
	require.NoError(t, err)

	spec, ok := crd.Object["spec"].(map[string]interface{})
	require.True(t, ok)

	replicaSpecs, ok := spec["tfReplicaSpecs"].(map[string]interface{})
	require.True(t, ok)

	_, hasChief := replicaSpecs["Chief"]
	assert.False(t, hasChief, "Chief should not be present when t.Chief is nil")

	_, hasWorker := replicaSpecs["Worker"]
	assert.True(t, hasWorker, "Worker should be present")
}

func TestTensorFlowBuildCRDWithChief(t *testing.T) {
	p := &TensorFlowProvider{}
	task := &task.Task{
		Name:      "tf-test",
		Image:     "tensorflow:2.15",
		Run:       "python train.py",
		Framework: task.Framework{Name: "tensorflow"},
		Worker:    &task.Worker{Replicas: 2, Resources: task.Resources{"nvidia.com/gpu": "1"}},
		Chief:     &task.RoleConfig{Resources: task.Resources{"cpu": "4"}},
	}

	crd, err := p.BuildCRD(task)
	require.NoError(t, err)

	spec, ok := crd.Object["spec"].(map[string]interface{})
	require.True(t, ok)

	replicaSpecs, ok := spec["tfReplicaSpecs"].(map[string]interface{})
	require.True(t, ok)

	_, hasChief := replicaSpecs["Chief"]
	assert.True(t, hasChief, "Chief should be present when t.Chief is set")
}

func TestTensorFlowBuildCRDWithEvaluator(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-eval",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker:    &task.Worker{Replicas: 2},
		Evaluator: &task.RoleConfig{},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	evaluator := replicaSpecs["Evaluator"].(map[string]interface{})
	assert.Equal(t, int64(1), evaluator["replicas"])
}

func TestTensorFlowBuildCRDWithSuccessPolicy(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-success",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{Replicas: 2},
		Lifecycle: task.Lifecycle{
			SuccessPolicy: "AllWorkers",
		},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	assert.Equal(t, "AllWorkers", spec["successPolicy"])
}

func TestTensorFlowBuildCRDInvalidFramework(t *testing.T) {
	tk := &task.Task{
		Name:  "wrong-fw",
		Image: "pytorch:1.13",
		Run:   "echo hello",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &TensorFlowProvider{}
	_, err := provider.BuildCRD(tk)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tensorflow")
}

func TestTensorFlowGetJobType(t *testing.T) {
	provider := &TensorFlowProvider{}
	assert.Equal(t, "TFJob", provider.GetJobType())
}

func TestTensorFlowBuildCRDWithRoleOverrides(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-roles",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{
			Replicas: 4,
			Resources: task.Resources{
				"nvidia.com/gpu": "1",
			},
		},
		Envs: map[string]task.EnvValue{
			"TF_DEBUG": {Value: "true"},
		},
		Chief: &task.RoleConfig{
			Resources: task.Resources{
				"nvidia.com/gpu": "2",
			},
		},
		PS: &task.RoleConfig{
			Replicas: 2,
			Resources: task.Resources{
				"cpu":    "4",
				"memory": "16Gi",
			},
		},
		Evaluator: &task.RoleConfig{
			Resources: task.Resources{
				"cpu": "2",
			},
		},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	// Chief should have custom resources
	chief := replicaSpecs["Chief"].(map[string]interface{})
	assert.Equal(t, int64(1), chief["replicas"])
	chiefTemplate := chief["template"].(map[string]interface{})
	chiefPodSpec := chiefTemplate["spec"].(map[string]interface{})
	chiefContainers := chiefPodSpec["containers"].([]interface{})
	chiefContainer := chiefContainers[0].(map[string]interface{})
	chiefRes := chiefContainer["resources"].(map[string]interface{})
	chiefReqs := chiefRes["requests"].(map[string]interface{})
	assert.Equal(t, "2", chiefReqs["nvidia.com/gpu"])

	// PS should have replicas=2 and CPU resources
	ps := replicaSpecs["PS"].(map[string]interface{})
	assert.Equal(t, int64(2), ps["replicas"])
	psTemplate := ps["template"].(map[string]interface{})
	psPodSpec := psTemplate["spec"].(map[string]interface{})
	psContainers := psPodSpec["containers"].([]interface{})
	psContainer := psContainers[0].(map[string]interface{})
	psRes := psContainer["resources"].(map[string]interface{})
	psReqs := psRes["requests"].(map[string]interface{})
	assert.Equal(t, "4", psReqs["cpu"])
	assert.Equal(t, "16Gi", psReqs["memory"])

	// Evaluator should have CPU resources
	evaluator := replicaSpecs["Evaluator"].(map[string]interface{})
	assert.Equal(t, int64(1), evaluator["replicas"])
	evalTemplate := evaluator["template"].(map[string]interface{})
	evalPodSpec := evalTemplate["spec"].(map[string]interface{})
	evalContainers := evalPodSpec["containers"].([]interface{})
	evalContainer := evalContainers[0].(map[string]interface{})
	evalRes := evalContainer["resources"].(map[string]interface{})
	evalReqs := evalRes["requests"].(map[string]interface{})
	assert.Equal(t, "2", evalReqs["cpu"])

	// Worker should still have GPU resources from t.Worker.Resources
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

func TestTensorFlowBuildCRDWithRoleEnvOverrides(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-env-roles",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{
			Replicas: 2,
			Envs: map[string]task.EnvValue{
				"WORKER_VAR": {Value: "worker-val"},
				"SHARED_VAR": {Value: "worker-shared"},
			},
		},
		Envs: map[string]task.EnvValue{
			"GLOBAL_VAR": {Value: "global-val"},
		},
		Chief: &task.RoleConfig{
			Envs: map[string]task.EnvValue{
				"CHIEF_VAR":  {Value: "chief-val"},
				"SHARED_VAR": {Value: "chief-override"},
			},
		},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	chief := replicaSpecs["Chief"].(map[string]interface{})
	chiefTemplate := chief["template"].(map[string]interface{})
	chiefPodSpec := chiefTemplate["spec"].(map[string]interface{})
	chiefContainers := chiefPodSpec["containers"].([]interface{})
	chiefContainer := chiefContainers[0].(map[string]interface{})
	chiefEnvVars := chiefContainer["env"].([]interface{})

	chiefEnvMap := make(map[string]string)
	for _, e := range chiefEnvVars {
		envVar := e.(map[string]interface{})
		if val, ok := envVar["value"]; ok {
			chiefEnvMap[envVar["name"].(string)] = val.(string)
		}
	}
	assert.Equal(t, "global-val", chiefEnvMap["GLOBAL_VAR"])
	assert.Equal(t, "chief-val", chiefEnvMap["CHIEF_VAR"])
	assert.Equal(t, "chief-override", chiefEnvMap["SHARED_VAR"])
	// Chief should NOT have worker envs
	_, hasWorkerVar := chiefEnvMap["WORKER_VAR"]
	assert.False(t, hasWorkerVar, "chief should not inherit worker envs")
}

func TestTensorFlowProviderImplementsInterface(_ *testing.T) {
	var _ Provider = &TensorFlowProvider{}
}

func TestTensorFlowPerRoleRunOverride(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-override",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{
			Replicas: 2,
			Run:      "python worker_train.py",
		},
		Chief: &task.RoleConfig{
			Run: "python chief_train.py",
		},
		PS: &task.RoleConfig{
			Replicas: 1,
			Run:      "python serve_ps.py",
		},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	// Worker uses its own run override
	worker := replicaSpecs["Worker"].(map[string]interface{})
	workerContainer := extractContainer(t, worker)
	assert.Equal(t, []interface{}{"python worker_train.py"}, workerContainer["args"])

	// Chief uses its own run override
	chief := replicaSpecs["Chief"].(map[string]interface{})
	chiefContainer := extractContainer(t, chief)
	assert.Equal(t, []interface{}{"python chief_train.py"}, chiefContainer["args"])

	// PS uses its own run override
	ps := replicaSpecs["PS"].(map[string]interface{})
	psContainer := extractContainer(t, ps)
	assert.Equal(t, []interface{}{"python serve_ps.py"}, psContainer["args"])
}

func extractContainer(t *testing.T, replicaSpec map[string]interface{}) map[string]interface{} {
	t.Helper()
	template := replicaSpec["template"].(map[string]interface{})
	podSpec := template["spec"].(map[string]interface{})
	containers := podSpec["containers"].([]interface{})
	return containers[0].(map[string]interface{})
}

func TestTensorFlowBuildCRDNoEvaluatorWhenNil(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-no-eval",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{Replicas: 2},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	_, hasEval := replicaSpecs["Evaluator"]
	assert.False(t, hasEval, "Evaluator should not exist when not specified")
}

func TestTensorFlowBuildCRDChiefNoInheritWorker(t *testing.T) {
	tk := &task.Task{
		Name:  "tf-chief-no-inherit",
		Image: "tensorflow:2.15",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "tensorflow",
		},
		Worker: &task.Worker{
			Replicas: 2,
			Resources: task.Resources{
				"nvidia.com/gpu": "4",
			},
			Envs: map[string]task.EnvValue{
				"WORKER_VAR": {Value: "from-worker"},
			},
		},
		Chief: &task.RoleConfig{
			Resources: task.Resources{
				"cpu": "2",
			},
			Envs: map[string]task.EnvValue{
				"CHIEF_VAR": {Value: "from-chief"},
			},
		},
	}

	provider := &TensorFlowProvider{}
	crd, err := provider.BuildCRD(tk)
	require.NoError(t, err)

	spec := crd.Object["spec"].(map[string]interface{})
	replicaSpecs := spec["tfReplicaSpecs"].(map[string]interface{})

	chief := replicaSpecs["Chief"].(map[string]interface{})
	chiefTemplate := chief["template"].(map[string]interface{})
	chiefPodSpec := chiefTemplate["spec"].(map[string]interface{})
	chiefContainers := chiefPodSpec["containers"].([]interface{})
	chiefContainer := chiefContainers[0].(map[string]interface{})

	// Chief should have its own resources, not worker's GPU
	chiefRes := chiefContainer["resources"].(map[string]interface{})
	chiefReqs := chiefRes["requests"].(map[string]interface{})
	assert.Equal(t, "2", chiefReqs["cpu"])
	_, hasGPU := chiefReqs["nvidia.com/gpu"]
	assert.False(t, hasGPU, "chief should not inherit worker GPU resources")

	// Chief should have its own envs, not worker's
	chiefEnvVars := chiefContainer["env"].([]interface{})
	envMap := make(map[string]string)
	for _, e := range chiefEnvVars {
		env := e.(map[string]interface{})
		if val, ok := env["value"]; ok {
			envMap[env["name"].(string)] = val.(string)
		}
	}
	assert.Equal(t, "from-chief", envMap["CHIEF_VAR"])
	_, hasWorkerVar := envMap["WORKER_VAR"]
	assert.False(t, hasWorkerVar, "chief should not inherit worker envs")
}

func TestTensorFlowBuildRBAC(t *testing.T) {
	provider := &TensorFlowProvider{}
	resources, err := provider.BuildRBAC(&task.Task{}, metav1.OwnerReference{})
	require.NoError(t, err)
	assert.Nil(t, resources)
}

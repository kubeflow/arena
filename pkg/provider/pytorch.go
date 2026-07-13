package provider

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/task"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PyTorch replica types
const (
	PyTorchReplicaTypeMaster = "Master"
	PyTorchReplicaTypeWorker = "Worker"
)

// PyTorch spec field names
const (
	PyTorchFieldNprocPerNode = "nprocPerNode"
)

// PyTorchProvider generates PyTorchJob CRDs.
type PyTorchProvider struct{}

func (p *PyTorchProvider) GetJobType() string {
	return constants.KindPyTorchJob
}

func (p *PyTorchProvider) GetLogPodSelector(jobName string) string {
	return fmt.Sprintf("%s=%s,%s=%s",
		constants.LabelJobName, jobName,
		constants.LabelReplicaType, constants.ReplicaRoleMaster)
}

func (p *PyTorchProvider) GetJobPodSelector(jobName string) string {
	return fmt.Sprintf("%s=%s", constants.LabelJobName, jobName)
}

func (p *PyTorchProvider) BuildCRD(t *task.Task) (*unstructured.Unstructured, error) {
	if t.Framework.Name != constants.FrameworkPyTorch {
		return nil, fmt.Errorf("PyTorchProvider requires framework.name=%s, got %s", constants.FrameworkPyTorch, t.Framework.Name)
	}

	restartPolicy := constants.RestartPolicyOnFailure
	if t.Restart != "" {
		restartPolicy = t.Restart
	}

	replicaSpecs := map[string]interface{}{}

	if t.Worker == nil {
		// Master-only mode: single-node training
		masterSpec, err := buildRoleReplicaSpec(constants.FrameworkPyTorch, t, t.Master.Resources, t.Master.Envs, 1, restartPolicy, true)
		if err != nil {
			return nil, fmt.Errorf("failed to build master replica spec: %w", err)
		}
		replicaSpecs[PyTorchReplicaTypeMaster] = masterSpec
	} else {
		// Worker present: always generate Worker replicaSpec
		workerSpec, err := buildRoleReplicaSpec(constants.FrameworkPyTorch, t, t.Worker.Resources, t.Worker.Envs, int64(t.Worker.Replicas), restartPolicy, true)
		if err != nil {
			return nil, fmt.Errorf("failed to build worker replica spec: %w", err)
		}
		replicaSpecs[PyTorchReplicaTypeWorker] = workerSpec

		// Master: inherit from worker if not explicitly configured
		var masterResources task.Resources
		var masterEnvs map[string]task.EnvValue
		if t.Master != nil {
			masterResources = t.Master.Resources
			masterEnvs = t.Master.Envs
		} else {
			masterResources = t.Worker.Resources
			masterEnvs = t.Worker.Envs
		}

		masterSpec, err := buildRoleReplicaSpec(constants.FrameworkPyTorch, t, masterResources, masterEnvs, 1, restartPolicy, true)
		if err != nil {
			return nil, fmt.Errorf("failed to build master replica spec: %w", err)
		}
		replicaSpecs[PyTorchReplicaTypeMaster] = masterSpec
	}

	spec := map[string]interface{}{
		"pytorchReplicaSpecs": replicaSpecs,
	}

	if t.Framework.Options.NprocPerNode != "" {
		spec[PyTorchFieldNprocPerNode] = t.Framework.Options.NprocPerNode
	}

	runPolicy, err := buildRunPolicy(t)
	if err != nil {
		return nil, fmt.Errorf("failed to build run policy: %w", err)
	}
	if runPolicy != nil {
		spec["runPolicy"] = runPolicy
	}

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": constants.KubeflowGroup + "/" + constants.KubeflowVersion,
			"kind":       constants.KindPyTorchJob,
			"metadata":   buildMetadata(t),
			"spec":       spec,
		},
	}

	return crd, nil
}

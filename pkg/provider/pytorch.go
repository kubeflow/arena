package provider

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/task"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PyTorch replica types
const (
	pyTorchReplicaTypeMaster = "Master"
	pyTorchReplicaTypeWorker = "Worker"
)

// PyTorch spec field names
const (
	pyTorchFieldNprocPerNode = "nprocPerNode"
)

// PyTorchProvider generates PyTorchJob CRDs.
type PyTorchProvider struct{}

func (p *PyTorchProvider) GetJobType() string {
	return constants.KindPyTorchJob
}

// GetLogPodSelector returns the label selector for the master pod.
// Uses metav1.LabelSelector for consistent label value escaping.
func (p *PyTorchProvider) GetLogPodSelector(jobName string) string {
	return buildLabelSelector(map[string]string{
		constants.LabelJobName:     jobName,
		constants.LabelReplicaType: constants.ReplicaRoleMaster,
	})
}

func (p *PyTorchProvider) GetJobPodSelector(jobName string) string {
	return buildLabelSelector(map[string]string{
		constants.LabelJobName: jobName,
	})
}

func (p *PyTorchProvider) BuildCRD(t *task.Task) (*unstructured.Unstructured, error) {
	if t.Framework.Name != constants.FrameworkPyTorch {
		return nil, fmt.Errorf("pytorch provider requires framework.name=%s, got %q", constants.FrameworkPyTorch, t.Framework.Name)
	}

	restartPolicy := constants.RestartPolicyOnFailure
	if t.Restart != "" {
		restartPolicy = t.Restart
	}

	replicaSpecs := map[string]interface{}{}

	if t.Worker == nil {
		// Master-only mode: single-node training
		masterSpec, err := buildRoleReplicaSpec(replicaSpecOptions{
			ContainerName:  constants.FrameworkPyTorch,
			Task:           t,
			Resources:      t.Master.Resources,
			Envs:           t.Master.Envs,
			Replicas:       1,
			RestartPolicy:  restartPolicy,
			IncludeVolumes: true,
			Run:            effectiveRun(t, t.Master.Run),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to build master replica spec: %w", err)
		}
		replicaSpecs[pyTorchReplicaTypeMaster] = masterSpec
	} else {
		// Worker present: always generate Worker replicaSpec
		workerSpec, err := buildRoleReplicaSpec(replicaSpecOptions{
			ContainerName:  constants.FrameworkPyTorch,
			Task:           t,
			Resources:      t.Worker.Resources,
			Envs:           t.Worker.Envs,
			Replicas:       int64(t.Worker.Replicas),
			RestartPolicy:  restartPolicy,
			IncludeVolumes: true,
			Run:            effectiveRun(t, t.Worker.Run),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to build worker replica spec: %w", err)
		}
		replicaSpecs[pyTorchReplicaTypeWorker] = workerSpec

		// Master: inherit from worker if not explicitly configured
		var masterResources task.Resources
		var masterEnvs map[string]task.EnvValue
		var masterRun string
		if t.Master != nil {
			masterResources = t.Master.Resources
			masterEnvs = t.Master.Envs
			masterRun = t.Master.Run
		} else {
			masterResources = t.Worker.Resources
			masterEnvs = t.Worker.Envs
			masterRun = t.Worker.Run
		}

		masterSpec, err := buildRoleReplicaSpec(replicaSpecOptions{
			ContainerName:  constants.FrameworkPyTorch,
			Task:           t,
			Resources:      masterResources,
			Envs:           masterEnvs,
			Replicas:       1,
			RestartPolicy:  restartPolicy,
			IncludeVolumes: true,
			Run:            effectiveRun(t, masterRun),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to build master replica spec: %w", err)
		}
		replicaSpecs[pyTorchReplicaTypeMaster] = masterSpec
	}

	spec := map[string]interface{}{
		"pytorchReplicaSpecs": replicaSpecs,
	}

	if t.Framework.Options.NprocPerNode != "" {
		spec[pyTorchFieldNprocPerNode] = t.Framework.Options.NprocPerNode
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

// BuildRBAC returns nil — PyTorchJob does not require auxiliary RBAC resources.
func (p *PyTorchProvider) BuildRBAC(_ *task.Task, _ metav1.OwnerReference) ([]*unstructured.Unstructured, error) {
	return nil, nil
}

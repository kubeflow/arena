package provider

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/task"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TensorFlowProvider generates TFJob CRDs.
type TensorFlowProvider struct{}

func (p *TensorFlowProvider) GetJobType() string {
	return constants.KindTFJob
}

// GetLogPodSelector returns the label selector for the chief pod.
// Uses metav1.LabelSelector for consistent label value escaping.
func (p *TensorFlowProvider) GetLogPodSelector(jobName string) string {
	return buildLabelSelector(map[string]string{
		constants.LabelJobName:     jobName,
		constants.LabelReplicaType: constants.ReplicaRoleChief,
	})
}

func (p *TensorFlowProvider) GetJobPodSelector(jobName string) string {
	return buildLabelSelector(map[string]string{
		constants.LabelJobName: jobName,
	})
}

func (p *TensorFlowProvider) BuildCRD(t *task.Task) (*unstructured.Unstructured, error) {
	if t.Framework.Name != constants.FrameworkTensorFlow {
		return nil, fmt.Errorf("tensorflow provider requires framework.name=%s, got %q", constants.FrameworkTensorFlow, t.Framework.Name)
	}

	restartPolicy := constants.RestartPolicyOnFailure
	if t.Restart != "" {
		restartPolicy = t.Restart
	}

	replicaSpecs := map[string]interface{}{}

	// Worker: always present (validated)
	workerSpec, err := buildRoleReplicaSpec(replicaSpecOptions{
		ContainerName:  constants.FrameworkTensorFlow,
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
	replicaSpecs["Worker"] = workerSpec

	// Chief: only if section present
	if t.Chief != nil {
		chiefSpec, err := buildRoleReplicaSpec(replicaSpecOptions{
			ContainerName:  constants.FrameworkTensorFlow,
			Task:           t,
			Resources:      t.Chief.Resources,
			Envs:           t.Chief.Envs,
			Replicas:       1,
			RestartPolicy:  restartPolicy,
			IncludeVolumes: true,
			Run:            effectiveRun(t, t.Chief.Run),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to build chief replica spec: %w", err)
		}
		replicaSpecs["Chief"] = chiefSpec
	}

	// PS: only if section present, uses own config only
	// Validator guarantees PS.Replicas >= 1.
	if t.PS != nil {
		psSpec, err := buildRoleReplicaSpec(replicaSpecOptions{
			ContainerName:  constants.FrameworkTensorFlow,
			Task:           t,
			Resources:      t.PS.Resources,
			Envs:           t.PS.Envs,
			Replicas:       int64(t.PS.Replicas),
			RestartPolicy:  restartPolicy,
			IncludeVolumes: true,
			Run:            effectiveRun(t, t.PS.Run),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to build PS replica spec: %w", err)
		}
		replicaSpecs["PS"] = psSpec
	}

	// Evaluator: only if section present, uses own config only
	if t.Evaluator != nil {
		evalSpec, err := buildRoleReplicaSpec(replicaSpecOptions{
			ContainerName:  constants.FrameworkTensorFlow,
			Task:           t,
			Resources:      t.Evaluator.Resources,
			Envs:           t.Evaluator.Envs,
			Replicas:       1,
			RestartPolicy:  restartPolicy,
			IncludeVolumes: true,
			Run:            effectiveRun(t, t.Evaluator.Run),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to build evaluator replica spec: %w", err)
		}
		replicaSpecs["Evaluator"] = evalSpec
	}

	spec := map[string]interface{}{
		"tfReplicaSpecs": replicaSpecs,
	}

	sp := t.Lifecycle.SuccessPolicy
	if sp == constants.SuccessPolicyChiefWorkerAlias {
		sp = constants.SuccessPolicyDefault
	}
	if sp != "" {
		spec["successPolicy"] = sp
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
			"kind":       constants.KindTFJob,
			"metadata":   buildMetadata(t),
			"spec":       spec,
		},
	}

	return crd, nil
}

// BuildRBAC returns nil — TFJob does not require auxiliary RBAC resources.
func (p *TensorFlowProvider) BuildRBAC(_ *task.Task, _ metav1.OwnerReference) ([]*unstructured.Unstructured, error) {
	return nil, nil
}

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
// Note: jobName is used directly as a label value and must conform to
// Kubernetes label value constraints (max 63 chars, alphanumeric, -, _, .).
func (p *TensorFlowProvider) GetLogPodSelector(jobName string) string {
	return fmt.Sprintf("%s=%s,%s=%s",
		constants.LabelJobName, jobName,
		constants.LabelReplicaType, constants.ReplicaRoleChief)
}

func (p *TensorFlowProvider) GetJobPodSelector(jobName string) string {
	return fmt.Sprintf("%s=%s", constants.LabelJobName, jobName)
}

func (p *TensorFlowProvider) BuildCRD(t *task.Task) (*unstructured.Unstructured, error) {
	if t.Framework.Name != constants.FrameworkTensorFlow {
		return nil, fmt.Errorf("TensorFlowProvider requires framework.name=%s, got %s", constants.FrameworkTensorFlow, t.Framework.Name)
	}

	restartPolicy := constants.RestartPolicyOnFailure
	if t.Restart != "" {
		restartPolicy = t.Restart
	}

	replicaSpecs := map[string]interface{}{}

	// Worker: always present (validated)
	workerSpec, err := buildRoleReplicaSpec(constants.FrameworkTensorFlow, t, t.Worker.Resources, t.Worker.Envs, int64(t.Worker.Replicas), restartPolicy, true, effectiveRun(t, t.Worker.Run))
	if err != nil {
		return nil, fmt.Errorf("failed to build worker replica spec: %w", err)
	}
	replicaSpecs["Worker"] = workerSpec

	// Chief: only if section present
	if t.Chief != nil {
		chiefSpec, err := buildRoleReplicaSpec(constants.FrameworkTensorFlow, t, t.Chief.Resources, t.Chief.Envs, 1, restartPolicy, true, effectiveRun(t, t.Chief.Run))
		if err != nil {
			return nil, fmt.Errorf("failed to build chief replica spec: %w", err)
		}
		replicaSpecs["Chief"] = chiefSpec
	}

	// PS: only if section present, uses own config only
	// Validator guarantees PS.Replicas >= 1.
	if t.PS != nil {
		psSpec, err := buildRoleReplicaSpec(constants.FrameworkTensorFlow, t, t.PS.Resources, t.PS.Envs, int64(t.PS.Replicas), restartPolicy, true, effectiveRun(t, t.PS.Run))
		if err != nil {
			return nil, fmt.Errorf("failed to build PS replica spec: %w", err)
		}
		replicaSpecs["PS"] = psSpec
	}

	// Evaluator: only if section present, uses own config only
	if t.Evaluator != nil {
		evalSpec, err := buildRoleReplicaSpec(constants.FrameworkTensorFlow, t, t.Evaluator.Resources, t.Evaluator.Envs, 1, restartPolicy, true, effectiveRun(t, t.Evaluator.Run))
		if err != nil {
			return nil, fmt.Errorf("failed to build evaluator replica spec: %w", err)
		}
		replicaSpecs["Evaluator"] = evalSpec
	}

	spec := map[string]interface{}{
		"tfReplicaSpecs": replicaSpecs,
	}

	if t.Lifecycle.SuccessPolicy != "" {
		spec["successPolicy"] = t.Lifecycle.SuccessPolicy
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
func (p *TensorFlowProvider) BuildRBAC(t *task.Task, ownerRef metav1.OwnerReference) ([]*unstructured.Unstructured, error) {
	return nil, nil
}

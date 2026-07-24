package provider

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/task"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MPI API versions
const (
	MPIAPIVersionV2beta1 = constants.MPIVersionV2beta1
	MPIAPIVersionV1      = constants.KubeflowVersion
)

// MPI implementations
const (
	mpiImplementationOpenMPI = "OpenMPI"
	mpiImplementationIntel   = "Intel"
)

// Launcher creation policies
const (
	launcherCreationPolicyAtStartup           = "AtStartup"
	launcherCreationPolicyWaitForWorkersReady = "WaitForWorkersReady"
)

// MPI replica types
const (
	mpiReplicaTypeLauncher = "Launcher"
	mpiReplicaTypeWorker   = "Worker"
)

// MPI default values
const (
	mpiDefaultSlotsPerWorker   = int64(1)
	mpiDefaultSSHAuthMountPath = "/root/.ssh"
)

// MPIProvider generates MPIJob CRDs.
type MPIProvider struct {
	// APIVersion is the MPIJob CRD version to use (e.g. "v2beta1", "v1").
	// Must be set by the CLI layer after cluster detection before calling BuildCRD.
	APIVersion string
}

func (p *MPIProvider) GetJobType() string {
	return constants.KindMPIJob
}

// GetLogPodSelector returns the label selector for the launcher pod.
// Uses metav1.LabelSelector for consistent label value escaping.
func (p *MPIProvider) GetLogPodSelector(jobName string) string {
	return buildLabelSelector(map[string]string{
		constants.LabelJobName:     jobName,
		constants.LabelReplicaType: constants.ReplicaRoleLauncher,
	})
}

func (p *MPIProvider) GetJobPodSelector(jobName string) string {
	return buildLabelSelector(map[string]string{
		constants.LabelJobName: jobName,
	})
}

func (p *MPIProvider) BuildCRD(t *task.Task) (*unstructured.Unstructured, error) {
	switch t.Framework.Name {
	case constants.FrameworkMPI, constants.FrameworkHorovod, constants.FrameworkDeepSpeed:
		// OK
	default:
		return nil, fmt.Errorf("mpi provider requires framework.name=mpi, horovod, or deepspeed, got %q", t.Framework.Name)
	}

	version := p.APIVersion
	if version == "" {
		return nil, errors.New("mpi provider apiversion is not set")
	}

	if !client.IsMPIVersionSupported(version) {
		return nil, fmt.Errorf("unsupported MPI API version: %s (arena supports: %s)",
			version, strings.Join(client.MPISupportedVersions, ", "))
	}

	return p.buildMPICRD(t, version)
}

// BuildRBAC returns RBAC resources (ServiceAccount, Role, RoleBinding) for the
// MPIJob launcher pod. If the user specified a ServiceAccount via t.ServiceAccount,
// SA creation is skipped and only Role + RoleBinding are returned.
func (p *MPIProvider) BuildRBAC(t *task.Task, ownerRef metav1.OwnerReference) ([]*unstructured.Unstructured, error) {
	saName := t.ServiceAccount
	if saName == "" {
		saName = t.Name + "-launcher"
	}

	workerReplicas := 0
	if t.Worker != nil {
		workerReplicas = t.Worker.Replicas
	}

	var resources []*unstructured.Unstructured

	if t.ServiceAccount == "" {
		resources = append(resources, buildLauncherServiceAccount(saName, t.Name, t.Namespace, ownerRef))
	}

	resources = append(resources, buildLauncherRole(t.Name, t.Namespace, workerReplicas, ownerRef))
	resources = append(resources, buildLauncherRoleBinding(t.Name, t.Namespace, saName, ownerRef))

	return resources, nil
}

// buildMPICRD generates an MPIJob CRD for the given API version.
// v1 and v2beta1 share the core structure; v2beta1 has additional top-level fields.
func (p *MPIProvider) buildMPICRD(t *task.Task, apiVersion string) (*unstructured.Unstructured, error) {
	// Work on a copy to avoid mutating the caller's task (GPUTopology sets HostNetwork and Labels).
	tCopy := *t
	if t.Labels != nil {
		tCopy.Labels = make(map[string]string, len(t.Labels))
		for k, v := range t.Labels {
			tCopy.Labels[k] = v
		}
	}
	t = &tCopy

	// GPUTopology implies host networking and topology labels per cli-flag-mapping.md.
	if t.Framework.Options.GPUTopology {
		t.HostNetwork = true
		if t.Labels == nil {
			t.Labels = map[string]string{}
		}
		if _, exists := t.Labels["gpu-topology"]; !exists {
			t.Labels["gpu-topology"] = "true"
		}
		if _, exists := t.Labels["gpu-topology-replica"]; !exists {
			t.Labels["gpu-topology-replica"] = "true"
		}
	}

	restartPolicy := constants.RestartPolicyOnFailure
	if t.Restart != "" {
		restartPolicy = t.Restart
	}

	launcherSpec, err := p.buildLauncherSpec(t, restartPolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to build launcher spec: %w", err)
	}
	workerSpec, err := p.buildWorkerReplicaSpec(t, int64(t.Worker.Replicas), restartPolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to build worker replica spec: %w", err)
	}

	slotsPerWorker := mpiDefaultSlotsPerWorker
	if t.Framework.Options.SlotsPerWorker > 0 {
		slotsPerWorker = int64(t.Framework.Options.SlotsPerWorker)
	}

	mpiImplementation := mpiImplementationOpenMPI
	if t.Framework.Options.MPIImplementation != "" {
		mpiImplementation = t.Framework.Options.MPIImplementation
	}

	launcherCreationPolicy := launcherCreationPolicyAtStartup
	if t.Framework.Options.LauncherCreationPolicy != "" {
		launcherCreationPolicy = t.Framework.Options.LauncherCreationPolicy
	}

	sshAuthMountPath := mpiDefaultSSHAuthMountPath
	if t.Framework.Options.SSHAuthMountPath != "" {
		sshAuthMountPath = t.Framework.Options.SSHAuthMountPath
	}

	spec := map[string]interface{}{
		"slotsPerWorker": slotsPerWorker,
		"mpiReplicaSpecs": map[string]interface{}{
			mpiReplicaTypeLauncher: launcherSpec,
			mpiReplicaTypeWorker:   workerSpec,
		},
	}

	// v2beta1-only fields
	if apiVersion == MPIAPIVersionV2beta1 {
		spec["mpiImplementation"] = mpiImplementation
		spec["sshAuthMountPath"] = sshAuthMountPath
		spec["launcherCreationPolicy"] = launcherCreationPolicy
		spec["runLauncherAsWorker"] = t.Framework.Options.RunLauncherAsWorker
	}

	runPolicy, err := buildRunPolicy(t)
	if err != nil {
		return nil, fmt.Errorf("failed to build run policy: %w", err)
	}
	if runPolicy != nil {
		spec["runPolicy"] = runPolicy
	}

	metadata := buildMetadata(t)

	if t.Framework.Options.GPUTopology {
		annotations, ok := metadata["annotations"].(map[string]interface{})
		if !ok {
			annotations = map[string]interface{}{}
			metadata["annotations"] = annotations
		}
		annotations["mpi.kubeflow.org/gpu-topology"] = "true"
	}

	crd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": constants.KubeflowGroup + "/" + apiVersion,
			"kind":       constants.KindMPIJob,
			"metadata":   metadata,
			"spec":       spec,
		},
	}

	return crd, nil
}

// buildLauncherSpec creates the Launcher replica spec.
// Three-way conditional:
//   - Launcher explicitly configured: use its own resources and envs only (no worker merge)
//   - No launcher config + run_launcher_as_worker: inherit from worker
//   - No launcher config + no run_launcher_as_worker: CPU-only (nil resources/envs)
//
// Global task envs (t.Envs) are always merged by buildEnvVars inside buildRoleReplicaSpec.
func (p *MPIProvider) buildLauncherSpec(t *task.Task, restartPolicy string) (map[string]interface{}, error) {
	var resources task.Resources
	var envs map[string]task.EnvValue
	var launcherRun string

	switch {
	case t.Launcher != nil:
		// Launcher explicitly configured: use its own config only
		resources = t.Launcher.Resources
		envs = t.Launcher.Envs
		launcherRun = t.Launcher.Run
	case t.Framework.Options.RunLauncherAsWorker:
		// No launcher config + run_launcher_as_worker: inherit from worker
		resources = t.Worker.Resources
		envs = t.Worker.Envs
	default:
		// No launcher config + no run_launcher_as_worker: CPU-only
		resources = nil
		envs = nil
	}

	includeVolumes := t.Framework.Options.MountsOnLauncher
	spec, err := buildRoleReplicaSpec(replicaSpecOptions{
		ContainerName:  constants.FrameworkMPI,
		Task:           t,
		Resources:      resources,
		Envs:           envs,
		Replicas:       1,
		RestartPolicy:  restartPolicy,
		IncludeVolumes: includeVolumes,
		Run:            effectiveRun(t, launcherRun),
	})
	if err != nil {
		return nil, err
	}

	// Inject launcher SA if user hasn't specified one.
	// Use unstructured helpers for safe nested access instead of raw type assertions.
	if t.ServiceAccount == "" {
		if _, found, err := unstructured.NestedMap(spec, "template", "spec"); err != nil || !found {
			return nil, errors.New("launcher spec missing template.spec field")
		}
		if err := unstructured.SetNestedField(spec, t.Name+"-launcher", "template", "spec", "serviceAccountName"); err != nil {
			return nil, fmt.Errorf("failed to set serviceAccountName on launcher spec: %w", err)
		}
	}

	return spec, nil
}

// buildWorkerReplicaSpec creates the Worker replica spec with full resources.
// MPI workers run as SSH daemons for the launcher to dispatch commands to.
// They must not receive the user's run command — only the launcher executes it.
func (p *MPIProvider) buildWorkerReplicaSpec(t *task.Task, replicas int64, restartPolicy string) (map[string]interface{}, error) {
	container := buildContainer(containerOptions{
		Name:      constants.FrameworkMPI,
		Image:     t.Image,
		Task:      t,
		Resources: t.Worker.Resources,
		RoleEnvs:  t.Worker.Envs,
		Run:       "",
		Mounts:    nil,
	})
	podSpec, err := buildPodSpec(t, container, true)
	if err != nil {
		return nil, err
	}

	template := map[string]interface{}{
		"spec": podSpec,
	}

	return map[string]interface{}{
		"replicas":      replicas,
		"restartPolicy": restartPolicy,
		"template":      template,
	}, nil
}

// MPISupportedVersions returns the list of supported MPIJob CRD versions.
// Exported for use by the CLI layer. Delegates to the client package to
// maintain a single source of truth for supported versions.
func MPISupportedVersions() []string {
	return client.MPISupportedVersions
}

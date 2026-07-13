package provider

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/task"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MPI API versions
const (
	MPIAPIVersionV2beta1 = constants.MPIVersionV2beta1
	MPIAPIVersionV1      = constants.KubeflowVersion
)

// mpiSupportedVersions lists all MPIJob CRD versions arena can construct.
var mpiSupportedVersions = []string{MPIAPIVersionV2beta1, MPIAPIVersionV1}

// MPI implementations
const (
	MPIImplementationOpenMPI = "OpenMPI"
	MPIImplementationIntel   = "Intel"
	MPIImplementationMPICH   = "MPICH"
)

// Launcher creation policies
const (
	LauncherCreationPolicyAtStartup           = "AtStartup"
	LauncherCreationPolicyWaitForWorkersReady = "WaitForWorkersReady"
)

// MPI replica types
const (
	MPIReplicaTypeLauncher = "Launcher"
	MPIReplicaTypeWorker   = "Worker"
)

// MPI default values
const (
	MPIDefaultSlotsPerWorker   = int64(1)
	MPIDefaultSSHAuthMountPath = "/root/.ssh"
)

// MPIProvider generates MPIJob CRDs.
type MPIProvider struct {
	// APIVersion is the MPIJob CRD version to use (e.g. "v2beta1", "v1").
	// Must be set by the CLI layer after cluster detection before calling BuildCRD.
	APIVersion string
}

// isMPIVersionSupported checks if a version is in the supported set.
func isMPIVersionSupported(version string) bool {
	for _, v := range mpiSupportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

// MPISupportedVersions returns the list of supported MPIJob CRD versions.
// Exported for use by the CLI layer.
func MPISupportedVersions() []string {
	return mpiSupportedVersions
}

func (p *MPIProvider) GetJobType() string {
	return constants.KindMPIJob
}

func (p *MPIProvider) GetLogPodSelector(jobName string) string {
	return fmt.Sprintf("%s=%s,%s=%s",
		constants.LabelJobName, jobName,
		constants.LabelReplicaType, constants.ReplicaRoleLauncher)
}

func (p *MPIProvider) GetJobPodSelector(jobName string) string {
	return fmt.Sprintf("%s=%s", constants.LabelJobName, jobName)
}

func (p *MPIProvider) BuildCRD(t *task.Task) (*unstructured.Unstructured, error) {
	if t.Framework.Name != constants.FrameworkMPI && t.Framework.Name != constants.FrameworkHorovod && t.Framework.Name != constants.FrameworkDeepSpeed {
		return nil, fmt.Errorf("MPIProvider requires framework.name=mpi, horovod, or deepspeed, got %s", t.Framework.Name)
	}

	version := p.APIVersion
	if version == "" {
		return nil, fmt.Errorf("MPIProvider.APIVersion must be set before calling BuildCRD")
	}

	if !isMPIVersionSupported(version) {
		return nil, fmt.Errorf("unsupported MPI API version: %s (arena supports: %s)",
			version, strings.Join(mpiSupportedVersions, ", "))
	}

	return p.buildMPICRD(t, version)
}

// buildMPICRD generates an MPIJob CRD for the given API version.
// v1 and v2beta1 share the core structure; v2beta1 has additional top-level fields.
func (p *MPIProvider) buildMPICRD(t *task.Task, apiVersion string) (*unstructured.Unstructured, error) {
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

	slotsPerWorker := MPIDefaultSlotsPerWorker
	if t.Framework.Options.SlotsPerWorker > 0 {
		slotsPerWorker = int64(t.Framework.Options.SlotsPerWorker)
	}

	mpiImplementation := MPIImplementationOpenMPI
	if t.Framework.Options.MPIImplementation != "" {
		mpiImplementation = t.Framework.Options.MPIImplementation
	}

	launcherCreationPolicy := LauncherCreationPolicyAtStartup
	if t.Framework.Options.LauncherCreationPolicy != "" {
		launcherCreationPolicy = t.Framework.Options.LauncherCreationPolicy
	}

	sshAuthMountPath := MPIDefaultSSHAuthMountPath
	if t.Framework.Options.SSHAuthMountPath != "" {
		sshAuthMountPath = t.Framework.Options.SSHAuthMountPath
	}

	spec := map[string]interface{}{
		"slotsPerWorker": slotsPerWorker,
		"mpiReplicaSpecs": map[string]interface{}{
			MPIReplicaTypeLauncher: launcherSpec,
			MPIReplicaTypeWorker:   workerSpec,
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

	if t.Launcher != nil {
		// Launcher explicitly configured: use its own config only
		resources = t.Launcher.Resources
		envs = t.Launcher.Envs
	} else if t.Framework.Options.RunLauncherAsWorker {
		// No launcher config + run_launcher_as_worker: inherit from worker
		resources = t.Worker.Resources
		envs = t.Worker.Envs
	} else {
		// No launcher config + no run_launcher_as_worker: CPU-only
		resources = nil
		envs = nil
	}

	includeVolumes := t.Framework.Options.MountsOnLauncher
	return buildRoleReplicaSpec(constants.FrameworkMPI, t, resources, envs, 1, restartPolicy, includeVolumes)
}

// buildWorkerReplicaSpec creates the Worker replica spec with full resources.
func (p *MPIProvider) buildWorkerReplicaSpec(t *task.Task, replicas int64, restartPolicy string) (map[string]interface{}, error) {
	container := buildContainer(constants.FrameworkMPI, t.Image, t, t.Worker.Envs)
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

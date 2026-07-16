package constants

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameworkNameConstants(t *testing.T) {
	assert.Equal(t, "pytorch", FrameworkPyTorch)
	assert.Equal(t, "tensorflow", FrameworkTensorFlow)
	assert.Equal(t, "mpi", FrameworkMPI)
	assert.Equal(t, "horovod", FrameworkHorovod)
	assert.Equal(t, "deepspeed", FrameworkDeepSpeed)
	assert.Equal(t, "ray", FrameworkRay)
}

func TestCRDKindConstants(t *testing.T) {
	assert.Equal(t, "PyTorchJob", KindPyTorchJob)
	assert.Equal(t, "TFJob", KindTFJob)
	assert.Equal(t, "MPIJob", KindMPIJob)
}

func TestCRDGroupAndVersionConstants(t *testing.T) {
	assert.Equal(t, "kubeflow.org", KubeflowGroup)
	assert.Equal(t, "v1", KubeflowVersion)
	assert.Equal(t, "v2beta1", MPIVersionV2beta1)
}

func TestRestartPolicyConstants(t *testing.T) {
	assert.Equal(t, "OnFailure", RestartPolicyOnFailure)
	assert.Equal(t, "Always", RestartPolicyAlways)
	assert.Equal(t, "Never", RestartPolicyNever)
}

func TestCleanPodPolicyConstants(t *testing.T) {
	assert.Equal(t, "None", CleanPodPolicyNone)
	assert.Equal(t, "Running", CleanPodPolicyRunning)
	assert.Equal(t, "All", CleanPodPolicyAll)
}

func TestJobStatusConstants(t *testing.T) {
	assert.Equal(t, "Pending", JobStatusPending)
	assert.Equal(t, "Running", JobStatusRunning)
	assert.Equal(t, "Suspended", JobStatusSuspended)
	assert.Equal(t, "Unknown", JobStatusUnknown)
}

func TestK8sResourceFieldConstants(t *testing.T) {
	assert.Equal(t, "Memory", EmptyDirMediumMemory)
	assert.Equal(t, "In", AffinityOperatorIn)
}

func TestClientConfigDefaults(t *testing.T) {
	assert.Equal(t, float32(10.0), DefaultQPS)
	assert.Equal(t, 20, DefaultBurst)
}

func TestDefaultRsyncImageNotLatest(t *testing.T) {
	assert.NotEqual(t, "latest", versionTag(DefaultRsyncImage),
		"DefaultRsyncImage must not use :latest tag for reproducibility")
}

// versionTag extracts the tag after the last colon in an image reference.
func versionTag(image string) string {
	for i := len(image) - 1; i >= 0; i-- {
		if image[i] == ':' {
			return image[i+1:]
		}
		if image[i] == '/' {
			break
		}
	}
	return ""
}

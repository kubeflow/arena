// Package constants defines shared default values used across Arena v2 packages.
// Centralising magic strings here keeps them discoverable and easy to change.
package constants

// Default images for sync init containers.
const (
	// DefaultGitSyncImage is the git-sync sidecar used to clone a Git repository
	// into the training pod before the main container starts.
	DefaultGitSyncImage = "registry.k8s.io/git-sync/git-sync:v3.3.5"

	// DefaultRsyncImage is the rsync image used to copy a remote directory
	// into the training pod via rsync.
	DefaultRsyncImage = "docker.io/instrumentisto/rsync-ssh:3.18"

	// DefaultHDFSImage is the Hadoop image used to copy files from HDFS
	// into the training pod.
	DefaultHDFSImage = "docker.io/apache/hadoop:3.5.0"
)

// Default images for auxiliary services.
const (
	// DefaultTensorBoardImage is the image used for the auto-created
	// TensorBoard Deployment when the user does not supply one.
	DefaultTensorBoardImage = "docker.io/tensorflow/tensorflow:2.21.0"
)

// Default shell interpreter.
const (
	// DefaultShell is the shell used in init containers and as the fallback
	// for the EffectiveShell helper when no valid shell is configured.
	DefaultShell = "/bin/sh"
)

// Default mount paths.
const (
	// DefaultSHMMountPath is the standard Linux shared-memory mount point.
	// Used when constructing a tmpfs/emptyDir storage entry for --shm.
	DefaultSHMMountPath = "/dev/shm"
)

// Default ports.
const (
	// DefaultTensorBoardPort is the port exposed by the TensorBoard container
	// and the port the auto-created Service listens on.
	DefaultTensorBoardPort = 6006
)

// Framework names used in task.Framework.Name and arena.io/framework labels.
const (
	FrameworkPyTorch    = "pytorch"
	FrameworkTensorFlow = "tensorflow"
	FrameworkMPI        = "mpi"
	FrameworkHorovod    = "horovod"
	FrameworkDeepSpeed  = "deepspeed"
	FrameworkRay        = "ray"
)

// CRD kinds for training job operators.
const (
	KindPyTorchJob = "PyTorchJob"
	KindTFJob      = "TFJob"
	KindMPIJob     = "MPIJob"
)

// CRD group and versions.
const (
	KubeflowGroup     = "kubeflow.org"
	KubeflowVersion   = "v1"
	MPIVersionV2beta1 = "v2beta1"
)

// Restart policies for training jobs.
const (
	RestartPolicyOnFailure = "OnFailure"
	RestartPolicyAlways    = "Always"
	RestartPolicyNever     = "Never"
)

// Clean pod policies.
const (
	CleanPodPolicyNone    = "None"
	CleanPodPolicyRunning = "Running"
	CleanPodPolicyAll     = "All"
)

// Job status values.
const (
	JobStatusPending   = "Pending"
	JobStatusRunning   = "Running"
	JobStatusSuspended = "Suspended"
	JobStatusUnknown   = "Unknown"
)

// K8s resource field values.
const (
	EmptyDirMediumMemory = "Memory"
	AffinityOperatorIn   = "In"
)

// Client configuration defaults.
const (
	DefaultQPS   float32 = 10.0
	DefaultBurst int     = 20
)

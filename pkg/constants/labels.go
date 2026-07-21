package constants

// Trainer label keys — matching Kubeflow Trainer's common label convention
// (training.kubeflow.org/*). All frameworks share these keys; the framework
// is distinguished by the operator-name value. Arena reads these; the
// training operator sets them on pods.
const (
	LabelJobName      = "training.kubeflow.org/job-name"
	LabelOperatorName = "training.kubeflow.org/operator-name"
	LabelReplicaType  = "training.kubeflow.org/replica-type"
	LabelReplicaIndex = "training.kubeflow.org/replica-index"
	LabelJobRole      = "training.kubeflow.org/job-role"
)

// Replica role values used with LabelReplicaType.
const (
	ReplicaRoleMaster    = "master"
	ReplicaRoleChief     = "chief"
	ReplicaRoleLauncher  = "launcher"
	ReplicaRoleWorker    = "worker"
	ReplicaRolePS        = "ps"
	ReplicaRoleEvaluator = "evaluator"
)

// Arena-owned labels — set by arena on resources it creates (e.g. TensorBoard).
const (
	LabelComponent    = "arena.io/component"
	LabelArenaJobName = "arena.io/job-name"

	ComponentTensorBoard = "tensorboard"
)

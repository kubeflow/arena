package k8saccesser

const (
	TensorflowCRDName             = "tfjobs.kubeflow.org"
	TensorflowCRDNameInDaemonMode = "TFJob.kubeflow.org"

	MPICRDName             = "mpijobs.kubeflow.org"
	MPICRDNameInDaemonMode = "MPIJob.kubeflow.org"

	PytorchCRDName             = "pytorchjobs.kubeflow.org"
	PytorchCRDNameInDaemonMode = "PyTorchJob.kubeflow.org"

	ETCRDName             = "trainingjobs.kai.alibabacloud.com"
	ETCRDNameInDaemonMode = "TrainingJob.kai.alibabacloud.com"

	VolcanoCRDName             = "jobs.batch.volcano.sh"
	VolcanoCRDNameInDaemonMode = "Job.batch.volcano.sh"

	SparkCRDNameInDaemonMode = "Sparkapplication.sparkoperator.k8s.io"
	SparkCRDName             = "sparkapplications.sparkoperator.k8s.io"
)

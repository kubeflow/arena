package kubectl

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/training"
	"github.com/kubeflow/arena/pkg/apis/types"
)

func AddTrainingJobLabel(job *training.Job, key string, value string) error {
	switch job.Type() {
	case types.TFTrainingJob:
		args := job.Args().(*types.SubmitTFJobArgs)
		_, err := kubectl([]string{
			"label",
			"-n",
			args.Namespace,
			"tfjobs.kubeflow.org",
			args.Name,
			fmt.Sprintf("%s=%s", key, value),
		})
		if err != nil {
			return err
		}
	case types.PytorchTrainingJob:
		args := job.Args().(*types.SubmitPyTorchJobArgs)
		_, err := kubectl([]string{
			"label",
			"-n",
			args.Namespace,
			"pytorchjobs.kubeflow.org",
			args.Name,
			fmt.Sprintf("%s=%s", key, value),
		})
		if err != nil {
			return err
		}
	case types.MPITrainingJob:
		args := job.Args().(*types.SubmitMPIJobArgs)
		_, err := kubectl([]string{
			"label",
			"-n",
			args.Namespace,
			"mpijobs.kubeflow.org",
			args.Name,
			fmt.Sprintf("%s=%s", key, value),
		})
		if err != nil {
			return err
		}
	case types.HorovodTrainingJob:
		args := job.Args().(*types.SubmitHorovodJobArgs)
		_, err := kubectl([]string{
			"label",
			"-n",
			args.Namespace,
			"mpijobs.kubeflow.org",
			args.Name,
			fmt.Sprintf("%s=%s", key, value),
		})
		if err != nil {
			return err
		}
	case types.VolcanoTrainingJob:
		args := job.Args().(*types.SubmitVolcanoJobArgs)
		_, err := kubectl([]string{
			"label",
			"-n",
			args.Namespace,
			"job.batch.volcano.sh",
			args.Name,
			fmt.Sprintf("%s=%s", key, value),
		})
		if err != nil {
			return err
		}
	case types.ETTrainingJob:
		args := job.Args().(*types.SubmitETJobArgs)
		_, err := kubectl([]string{
			"label",
			"-n",
			args.Namespace,
			"trainingjobs.kai.alibabacloud.com",
			args.Name,
			fmt.Sprintf("%s=%s", key, value),
		})
		if err != nil {
			return err
		}
	case types.SparkTrainingJob:
		args := job.Args().(*types.SubmitSparkJobArgs)
		_, err := kubectl([]string{
			"label",
			"-n",
			args.Namespace,
			"sparkapplications.sparkoperator.k8s.io",
			args.Name,
			fmt.Sprintf("%s=%s", key, value),
		})
		if err != nil {
			return err
		}
	case types.DeepSpeedTrainingJob:
		args := job.Args().(*types.SubmitDeepSpeedJobArgs)
		_, err := kubectl([]string{
			"label",
			"-n",
			args.Namespace,
			"trainingjobs.kai.alibabacloud.com",
			args.Name,
			fmt.Sprintf("%s=%s", key, value),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

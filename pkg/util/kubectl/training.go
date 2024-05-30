// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

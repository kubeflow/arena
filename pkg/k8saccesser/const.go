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

	RayJobCRDName             = "rayjobs.ray.io"
	RayJobCRDNameInDaemonMode = "RayJob.ray.io"

	LWSCRDName             = "leaderworkersets.leaderworkerset.x-k8s.io"
	LWSCRDNameInDaemonMode = "Leaderworkerset.leaderworkerset.x-k8s.io"
)

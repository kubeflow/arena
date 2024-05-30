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

package argsbuilder

const (
	// ShareDataPrefix is used to defines sharing data from parent builder to children builder
	ShareDataPrefix = "share-"

	gangSchdName = "kube-batch"

	aliyunENIAnnotation = "k8s.aliyun.com/eni"

	jobSuspend = "scheduling.x-k8s.io/suspend"

	spotInstanceAnnotation = "job-supervisor.kube-ai.io/spot-instance"

	maxWaitTimeAnnotation = "job-supervisor.kube-ai.io/max-wait-time"

	NCCLAsyncErrorHanding = "NCCL_ASYNC_ERROR_HANDLING"
)

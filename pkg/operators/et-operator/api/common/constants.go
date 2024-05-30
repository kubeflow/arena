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

package common

const (
	// ReplicaIndexLabel represents the label key for the replica-index, e.g. the value is 0, 1, 2.. etc
	ReplicaIndexLabel = "replica-index"

	// ReplicaTypeLabel represents the label key for the replica-type, e.g. the value is ps , worker etc.
	ReplicaTypeLabel = "replica-type"

	// GroupNameLabel represents the label key for group name, e.g. the value is kubeflow.org
	GroupNameLabel = "group-name"

	// JobNameLabel represents the label key for the job name, the value is job name
	JobNameLabel = "job-name"

	// JobRoleLabel represents the label key for the job role, e.g. the value is master
	JobRoleLabel = "job-role"
)

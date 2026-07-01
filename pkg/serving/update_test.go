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

package serving

import (
	"strings"
	"testing"

	"github.com/kubeflow/arena/pkg/apis/types"
)

func TestUpdateDistributedServingReturnsFindAndBuildLWSJobErrors(t *testing.T) {
	args := &types.UpdateDistributedServingArgs{
		CommonUpdateServingArgs: types.CommonUpdateServingArgs{
			Name: "demo",
			Type: types.UnknownServingJob,
		},
	}

	err := UpdateDistributedServing(args)
	if err == nil {
		t.Fatal("expected UpdateDistributedServing to return an error")
	}
	if !strings.Contains(err.Error(), "unknown serving job type") {
		t.Fatalf("expected unknown serving job type error, got %q", err)
	}
}

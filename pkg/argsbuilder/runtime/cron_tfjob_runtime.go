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

package runtime

import (
	"github.com/kubeflow/arena/pkg/apis/types"
)

type defaultCronTFRuntime struct {
	name string
}

func NewDefaultCronTFJobRuntime() types.TFRuntime {
	return &defaultCronTFRuntime{
		name: "cron",
	}
}

func (d *defaultCronTFRuntime) Check(tf *types.SubmitTFJobArgs) error {
	return nil
}

func (d *defaultCronTFRuntime) Transform(tf *types.SubmitTFJobArgs) error {
	return nil
}

func (d *defaultCronTFRuntime) GetChartName() string {
	return d.name
}

func (d *defaultCronTFRuntime) IsDefault() bool {
	return true
}

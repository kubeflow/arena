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

package analyze

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
)

func SubmitModelProfileJob(namespace string, args *types.ModelProfileArgs) error {
	args.Namespace = namespace

	if args.Command == "" {
		args.Command = fmt.Sprintf("python easy_inference/main.py profile --model-config-file=%s --report-path=%s",
			args.ModelConfigFile, args.ReportPath)
	}

	modelJobChart := util.GetChartsFolder() + "/modeljob"
	err := workflow.SubmitJob(args.Name, string(types.ModelProfileJob), namespace, args, modelJobChart, args.HelmOptions...)
	if err != nil {
		return err
	}
	log.Infof("The model profile job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena model analyze get %s` to check the job status", args.Name)
	return nil
}

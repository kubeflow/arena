// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License
package argsbuilder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kubeflow/arena/pkg/apis/types"
)

type SubmitPytorchJobArgsBuilder struct {
	args        *types.SubmitPyTorchJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitPytorchJobArgsBuilder(args *types.SubmitPyTorchJobArgs) ArgsBuilder {
	args.TrainingType = types.PytorchTrainingJob
	s := &SubmitPytorchJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewSubmitArgsBuilder(&s.args.CommonSubmitArgs),
		NewSubmitSyncCodeArgsBuilder(&s.args.SubmitSyncCodeArgs),
		NewSubmitTensorboardArgsBuilder(&s.args.SubmitTensorboardArgs),
	)
	return s
}

func (s *SubmitPytorchJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitPytorchJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitPytorchJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitPytorchJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}

	var (
		runningTimeout   time.Duration
		ttlAfterFinished time.Duration
	)

	command.Flags().StringVar(&s.args.CleanPodPolicy, "clean-task-policy", "Running", "How to clean tasks after Training is done, support None, Running, All.")
	command.Flags().StringVar(&s.args.Cpu, "cpu", "", "the cpu resource to use for the training, like 1 for 1 core.")
	command.Flags().StringVar(&s.args.Memory, "memory", "", "the memory resource to use for the training, like 1Gi.")
	command.Flags().DurationVar(&runningTimeout, "running-timeout", runningTimeout, "Specifies the duration since startTime during which the job can remain active before it is terminated(e.g. '5s', '1m', '2h22m').")
	command.Flags().DurationVar(&ttlAfterFinished, "ttl-after-finished", ttlAfterFinished, "Defines the TTL for cleaning up finished PytorchJobs(e.g. '5s', '1m', '2h22m'). Defaults to infinite.")
	command.Flags().StringVar(&s.args.ShareMemory, "share-memory", "2Gi", "the shared memory of each replica to run the job, default 2Gi.")
	command.Flags().StringVar(&s.args.NprocPerNode, "nproc-per-node", "", "The number of processes per node, available values are \"auto\", \"cpu\", \"gpu\" and a number (e.g. 4).")

	s.AddArgValue("running-timeout", &runningTimeout).
		AddArgValue("ttl-after-finished", &ttlAfterFinished)
}

func (s *SubmitPytorchJobArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	s.AddArgValue(ShareDataPrefix+"dataset", s.args.DataSet)
	return nil
}

func (s *SubmitPytorchJobArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.setRunPolicy(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	if err := s.addEnv(); err != nil {
		return err
	}
	return nil
}

func (s *SubmitPytorchJobArgsBuilder) setRunPolicy() error {
	// Get active deadline
	if rt, ok := s.argValues["running-timeout"]; ok {
		runningTimeout := rt.(*time.Duration)
		s.args.ActiveDeadlineSeconds = int64(runningTimeout.Seconds())
	}

	// Get ttlSecondsAfterFinished
	if ft, ok := s.argValues["ttl-after-finished"]; ok {
		ttlAfterFinished := ft.(*time.Duration)
		s.args.TTLSecondsAfterFinished = int32(ttlAfterFinished.Seconds())
	}
	return nil
}

func (s *SubmitPytorchJobArgsBuilder) check() error {
	if s.args.Image == "" {
		return fmt.Errorf("--image must be set ")
	}
	// check clean-task-policy
	switch s.args.CleanPodPolicy {
	case "None", "Running", "All":
		log.Debugf("Supported cleanTaskPolicy: %s", s.args.CleanPodPolicy)
	default:
		return fmt.Errorf("Unsupported cleanTaskPolicy %s", s.args.CleanPodPolicy)
	}

	if s.args.GPUCount < 0 {
		return fmt.Errorf("--gpus is invalid")
	}
	if s.args.Cpu != "" {
		_, err := resource.ParseQuantity(s.args.Cpu)
		if err != nil {
			return fmt.Errorf("--cpu is invalid")
		}
	}
	if s.args.Memory != "" {
		_, err := resource.ParseQuantity(s.args.Memory)
		if err != nil {
			return fmt.Errorf("--memory is invalid")
		}
	}
	if s.args.ActiveDeadlineSeconds < 0 {
		return fmt.Errorf("--running-timeout is invalid")
	}
	if s.args.TTLSecondsAfterFinished < 0 {
		return fmt.Errorf("--ttl-after-finished is invalid")
	}
	if s.args.ShareMemory != "" {
		_, err := resource.ParseQuantity(s.args.ShareMemory)
		if err != nil {
			return fmt.Errorf("--share-memory is invalid")
		}
	}

	// Check whether nprocPerNode is valid
	switch s.args.NprocPerNode {
	case "auto", "cpu", "gpu":
		log.Debugf("Supported nprocPerNode: %s", s.args.NprocPerNode)
	case "":
		log.Debugf("--nproc-per-node is not set")
	default:
		nprocPerNode, err := strconv.Atoi(s.args.NprocPerNode)
		if err != nil {
			return fmt.Errorf("--nproc-per-node is invalid")
		}
		log.Debugf("Supported nprocPerNode: %d", nprocPerNode)
	}

	return nil
}

func (s *SubmitPytorchJobArgsBuilder) addEnv() error {
	if s.args.Envs == nil {
		s.args.Envs = map[string]string{}
	}

	if s.args.EnableRDMA {
		s.args.Envs["MASTER_ADDR"] = fmt.Sprintf("%v-master-0", s.args.Name)
	}

	if s.args.NprocPerNode != "" {
		s.args.Envs["PET_NPROC_PER_NODE"] = s.args.NprocPerNode
	}

	return nil
}

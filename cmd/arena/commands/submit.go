// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/kubeflow/arena/cmd/arena/commands/flags"
	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	nameParameter string
	dryRun        bool

	envs        []string
	selectors   []string
	tolerations []string
	dataset     []string
	dataDirs    []string
	annotations []string
)

// The common parts of the submitAthd
type submitArgs struct {
	// Name       string   `yaml:"name"`       // --name
	NodeSelectors map[string]string `yaml:"nodeSelectors"` // --selector
	Tolerations   []string          `yaml:"tolerations"`   // --toleration
	Image         string            `yaml:"image"`         // --image
	Envs          map[string]string `yaml:"envs"`          // --envs
	// for horovod
	Mode string `yaml:"mode"`
	// --mode
	// SSHPort     int               `yaml:"sshPort"`  // --sshPort
	Retry int `yaml:"retry"` // --retry
	// DataDir  string            `yaml:"dataDir"`  // --dataDir
	DataSet  map[string]string `yaml:"dataset"`
	DataDirs []dataDirVolume   `yaml:"dataDirs"`

	EnableRDMA bool `yaml:"enableRDMA"` // --rdma
	UseENI     bool `yaml:"useENI"`

	Annotations map[string]string `yaml:"annotations"`

	IsNonRoot          bool                      `yaml:"isNonRoot"`
	PodSecurityContext limitedPodSecurityContext `yaml:"podSecurityContext"`
	Project            string                    `yaml:"project,omitempty"`
	User               string                    `yaml:"user,omitempty"`
	PriorityClassName  string                    `yaml:"priorityClassName"`
	// Name       string   `yaml:"name"`       // --name
	Name                string
	Namespace           string
	GPU                 *float64 `yaml:"gpu,omitempty"`
	NodeType            string   `yaml:"node_type,omitempty"`
	Args                []string `yaml:"args,omitempty"`
	CPU                 string   `yaml:"cpu,omitempty"`
	Memory              string   `yaml:"memory,omitempty"`
	EnvironmentVariable []string `yaml:"environment,omitempty"`
}

type dataDirVolume struct {
	HostPath      string `yaml:"hostPath"`
	ContainerPath string `yaml:"containerPath"`
	Name          string `yaml:"name"`
}

type limitedPodSecurityContext struct {
	RunAsUser          int64   `yaml:"runAsUser"`
	RunAsNonRoot       bool    `yaml:"runAsNonRoot"`
	RunAsGroup         int64   `yaml:"runAsGroup"`
	SupplementalGroups []int64 `yaml:"supplementalGroups"`
}

func (s submitArgs) check() error {
	if name == "" {
		return fmt.Errorf("--name must be set")
	}

	// return fmt.Errorf("must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.")
	err := util.ValidateJobName(name)
	if err != nil {
		return err
	}

	if s.PriorityClassName != "" {
		err = util.ValidatePriorityClassName(s.PriorityClassName)
		if err != nil {
			return err
		}
	}

	// if s.DataDir == "" {
	// 	return fmt.Errorf("--dataDir must be set")
	// }

	return nil
}

// transform common parts of submitArgs
func (s *submitArgs) transform() (err error) {
	// 1. handle data dirs
	log.Debugf("dataDir: %v", dataDirs)
	if len(dataDirs) > 0 {
		s.DataDirs = []dataDirVolume{}
		for i, dataDir := range dataDirs {
			hostPath, containerPath, err := util.ParseDataDirRaw(dataDir)
			if err != nil {
				return err
			}
			s.DataDirs = append(s.DataDirs, dataDirVolume{
				Name:          fmt.Sprintf("training-data-%d", i),
				HostPath:      hostPath,
				ContainerPath: containerPath,
			})
		}
	}
	// 2. handle data sets
	log.Debugf("dataset: %v", dataset)
	if len(dataset) > 0 {
		err = util.ValidateDatasets(dataset)
		if err != nil {
			return err
		}
		s.DataSet = transformSliceToMap(dataset, ":")
	}
	// 3. handle annotations
	log.Debugf("annotations: %v", annotations)
	if len(annotations) > 0 {
		s.Annotations = transformSliceToMap(annotations, "=")
		if value, _ := s.Annotations[aliyunENIAnnotation]; value == "true" {
			s.UseENI = true
		}
	}
	// 4. handle PodSecurityContext: runAsUser, runAsGroup, supplementalGroups, runAsNonRoot
	callerUid := os.Getuid()
	callerGid := os.Getgid()
	log.Debugf("Current user: %d", callerUid)
	if callerUid != 0 {
		// only config PodSecurityContext for non-root user
		s.IsNonRoot = true
		s.PodSecurityContext.RunAsNonRoot = true
		s.PodSecurityContext.RunAsUser = int64(callerUid)
		s.PodSecurityContext.RunAsGroup = int64(callerGid)
		groups, _ := os.Getgroups()
		if len(groups) > 0 {
			sg := make([]int64, 0)
			for _, group := range groups {
				sg = append(sg, int64(group))
			}
			s.PodSecurityContext.SupplementalGroups = sg
		}
		log.Debugf("PodSecurityContext %v ", s.PodSecurityContext)
	}
	return nil
}

// get node selectors
func (submitArgs *submitArgs) addNodeSelectors() {
	log.Debugf("node selectors: %v", selectors)
	if len(selectors) == 0 {
		submitArgs.NodeSelectors = map[string]string{}
		return
	}
	submitArgs.NodeSelectors = transformSliceToMap(selectors, "=")
}

// get tolerations labels
func (submitArgs *submitArgs) addTolerations() {
	log.Debugf("tolerations: %v", tolerations)
	if len(tolerations) == 0 {
		submitArgs.Tolerations = []string{}
		return
	}
	submitArgs.Tolerations = []string{}
	for _, taintKey := range tolerations {
		if taintKey == "all" {
			submitArgs.Tolerations = []string{"all"}
			return
		}
		submitArgs.Tolerations = append(submitArgs.Tolerations, taintKey)
	}
}

func (submitArgs *submitArgs) addCommonFlags(command *cobra.Command) {
	var defaultUser string
	currentUser, err := user.Current()
	if err != nil {
		defaultUser = ""
	} else {
		defaultUser = currentUser.Username
	}

	command.Flags().StringVar(&nameParameter, "name", "", "Job name")
	command.Flags().MarkDeprecated("name", "please use positional argument instead")

	flags.AddFloat64NullableFlagP(command.Flags(), &(submitArgs.GPU), "gpu", "g", "Number of GPUs to allocation to the Job.")
	command.Flags().StringVar(&(submitArgs.CPU), "cpu", "", "CPU units to allocate for the job (0.5, 1, .etc)")
	command.Flags().StringVar(&(submitArgs.Memory), "memory", "", "CPU Memory to allocate for this job (1G, 20M, .etc)")
	command.Flags().StringVarP(&(submitArgs.Project), "project", "p", "", "Specifies the Run:AI project to use for this Job.")
	command.Flags().StringVarP(&(submitArgs.User), "user", "u", defaultUser, "Use different user to run the Job.")
	command.Flags().StringVarP(&(submitArgs.Image), "image", "i", "", "Image to use when creating the container for this Job.")
	command.Flags().StringArrayVar(&(submitArgs.Args), "args", []string{}, "Arguments to pass to the command run on container start. Use together with --command.")
	command.Flags().StringVar(&(submitArgs.NodeType), "node-type", "", "Enforce node type affinity by setting a node-type label.")
	command.Flags().StringArrayVarP(&(submitArgs.EnvironmentVariable), "environment", "e", []string{}, "Define environment variable to be set in the container.")
	command.Flags().MarkHidden("user")
	// Will not submit the job to the cluster, just print the template to the screen
	command.Flags().BoolVar(&dryRun, "dry-run", false, "run as dry run")
	command.Flags().MarkHidden("dry-run")

}

func (submitArgs *submitArgs) setCommonRun(cmd *cobra.Command, args []string, kubeClient *client.Client) {
	util.SetLogLevel(logLevel)
	if nameParameter == "" && len(args) >= 1 {
		name = args[0]
	} else {
		name = nameParameter
	}

	submitArgs.Name = name

	namespace, err := flags.GetNamespaceToUseFromProjectFlag(cmd, kubeClient)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	submitArgs.Namespace = namespace
}

var (
	submitLong = `Submit a job.

Available Commands:
  tfjob,tf             Submit a TFJob.
  horovod,hj           Submit a Horovod Job.
  mpijob,mpi           Submit a MPIJob.
  standalonejob,sj     Submit a standalone Job.
  tfserving,tfserving  Submit a Serving Job.
  volcanojob,vj        Submit a VolcanoJob.
    `
)

func transformSliceToMap(sets []string, split string) (valuesMap map[string]string) {
	valuesMap = map[string]string{}
	for _, member := range sets {
		splits := strings.SplitN(member, split, 2)
		if len(splits) == 2 {
			valuesMap[splits[0]] = splits[1]
		}
	}

	return valuesMap
}

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
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kubeflow/arena/pkg/commands/training"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	envs             []string
	selectors        []string
	configFiles      []string
	tolerations      []string
	dataset          []string
	dataDirs         []string
	annotations      []string
	imagePullSecrets []string
)

// The common parts of the submitAthd
type submitArgs struct {
	// Name       string   `yaml:"name"`       // --name
	NodeSelectors map[string]string `yaml:"nodeSelectors"` // --selector
	// key is container path
	ConfigFiles map[string]map[string]configFileInfo `yaml:"configFiles"` // --config-file
	Tolerations []string                             `yaml:"tolerations"` // --toleration
	Image       string                               `yaml:"image"`       // --image
	GPUCount    int                                  `yaml:"gpuCount"`    // --gpuCount
	Envs        map[string]string                    `yaml:"envs"`        // --envs
	WorkingDir  string                               `yaml:"workingDir"`  // --workingDir
	Command     string                               `yaml:"command"`
	// for horovod
	Mode        string `yaml:"mode"`    // --mode
	WorkerCount int    `yaml:"workers"` // --workers
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

	PriorityClassName string `yaml:"priorityClassName"`

	Conscheduling        bool
	PodGroupName         string `yaml:"podGroupName"`
	PodGroupMinAvailable string `yaml:"podGroupMinAvailable"`

	ImagePullSecrets []string `yaml:"imagePullSecrets"` // --image-pull-secrets
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

type configFileInfo struct {
	ContainerFileName string `yaml:"containerFileName"`
	HostFile          string `yaml:"hostFile"`
	Key               string `yaml:"key"`
	ContainerFilePath string `yaml:"containerFilePath"`
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
		err = util.ValidatePriorityClassName(clientset, s.PriorityClassName)
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
// e.g. --data-dir、--data、--annotation、PodSecurityContext
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

	// 5. handle imagePullSecrets, TODO: add validation to check name of imagePullSecrets exists
	if len(s.ImagePullSecrets) > 0 {
		log.Debugf("imagePullSecrets: %v", s.ImagePullSecrets)
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

// get imagePullSecrets
func (submitArgs *submitArgs) addImagePullSecrets() {
	submitArgs.ImagePullSecrets = []string{}
	if len(imagePullSecrets) == 0 {
		if temp, found := arenaConfigs["imagePullSecrets"]; found {
			log.Debugf("imagePullSecrets load from arenaConfigs: %v", temp)
			submitArgs.ImagePullSecrets = strings.Split(temp, ",")
		}
	} else {
		submitArgs.ImagePullSecrets = imagePullSecrets
	}
	log.Debugf("imagePullSecrets: %v", submitArgs.ImagePullSecrets)

	return
}

// this function is used to create config file information
// if the contianer path of file is the same,they will be merged to a configmap
func (submitArgs *submitArgs) addJobConfigFiles() error {
	if len(submitArgs.ConfigFiles) == 0 {
		submitArgs.ConfigFiles = map[string]map[string]configFileInfo{}
	}
	exists := map[string]bool{}
	for ind, val := range configFiles {
		var (
			containerFile string
			err           error
		)
		// use md5 rather than index,the reason is that if user gives a option twice,index can't filter it
		configFileKey := fmt.Sprintf("config-%v", ind)
		files := strings.Split(val, ":")
		hostFile := files[0]
		// change ~ to user home directory
		if strings.Index(hostFile, "~/") == 0 {
			hostFile = strings.Replace(hostFile, "~", os.Getenv("HOME"), -1)
		}
		// change relative path to absolute path
		hostFile, err = filepath.Abs(hostFile)
		if err != nil {
			return err
		}
		// the option gives container path or not,if not, we see the container path is same as host path
		switch len(files) {
		case 1:
			containerFile = hostFile
		case 2:
			containerFile = files[1]
		default:
			return fmt.Errorf("invalid format for assigning config file,it should be '--config-file <host_path_file>:<container_path_file>'")
		}
		if _, ok := exists[fmt.Sprintf("%v:%v", hostFile, containerFile)]; ok {
			continue
		}
		exists[fmt.Sprintf("%v:%v", hostFile, containerFile)] = true
		// if the container path is not absolute path,return error
		if !path.IsAbs(containerFile) {
			return fmt.Errorf("the path of file in container must be absolute path")
		}
		// check the host path file is exist or not
		_, err = os.Stat(hostFile)
		if os.IsNotExist(err) {
			return err
		}
		info := configFileInfo{
			Key:               configFileKey,
			ContainerFileName: path.Base(containerFile),
			ContainerFilePath: path.Dir(containerFile),
			HostFile:          hostFile,
		}
		// classify the files by container path
		containerPathKey := util.Md5(path.Dir(containerFile))[0:15]
		if _, ok := submitArgs.ConfigFiles[containerPathKey]; !ok {
			submitArgs.ConfigFiles[containerPathKey] = map[string]configFileInfo{}
		}
		submitArgs.ConfigFiles[containerPathKey][configFileKey] = info
	}
	return nil
}

// add --set-file flag to 'helm template'
func (submitArgs *submitArgs) addHelmOptions() []string {
	options := []string{}
	for containerPathkey, val := range submitArgs.ConfigFiles {
		for configFileKey, info := range val {
			options = append(options,
				fmt.Sprintf("--set-file configFiles.%v.%v.content=%v", containerPathkey, configFileKey, info.HostFile))
		}
	}
	return options
}

func (submitArgs *submitArgs) addJobInfoToEnv() {
	if len(submitArgs.Envs) == 0 {
		submitArgs.Envs = map[string]string{}
	}
	submitArgs.Envs["workers"] = strconv.Itoa(submitArgs.WorkerCount)
	submitArgs.Envs["gpus"] = strconv.Itoa(submitArgs.GPUCount)
}

// process general parameters for submiting job, like: --image-pull-secrets
func (submitArgs *submitArgs) processCommonFlags() {
	// add tolerations, if given
	submitArgs.addTolerations()
	// add jobinfo to env
	submitArgs.addJobInfoToEnv()
	// add imagePullSecrets ,if given
	submitArgs.addImagePullSecrets()
	// add node selectors, if given
	submitArgs.addNodeSelectors()

}

func (submitArgs *submitArgs) addCommonFlags(command *cobra.Command) {

	// create subcommands
	command.Flags().StringVar(&name, "name", "", "override name")
	command.MarkFlagRequired("name")
	command.Flags().StringVar(&submitArgs.Image, "image", "", "the docker image name of training job")
	// command.MarkFlagRequired("image")
	command.Flags().IntVar(&submitArgs.GPUCount, "gpus", 0,
		"the GPU count of each worker to run the training.")
	// command.Flags().StringVar(&submitArgs.DataDir, "dataDir", "", "the data dir. If you specify /data, it means mounting hostpath /data into container path /data")
	command.Flags().IntVar(&submitArgs.WorkerCount, "workers", 1,
		"the worker number to run the distributed training.")
	command.Flags().IntVar(&submitArgs.Retry, "retry", 0,
		"retry times.")
	// command.MarkFlagRequired("syncSource")
	command.Flags().StringVar(&submitArgs.WorkingDir, "workingDir", "/root", "working directory to extract the code. If using syncMode, the $workingDir/code contains the code")
	command.Flags().MarkDeprecated("workingDir", "please use --working-dir instead")
	command.Flags().StringVar(&submitArgs.WorkingDir, "working-dir", "/root", "working directory to extract the code. If using syncMode, the $workingDir/code contains the code")

	// command.MarkFlagRequired("workingDir")
	command.Flags().StringArrayVarP(&envs, "env", "e", []string{}, "the environment variables")
	command.Flags().StringArrayVarP(&dataset, "data", "d", []string{}, "specify the datasource to mount to the job, like <name_of_datasource>:<mount_point_on_job>")
	command.Flags().StringArrayVar(&dataDirs, "dataDir", []string{}, "the data dir. If you specify /data, it means mounting hostpath /data into container path /data")
	command.Flags().MarkDeprecated("dataDir", "please use --data-dir instead")
	command.Flags().StringArrayVar(&dataDirs, "data-dir", []string{}, "the data dir. If you specify /data, it means mounting hostpath /data into container path /data")

	command.Flags().StringArrayVarP(&annotations, "annotation", "a", []string{}, "the annotations")
	// enable RDMA or not, support hostnetwork for now
	command.Flags().BoolVar(&submitArgs.EnableRDMA, "rdma", false, "enable RDMA")

	// use priority
	command.Flags().StringVarP(&submitArgs.PriorityClassName, "priority", "p", "", "priority class name")
	// toleration
	command.Flags().StringArrayVarP(&tolerations, "toleration", "", []string{}, `tolerate some k8s nodes with taints,usage: "--toleration taint-key" or "--toleration all" `)
	command.Flags().StringArrayVarP(&selectors, "selector", "", []string{}, `assigning jobs to some k8s particular nodes, usage: "--selector=key=value" or "--selector key=value" `)
	command.Flags().StringArrayVarP(&configFiles, "config-file", "", []string{}, `giving configuration files when submiting jobs,usage:"--config-file <host_path_file>:<container_path_file>"`)

	// Using a Private Registry
	command.Flags().StringArrayVarP(&imagePullSecrets, "image-pull-secrets", "", []string{}, `giving names of imagePullSecrets when you want to use a private registry, usage:"--image-pull-secrets <name1>"`)
}

func init() {
	if os.Getenv(CHART_PKG_LOC) != "" {
		standalone_training_chart = filepath.Join(os.Getenv(CHART_PKG_LOC), "training")
	}
}

var (
	submitLong = `Submit a job.

Available Commands:
  tfjob,tf             Submit a TFJob.
  horovod,hj           Submit a Horovod Job.
  mpijob,mpi           Submit a MPIJob.
  pytorchjob,pytorch   Submit a PyTorchJob.
  standalonejob,sj     Submit a standalone Job.
  tfserving,tfserving  Submit a Serving Job.
  volcanojob,vj        Submit a VolcanoJob.
  etjob,et           Submit a ETJob.
    `
)

func NewSubmitCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "submit",
		Short: "Submit a job.",
		Long:  submitLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(training.NewSubmitTFJobCommand())
	command.AddCommand(training.NewSubmitMPIJobCommand())
	// support pytorchjob
	command.AddCommand(training.NewSubmitPytorchJobCommand())
	command.AddCommand(NewSubmitHorovodJobCommand())
	// This will be deprecated soon.
	command.AddCommand(NewSubmitStandaloneJobCommand())
	command.AddCommand(NewSparkApplicationCommand())

	command.AddCommand(NewVolcanoJobCommand())
	command.AddCommand(NewSubmitETJobCommand())

	return command
}

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

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
// limitations under the License
package argsbuilder

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type SubmitArgsBuilder struct {
	args        *types.CommonSubmitArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitArgsBuilder(args *types.CommonSubmitArgs) ArgsBuilder {
	return &SubmitArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
}

func (s *SubmitArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	var (
		envs             []string
		dataSet          []string
		dataDir          []string
		annotations      []string
		labels           []string
		tolerations      []string
		nodeSelectors    []string
		configFiles      []string
		imagePullSecrets []string
	)
	// create subcommands
	// add option --name
	command.Flags().StringVar(&s.args.Name, "name", "", "override name")
	// --name is required
	command.MarkFlagRequired("name")
	// add option --image
	command.Flags().StringVar(&s.args.Image, "image", "", "the docker image name of training job")
	// command.MarkFlagRequired("image")
	// add option --gpus
	command.Flags().IntVar(&s.args.GPUCount, "gpus", 0,
		"the GPU count of each worker to run the training.")
	// add option --workers
	command.Flags().IntVar(&s.args.WorkerCount, "workers", 1,
		"the worker number to run the distributed training.")
	// add option --retry
	command.Flags().IntVar(&s.args.Retry, "retry", 0,
		"retry times.")
	// command.MarkFlagRequired("syncSource")
	// add option --working-dir
	command.Flags().StringVar(&s.args.WorkingDir, "workingDir", "/root", "working directory to extract the code. If using syncMode, the $workingDir/code contains the code")
	command.Flags().MarkDeprecated("workingDir", "please use --working-dir instead")
	command.Flags().StringVar(&s.args.WorkingDir, "working-dir", "/root", "working directory to extract the code. If using syncMode, the $workingDir/code contains the code")

	// command.MarkFlagRequired("workingDir")
	// add option --env,its' value will be get from viper
	command.Flags().StringSliceVarP(&envs, "env", "e", []string{}, "the environment variables")
	// add option --data,its' value will be get from viper
	command.Flags().StringSliceVarP(&dataSet, "data", "d", []string{}, "specify the datasource to mount to the job, like <name_of_datasource>:<mount_point_on_job>")
	// add option --data-dir,its' value will be get from viper
	command.Flags().StringSliceVar(&dataDir, "dataDir", []string{}, "the data dir. If you specify /data, it means mounting hostpath /data into container path /data")
	command.Flags().MarkDeprecated("dataDir", "please use --data-dir instead")
	command.Flags().StringSliceVar(&dataDir, "data-dir", []string{}, "the data dir. If you specify /data, it means mounting hostpath /data into container path /data")
	// add option --annotation,its' value will be get from viper
	command.Flags().StringSliceVarP(&annotations, "annotation", "a", []string{}, "the annotations")
	command.Flags().StringSliceVarP(&labels, "label", "l", []string{}, "specify the label")
	// enable RDMA or not, support hostnetwork for now
	// add option --rdma
	command.Flags().BoolVar(&s.args.EnableRDMA, "rdma", false, "enable RDMA")
	// enable Conscheduling
	command.Flags().BoolVar(&s.args.Conscheduling, "gang", false, "enable gang scheduling")
	// use priority
	command.Flags().StringVarP(&s.args.PriorityClassName, "priority", "p", "", "priority class name")
	// add option --toleration,its' value will be get from viper
	command.Flags().StringSliceVar(&tolerations, "toleration", []string{}, `tolerate some k8s nodes with taints,usage: "--toleration taint-key" or "--toleration all" `)
	// add option --selector,its' value will be get from viper
	command.Flags().StringSliceVar(&nodeSelectors, "selector", []string{}, `assigning jobs to some k8s particular nodes, usage: "--selector=key=value" or "--selector key=value" `)
	// add option --config-file its' value will be get from viper
	command.Flags().StringSliceVar(&configFiles, "config-file", []string{}, `giving configuration files when submiting jobs,usage:"--config-file <host_path_file>:<container_path_file>"`)
	// add option --image-pull-secret its' value will be get from viper,Using a Private Registry
	command.Flags().StringSliceVar(&imagePullSecrets, "image-pull-secret", []string{}, `giving names of imagePullSecret when you want to use a private registry, usage:"--image-pull-secret <name1>"`)
	// add option --shell
	command.Flags().StringVarP(&s.args.Shell, "shell", "", "sh", "specify the linux shell, usage: bash or sh")

	s.AddArgValue("image-pull-secret", &imagePullSecrets).
		AddArgValue("config-file", &configFiles).
		AddArgValue("selector", &nodeSelectors).
		AddArgValue("toleration", &tolerations).
		AddArgValue("annotation", &annotations).
		AddArgValue("data-dir", &dataDir).
		AddArgValue("data", &dataSet).
		AddArgValue("label", &labels).
		AddArgValue("env", &envs)
}

func (s *SubmitArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	// check name
	if err := s.checkNameAndPriorityClassName(); err != nil {
		return err
	}
	// set data set
	if err := s.setDataSet(); err != nil {
		return err
	}
	return nil
}

// Build builds the common submit args
func (s *SubmitArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	// set data dir
	if err := s.setDataDirs(); err != nil {
		return err
	}
	// set annotation
	if err := s.setAnnotations(); err != nil {
		return err
	}
	if err := s.setLabels(); err != nil {
		return err
	}
	if err := s.setUserNameAndUserId(); err != nil {
		return err
	}
	// set config files
	if err := s.setConfigFiles(); err != nil {
		return err
	}
	// set image pull secrets
	if err := s.setImagePullSecrets(); err != nil {
		return err
	}
	if err := s.setEnvs(); err != nil {
		return err
	}
	// add job info to env
	if err := s.setJobInfoToEnv(); err != nil {
		return err
	}
	// set node selectors
	if err := s.setNodeSelectors(); err != nil {
		return err
	}
	// set pod security context
	if err := s.setPodSecurityContext(); err != nil {
		return err
	}
	// set toleration
	if err := s.setTolerations(); err != nil {
		return err
	}
	if err := s.addPodGroupLabel(); err != nil {
		return err
	}
	if err := s.addRequestGPUsToAnnotation(); err != nil {
		return err
	}
	return nil
}

// UpdateArgs is used to update args,this function will be invoked by api
func (s *SubmitArgsBuilder) UpdateArgs(args *types.CommonSubmitArgs) {
	s.args = args
}

// getArgs returns the CommonSubmitArgs
func (s *SubmitArgsBuilder) getArgs() *types.CommonSubmitArgs {
	return s.args
}

// checkNameAndPriorityClassName is used to check the name
func (s *SubmitArgsBuilder) checkNameAndPriorityClassName() error {
	if s.args.Name == "" {
		return fmt.Errorf("--name must be set")
	}
	err := util.ValidateJobName(s.args.Name)
	if err != nil {
		return err
	}
	if s.args.PriorityClassName != "" {
		arenaConfiger := config.GetArenaConfiger()
		err = util.ValidatePriorityClassName(arenaConfiger.GetClientSet(), s.args.PriorityClassName)
		if err != nil {
			return err
		}
	}
	return nil
}

// setDataDirs is used to handle option --data-dir
func (s *SubmitArgsBuilder) setDataDirs() error {
	s.args.DataDirs = []types.DataDirVolume{}
	argKey := "data-dir"
	var dataDirs *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	dataDirs = value.(*[]string)
	log.Debugf("dataDir: %v", *dataDirs)
	for i, dataDir := range *dataDirs {
		hostPath, containerPath, err := util.ParseDataDirRaw(dataDir)
		if err != nil {
			return err
		}
		s.args.DataDirs = append(s.args.DataDirs, types.DataDirVolume{
			Name:          fmt.Sprintf("training-data-%d", i),
			HostPath:      hostPath,
			ContainerPath: containerPath,
		})
	}
	return nil
}

// setDataSets is used to handle option --data
func (s *SubmitArgsBuilder) setDataSet() error {
	s.args.DataSet = map[string]string{}
	argKey := "data"
	var dataSet *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	dataSet = value.(*[]string)
	log.Debugf("dataset: %v", *dataSet)
	if len(*dataSet) <= 0 {
		return nil
	}
	err := util.ValidateDatasets(*dataSet)
	if err != nil {
		return err
	}
	s.args.DataSet = transformSliceToMap(*dataSet, ":")
	return nil
}

// setAnnotations is used to handle option --annotation
func (s *SubmitArgsBuilder) setAnnotations() error {
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	argKey := "annotation"
	var annotations *[]string
	item, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	annotations = item.(*[]string)
	if len(*annotations) <= 0 {
		return nil
	}
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	for key, val := range transformSliceToMap(*annotations, "=") {
		s.args.Annotations[key] = val
	}
	value := s.args.Annotations[aliyunENIAnnotation]
	if value == "true" {
		s.args.UseENI = true
	}
	return nil
}

// setAnnotations is used to handle option --annotation
func (s *SubmitArgsBuilder) setLabels() error {
	if s.args.Labels == nil {
		s.args.Labels = map[string]string{}
	}
	argKey := "label"
	var labels *[]string
	item, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	labels = item.(*[]string)
	if len(*labels) <= 0 {
		return nil
	}
	if s.args.Labels == nil {
		s.args.Labels = map[string]string{}
	}
	for key, val := range transformSliceToMap(*labels, "=") {
		s.args.Labels[key] = val
	}
	return nil
}

func (s *SubmitArgsBuilder) setUserNameAndUserId() error {
	if s.args.Labels == nil {
		s.args.Labels = map[string]string{}
	}
	if s.args.Annotations == nil {
		s.args.Annotations = map[string]string{}
	}
	arenaConfiger := config.GetArenaConfiger()
	user := arenaConfiger.GetUser()
	s.args.Labels[types.UserNameIdLabel] = user.GetId()
	s.args.Annotations[types.UserNameNameLabel] = user.GetName()
	return nil
}

// setPodSecurityContext is used to set pod security context,read uid from os
func (s *SubmitArgsBuilder) setPodSecurityContext() error {
	// handle PodSecurityContext: runAsUser, runAsGroup, supplementalGroups, runAsNonRoot
	callerUid := os.Getuid()
	callerGid := os.Getgid()
	s.args.PodSecurityContext = types.LimitedPodSecurityContext{}
	log.Debugf("Current user: %d", callerUid)
	// if user is root,return
	if callerUid == 0 {
		return nil
	}
	// only config PodSecurityContext for non-root user
	s.args.IsNonRoot = true
	s.args.PodSecurityContext.RunAsNonRoot = true
	s.args.PodSecurityContext.RunAsUser = int64(callerUid)
	s.args.PodSecurityContext.RunAsGroup = int64(callerGid)
	groups, _ := os.Getgroups()
	if len(groups) <= 0 {
		return nil
	}
	sg := make([]int64, 0)
	for _, group := range groups {
		sg = append(sg, int64(group))
	}
	s.args.PodSecurityContext.SupplementalGroups = sg
	log.Debugf("PodSecurityContext %v ", s.args.PodSecurityContext)
	return nil
}

// setNodeSelectors is used to handle option --selector
func (s *SubmitArgsBuilder) setNodeSelectors() error {
	if s.args.NodeSelectors == nil {
		s.args.NodeSelectors = map[string]string{}
	}
	argKey := "selector"
	var nodeSelectors *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	nodeSelectors = value.(*[]string)
	log.Debugf("node selectors: %v", *nodeSelectors)
	s.args.NodeSelectors = transformSliceToMap(*nodeSelectors, "=")
	return nil
}

// setConfigFiles is used to handle option --config-file
func (s *SubmitArgsBuilder) setConfigFiles() error {
	s.args.ConfigFiles = map[string]map[string]types.ConfigFileInfo{}
	if s.args.HelmOptions == nil {
		s.args.HelmOptions = []string{}
	}
	argKey := "config-file"
	var configFiles *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	configFiles = value.(*[]string)
	exists := map[string]bool{}
	for ind, val := range *configFiles {
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
		info := types.ConfigFileInfo{
			Key:               configFileKey,
			ContainerFileName: path.Base(containerFile),
			ContainerFilePath: path.Dir(containerFile),
			HostFile:          hostFile,
		}
		// classify the files by container path
		containerPathKey := util.Md5(path.Dir(containerFile))[0:15]
		if _, ok := s.args.ConfigFiles[containerPathKey]; !ok {
			s.args.ConfigFiles[containerPathKey] = map[string]types.ConfigFileInfo{}
		}
		s.args.ConfigFiles[containerPathKey][configFileKey] = info
	}
	for containerPathkey, val := range s.args.ConfigFiles {
		for configFileKey, info := range val {
			s.args.HelmOptions = append(s.args.HelmOptions,
				fmt.Sprintf("--set-file configFiles.%v.%v.content=%v", containerPathkey, configFileKey, info.HostFile))
		}
	}
	return nil
}

// setTolerations is used to handle option --toleration
func (s *SubmitArgsBuilder) setTolerations() error {
	if s.args.Tolerations == nil {
		s.args.Tolerations = []string{}
	}
	argKey := "toleration"
	var tolerations *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	tolerations = value.(*[]string)
	log.Debugf("tolerations: %v", *tolerations)
	for _, taintKey := range *tolerations {
		if taintKey == "all" {
			s.args.Tolerations = []string{"all"}
			return nil
		}
		s.args.Tolerations = append(s.args.Tolerations, taintKey)
	}
	return nil
}

// setImagePullSecrets is used to set
func (s *SubmitArgsBuilder) setImagePullSecrets() error {
	s.args.ImagePullSecrets = []string{}
	argKey := "image-pull-secret"
	var imagePullSecrets *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	imagePullSecrets = value.(*[]string)

	if len(*imagePullSecrets) == 0 {
		arenaConfig := config.GetArenaConfiger().GetConfigsFromConfigFile()
		if temp, found := arenaConfig["imagePullSecrets"]; found {
			log.Debugf("imagePullSecrets load from arenaConfigs: %v", temp)
			s.args.ImagePullSecrets = strings.Split(temp, ",")
		}
	} else {
		s.args.ImagePullSecrets = *imagePullSecrets
	}
	log.Debugf("imagePullSecrets: %v", s.args.ImagePullSecrets)
	return nil
}

func (s *SubmitArgsBuilder) setEnvs() error {
	argKey := "env"
	var envs *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	envs = value.(*[]string)
	if s.args.Envs == nil {
		s.args.Envs = map[string]string{}
	}
	for key, val := range transformSliceToMap(*envs, "=") {
		s.args.Envs[key] = val
	}
	return nil
}

func (s *SubmitArgsBuilder) setJobInfoToEnv() error {
	if s.args.Envs == nil {
		s.args.Envs = map[string]string{}
	}
	s.args.Envs["workers"] = strconv.Itoa(s.args.WorkerCount)
	s.args.Envs["gpus"] = strconv.Itoa(s.args.GPUCount)
	return nil
}

func (s *SubmitArgsBuilder) addPodGroupLabel() error {
	if s.args.Conscheduling {
		s.args.PodGroupName = fmt.Sprintf("%v_%v", s.args.TrainingType, s.args.Name)
		s.args.PodGroupMinAvailable = fmt.Sprintf("%v", s.args.WorkerCount)
	}
	return nil
}

func (s *SubmitArgsBuilder) addRequestGPUsToAnnotation() error {
	s.args.Annotations[types.RequestGPUsOfJobAnnoKey] = fmt.Sprintf("%v", s.args.WorkerCount*s.args.GPUCount)
	return nil
}

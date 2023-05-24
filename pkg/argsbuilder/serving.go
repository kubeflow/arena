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
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServingArgsBuilder struct {
	args        *types.CommonServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewServingArgsBuilder(args *types.CommonServingArgs) ArgsBuilder {
	s := &ServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	return s
}

func (s *ServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *ServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *ServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *ServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	var (
		envs               []string
		dataset            []string
		datadir            []string
		dataSubpathExpr    []string
		tempDir            []string
		tempDirSubpathExpr []string
		annotations        []string
		tolerations        []string
		labels             []string
		selectors          []string
		configFiles        []string
		imagePullSecrets   []string
	)
	defaultImage := ""
	item, ok := s.argValues["default-image"]
	if ok {
		defaultImage = item.(string)
	}
	command.Flags().StringVar(&s.args.Image, "image", defaultImage, "the docker image name of serving job")
	command.Flags().StringVar(&s.args.ImagePullPolicy, "imagePullPolicy", "IfNotPresent", "the policy to pull the image, and the default policy is IfNotPresent")
	command.Flags().MarkDeprecated("imagePullPolicy", "please use --image-pull-policy instead")
	command.Flags().StringVar(&s.args.ImagePullPolicy, "image-pull-policy", "IfNotPresent", "the policy to pull the image, and the default policy is IfNotPresent")

	command.Flags().IntVar(&s.args.GPUCount, "gpus", 0, "the limit GPU count of each replica to run the serve.")
	command.Flags().IntVar(&s.args.GPUMemory, "gpumemory", 0, "the limit GPU memory of each replica to run the serve.")
	command.Flags().IntVar(&s.args.GPUCore, "gpucore", 0, "the limit GPU core of each replica to run the serve.")
	command.Flags().StringVar(&s.args.Cpu, "cpu", "", "the request cpu of each replica to run the serve.")
	command.Flags().StringVar(&s.args.Memory, "memory", "", "the request memory of each replica to run the serve.")
	command.Flags().IntVar(&s.args.Replicas, "replicas", 1, "the replicas number of the serve job.")
	command.Flags().StringVar(&s.args.ShareMemory, "share-memory", "", "the request share memory of each replica to run the serve.")
	// add option --image-pull-secret its' value will be get from viper,Using a Private Registry
	command.Flags().StringArrayVar(&imagePullSecrets, "image-pull-secret", []string{}, `giving names of imagePullSecret when you want to use a private registry, usage:"--image-pull-secret <name1>"`)

	command.Flags().StringArrayVarP(&envs, "env", "e", []string{}, "the environment variables")

	command.Flags().BoolVar(&s.args.EnableIstio, "enableIstio", false, "enable Istio for serving or not (disable Istio by default)")
	command.Flags().MarkDeprecated("enableIstio", "please use --enable-istio instead")
	command.Flags().BoolVar(&s.args.EnableIstio, "enable-istio", false, "enable Istio for serving or not (disable Istio by default)")

	command.Flags().BoolVar(&s.args.ExposeService, "exposeService", false, "expose service using Istio gateway for external access or not (not expose by default)")
	command.Flags().MarkDeprecated("exposeService", "please use --expose-service instead")
	command.Flags().BoolVar(&s.args.ExposeService, "expose-service", false, "expose service using Istio gateway for external access or not (not expose by default)")

	command.Flags().StringVar(&s.args.Name, "servingName", "", "the serving name")
	command.Flags().MarkDeprecated("servingName", "please use --name instead")
	command.Flags().StringVar(&s.args.Name, "name", "", "the serving name")

	command.Flags().StringVar(&s.args.Version, "servingVersion", "", "the serving version")
	command.Flags().MarkDeprecated("servingVersion", "please use --version instead")
	command.Flags().StringVar(&s.args.Version, "version", "", "the serving version")

	command.Flags().StringArrayVarP(&dataset, "data", "d", []string{}, "specify the trained models datasource to mount for serving, like <name_of_datasource>:<mount_point_on_job>")
	command.Flags().StringArrayVarP(&dataSubpathExpr, "data-subpath-expr", "", []string{}, "specify the datasource subpath to mount to the job by expression, like <name_of_datasource>:<mount_subpath_expr>")
	command.Flags().StringArrayVarP(&datadir, "data-dir", "", []string{}, "specify the trained models datasource on host to mount for serving, like <host_path>:<mount_point_on_job>")
	command.Flags().StringArrayVarP(&tempDirSubpathExpr, "temp-dir-subpath-expr", "", []string{}, "specify the datasource subpath to mount to the pod by expression, like <empty_dir_name>:<mount_subpath_expr>")
	command.Flags().StringArrayVarP(&tempDir, "temp-dir", "", []string{}, "specify the deployment empty dir, like <empty_dir_name>:<mount_point_on_pod>")
	command.MarkFlagRequired("name")

	command.Flags().StringArrayVarP(&annotations, "annotation", "a", []string{}, `specify the annotations, usage: "--annotation=key=value" or "--annotation key=value"`)
	command.Flags().StringArrayVarP(&labels, "label", "l", []string{}, "specify the labels")
	command.Flags().StringArrayVarP(&tolerations, "toleration", "", []string{}, `tolerate some k8s nodes with taints,usage: "--toleration key=value:effect,operator" or "--toleration all" `)
	command.Flags().StringArrayVarP(&selectors, "selector", "", []string{}, `assigning jobs to some k8s particular nodes, usage: "--selector=key=value" or "--selector key=value" `)

	// add option --config-file its' value will be get from viper
	command.Flags().StringArrayVar(&configFiles, "config-file", []string{}, `giving configuration files when serving model, usage:"--config-file <host_path_file>:<container_path_file>"`)

	// add option --shell
	command.Flags().StringVarP(&s.args.Shell, "shell", "", "sh", "specify the linux shell, usage: bash or sh")

	s.AddArgValue("annotation", &annotations).
		AddArgValue("toleration", &tolerations).
		AddArgValue("label", &labels).
		AddArgValue("selector", &selectors).
		AddArgValue("data", &dataset).
		AddArgValue("data-subpath-expr", &dataSubpathExpr).
		AddArgValue("data-dir", &datadir).
		AddArgValue("temp-dir-subpath-expr", &tempDirSubpathExpr).
		AddArgValue("temp-dir", &tempDir).
		AddArgValue("env", &envs).
		AddArgValue("config-file", &configFiles).
		AddArgValue("image-pull-secret", &imagePullSecrets)
}

func (s *ServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	if err := s.checkNamespace(); err != nil {
		return err
	}
	if err := s.validateIstioEnablement(); err != nil {
		return err
	}
	if err := s.setDataSet(); err != nil {
		return err
	}
	if err := s.setDataSubpathExprs(); err != nil {
		return err
	}
	if err := s.setDataDirs(); err != nil {
		return err
	}
	if err := s.setTempDirSubpathExprs(); err != nil {
		return err
	}
	if err := s.setTempDirs(); err != nil {
		return err
	}
	if err := s.setEnvs(); err != nil {
		return err
	}
	if err := s.setAnnotations(); err != nil {
		return err
	}
	if err := s.setLabels(); err != nil {
		return err
	}
	if err := s.setNodeSelectors(); err != nil {
		return err
	}
	if err := s.setTolerations(); err != nil {
		return err
	}
	if err := s.setServingVersion(); err != nil {
		return err
	}
	if err := s.setUserNameAndUserId(); err != nil {
		return err
	}
	// set image pull secrets
	if err := s.setImagePullSecrets(); err != nil {
		return err
	}
	// set config files
	if err := s.setConfigFiles(); err != nil {
		return err
	}
	if err := s.disabledNvidiaENVWithNoneGPURequest(); err != nil {
		return err
	}
	if err := s.check(); err != nil {
		return err
	}
	if err := s.checkGPUCore(); err != nil {
		return err
	}
	return nil
}

func (s *ServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}

	return nil
}

func (s *ServingArgsBuilder) check() error {
	if s.args.GPUCount < 0 {
		return fmt.Errorf("--gpus is invalid")
	}
	if s.args.GPUMemory < 0 {
		return fmt.Errorf("--gpumemory is invalid")
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
	if s.args.ShareMemory != "" {
		_, err := resource.ParseQuantity(s.args.ShareMemory)
		if err != nil {
			return fmt.Errorf("--share-memory is invalid")
		}
	}

	return s.checkServiceExists()
}

// setImagePullSecrets is used to set
func (s *ServingArgsBuilder) setImagePullSecrets() error {
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

// setDataSets is used to handle option --data
func (s *ServingArgsBuilder) setDataSet() error {
	s.args.ModelDirs = map[string]string{}
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
	s.args.ModelDirs = transformSliceToMap(*dataSet, ":")
	return nil
}

// setDataSubpathExprs is used to handle option --data-subpath-expr
func (s *ServingArgsBuilder) setDataSubpathExprs() error {
	s.args.DataSubpathExprs = map[string]string{}
	argKey := "data-subpath-expr"
	var dataSubPathExprs *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	dataSubPathExprs = value.(*[]string)
	log.Debugf("setDataSubpathExprs: %v", *dataSubPathExprs)
	if len(*dataSubPathExprs) <= 0 {
		return nil
	}
	s.args.DataSubpathExprs = transformSliceToMap(*dataSubPathExprs, ":")
	return nil
}

// setDataSubpathExprs is used to handle option --temp-dir-subpath-expr
func (s *ServingArgsBuilder) setTempDirSubpathExprs() error {
	s.args.TempDirSubpathExpr = map[string]string{}
	argKey := "temp-dir-subpath-expr"
	var tempDirSubPathExprs *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	tempDirSubPathExprs = value.(*[]string)
	log.Debugf("setDataSubpathExprs: %v", *tempDirSubPathExprs)
	if len(*tempDirSubPathExprs) <= 0 {
		return nil
	}
	s.args.TempDirSubpathExpr = transformSliceToMap(*tempDirSubPathExprs, ":")
	return nil
}

// setDataDirs is used to handle option --data-dir
func (s *ServingArgsBuilder) setDataDirs() error {
	s.args.HostVolumes = []types.DataDirVolume{}
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
		s.args.HostVolumes = append(s.args.HostVolumes, types.DataDirVolume{
			Name:          fmt.Sprintf("serving-data-%d", i),
			HostPath:      hostPath,
			ContainerPath: containerPath,
		})
	}
	return nil
}

// setTempDirs is used to handle option --temp-dir
// setDataSets is used to handle option --data
func (s *ServingArgsBuilder) setTempDirs() error {
	s.args.TempDirs = map[string]string{}
	argKey := "temp-dir"
	var tempDirs *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	tempDirs = value.(*[]string)
	log.Debugf("tempDirs: %v", *tempDirs)
	if len(*tempDirs) <= 0 {
		return nil
	}
	s.args.TempDirs = transformSliceToMap(*tempDirs, ":")
	return nil
}

// setAnnotations is used to handle option --annotation
func (s *ServingArgsBuilder) setAnnotations() error {
	s.args.Annotations = map[string]string{}
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
	s.args.Annotations = transformSliceToMap(*annotations, "=")
	return nil
}

// setLabels is used to handle option --label
func (s *ServingArgsBuilder) setLabels() error {
	s.args.Labels = map[string]string{}
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
	s.args.Labels = transformSliceToMap(*labels, "=")
	return nil
}

func (s *ServingArgsBuilder) setUserNameAndUserId() error {
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

// setNodeSelectors is used to handle option --selector
func (s *ServingArgsBuilder) setNodeSelectors() error {
	s.args.NodeSelectors = map[string]string{}
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

// setTolerations is used to handle option --toleration
func (s *ServingArgsBuilder) setTolerations() error {
	if s.args.Tolerations == nil {
		s.args.Tolerations = []types.TolerationArgs{}
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
			s.args.Tolerations = append(s.args.Tolerations, types.TolerationArgs{
				Operator: "Exists",
			})
			return nil
		}
		tolerationArg, err := parseTolerationString(taintKey)
		if err != nil {
			log.Debugf(err.Error())
			continue
		}
		s.args.Tolerations = append(s.args.Tolerations, *tolerationArg)
	}
	return nil
}

func (s *ServingArgsBuilder) setEnvs() error {
	argKey := "env"
	var envs *[]string
	value, ok := s.argValues[argKey]
	if !ok {
		return nil
	}
	envs = value.(*[]string)
	s.args.Envs = transformSliceToMap(*envs, "=")
	return nil
}

func (s *ServingArgsBuilder) setServingVersion() error {
	if s.args.Version == "" {
		t := time.Now()
		s.args.Version = fmt.Sprint(t.Format("200601021504"))
	}
	return nil
}

// setConfigFiles is used to handle option --config-file
func (s *ServingArgsBuilder) setConfigFiles() error {
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
	for containerPathKey, val := range s.args.ConfigFiles {
		for configFileKey, info := range val {
			s.args.HelmOptions = append(s.args.HelmOptions,
				fmt.Sprintf("--set-file configFiles.%v.%v.content=%v", containerPathKey, configFileKey, info.HostFile))
			tmp := fmt.Sprintf("--set-file configFiles.%v.%v.content=%v", containerPathKey, configFileKey, info.HostFile)
			fmt.Println(tmp)
		}
	}
	return nil
}

func (s *ServingArgsBuilder) validateIstioEnablement() error {
	log.Debugf("--enable-Istio=%t is specified.", s.args.EnableIstio)
	if !s.args.EnableIstio {
		return nil
	}
	var reg *regexp.Regexp
	reg = regexp.MustCompile(regexp4serviceName)
	matched := reg.MatchString(s.args.Name)
	if !matched {
		return fmt.Errorf("--name should be numbers, letters, dashes, and underscores ONLY")
	}
	log.Debugf("--version=%s is specified.", s.args.Version)
	if s.args.Version == "" {
		return fmt.Errorf("--version must be specified if --enable-istio=true")
	}
	return nil
}

// checkServiceExists is used to check services,must execute after function checkNamespace
func (s *ServingArgsBuilder) checkServiceExists() error {
	client := config.GetArenaConfiger().GetClientSet()
	_, err := client.CoreV1().Services(s.args.Namespace).Get(context.TODO(), s.args.Name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		s.args.ModelServiceExists = false
	} else {
		s.args.ModelServiceExists = true
	}
	return nil
}

func (s *ServingArgsBuilder) checkNamespace() error {
	if s.args.Namespace == "" {
		return fmt.Errorf("not set namespace,please set it")
	}
	log.Debugf("namespace is %v", s.args.Namespace)
	return nil
}

func (s *ServingArgsBuilder) disabledNvidiaENVWithNoneGPURequest() error {
	if s.args.Envs == nil {
		s.args.Envs = map[string]string{}
	}
	if s.args.GPUCount == 0 && s.args.GPUMemory == 0 && s.args.GPUCore == 0 {
		s.args.Envs["NVIDIA_VISIBLE_DEVICES"] = "void"
	}
	return nil
}

func (s *ServingArgsBuilder) checkGPUCore() error {
	if s.args.GPUCore%5 != 0 {
		return fmt.Errorf("GPUCore should be the multiple of 5")
	}
	return nil
}

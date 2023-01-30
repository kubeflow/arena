package argsbuilder

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"reflect"
	"strings"
)

type EvaluateJobArgsBuilder struct {
	args        *types.EvaluateJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewEvaluateJobArgsBuilder(args *types.EvaluateJobArgs) ArgsBuilder {
	e := &EvaluateJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}

	e.AddSubBuilder(
		NewSubmitSyncCodeArgsBuilder(&e.args.SubmitSyncCodeArgs),
	)

	return e
}

func (e *EvaluateJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*e)), ".")
	return items[len(items)-1]
}

func (e *EvaluateJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		e.subBuilders[b.GetName()] = b
	}
	return e
}

func (e *EvaluateJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range e.subBuilders {
		e.subBuilders[name].AddArgValue(key, value)
	}
	e.argValues[key] = value
	return e
}

func (e *EvaluateJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range e.subBuilders {
		e.subBuilders[name].AddCommandFlags(command)
	}

	var (
		envs             []string
		dataDir          []string
		dataSources      []string
		annotations      []string
		labels           []string
		tolerations      []string
		nodeSelectors    []string
		imagePullSecrets []string
	)

	// evaluate job arguments
	command.Flags().StringVar(&e.args.Name, "name", "", "the evaluate job name")
	command.Flags().StringVar(&e.args.Namespace, "namespace", "", "the evaluate job namespace")
	command.Flags().StringVar(&e.args.ModelName, "model-name", "", "the model name to evaluate")
	command.Flags().StringVar(&e.args.ModelVersion, "model-version", "", "the model version to evaluate")
	command.Flags().StringVar(&e.args.ModelPath, "model-path", "", "the model path to evaluate in the container")
	command.Flags().StringVar(&e.args.DatasetPath, "dataset-path", "", "the model version to evaluate")
	command.Flags().StringVar(&e.args.MetricsPath, "metrics-path", "", "the evaluate result saved path")
	command.Flags().StringVar(&e.args.Image, "image", "", "the evaluate image")
	command.Flags().StringVar(&e.args.WorkingDir, "working-dir", "/root", "working directory to extract the code. If using syncMode, the $workingDir/code contains the code")
	command.Flags().IntVar(&e.args.GPUCount, "gpus", 0, "the limit GPU count of each replica to run the evaluate job.")
	command.Flags().StringVar(&e.args.Cpu, "cpu", "", "the request cpu of each replica to run the evaluate job.")
	command.Flags().StringVar(&e.args.Memory, "memory", "", "the request memory of each replica to run the evaluate job.")

	command.Flags().StringArrayVarP(&envs, "env", "e", []string{}, "the environment variables")
	command.Flags().StringArrayVarP(&annotations, "annotation", "a", []string{}, "the annotations")
	command.Flags().StringArrayVarP(&labels, "label", "", []string{}, "the labels")
	command.Flags().StringArrayVar(&tolerations, "toleration", []string{}, `tolerate some k8s nodes with taints,usage: "--toleration key=value:effect,operator" or "--toleration all" `)
	// add option --selector, it's value will be get from viper
	command.Flags().StringArrayVar(&nodeSelectors, "selector", []string{}, `assigning jobs to some k8s particular nodes, usage: "--selector=key=value" or "--selector key=value" `)
	// add option --image-pull-secret it's value will be get from viper,Using a Private Registry
	command.Flags().StringArrayVar(&imagePullSecrets, "image-pull-secret", []string{}, `giving names of imagePullSecret when you want to use a private registry, usage:"--image-pull-secret <name1>"`)
	command.Flags().StringArrayVar(&dataDir, "data-dir", []string{}, "the data dir. If you specify /data, it means mounting hostpath /data into container path /data")
	command.Flags().StringArrayVarP(&dataSources, "data", "d", []string{}, "specify the datasource to mount to the job, like <name_of_datasource>:<mount_point_on_job>")

	e.AddArgValue("image-pull-secret", &imagePullSecrets).
		AddArgValue("selector", &nodeSelectors).
		AddArgValue("toleration", &tolerations).
		AddArgValue("annotation", &annotations).
		AddArgValue("label", &labels).
		AddArgValue("data-dir", &dataDir).
		AddArgValue("data", &dataSources).
		AddArgValue("env", &envs)
}

func (e *EvaluateJobArgsBuilder) PreBuild() error {
	for name := range e.subBuilders {
		if err := e.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}

	// set data set
	if err := e.setDataSources(); err != nil {
		return err
	}

	return nil
}

func (e *EvaluateJobArgsBuilder) Build() error {
	for name := range e.subBuilders {
		if err := e.subBuilders[name].Build(); err != nil {
			return err
		}
	}

	if err := e.check(); err != nil {
		return err
	}

	// set data dir
	if err := e.setDataDirs(); err != nil {
		return err
	}
	// set annotations
	if err := e.setAnnotations(); err != nil {
		return err
	}
	// set labels
	if err := e.setLabels(); err != nil {
		return err
	}
	// set image pull secrets
	if err := e.setImagePullSecrets(); err != nil {
		return err
	}
	if err := e.setEnvs(); err != nil {
		return err
	}
	// set node selectors
	if err := e.setNodeSelectors(); err != nil {
		return err
	}
	// set toleration
	if err := e.setTolerations(); err != nil {
		return err
	}

	return nil
}

func (e *EvaluateJobArgsBuilder) check() error {
	if e.args.ModelName == "" {
		return fmt.Errorf("--model-name must be set ")
	}
	if e.args.ModelPath == "" {
		return fmt.Errorf("--model-path must be set ")
	}
	if e.args.DatasetPath == "" {
		return fmt.Errorf("--dataset-path must be set ")
	}
	if e.args.MetricsPath == "" {
		return fmt.Errorf("--metrics-path must be set ")
	}
	return nil
}

// setDataSources is used to handle option --data
func (e *EvaluateJobArgsBuilder) setDataSources() error {
	e.args.DataSources = map[string]string{}
	argKey := "data"
	var dataSet *[]string
	value, ok := e.argValues[argKey]
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
	e.args.DataSources = transformSliceToMap(*dataSet, ":")
	return nil
}

// setDataDirs is used to handle option --data-dir
func (e *EvaluateJobArgsBuilder) setDataDirs() error {
	e.args.DataDirs = []types.DataDirVolume{}
	argKey := "data-dir"
	var dataDirs *[]string
	value, ok := e.argValues[argKey]
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
		e.args.DataDirs = append(e.args.DataDirs, types.DataDirVolume{
			Name:          fmt.Sprintf("evaluate-data-%d", i),
			HostPath:      hostPath,
			ContainerPath: containerPath,
		})
	}
	return nil
}

// setAnnotations is used to handle option --annotation
func (e *EvaluateJobArgsBuilder) setAnnotations() error {
	e.args.Annotations = map[string]string{}
	argKey := "annotation"
	var annotations *[]string
	item, ok := e.argValues[argKey]
	if !ok {
		return nil
	}
	annotations = item.(*[]string)
	if len(*annotations) <= 0 {
		return nil
	}
	if e.args.Annotations == nil {
		e.args.Annotations = map[string]string{}
	}
	for key, val := range transformSliceToMap(*annotations, "=") {
		e.args.Annotations[key] = val
	}
	return nil
}

// setLabels is used to handle option --label
func (e *EvaluateJobArgsBuilder) setLabels() error {
	e.args.Labels = map[string]string{}
	argKey := "label"
	var labels *[]string
	item, ok := e.argValues[argKey]
	if !ok {
		return nil
	}
	labels = item.(*[]string)
	if len(*labels) <= 0 {
		return nil
	}
	e.args.Labels = transformSliceToMap(*labels, "=")
	return nil
}

// setNodeSelectors is used to handle option --selector
func (e *EvaluateJobArgsBuilder) setNodeSelectors() error {
	e.args.NodeSelectors = map[string]string{}
	argKey := "selector"
	var nodeSelectors *[]string
	value, ok := e.argValues[argKey]
	if !ok {
		return nil
	}
	nodeSelectors = value.(*[]string)
	log.Debugf("node selectors: %v", *nodeSelectors)
	e.args.NodeSelectors = transformSliceToMap(*nodeSelectors, "=")
	return nil
}

// setTolerations is used to handle option --toleration
func (e *EvaluateJobArgsBuilder) setTolerations() error {
	if e.args.Tolerations == nil {
		e.args.Tolerations = []types.TolerationArgs{}
	}
	argKey := "toleration"
	var tolerations *[]string
	value, ok := e.argValues[argKey]
	if !ok {
		return nil
	}
	tolerations = value.(*[]string)
	log.Debugf("tolerations: %v", *tolerations)
	for _, taintKey := range *tolerations {
		if taintKey == "all" {
			e.args.Tolerations = append(e.args.Tolerations, types.TolerationArgs{
				Operator: "Exists",
			})
			return nil
		}
		tolerationArg, err := parseTolerationString(taintKey)
		if err != nil {
			log.Debugf(err.Error())
			continue
		}
		e.args.Tolerations = append(e.args.Tolerations, *tolerationArg)
	}
	return nil
}

// setImagePullSecrets is used to set
func (e *EvaluateJobArgsBuilder) setImagePullSecrets() error {
	e.args.ImagePullSecrets = []string{}
	argKey := "image-pull-secret"
	var imagePullSecrets *[]string
	value, ok := e.argValues[argKey]
	if !ok {
		return nil
	}
	imagePullSecrets = value.(*[]string)

	if len(*imagePullSecrets) == 0 {
		arenaConfig := config.GetArenaConfiger().GetConfigsFromConfigFile()
		if temp, found := arenaConfig["imagePullSecrets"]; found {
			log.Debugf("imagePullSecrets load from arenaConfigs: %v", temp)
			e.args.ImagePullSecrets = strings.Split(temp, ",")
		}
	} else {
		e.args.ImagePullSecrets = *imagePullSecrets
	}
	log.Debugf("imagePullSecrets: %v", e.args.ImagePullSecrets)
	return nil
}

func (e *EvaluateJobArgsBuilder) setEnvs() error {
	argKey := "env"
	var envs *[]string
	value, ok := e.argValues[argKey]
	if !ok {
		return nil
	}
	envs = value.(*[]string)
	if e.args.Envs == nil {
		e.args.Envs = map[string]string{}
	}
	for key, val := range transformSliceToMap(*envs, "=") {
		e.args.Envs[key] = val
	}
	return nil
}

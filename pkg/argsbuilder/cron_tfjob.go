package argsbuilder

import (
	"context"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/argsbuilder/runtime"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
)

type CronTFJobArgsBuilder struct {
	args         *types.CronTFJobArgs
	argValues    map[string]interface{}
	subBuilders  map[string]ArgsBuilder
	tfjobBuilder ArgsBuilder
}

func NewCronTFJobArgsBuilder(args *types.CronTFJobArgs) ArgsBuilder {
	args.TrainingType = types.TFTrainingJob
	c := &CronTFJobArgsBuilder{
		args:         args,
		argValues:    map[string]interface{}{},
		subBuilders:  map[string]ArgsBuilder{},
		tfjobBuilder: NewSubmitTFJobArgsBuilder(&args.SubmitTFJobArgs),
	}
	c.AddSubBuilder(
		NewSubmitArgsBuilder(&c.args.CommonSubmitArgs),
		NewSubmitSyncCodeArgsBuilder(&c.args.SubmitSyncCodeArgs),
		NewSubmitTensorboardArgsBuilder(&c.args.SubmitTensorboardArgs),
	)
	return c
}

func (c *CronTFJobArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*c)), ".")
	return items[len(items)-1]
}

func (c *CronTFJobArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		c.subBuilders[b.GetName()] = b
	}
	return c
}

func (c *CronTFJobArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range c.subBuilders {
		c.subBuilders[name].AddArgValue(key, value)
	}
	c.argValues[key] = value
	return c
}

func (c *CronTFJobArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range c.subBuilders {
		c.subBuilders[name].AddCommandFlags(command)
	}
	var (
		workerSelectors    []string
		chiefSelectors     []string
		psSelectors        []string
		evaluatorSelectors []string
		roleSequence       string
	)

	// cron task arguments
	command.Flags().StringVar(&c.args.Schedule, "schedule", "", "the schedule of cron task")
	command.Flags().StringVar(&c.args.ConcurrencyPolicy, "concurrency-policy", "", "specifies how to treat concurrent executions of a task")
	command.Flags().BoolVar(&c.args.Suspend, "suspend", false, "if suspend the cron task")
	command.Flags().StringVar(&c.args.Deadline, "deadline", "", "the timestamp that a cron job can keep scheduling util then")
	command.Flags().IntVar(&c.args.HistoryLimit, "history-limit", 0, "the number of finished job history to retain")

	// tfjob arguments
	command.Flags().StringVar(&c.args.WorkerImage, "worker-image", "", "the docker image for tensorflow workers")

	command.Flags().StringVar(&c.args.PSImage, "ps-image", "", "the docker image for tensorflow workers")

	command.Flags().IntVar(&c.args.PSCount, "ps", 0, "the number of the parameter servers.")

	command.Flags().IntVar(&c.args.PSPort, "ps-port", 0, "the port of the parameter server.")

	command.Flags().IntVar(&c.args.WorkerPort, "worker-port", 0, "the port of the worker.")

	command.Flags().StringVar(&c.args.WorkerCpu, "worker-cpu", "", "the cpu resource to use for the worker, like 1 for 1 core.")

	command.Flags().StringVar(&c.args.WorkerMemory, "worker-memory", "", "the memory resource to use for the worker, like 1Gi.")

	command.Flags().StringVar(&c.args.PSCpu, "ps-cpu", "", "the cpu resource to use for the parameter servers, like 1 for 1 core.")

	command.Flags().IntVar(&c.args.PSGpu, "ps-gpus", 0, "the gpu resource to use for the parameter servers, like 1 for 1 gpu.")

	command.Flags().StringVar(&c.args.PSMemory, "ps-memory", "", "the memory resource to use for the parameter servers, like 1Gi.")
	// How to clean up Task
	command.Flags().StringVar(&c.args.CleanPodPolicy, "clean-task-policy", "Running", "How to clean tasks after Training is done, only support Running, None.")

	// Estimator
	command.Flags().BoolVar(&c.args.UseChief, "chief", false, "enable chief, which is required for estimator.")
	command.Flags().BoolVar(&c.args.UseEvaluator, "evaluator", false, "enable evaluator, which is optional for estimator.")
	command.Flags().StringVar(&c.args.ChiefCpu, "chief-cpu", "", "the cpu resource to use for the Chief, like 1 for 1 core.")

	command.Flags().StringVar(&c.args.ChiefMemory, "chief-memory", "", "the memory resource to use for the Chief, like 1Gi.")

	command.Flags().StringVar(&c.args.EvaluatorCpu, "evaluator-cpu", "", "the cpu resource to use for the evaluator, like 1 for 1 core.")

	command.Flags().StringVar(&c.args.EvaluatorMemory, "evaluator-memory", "", "the memory resource to use for the evaluator, like 1Gi.")

	command.Flags().IntVar(&c.args.ChiefPort, "chief-port", 0, "the port of the chief.")
	command.Flags().StringSliceVar(&workerSelectors, "worker-selector", []string{}, `assigning jobs with "Worker" role to some k8s particular nodes(this option would cover --selector), usage: "--worker-selector=key=value"`)
	command.Flags().StringSliceVar(&chiefSelectors, "chief-selector", []string{}, `assigning jobs with "Chief" role to some k8s particular nodes(this option would cover --selector), usage: "--chief-selector=key=value"`)
	command.Flags().StringSliceVar(&evaluatorSelectors, "evaluator-selector", []string{}, `assigning jobs with "Evaluator" role to some k8s particular nodes(this option would cover --selector), usage: "--evaluator-selector=key=value"`)
	command.Flags().StringSliceVar(&psSelectors, "ps-selector", []string{}, `assigning jobs with "PS" role to some k8s particular nodes(this option would cover --selector), usage: "--ps-selector=key=value"`)
	command.Flags().StringVar(&roleSequence, "role-sequence", "", `specify the tfjob role sequence,like: "Worker,PS,Chief,Evaluator" or "w,p,c,e"`)

	c.AddArgValue("worker-selector", &workerSelectors).
		AddArgValue("chief-selector", &chiefSelectors).
		AddArgValue("evaluator-selector", &evaluatorSelectors).
		AddArgValue("ps-selector", &psSelectors).
		AddArgValue("role-sequence", &roleSequence)
}

func (c *CronTFJobArgsBuilder) PreBuild() error {
	for name := range c.subBuilders {
		if err := c.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	c.AddArgValue(ShareDataPrefix+"dataset", c.args.DataSet)
	return nil
}

func (c *CronTFJobArgsBuilder) Build() error {
	for name := range c.subBuilders {
		if err := c.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := c.setStandaloneMode(); err != nil {
		return err
	}
	if err := c.transform(); err != nil {
		return err
	}
	if err := c.setTFNodeSelectors(); err != nil {
		return err
	}
	if err := c.checkGangCapablitiesInCluster(); err != nil {
		return err
	}
	if err := c.setRuntime(); err != nil {
		return err
	}
	if err := c.checkRoleSequence(); err != nil {
		return err
	}
	if err := c.addRequestGPUsToAnnotation(); err != nil {
		return err
	}
	if err := c.check(); err != nil {
		return err
	}
	return nil
}

func (c *CronTFJobArgsBuilder) setStandaloneMode() error {
	if c.args.PSCount < 1 && c.args.WorkerCount == 1 {
		c.args.UseChief = true
		c.args.WorkerCount = 0
	}
	return nil
}

func (c *CronTFJobArgsBuilder) transform() error {
	arenaConfiger := config.GetArenaConfiger()
	if c.args.WorkerImage == "" {
		c.args.WorkerImage = c.args.Image
	}

	if c.args.WorkerCount > 0 {
		autoSelectWorkerPort, err := util.SelectAvailablePortWithDefault(arenaConfiger.GetClientSet(), c.args.WorkerPort)
		if err != nil {
			return fmt.Errorf("failed to select worker port: %++v", err)
		}
		c.args.WorkerPort = autoSelectWorkerPort
	}

	if c.args.UseChief {
		autoSelectChiefPort, err := util.SelectAvailablePortWithDefault(arenaConfiger.GetClientSet(), c.args.ChiefPort)
		if err != nil {
			return fmt.Errorf("failed to select chief port: %++v", err)
		}
		c.args.ChiefPort = autoSelectChiefPort
		c.args.ChiefCount = 1
	}

	if c.args.PSCount > 0 {
		autoSelectPsPort, err := util.SelectAvailablePortWithDefault(arenaConfiger.GetClientSet(), c.args.PSPort)
		if err != nil {
			return fmt.Errorf("failed to select ps port: %++v", err)
		}
		c.args.PSPort = autoSelectPsPort
		if c.args.PSImage == "" {
			c.args.PSImage = c.args.Image
		}
	}

	if c.args.UseEvaluator {
		c.args.EvaluatorCount = 1
	}
	return nil
}

func (c *CronTFJobArgsBuilder) checkGangCapablitiesInCluster() error {
	c.args.HasGangScheduler = false
	arenaConfiger := config.GetArenaConfiger()
	podList, err := arenaConfiger.GetClientSet().CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%v", gangSchdName),
	})
	if err != nil {
		log.Debugf("Failed to find %s due to %v", gangSchdName, err)
		return nil
	}
	if len(podList.Items) == 0 {
		log.Debugf("not found %v scheduler,it represents that you don't deploy it", gangSchdName)
		return nil
	}
	log.Debugf("Found %s successfully, the gang scheduler is enabled in the cluster.", gangSchdName)
	c.args.HasGangScheduler = true
	return nil
}

// add node selectors
func (c *CronTFJobArgsBuilder) setTFNodeSelectors() error {
	c.args.TFNodeSelectors = map[string]map[string]string{}
	var (
		psSelectors        *[]string
		workerSelectors    *[]string
		chiefSelectors     *[]string
		evaluatorSelectors *[]string
	)
	item1, ok := c.argValues["ps-selector"]
	if ok {
		psSelectors = item1.(*[]string)
	}
	item2, ok := c.argValues["worker-selector"]
	if ok {
		workerSelectors = item2.(*[]string)
	}
	item3, ok := c.argValues["chief-selector"]
	if ok {
		chiefSelectors = item3.(*[]string)
	}
	item4, ok := c.argValues["chief-selector"]
	if ok {
		evaluatorSelectors = item4.(*[]string)
	}
	for _, role := range []string{"PS", "Worker", "Evaluator", "Chief"} {
		switch {
		case role == "PS":
			c.transformSelectorArrayToMap(psSelectors, "PS")
		case role == "Worker":
			c.transformSelectorArrayToMap(workerSelectors, "Worker")
		case role == "Chief":
			c.transformSelectorArrayToMap(chiefSelectors, "Chief")
		case role == "Evaluator":
			c.transformSelectorArrayToMap(evaluatorSelectors, "Evaluator")
		}
	}
	return nil
}

func (c *CronTFJobArgsBuilder) transformSelectorArrayToMap(selectorArray *[]string, role string) {
	c.args.TFNodeSelectors[role] = map[string]string{}
	if selectorArray != nil && len(*selectorArray) != 0 {
		log.Debugf("%v Selectors: %v", role, selectorArray)
		c.args.TFNodeSelectors[role] = transformSliceToMap(*selectorArray, "=")
		return
	}
	// set the default node selectors to tf role node selectors
	log.Debugf("use to Node Selectors %v to %v Selector", c.args.NodeSelectors, role)
	c.args.TFNodeSelectors[role] = c.args.NodeSelectors

}

func (c *CronTFJobArgsBuilder) addRequestGPUsToAnnotation() error {
	gpus := 0
	gpus += c.args.ChiefCount
	gpus += c.args.EvaluatorCount
	gpus += c.args.WorkerCount * c.args.GPUCount
	if c.args.Annotations == nil {
		c.args.Annotations = map[string]string{}
	}
	c.args.Annotations[types.RequestGPUsOfJobAnnoKey] = fmt.Sprintf("%v", gpus)
	return nil
}

func (c *CronTFJobArgsBuilder) checkRoleSequence() error {
	roleSequence := ""
	var getRoleSeqFromConfig = func() {
		configs := config.GetArenaConfiger().GetConfigsFromConfigFile()
		value, ok := configs["tfjob_role_sequence"]
		if !ok {
			return
		}
		log.Debugf("read tfjob role sequence from config file")
		roleSequence = strings.Trim(value, " ")
	}
	item, ok := c.argValues["role-sequence"]
	if !ok {
		getRoleSeqFromConfig()
	} else {
		v := item.(*string)
		roleSequence = *v
		if roleSequence == "" {
			getRoleSeqFromConfig()
		} else {
			log.Debugf("read tfjob role sequence from command option")
		}
	}
	if roleSequence == "" {
		return nil
	}
	roles := []string{}
	for _, r := range strings.Split(roleSequence, ",") {
		switch strings.ToLower(strings.Trim(r, " ")) {
		case "worker", "w":
			roles = append(roles, "Worker")
		case "ps", "p":
			roles = append(roles, "PS")
		case "evaluator", "e":
			roles = append(roles, "Evaluator")
		case "chief", "c":
			roles = append(roles, "Chief")
		default:
			return fmt.Errorf("Unknown role: %v, the tfjob only supports:[Worker(w)|PS(p)|Evaluator(e)|Chief(c)]", r)
		}
	}
	if c.args.Annotations == nil {
		c.args.Annotations = map[string]string{}
	}
	c.args.Annotations["job-role-sequence"] = strings.Join(roles, ",")
	return nil
}

func (c *CronTFJobArgsBuilder) setRuntime() error {
	// Get the runtime name
	annotations := c.args.CommonSubmitArgs.Annotations
	name := annotations["runtime"]
	c.args.TFRuntime = runtime.GetTFRuntime(name)
	return c.args.TFRuntime.Check(&c.args.SubmitTFJobArgs)
}

func (c *CronTFJobArgsBuilder) check() error {
	switch c.args.CleanPodPolicy {
	case "None", "Running":
		log.Debugf("Supported cleanTaskPolicy: %s", c.args.CleanPodPolicy)
	default:
		return fmt.Errorf("Unsupported cleanTaskPolicy %s", c.args.CleanPodPolicy)
	}
	if c.args.WorkerCount == 0 && !c.args.UseChief {
		return fmt.Errorf("--workers must be greater than 0 in distributed training")
	}
	if c.args.WorkerImage == "" {
		return fmt.Errorf("--image or --workerImage must be set")
	}
	if c.args.PSCount > 0 {
		if c.args.PSImage == "" {
			return fmt.Errorf("--image or --psImage must be set")
		}
	}
	return nil
}

package argsbuilder

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
	"reflect"
	"strings"
)

type CronTFJobArgsBuilder struct {
	args        *types.CronTFJobArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewCronTFJobArgsBuilder(args *types.CronTFJobArgs) ArgsBuilder {
	args.TrainingType = types.TFTrainingJob
	c := &CronTFJobArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
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
	command.Flags().IntVar(&c.args.JobsHistoryLimit, "jobs-history-limit", 0, "the number of finished job history to retain")

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
	/*
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
	*/
	return nil
}

package commands

import (
	"github.com/kubeflow/arena/pkg/util"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"github.com/kubeflow/arena/pkg/workflow"
	"os"
	"fmt"
	"errors"
)

var (
	sparkChart = util.GetChartsFolder() + "/sparkjob"
)

const (
	defaultSparkJobTrainingType = "sparkjob"
)

/**
	https://github.com/GoogleCloudPlatform/spark-on-k8s-operator

	sparkApplication is the supported as default
	scheduledSparkApplication is not supported.
 */
func NewSparkApplicationCommand() *cobra.Command {
	submitArgs := NewSubmitSparkJobArgs()
	var command = &cobra.Command{
		Use:     "sparkjob",
		Short:   "Submit a common spark application job.",
		Aliases: []string{"spark"},
		Run: func(cmd *cobra.Command, args []string) {

			util.SetLogLevel(logLevel)
			setupKubeconfig()

			_, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = updateNamespace(cmd)
			if err != nil {
				log.Debugf("Failed due to %v", err)
				fmt.Println(err)
				os.Exit(1)
			}

			err = submitSparkApplicationJob(args, submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	submitArgs.addFlags(command)

	return command
}

func NewSubmitSparkJobArgs() *submitSparkJobArgs {
	return &submitSparkJobArgs{
		Driver:   &Driver{},
		Executor: &Executor{},
	}
}

type submitSparkJobArgs struct {
	Image     string    `yaml:"Image"`
	MainClass string    `yaml:"MainClass"`
	Jar       string    `yaml:"Jar"`
	Executor  *Executor `yaml:"Executor"`
	Driver    *Driver   `yaml:"Driver"`
}

type Driver struct {
	CPURequest    int    `yaml:"CPURequest"`
	MemoryRequest string `yaml:"MemoryRequest"`
}

type Executor struct {
	Replicas      int    `yaml:"Replicas"`
	CPURequest    int    `yaml:"CPURequest"`
	MemoryRequest string `yaml:"MemoryRequest"`
}

// add flags to submit spark args
func (sa *submitSparkJobArgs) addFlags(command *cobra.Command) {
	command.Flags().StringVar(&name, "name", "", "override name")
	command.MarkFlagRequired("name")

	command.Flags().StringVar(&sa.Image, "image", "registry.aliyuncs.com/acs/spark:v2.4.0", "the docker image name of training job")
	command.Flags().IntVar(&sa.Executor.Replicas, "replicas", 1, "the executor's number to run the distributed training.")
	command.Flags().StringVar(&sa.MainClass, "main-class", "org.apache.spark.examples.SparkPi", "main class of your jar")
	command.Flags().StringVar(&sa.Jar, "jar", "local:///opt/spark/examples/jars/spark-examples_2.11-2.4.0.jar", "jar path in image")

	// cpu and memory request
	command.Flags().IntVar(&sa.Driver.CPURequest, "driver-cpu-request", 1, "cpu request for driver pod")
	command.Flags().StringVar(&sa.Driver.MemoryRequest, "driver-memory-request", "500m", "memory request for driver pod (min is 500m)")
	command.Flags().IntVar(&sa.Executor.CPURequest, "executor-cpu-request", 1, "cpu request for executor pod")
	command.Flags().StringVar(&sa.Executor.MemoryRequest, "executor-memory-request", "500m", "memory request for executor pod (min is 500m)")
}

// TODO add more check
// check params
func (sa *submitSparkJobArgs) isValid() error {
	if sa.Executor != nil && sa.Executor.Replicas == 0 {
		return errors.New("WorkersMustMoreThanOne")
	}

	return nil
}

func submitSparkApplicationJob(args []string, submitArgs *submitSparkJobArgs) error {
	err := submitArgs.isValid()
	if err != nil {
		return err
	}

	trainer := NewSparkJobTrainer(clientset)

	job, err := trainer.GetTrainingJob(name, namespace)
	if err != nil {
		return fmt.Errorf("failed to create sparkjob %s due to error %v", name, err)
	}

	if job != nil {
		return fmt.Errorf("the job %s already exist, please delete it first. use 'arena delete %s'", name, name)
	}

	err = workflow.SubmitJob(name, defaultSparkJobTrainingType, namespace, submitArgs, sparkChart)
	if err != nil {
		return err
	}

	log.Infof("The Job %s has been submitted successfully", name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", name, defaultSparkJobTrainingType)
	return nil
}

package commands

import (
	"errors"
	"fmt"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var (
	volcanoChart = util.GetChartsFolder() + "/volcanojob"
)

const (
	defaultVolcanoJobTrainingType = "volcanojob"
)

func NewVolcanoJobCommand() *cobra.Command {
	submitArgs := NewSubmitVolcanoJobArgs()
	var command = &cobra.Command{
		Use:     "volcanojob",
		Short:   "Submit a Volcano job.",
		Aliases: []string{"vj"},
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

			err = submitVolcanoJob(args, submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	submitArgs.addFlags(command)

	return command
}

func NewSubmitVolcanoJobArgs() *submitVolcanoJobArgs {
	return &submitVolcanoJobArgs{}
}

type submitVolcanoJobArgs struct {
	// The MinAvailable available pods to run for this Job
	MinAvailable int `yaml:"minAvailable"`
	// Specifies the queue that will be used in the scheduler, "default" queue is used this leaves empty.
	Queue string `yaml:"queue"`
	// SchedulerName is the default value of `tasks.template.spec.schedulerName`.
	SchedulerName string `yaml:"schedulerName"`
	// TaskName specifies the name of task
	TaskName   string   `yaml:"taskName"`
	TaskImages []string `yaml:"taskImages"`
	// TaskReplicas specifies the replicas of this Task in Job
	TaskReplicas int `yaml:"taskReplicas"`
	// TaskCPU specifies the cpu resource required for each replica of Task in Job. default is 250m
	TaskCPU string `yaml:"taskCPU"`
	// TaskMemory specifies the memory resource required for each replica of Task in Job. default is 128Mi
	TaskMemory string `yaml:"taskMemory"`
	TaskPort   int    `yaml:"taskPort"`
}

// add flags to submit spark args
func (sa *submitVolcanoJobArgs) addFlags(command *cobra.Command) {
	command.Flags().StringVar(&name, "name", "", "override name")
	command.MarkFlagRequired("name")

	command.Flags().IntVar(&(sa.MinAvailable), "minAvailable", 1, "The minimal available pods to run for this Job. default value is 1")
	command.Flags().StringVar(&(sa.Queue), "queue", "default", "Specifies the queue that will be used in the scheduler, default queue is used this leaves empty")
	command.Flags().StringVar(&(sa.SchedulerName), "schedulerName", "volcano", "Specifies the scheduler Name, default is volcano when not specified")
	// each task related information name,image,replica number
	command.Flags().StringVar(&(sa.TaskName), "taskName", "task", "the task name of volcano job, default value is task")
	command.Flags().StringSliceVar(&(sa.TaskImages), "taskImages", []string{"ubuntu", "nginx", "busybox"}, "the docker images of different tasks of volcano job. default used 3 tasks with ubuntu,nginx and busybox images")
	command.Flags().IntVar(&(sa.TaskReplicas), "taskReplicas", 1, "the task replica's number to run the distributed tasks. default value is 1")
	// cpu and memory request
	command.Flags().StringVar(&(sa.TaskCPU), "taskCPU", "250m", "cpu request for each task replica / pod. default value is 250m")
	command.Flags().StringVar(&(sa.TaskMemory), "taskMemory", "128Mi", "memory request for each task replica/pod.default value is 128Mi)")
	command.Flags().IntVar(&(sa.TaskPort), "taskPort", 2222, "the task port number. default value is 2222")

}

// check params
func (sa *submitVolcanoJobArgs) isValid() error {

	if len(sa.TaskName) == 0 {
		return errors.New("Default task Name should be there")
	}

	if len(sa.TaskImages) == 0 {
		return errors.New("TaskImages should be there")
	}

	return nil
}

func submitVolcanoJob(args []string, submitArgs *submitVolcanoJobArgs) error {
	err := submitArgs.isValid()
	if err != nil {
		return err
	}

	trainer := NewVolcanoJobTrainerSubmit(clientset)

	job, err := trainer.GetTrainingJobAtSubmit(name, namespace)
	if err != nil {
		return fmt.Errorf("failed to create volcano %s due to error %v", name, err)
	}

	if job != nil {
		return fmt.Errorf("the job %s already exist, please delete it first. use 'arena delete %s'", name, name)
	}

	err = workflow.SubmitJob(name, defaultVolcanoJobTrainingType, namespace, submitArgs, volcanoChart)
	if err != nil {
		return err
	}

	log.Infof("The Job %s has been submitted successfully", name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", name, defaultVolcanoJobTrainingType)
	return nil
}

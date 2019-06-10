package commands

import (
	"errors"
	"fmt"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strconv"
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
	return &submitVolcanoJobArgs{
		Tasks: make([]Task, 3, 10),
	}
}

type submitVolcanoJobArgs struct {
	MinAvailable  int    `yaml:"minAvailable"`
	Queue         string `yaml:"queue"`
	SchedulerName string `yaml:"schedulerName"`
	Tasks         []Task `yaml:"tasks"`
}

type Task struct {
	TaskName     string `yaml:"taskName"`
	TaskImage    string `yaml:"taskImage"`
	TaskReplicas int    `yaml:"taskReplicas"`
	TaskCPU      string `yaml:"taskCPU"`
	TaskMemory   string `yaml:"taskMemory"`
	TaskPort     int    `yaml:"taskPort"`
}

// add flags to submit spark args
func (sa *submitVolcanoJobArgs) addFlags(command *cobra.Command) {
	command.Flags().StringVar(&name, "name", "", "override name")
	command.MarkFlagRequired("name")

	command.Flags().IntVar(&(sa.MinAvailable), "minAvailable", 1, "The minimal available pods to run for this Job.")
	command.Flags().StringVar(&(sa.Queue), "queue", "default", "Specifies the queue that will be used in the scheduler, default queue is used this leaves empty")
	command.Flags().StringVar(&(sa.SchedulerName), "schedulerName", "kube-batch", "Specifies the scheduler Name, default  is kube-batch used this leaves empty")

	for i := 0; i < 3; i++ {

		command.Flags().StringVar(&(sa.Tasks[i].TaskName), "taskName"+strconv.Itoa(i), "task"+strconv.Itoa(i), "the task name of volcano job")
		command.Flags().StringVar(&(sa.Tasks[i].TaskImage), "taskImage"+strconv.Itoa(i), "nginx", "the docker image name of task job")
		command.Flags().IntVar(&(sa.Tasks[i].TaskReplicas), "taskReplicas"+strconv.Itoa(i), 1, "the task replica's number to run the distributed task.")
		// cpu and memory request
		command.Flags().StringVar(&(sa.Tasks[i].TaskCPU), "taskCPU"+strconv.Itoa(i), "250m", "cpu request for task pod")
		command.Flags().StringVar(&(sa.Tasks[i].TaskMemory), "taskMemory"+strconv.Itoa(i), "128Mi", "memory request for task pod (min is 128Mi)")
		command.Flags().IntVar(&(sa.Tasks[i].TaskPort), "taskPort"+strconv.Itoa(i), 2222+i, "the task port number.")
	}

}

// TODO add more check
// check params
func (sa *submitVolcanoJobArgs) isValid() error {

	if len(sa.Tasks) == 0 {
		return errors.New("tasks should be there")
	}

	for i := 0; i < len(sa.Tasks); i++ {
		if sa.Tasks[i].TaskImage == "" {
			return errors.New("Image should be there in each task")
		}
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

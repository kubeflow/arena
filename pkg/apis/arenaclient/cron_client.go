package arenaclient

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	apiscron "github.com/kubeflow/arena/pkg/apis/cron"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/cron"
)

type CronTaskClient struct {
	namespace string
	configer  *config.ArenaConfiger
}

// NewCronTaskClient creates a CronTaskClient
func NewCronTaskClient(namespace string, configer *config.ArenaConfiger) *CronTaskClient {
	return &CronTaskClient{
		namespace: namespace,
		configer:  configer,
	}
}

// Submit submits a training job
func (c *CronTaskClient) SubmitCronTrainingJob(job *apiscron.Job) error {
	switch job.Type() {
	case types.CronTFTrainingJob:
		args := job.Args().(*types.CronTFJobArgs)
		return cron.SubmitCronTFJob(c.namespace, args)
	}
	return nil
}

// Namespace sets the namespace,this operation does not change the default namespace
func (c *CronTaskClient) Namespace(namespace string) *CronTaskClient {
	copyCronTaskClient := &CronTaskClient{
		namespace: namespace,
		configer:  c.configer,
	}
	return copyCronTaskClient
}

// List return all cron task
func (c *CronTaskClient) List(allNamespaces bool) ([]*types.CronTaskInfo, error) {
	return cron.ListCronTask(c.namespace, allNamespaces)
}

// ListAndPrint lists and prints the job informations
func (c *CronTaskClient) ListAndPrint(allNamespaces bool, format string) error {
	if utils.TransferPrintFormat(format) == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}
	tasks, err := cron.ListCronTask(c.namespace, allNamespaces)
	if err != nil {
		return err
	}
	cron.DisplayAllCronTasks(tasks, allNamespaces, utils.TransferPrintFormat(format))
	return nil
}

func (c *CronTaskClient) Get(name string) (*types.CronTaskInfo, error) {
	return cron.GetCronTask(name, c.namespace)
}

func (c *CronTaskClient) GetAndPrint(name string) error {
	fmt.Println(name)
	info, err := cron.GetCronTask(name, c.namespace)
	if err != nil {
		return err
	}

	b, _ := json.Marshal(info)
	fmt.Println(string(b))
	return nil
}

func (c *CronTaskClient) Delete(names ...string) error {
	fmt.Println("====== delete cron ")
	return nil
}

package arenaclient

import (
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/kubeflow/arena/pkg/apis/config"
	apievaluate "github.com/kubeflow/arena/pkg/apis/evaluate"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/evaluate"
)

type EvaluateClient struct {
	namespace string
	configer  *config.ArenaConfiger
}

// NewEvaluateClient creates a EvaluateClient
func NewEvaluateClient(namespace string, configer *config.ArenaConfiger) *EvaluateClient {
	return &EvaluateClient{
		namespace: namespace,
		configer:  configer,
	}
}

// Namespace sets the namespace,this operation does not change the default namespace
func (c *EvaluateClient) Namespace(namespace string) *EvaluateClient {
	copyEvaluateClient := &EvaluateClient{
		namespace: namespace,
		configer:  c.configer,
	}
	return copyEvaluateClient
}

// SubmitEvaluateJob submits a evaluate job
func (c *EvaluateClient) SubmitEvaluateJob(job *apievaluate.EvaluateJob) error {
	args := job.Args().(*types.EvaluateJobArgs)

	// generate uuid v4
	u4 := uuid.New()
	jobId := u4.String()

	if args.Envs == nil {
		args.Envs = make(map[string]string)
	}
	args.Envs["JOB_ID"] = jobId
	args.Envs["MODEL_NAME"] = args.ModelName
	args.Envs["MODEL_PATH"] = args.ModelPath
	args.Envs["DATASET_PATH"] = args.DatasetPath
	args.Envs["METRICS_PATH"] = args.MetricsPath
	if args.ModelVersion != "" {
		args.Envs["MODEL_VERSION"] = args.ModelVersion
	}

	if args.Labels == nil {
		args.Labels = make(map[string]string)
	}
	args.Labels["jobId"] = jobId

	return evaluate.SubmitEvaluateJob(c.namespace, args)
}

func (c *EvaluateClient) Get(name, namespace string) (*types.EvaluateJobInfo, error) {
	return evaluate.GetEvaluateJob(name, namespace)
}

func (c *EvaluateClient) GetAndPrint(name string, format string) error {
	outputFormat := utils.TransferPrintFormat(format)
	if outputFormat == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}

	job, err := evaluate.GetEvaluateJob(name, c.namespace)
	if err != nil {
		return err
	}
	evaluate.DisplayEvaluateJob(job, outputFormat)
	return nil
}

func (c *EvaluateClient) List(allNamespaces bool) ([]*types.EvaluateJobInfo, error) {
	return evaluate.ListEvaluateJobs(c.namespace, allNamespaces)
}

func (c *EvaluateClient) ListAndPrint(allNamespaces bool, format string) error {
	outputFormat := utils.TransferPrintFormat(format)
	if outputFormat == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}

	jobs, err := evaluate.ListEvaluateJobs(c.namespace, allNamespaces)
	if err != nil {
		return err
	}
	evaluate.DisplayAllEvaluateJobs(jobs, allNamespaces, outputFormat)
	return nil
}

func (c *EvaluateClient) Delete(names ...string) error {
	for _, name := range names {
		err := evaluate.DeleteEvaluateJob(name, c.namespace)
		if err != nil {
			log.Errorf("failed to delete evaluate job, name:%s ns:%s, reason:%v", name, c.namespace, err)
		}
	}

	return nil
}

package arenaclient

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/util"
)

// ArenaClient is a client which includes operations:
// 1.manage training jobs,like:
//   - submit a training job
//   - get a training job information
//   - get job logs
//   - delete a job
//
// TODO: 2.manage serving job
// TODO: 3.manage node
// it serves for commands and apis
type ArenaClient struct {
	namespace            string
	arenaSystemNamespace string
	arenaConfiger        *config.ArenaConfiger
}

// NewArenaClient creates a ArenaClient
func NewArenaClient(args types.ArenaClientArgs) (*ArenaClient, error) {
	// check loglevel is valid or not
	if utils.TransferLogLevel(args.LogLevel) == types.LogUnknown {
		return nil, fmt.Errorf("unknown loglevel,only support:[info,debug,error,warn]")
	}
	// set log level
	util.SetLogLevel(args.LogLevel)
	// if arenaSystemNamespace is null,transfer it to "arena-system"
	if args.ArenaNamespace == "" {
		args.ArenaNamespace = "arena-system"
	}
	client := &ArenaClient{
		arenaSystemNamespace: args.ArenaNamespace,
	}
	// InitArenaConfiger creates and init ArenaConfiger
	// Warning: this function only init one time
	configer, err := config.InitArenaConfiger(args)
	if err != nil {
		return nil, err
	}
	if err := k8saccesser.InitK8sResourceAccesser(configer.GetRestConfig(), configer.GetClientSet(), configer.IsDaemonMode()); err != nil {
		return client, err
	}
	client.arenaConfiger = configer
	// the namespace may be updated
	client.namespace = configer.GetNamespace()
	return client, err
}

// Training returns the Training Job Client
func (a *ArenaClient) Training() *TrainingJobClient {
	return NewTrainingJobClient(a.namespace, a.arenaSystemNamespace, a.arenaConfiger)
}

// Serving returns the Serving job client
func (a *ArenaClient) Serving() *ServingJobClient {
	return NewServingJobClient(a.namespace, a.arenaConfiger)
}

// Serving returns the Cron client
func (a *ArenaClient) Cron() *CronClient {
	return NewCronClient(a.namespace, a.arenaConfiger)
}

func (a *ArenaClient) Node() *NodeClient {
	return NewNodeClient(a.namespace, a.arenaConfiger)
}

func (a *ArenaClient) Data() *DataClient {
	return NewDataClient(a.namespace, a.arenaConfiger)
}

func (a *ArenaClient) Evaluate() *EvaluateClient {
	return NewEvaluateClient(a.namespace, a.arenaConfiger)
}

func (a *ArenaClient) Analyze() *AnalyzeClient {
	return NewAnalyzeClient(a.namespace, a.arenaConfiger)
}

package arenaclient

import (
	"fmt"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/util"
)

// ArenaClient is a client which includes operations:
// 1.manage training jobs,like:
//   * submit a training job
//   * get a training job information
//   * get job logs
//   * delete a job
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
	// if namespace is null,transfer it to "default"
	if args.Namespace == "" {
		args.Namespace = "default"
	}
	// if arenaSystemNamespace is null,transfer it to "arena-system"
	if args.ArenaNamespace == "" {
		args.ArenaNamespace = "arena-system"
	}
	client := &ArenaClient{
		namespace:            args.Namespace,
		arenaSystemNamespace: args.ArenaNamespace,
	}
	// InitArenaConfiger creates and init ArenaConfiger
	// Warning: this function only init one time
	configer, err := config.InitArenaConfiger(args)
	if err != nil {
		return nil, err
	}
	client.arenaConfiger = configer
	return client, err
}

// Training returns the Training Job Client
func (a *ArenaClient) Training() *TrainingJobClient {
	return NewTrainingJobClient(a.namespace, a.arenaSystemNamespace, a.arenaConfiger)
}

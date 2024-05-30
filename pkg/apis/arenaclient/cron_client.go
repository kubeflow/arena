// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package arenaclient

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	apiscron "github.com/kubeflow/arena/pkg/apis/cron"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	"github.com/kubeflow/arena/pkg/cron"
	log "github.com/sirupsen/logrus"
)

type CronClient struct {
	namespace string
	configer  *config.ArenaConfiger
}

// NewCronClient creates a CronClient
func NewCronClient(namespace string, configer *config.ArenaConfiger) *CronClient {
	return &CronClient{
		namespace: namespace,
		configer:  configer,
	}
}

// Submit submits a training job
func (c *CronClient) SubmitCronTrainingJob(job *apiscron.Job) error {
	switch job.Type() {
	case types.CronTFTrainingJob:
		args := job.Args().(*types.CronTFJobArgs)
		return cron.SubmitCronTFJob(c.namespace, args)
	}
	return nil
}

// Namespace sets the namespace,this operation does not change the default namespace
func (c *CronClient) Namespace(namespace string) *CronClient {
	copyCronClient := &CronClient{
		namespace: namespace,
		configer:  c.configer,
	}
	return copyCronClient
}

// List return all cron task
func (c *CronClient) List(allNamespaces bool) ([]*types.CronInfo, error) {
	return cron.ListCrons(c.namespace, allNamespaces)
}

// ListAndPrint lists and prints the job informations
func (c *CronClient) ListAndPrint(allNamespaces bool, format string) error {
	outputFormat := utils.TransferPrintFormat(format)
	if outputFormat == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}
	cronInfos, err := cron.ListCrons(c.namespace, allNamespaces)
	if err != nil {
		return err
	}
	cron.DisplayAllCrons(cronInfos, allNamespaces, outputFormat)
	return nil
}

func (c *CronClient) Get(name string) (*types.CronInfo, error) {
	return cron.GetCronInfo(name, c.namespace)
}

func (c *CronClient) GetAndPrint(name string, format string) error {
	outputFormat := utils.TransferPrintFormat(format)
	if outputFormat == types.UnknownFormat {
		return fmt.Errorf("Unknown output format,only support:[wide|json|yaml]")
	}

	cronInfo, err := cron.GetCronInfo(name, c.namespace)
	if err != nil {
		return err
	}

	cron.DisplayCron(cronInfo, outputFormat)
	return nil
}

func (c *CronClient) Suspend(name string) error {
	return cron.SuspendCron(name, c.namespace, true)
}

func (c *CronClient) Resume(name string) error {
	return cron.SuspendCron(name, c.namespace, false)
}

func (c *CronClient) Delete(names ...string) error {
	for _, name := range names {
		cronInfo, err := cron.GetCronInfo(name, c.namespace)
		if err != nil {
			log.Errorf("failed to get cron info of %s, reason: %v", name, err)
			continue
		}

		_ = cron.DeleteCron(name, c.namespace, cronInfo.Type)
	}

	return nil
}

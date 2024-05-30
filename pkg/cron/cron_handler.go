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

package cron

import (
	"context"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/operators/kubedl-operator/apis/apps/v1alpha1"
	"github.com/kubeflow/arena/pkg/operators/kubedl-operator/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var cronHandler *CronHandler

// CronHandler cron handler
type CronHandler struct {
	// the k8s client
	client *kubernetes.Clientset
	// cronClient client
	cronClient *versioned.Clientset
	// check if it's enabled
	enabled bool
}

func NewCronHandler() *CronHandler {
	arenaConfiger := config.GetArenaConfiger()
	cronClient := versioned.NewForConfigOrDie(arenaConfiger.GetRestConfig())
	enable := false
	_, err := arenaConfiger.GetAPIExtensionClientSet().ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), k8saccesser.TensorflowCRDName, metav1.GetOptions{})
	if err == nil {
		log.Debugf("CronHandler is enabled")
		enable = true
	} else {
		log.Debugf("CronHandler is disabled,reason: %v", err)
	}
	log.Debugf("Succeed to init CronHandler")
	return &CronHandler{
		client:     arenaConfiger.GetClientSet(),
		cronClient: cronClient,
		enabled:    enable,
	}
}

func GetCronHandler() *CronHandler {
	if cronHandler == nil {
		cronHandler = NewCronHandler()
	}
	return cronHandler
}

func (ch *CronHandler) ListCrons(namespace string, allNamespaces bool) ([]*types.CronInfo, error) {
	if allNamespaces {
		namespace = metav1.NamespaceAll
	}

	crons, err := k8saccesser.GetK8sResourceAccesser().ListCrons(ch.cronClient, namespace)
	if err != nil {
		return nil, err
	}

	var cronInfos []*types.CronInfo
	for _, cron := range crons {
		cronInfo := ch.buildCronInfo(cron)
		cronInfos = append(cronInfos, cronInfo)
	}

	return cronInfos, nil
}

func (ch *CronHandler) GetCron(namespace string, name string) (*types.CronInfo, error) {
	cron, err := k8saccesser.GetK8sResourceAccesser().GetCron(ch.cronClient, namespace, name)
	if err != nil {
		return nil, err
	}

	cronInfo := ch.buildCronInfo(cron)
	return cronInfo, nil
}

func (ch *CronHandler) DeleteCron(namespace string, name string) error {
	return ch.cronClient.AppsV1alpha1().Crons(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
}

func (ch *CronHandler) UpdateCron(namespace string, name string, suspend bool) error {
	cron, err := ch.cronClient.AppsV1alpha1().Crons(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	cron.Spec.Suspend = &suspend
	_, err = ch.cronClient.AppsV1alpha1().Crons(namespace).Update(context.TODO(), cron, metav1.UpdateOptions{})
	return err
}

func (ch *CronHandler) buildCronInfo(cron *v1alpha1.Cron) *types.CronInfo {
	cronInfo := &types.CronInfo{
		UUID:              string(cron.UID),
		Name:              cron.Name,
		Namespace:         cron.Namespace,
		Type:              cron.Spec.CronTemplate.Kind,
		Schedule:          cron.Spec.Schedule,
		ConcurrencyPolicy: string(cron.Spec.ConcurrencyPolicy),
		HistoryLimit:      int64(*cron.Spec.HistoryLimit),
		CreationTimestamp: formatTime(cron.CreationTimestamp.Time),
	}

	suspend := false
	if cron.Spec.Suspend != nil {
		suspend = *cron.Spec.Suspend
	}
	cronInfo.Suspend = suspend

	if cron.Spec.Deadline != nil {
		cronInfo.Deadline = formatTime(cron.Spec.Deadline.Time)
	}

	if cron.Status.LastScheduleTime != nil {
		cronInfo.LastScheduleTime = formatTime(cron.Status.LastScheduleTime.Time)
	}

	if len(cron.Status.History) > 0 {
		var histories []types.CronHistoryInfo
		for _, item := range cron.Status.History {
			history := types.CronHistoryInfo{
				Namespace: cron.Namespace,
				Name:      item.Object.Name,
				Group:     *item.Object.APIGroup,
				Kind:      item.Object.Kind,
				Status:    string(item.Status),
			}

			if item.Created != nil {
				history.CreateTime = formatTime(item.Created.Time)
			}

			if item.Finished != nil {
				history.FinishTime = formatTime(item.Finished.Time)
			}

			histories = append(histories, history)
		}

		cronInfo.History = histories
	}

	return cronInfo
}

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

package podlogs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	ErrPodNotFound      = errors.New(`no logs return,because not found instance`)
	ErrTooManyPodsFound = errors.New(`too many pods found in the job,`)
	ErrPodNotRunning    = errors.New(`is not running`)
	ErrPodStatusUnknown = errors.New(`status is unknown`)
)

type PodLogger struct {
	clientset kubernetes.Interface
	*types.LogArgs
	pipe
}

type pipe struct {
	Reader io.ReadCloser
	Writer io.WriteCloser
}

func NewPodLogger(args *types.LogArgs) *PodLogger {
	piper, pipew := io.Pipe()
	pipe := pipe{Reader: piper, Writer: pipew}
	return &PodLogger{
		clientset: kubernetes.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig()),
		LogArgs:   args,
		pipe:      pipe,
	}
}

func (p *PodLogger) Print() (int, error) {
	return p.AcceptLogs()
}

func (p *PodLogger) AcceptLogs() (int, error) {
	defer p.Reader.Close()
	if err := p.getLogs(func(reader io.ReadCloser) {
		defer p.Writer.Close()
		defer reader.Close()
		if _, err := io.Copy(p.Writer, reader); err != nil {
			log.Debugf("get logs failed, err: %s", err)
		}
	}); err != nil {
		return 1, err
	}
	if _, err := io.Copy(p.WriterCloser, p.Reader); err != nil {
		log.Debugf("get logs failed, err: %s", err)
	}
	return 0, nil
}

func (p *PodLogger) getLogs(accept func(io.ReadCloser)) error {
	err := p.ensureContainerStarted()
	if err != nil {
		return err
	}
	podLogOption := &corev1.PodLogOptions{
		// Container:    p.container,
		Follow:       p.Follow,
		Timestamps:   p.Timestamps,
		SinceSeconds: p.SinceSeconds,
		SinceTime:    p.SinceTime,
		TailLines:    p.Tail,
	}
	if p.ContainerName != "" {
		podLogOption.Container = p.ContainerName
	}
	readCloser, err := p.clientset.CoreV1().Pods(p.Namespace).GetLogs(p.InstanceName, podLogOption).Stream(context.TODO())

	if err != nil {
		return err
	}
	// warning: readCloser should execute readCloser.Close() in accept function.
	go accept(readCloser)
	return nil
}

func (p *PodLogger) ensureContainerStarted() error {
	for p.RetryCnt > 0 {
		pod, err := p.clientset.CoreV1().Pods(p.Namespace).Get(context.TODO(), p.InstanceName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		status, _, _, _ := utils.DefinePodPhaseStatus(*pod)
		log.Debugf("pod:%s,pod phase: %v\n", p.InstanceName, pod.Status.Phase)
		log.Debugf("pod print status: %s\n", status)
		switch podPhase := pod.Status.Phase; {
		case podPhase == corev1.PodPending && strings.Index(status, "Init:") == 0:
			return nil
		case podPhase == corev1.PodPending && strings.Index(status, "PodInitializing") == 0:
			return nil
		case podPhase == corev1.PodRunning && status != "Unknown":
			return nil
		case podPhase == corev1.PodFailed || podPhase == corev1.PodSucceeded:
			return nil
		}
		p.RetryCnt--
	}
	return fmt.Errorf("instance %s %s", p.InstanceName, ErrPodNotRunning.Error())
}

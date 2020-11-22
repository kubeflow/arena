package podlogs

import (
	"errors"
	"fmt"
	"io"

	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	ErrPodNotFound      = errors.New(`no logs return,because not found instance`)
	ErrTooManyPodsFound = errors.New(`too many pods found in the job,`)
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
		io.Copy(p.Writer, reader)
	}); err != nil {
		return 1, err
	}
	io.Copy(p.WriterCloser, p.Reader)
	return 0, nil
}

func (p *PodLogger) getLogs(accept func(io.ReadCloser)) error {
	err := p.ensureContainerStarted()
	if err != nil {
		return err
	}
	readCloser, err := p.clientset.CoreV1().Pods(p.Namespace).GetLogs(p.InstanceName, &v1.PodLogOptions{
		// Container:    p.container,
		Follow:       p.Follow,
		Timestamps:   p.Timestamps,
		SinceSeconds: p.SinceSeconds,
		SinceTime:    p.SinceTime,
		TailLines:    p.Tail,
	}).Stream()

	if err != nil {
		return err
	}
	// warning: readCloser should execute readCloser.Close() in accept function.
	go accept(readCloser)
	return nil
}

func (p *PodLogger) ensureContainerStarted() error {
	for p.RetryCnt > 0 {
		pod, err := p.clientset.CoreV1().Pods(p.Namespace).Get(p.InstanceName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		status, _, _, _ := utils.DefinePodPhaseStatus(*pod)
		log.Debugf("pod:%s,pod phase: %v\n", p.InstanceName, pod.Status.Phase)
		log.Debugf("pod print status: %s\n", status)
		switch podPhase := pod.Status.Phase; {
		case podPhase == v1.PodRunning && status != "Unknown":
			return nil
		case podPhase == v1.PodFailed || podPhase == v1.PodSucceeded:
			return nil
		}
		p.RetryCnt--
	}
	return fmt.Errorf("instance %s %s", p.InstanceName, ErrPodNotRunning.Error())
}

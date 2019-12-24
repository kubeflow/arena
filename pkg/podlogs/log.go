package podlogs

import (
	"errors"
	"fmt"
	"io"
	"time"

	servejob "github.com/kubeflow/arena/pkg/jobs/serving"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	ErrPodNotRunning    = errors.New(`is not running`)
	ErrPodStatusUnknown = errors.New(`status is unknown`)
)

type PodLogArgs struct {
	Namespace    string
	PodName      string
	Follow       bool
	SinceSeconds *int64
	SinceTime    *metav1.Time
	Tail         *int64
	Timestamps   bool
	RetryCnt     int
	RetryTimeout time.Duration
	KubeClient   kubernetes.Interface
}

type PodLog struct {
	Args *PodLogArgs
}

func NewPodLog(args *OuterRequestArgs) (*PodLog, error) {
	podLogArgs, err := checkAndTransferArgs(args)
	if err != nil {
		return nil, err
	}
	return &PodLog{
		Args: podLogArgs,
	}, nil
}

func (pl *PodLog) GetPodLogEntry(accept func(io.ReadCloser)) error {
	err := pl.ensureContainerStarted()
	if err != nil {
		return err
	}
	readCloser, err := pl.Args.KubeClient.CoreV1().Pods(pl.Args.Namespace).GetLogs(pl.Args.PodName, &v1.PodLogOptions{
		// Container:    p.container,
		Follow:       pl.Args.Follow,
		Timestamps:   pl.Args.Timestamps,
		SinceSeconds: pl.Args.SinceSeconds,
		SinceTime:    pl.Args.SinceTime,
		TailLines:    pl.Args.Tail,
	}).Stream()

	if err != nil {
		return err
	}
	// warning: readCloser should execute readCloser.Close() in accept function.
	go accept(readCloser)
	return nil
}

func (pl *PodLog) ensureContainerStarted() error {
	for pl.Args.RetryCnt > 0 {
		pod, err := pl.Args.KubeClient.CoreV1().Pods(pl.Args.Namespace).Get(pl.Args.PodName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		status, _, _, _ := servejob.DefinePodPhaseStatus(*pod)
		log.Debugf("pod:%s,pod phase: %v\n", pl.Args.PodName, pod.Status.Phase)
		log.Debugf("pod print status: %s\n", status)
		switch podPhase := pod.Status.Phase; {
		case podPhase == v1.PodRunning && status != "Unknown":
			return nil
		case podPhase == v1.PodFailed || podPhase == v1.PodSucceeded:
			return nil
		}
		pl.Args.RetryCnt--
	}
	return fmt.Errorf("instance %s %s", pl.Args.PodName, ErrPodNotRunning.Error())
}
func checkAndTransferArgs(out *OuterRequestArgs) (*PodLogArgs, error) {
	podLogArgs := &PodLogArgs{
		PodName:    out.PodName,
		Namespace:  out.Namespace,
		KubeClient: out.KubeClient,
		Follow:     out.Follow,
		RetryCnt:   out.RetryCount,
		Timestamps: out.Timestamps,
	}
	if out.Tail > 0 {
		t := int64(out.Tail)
		podLogArgs.Tail = &t
	}
	if out.SinceTime != "" {
		st, err := ParseSinceTime(out.SinceTime)
		if err != nil {
			return nil, err
		}
		podLogArgs.SinceTime = st
	}
	if out.SinceSeconds != 0 {
		rounded := RoundSeconds(out.SinceSeconds)
		podLogArgs.SinceSeconds = &rounded
	}
	return podLogArgs, nil
}

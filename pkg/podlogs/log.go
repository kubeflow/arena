package podlogs

import (
	"fmt"
	"io"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
		if len(pod.Status.ContainerStatuses) == 0 {
			time.Sleep(pl.Args.RetryTimeout)
			pl.Args.RetryCnt--
			continue
		}
		// warning: would be a bug if container in a pod more than one.
		var containerStatus *v1.ContainerStatus = &pod.Status.ContainerStatuses[0]
		// for _, status := range pod.Status.ContainerStatuses {
		// 	if status.Name == container {
		// 		containerStatus = &status
		// 		break
		// 	}
		// }
		if containerStatus == nil || containerStatus.State.Waiting != nil {
			time.Sleep(pl.Args.RetryTimeout)
			pl.Args.RetryCnt--
		} else {
			return nil
		}
	}
	return fmt.Errorf("pod '%s' has not been started.", pl.Args.PodName)
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
	if out.SinceSeconds != "" {
		ss, err := ParseSinceSeconds(out.SinceSeconds)
		if err != nil {
			return nil, err
		}
		podLogArgs.SinceSeconds = ss
	}
	return podLogArgs, nil
}

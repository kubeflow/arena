package commands

import (
	"io"
	"path"
	"time"

	podlogs "github.com/kubeflow/arena/pkg/podlogs"
	tlogs "github.com/kubeflow/arena/pkg/printer/base/logs"
	"k8s.io/client-go/kubernetes"
)

// AcceptJobLog is used for arena-go-sdk
func AcceptJobLog(kubeconfig, logLevel, ns, jobName, jobType, sinceSeconds, sinceTime, instance string, follow, timestamps bool, tail int, writeCloser io.WriteCloser) (int, error) {
	if err := InitCommonConfig(kubeconfig, logLevel, ns); err != nil {
		return 1, err
	}
	job, err := searchTrainingJob(jobName, jobType, namespace)
	if err != nil {
		return 2, err
	}
	conf, err := clientConfig.ClientConfig()
	if err != nil {
		return 3, err
	}
	var outerArgs = &podlogs.OuterRequestArgs{}
	outerArgs.Namespace = namespace
	outerArgs.RetryCount = 5
	outerArgs.RetryTimeout = time.Millisecond
	outerArgs.SinceSeconds = sinceSeconds
	outerArgs.SinceTime = sinceTime
	outerArgs.Follow = follow
	outerArgs.Tail = tail
	outerArgs.PodName = instance
	outerArgs.Timestamps = timestamps
	outerArgs.KubeClient = kubernetes.NewForConfigOrDie(conf)
	names := []string{}
	for _, pod := range job.AllPods() {
		names = append(names, path.Base(pod.ObjectMeta.SelfLink))
	}
	chiefPod := job.ChiefPod()
	if len(names) > 1 && outerArgs.PodName == "" {
		names = []string{path.Base(chiefPod.ObjectMeta.SelfLink)}
	}
	logPrinter, err := tlogs.NewPodLogPrinter(names, outerArgs)
	if err != nil {
		return 3, err
	}
	return logPrinter.AcceptLogs(writeCloser)
}

package serving

import (
	"fmt"
	"strings"

	"github.com/kubeflow/arena/pkg/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Get all jobs under the assigned conditons.
func NewServingJobList(client *kubernetes.Clientset, servingName string, ns string) ([]Serving, error) {
	jobs := []Serving{}
	deployments, err := client.AppsV1().Deployments(ns).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("serviceName=%s", servingName),
	})
	if err != nil {
		return nil, fmt.Errorf("Failed due to %v", err)
	}
	podListObject, err := client.CoreV1().Pods(ns).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("serviceName=%s", servingName),
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to get pods by label serviceName=%s,reason=%s", servingName, err.Error())
	}

	for _, deploy := range deployments.Items {
		jobs = append(jobs, NewServingJob(client, deploy, podListObject.Items))
	}

	if len(jobs) == 0 {
		return nil, types.ErrNotFoundJobs
	}
	return jobs, nil
}

// filter jobs under the assigned conditions.
func FilterJobs(namespace, version, servingTypeKey string, jobs []Serving) []Serving {
	filterJobs := []Serving{}
	for _, job := range jobs {
		isMatchedNamespace := job.IsMatchedGivenCondition(namespace, "NAMESPACE")
		isMatchedVersion := job.IsMatchedGivenCondition(version, "VERSION")
		isMatchedType := job.IsMatchedGivenCondition(servingTypeKey, "TYPE")
		if isMatchedNamespace && isMatchedVersion && isMatchedType {
			filterJobs = append(filterJobs, job)
		}
	}
	return filterJobs
}

// print the help info  when jobs more than one
func GetMultiJobsHelpInfo(jobs []Serving) string {
	header := fmt.Sprintf("There is %d jobs have been found:", len(jobs))
	tableHeader := "NAME\tTYPE\tVERSION"
	printLines := []string{tableHeader}
	footer := fmt.Sprintf("please use \"--type\" or \"--version\" to filter.")
	for _, job := range jobs {
		line := fmt.Sprintf("%s\t%s\t%s",
			job.Name,
			string(job.ServeType),
			job.Version,
		)
		printLines = append(printLines, line)
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(printLines, "\n"), footer)
}

func GetOnlyOneJob(client *kubernetes.Clientset, ns, servingName, servingTypeKey, version string) (Serving, string, error) {
	allJobs, err := NewServingJobList(client, servingName, ns)
	if err != nil {
		return Serving{}, "", err
	}
	filterJobs := FilterJobs(ns, version, servingTypeKey, allJobs)
	if len(filterJobs) == 0 {
		return Serving{}, "", types.ErrNotFoundJobs
	} else if len(filterJobs) > 1 {
		return Serving{}, GetMultiJobsHelpInfo(filterJobs), types.ErrTooManyJobs
	}
	return filterJobs[0], "", nil
}
func DefinePodPhaseStatus(pod v1.Pod) (string, int, int, int) {
	restarts := 0
	totalContainers := len(pod.Spec.Containers)
	readyContainers := 0

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}
	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			reason = "Running"
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}
	return reason, totalContainers, restarts, readyContainers
}

func KeyMapServingType(servingKey string) types.ServingType {
	switch servingKey {
	case "tf", "tf-serving", "tensorflow-serving":
		return types.ServingTF
	case "trt", "trt-serving", "tensorrt-serving":
		return types.ServingTRT
	case "custom", "custom-serving":
		return types.ServingCustom
	default:
		return types.ServingType("")
	}
}

func CheckServingTypeIsOk(stype string) error {
	if stype == "" {
		return nil
	}
	if KeyMapServingType(stype) == types.ServingType("") {
		return fmt.Errorf("unknow serving type: %s", stype)
	}
	return nil
}

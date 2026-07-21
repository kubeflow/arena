package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	outputpkg "github.com/kubeflow/arena/pkg/output"
	"github.com/kubeflow/arena/pkg/provider"
	"github.com/kubeflow/arena/pkg/task"
)

var getDetails bool

var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get detailed information about a training job",
	Long:  `Retrieve and display detailed information about a training job, including its status and pod details.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate format
		if err := outputpkg.OutputFormat(outputFormat).Validate(); err != nil {
			return err
		}

		name := args[0]

		k8sClient, err := client.NewClient(kubeconfig, kubeContext)
		if err != nil {
			return fmt.Errorf("failed to create K8s client: %w", err)
		}

		ns := resolveNS("")
		jobKind, err := detectJobType(cmdContext(cmd), k8sClient, ns, name)
		if err != nil {
			return err
		}

		job, err := k8sClient.Get(cmdContext(cmd), jobKind, ns, name)
		if err != nil {
			return fmt.Errorf("failed to get %s %s: %w", jobKind, name, err)
		}

		status := extractJobStatus(job, jobKind)
		if fw, ok := job.GetLabels()[FrameworkLabel]; ok && fw != "" {
			status.Framework = fw
		} else {
			status.Framework = kindToFramework(jobKind)
		}
		status.GPURequested = extractGPURequested(job)

		// Get real pods via typed client using provider selector
		p, pErr := providerForKind(jobKind)
		var podList []client.PodInfo
		if pErr == nil {
			selector := p.GetJobPodSelector(name)
			podList = getRealPods(cmdContext(cmd), ns, selector)
		}
		// Fallback to synthetic pods from CRD spec if no real pods found
		if len(podList) == 0 {
			podList = extractPods(job)
		}

		info := &client.JobInfo{
			Status: status,
			Pods:   podList,
		}

		if getDetails {
			cm, err := k8sClient.Get(cmdContext(cmd), "ConfigMap", ns, name)
			if err != nil {
				if !apierrors.IsNotFound(err) {
					return fmt.Errorf("failed to get ConfigMap: %w", err)
				}
				// ConfigMap not found, skip configuration display
			} else {
				data, found, err := unstructured.NestedMap(cm.Object, "data")
				if err == nil && found {
					yamlContent, ok := data["arena-v2.yaml"].(string)
					if ok && yamlContent != "" {
						var config task.Task
						if err := yaml.Unmarshal([]byte(yamlContent), &config); err == nil {
							info.Configuration = &config
						}
					}
				}
			}
		}

		renderer := &outputpkg.TableRenderer{}
		opts := outputpkg.RenderOptions{
			TableFn: func() string { return renderer.RenderJobDetail(info) },
		}
		if err := outputpkg.OutputFormat(outputFormat).Render(info, opts); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	jobCmd.AddCommand(getCmd)
	getCmd.Flags().BoolVar(&getDetails, "details", false, "show job configuration details")
}

// providerForKind returns the Provider for a given CRD kind.
func providerForKind(kind string) (provider.Provider, error) {
	switch kind {
	case constants.KindPyTorchJob:
		return &provider.PyTorchProvider{}, nil
	case constants.KindTFJob:
		return &provider.TensorFlowProvider{}, nil
	case constants.KindMPIJob:
		return &provider.MPIProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported kind: %s", kind)
	}
}

// getRealPods queries the typed client for pods matching the given label selector.
// Returns PodInfo for each pod found, or nil if the query fails or returns no pods.
// Errors are logged at Warning level before returning nil.
func getRealPods(ctx context.Context, namespace, selector string) []client.PodInfo {
	config, err := client.LoadRestConfig(kubeconfig, kubeContext)
	if err != nil {
		log.Warning("failed to load REST config for pod lookup", "error", err.Error())
		return nil
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Warning("failed to create clientset for pod lookup", "error", err.Error())
		return nil
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(
		ctx,
		metav1.ListOptions{LabelSelector: selector},
	)
	if err != nil {
		log.Warning("failed to list pods", "namespace", namespace, "selector", selector, "error", err.Error())
		return nil
	}
	if len(pods.Items) == 0 {
		return nil
	}

	var result []client.PodInfo
	for _, pod := range pods.Items {
		result = append(result, client.PodInfo{
			Name:   pod.Name,
			Status: podDisplayStatus(&pod),
			IP:     pod.Status.PodIP,
			Node:   pod.Spec.NodeName,
		})
	}
	return result
}

// podDisplayStatus returns the most informative status string for a pod,
// mirroring kubectl's STATUS column logic. Container waiting reasons
// (e.g. ImagePullBackOff, CrashLoopBackOff) take priority over the
// coarse pod phase when the pod is not in a terminal state.
func podDisplayStatus(pod *corev1.Pod) string {
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return string(pod.Status.Phase)
	}
	for _, c := range pod.Status.InitContainerStatuses {
		if c.State.Waiting != nil && c.State.Waiting.Reason != "" {
			return c.State.Waiting.Reason
		}
	}
	for _, c := range pod.Status.ContainerStatuses {
		if c.State.Waiting != nil && c.State.Waiting.Reason != "" {
			return c.State.Waiting.Reason
		}
	}
	return string(pod.Status.Phase)
}

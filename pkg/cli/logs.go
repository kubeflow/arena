package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
)

var (
	logsFollow    bool
	logsTail      int
	logsPod       string // pod name (skip label selector)
	logsContainer string // container name (default: first container)
)

var logsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "View logs from a training job",
	Long: `Stream logs from a training job pod. By default, streams from the primary pod
(master/chief/launcher). Use --pod to specify a pod by name and --container to select
a specific container within the pod.

Examples:
  # View master pod logs
  arena job logs my-job

  # View worker pod logs
  arena job logs my-job --pod my-job-worker-0

  # View specific container logs
  arena job logs my-job --pod my-job-worker-0 --container tensorboard

  # Follow logs with tail
  arena job logs my-job -f --tail 100`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Use arena client to find the job kind
		k8sClient, err := client.NewClient(kubeconfig, kubeContext)
		if err != nil {
			return fmt.Errorf("failed to create K8s client: %w", err)
		}

		ns := resolveNS("")
		// Determine job kind using ConfigMap anchor (consistent with other commands)
		jobKind, err := detectJobType(cmdContext(cmd), k8sClient, ns, name)
		if err != nil {
			return err
		}

		// Get the matching provider for this job kind
		frameworkName := kindToFramework(jobKind)
		p, err := getProvider(frameworkName)
		if err != nil {
			return fmt.Errorf("failed to get provider for %s: %w", jobKind, err)
		}

		// Build the clientset for pod log streaming
		config, err := client.LoadRestConfig(kubeconfig, kubeContext)
		if err != nil {
			return fmt.Errorf("failed to build config: %w", err)
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("failed to create clientset: %w", err)
		}

		var podName string
		var selectedPod *corev1.Pod

		if logsPod != "" {
			// User specified pod name - validate it exists and belongs to job
			pod, err := clientset.CoreV1().Pods(ns).Get(
				cmdContext(cmd),
				logsPod,
				metav1.GetOptions{},
			)
			if err != nil {
				if apierrors.IsNotFound(err) {
					return fmt.Errorf("pod %q not found in namespace %q", logsPod, ns)
				}
				return fmt.Errorf("failed to get pod %q: %w", logsPod, err)
			}

			// Validate pod belongs to job using provider's label convention
			if !podBelongsToJob(pod, name) {
				return fmt.Errorf("pod %q does not belong to job %q", logsPod, name)
			}

			podName = pod.Name
			selectedPod = pod
		} else {
			// Use the provider's log pod selector
			selector := p.GetLogPodSelector(name)

			// Find pod using provider's selector
			pods, err := clientset.CoreV1().Pods(ns).List(
				cmdContext(cmd),
				metav1.ListOptions{
					LabelSelector: selector,
				},
			)
			if err != nil {
				return fmt.Errorf("failed to list pods: %w", err)
			}

			// For TFJob, if chief selector found zero pods, fall back to worker replica
			// (chief is conditional in TFJob; when absent, worker is the primary log target)
			if len(pods.Items) == 0 && jobKind == constants.KindTFJob {
				fallbackSelector := fmt.Sprintf("%s=%s,%s=%s,%s=%s",
					constants.LabelJobName, name,
					constants.LabelReplicaType, constants.ReplicaRoleWorker,
					constants.LabelReplicaIndex, "0")
				pods, err = clientset.CoreV1().Pods(ns).List(
					cmdContext(cmd),
					metav1.ListOptions{
						LabelSelector: fallbackSelector,
					},
				)
				if err != nil {
					return fmt.Errorf("failed to list pods with fallback selector: %w", err)
				}
				selector = fallbackSelector
			}

			if len(pods.Items) == 0 {
				return fmt.Errorf("no pods found for job %q in namespace %q using selector %q", name, ns, selector)
			}

			podName = pods.Items[0].Name
			selectedPod = &pods.Items[0]
		}

		// Build log options.
		logOptions := &corev1.PodLogOptions{
			Follow: logsFollow,
		}
		if logsTail > 0 {
			tail := int64(logsTail)
			logOptions.TailLines = &tail
		}

		// Validate container if specified.
		if logsContainer != "" {
			if !containerExists(selectedPod, logsContainer) {
				available := getAvailableContainers(selectedPod)
				return fmt.Errorf("container %q not found in pod %q (available: %s)",
					logsContainer, podName, strings.Join(available, ", "))
			}

			logOptions.Container = logsContainer
		}

		// Stream logs from the pod.
		req := clientset.CoreV1().Pods(ns).GetLogs(podName, logOptions)
		stream, err := req.Stream(cmdContext(cmd))
		if err != nil {
			return fmt.Errorf("failed to stream logs from pod %s: %w", podName, err)
		}
		defer stream.Close()

		scanner := bufio.NewScanner(stream)
		// Increase buffer from default 64KB to 1MB — training logs can emit
		// long JSON lines or stack traces that exceed the default token limit.
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

		if err := scanner.Err(); err != nil && err != io.EOF {
			return fmt.Errorf("error reading log stream: %w", err)
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "follow log output")
	logsCmd.Flags().IntVar(&logsTail, "tail", -1, "number of lines to show from end")
	logsCmd.Flags().StringVar(&logsPod, "pod", "", "pod name (skip label selector)")
	logsCmd.Flags().StringVar(&logsContainer, "container", "", "container name (default: first container)")
	jobCmd.AddCommand(logsCmd)
}

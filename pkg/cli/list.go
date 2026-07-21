package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	outputpkg "github.com/kubeflow/arena/pkg/output"
)

// supportedJobKinds lists the CRD kinds that arena manages.
var supportedJobKinds = []string{constants.KindPyTorchJob, constants.KindTFJob, constants.KindMPIJob}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all training jobs",
	Long:  `List all training jobs across PyTorchJob, TFJob, and MPIJob CRD kinds.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate format
		if err := outputpkg.OutputFormat(outputFormat).Validate(); err != nil {
			return err
		}

		k8sClient, err := client.NewClient(kubeconfig, kubeContext)
		if err != nil {
			return fmt.Errorf("failed to create K8s client: %w", err)
		}

		mpiAvailable := true
		if err := k8sClient.ResolveMPIVersion(cmdContext(cmd)); err != nil {
			log.Debug("MPIJob CRD not available", "error", err.Error())
			mpiAvailable = false
		}

		ns := resolveNS("")
		var allJobs []client.JobStatus
		for _, kind := range supportedJobKinds {
			if kind == constants.KindMPIJob && !mpiAvailable {
				continue
			}
			jobs, err := k8sClient.List(cmdContext(cmd), kind, ns, V2LabelSelector)
			if err != nil {
				if apierrors.IsNotFound(err) {
					log.Debug("CRD not installed", "kind", kind)
				} else {
					apiVer, _ := k8sClient.KindToAPIVersion(kind) // best-effort for logging
					log.Warning("failed to list CRD kind", "kind", kind, "apiVersion", apiVer, "error", err.Error())
				}
				continue
			}
			for _, job := range jobs {
				status := extractJobStatus(job, kind)
				if fw, ok := job.GetLabels()[FrameworkLabel]; ok && fw != "" {
					status.Framework = fw
				} else {
					status.Framework = kindToFramework(kind)
				}
				status.GPURequested = extractGPURequested(job)
				allJobs = append(allJobs, status)
			}
		}

		renderer := &outputpkg.TableRenderer{}
		opts := outputpkg.RenderOptions{
			TableFn: func() string { return renderer.RenderJobList(allJobs) },
			WideFn:  func() string { return renderer.RenderJobListWide(allJobs) },
		}
		if err := outputpkg.OutputFormat(outputFormat).Render(allJobs, opts); err != nil {
			return err
		}
		return nil
	},
}

// extractJobStatus converts an unstructured CRD object into a JobStatus.
func extractJobStatus(obj *unstructured.Unstructured, kind string) client.JobStatus {
	return client.JobStatus{
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(),
		Status:     extractJobPhase(obj),
		APIVersion: obj.GetAPIVersion(),
		Replicas:   extractReplicas(obj),
		Ready:      extractReady(obj),
		Age:        formatAge(obj.GetCreationTimestamp().Time),
	}
}

// extractJobPhase reads the job phase/status from the CRD status field.
// Kubeflow training operators maintain a list of conditions where multiple
// conditions can have status "True" simultaneously (cumulative, not exclusive).
// For example, a completed job has both Created=True and Succeeded=True.
// Conditions are appended chronologically, so we iterate in reverse to find
// the most recent (current) state.
//
// Fallback chain:
//  1. Last condition with status=="True" (reverse scan) → return its type
//  2. spec.runPolicy.suspend==true → "Suspended"
//  3. No conditions (empty or missing) → "Pending"
//  4. Conditions exist but none True → "Unknown"
func extractJobPhase(obj *unstructured.Unstructured) string {
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err != nil {
		return constants.JobStatusUnknown
	}

	// Reverse scan: find the last condition with status=="True".
	// Kubeflow appends conditions chronologically, so the last True is current.
	for i := len(conditions) - 1; i >= 0; i-- {
		cond, ok := conditions[i].(map[string]interface{})
		if !ok {
			continue
		}
		status, _, _ := unstructured.NestedString(cond, "status")
		if status == "True" {
			condType, _, _ := unstructured.NestedString(cond, "type")
			if condType != "" {
				return condType
			}
		}
	}

	// No True condition found. Check if the job is suspended via runPolicy.
	suspended, suspFound, _ := unstructured.NestedBool(obj.Object, "spec", "runPolicy", "suspend")
	if suspFound && suspended {
		return constants.JobStatusSuspended
	}

	// If no conditions exist at all, the job hasn't started yet.
	if !found || len(conditions) == 0 {
		return constants.JobStatusPending
	}

	// Conditions exist but none are True.
	return constants.JobStatusUnknown
}

// extractReplicas reads the total desired replica count from the CRD spec.
// Kubeflow CRDs nest replica counts under spec.<framework>ReplicaSpecs.<Role>.replicas.
func extractReplicas(obj *unstructured.Unstructured) int {
	spec, found, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil || !found {
		return 0
	}

	total := 0
	// Iterate over spec keys (e.g., pytorchReplicaSpecs, tfReplicaSpecs, mpiReplicaSpecs)
	for _, val := range spec {
		replicaSpecs, ok := val.(map[string]interface{})
		if !ok {
			continue
		}
		// Iterate over role keys (e.g., Worker, Master, Launcher)
		for _, roleVal := range replicaSpecs {
			roleSpec, ok := roleVal.(map[string]interface{})
			if !ok {
				continue
			}
			replicas, found, err := unstructured.NestedInt64(roleSpec, "replicas")
			if err == nil && found {
				total += int(replicas)
			}
		}
	}

	return total
}

// extractReady reads the number of ready replicas from the CRD status.
func extractReady(obj *unstructured.Unstructured) int {
	replicaStatuses, found, err := unstructured.NestedMap(obj.Object, "status", "replicaStatuses")
	if err != nil || !found {
		return 0
	}

	total := 0
	for _, val := range replicaStatuses {
		statusMap, ok := val.(map[string]interface{})
		if !ok {
			continue
		}
		active, _, _ := unstructured.NestedInt64(statusMap, "active")
		succeeded, _, _ := unstructured.NestedInt64(statusMap, "succeeded")
		total += int(active) + int(succeeded)
	}

	return total
}

// extractPods builds synthetic pod info entries from the aggregate replica
// counts (active, succeeded, failed) in the CRD status.replicaStatuses.
//
// Kubeflow training operators store only aggregate counts per role in
// replicaStatuses — not individual pod entries. This function synthesizes
// pod names (e.g. "Worker-0", "Worker-1") and maps them to statuses derived
// from the counters.
//
// IMPORTANT: This is a fallback of last resort. The primary pod discovery path
// is getRealPods (in get.go), which queries actual pods via the Kubernetes API
// using a label selector. Synthetic pod names do not correspond to real pod
// names and should only be displayed when the API query fails or returns no
// results (e.g. insufficient RBAC permissions, API server unreachable).
func extractPods(obj *unstructured.Unstructured) []client.PodInfo {
	replicaStatuses, found, err := unstructured.NestedMap(obj.Object, "status", "replicaStatuses")
	if err != nil || !found {
		return nil
	}

	var pods []client.PodInfo
	for role, val := range replicaStatuses {
		statusMap, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		active, _, _ := unstructured.NestedInt64(statusMap, "active")
		succeeded, _, _ := unstructured.NestedInt64(statusMap, "succeeded")
		failed, _, _ := unstructured.NestedInt64(statusMap, "failed")

		idx := 0
		for i := int64(0); i < active; i++ {
			pods = append(pods, client.PodInfo{
				Name:   fmt.Sprintf("%s-%d", role, idx),
				Status: "Running",
			})
			idx++
		}
		for i := int64(0); i < succeeded; i++ {
			pods = append(pods, client.PodInfo{
				Name:   fmt.Sprintf("%s-%d", role, idx),
				Status: "Succeeded",
			})
			idx++
		}
		for i := int64(0); i < failed; i++ {
			pods = append(pods, client.PodInfo{
				Name:   fmt.Sprintf("%s-%d", role, idx),
				Status: "Failed",
			})
			idx++
		}
	}

	return pods
}

// formatAge returns a human-readable duration string from a creation timestamp.
func formatAge(creationTime time.Time) string {
	if creationTime.IsZero() {
		return "<unknown>"
	}
	return duration.HumanDuration(time.Since(creationTime))
}

func init() {
	jobCmd.AddCommand(listCmd)
}

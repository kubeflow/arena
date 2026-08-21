package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	outputpkg "github.com/kubeflow/arena/pkg/output"
)

// supportedJobKinds lists the CRD kinds that arena manages.
var supportedJobKinds = []string{constants.KindPyTorchJob, constants.KindTFJob, constants.KindMPIJob}

var (
	listSelector      string // -l/--selector: kubectl label selector expression
	listType          string // --type: framework filter (translated to a label)
	listStatus        string // --status: client-side phase filter
	listLast          int    // --last: newest-first cap (0 = unlimited)
	listAllNamespaces bool   // -A/--all-namespaces: list across all namespaces
)

// defaultListLast bounds job list output so agent context windows are not
// flooded by large clusters.
const defaultListLast = 50

// validJobStatuses enumerates the phase values `--status` accepts: Kubeflow
// training operator condition types plus arena's fallback phases (see
// extractJobPhase).
var validJobStatuses = []string{"Created", "Running", "Succeeded", "Failed", "Restarting", "Pending", "Suspended", "Unknown"}

// containsFold reports whether s matches one of values case-insensitively.
func containsFold(values []string, s string) bool {
	for _, v := range values {
		if strings.EqualFold(v, s) {
			return true
		}
	}
	return false
}

// validateListType validates --type case-insensitively against the supported
// frameworks and returns the canonical framework name used as label value.
func validateListType(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	for _, fw := range v2FrameworkNames() {
		if strings.EqualFold(fw, v) {
			return fw, nil
		}
	}
	return "", &CLIError{
		Message:     fmt.Sprintf("invalid value %q for --type", v),
		ValidValues: v2FrameworkNames(),
	}
}

// validateListStatus validates --status case-insensitively against the phases
// extractJobPhase can derive.
func validateListStatus(v string) error {
	if v == "" {
		return nil
	}
	if !containsFold(validJobStatuses, v) {
		return &CLIError{
			Message:     fmt.Sprintf("invalid value %q for --status", v),
			ValidValues: validJobStatuses,
		}
	}
	return nil
}

// validateListLast rejects negative --last values.
func validateListLast(n int) error {
	if n < 0 {
		return &CLIError{
			Message: fmt.Sprintf("invalid value %d for --last", n),
			Hint:    "use a positive number of jobs, or 0 for unlimited",
		}
	}
	return nil
}

// validateListSelector pre-validates -l/--selector syntax client-side so a
// typo fails fast with a teaching error instead of an API server 400.
func validateListSelector(selector string) error {
	if selector == "" {
		return nil
	}
	if _, err := labels.Parse(selector); err != nil {
		return &CLIError{
			Message: fmt.Sprintf("invalid label selector %q: %v", selector, err),
			Hint:    "expected kubectl selector syntax, e.g. -l app=training or -l 'tier in (frontend, backend)'",
		}
	}
	return nil
}

// buildListSelector composes the server-side label selector for job list.
// The arena.io/framework key-exists requirement scopes results to
// arena-v2-created jobs; an explicit --type replaces it with the equality
// form. A user -l selector is AND-combined with a comma.
func buildListSelector(jobType, userSelector string) string {
	parts := make([]string, 0, 2)
	if jobType != "" {
		parts = append(parts, frameworkLabel+"="+jobType)
	} else {
		parts = append(parts, v2LabelSelector)
	}
	if userSelector != "" {
		parts = append(parts, userSelector)
	}
	return strings.Join(parts, ",")
}

// applyStatusFilter keeps jobs whose derived phase matches status
// case-insensitively; an empty status keeps everything.
func applyStatusFilter(jobs []client.JobStatus, status string) []client.JobStatus {
	if status == "" {
		return jobs
	}
	result := make([]client.JobStatus, 0, len(jobs))
	for _, job := range jobs {
		if strings.EqualFold(job.Status, status) {
			result = append(result, job)
		}
	}
	return result
}

// listedJob pairs a listed CRD object with the kind it was listed under, so
// sorting and extraction survive the cross-kind merge.
type listedJob struct {
	kind string
	obj  *unstructured.Unstructured
}

// sortListedJobsNewestFirst orders jobs by creation timestamp, newest first,
// so a bounded list shows the most recent jobs.
func sortListedJobsNewestFirst(all []listedJob) {
	sort.SliceStable(all, func(i, j int) bool {
		return all[i].obj.GetCreationTimestamp().After(all[j].obj.GetCreationTimestamp().Time)
	})
}

// extractAllJobStatuses converts listed CRD objects into enriched JobStatus
// values, preserving the input order.
func extractAllJobStatuses(all []listedJob) []client.JobStatus {
	jobs := make([]client.JobStatus, 0, len(all))
	for _, lj := range all {
		status := extractJobStatus(lj.obj, lj.kind)
		if fw, ok := lj.obj.GetLabels()[frameworkLabel]; ok && fw != "" {
			status.Framework = fw
		} else {
			status.Framework = kindToFramework(lj.kind)
		}
		status.GPURequested = extractGPURequested(lj.obj)
		jobs = append(jobs, status)
	}
	return jobs
}

// limitJobList truncates jobs to limit (0 = unlimited). When entries are
// dropped it writes a narrowing hint to w so output stays bounded at every
// layer without polluting stdout.
func limitJobList(w io.Writer, jobs []client.JobStatus, limit int) []client.JobStatus {
	if limit <= 0 || len(jobs) <= limit {
		return jobs
	}
	fmt.Fprintf(w, "Showing %d of %d jobs. Use --status or --selector to narrow, or --last=0 to show all.\n", limit, len(jobs))
	return jobs[:limit]
}

// listJobsAcrossKinds lists every supported job kind with the given label
// selector and merges the results. Missing CRDs are tolerated (logged and
// skipped); MPIJob is skipped when the MPI CRD version is unresolved. Real
// API errors are collected per kind: when every kind fails, the run fails;
// partial failures are returned so the caller can warn that results may be
// incomplete.
func listJobsAcrossKinds(ctx context.Context, k8sClient *client.Client, ns, selector string) ([]listedJob, []string, error) {
	all := make([]listedJob, 0)
	apiErrors := make([]string, 0)
	anySucceeded := false
	for _, kind := range supportedJobKinds {
		if kind == constants.KindMPIJob && k8sClient.GetMPIVersion() == "" {
			continue
		}
		jobs, err := k8sClient.List(ctx, kind, ns, selector)
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Debug("CRD not installed", "kind", kind)
				continue
			}
			apiVer, _ := k8sClient.KindToAPIVersion(kind)
			log.Warning("failed to list CRD kind", "kind", kind, "apiVersion", apiVer, "error", err.Error())
			apiErrors = append(apiErrors, fmt.Sprintf("%s: %s", kind, err.Error()))
			continue
		}
		anySucceeded = true
		for _, job := range jobs {
			all = append(all, listedJob{kind: kind, obj: job})
		}
	}
	if !anySucceeded && len(apiErrors) > 0 {
		return nil, apiErrors, fmt.Errorf("failed to list any job types; checked %d kind(s)", len(apiErrors))
	}
	return all, apiErrors, nil
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all training jobs",
		Long: `List arena-v2 training jobs across PyTorchJob, TFJob, and MPIJob CRD kinds.

Narrow results with --type (framework), --status (phase), or -l/--selector
(kubectl label selector syntax). --last bounds the output to the N most
recent jobs (default 50, 0 = unlimited). -A/--all-namespaces lists jobs in
every namespace and adds a NAMESPACE column to table output.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := validateOutputFormat(); err != nil {
				return err
			}
			jobType, err := validateListType(listType)
			if err != nil {
				return err
			}
			if err := validateListStatus(listStatus); err != nil {
				return err
			}
			if err := validateListLast(listLast); err != nil {
				return err
			}
			if err := validateListSelector(listSelector); err != nil {
				return err
			}

			k8sClient, err := client.NewClient(kubeconfig, kubeContext)
			if err != nil {
				return fmt.Errorf("failed to create K8s client: %w", err)
			}
			if err := k8sClient.ResolveMPIVersion(cmdContext(cmd)); err != nil {
				log.Debug("MPIJob CRD not available", "error", err.Error())
			}

			ns := resolveListNamespace(listAllNamespaces)
			selector := buildListSelector(jobType, listSelector)
			all, apiErrors, err := listJobsAcrossKinds(cmdContext(cmd), k8sClient, ns, selector)
			if err != nil {
				return err
			}
			if len(apiErrors) > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(),
					"\nWarning: failed to list %d job type(s) due to API errors; results may be incomplete:\n", len(apiErrors))
				for _, e := range apiErrors {
					fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", e)
				}
			}

			sortListedJobsNewestFirst(all)
			allJobs := extractAllJobStatuses(all)
			allJobs = applyStatusFilter(allJobs, listStatus)
			allJobs = limitJobList(cmd.ErrOrStderr(), allJobs, listLast)

			renderer := &outputpkg.TableRenderer{}
			opts := outputpkg.RenderOptions{
				TableFn: func() string { return renderer.RenderJobList(allJobs) },
				WideFn:  func() string { return renderer.RenderJobListWide(allJobs) },
			}
			if listAllNamespaces {
				opts.TableFn = func() string { return renderer.RenderJobListAllNamespaces(allJobs) }
			}
			return outputpkg.Format(outputFormat).Render(cmd.OutOrStdout(), allJobs, opts)
		},
	}
	cmd.Flags().StringVarP(&listSelector, "selector", "l", "", "label selector (kubectl syntax, AND-combined with arena's job scope)")
	cmd.Flags().StringVar(&listType, "type", "", "show only jobs of this framework type")
	cmd.Flags().StringVar(&listStatus, "status", "", "show only jobs in this status phase")
	cmd.Flags().IntVar(&listLast, "last", defaultListLast, "number of most recent jobs to show (0 = unlimited)")
	cmd.Flags().BoolVarP(&listAllNamespaces, "all-namespaces", "A", false, "list jobs across all namespaces")

	_ = cmd.RegisterFlagCompletionFunc("type", completeStaticChoices(v2FrameworkNames()...))
	_ = cmd.RegisterFlagCompletionFunc("status", completeStaticChoices(validJobStatuses...))
	_ = cmd.RegisterFlagCompletionFunc("selector", cobra.NoFileCompletions)

	registerOutputFlag(cmd)
	cmd.ValidArgsFunction = cobra.NoFileCompletions
	return cmd
}

// extractJobStatus converts an unstructured CRD object into a JobStatus.
func extractJobStatus(obj *unstructured.Unstructured, _ string) client.JobStatus {
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

// extractJobPhase returns the last True condition (reverse scan), or "Suspended"/"Pending"/"Unknown" as fallbacks.
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

// extractPods synthesizes pod info from CRD replicaStatuses counters (fallback when real pods are unavailable).
func extractPods(obj *unstructured.Unstructured) []client.PodInfo {
	replicaStatuses, found, err := unstructured.NestedMap(obj.Object, "status", "replicaStatuses")
	if err != nil || !found {
		return nil
	}

	pods := make([]client.PodInfo, 0)
	for role, val := range replicaStatuses {
		statusMap, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		active, _, _ := unstructured.NestedInt64(statusMap, "active")
		succeeded, _, _ := unstructured.NestedInt64(statusMap, "succeeded")
		failed, _, _ := unstructured.NestedInt64(statusMap, "failed")

		idx := 0
		for range active {
			pods = append(pods, client.PodInfo{
				Name:   fmt.Sprintf("%s-%d", role, idx),
				Status: "Running",
			})
			idx++
		}
		for range succeeded {
			pods = append(pods, client.PodInfo{
				Name:   fmt.Sprintf("%s-%d", role, idx),
				Status: "Succeeded",
			})
			idx++
		}
		for range failed {
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

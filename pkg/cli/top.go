package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/constants"
	"github.com/kubeflow/arena/pkg/log"
	outputpkg "github.com/kubeflow/arena/pkg/output"
)

var topOutputFormat string

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Display Resource (GPU) usage.",
	Long: `Display Resource (GPU) usage.

Available Commands:
  job         Display Resource (GPU) usage of jobs`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.HelpFunc()(cmd, args)
	},
}

var topJobCmd = &cobra.Command{
	Use:   "job",
	Short: "Display Resource (GPU) usage of jobs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate format
		if err := outputpkg.OutputFormat(topOutputFormat).Validate(); err != nil {
			return err
		}

		k8sClient, err := client.NewClient(kubeconfig, kubeContext)
		if err != nil {
			return fmt.Errorf("failed to create K8s client: %w", err)
		}

		mpiAvailable := true
		if err := k8sClient.ResolveMPIVersion(cmdContext(cmd)); err != nil {
			log.Warning("MPIJob CRD not available", "error", err.Error())
			mpiAvailable = false
		}

		ns := resolveNS("")
		var allJobs []client.JobStatus
		anySucceeded := false
		failedKindCount := 0
		var apiErrors []string
		for _, kind := range supportedJobKinds {
			if kind == constants.KindMPIJob && !mpiAvailable {
				continue
			}
			jobs, err := k8sClient.List(cmdContext(cmd), kind, ns, V2LabelSelector)
			if err != nil {
				if apierrors.IsNotFound(err) {
					// CRD not installed — not an error, just skip silently
					log.Warning("CRD not installed", "kind", kind)
				} else {
					// Real API error (permission, network, etc.) — track and report
					apiVer, _ := k8sClient.KindToAPIVersion(kind) // best-effort for logging
					log.Warning("failed to list CRD kind", "kind", kind, "apiVersion", apiVer, "error", err.Error())
					apiErrors = append(apiErrors, fmt.Sprintf("%s: %s", kind, err.Error()))
					failedKindCount++
				}
				continue
			}
			anySucceeded = true
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

		// If all kinds failed with API errors (not just missing CRDs), return
		// a clear error so the user understands the issue.
		if !anySucceeded && failedKindCount > 0 {
			return fmt.Errorf("failed to list any job types; checked %d kind(s) — check permissions or cluster connectivity", failedKindCount)
		}

		renderer := &outputpkg.TableRenderer{}
		opts := outputpkg.RenderOptions{
			TableFn: func() string { return renderer.RenderTopJob(allJobs) },
			WideFn:  func() string { return renderer.RenderTopJobWide(allJobs) },
		}
		if err := outputpkg.OutputFormat(topOutputFormat).Render(allJobs, opts); err != nil {
			return err
		}
		// Warn the user when some job types could not be listed so they
		// understand the results may be incomplete.
		if failedKindCount > 0 && anySucceeded {
			fmt.Fprintf(cmd.ErrOrStderr(),
				"\nWarning: failed to list %d job type(s) due to API errors; results may be incomplete:\n", failedKindCount)
			for _, e := range apiErrors {
				fmt.Fprintf(cmd.ErrOrStderr(), "  - %s\n", e)
			}
		}
		return nil
	},
}

func init() {
	topJobCmd.Flags().StringVarP(
		&topOutputFormat,
		"output",
		"o",
		string(outputpkg.DefaultFormat),
		outputpkg.FormatHelpText,
	)
	topCmd.AddCommand(topJobCmd)
	rootCmd.AddCommand(topCmd)
}

package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktesting "k8s.io/client-go/testing"

	"github.com/kubeflow/arena/pkg/client"
)

func TestValidateListType(t *testing.T) {
	for _, fw := range v2FrameworkNames() {
		got, err := validateListType(fw)
		require.NoError(t, err, "framework %q should be valid", fw)
		assert.Equal(t, fw, got)
	}

	// Case-insensitive validation, canonical lowercase output.
	got, err := validateListType("PyTorch")
	require.NoError(t, err)
	assert.Equal(t, "pytorch", got)

	got, err = validateListType("")
	require.NoError(t, err)
	assert.Empty(t, got)

	_, err = validateListType("jax")
	var cliErr *CLIError
	require.Error(t, err)
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, []string{"pytorch", "tensorflow", "mpi", "horovod", "deepspeed"}, cliErr.ValidValues)
	assert.Contains(t, err.Error(), `invalid value "jax" for --type`)
}

func TestValidateListStatus(t *testing.T) {
	for _, s := range validJobStatuses {
		require.NoError(t, validateListStatus(s), "status %q should be valid", s)
	}
	require.NoError(t, validateListStatus(""))
	require.NoError(t, validateListStatus("succeeded")) // case-insensitive

	err := validateListStatus("Zombie")
	var cliErr *CLIError
	require.Error(t, err)
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, validJobStatuses, cliErr.ValidValues)
	assert.Contains(t, err.Error(), `invalid value "Zombie" for --status`)
}

func TestValidateListLast(t *testing.T) {
	require.NoError(t, validateListLast(0))
	require.NoError(t, validateListLast(50))

	err := validateListLast(-1)
	var cliErr *CLIError
	require.Error(t, err)
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, "use a positive number of jobs, or 0 for unlimited", cliErr.Hint)
	assert.Contains(t, err.Error(), `invalid value -1 for --last`)
}

func TestValidateListSelector(t *testing.T) {
	require.NoError(t, validateListSelector(""))
	for _, s := range []string{
		"app=training",
		"app==training",
		"tier!=frontend",
		"tier in (frontend, backend)",
		"environment", // key-exists
	} {
		require.NoError(t, validateListSelector(s), "selector %q should be valid", s)
	}

	err := validateListSelector("a=b=c")
	var cliErr *CLIError
	require.Error(t, err)
	require.True(t, errors.As(err, &cliErr))
	assert.Contains(t, err.Error(), `invalid label selector "a=b=c"`)
	assert.Equal(t, "expected kubectl selector syntax, e.g. -l app=training or -l 'tier in (frontend, backend)'", cliErr.Hint)
}

func TestBuildListSelector(t *testing.T) {
	tests := []struct {
		name         string
		jobType      string
		userSelector string
		want         string
	}{
		{name: "empty keeps key-exists scope", want: "arena.io/framework"},
		{name: "type only becomes equality", jobType: "pytorch", want: "arena.io/framework=pytorch"},
		{name: "user selector only", userSelector: "app=training", want: "arena.io/framework,app=training"},
		{name: "type and user selector AND-combined", jobType: "mpi", userSelector: "team=ai", want: "arena.io/framework=mpi,team=ai"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, buildListSelector(tt.jobType, tt.userSelector))
		})
	}
}

func TestBuildListSelector_IsParseable(t *testing.T) {
	for _, sel := range []string{
		buildListSelector("", ""),
		buildListSelector("pytorch", ""),
		buildListSelector("", "app=training"),
		buildListSelector("deepspeed", "tier in (frontend, backend)"),
	} {
		_, err := labels.Parse(sel)
		require.NoError(t, err, "selector %q must be parseable", sel)
	}
}

func jobNames(jobs []client.JobStatus) []string {
	names := make([]string, 0, len(jobs))
	for _, j := range jobs {
		names = append(names, j.Name)
	}
	return names
}

func TestApplyStatusFilter(t *testing.T) {
	jobs := []client.JobStatus{
		{Name: "a", Status: "Running"},
		{Name: "b", Status: "Succeeded"},
		{Name: "c", Status: "Pending"},
	}

	assert.Equal(t, jobs, applyStatusFilter(jobs, ""), "empty status keeps all")
	assert.Equal(t, []string{"a"}, jobNames(applyStatusFilter(jobs, "running")))
	assert.Equal(t, []string{"b"}, jobNames(applyStatusFilter(jobs, "Succeeded")))
	assert.Empty(t, applyStatusFilter(jobs, "Failed"))
}

func TestListJobsAcrossKinds_PassesComposedSelector(t *testing.T) {
	c, fakeDynamic := newListTestClient(t)

	selector := buildListSelector("pytorch", "app=training")
	_, _, err := listJobsAcrossKinds(context.Background(), c, "default", selector)
	require.NoError(t, err)

	var got []string
	for _, action := range fakeDynamic.Actions() {
		if action.GetVerb() != "list" {
			continue
		}
		listAction, ok := action.(ktesting.ListActionImpl)
		require.True(t, ok, "list action should be ListActionImpl, got %T", action)
		got = append(got, listAction.GetListRestrictions().Labels.String())
	}
	require.Len(t, got, 3, "one list call per job kind")

	// labels.Selector.String() renders requirements in canonical (sorted)
	// order, so compare against the parsed expected selector rather than the
	// raw composed string.
	want, err := labels.Parse("arena.io/framework=pytorch,app=training")
	require.NoError(t, err)
	for _, sel := range got {
		assert.Equal(t, want.String(), sel)
	}
}

func TestListCmd_RemovedFlagsAreRejected(t *testing.T) {
	orig := outputFormat
	t.Cleanup(func() { outputFormat = orig })

	for _, flag := range []string{"--filter", "--limit"} {
		err := ExecuteWithArgs([]string{"job", "list", flag, "name=x"})
		require.Error(t, err, "%s should be rejected as unknown flag", flag)
		assert.Contains(t, err.Error(), "unknown flag")
	}
}

func TestResolveListNamespace(t *testing.T) {
	orig := namespace
	t.Cleanup(func() { namespace = orig })

	namespace = "team-a"
	assert.Empty(t, resolveListNamespace(true), "-A must return the all-namespaces scope")
	assert.Equal(t, "team-a", resolveListNamespace(false), "without -A the resolveNS chain applies")

	namespace = "e2e-b"
	assert.Empty(t, resolveListNamespace(true), "-A silently overrides -n")
}

func TestListJobsAcrossKinds_EmptyNamespaceIsPassedThrough(t *testing.T) {
	c, fakeDynamic := newListTestClient(t)

	_, _, err := listJobsAcrossKinds(context.Background(), c, "", "arena.io/framework")
	require.NoError(t, err)

	listed := 0
	for _, action := range fakeDynamic.Actions() {
		if action.GetVerb() != "list" {
			continue
		}
		listAction, ok := action.(ktesting.ListActionImpl)
		require.True(t, ok, "list action should be ListActionImpl, got %T", action)
		assert.Empty(t, listAction.GetNamespace(), "all-namespaces listing must pass an empty namespace to the API")
		listed++
	}
	require.Equal(t, 3, listed, "one list call per job kind")
}

func TestListCmd_AllNamespacesFlagRegistered(t *testing.T) {
	flag := newListCmd().Flags().Lookup("all-namespaces")
	require.NotNil(t, flag, "job list must register --all-namespaces")
	assert.Equal(t, "A", flag.Shorthand)
}

func TestListCmd_AllNamespacesParses(t *testing.T) {
	err := ExecuteWithArgs([]string{"job", "list", "-A", "--kubeconfig", "/nonexistent/kubeconfig"})
	require.Error(t, err, "nonexistent kubeconfig must fail the run")
	assert.NotContains(t, err.Error(), "unknown", "-A must parse as a known flag")
	assert.Contains(t, err.Error(), "failed to create K8s client")
}

func forbiddenReactor(_ ktesting.Action) (bool, runtime.Object, error) {
	return true, nil, apierrors.NewForbidden(
		schema.GroupResource{Group: "kubeflow.org", Resource: "jobs"},
		"list",
		errors.New("insufficient permissions"),
	)
}

func notFoundReactor(_ ktesting.Action) (bool, runtime.Object, error) {
	return true, nil, apierrors.NewNotFound(
		schema.GroupResource{Group: "kubeflow.org", Resource: "pytorchjobs"},
		"",
	)
}

func TestListJobsAcrossKinds_AllKindsAPIErrorsFailLoudly(t *testing.T) {
	c, fakeDynamic := newListTestClient(t)
	fakeDynamic.PrependReactor("list", "*", forbiddenReactor)

	_, apiErrs, err := listJobsAcrossKinds(context.Background(), c, "default", "arena.io/framework")
	require.Error(t, err, "all kinds failing with API errors must surface an error")
	assert.Contains(t, err.Error(), "failed to list any job types")
	assert.Len(t, apiErrs, 3, "one API error per job kind")
}

func TestListJobsAcrossKinds_PartialFailureStillListsJobs(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "kubeflow.org/v1",
		"kind":       "PyTorchJob",
		"metadata": map[string]interface{}{
			"name":      "job-a",
			"namespace": "default",
			"labels":    map[string]interface{}{frameworkLabel: "pytorch"},
		},
	}}
	c, fakeDynamic := newListTestClient(t, obj)
	fakeDynamic.PrependReactor("list", "tfjobs", forbiddenReactor)

	jobs, apiErrs, err := listJobsAcrossKinds(context.Background(), c, "default", "arena.io/framework")
	require.NoError(t, err, "partial failure must not fail the whole listing")
	assert.Len(t, jobs, 1, "jobs from succeeding kinds are still listed")
	assert.Len(t, apiErrs, 1)
	assert.Contains(t, apiErrs[0], "TFJob")
}

func TestListJobsAcrossKinds_AllCRDsMissingReturnsEmpty(t *testing.T) {
	c, fakeDynamic := newListTestClient(t)
	fakeDynamic.PrependReactor("list", "*", notFoundReactor)

	jobs, apiErrs, err := listJobsAcrossKinds(context.Background(), c, "default", "arena.io/framework")
	require.NoError(t, err, "missing CRDs are not API failures")
	assert.Empty(t, jobs)
	assert.Empty(t, apiErrs)
}

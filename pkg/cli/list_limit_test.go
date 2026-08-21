package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"

	"github.com/kubeflow/arena/pkg/client"
	outputpkg "github.com/kubeflow/arena/pkg/output"
)

func TestListCmd_HasSelectionFlags(t *testing.T) {
	cmd := newListCmd()

	sel := cmd.Flags().Lookup("selector")
	require.NotNil(t, sel, "selector flag should be registered")
	assert.Equal(t, "l", sel.Shorthand)
	assert.Equal(t, "", sel.DefValue)

	tp := cmd.Flags().Lookup("type")
	require.NotNil(t, tp, "type flag should be registered")
	assert.Equal(t, "", tp.DefValue)

	st := cmd.Flags().Lookup("status")
	require.NotNil(t, st, "status flag should be registered")
	assert.Equal(t, "", st.DefValue)

	last := cmd.Flags().Lookup("last")
	require.NotNil(t, last, "last flag should be registered")
	assert.Equal(t, "50", last.DefValue)

	assert.Nil(t, cmd.Flags().Lookup("filter"), "--filter should be removed")
	assert.Nil(t, cmd.Flags().Lookup("limit"), "--limit should be removed")
}

func TestSortListedJobsNewestFirst(t *testing.T) {
	mk := func(name string, age time.Duration) listedJob {
		obj := &unstructured.Unstructured{}
		obj.SetName(name)
		obj.SetCreationTimestamp(metav1.NewTime(time.Now().Add(-age)))
		return listedJob{kind: "PyTorchJob", obj: obj}
	}
	all := []listedJob{mk("old", 3*time.Hour), mk("new", 1*time.Minute), mk("mid", 30*time.Minute)}
	sortListedJobsNewestFirst(all)
	names := []string{all[0].obj.GetName(), all[1].obj.GetName(), all[2].obj.GetName()}
	assert.Equal(t, []string{"new", "mid", "old"}, names)
}

func TestExtractAllJobStatuses(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "kubeflow.org/v1",
		"kind":       "PyTorchJob",
		"metadata": map[string]interface{}{
			"name":      "j1",
			"namespace": "default",
			"labels":    map[string]interface{}{frameworkLabel: "deepspeed"},
		},
	}}
	jobs := extractAllJobStatuses([]listedJob{{kind: "MPIJob", obj: obj}})
	require.Len(t, jobs, 1)
	assert.Equal(t, "j1", jobs[0].Name)
	assert.Equal(t, "deepspeed", jobs[0].Framework, "framework label should win over kind fallback")
}

func TestExtractAllJobStatuses_FallbackFramework(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "kubeflow.org/v1",
		"kind":       "TFJob",
		"metadata": map[string]interface{}{
			"name":      "j2",
			"namespace": "default",
		},
	}}
	jobs := extractAllJobStatuses([]listedJob{{kind: "TFJob", obj: obj}})
	require.Len(t, jobs, 1)
	assert.Equal(t, "tensorflow", jobs[0].Framework)
}

func TestLimitJobList_TruncatesWithHint(t *testing.T) {
	jobs := make([]client.JobStatus, 0, 60)
	for i := 0; i < 60; i++ {
		jobs = append(jobs, client.JobStatus{Name: fmt.Sprintf("job-%d", i)})
	}
	var buf bytes.Buffer
	got := limitJobList(&buf, jobs, 50)
	assert.Len(t, got, 50)
	assert.Equal(t, "Showing 50 of 60 jobs. Use --status or --selector to narrow, or --last=0 to show all.\n", buf.String())
}

func TestLimitJobList_ZeroIsUnlimited(t *testing.T) {
	jobs := make([]client.JobStatus, 60)
	var buf bytes.Buffer
	got := limitJobList(&buf, jobs, 0)
	assert.Len(t, got, 60)
	assert.Empty(t, buf.String())
}

func TestLimitJobList_UnderLimitNoHint(t *testing.T) {
	jobs := make([]client.JobStatus, 10)
	var buf bytes.Buffer
	got := limitJobList(&buf, jobs, 50)
	assert.Len(t, got, 10)
	assert.Empty(t, buf.String())
}

func TestLimitJobList_AtLimitNoHint(t *testing.T) {
	jobs := make([]client.JobStatus, 50)
	var buf bytes.Buffer
	got := limitJobList(&buf, jobs, 50)
	assert.Len(t, got, 50)
	assert.Empty(t, buf.String())
}

// newListTestClient builds a fake dynamic client with all three job-kind
// GVRs registered, the MPI version pre-resolved, and the given objects
// preloaded.
func newListTestClient(t *testing.T, objs ...runtime.Object) (*client.Client, *fake.FakeDynamicClient) {
	t.Helper()
	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		{Group: "kubeflow.org", Version: "v1", Resource: "pytorchjobs"}: "PyTorchJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "tfjobs"}:      "TFJobList",
		{Group: "kubeflow.org", Version: "v1", Resource: "mpijobs"}:     "MPIJobList",
	}
	fakeDynamic := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds, objs...)
	c := client.NewClientForInterface(fakeDynamic)
	c.SetMPIVersion("v1")
	return c, fakeDynamic
}

func TestListPipeline_DefaultPathKeepsNewest50(t *testing.T) {
	objs := make([]runtime.Object, 0, 60)
	for i := 0; i < 60; i++ {
		obj := &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "kubeflow.org/v1",
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("job-%02d", i),
				"namespace": "default",
				"labels":    map[string]interface{}{frameworkLabel: "pytorch"},
			},
		}}
		obj.SetCreationTimestamp(metav1.NewTime(time.Now().Add(-time.Duration(60-i) * time.Minute)))
		objs = append(objs, obj)
	}
	c, _ := newListTestClient(t, objs...)

	all, _, err := listJobsAcrossKinds(context.Background(), c, "default", buildListSelector("", ""))
	require.NoError(t, err)
	require.Len(t, all, 60)

	sortListedJobsNewestFirst(all)
	allJobs := extractAllJobStatuses(all)
	allJobs = applyStatusFilter(allJobs, "")
	var buf bytes.Buffer
	allJobs = limitJobList(&buf, allJobs, defaultListLast)

	require.Len(t, allJobs, 50)
	assert.Equal(t, "job-59", allJobs[0].Name)
	assert.Equal(t, "job-10", allJobs[49].Name)
	assert.Equal(t, "Showing 50 of 60 jobs. Use --status or --selector to narrow, or --last=0 to show all.\n", buf.String())

	renderer := &outputpkg.TableRenderer{}
	out := renderer.RenderJobList(allJobs)
	assert.Contains(t, out, "job-59")
	assert.NotContains(t, out, "job-09")
}

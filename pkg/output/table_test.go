package output

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/task"
)

func TestRenderJobList(t *testing.T) {
	jobs := []client.JobStatus{
		{Name: "job-1", Status: "Running", Replicas: 4, Ready: 3, Age: "5m"},
		{Name: "job-2", Status: "Succeeded", Replicas: 2, Ready: 0, Age: "1h"},
	}

	renderer := &TableRenderer{}
	output := renderer.RenderJobList(jobs)

	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "REPLICAS")
	assert.Contains(t, output, "AGE")
	assert.Contains(t, output, "job-1")
	assert.Contains(t, output, "Running")
	assert.Contains(t, output, "3/4")
	assert.Contains(t, output, "job-2")
	assert.Contains(t, output, "Succeeded")
	assert.Contains(t, output, "0/2")
}

func TestRenderJobListEmpty(t *testing.T) {
	renderer := &TableRenderer{}
	output := renderer.RenderJobList(nil)

	assert.Contains(t, output, "No jobs found")
}

func TestRenderJobListEmptySlice(t *testing.T) {
	renderer := &TableRenderer{}
	output := renderer.RenderJobList([]client.JobStatus{})

	assert.Contains(t, output, "No jobs found")
}

func TestRenderJobDetail(t *testing.T) {
	info := &client.JobInfo{
		Status: client.JobStatus{
			Name:      "job-1",
			Namespace: "default",
			Status:    "Running",
			Replicas:  4,
			Ready:     3,
			Age:       "5m",
		},
		Pods: []client.PodInfo{
			{Name: "job-1-pod-0", Status: "Running", IP: "10.0.0.1", Node: "node-1"},
			{Name: "job-1-pod-1", Status: "Running", IP: "10.0.0.2", Node: "node-2"},
		},
	}

	renderer := &TableRenderer{}
	output := renderer.RenderJobDetail(info)

	assert.Contains(t, output, "Name:")
	assert.Contains(t, output, "job-1")
	assert.Contains(t, output, "Namespace:")
	assert.Contains(t, output, "default")
	assert.Contains(t, output, "Status:")
	assert.Contains(t, output, "Running")
	assert.Contains(t, output, "3/4")
	assert.Contains(t, output, "Pods:")
	assert.Contains(t, output, "job-1-pod-0")
	assert.Contains(t, output, "10.0.0.1")
	assert.Contains(t, output, "node-1")
}

func TestRenderJobDetailNoPods(t *testing.T) {
	info := &client.JobInfo{
		Status: client.JobStatus{
			Name:      "job-1",
			Namespace: "default",
			Status:    "Pending",
			Replicas:  2,
			Ready:     0,
			Age:       "1m",
		},
		Pods: nil,
	}

	renderer := &TableRenderer{}
	output := renderer.RenderJobDetail(info)

	assert.Contains(t, output, "Name:")
	assert.Contains(t, output, "job-1")
	assert.Contains(t, output, "Pending")
	assert.NotContains(t, output, "Pods:")
}

func TestRenderJobDetailEmptyPods(t *testing.T) {
	info := &client.JobInfo{
		Status: client.JobStatus{
			Name:      "job-1",
			Namespace: "default",
			Status:    "Running",
			Replicas:  1,
			Ready:     1,
			Age:       "10m",
		},
		Pods: []client.PodInfo{},
	}

	renderer := &TableRenderer{}
	output := renderer.RenderJobDetail(info)

	assert.Contains(t, output, "job-1")
	assert.NotContains(t, output, "Pods:")
}

func TestRenderJobListDynamicWidth(t *testing.T) {
	renderer := &TableRenderer{}
	jobs := []client.JobStatus{
		{Name: "short", Status: "Running", Ready: 1, Replicas: 1, Age: "1h"},
		{Name: "medium-length-job-name", Status: "Pending", Ready: 0, Replicas: 2, Age: "2h"},
	}
	output := renderer.RenderJobList(jobs)

	// Both job names should be fully visible
	assert.Contains(t, output, "short")
	assert.Contains(t, output, "medium-length-job-name")
}

func TestRenderJobListTruncation(t *testing.T) {
	renderer := &TableRenderer{}
	longName := "extremely-long-job-name-that-exceeds-fifty-characters-and-should-be-truncated"
	jobs := []client.JobStatus{
		{Name: longName, Status: "Running", Ready: 1, Replicas: 1, Age: "1h"},
	}
	output := renderer.RenderJobList(jobs)

	assert.Contains(t, output, "...")
	assert.NotContains(t, output, longName)
}

func TestRenderJobDetailPodsDynamicWidth(t *testing.T) {
	renderer := &TableRenderer{}
	info := &client.JobInfo{
		Status: client.JobStatus{
			Name: "test-job", Status: "Running", Replicas: 2, Ready: 1, Age: "1h",
		},
		Pods: []client.PodInfo{
			{Name: "pod-1", Status: "Running", IP: "10.0.0.1", Node: "node-1"},
			{Name: "pod-2-with-longer-name", Status: "Pending", IP: "10.0.0.2", Node: "node-2"},
		},
	}
	output := renderer.RenderJobDetail(info)

	assert.Contains(t, output, "pod-1")
	assert.Contains(t, output, "pod-2-with-longer-name")
}

func TestRenderJobListWide(t *testing.T) {
	jobs := []client.JobStatus{
		{Name: "job1", Namespace: "default", Status: "Running", Framework: "pytorch", GPURequested: 8, Replicas: 4, Ready: 4, Age: "5m"},
		{Name: "job2", Namespace: "ml-team", Status: "Succeeded", Framework: "tensorflow", GPURequested: 4, Replicas: 2, Ready: 2, Age: "1h"},
	}

	r := NewTableRenderer()
	output := r.RenderJobListWide(jobs)

	// Check headers
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "NAMESPACE")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "FRAMEWORK")
	assert.Contains(t, output, "GPU")
	assert.Contains(t, output, "REPLICAS")
	assert.Contains(t, output, "AGE")

	// Check data
	assert.Contains(t, output, "job1")
	assert.Contains(t, output, "default")
	assert.Contains(t, output, "pytorch")
	assert.Contains(t, output, "4/4")
	assert.Contains(t, output, "job2")
	assert.Contains(t, output, "ml-team")
	assert.Contains(t, output, "tensorflow")
	assert.Contains(t, output, "2/2")
}

func TestRenderJobListWideEmpty(t *testing.T) {
	r := NewTableRenderer()
	output := r.RenderJobListWide(nil)
	assert.Contains(t, output, "No jobs found")
}

func TestRenderJobListWideEmptySlice(t *testing.T) {
	r := NewTableRenderer()
	output := r.RenderJobListWide([]client.JobStatus{})
	assert.Contains(t, output, "No jobs found")
}

func TestRenderTopJob(t *testing.T) {
	jobs := []client.JobStatus{
		{Name: "job1", Status: "Running", GPURequested: 32, Replicas: 4, Ready: 4, Age: "5m"},
		{Name: "job2", Status: "Succeeded", GPURequested: 0, Replicas: 2, Ready: 2, Age: "1h"},
	}

	r := NewTableRenderer()
	output := r.RenderTopJob(jobs)

	// Check GPU_REQUESTED header
	assert.Contains(t, output, "GPU_REQUESTED")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "STATUS")
	assert.Contains(t, output, "REPLICAS")
	assert.Contains(t, output, "AGE")

	// Check data
	assert.Contains(t, output, "32")
	assert.Contains(t, output, "0")
	assert.Contains(t, output, "job1")
	assert.Contains(t, output, "4/4")
}

func TestRenderTopJobEmpty(t *testing.T) {
	r := NewTableRenderer()
	output := r.RenderTopJob(nil)
	assert.Contains(t, output, "No jobs found")
}

func TestRenderTopJobWide(t *testing.T) {
	jobs := []client.JobStatus{
		{Name: "job1", Namespace: "default", Status: "Running", Framework: "pytorch", GPURequested: 32, Replicas: 4, Ready: 4, Age: "5m"},
	}

	r := NewTableRenderer()
	output := r.RenderTopJobWide(jobs)

	// Check all headers present
	headers := []string{"NAME", "NAMESPACE", "STATUS", "FRAMEWORK", "GPU_REQUESTED", "REPLICAS", "AGE"}
	for _, h := range headers {
		assert.Contains(t, output, h)
	}

	// Check data
	assert.Contains(t, output, "job1")
	assert.Contains(t, output, "default")
	assert.Contains(t, output, "pytorch")
	assert.Contains(t, output, "32")
	assert.Contains(t, output, "4/4")
}

func TestRenderTopJobWideEmpty(t *testing.T) {
	r := NewTableRenderer()
	output := r.RenderTopJobWide(nil)
	assert.Contains(t, output, "No jobs found")
}

func TestRenderTopJobWideEmptySlice(t *testing.T) {
	r := NewTableRenderer()
	output := r.RenderTopJobWide([]client.JobStatus{})
	assert.Contains(t, output, "No jobs found")
}

func TestRenderJobDetailWithConfiguration(t *testing.T) {
	config := &task.Task{
		Name:  "test-job",
		Image: "pytorch:2.1",
		Run:   "python train.py",
		Framework: task.Framework{
			Name: "pytorch",
		},
		Worker: &task.Worker{
			Replicas: 2,
		},
	}

	info := &client.JobInfo{
		Status: client.JobStatus{
			Name:      "test-job",
			Namespace: "default",
			Status:    "Running",
			Replicas:  2,
			Ready:     2,
			Age:       "5m",
		},
		Pods: []client.PodInfo{
			{Name: "test-job-worker-0", Status: "Running", IP: "10.0.0.1", Node: "node-1"},
		},
		Configuration: config,
	}

	renderer := &TableRenderer{}
	output := renderer.RenderJobDetail(info)

	assert.Contains(t, output, "Configuration:")
	assert.Contains(t, output, "test-job")
	assert.Contains(t, output, "pytorch:2.1")
	assert.Contains(t, output, "python train.py")
	assert.Contains(t, output, "pytorch")

	// Every non-empty line in the Pods sub-table must be indented by 2 spaces.
	lines := strings.Split(output, "\n")
	podsStart := -1
	for i, line := range lines {
		if line == "Pods:" {
			podsStart = i + 1
			break
		}
	}
	require.GreaterOrEqual(t, podsStart, 0, "Pods: header should be present")
	for i := podsStart; i < len(lines); i++ {
		if lines[i] == "" || lines[i] == "Configuration:" {
			break
		}
		assert.True(t, strings.HasPrefix(lines[i], "  "),
			"Pods sub-table line %q should start with 2 spaces", lines[i])
	}

	// Every non-empty line in the Configuration YAML section must be indented by 2 spaces.
	configStart := -1
	for i, line := range lines {
		if line == "Configuration:" {
			configStart = i + 1
			break
		}
	}
	require.GreaterOrEqual(t, configStart, 0, "Configuration: header should be present")
	for i := configStart; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		assert.True(t, strings.HasPrefix(lines[i], "  "),
			"Configuration YAML line %q should start with 2 spaces", lines[i])
	}
}

func TestRenderJobDetailWithoutConfiguration(t *testing.T) {
	info := &client.JobInfo{
		Status: client.JobStatus{
			Name:      "test-job",
			Namespace: "default",
			Status:    "Running",
			Replicas:  1,
			Ready:     1,
			Age:       "1m",
		},
		Configuration: nil,
	}

	renderer := &TableRenderer{}
	output := renderer.RenderJobDetail(info)

	assert.NotContains(t, output, "Configuration:")
	assert.Contains(t, output, "test-job")
}

func TestIndentLines(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		prefix string
		want   string
	}{
		{
			name:   "single line",
			input:  "hello",
			prefix: "  ",
			want:   "  hello",
		},
		{
			name:   "multi line",
			input:  "a\nb\nc",
			prefix: "  ",
			want:   "  a\n  b\n  c",
		},
		{
			name:   "empty string",
			input:  "",
			prefix: "  ",
			want:   "",
		},
		{
			name:   "trailing newline",
			input:  "a\n",
			prefix: "  ",
			want:   "  a\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, indentLines(tt.input, tt.prefix))
		})
	}
}

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kubeflow/arena/pkg/client"
	"github.com/kubeflow/arena/pkg/output"
)

func TestOutputRendering(t *testing.T) {
	renderer := &output.TableRenderer{}

	t.Run("empty list", func(t *testing.T) {
		result := renderer.RenderJobList(nil)
		assert.Contains(t, result, "No jobs found")
	})

	t.Run("job list", func(t *testing.T) {
		jobs := []client.JobStatus{
			{Name: "job-1", Status: "Running", Replicas: 4, Ready: 3, Age: "5m"},
			{Name: "job-2", Status: "Succeeded", Replicas: 2, Ready: 2, Age: "1h"},
		}
		result := renderer.RenderJobList(jobs)
		assert.Contains(t, result, "job-1")
		assert.Contains(t, result, "job-2")
		assert.Contains(t, result, "Running")
		assert.Contains(t, result, "Succeeded")
		assert.Contains(t, result, "3/4")
		assert.Contains(t, result, "2/2")
	})

	t.Run("job detail", func(t *testing.T) {
		info := &client.JobInfo{
			Status: client.JobStatus{
				Name:      "detail-job",
				Namespace: "default",
				Status:    "Running",
				Replicas:  4,
				Ready:     2,
				Age:       "10m",
			},
			Pods: []client.PodInfo{
				{Name: "pod-0", Status: "Running", IP: "10.0.0.1", Node: "node-1"},
				{Name: "pod-1", Status: "Pending", IP: "", Node: ""},
			},
		}
		result := renderer.RenderJobDetail(info)
		assert.Contains(t, result, "detail-job")
		assert.Contains(t, result, "Running")
		assert.Contains(t, result, "pod-0")
		assert.Contains(t, result, "10.0.0.1")
	})
}

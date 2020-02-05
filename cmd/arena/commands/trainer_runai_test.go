package commands

import (
	appsv1 "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

var (
	NAMESPACE string = "default"
)

func TestJobInclusionInResourcesListCommand(t *testing.T) {
	jobName := "job-name"
	jobUUID := "id1"
	controller := true
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					UID:        types.UID(jobUUID),
					Kind:       string(ResourceTypeJob),
					Name:       jobName,
					Controller: &controller,
				},
			}},
		Spec: v1.PodSpec{
			SchedulerName: "runai-scheduler",
		},
	}

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      jobName,
			UID:       types.UID(jobUUID),
		},
	}

	objects := []runtime.Object{pod, job}
	client := fake.NewSimpleClientset(objects...)

	trainer := NewRunaiTrainer(client)
	jobs, _ := trainer.ListTrainingJobs(NAMESPACE)

	trainJob := jobs[0]
	resources := trainJob.Resources()

	if !testResourceIncluded(resources, jobName, ResourceTypeJob) {
		t.Errorf("Could not find related job in training job resources")
	}
}

func TestJobInclusionInResourcesGetCommand(t *testing.T) {
	jobName := "job-name"
	jobUUID := "id1"

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      jobName,
			UID:       types.UID(jobUUID),
		},
	}

	objects := []runtime.Object{job}
	client := fake.NewSimpleClientset(objects...)

	trainer := NewRunaiTrainer(client)
	trainJob, _ := trainer.GetTrainingJob(jobName, NAMESPACE)

	resources := trainJob.Resources()

	if !testResourceIncluded(resources, jobName, ResourceTypeJob) {
		t.Errorf("Could not find related job in training job resources")
	}
}

func TestStatefulSetInclusionInResourcesGetCommand(t *testing.T) {
	jobName := "job-name"
	jobUUID := "id1"

	labelSelector := make(map[string]string)

	job := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      jobName,
			UID:       types.UID(jobUUID),
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
		},
	}

	objects := []runtime.Object{job}
	client := fake.NewSimpleClientset(objects...)

	trainer := NewRunaiTrainer(client)
	trainJob, _ := trainer.GetTrainingJob(jobName, NAMESPACE)

	resources := trainJob.Resources()

	if !testResourceIncluded(resources, jobName, ResourceTypeStatefulSet) {
		t.Errorf("Could not find related job in training job resources")
	}
}

func TestReplicaSetInclusionInResourcesGetCommand(t *testing.T) {
	jobName := "job-name"
	jobUUID := "id1"

	labelSelector := make(map[string]string)

	job := &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      jobName,
			UID:       types.UID(jobUUID),
		},
		Spec: appsv1.ReplicaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
		},
	}

	objects := []runtime.Object{job}
	client := fake.NewSimpleClientset(objects...)

	trainer := NewRunaiTrainer(client)
	trainJob, _ := trainer.GetTrainingJob(jobName, NAMESPACE)

	resources := trainJob.Resources()

	if !testResourceIncluded(resources, jobName, ResourceTypeReplicaset) {
		t.Errorf("Could not find related job in training job resources")
	}
}

func testResourceIncluded(resources []Resource, name string, resourceType ResourceType) bool {
	for _, resource := range resources {
		if resource.ResourceType == resourceType && resource.Name == name {
			return true
		}
	}
	return false
}

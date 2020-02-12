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

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      jobName,
			UID:       types.UID(jobUUID),
		},
	}

	pod := createPodOwnedBy("pod", nil, jobUUID, string(ResourceTypeJob), jobName)

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

func TestIncludeMultiplePodsInReplicaset(t *testing.T) {
	jobName := "job-name"
	jobUUID := "id1"

	labelSelector := make(map[string]string)
	labelSelector["job-name"] = jobName

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

	pod1 := createPodOwnedBy("pod1", labelSelector, jobUUID, string(ResourceTypeJob), jobName)
	pod2 := createPodOwnedBy("pod2", labelSelector, jobUUID, string(ResourceTypeJob), jobName)

	objects := []runtime.Object{job, pod1, pod2}
	client := fake.NewSimpleClientset(objects...)

	trainer := NewRunaiTrainer(client)
	trainJob, _ := trainer.GetTrainingJob(jobName, NAMESPACE)

	if len(trainJob.AllPods()) != 2 {
		t.Errorf("Did not get all pod owned by job")
	}
}

func TestIncludeMultiplePodsInStatefulset(t *testing.T) {
	jobName := "job-name"
	jobUUID := "id1"

	labelSelector := make(map[string]string)
	labelSelector["job-name"] = jobName

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

	pod1 := createPodOwnedBy("pod1", labelSelector, jobUUID, string(ResourceTypeJob), jobName)
	pod2 := createPodOwnedBy("pod2", labelSelector, jobUUID, string(ResourceTypeJob), jobName)

	objects := []runtime.Object{job, pod1, pod2}
	client := fake.NewSimpleClientset(objects...)

	trainer := NewRunaiTrainer(client)
	trainJob, _ := trainer.GetTrainingJob(jobName, NAMESPACE)

	if len(trainJob.AllPods()) != 2 {
		t.Errorf("Did not get all pod owned by job")
	}
}

func TestIncludeMultiplePodsInJob(t *testing.T) {
	jobName := "job-name"
	jobUUID := "id1"

	labelSelector := make(map[string]string)
	labelSelector["job-name"] = jobName

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: NAMESPACE,
			Name:      jobName,
			UID:       types.UID(jobUUID),
		},
	}

	pod1 := createPodOwnedBy("pod1", labelSelector, jobUUID, string(ResourceTypeJob), jobName)
	pod2 := createPodOwnedBy("pod2", labelSelector, jobUUID, string(ResourceTypeJob), jobName)

	objects := []runtime.Object{job, pod1, pod2}
	client := fake.NewSimpleClientset(objects...)

	trainer := NewRunaiTrainer(client)
	trainJob, _ := trainer.GetTrainingJob(jobName, NAMESPACE)

	if len(trainJob.AllPods()) != 2 {
		t.Errorf("Did not get all pod owned by job")
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

func createPodOwnedBy(podName string, labelSelector map[string]string, ownerUUID string, ownerKind string, ownerName string) *v1.Pod {
	controller := true
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: NAMESPACE,
			Labels:    labelSelector,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					UID:        types.UID(ownerUUID),
					Kind:       ownerKind,
					Name:       ownerName,
					Controller: &controller,
				},
			}},
		Spec: v1.PodSpec{
			SchedulerName: SchedulerName,
		},
	}
}

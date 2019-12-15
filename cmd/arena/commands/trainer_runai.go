package commands

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type RunaiTrainer struct {
	client *kubernetes.Clientset
}

func NewRunaiTrainer(client *kubernetes.Clientset) Trainer {
	return &RunaiTrainer{
		client: client,
	}
}

func (rt *RunaiTrainer) IsSupported(name, ns string) bool {
	runaiJobList, err := rt.client.Batch().Jobs(ns).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s,app=runaijob", name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiJobList.Items) > 0 {
		return true
	}

	runaiStatefulSetsList, err := rt.client.Apps().StatefulSets(ns).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s,app=runaijob", name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiStatefulSetsList.Items) > 0 {
		return true
	}

	return false
}

func (rt *RunaiTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {

	runaiJobList, err := rt.client.Batch().Jobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s,app=runaijob", name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiJobList.Items) > 0 {
		return rt.getTrainingJob(runaiJobList.Items[0])
	}

	runaiStatufulsetList, err := rt.client.Apps().StatefulSets(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("release=%s,app=runaijob", name),
	})

	if len(runaiStatufulsetList.Items) > 0 {
		return rt.getTrainingStatefulset(runaiStatufulsetList.Items[0])
	}

	return nil, fmt.Errorf("Failed to find the job for %s", name)
}

func (rt *RunaiTrainer) Type() string {
	return defaultRunaiTrainingType
}

func (rt *RunaiTrainer) getTrainingStatefulset(statefulset appsv1.StatefulSet) (TrainingJob, error) {
	podList, err := rt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: fmt.Sprintf("release=%s,app=runaijob", statefulset.Name),
	})

	if err != nil {
		return nil, err
	}

	// Last created pod will be the chief pod
	pods := podList.Items
	return NewRunaiJob(pods, statefulset.CreationTimestamp, rt.Type(), statefulset.Name, false), nil
}

func (rt *RunaiTrainer) getTrainingJob(job batchv1.Job) (TrainingJob, error) {
	podList, err := rt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: fmt.Sprintf("job-name=%s", job.Name),
	})

	if err != nil {
		return nil, err
	}

	// Last created pod will be the chief pod
	pods := podList.Items
	return NewRunaiJob(pods, job.CreationTimestamp, rt.Type(), job.Name, true), nil
}

func (rt *RunaiTrainer) ListTrainingJobs() ([]TrainingJob, error) {
	runaiJobs := []TrainingJob{}

	runaiJobList, err := rt.client.Batch().Jobs(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", "runaijob"),
	})

	if err != nil {
		return nil, err
	}

	runaiStatefulSetList, err := rt.client.Apps().StatefulSets(namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", "runaijob"),
	})

	if err != nil {
		return nil, err
	}

	for _, item := range runaiJobList.Items {
		runaiJob, err := rt.getTrainingJob(item)

		if err != nil {
			return nil, err
		}

		runaiJobs = append(runaiJobs, runaiJob)
	}

	for _, item := range runaiStatefulSetList.Items {
		runaiStatefulSet, err := rt.getTrainingStatefulset(item)

		if err != nil {
			return nil, err
		}

		runaiJobs = append(runaiJobs, runaiStatefulSet)
	}

	return runaiJobs, nil
}

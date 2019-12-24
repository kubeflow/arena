package commands

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type RunaiTrainer struct {
	client *kubernetes.Clientset
}

func NewRunaiTrainer(client *kubernetes.Clientset) Trainer {
	return &RunaiTrainer{
		client: client,
	}
}

func fieldSelectorByName(name string) string {
	return fmt.Sprintf("metadata.name=%s", name)
}

func (rt *RunaiTrainer) IsSupported(name, ns string) bool {
	runaiJobList, err := rt.client.Batch().Jobs(ns).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiJobList.Items) > 0 {
		for _, item := range runaiJobList.Items {
			if item.Spec.Template.Spec.SchedulerName == "runai-scheduler" {
				return true
			}
		}
	}

	runaiStatefulSetsList, err := rt.client.Apps().StatefulSets(ns).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiStatefulSetsList.Items) > 0 {
		for _, item := range runaiStatefulSetsList.Items {
			if item.Spec.Template.Spec.SchedulerName == "runai-scheduler" {
				return true
			}
		}
	}

	runaiReplicaSetsList, err := rt.client.Apps().ReplicaSets(ns).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
	}

	if len(runaiReplicaSetsList.Items) > 0 {
		for _, item := range runaiReplicaSetsList.Items {
			if item.Spec.Template.Spec.SchedulerName == "runai-scheduler" {
				return true
			}
		}
	}

	return false
}

func (rt *RunaiTrainer) GetTrainingJob(name, namespace string) (TrainingJob, error) {

	runaiJobList, err := rt.client.Batch().Jobs(namespace).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiJobList.Items) > 0 {
		return rt.getTrainingJob(runaiJobList.Items[0])
	}

	runaiStatufulsetList, err := rt.client.Apps().StatefulSets(namespace).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiStatufulsetList.Items) > 0 {
		return rt.getTrainingStatefulset(runaiStatufulsetList.Items[0])
	}

	runaiReplicaSetsList, err := rt.client.Apps().ReplicaSets(namespace).List(metav1.ListOptions{
		FieldSelector: fieldSelectorByName(name),
	})

	if err != nil {
		log.Debugf("failed to search job %s in namespace %s due to %v", name, namespace, err)
	}

	if len(runaiReplicaSetsList.Items) > 0 {
		return rt.getTrainingReplicaSet(runaiReplicaSetsList.Items[0])
	}

	return nil, fmt.Errorf("Failed to find the job for %s", name)
}

func (rt *RunaiTrainer) Type() string {
	return defaultRunaiTrainingType
}

func (rt *RunaiTrainer) getTrainingReplicaSet(replicaSet appsv1.ReplicaSet) (TrainingJob, error) {
	labels := []string{}
	for key, value := range replicaSet.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", key, value))
	}

	podList, err := rt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: strings.Join(labels, ","),
	})

	if err != nil {
		return nil, err
	}

	// Last created pod will be the chief pod
	pods := podList.Items
	return NewRunaiJob(pods, replicaSet.CreationTimestamp, rt.Type(), replicaSet.Name, false, replicaSet.Labels["app"] == "runai"), nil
}

func (rt *RunaiTrainer) getTrainingStatefulset(statefulset appsv1.StatefulSet) (TrainingJob, error) {
	labels := []string{}
	for key, value := range statefulset.Spec.Selector.MatchLabels {
		labels = append(labels, fmt.Sprintf("%s=%s", key, value))
	}

	podList, err := rt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		},
		LabelSelector: strings.Join(labels, ","),
	})

	if err != nil {
		return nil, err
	}

	// Last created pod will be the chief pod
	pods := podList.Items
	return NewRunaiJob(pods, statefulset.CreationTimestamp, rt.Type(), statefulset.Name, false, statefulset.Labels["app"] == "runai"), nil
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
	return NewRunaiJob(pods, job.CreationTimestamp, rt.Type(), job.Name, true, job.Labels["app"] == "runai"), nil
}

func (rt *RunaiTrainer) ListTrainingJobs() ([]TrainingJob, error) {
	runaiJobs := []TrainingJob{}

	// Get all pods running with runai scheduler
	runaiPods, err := rt.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		FieldSelector: "spec.schedulerName=runai-scheduler",
	})

	if err != nil {
		return nil, err
	}

	jobPodMap := make(map[string]*RunaiJobInfo)

	// Group the pods by their controller
	for _, pod := range runaiPods.Items {
		controller := ""
		kind := ""

		for _, owner := range pod.OwnerReferences {
			if *owner.Controller {
				controller = owner.Name
				kind = owner.Kind
			}
		}

		if jobPodMap[controller] == nil {
			jobPodMap[controller] = &RunaiJobInfo{
				name: controller,
				pods: []v1.Pod{},
				kind: kind,
			}
		}

		// If controller exists for pod than add it to the map
		if controller != "" {
			jobPodMap[controller].pods = append(jobPodMap[controller].pods, pod)
		}
	}

	// Find more info on each of the controllers
	runaiJobList, err := rt.client.Batch().Jobs(namespace).List(metav1.ListOptions{})

	for _, job := range runaiJobList.Items {
		if jobPodMap[job.Name] != nil {
			jobPodMap[job.Name].creationTimestamp = job.CreationTimestamp

			if job.Labels["app"] == "runaijob" {
				jobPodMap[job.Name].createdByCLI = true
			}
		}
	}

	runaiStatefulSetsList, err := rt.client.Apps().StatefulSets(namespace).List(metav1.ListOptions{})

	for _, statefulSet := range runaiStatefulSetsList.Items {
		if jobPodMap[statefulSet.Name] != nil {
			jobPodMap[statefulSet.Name].creationTimestamp = statefulSet.CreationTimestamp
			
			if statefulSet.Labels["app"] == "runaijob" {
				jobPodMap[statefulSet.Name].createdByCLI = true
			}
		}
	}

	runaiReplicaSetsList, err := rt.client.Apps().ReplicaSets(namespace).List(metav1.ListOptions{})

	for _, replicaSet := range runaiReplicaSetsList.Items {
		if jobPodMap[replicaSet.Name] != nil {
			jobPodMap[replicaSet.Name].creationTimestamp = replicaSet.CreationTimestamp

			if replicaSet.Labels["app"] == "runaijob" {
				jobPodMap[replicaSet.Name].createdByCLI = true
			}
		}
	}

	for _, jobInfo := range jobPodMap {
		runaiJobs = append(runaiJobs, NewRunaiJobFromInfo(jobInfo))
	}

	return runaiJobs, nil
}

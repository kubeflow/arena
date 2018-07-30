package commands

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MPI Job Information
type MPIJob struct {
	*JobInfo
}

// Get the chief Pod of the Job.
func (mj *MPIJob) ChiefPod() v1.Pod {
	return mj.jobPod
}

// Get the name of the Training Job
// func (mj *MPIJob) Name() string {
// 	return
// }

// Get all the pods of the Training Job
func (mj *MPIJob) AllPods() []v1.Pod {
	return mj.pods
}

// Get Dashboard url of the job
func (mj *MPIJob) GetJobDashboards(client *kubernetes.Clientset) ([]string, error) {
	urls := []string{}
	dashboardURL, err := dashboard(client, arenaNamespace, "kubernetes-dashboard")

	if err != nil {
		log.Debugf("Get dashboard failed due to %v", err)
		// retry for the existing customers, will be deprecated in the future
		dashboardURL, err = dashboard(client, "kube-system", "kubernetes-dashboard")
		if err != nil {
			log.Debugf("Get dashboard failed due to %v", err)
		}
	}

	if dashboardURL == "" {
		return urls, fmt.Errorf("No LOGVIEWER Installed.")
	}

	spec := mj.jobPod.Spec
	job := mj.job
	url := fmt.Sprintf("%s/#!/log/%s/%s/%s?namespace=%s\n",
		dashboardURL,
		job.Namespace,
		mj.jobPod.Name,
		spec.Containers[0].Name,
		job.Namespace)

	urls = append(urls, url)

	return urls, nil
}

// Get the hostIP of the chief Pod
func (mj *MPIJob) HostIPOfChief() (hostIP string) {
	hostIP = "N/A"
	if mj.GetStatus() == "RUNNING" {
		hostIP = mj.jobPod.Status.HostIP
	}

	return hostIP
}

// MPI Job trainer
type MPIJobTrainer struct {
	client      *kubernetes.Clientset
	trainerType string
}

func NewMPIJobTrainer(client *kubernetes.Clientset) Trainer {
	log.Debugf("Init MPI job trainer")

	return &MPIJobTrainer{
		client:      client,
		trainerType: "mpijob",
	}
}

// check if it's mpi job
func (m *MPIJobTrainer) IsSupported(name, ns string) bool {
	isMPI := false

	if len(allJobs) > 0 {
		for _, job := range allJobs {
			if isMPIJob(name, ns, job) {
				isMPI = true
				log.Debugf("the job %s for %s in namespace %s is found.", job.Name, name, ns)
				break
			}
		}
	} else {
		jobList, err := m.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			}, LabelSelector: fmt.Sprintf("release=%s", name),
		})
		if err != nil {
			log.Debugf("failed to search job %s in namespace %s due to %v", name, ns, err)
		}

		if len(jobList.Items) > 0 {
			isMPI = true
		}
	}

	return isMPI
}

func (m *MPIJobTrainer) Type() string {
	return m.trainerType
}

func (m *MPIJobTrainer) GetTrainingJob(name, namespace string) (tj TrainingJob, err error) {
	if len(allPods) > 0 {
		tj, err = m.getTrainingJobFromCache(name, namespace)
	} else {
		tj, err = m.getTrainingJob(name, namespace)
	}

	return tj, err
}

func (m *MPIJobTrainer) getTrainingJob(name, namespace string) (TrainingJob, error) {
	var (
		jobPod v1.Pod
		job    batchv1.Job
		latest metav1.Time
	)

	// 1. Get the batchJob of trainig Job
	pods := []v1.Pod{}
	jobList, err := m.client.BatchV1().Jobs(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", name),
	})
	if err != nil {
		return nil, err
	}

	if len(jobList.Items) == 0 {
		return nil, fmt.Errorf("Failed to find the job for %s", name)
	} else {
		job = jobList.Items[0]
	}

	// 2. Find the pod list, and determine the pod of the job
	podList, err := m.client.CoreV1().Pods(namespace).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("release=%s", name),
	})

	if err != nil {
		return nil, err
	}

	for _, item := range podList.Items {
		meta := item.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" {
				isJob = true
				log.Debugf("find job pod %v, break", item)
				break
			}
		}

		if !isJob {
			pods = append(pods, item)
			log.Debugf("add pod %v to pods", item)
		} else {
			if jobPod.Name == "" {
				latest = item.CreationTimestamp
				jobPod = item
				log.Debugf("set pod %s as first jobpod, and it's time is %v", jobPod.Name, jobPod.CreationTimestamp)
			} else {
				log.Debugf("current jobpod %s , and it's time is %v", jobPod.Name, latest)
				log.Debugf("candidiate jobpod %s , and it's time is %v", item.Name, item.CreationTimestamp)
				current := item.CreationTimestamp
				if latest.Before(&current) {
					jobPod = item
					latest = current
					log.Debugf("replace")
				} else {
					log.Debugf("no replace")
				}
			}
		}
	}

	pods = append(pods, jobPod)

	return &MPIJob{
		JobInfo: &JobInfo{
			job:         job,
			jobPod:      jobPod,
			pods:        pods,
			name:        name,
			trainerType: m.Type(),
		},
	}, nil

}

// Get the training job from Cache
func (m *MPIJobTrainer) getTrainingJobFromCache(name, ns string) (TrainingJob, error) {

	var (
		jobPod v1.Pod
		job    batchv1.Job
		latest metav1.Time
	)

	pods := []v1.Pod{}

	// 1. Find the batch job
	for _, item := range allJobs {
		if isMPIJob(name, ns, item) {
			job = item
			break
		}
	}

	// 2. Find the pods, and determine the pod of the job
	for _, item := range allPods {

		if !isMPIPod(name, ns, item) {
			continue
		}

		meta := item.ObjectMeta
		isJob := false
		owners := meta.OwnerReferences
		for _, owner := range owners {
			if owner.Kind == "Job" {
				isJob = true
				log.Debugf("find job pod %v, break", item)
				break
			}
		}

		if !isJob {
			// for non-job pod, add it into the pod list
			pods = append(pods, item)
			log.Debugf("add pod %v to pods", item)
		} else {
			if jobPod.Name == "" {
				latest = item.CreationTimestamp
				jobPod = item
				log.Debugf("set pod %s as first jobpod, and it's time is %v", jobPod.Name, jobPod.CreationTimestamp)
			} else {
				log.Debugf("current jobpod %s , and it's time is %v", jobPod.Name, latest)
				log.Debugf("candidiate jobpod %s , and it's time is %v", item.Name, item.CreationTimestamp)
				current := item.CreationTimestamp
				if latest.Before(&current) {
					jobPod = item
					latest = current
					log.Debugf("replace")
				} else {
					log.Debugf("no replace")
				}
			}
		}
	}

	pods = append(pods, jobPod)

	return &MPIJob{
		JobInfo: &JobInfo{
			job:         job,
			jobPod:      jobPod,
			pods:        pods,
			name:        name,
			trainerType: m.Type(),
		},
	}, nil
}

func isMPIJob(name, ns string, item batchv1.Job) bool {

	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the job %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "tf-horovod") {
		log.Debugf("the job %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

func isMPIPod(name, ns string, item v1.Pod) bool {
	if val, ok := item.Labels["release"]; ok && (val == name) {
		log.Debugf("the pod %s with labels %s", item.Name, val)
	} else {
		return false
	}

	if val, ok := item.Labels["app"]; ok && (val == "tf-horovod") {
		log.Debugf("the pod %s with labels %s is found.", item.Name, val)
	} else {
		return false
	}

	if item.Namespace != ns {
		return false
	}
	return true
}

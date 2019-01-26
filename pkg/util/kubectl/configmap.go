package kubectl

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/kubeflow/arena/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const JOB_CONFIG_LABEL = "createdBy=arena"

/**
*
* list configMaps by using namespace
**/
func ListAppConfigMaps(clientset *kubernetes.Clientset, namespace string, trainingTypes []string) (jobs []types.TrainingJobInfo, err error) {

	jobs = []types.TrainingJobInfo{}
	cmList, err := clientset.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{LabelSelector: JOB_CONFIG_LABEL})
	if err != nil {
		return nil, err
	}

	for _, cm := range cmList.Items {
		found := false

		job := types.TrainingJobInfo{}

	innerLoop:
		for _, trainingType := range trainingTypes {
			// if item.n.HasSuffix(line, fmt.Sprintf("-%s", trainingType)) {
			// 	found = true
			// 	name.Name = strings.TrimSuffix(line, fmt.Sprintf("-%s", trainingType))
			// 	name.Type = trainingType
			// 	break innerLoop
			// }

			if strings.HasSuffix(cm.Name, fmt.Sprintf("-%s", trainingType)) {
				found = true
				job.Name = strings.TrimSuffix(cm.Name, fmt.Sprintf("-%s", trainingType))
				job.Type = trainingType
				job.Namespace = cm.Namespace
			}
		}

		if found {
			jobs = append(jobs, job)
		} else {
			log.Debugf("drop %s in training configmap", job)
		}

	}

	log.Debugf("the job training configmap: %q", jobs)

	return jobs, nil
}

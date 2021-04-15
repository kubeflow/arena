// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/kubeflow/arena/pkg/util"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	std_errors "errors"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
)

var (
	namespace       string
	jobName         string
	statefulSetName string

	ErrNoNeedUpgrade = std_errors.New("No need upgrade. It's already the latest version.")
	ErrStillNeedWait = std_errors.New("Need waited.")
)

// Receive Namespace, Job Name, Statefulset name
func main() {

	// 1. Get the job, statefulset and namespace
	err := updateInputFromEnv()
	if err != nil {
		panic(err.Error())
	}

	log.Infof("jobName: %s, namespace: %s, statefulset %s", jobName, namespace, statefulSetName)
	// 2. creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// 3. creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 4. Wait the job complete, only when job completes, clean up the statefulset
	err = WaitJobComplete(clientset, namespace, jobName, math.MaxInt64, 5*time.Second)
	if err != nil {
		panic(err.Error())
	}

	// 5. Delete the statefulset
	err = DeleteStatefulSet(clientset, namespace, statefulSetName)
	if err != nil {
		panic(err.Error())
	}

}

// get the job to monitor from environment
func updateInputFromEnv() error {
	namespace = os.Getenv("NAMESPACE")
	if len(namespace) == 0 {
		return fmt.Errorf("Failed to get namespace from env NAMESPACE")
	}

	jobName = os.Getenv("JOBNAME")
	if len(jobName) == 0 {
		return fmt.Errorf("Failed to get jobName from env JOBNAME")
	}

	statefulSetName = os.Getenv("STATEFULSETNAME")
	if len(statefulSetName) == 0 {
		return fmt.Errorf("Failed to get statefulsetName from env STATEFULSETNAME")
	}

	return nil
}

func WaitJobComplete(client *kubernetes.Clientset, namespace string, jobName string, duration time.Duration, tick time.Duration) error {

	log.Infof("Duration %v, tick %v", duration, tick)

	return util.RetryDuring(duration, tick, func() (err error) {
		job, err := client.BatchV1().Jobs(namespace).Get(context.TODO(), jobName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				log.Infof("Job %s doesn't exist, need to wait.", jobName)
				return ErrStillNeedWait
			}
		}

		output, err := yaml.Marshal(job)
		if err != nil {
			log.Warnf("Failed to parse the job %s due to %v", job.Name, err)
		}

		log.Infof("Check the job status %s", string(output))
		// wait the job to start
		startTime := job.Status.StartTime
		if startTime.IsZero() {
			return ErrStillNeedWait
		}

		succeed := false
		if job.Status.Succeeded > 0 {
			succeed = true
		}

		if !succeed {
			log.Warnf("Failed due to %v", job.Status.Conditions)
			return ErrStillNeedWait
		} else {
			return nil
		}

	})

}

func DeleteStatefulSet(client *kubernetes.Clientset, namespace string, stsName string) error {
	sts, err := client.AppsV1beta1().StatefulSets(namespace).Get(context.TODO(), stsName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("The statefulset %s in namespace %s is not found, it has been deleted.", stsName, namespace)
			return nil
		} else {
			return err
		}
	}

	// Delete statefulset when the job is completed.
	deletePolicy := metav1.DeletePropagationForeground
	svcName := sts.Spec.ServiceName
	err = client.CoreV1().Services(namespace).Delete(context.TODO(), svcName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("The svcName %s in namespace %s has been deleted.", svcName, namespace)
			return nil
		} else {
			return err
		}
	}

	err = client.AppsV1beta1().StatefulSets(namespace).Delete(context.TODO(), stsName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		return err
	}

	return util.RetryDuring(10*time.Minute, 10*time.Second, func() (err error) {
		sts, err = client.AppsV1beta1().StatefulSets(namespace).Get(context.TODO(), stsName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				log.Infof("The statefulset %s in namespace %s has been deleted.", stsName, namespace)
				return nil
			} else {
				log.Infof("Unexpected the err %v", err)
				return err
			}
		}
		log.Errorf("still can get statefulset: %s/%s, [%+v]\n", namespace, stsName, sts)
		return ErrStillNeedWait
	})

}

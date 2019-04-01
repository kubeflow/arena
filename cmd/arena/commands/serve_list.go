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

package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/helm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var serving_charts = map[string]string{
	"tensorflow-serving-0.2.0":  "Tensorflow",
	"tensorrt-inference-server-0.0.1": "TensorRT",
}

func NewServingListCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "list",
		Short: "list all the serving jobs",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)

			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			fmt.Fprintf(w, "NAME\tTYPE\tVERSION\tSTATUS\tCLUSTER-IP\n")
			jobs, err := ListServingJobs(client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, servingJob := range jobs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					servingJob.GetName(),
					servingJob.ServeType,
					servingJob.Version,
					servingJob.GetStatus(),
					servingJob.GetClusterIP(),
						)
			}

			_ = w.Flush()
		},
	}

	return command
}

func ListServingJobs(client *kubernetes.Clientset) ([]ServingJob, error) {
	jobs := []ServingJob{}
	ns := namespace
	if allNamespaces {
		ns = metav1.NamespaceAll
	}
	serviceNameLabel := "servingName"
	deployments, err := client.AppsV1().Deployments(ns).List(metav1.ListOptions{
		LabelSelector: serviceNameLabel,
	})
	if err != nil {
		log.Errorf("Failed due to %v", err)
		os.Exit(1)
	}

	allPods, err = acquireAllPods(client)
	if err != nil {
		log.Errorf("Failed to acquireAllPods due to %v", err)
		os.Exit(1)
	}

	services, err := client.CoreV1().Services(ns).List(metav1.ListOptions{
		LabelSelector: "servingName",
	})
	if err != nil {
		log.Errorf("Failed to list services due to %v", err)
		os.Exit(1)
	}
	allServices = services.Items
	log.Debugf("All Services: %++v", allServices)

	log.Debugf("Serving deployments Items is %++v", deployments.Items)
	for _, deploy := range deployments.Items {
		jobs = append(jobs, NewServingJob(deploy))
	}
	log.Debugf("Serving jobs list is %++v", jobs)
	return jobs, nil
}

func ListServingJobsByHelm() ([]ServingJob, error) {
	releaseMap, err := helm.ListAllReleasesWithDetail()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	servingJobs := []ServingJob{}
	for name, cols := range releaseMap {
		log.Debugf("name: %s, cols: %s", name, cols)
		namespace := cols[len(cols)-1]
		chart := cols[len(cols)-2]
		status := cols[len(cols)-3]
		log.Debugf("namespace: %s, chart: %s, status:%s", namespace, chart, status)
		if serveType, ok := serving_charts[chart]; ok {
			index := strings.Index(name, "-")
			//serviceName := name[0:index]
			serviceVersion := ""
			if index > -1 {
				serviceVersion = name[index+1:]
			}
			nameAndVersion := strings.Split(name, "-")
			log.Debugf("nameAndVersion: %s, len(nameAndVersion): %d", nameAndVersion, len(nameAndVersion))
			servingJobs = append(servingJobs, ServingJob{
				Name: nameAndVersion[0],
				Namespace: namespace,
				Version: serviceVersion,
				ServeType: serveType,
				//Status: status,
			})
		}
	}
	return servingJobs, nil
}

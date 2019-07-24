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

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/util/helm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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
			jobs, err := ListServing(client)
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

// ListServing returns a list of serving
func ListServing(client *kubernetes.Clientset) ([]types.Serving, error) {
	jobs := []types.Serving{}
	ns := GetNamespace()
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

	log.Debugf("Serving deployments Items is %++v", deployments.Items)
	for _, deploy := range deployments.Items {
		jobs = append(jobs, types.NewServingJob(client, deploy, allPods))
	}
	log.Debugf("Serving jobs list is %++v", jobs)
	return jobs, nil
}

// List Servings by name
func ListServingsByName(client *kubernetes.Clientset, name string) (servings []types.Serving, err error) {
	ns := GetNamespace()
	labels := fmt.Sprintf("servingName=%s", name)
	deployList, err := client.AppsV1().Deployments(ns).List(metav1.ListOptions{
		LabelSelector: labels,
	})
	if err != nil {
		log.Debugf("Failed to ListServingsByName due to %v", err)
		return nil, err
	}

	servings = []types.Serving{}
	for _, deploy := range deployList.Items {
		servingType := deploy.Labels["servingType"]
		servingVersion := deploy.Labels["servingVersion"]
		// servingName := deploy.Labels["servingName"]
		servings = append(servings, types.Serving{
			Name:      name,
			ServeType: servingType,
			Version:   servingVersion,
		})
	}
	return servings, nil
}

func ListServingJobsByHelm() ([]types.Serving, error) {
	releaseMap, err := helm.ListAllReleasesWithDetail()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	servings := []types.Serving{}
	for name, cols := range releaseMap {
		log.Debugf("name: %s, cols: %s", name, cols)
		namespace := cols[len(cols)-1]
		chart := cols[len(cols)-2]
		status := cols[len(cols)-3]
		log.Debugf("namespace: %s, chart: %s, status:%s", namespace, chart, status)
		if serveType, ok := types.SERVING_CHARTS[chart]; ok {
			index := strings.Index(name, "-")
			//serviceName := name[0:index]
			servingVersion := ""
			if index > -1 {
				servingVersion = name[index+1:]
			}
			nameAndVersion := strings.Split(name, "-")
			log.Debugf("nameAndVersion: %s, len(nameAndVersion): %d", nameAndVersion, len(nameAndVersion))
			servings = append(servings, types.Serving{
				Name:      nameAndVersion[0],
				Namespace: namespace,
				Version:   servingVersion,
				ServeType: serveType,
				//Status: status,
			})
		}
	}
	return servings, nil
}

func GetNamespace() string {
	ns := namespace
	if allNamespaces {
		ns = metav1.NamespaceAll
	}
	return ns
}

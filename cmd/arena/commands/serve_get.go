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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"

	//"io"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// define serving type
type ServingType string

// three serving types
const (
	ServingTF     ServingType = "TENSORFLOW"
	ServingTRT    ServingType = "TENSORRT"
	ServingCustom ServingType = "CUSTOM"
)

var (
	// get format from command option
	printFormat string
	// format template for "wide"
	tablePrintTemplate = `Query Result %d:
%s
NAME:             %s
NAMESPACE:        %s	
VERSION:          %s
STATUS:           %s
DESIRED:          %d
AVAILABLE:        %d
SERVING TYPE:     %s
ENDPOINT ADDRESS: %s
ENDPOINT PORTS:   %s
SERVING DURATION: %s
%s
%s
%s
`

	// table header
	subtableHeader = "INSTANCE\tSTATUS\tAGE\tREADY\tRESTARTS\tNODE"
	// how many space equal a table?
	oneTableEqualManySpace = 2
	extraSpaces            = 8 * oneTableEqualManySpace
	printAllJobs           bool
	// define a map for serving type
	SERVINGTYPE = map[string]ServingType{
		"tf-serving":     ServingTF,
		"trt-serving":    ServingTRT,
		"custom-serving": ServingCustom,
	}
	// get serving type from command option
	stype           string
	ErrNotFoundJobs = errors.New(`not found jobs under the assigned conditions.`)
)

// ServingJobPrint defines the print format
type ServingJobPrint struct {
	Name            string         `yaml:"name" json:"name"`
	Namespace       string         `yaml:"namespace" json:"namespace"`
	Version         string         `yaml:"version" json:"version"`
	Status          string         `yaml:"status" json:"status"`
	Desired         int32          `yaml:"desired" json:"desired"`
	Available       int32          `yaml:"available" json:"available"`
	ServingDuration string         `yaml:"serving_duration" json:"serving_duration"`
	ServingType     string         `yaml:"serving_type" json:"serving_type"`
	EndpointAddress string         `yaml:"endpoint_address" json:"endpoint_address"`
	EndpointPorts   string         `yaml:"endpoint_ports" json:"endpoint_ports"`
	Pods            []PodPrintInfo `yaml:"instances" json:"instances"`
}
type PodPrintInfo struct {
	PodName string `yaml:"pod_name" json:"pod_name"` // selfLink
	// create timestamp
	CreationTimestamp metav1.Time `yaml:"-" json:"-"`
	// how long the pod is running
	Age string `yaml:"age" json:"age"`
	// pod' status,there is "Running" and "Pending"
	Status v1.PodPhase `yaml:"status" json:"status"`
	// the node ip
	HostIP       string `yaml:"host_ip" json:"host_ip"`
	Ready        string `yaml:"ready" json:"ready"`
	RestartCount string `yaml:"restart_count" json:"restart_count"`
}

// NewServingGetCommand starts the command
func NewServingGetCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "get ServingJobName",
		Short: "display details of a serving job",
		Run: func(cmd *cobra.Command, args []string) {
			// no serving name is an error
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			// set loglevel
			util.SetLogLevel(logLevel)
			// initate kubenetes client
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				log.Errorf(err.Error())
				os.Exit(1)
			}
			servingName := args[0]
			ServingGetExecute(client, servingName)
		},
	}
	//command.Flags().BoolVar(&allNamespaces, "all-namespaces", false, "all namespace")
	command.Flags().StringVar(&servingVersion, "version", "", "assign the serving job version")
	command.Flags().BoolVar(&printAllJobs, "all", false, "display all jobs whose name is matched the one user gives.")
	command.Flags().StringVar(&printFormat, "format", "wide", `set the print format,format can be "yaml" or "json"`)
	command.Flags().StringVar(&stype, "type", "", `assign the serving job type,type can be "tf"("tensorflow"),"trt"("tensorrt"),"custom"`)

	return command

}

// check the serving type and transfer it
func checkServingTypeIsOk() error {
	var st = map[string]ServingType{
		"trt":    ServingTRT,
		"tf":     ServingTF,
		"custom": ServingCustom,
	}
	if stype == "" {
		return nil
	}
	if _, ok := st[stype]; !ok {
		return fmt.Errorf("unknow serving type: %s", stype)
	}
	stype = string(st[stype])
	return nil
}

// entry function for "serve get"
func ServingGetExecute(client *kubernetes.Clientset, servingName string) {
	// check some conditions are ok
	if err := PrepareCheck(); err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	// get all jobs,the jobs' name are the same,but version or type maybe not.
	servingJobs, err := GetServingJobsByName(client, servingName)
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	// filter all jobs and get the job which meets the criteria
	filterJobs, err := FilterServingJobs(servingJobs, servingVersion, stype)
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	// create the print format
	servingJobsForPrinting, err := NewServingJobsPrintList(filterJobs)
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	// if jobs count is large than 1,print the help information
	if len(servingJobsForPrinting) > 1 && !printAllJobs {
		printString := GetMulJobsHelpInfo(servingJobsForPrinting)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, oneTableEqualManySpace, ' ', 0)
		fmt.Fprintf(w, printString)
		os.Exit(1)
	}
	printInfoToBytes, err := FormatServingJobs(printFormat, servingJobsForPrinting)
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	if printFormat == "wide" {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, oneTableEqualManySpace, ' ', 0)
		fmt.Fprintf(w, string(printInfoToBytes))
	} else {
		fmt.Println(string(printInfoToBytes))
	}
}

// if there is some jobs match the conditons given by user,print the all jobs and make user to chose one job.
func GetMulJobsHelpInfo(jobs []ServingJobPrint) string {
	header := fmt.Sprintf("%d jobs have been found:", len(jobs))
	tableHeader := "NAME\tTYPE\tVERSION"
	printLines := []string{tableHeader}
	footer := fmt.Sprintf("please use \"--type\" or \"--version\" to filter,or use \"--all\" to display all jobs.")
	for _, job := range jobs {
		line := fmt.Sprintf("%s\t%s\t%s",
			job.Name,
			job.ServingType,
			job.Version,
		)
		printLines = append(printLines, line)
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(printLines, "\n"), footer)
}

// make some checks firstly
func PrepareCheck() error {
	err := checkServingTypeIsOk()
	if err != nil {
		return err
	}
	return nil
}

// get all jobs whose name meet the  given one.
func GetServingJobsByName(client *kubernetes.Clientset, servingName string) ([]types.Serving, error) {
	jobs := []types.Serving{}
	ns := GetNamespace()
	deployments, err := client.AppsV1().Deployments(ns).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("serviceName=%s", servingName),
	})
	if err != nil {
		log.Errorf("Failed due to %v", err)
		os.Exit(1)
	}
	podListObject, err := client.CoreV1().Pods(GetNamespace()).List(metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ListOptions",
			APIVersion: "v1",
		}, LabelSelector: fmt.Sprintf("serviceName=%s", servingName),
	})
	if err != nil {
		log.Errorf("Failed to get pods by label serviceName=%s,reason=%s", servingName, err.Error())
		os.Exit(1)
	}

	for _, deploy := range deployments.Items {
		jobs = append(jobs, types.NewServingJob(client, deploy, podListObject.Items))
	}

	log.Debugf("Serving jobs list is %++v", jobs)
	if len(jobs) == 0 {
		return nil, ErrNotFoundJobs
	}
	return jobs, nil
}

// check namespace,version,type is matched the given one or not
func servingJobIsMatched(target string, matchType string, servingJob types.Serving) bool {
	switch {
	case target == "":
		return true
	case matchType == "NAMESPACE" && target == servingJob.Namespace:
		return true
	case matchType == "VERSION" && target == servingJob.Version:
		return true
	case matchType == "TYPE" && target == servingJob.ServeType:
		return true
	default:
		return false
	}
}

// FilterServingJobs filters serving jobs under some conditions
func FilterServingJobs(jobs []types.Serving, servingVersion, servingType string) ([]types.Serving, error) {
	filterJobs := []types.Serving{}
	for _, job := range jobs {
		namespaceIsMatched := servingJobIsMatched(namespace, "NAMESPACE", job)
		versionIsMatched := servingJobIsMatched(servingVersion, "VERSION", job)
		typeIsMatched := servingJobIsMatched(servingType, "TYPE", job)
		log.Debugf("name: %v,namespaceIsOk: %v,versionIsOk: %v,typeIsMatched: %v", job.Name, namespaceIsMatched, versionIsMatched, typeIsMatched)
		if namespaceIsMatched && versionIsMatched && typeIsMatched {
			filterJobs = append(filterJobs, job)
		}
	}
	if len(filterJobs) == 0 {
		return nil, ErrNotFoundJobs
	}
	return filterJobs, nil
}

// create a list to store the jobs with printing format
func NewServingJobsPrintList(jobs []types.Serving) ([]ServingJobPrint, error) {
	servingJobPrintList := []ServingJobPrint{}
	for _, job := range jobs {
		podPrintList := []PodPrintInfo{}
		for _, pod := range job.AllPods() {
			var hostIP, age, restartCount, ready string
			switch pod.Status.Phase {
			case v1.PodPending:
				age = "N/A"
				hostIP = "N/A"
				restartCount = "N/A"
				ready = "N/A"
			default:
				hostIP = pod.Status.HostIP
				age = util.ShortHumanDuration(time.Now().Sub(pod.ObjectMeta.CreationTimestamp.Time))
			}
			if restartCount != "N/A" {
				container, err := GetContainerStatus(pod.Status.ContainerStatuses, pod.ObjectMeta.Labels["app"])
				if err != nil {
					return nil, err
				}
				restartCount = fmt.Sprintf("%d", container.RestartCount)
			}
			if ready != "N/A" {
				ready = GetReadyContainerCount(pod.Status.ContainerStatuses)
			}
			podPrintInfo := PodPrintInfo{
				PodName:           path.Base(pod.ObjectMeta.SelfLink),
				CreationTimestamp: pod.ObjectMeta.CreationTimestamp,
				Age:               age,
				Status:            pod.Status.Phase,
				HostIP:            hostIP,
				RestartCount:      restartCount,
				Ready:             ready,
			}
			podPrintList = append(podPrintList, podPrintInfo)
		}
		jobPrintObj := ServingJobPrint{
			Name:            job.Name,
			Namespace:       job.Namespace,
			Status:          job.GetStatus(),
			Desired:         job.DesiredInstances(),
			Available:       job.AvailableInstances(),
			EndpointAddress: job.GetClusterIP(),
			EndpointPorts:   job.GetPorts(),
			ServingDuration: job.GetAge(),
			Version:         job.Version,
			ServingType:     job.ServeType,
			Pods:            podPrintList,
		}
		servingJobPrintList = append(servingJobPrintList, jobPrintObj)
	}
	return servingJobPrintList, nil

}

// get count of containers which status is ready in a pod
func GetReadyContainerCount(containers []v1.ContainerStatus) string {
	if len(containers) == 0 {
		return "0/0"
	}
	readyCount := 0
	for _, con := range containers {
		if con.Ready {
			readyCount++
		}
	}
	return fmt.Sprintf("%v/%v", readyCount, len(containers))
}

// get the container status
func GetContainerStatus(containers []v1.ContainerStatus, containerName string) (v1.ContainerStatus, error) {
	for _, con := range containers {
		if con.Name == containerName {
			return con, nil
		}
		log.Debugf("container name: %s,target name: %s\n", con.Name, containerName)
	}
	return v1.ContainerStatus{}, fmt.Errorf("not found container status whose name is %s", containerName)
}

// matchTargetCondition checks serving job is matching the namespace or version

// format the serving jobs information
func FormatServingJobs(format string, servingJobPrintInfoList []ServingJobPrint) ([]byte, error) {
	switch format {
	case "json":
		return json.Marshal(servingJobPrintInfoList)
	case "yaml":
		return yaml.Marshal(servingJobPrintInfoList)
	default:
		return []byte(customFormat(servingJobPrintInfoList)), nil
	}
}

// if format type is "wide",define our printable string
// 	subtableHeader = "INSTANCE\tSTATUS\tAGE\tRESTARTS\tNODE"

func customFormat(servingJobList []ServingJobPrint) string {
	var printInfoList = []string{}
	for index, job := range servingJobList {
		podInfoStringArray := []string{subtableHeader}
		maxStringLen := len(subtableHeader) + extraSpaces
		for _, pod := range job.Pods {
			podInfoStringLine := fmt.Sprintf("%s\t%v\t%s\t%s\t%s\t%s",
				pod.PodName,
				pod.Status,
				pod.Age,
				pod.Ready,
				pod.RestartCount,
				pod.HostIP,
			)
			if len(podInfoStringLine)+extraSpaces > maxStringLen {
				maxStringLen = len(podInfoStringLine) + extraSpaces
			}
			podInfoStringArray = append(podInfoStringArray, podInfoStringLine)
		}
		singleJobPrintInfo := fmt.Sprintf(
			tablePrintTemplate,
			index,
			JoinCharToLine("=", maxStringLen),
			job.Name,
			job.Namespace,
			job.Version,
			job.Status,
			job.Desired,
			job.Available,
			job.ServingType,
			job.EndpointAddress,
			job.EndpointPorts,
			job.ServingDuration,
			JoinCharToLine("-", maxStringLen),
			strings.Join(podInfoStringArray, "\n"),
			JoinCharToLine("=", maxStringLen),
		)
		printInfoList = append(printInfoList, singleJobPrintInfo)

	}
	return strings.Join(printInfoList, "\n")
}

// get a string with expect length
func JoinCharToLine(ch string, length int) string {
	if length == 0 {
		return ""
	}
	var line = []string{}
	for i := 0; i < length; i++ {
		line = append(line, ch)
	}
	return strings.Join(line, "")

}

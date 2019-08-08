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
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/printer"

	"github.com/kubeflow/arena/pkg/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	//"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	// get format from command option
	printFormat string
	// format template for "wide"
	tablePrintTemplate = `NAME:             %s
NAMESPACE:        %s	
VERSION:          %s
DESIRED:          %d
AVAILABLE:        %d
SERVING TYPE:     %s
ENDPOINT ADDRESS: %s
ENDPOINT PORTS:   %s
AGE:              %s

%s
`
	// table header
	subtableHeader = "INSTANCE\tSTATUS\tAGE\tREADY\tRESTARTS\tNODE"
	// how many space equal a table?
	oneTableEqualManySpace = 2
	// get serving type from command option
	stype           string
	ErrNotFoundJobs = errors.New(`not found jobs under the assigned conditions.`)
	ErrTooManyJobs  = errors.New(`found jobs more than one,please use --version or --type to filter.`)
)

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
	command.Flags().StringVar(&printFormat, "format", "wide", `set the print format,format can be "yaml" or "json"`)
	command.Flags().StringVar(&stype, "type", "", `assign the serving job type,type can be "tf"("tensorflow"),"trt"("tensorrt"),"custom"`)

	return command

}

// check the serving type and transfer it
func checkServingTypeIsOk() error {
	if stype == "" {
		return nil
	}
	if types.KeyMapServingType(stype) == types.ServingType("") {
		return fmt.Errorf("unknow serving type: %s", stype)
	}
	stype = string(types.KeyMapServingType(stype))
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
	// if jobs count is large than 1,print the help information
	if len(filterJobs) > 1 {
		printString := GetMulJobsHelpInfo(filterJobs)
		w := tabwriter.NewWriter(os.Stdout, 0, 0, oneTableEqualManySpace, ' ', 0)
		fmt.Fprintf(w, printString)
		os.Exit(1)
	}
	printInfoToBytes, err := FormatServingJobs(printFormat, filterJobs[0])
	if err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	if printFormat == "wide" {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, oneTableEqualManySpace, ' ', 0)
		fmt.Fprintf(w, string(printInfoToBytes))
		w.Flush()
	} else {
		fmt.Println(string(printInfoToBytes))
	}
}

// if there is some jobs match the conditons given by user,print the all jobs and make user to chose one job.
func GetMulJobsHelpInfo(jobs []printer.ServingJobPrinter) string {
	header := fmt.Sprintf("There is %d jobs have been found:", len(jobs))
	tableHeader := "NAME\tTYPE\tVERSION"
	printLines := []string{tableHeader}
	footer := fmt.Sprintf("please use \"--type\" or \"--version\" to filter.")
	for _, job := range jobs {
		line := fmt.Sprintf("%s\t%s\t%s",
			job.GetName(),
			job.GetType(),
			job.GetVersion(),
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
func GetServingJobsByName(client *kubernetes.Clientset, servingName string) ([]printer.ServingJobPrinter, error) {
	jobs := []printer.ServingJobPrinter{}
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
		jobs = append(jobs, printer.NewServingJobPrinter(client, deploy, podListObject.Items))
	}

	log.Debugf("Serving jobs list is %++v", jobs)
	if len(jobs) == 0 {
		return nil, ErrNotFoundJobs
	}
	return jobs, nil
}

// FilterServingJobs filters serving jobs under some conditions
func FilterServingJobs(jobs []printer.ServingJobPrinter, servingVersion, servingType string) ([]printer.ServingJobPrinter, error) {
	filterJobs := []printer.ServingJobPrinter{}
	for _, job := range jobs {
		namespaceIsMatched := job.IsMatchedGivenNamespace(namespace)
		versionIsMatched := job.IsMatchedGivenVersion(servingVersion)
		typeIsMatched := job.IsMatchedGivenType(servingType)
		log.Debugf("name: %v,namespaceIsOk: %v,versionIsOk: %v,typeIsMatched: %v", job.GetName(), namespaceIsMatched, versionIsMatched, typeIsMatched)
		if namespaceIsMatched && versionIsMatched && typeIsMatched {
			filterJobs = append(filterJobs, job)
		}
	}
	if len(filterJobs) == 0 {
		return nil, ErrNotFoundJobs
	}
	return filterJobs, nil
}

// format the serving jobs information
func FormatServingJobs(format string, job printer.ServingJobPrinter) ([]byte, error) {
	switch format {
	case "json":
		return job.GetJson()
	case "yaml":
		return job.GetYaml()
	default:
		return []byte(customFormat(job)), nil
	}
}

// if format type is "wide",define our printable string
// 	subtableHeader = "INSTANCE\tSTATUS\tAGE\tRESTARTS\tNODE"

func customFormat(job printer.ServingJobPrinter) string {
	podInfoStringArray := []string{subtableHeader}
	for _, pod := range job.Pods {
		podInfoStringLine := fmt.Sprintf("%s\t%v\t%s\t%s\t%s\t%s",
			pod.PodName,
			pod.Status,
			pod.Age,
			pod.Ready,
			pod.RestartCount,
			pod.HostIP,
		)
		podInfoStringArray = append(podInfoStringArray, podInfoStringLine)
	}
	jobPrintString := fmt.Sprintf(
		tablePrintTemplate,
		job.GetName(),
		job.GetNamespace(),
		job.GetVersion(),
		job.Desired,
		job.Available,
		job.GetType(),
		job.EndpointAddress,
		job.EndpointPorts,
		job.Age,
		strings.Join(podInfoStringArray, "\n"),
	)
	return jobPrintString
}

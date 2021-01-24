package serving

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v2"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	log "github.com/sirupsen/logrus"
)

var (
	errNotFoundServingJobMessage = "Not found serving job %v, please check it with `arena serve list | grep %v`"
)

var getJobTemplate = `
Name:       %v
Namespace:  %v
Type:       %v
Version:    %v
Desired:    %v
Available:  %v
Age:        %v
Address:    %v
Port:       %v
%v
`

func SearchServingJob(namespace, name, version string, servingType types.ServingJobType) (ServingJob, error) {
	if servingType == types.UnknownServingJob {
		return nil, fmt.Errorf("Unknown serving job type,arena only supports: [%s]", utils.GetSupportServingJobTypesInfo())
	}
	processers := GetAllProcesser()
	if servingType != types.AllServingJob {
		processer, ok := processers[servingType]
		if !ok {
			return nil, fmt.Errorf("unknown processer %v,please define it", servingType)
		}
		servingJobs, err := processer.GetServingJobs(namespace, name, version)
		if err != nil {
			return nil, err
		}
		if len(servingJobs) == 0 {
			return nil, fmt.Errorf(errNotFoundServingJobMessage, name, name)
		}
		if len(servingJobs) > 1 {
			return nil, fmt.Errorf("%v", moreThanOneJobHelpInfo(servingJobs))
		}
		return servingJobs[0], nil
	}
	jobs := []ServingJob{}
	for _, p := range processers {
		servingJobs, err := p.GetServingJobs(namespace, name, version)
		if err != nil {
			log.Debugf("processer %v does not support the serving job %v", p.Type(), name)
			continue
		}
		jobs = append(jobs, servingJobs...)
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf(errNotFoundServingJobMessage, name, name)
	}
	if len(jobs) > 1 {
		return nil, fmt.Errorf("%v", moreThanOneJobHelpInfo(jobs))
	}
	return jobs[0], nil
}

func PrintServingJob(job ServingJob, format types.FormatStyle) {
	switch format {
	case types.JsonFormat:
		data, _ := json.MarshalIndent(job.Convert2JobInfo(), "", "    ")
		fmt.Printf("%v", string(data))
		return
	case types.YamlFormat:
		data, _ := yaml.Marshal(job.Convert2JobInfo())
		fmt.Printf("%v", string(data))
		return
	}
	jobInfo := job.Convert2JobInfo()
	endpointAddress := jobInfo.IPAddress
	ports := []string{}
	for _, e := range jobInfo.Endpoints {
		port := ""
		if e.NodePort != 0 {
			port = fmt.Sprintf("%v:%v->%v", strings.ToUpper(e.Name), e.NodePort, e.Port)
		} else {
			port = fmt.Sprintf("%v:%v", strings.ToUpper(e.Name), e.Port)
		}
		ports = append(ports, port)
	}
	isGPUShare := false
	title := "GPUs"
	step := "----"
	gpuAllocation := fmt.Sprintf("GPUs:       %v", jobInfo.RequestGPU)
	if jobInfo.RequestGPUMemory != 0 {
		isGPUShare = true
		title = "GPU_MEMORY"
		step = "----------"
		gpuAllocation = fmt.Sprintf("GPU Memory: %vGiB", jobInfo.RequestGPUMemory)
	}
	lines := []string{gpuAllocation, "", "Instances:", fmt.Sprintf("  NAME\tSTATUS\tAGE\tREADY\tRESTARTS\t%v\tNODE", title)}
	lines = append(lines, fmt.Sprintf("  ----\t------\t---\t-----\t--------\t%v\t----", step))
	for _, i := range jobInfo.Instances {
		value := fmt.Sprintf("%v", i.RequestGPU)
		if isGPUShare {
			value = fmt.Sprintf("%vGiB", i.RequestGPUMemory)
		}
		items := []string{
			fmt.Sprintf("  %v", i.Name),
			fmt.Sprintf("%v", i.Status),
			fmt.Sprintf("%v", i.Age),
			fmt.Sprintf("%v/%v", i.ReadyContainer, i.TotalContainer),
			fmt.Sprintf("%v", i.RestartCount),
			fmt.Sprintf("%v", value),
			fmt.Sprintf("%v", i.NodeName),
		}
		lines = append(lines, strings.Join(items, "\t"))
	}
	lines = append(lines, "")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	output := fmt.Sprintf(strings.Trim(getJobTemplate, "\n"),
		jobInfo.Name,
		jobInfo.Namespace,
		jobInfo.Type,
		jobInfo.Version,
		jobInfo.Desired,
		jobInfo.Available,
		jobInfo.Age,
		endpointAddress,
		strings.Join(ports, ","),
		strings.Join(lines, "\n"),
	)
	fmt.Fprintf(w, output)
	w.Flush()
}

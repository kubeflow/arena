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
			log.Debugf("processer %v does not support the serving job %v", name)
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
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	jobInfo := job.Convert2JobInfo()
	fields := []string{
		fmt.Sprintf(`NAME:             %v`, jobInfo.Name),
		fmt.Sprintf(`NAMESPACE:        %v`, jobInfo.Namespace),
		fmt.Sprintf(`VERSION:          %v`, jobInfo.Version),
		fmt.Sprintf(`SERVING TYPE:     %v`, jobInfo.Type),
		fmt.Sprintf(`DESIRED:          %v`, jobInfo.Desired),
		fmt.Sprintf(`AVAILABLE:        %v`, jobInfo.Available),
		fmt.Sprintf(`AGE:              %v`, jobInfo.Age),
	}
	if jobInfo.RequestGPU != 0 {
		fields = append(fields, fmt.Sprintf("GPUS:             %v", jobInfo.RequestGPU))
	}
	if jobInfo.RequestGPUMemory != 0 {
		fields = append(fields, fmt.Sprintf("GPU MEMORY:       %v", jobInfo.RequestGPUMemory))
	}
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
	fields = append(fields, fmt.Sprintf(`ENDPOINT ADDRESS: %v`, endpointAddress))
	fields = append(fields, fmt.Sprintf(`ENDPOINT PORT:    %v`, strings.Join(ports, ",")))
	fields = append(fields, "")
	fields = append(fields, "INSTANCE\tSTATUS\tAGE\tREADY\tRESTARTS\tNODE")
	for _, i := range jobInfo.Instances {
		items := []string{
			fmt.Sprintf("%v", i.Name),
			fmt.Sprintf("%v", i.Status),
			fmt.Sprintf("%v", i.Age),
			fmt.Sprintf("%v/%v", i.ReadyContainer, i.TotalContainer),
			fmt.Sprintf("%v", i.RestartCount),
			fmt.Sprintf("%v", i.NodeName),
		}
		fields = append(fields, strings.Join(items, "\t"))
	}
	fmt.Fprintf(w, strings.Join(fields, "\n"))
	w.Flush()
	fmt.Println()
}

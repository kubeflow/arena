package serving

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	errUnknownServingJobType = fmt.Errorf("unknown serving types,only support: %v", utils.GetServingJobTypes())
)

func ListServingJobs(namespace string, allNamespace bool, servingType types.ServingJobType) ([]ServingJob, error) {
	if servingType == types.UnknownServingJob {
		return nil, fmt.Errorf("Unknown serving job type,arena only supports: [%s]", utils.GetSupportServingJobTypesInfo())
	}
	processers := GetAllProcesser()
	if servingType != types.AllServingJob {
		processer, ok := processers[servingType]
		if !ok {
			return nil, fmt.Errorf("unknown processer %v,please define it", servingType)
		}
		return processer.ListServingJobs(namespace, allNamespace)
	}
	servingJobs := []ServingJob{}
	for _, p := range processers {
		jobs, err := p.ListServingJobs(namespace, allNamespace)
		if err != nil {
			log.Debugf("failed to get serving jobs whose type are %v", p.Type())
			continue
		}
		servingJobs = append(servingJobs, jobs...)
	}
	return servingJobs, nil
}

func DisplayAllServingJobs(jobs []ServingJob, allNamespace bool, format types.FormatStyle) {
	jobInfos := []types.ServingJobInfo{}
	for _, job := range jobs {
		jobInfos = append(jobInfos, job.Convert2JobInfo())
	}
	switch format {
	case types.JsonFormat:
		data, _ := json.MarshalIndent(jobInfos, "", "    ")
		fmt.Printf("%v", string(data))
		return
	case types.YamlFormat:
		data, _ := yaml.Marshal(jobInfos)
		fmt.Printf("%v", string(data))
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := []string{}
	if allNamespace {
		header = append(header, "NAMESPACE")
	}
	fields := []string{"NAME", "TYPE", "VERSION", "DESIRED", "AVAILABLE", "ADDRESS", "PORTS"}
	header = append(header, fields...)
	PrintLine(w, header...)
	for _, jobInfo := range jobInfos {
		line := []string{}
		if allNamespace {
			line = append(line, jobInfo.Namespace)
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
		if len(ports) == 0 {
			ports = append(ports, "N/A")
		}
		items := []string{
			jobInfo.Name,
			fmt.Sprintf("%v", jobInfo.Type),
			jobInfo.Version,
			fmt.Sprintf("%v", jobInfo.Desired),
			fmt.Sprintf("%v", jobInfo.Available),
			endpointAddress,
			strings.Join(ports, ","),
		}
		line = append(line, items...)
		PrintLine(w, line...)
	}
	_ = w.Flush()
}

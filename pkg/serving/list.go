package serving

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
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
	var wg sync.WaitGroup
	locker := new(sync.RWMutex)
	noPrivileges := false
	for _, pr := range processers {
		wg.Add(1)
		p := pr
		go func() {
			defer wg.Done()
			jobs, err := p.ListServingJobs(namespace, allNamespace)
			if err != nil {
				if strings.Contains(err.Error(), "forbidden: User") {
					log.Debugf("the user has no privileges to get the serving job %v,reason: %v", p.Type(), err)
					noPrivileges = true
					return
				}
				log.Debugf("failed to get serving jobs whose type are %v", p.Type())
				return
			}
			locker.Lock()
			servingJobs = append(servingJobs, jobs...)
			locker.Unlock()
		}()
	}
	wg.Wait()
	if noPrivileges {
		item := fmt.Sprintf("namespace %v", namespace)
		if allNamespace {
			item = "all namespaces"
		}
		return nil, fmt.Errorf("the user has no privileges to list the serving jobs in %v", item)
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
	gpus := float64(0)
	for _, jobinfo := range jobInfos {
		for _, instance := range jobinfo.Instances {
			gpus += instance.RequestGPUs
		}
	}
	fields := []string{"NAME", "TYPE", "VERSION", "DESIRED", "AVAILABLE", "ADDRESS", "PORTS"}
	header = append(header, fields...)
	if gpus != float64(0) {
		header = append(header, "GPU")
	}
	PrintLine(w, header...)
	jobInfosMap := addTrafficWeight(jobInfos)
	for _, jobInfo := range jobInfos {
		jobInfo, ok := jobInfosMap[genServingJobKey(jobInfo)]
		if !ok {
			log.Debugf("not found job in job map,skip to print it")
			continue
		}
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
		if gpus != float64(0) {
			jobGPUs := float64(0)
			for _, instance := range jobInfo.Instances {
				jobGPUs += instance.RequestGPUs
			}
			items = append(items, fmt.Sprintf("%v", jobGPUs))

		}
		line = append(line, items...)
		PrintLine(w, line...)
	}
	_ = w.Flush()
}

type ServingJobGroup struct {
	Id        string
	Namespace string
	JobType   types.ServingJobType
	JobName   string
}

func addTrafficWeight(allJobInfos []types.ServingJobInfo) map[string]types.ServingJobInfo {
	type ServingJobGroup struct {
		id        string
		namespace string
		jobType   types.ServingJobType
		jobName   string
		ip        string
		endpoints []types.Endpoint
		items     []types.ServingJobInfo
	}
	servingJobsGroup := map[string]*ServingJobGroup{}
	servingJobMap := map[string]types.ServingJobInfo{}
	for _, jobInfo := range allJobInfos {
		key := genServingJobGroupKey(jobInfo)
		if servingJobsGroup[key] == nil {
			servingJobsGroup[key] = &ServingJobGroup{
				id:        key,
				namespace: jobInfo.Namespace,
				jobName:   jobInfo.Name,
				jobType:   types.ServingJobType(jobInfo.Type),
				items:     []types.ServingJobInfo{},
			}
		}
		servingJobMap[genServingJobKey(jobInfo)] = jobInfo
		if jobInfo.IPAddress != "" && jobInfo.IPAddress != "N/A" {
			servingJobsGroup[key].ip = jobInfo.IPAddress
		}
		if jobInfo.Endpoints != nil && len(jobInfo.Endpoints) != 0 {
			servingJobsGroup[key].endpoints = jobInfo.Endpoints
		}
		servingJobsGroup[key].items = append(servingJobsGroup[key].items, jobInfo)
	}
	if len(servingJobsGroup) == len(allJobInfos) {
		return servingJobMap
	}
	istioClient, err := initIstioClient()
	if err != nil {
		log.Debugf("failed to get istio client when querying traffic weight,reason: %v", err)
		return servingJobMap
	}
	for key, group := range servingJobsGroup {
		if len(group.items) == 1 {
			continue
		}
		weights, err := getVirtualServiceWeight(istioClient, group.namespace, group.jobName)
		if err != nil {
			log.Debugf("failed to get virtual service weight,reason: %v", err)
			continue
		}
		// if the weight is 0,fix it with 100
		if len(weights) == 1 {
			for version, weight := range weights {
				if weight == int32(0) {
					weights[version] = 100
				}
			}
		}
		jobHasBeenChanged := map[string]bool{}
		for version, weight := range weights {
			completeName := fmt.Sprintf("%v/%v", key, version)
			target := servingJobMap[completeName]
			switch weight {
			case int32(0):
				target.IPAddress = "N/A"
				target.Endpoints = []types.Endpoint{}
			case int32(100):
				target.IPAddress = fmt.Sprintf("%v", target.IPAddress)
				target.Endpoints = group.endpoints
			default:
				target.IPAddress = fmt.Sprintf("%v(%v%%)", target.IPAddress, weight)
				target.Endpoints = group.endpoints
			}
			jobHasBeenChanged[completeName] = true
			servingJobMap[completeName] = target
		}
		for _, item := range group.items {
			jobKey := genServingJobKey(item)
			if jobHasBeenChanged[jobKey] {
				continue
			}
			t := servingJobMap[jobKey]
			t.IPAddress = "N/A"
			t.Endpoints = []types.Endpoint{}
			servingJobMap[jobKey] = t
		}
	}
	return servingJobMap
}

func genServingJobKey(jobInfo types.ServingJobInfo) string {
	return fmt.Sprintf("%v/%v", genServingJobGroupKey(jobInfo), jobInfo.Version)
}

func genServingJobGroupKey(jobInfo types.ServingJobInfo) string {
	return fmt.Sprintf("%v/%v/%v", jobInfo.Namespace, jobInfo.Type, jobInfo.Name)
}

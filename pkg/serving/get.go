package serving

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/apis/utils"
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
		if err := validateJobs(servingJobs, name); err != nil {
			return nil, err
		}
		return servingJobs[0], nil
	}
	jobs := []ServingJob{}
	var wg sync.WaitGroup
	locker := new(sync.RWMutex)
	noPrivileges := false
	for _, pr := range processers {
		wg.Add(1)
		p := pr
		go func() {
			defer wg.Done()
			servingJobs, err := p.GetServingJobs(namespace, name, version)
			if err != nil {
				if strings.Contains(err.Error(), "forbidden: User") {
					log.Debugf("the user has no privileges to get the serving job %v,reason: %v", p.Type(), err)
					noPrivileges = true
					return
				}
				log.Debugf("processer %v does not support the serving job %v", p.Type(), name)
				return
			}
			locker.Lock()
			jobs = append(jobs, servingJobs...)
			locker.Unlock()
		}()
	}
	wg.Wait()
	if noPrivileges {
		return nil, fmt.Errorf("the user has no privileges to get the serving job in namespace %v", namespace)
	}
	if err := validateJobs(jobs, name); err != nil {
		return nil, err
	}
	return jobs[0], nil
}

func validateJobs(jobs []ServingJob, name string) error {
	if len(jobs) == 0 {
		return fmt.Errorf(errNotFoundServingJobMessage, name, name)
	}
	knownJobs := []ServingJob{}
	unknownJobs := []ServingJob{}
	for _, s := range jobs {
		var labels map[string]string
		if ksjob, ok := s.(*kserveJob); ok {
			labels = ksjob.inferenceService.Labels
		} else {
			labels = s.Deployment().Labels
		}
		if CheckJobIsOwnedByProcesser(labels) {
			knownJobs = append(knownJobs, s)
		} else {
			unknownJobs = append(unknownJobs, s)
		}
	}
	log.Debugf("total known jobs: %v,total unknown jobs: %v", len(knownJobs), len(unknownJobs))
	if len(knownJobs) > 1 {
		return fmt.Errorf("%v", moreThanOneJobHelpInfo(jobs))
	}
	if len(unknownJobs) > 0 {
		return types.ErrNoPrivilegesToOperateJob
	}
	return nil
}

func PrintServingJob(job ServingJob, mv *types.ModelVersion, format types.FormatStyle) {
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
	totalGPUs := float64(0)
	for _, i := range jobInfo.Instances {
		totalGPUs += i.RequestGPUs
	}
	title := ""
	step := ""
	gpuLine := ""
	var lines []string
	if totalGPUs != 0 {
		title = "\tGPU"
		step = "\t---"
		gpuLine = fmt.Sprintf("GPU:        %v", totalGPUs)
		lines = append(lines, gpuLine)
	}

	if job.Type() == types.KServeJob {
		if ksjob, ok := job.(*kserveJob); ok {
			lines = append(lines, "")
			for _, traffic := range ksjob.inferenceService.Status.Components["predictor"].Traffic {
				revision := traffic.RevisionName
				percent := *traffic.Percent
				if traffic.Tag == "prev" {
					lines = append(lines, fmt.Sprintf("PrevRevision:       %v", revision))
					lines = append(lines, fmt.Sprintf("PrevPrecent:        %v", percent))
				} else {
					lines = append(lines, fmt.Sprintf("LatestRevision:     %v", revision))
					lines = append(lines, fmt.Sprintf("LatestPrecent:      %v", percent))
				}
			}
		}
	}

	lines = append(lines, "", "Instances:", fmt.Sprintf("  NAME\tSTATUS\tAGE\tREADY\tRESTARTS%v\tNODE", title))
	lines = append(lines, fmt.Sprintf("  ----\t------\t---\t-----\t--------%v\t----", step))
	for _, i := range jobInfo.Instances {
		value := fmt.Sprintf("%v", i.RequestGPUs)
		items := []string{
			fmt.Sprintf("  %v", i.Name),
			fmt.Sprintf("%v", i.Status),
			fmt.Sprintf("%v", i.Age),
			fmt.Sprintf("%v/%v", i.ReadyContainer, i.TotalContainer),
			fmt.Sprintf("%v", i.RestartCount),
		}
		if totalGPUs != 0 {
			items = append(items, value)
		}
		items = append(items, i.NodeName)
		lines = append(lines, strings.Join(items, "\t"))
	}
	lines = append(lines, "")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Name:\t%v\n", jobInfo.Name)
	fmt.Fprintf(w, "Namespace:\t%v\n", jobInfo.Namespace)
	fmt.Fprintf(w, "Type:\t%v\n", jobInfo.Type)
	fmt.Fprintf(w, "Version:\t%v\n", jobInfo.Version)
	fmt.Fprintf(w, "Desired:\t%v\n", jobInfo.Desired)
	fmt.Fprintf(w, "Available:\t%v\n", jobInfo.Available)
	fmt.Fprintf(w, "Age:\t%v\n", jobInfo.Age)
	fmt.Fprintf(w, "Address:\t%v\n", endpointAddress)
	fmt.Fprintf(w, "Port:\t%v\n", strings.Join(ports, ","))
	if mv != nil {
		if mv.Name != "" {
			fmt.Fprintf(w, "ModelName:\t%v\n", mv.Name)
		}
		if mv.Version != "" {
			fmt.Fprintf(w, "ModelVersion:\t%v\n", mv.Version)
		}
		if mv.Source != "" {
			fmt.Fprintf(w, "ModelSource:\t%v\n", mv.Source)
		}
	}
	fmt.Fprintf(w, "%v\n", strings.Join(lines, "\n"))
	w.Flush()
}

package serving

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/k8saccesser"
	"github.com/kubeflow/arena/pkg/operators/kserve-operator/client/clientset/versioned"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/kubeflow/arena/pkg/workflow"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// TensorflowServingProcesser use the default processer
type KFServingProcesser struct {
	kserveclient *versioned.Clientset
	*processer
}

type kServeJob struct {
	*servingJob
}

func NewKFServingProcesser() Processer {
	p := &processer{
		processerType:   types.KFServingJob,
		client:          config.GetArenaConfiger().GetClientSet(),
		enable:          true,
		useIstioGateway: true,
	}
	return &KFServingProcesser{
		processer: p,
		kserveclient: versioned.NewForConfigOrDie(config.GetArenaConfiger().GetRestConfig()),
	}
}

func (p *KFServingProcesser) GetServingJobs(namespace, name, version string) ([]ServingJob, error) {
	//todo:refator this method to get real job item
	selector := []string{
		fmt.Sprintf("%v=%v", servingNameLabelKey, name),
		fmt.Sprintf("%v=%v", servingTypeLabelKey, p.processerType),
	}
	if version != "" {
		selector = append(selector, fmt.Sprintf("%v=%v", servingVersionLabelKey, version))
	}
	return p.FilterServingJobs(namespace, false, strings.Join(selector, ","))
}

func (p *KFServingProcesser) ListServingJobs(namespace string, allNamespace bool) ([]ServingJob, error) {
	selector := ""
	arenaConfiger := config.GetArenaConfiger()
	if arenaConfiger.IsIsolateUserInNamespace() {
		selector = fmt.Sprintf("%v,%v=%v", selector, types.UserNameIdLabel, arenaConfiger.GetUser().GetId())
	}
	return p.FilterServingJobs(namespace, allNamespace, selector)
}

func (s *kServeJob) Convert2JobInfo() types.ServingJobInfo {
	servingType := types.ServingTypeMap[s.servingType].Alias
	servingJobInfo := types.ServingJobInfo{
		UUID:              s.name,
		Name:              s.name,
		Namespace:         s.namespace,
		Version:           s.version,
		Type:              servingType,
	}
	return servingJobInfo
}

func (p *KFServingProcesser) FilterServingJobs(namespace string, allNamespace bool, label string) ([]ServingJob, error) {
	if allNamespace {
		namespace = metav1.NamespaceAll
	}

	isvcList, err := k8saccesser.GetK8sResourceAccesser().ListInferenceService(p.kserveclient,namespace, label)
	if err != nil {
		return nil, err
	}

	istioGatewayServices := p.getIstioGatewayService()
	servingJobs := []ServingJob{}

	for _, isvc := range isvcList {
		job := &servingJob{
			name:          isvc.Name,
			namespace:     isvc.Namespace,
			servingType:   p.processerType,
			version:       isvc.ResourceVersion,
			istioServices: istioGatewayServices,
		}

		servingJobs = append(servingJobs, &kServeJob{
			servingJob: job,
		})
	}
	return servingJobs, nil
}

func SubmitKFServingJob(namespace string, args *types.KFServingArgs) (err error) {
	log.Infof("The Job %s has been submitted successfully", args.Name)
	nameWithVersion := fmt.Sprintf("%v-%v", args.Name, args.Version)
	args.Namespace = namespace
	processers := GetAllProcesser()
	processer, ok := processers[args.Type]
	if !ok {
		return fmt.Errorf("not found processer whose type is %v", args.Type)
	}
	jobs, err := processer.GetServingJobs(args.Namespace, args.Name, args.Version)
	if err != nil {
		return err
	}
	if err := ValidateJobsBeforeSubmiting(jobs, args.Name); err != nil {
		return err
	}
	// the master is also considered as a worker
	chart := util.GetChartsFolder() + "/kfserving"
	err = workflow.SubmitJob(nameWithVersion, string(types.KFServingJob), namespace, args, chart)
	if err != nil {
		return err
	}
	log.Infof("The Job %s has been submitted successfully", args.Name)
	log.Infof("You can run `arena get %s --type %s` to check the job status", args.Name, args.Type)
	return nil
}
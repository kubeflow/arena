package k8saccesser

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"

	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	v1alpha12 "github.com/kubeflow/arena/pkg/operators/et-operator/api/v1alpha1"
	etversioned "github.com/kubeflow/arena/pkg/operators/et-operator/client/clientset/versioned"
	cron_v1alpha1 "github.com/kubeflow/arena/pkg/operators/kubedl-operator/apis/apps/v1alpha1"
	cronversioned "github.com/kubeflow/arena/pkg/operators/kubedl-operator/client/clientset/versioned"
	"github.com/kubeflow/arena/pkg/operators/mpi-operator/apis/kubeflow/v1alpha1"
	mpiversioned "github.com/kubeflow/arena/pkg/operators/mpi-operator/client/clientset/versioned"
	pytorch_v1 "github.com/kubeflow/arena/pkg/operators/pytorch-operator/apis/pytorch/v1"
	pyversioned "github.com/kubeflow/arena/pkg/operators/pytorch-operator/client/clientset/versioned"
	spark_v1beta2 "github.com/kubeflow/arena/pkg/operators/spark-operator/apis/sparkoperator.k8s.io/v1beta2"
	sparkversioned "github.com/kubeflow/arena/pkg/operators/spark-operator/client/clientset/versioned"
	tfv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/tensorflow/v1"
	tfversioned "github.com/kubeflow/arena/pkg/operators/tf-operator/client/clientset/versioned"
	volcano_v1alpha1 "github.com/kubeflow/arena/pkg/operators/volcano-operator/apis/batch/v1alpha1"
	volcanovesioned "github.com/kubeflow/arena/pkg/operators/volcano-operator/client/clientset/versioned"
)

var accesser *k8sResourceAccesser
var once sync.Once

func init() {
	tfv1.AddToScheme(scheme.Scheme)
	v1alpha1.AddToScheme(scheme.Scheme)
	v1alpha12.AddToScheme(scheme.Scheme)
	pytorch_v1.AddToScheme(scheme.Scheme)
	spark_v1beta2.AddToScheme(scheme.Scheme)
	volcano_v1alpha1.AddToScheme(scheme.Scheme)
	cron_v1alpha1.AddToScheme(scheme.Scheme)
}

func InitK8sResourceAccesser(config *rest.Config, clientset *kubernetes.Clientset, isDaemonMode bool) error {
	var err error
	once.Do(func() {
		accesser, err = NewK8sResourceAccesser(config, clientset, isDaemonMode)
		if err == nil {
			err = accesser.Run()
		}
	})
	return err
}

func GetK8sResourceAccesser() *k8sResourceAccesser {
	return accesser
}

type k8sResourceAccesser struct {
	cacheClient  cache.Cache
	clientset    *kubernetes.Clientset
	cacheEnabled bool
}

func NewK8sResourceAccesser(config *rest.Config, clientset *kubernetes.Clientset, isDaemonMode bool) (*k8sResourceAccesser, error) {
	var cacheClient cache.Cache
	var err error
	if isDaemonMode {
		mapper, err := apiutil.NewDynamicRESTMapper(config)
		if err != nil {
			log.Errorf("failed to create cacheClient mapper, reason: %v", err)
			// if create dynamic mapper failed, use default restMapper
			mapper = nil
		}
		cacheClient, err = cache.New(config, cache.Options{Mapper: mapper, Resync: nil, Namespace: ""})
		if err != nil {
			log.Errorf("failed to create cacheClient, reason: %v", err)
			return nil, err
		}
		cacheClient.IndexField(context.TODO(), &v1.Pod{}, "spec.nodeName", func(o runtime.Object) []string {
			if pod, ok := o.(*v1.Pod); ok {
				return []string{pod.Spec.NodeName}
			}
			return []string{}
		})
	}
	return &k8sResourceAccesser{
		cacheClient:  cacheClient,
		clientset:    clientset,
		cacheEnabled: isDaemonMode,
	}, err
}

func (k *k8sResourceAccesser) Run() (err error) {
	if !k.cacheEnabled {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err = k.cacheClient.Start(context.Background().Done()); err != nil {
			cancel()
			log.Errorf("failed to start cacheClient, reason: %v", err)
		}
	}()
	k.cacheClient.WaitForCacheSync(ctx.Done())
	return err
}

func (k *k8sResourceAccesser) GetCacheClient() cache.Cache {
	return k.cacheClient
}

func (k *k8sResourceAccesser) ListPods(namespace string, filterLabels string, filterFields string, filterFunc func(*v1.Pod) bool) ([]*v1.Pod, error) {
	pods := []*v1.Pod{}
	podList := &v1.PodList{}
	labelSelector, err := parseLabelSelector(filterLabels)
	if err != nil {
		return nil, err
	}
	fieldSelector, err := parseFieldSelector(filterFields)
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			podList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			},
		)
	} else {
		podList, err = k.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: labelSelector.String(),
			FieldSelector: fieldSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		copyPod := pod.DeepCopy()
		if filterFunc != nil && !filterFunc(copyPod) {
			continue
		}
		pods = append(pods, copyPod)
	}
	return pods, nil
}

func (k *k8sResourceAccesser) ListStatefulSets(namespace string, filterLabels string) ([]*appv1.StatefulSet, error) {
	statefulsets := []*appv1.StatefulSet{}
	stsList := &appv1.StatefulSetList{}
	labelSelector, err := parseLabelSelector(filterLabels)
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			stsList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		stsList, err = k.clientset.AppsV1().StatefulSets(namespace).List(context.TODO(), metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, sts := range stsList.Items {
		statefulsets = append(statefulsets, sts.DeepCopy())
	}
	return statefulsets, nil
}

func (k *k8sResourceAccesser) ListDeployments(namespace string, filterLabels string) ([]*appv1.Deployment, error) {
	deployments := []*appv1.Deployment{}
	deployList := &appv1.DeploymentList{}
	labelSelector, err := parseLabelSelector(filterLabels)
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			deployList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		deployList, err = k.clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, d := range deployList.Items {
		deployments = append(deployments, d.DeepCopy())
	}
	return deployments, nil
}

func (k *k8sResourceAccesser) ListBatchJobs(namespace string, filterLabels string) ([]*batchv1.Job, error) {
	jobs := []*batchv1.Job{}
	jobList := &batchv1.JobList{}
	labelSelector, err := parseLabelSelector(filterLabels)
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			jobList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		jobList, err = k.clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, j := range jobList.Items {
		jobs = append(jobs, j.DeepCopy())
	}
	return jobs, nil
}

func (k *k8sResourceAccesser) ListServices(namespace string, filterLabels string) ([]*v1.Service, error) {
	services := []*v1.Service{}
	serviceList := &v1.ServiceList{}
	labelSelector, err := parseLabelSelector(filterLabels)
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			serviceList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		serviceList, err = k.clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, svc := range serviceList.Items {
		services = append(services, svc.DeepCopy())
	}
	return services, nil
}

func (k *k8sResourceAccesser) ListConfigMaps(namespace string, filterLabels string) ([]*v1.ConfigMap, error) {
	configmaps := []*v1.ConfigMap{}
	configmapList := &v1.ConfigMapList{}
	labelSelector, err := parseLabelSelector(filterLabels)
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			configmapList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		configmapList, err = k.clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, configmap := range configmapList.Items {
		configmaps = append(configmaps, configmap.DeepCopy())
	}
	return configmaps, nil
}

func (k *k8sResourceAccesser) ListNodes(filterLabels string) ([]*v1.Node, error) {
	nodeList := &v1.NodeList{}
	nodes := []*v1.Node{}
	labelSelector, err := parseLabelSelector(filterLabels)
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			nodeList,
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		nodeList, err = k.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ListOptions",
				APIVersion: "v1",
			},
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, node := range nodeList.Items {
		nodes = append(nodes, node.DeepCopy())
	}
	return nodes, nil
}

func (k *k8sResourceAccesser) ListCrons(cronClient *cronversioned.Clientset, namespace string) ([]*cron_v1alpha1.Cron, error) {
	var crons []*cron_v1alpha1.Cron
	cronList := &cron_v1alpha1.CronList{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			cronList,
			client.InNamespace(namespace),
			&client.ListOptions{},
		)
	} else {
		cronList, err = cronClient.AppsV1alpha1().Crons(namespace).List(context.Background(), metav1.ListOptions{})
	}

	if err != nil {
		return nil, err
	}
	for _, cron := range cronList.Items {
		crons = append(crons, cron.DeepCopy())
	}
	return crons, nil
}

func (k *k8sResourceAccesser) ListTensorflowJobs(tfjobClient *tfversioned.Clientset, namespace string) ([]*tfv1.TFJob, error) {
	jobs := []*tfv1.TFJob{}
	jobList := &tfv1.TFJobList{}
	var err error
	labelSelector, err := parseLabelSelector(fmt.Sprintf("app=%v,release", types.TFTrainingJob))
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			jobList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			},
		)

	} else {
		jobList, err = tfjobClient.KubeflowV1().TFJobs(namespace).List(metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, job := range jobList.Items {
		jobs = append(jobs, job.DeepCopy())
	}
	return jobs, nil
}

func (k *k8sResourceAccesser) ListMPIJobs(mpijobClient *mpiversioned.Clientset, namespace string) ([]*v1alpha1.MPIJob, error) {
	jobs := []*v1alpha1.MPIJob{}
	jobList := &v1alpha1.MPIJobList{}
	var err error
	labelSelector, err := parseLabelSelector(fmt.Sprintf("app=%v,release", types.MPITrainingJob))
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			jobList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		jobList, err = mpijobClient.KubeflowV1alpha1().MPIJobs(namespace).List(metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, job := range jobList.Items {
		jobs = append(jobs, job.DeepCopy())
	}
	return jobs, nil
}

func (k *k8sResourceAccesser) ListPytorchJobs(pytorchjobClient *pyversioned.Clientset, namespace string) ([]*pytorch_v1.PyTorchJob, error) {
	jobs := []*pytorch_v1.PyTorchJob{}
	jobList := &pytorch_v1.PyTorchJobList{}
	var err error
	labelSelector, err := parseLabelSelector(fmt.Sprintf("app=%v,release", types.PytorchTrainingJob))
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			jobList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		jobList, err = pytorchjobClient.KubeflowV1().PyTorchJobs(namespace).List(metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, job := range jobList.Items {
		jobs = append(jobs, job.DeepCopy())
	}
	return jobs, nil
}

func (k *k8sResourceAccesser) ListETJobs(etjobClient *etversioned.Clientset, namespace string) ([]*v1alpha12.TrainingJob, error) {
	jobs := []*v1alpha12.TrainingJob{}
	jobList := &v1alpha12.TrainingJobList{}
	var err error
	labelSelector, err := parseLabelSelector(fmt.Sprintf("app=%v,release", types.ETTrainingJob))
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			jobList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		jobList, err = etjobClient.EtV1alpha1().TrainingJobs(namespace).List(metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, job := range jobList.Items {
		jobs = append(jobs, job.DeepCopy())
	}
	return jobs, nil
}

func (k *k8sResourceAccesser) ListVolcanoJobs(volcanojobClient *volcanovesioned.Clientset, namespace string) ([]*volcano_v1alpha1.Job, error) {
	jobs := []*volcano_v1alpha1.Job{}
	jobList := &volcano_v1alpha1.JobList{}
	var err error
	labelSelector, err := parseLabelSelector(fmt.Sprintf("app=%v,release", types.VolcanoTrainingJob))
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			jobList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		jobList, err = volcanojobClient.BatchV1alpha1().Jobs(namespace).List(metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, job := range jobList.Items {
		jobs = append(jobs, job.DeepCopy())
	}
	return jobs, nil
}

func (k *k8sResourceAccesser) ListSparkJobs(sparkjobClient *sparkversioned.Clientset, namespace string) ([]*spark_v1beta2.SparkApplication, error) {
	jobs := []*spark_v1beta2.SparkApplication{}
	jobList := &spark_v1beta2.SparkApplicationList{}
	var err error
	labelSelector, err := parseLabelSelector(fmt.Sprintf("app=%v,release", types.SparkTrainingJob))
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			context.Background(),
			jobList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			})
	} else {
		jobList, err = sparkjobClient.SparkoperatorV1beta2().SparkApplications(namespace).List(metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		})
	}
	if err != nil {
		return nil, err
	}
	for _, job := range jobList.Items {
		jobs = append(jobs, job.DeepCopy())
	}
	return jobs, nil
}

func (k *k8sResourceAccesser) GetCron(cronClient *cronversioned.Clientset, namespace string, name string) (*cron_v1alpha1.Cron, error) {
	cron := &cron_v1alpha1.Cron{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, cron)
		if err != nil {
			return nil, fmt.Errorf("failed to find cron %v from cache, reason: %v", name, err)
		}
	} else {
		cron, err = cronClient.AppsV1alpha1().Crons(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to find cron %v from api server, reason: %v", name, err)
		}
	}
	return cron, nil
}

func (k *k8sResourceAccesser) GetTensorflowJob(tfjobClient *tfversioned.Clientset, namespace string, name string) (*tfv1.TFJob, error) {
	tfjob := &tfv1.TFJob{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, tfjob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, TensorflowCRDNameInDaemonMode, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find tfjob %v from cache,reason: %v", name, err)
		}
	} else {
		tfjob, err = tfjobClient.KubeflowV1().TFJobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, TensorflowCRDName, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find tfjob %v from api server,reason: %v", name, err)
		}
	}
	return tfjob, nil
}

func (k *k8sResourceAccesser) GetMPIJob(mpijobClient *mpiversioned.Clientset, namespace string, name string) (*v1alpha1.MPIJob, error) {
	mpijob := &v1alpha1.MPIJob{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, mpijob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, MPICRDNameInDaemonMode, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find mpijob %v from cache,reason: %v", name, err)
		}
	} else {
		mpijob, err = mpijobClient.KubeflowV1alpha1().MPIJobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, MPICRDName, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find mpijob %v from api server,reason: %v", name, err)
		}
	}
	return mpijob, err
}

func (k *k8sResourceAccesser) GetPytorchJob(pytorchjobClient *pyversioned.Clientset, namespace string, name string) (*pytorch_v1.PyTorchJob, error) {
	pytorchjob := &pytorch_v1.PyTorchJob{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, pytorchjob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, PytorchCRDNameInDaemonMode, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find pytorchjob %v from cache,reason: %v", name, err)
		}

	} else {
		pytorchjob, err = pytorchjobClient.KubeflowV1().PyTorchJobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, PytorchCRDName, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find pytorchjob %v from api server,reason: %v", name, err)
		}
	}
	return pytorchjob, err
}

func (k *k8sResourceAccesser) GetETJob(etjobClient *etversioned.Clientset, namespace string, name string) (*v1alpha12.TrainingJob, error) {
	etjob := &v1alpha12.TrainingJob{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, etjob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, ETCRDNameInDaemonMode, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find elastic job %v from cache,reason: %v", name, err)
		}
	} else {
		etjob, err = etjobClient.EtV1alpha1().TrainingJobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, ETCRDName, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find elastic job %v from api server,reason: %v", name, err)
		}
	}
	return etjob, err
}

func (k *k8sResourceAccesser) GetVolcanoJob(volcanojobClient *volcanovesioned.Clientset, namespace string, name string) (*volcano_v1alpha1.Job, error) {
	volcanoJob := &volcano_v1alpha1.Job{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, volcanoJob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, VolcanoCRDNameInDaemonMode, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find volcanojob %v from cache,reason: %v", name, err)
		}
	} else {
		volcanoJob, err = volcanojobClient.BatchV1alpha1().Jobs(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, VolcanoCRDName, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find volcanojob from api server,reason: %v", err)
		}
	}
	return volcanoJob, err
}

func (k *k8sResourceAccesser) GetSparkJob(sparkjobClient *sparkversioned.Clientset, namespace string, name string) (*spark_v1beta2.SparkApplication, error) {
	sparkJob := &spark_v1beta2.SparkApplication{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, sparkJob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, SparkCRDNameInDaemonMode, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find sparkjob %v from cache,reason: %v", name, err)
		}

	} else {
		sparkJob, err = sparkjobClient.SparkoperatorV1beta2().SparkApplications(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, SparkCRDName, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find sparkjob %v from api server,reason: %v", name, err)
		}
	}
	return sparkJob, err
}

func (k *k8sResourceAccesser) GetService(namespace, name string) (*v1.Service, error) {
	service := &v1.Service{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, service)
	} else {
		service, err = k.clientset.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (k *k8sResourceAccesser) GetEndpoints(namespace, name string) (*v1.Endpoints, error) {
	endpoints := &v1.Endpoints{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, endpoints)
	} else {
		endpoints, err = k.clientset.CoreV1().Endpoints(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (k *k8sResourceAccesser) GetNode(nodeName string) (*v1.Node, error) {
	node := &v1.Node{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Name: nodeName}, node)
	} else {
		node, err = k.clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}
	return node, nil
}

func parseLabelSelector(item string) (labels.Selector, error) {
	if item == "" {
		return labels.Everything(), nil
	}
	selector, err := labels.Parse(item)
	if err != nil {
		log.Errorf("failed to parse label selectors,reason: %v", err)
		return nil, err
	}
	return selector, nil
}

func parseFieldSelector(item string) (fields.Selector, error) {
	if item == "" {
		return fields.Everything(), nil
	}
	selector, err := fields.ParseSelector(item)
	if err != nil {
		log.Errorf("failed to parse fields selectors,reason: %v", err)
		return nil, err
	}
	return selector, nil
}

func createClientListOptions(labelSelector labels.Selector, fieldSelector fields.Selector) *client.ListOptions {
	options := &client.ListOptions{
		LabelSelector: labelSelector,
	}
	if !fieldSelector.Empty() {
		options.FieldSelector = fieldSelector
	}
	return options
}

// Copyright 2024 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8saccesser

import (
	"context"
	"fmt"
	"strings"
	"sync"

	kservev1beta1 "github.com/kserve/kserve/pkg/apis/serving/v1beta1"
	kserveClient "github.com/kserve/kserve/pkg/client/clientset/versioned"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/kubeflow/arena/pkg/apis/types"
	v1alpha12 "github.com/kubeflow/arena/pkg/operators/et-operator/api/v1alpha1"
	etversioned "github.com/kubeflow/arena/pkg/operators/et-operator/client/clientset/versioned"
	cron_v1alpha1 "github.com/kubeflow/arena/pkg/operators/kubedl-operator/apis/apps/v1alpha1"
	cronversioned "github.com/kubeflow/arena/pkg/operators/kubedl-operator/client/clientset/versioned"
	mpi_v1 "github.com/kubeflow/arena/pkg/operators/mpi-operator/apis/kubeflow/v1"
	mpiversioned "github.com/kubeflow/arena/pkg/operators/mpi-operator/client/clientset/versioned"
	pytorch_v1 "github.com/kubeflow/arena/pkg/operators/pytorch-operator/apis/pytorch/v1"
	pyversioned "github.com/kubeflow/arena/pkg/operators/pytorch-operator/client/clientset/versioned"
	spark_v1beta2 "github.com/kubeflow/arena/pkg/operators/spark-operator/apis/sparkoperator.k8s.io/v1beta2"
	sparkversioned "github.com/kubeflow/arena/pkg/operators/spark-operator/client/clientset/versioned"
	tfv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/tensorflow/v1"
	tfversioned "github.com/kubeflow/arena/pkg/operators/tf-operator/client/clientset/versioned"
	volcano_v1alpha1 "github.com/kubeflow/arena/pkg/operators/volcano-operator/apis/batch/v1alpha1"
	volcanovesioned "github.com/kubeflow/arena/pkg/operators/volcano-operator/client/clientset/versioned"
	ray_v1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	rayversioned "github.com/ray-project/kuberay/ray-operator/pkg/client/clientset/versioned"
	lws_v1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
	lwsversioned "sigs.k8s.io/lws/client-go/clientset/versioned"
)

var accesser *k8sResourceAccesser
var once sync.Once

func init() {
	utilruntime.Must(tfv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(mpi_v1.AddToScheme(scheme.Scheme))
	utilruntime.Must(v1alpha12.AddToScheme(scheme.Scheme))
	utilruntime.Must(pytorch_v1.AddToScheme(scheme.Scheme))
	utilruntime.Must(spark_v1beta2.AddToScheme(scheme.Scheme))
	utilruntime.Must(volcano_v1alpha1.AddToScheme(scheme.Scheme))
	utilruntime.Must(cron_v1alpha1.AddToScheme(scheme.Scheme))
	utilruntime.Must(ray_v1.AddToScheme(scheme.Scheme))
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
		httpClient, err := rest.HTTPClientFor(config)
		if err != nil {
			log.Errorf("failed to create httpClient, reason: %v", err)
			return nil, err
		}
		mapper, err := apiutil.NewDynamicRESTMapper(config, httpClient)
		if err != nil {
			log.Errorf("failed to create cacheClient mapper, reason: %v", err)
			// if create dynamic mapper failed, use default restMapper
			mapper = nil
		}
		cacheClient, err = cache.New(config, cache.Options{Mapper: mapper})
		if err != nil {
			log.Errorf("failed to create cacheClient, reason: %v", err)
			return nil, err
		}
		_ = cacheClient.IndexField(context.TODO(), &corev1.Pod{}, "spec.nodeName", func(o client.Object) []string {
			if pod, ok := o.(*corev1.Pod); ok {
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
		if err = k.cacheClient.Start(context.Background()); err != nil {
			cancel()
			log.Errorf("failed to start cacheClient, reason: %v", err)
		}
	}()
	k.cacheClient.WaitForCacheSync(ctx)
	return err
}

func (k *k8sResourceAccesser) GetCacheClient() cache.Cache {
	return k.cacheClient
}

func (k *k8sResourceAccesser) ListPods(namespace string, filterLabels string, filterFields string, filterFunc func(*corev1.Pod) bool) ([]*corev1.Pod, error) {
	pods := []*corev1.Pod{}
	podList := &corev1.PodList{}
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

func (k *k8sResourceAccesser) ListStatefulSets(namespace string, filterLabels string) ([]*appsv1.StatefulSet, error) {
	statefulsets := []*appsv1.StatefulSet{}
	stsList := &appsv1.StatefulSetList{}
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

func (k *k8sResourceAccesser) ListDeployments(namespace string, filterLabels string) ([]*appsv1.Deployment, error) {
	deployments := []*appsv1.Deployment{}
	deployList := &appsv1.DeploymentList{}
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

func (k *k8sResourceAccesser) ListServices(namespace string, filterLabels string) ([]*corev1.Service, error) {
	services := []*corev1.Service{}
	serviceList := &corev1.ServiceList{}
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

func (k *k8sResourceAccesser) ListConfigMaps(namespace string, filterLabels string) ([]*corev1.ConfigMap, error) {
	configmaps := []*corev1.ConfigMap{}
	configmapList := &corev1.ConfigMapList{}
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

func (k *k8sResourceAccesser) ListNodes(filterLabels string) ([]*corev1.Node, error) {
	nodeList := &corev1.NodeList{}
	nodes := []*corev1.Node{}
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

func (k *k8sResourceAccesser) ListKServeInferenceService(ctx context.Context, kserveClient *kserveClient.Clientset, namespace string, label string) ([]*kservev1beta1.InferenceService, error) {
	jobs := []*kservev1beta1.InferenceService{}
	jobList := &kservev1beta1.InferenceServiceList{}
	var err error
	labelSelector, err := parseLabelSelector(label)
	if err != nil {
		return nil, err
	}
	if k.cacheEnabled {
		err = k.cacheClient.List(
			ctx,
			jobList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			},
		)
	} else {
		jobList, err = kserveClient.ServingV1beta1().InferenceServices(namespace).List(ctx, metav1.ListOptions{
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

func (k *k8sResourceAccesser) ListTensorflowJobs(tfjobClient *tfversioned.Clientset, namespace string, tfjobLabel string) ([]*tfv1.TFJob, error) {
	jobs := []*tfv1.TFJob{}
	jobList := &tfv1.TFJobList{}
	var err error
	labelSelector, err := parseLabelSelector(tfjobLabel)
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

func (k *k8sResourceAccesser) ListMPIJobs(mpijobClient *mpiversioned.Clientset, namespace string, labels string) ([]*mpi_v1.MPIJob, error) {
	jobs := []*mpi_v1.MPIJob{}
	jobList := &mpi_v1.MPIJobList{}
	var err error
	labelSelector, err := parseLabelSelector(labels)
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
		jobList, err = mpijobClient.KubeflowV1().MPIJobs(namespace).List(context.TODO(), metav1.ListOptions{
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

func (k *k8sResourceAccesser) ListPytorchJobs(pytorchjobClient *pyversioned.Clientset, namespace string, labels string) ([]*pytorch_v1.PyTorchJob, error) {
	jobs := []*pytorch_v1.PyTorchJob{}
	jobList := &pytorch_v1.PyTorchJobList{}
	var err error
	labelSelector, err := parseLabelSelector(labels)
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

func (k *k8sResourceAccesser) ListRayJobs(rayClient *rayversioned.Clientset, namespace string, labels string) ([]*ray_v1.RayJob, error) {
	jobs := []*ray_v1.RayJob{}
	jobList := &ray_v1.RayJobList{}
	var err error
	labelSelector, err := parseLabelSelector(labels)
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
		jobList, err = rayClient.RayV1().RayJobs(namespace).List(context.Background(), metav1.ListOptions{
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

func (k *k8sResourceAccesser) ListETJobs(etjobClient *etversioned.Clientset, namespace string, labels string) ([]*v1alpha12.TrainingJob, error) {
	jobs := []*v1alpha12.TrainingJob{}
	jobList := &v1alpha12.TrainingJobList{}
	var err error
	labelSelector, err := parseLabelSelector(labels)
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

func (k *k8sResourceAccesser) ListVolcanoJobs(volcanojobClient *volcanovesioned.Clientset, namespace string, labels string) ([]*volcano_v1alpha1.Job, error) {
	jobs := []*volcano_v1alpha1.Job{}
	jobList := &volcano_v1alpha1.JobList{}
	var err error
	labelSelector, err := parseLabelSelector(labels)
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

func (k *k8sResourceAccesser) ListSparkJobs(sparkjobClient *sparkversioned.Clientset, namespace string, labels string) ([]*spark_v1beta2.SparkApplication, error) {
	jobs := []*spark_v1beta2.SparkApplication{}
	jobList := &spark_v1beta2.SparkApplicationList{}
	var err error
	labelSelector, err := parseLabelSelector(labels)
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

func (k *k8sResourceAccesser) ListLWSJobs(lwsClient *lwsversioned.Clientset, namespace string, labels string) ([]*lws_v1.LeaderWorkerSet, error) {
	jobs := []*lws_v1.LeaderWorkerSet{}
	jobList := &lws_v1.LeaderWorkerSetList{}
	var err error
	labelSelector, err := parseLabelSelector(labels)
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
		jobList, err = lwsClient.LeaderworkersetV1().LeaderWorkerSets(namespace).List(
			context.Background(),
			metav1.ListOptions{LabelSelector: labelSelector.String()},
		)
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

func (k *k8sResourceAccesser) GetMPIJob(mpijobClient *mpiversioned.Clientset, namespace string, name string) (*mpi_v1.MPIJob, error) {
	mpijob := &mpi_v1.MPIJob{}
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
		mpijob, err = mpijobClient.KubeflowV1().MPIJobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
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

func (k *k8sResourceAccesser) GetRayJob(rayClient *rayversioned.Clientset, namespace string, name string) (*ray_v1.RayJob, error) {
	rayJob := &ray_v1.RayJob{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, rayJob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, RayJobCRDNameInDaemonMode, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find rayjob %v from cache,reason: %v", name, err)
		}

	} else {
		rayJob, err = rayClient.RayV1().RayJobs(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, RayJobCRDName, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find rayjob %v from api server,reason: %v", name, err)
		}
	}
	return rayJob, err
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

func (k *k8sResourceAccesser) GetLWSJob(lwsClient *lwsversioned.Clientset, namespace string, name string) (*lws_v1.LeaderWorkerSet, error) {
	lwsJob := &lws_v1.LeaderWorkerSet{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, lwsJob)
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, LWSCRDNameInDaemonMode, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find lwsjob %v from cache,reason: %v", name, err)
		}
	} else {
		lwsJob, err = lwsClient.LeaderworkersetV1().LeaderWorkerSets(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf(`%v "%v" not found`, SparkCRDName, name)) {
				return nil, types.ErrTrainingJobNotFound
			}
			return nil, fmt.Errorf("failed to find lws %v from api server,reason: %v", name, err)
		}
	}
	return lwsJob, err
}

func (k *k8sResourceAccesser) GetService(namespace, name string) (*corev1.Service, error) {
	service := &corev1.Service{}
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

func (k *k8sResourceAccesser) GetEndpoints(namespace, name string) (*corev1.Endpoints, error) {
	endpoints := &corev1.Endpoints{}
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

func (k *k8sResourceAccesser) GetNode(nodeName string) (*corev1.Node, error) {
	node := &corev1.Node{}
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

func (k *k8sResourceAccesser) GetJob(name, namespace string) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	var err error
	if k.cacheEnabled {
		err = k.cacheClient.Get(context.TODO(), client.ObjectKey{Name: name, Namespace: namespace}, job)
	} else {
		job, err = k.clientset.BatchV1().Jobs(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	}

	if err != nil {
		return nil, err
	}
	return job, nil
}

func (k *k8sResourceAccesser) ListJobs(namespace string, filterLabels string, filterFields string, filterFunc func(*batchv1.Job) bool) ([]*batchv1.Job, error) {
	jobs := []*batchv1.Job{}
	jobList := &batchv1.JobList{}
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
			jobList,
			client.InNamespace(namespace),
			&client.ListOptions{
				LabelSelector: labelSelector,
			},
		)
	} else {
		jobList, err = k.clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Job",
				APIVersion: "batch/v1",
			},
			LabelSelector: labelSelector.String(),
			FieldSelector: fieldSelector.String(),
		})
	}

	if err != nil {
		return nil, err
	}
	for _, job := range jobList.Items {
		copyJob := job.DeepCopy()
		if filterFunc != nil && !filterFunc(copyJob) {
			continue
		}
		jobs = append(jobs, copyJob)
	}
	return jobs, nil
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

package arenacache

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sync"

	v1alpha12 "github.com/kubeflow/arena/pkg/operators/et-operator/api/v1alpha1"
	"github.com/kubeflow/arena/pkg/operators/mpi-operator/apis/kubeflow/v1alpha1"
	pytorch_v1 "github.com/kubeflow/arena/pkg/operators/pytorch-operator/apis/pytorch/v1"
	spark_v1beta1 "github.com/kubeflow/arena/pkg/operators/spark-operator/apis/sparkoperator.k8s.io/v1beta1"
	tfv1 "github.com/kubeflow/arena/pkg/operators/tf-operator/apis/tensorflow/v1"
	volcano_v1alpha1 "github.com/kubeflow/arena/pkg/operators/volcano-operator/apis/batch/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var cacheClient *CacheClient
var cacheClientOnce sync.Once

func init() {
	tfv1.AddToScheme(scheme.Scheme)
	v1alpha1.AddToScheme(scheme.Scheme)
	v1alpha12.AddToScheme(scheme.Scheme)
	pytorch_v1.AddToScheme(scheme.Scheme)
	spark_v1beta1.AddToScheme(scheme.Scheme)
	volcano_v1alpha1.AddToScheme(scheme.Scheme)
}

func InitCacheClient(config *rest.Config) error {
	if cacheClient != nil {
		return nil
	}
	var err error
	cacheClientOnce.Do(func() {
		cacheClient, err = NewCacheClient(config)
		if err == nil {
			err = cacheClient.Run()
		}
	})
	return err
}

func GetCacheClient() *CacheClient {
	return cacheClient
}
func NewCacheClient(config *rest.Config) (*CacheClient, error) {
	mapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		log.Errorf("failed to create cacheClient mapper, reason: %v", err)
		// if create dynamic mapper failed, use default restMapper
		mapper = nil
	}
	cc, err := cache.New(config, cache.Options{Mapper: mapper, Resync: nil, Namespace: ""})
	if err != nil {
		log.Errorf("failed to create cacheClient, reason: %v", err)
		return nil, err
	}
	cc.IndexField(&corev1.Pod{}, "spec.nodeName", func(o runtime.Object) []string {
		if pod, ok := o.(*corev1.Pod); ok {
			return []string{pod.Spec.NodeName}
		}
		return []string{}
	})
	return &CacheClient{Cache: cc}, nil
}

type CacheClient struct {
	cache.Cache
}

var releaseSelector = client.HasLabels{
	"release",
}

func matchLabelsSelector(name, jobType string) client.ListOption {
	return client.MatchingLabels{
		"release": name,
		"app":     jobType,
	}
}

func (c *CacheClient) Run() (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err = c.Cache.Start(context.Background().Done()); err != nil {
			cancel()
			log.Errorf("failed to start cacheClient, reason: %v", err)
		}
	}()
	c.Cache.WaitForCacheSync(ctx.Done())
	return err
}

func (c *CacheClient) ListTrainingJobs(list runtime.Object, namespace string) error {
	if err := cacheClient.List(context.Background(), list, client.InNamespace(namespace), releaseSelector); err != nil {
		return err
	}
	return nil
}

func (c *CacheClient) ListTrainingJobResources(list runtime.Object, namespace, name, trainingType string) error {
	return cacheClient.List(context.Background(), list, client.InNamespace(namespace), matchLabelsSelector(name, trainingType))
}

func (c *CacheClient) ListNodeRunningPods(nodeName string) ([]*corev1.Pod, error) {
	selector := fmt.Sprintf("spec.nodeName=%v", nodeName)
	fieldSelector, err := fields.ParseSelector(selector)
	if err != nil {
		return nil, err
	}
	podList := &corev1.PodList{}
	if err = cacheClient.List(context.Background(), podList, client.MatchingFieldsSelector{Selector: fieldSelector}); err != nil {
		return nil, err
	}
	pods := []*corev1.Pod{}
	for _, pod := range podList.Items {
		pods = append(pods, pod.DeepCopy())
	}
	return filterRunningPods(pods), nil
}

func filterRunningPods(podList []*corev1.Pod) []*corev1.Pod {
	pods := []*corev1.Pod{}
	for i, _ := range podList {
		pod := podList[i]
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		pods = append(pods, pod)
	}
	return pods
}

func (c *CacheClient) ListResources(list runtime.Object, namespace string, options metav1.ListOptions) error {
	ops := []client.ListOption{}
	if namespace != metav1.NamespaceAll {
		ops = append(ops, client.InNamespace(namespace))
	}
	if options.LabelSelector != "" {
		labelSelector, err := labels.Parse(options.LabelSelector)
		if err != nil {
			panic(fmt.Errorf("invalid selector %q: %v", options.LabelSelector, err))
		}
		ops = append(ops, client.MatchingLabelsSelector{Selector: labelSelector})
	}
	if options.FieldSelector != "" {
		fieldSelector, err := fields.ParseSelector(options.FieldSelector)
		if err != nil {
			panic(fmt.Errorf("invalid selector %q: %v", options.FieldSelector, err))
		}
		ops = append(ops, client.MatchingFieldsSelector{Selector: fieldSelector})
	}
	if err := cacheClient.List(context.Background(), list, ops...); err != nil {
		return err
	}
	return nil
}

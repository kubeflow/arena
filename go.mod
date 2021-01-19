module github.com/kubeflow/arena

go 1.12

require (
	github.com/AliyunContainerService/et-operator v0.0.0-00010101000000-000000000000
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-00010101000000-000000000000
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-openapi/spec v0.20.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/kubeflow/common v0.3.1
	github.com/kubeflow/mpi-operator v0.0.0-00010101000000-000000000000
	github.com/kubeflow/pytorch-operator v0.0.0-00010101000000-000000000000
	github.com/kubeflow/tf-operator v0.0.0-00010101000000-000000000000
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/volcano.sh/volcano v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.4.0
	istio.io/api v0.0.0-20180824201241-76349c53b87f
	k8s.io/api v0.16.9
	k8s.io/apimachinery v0.16.9
	k8s.io/client-go v0.16.9
)

replace (
	github.com/AliyunContainerService/et-operator => ./dependency/et-operator
	github.com/GoogleCloudPlatform/spark-on-k8s-operator => ./dependency/spark-on-k8s-operator
	github.com/kubeflow/mpi-operator => ./dependency/mpi-operator
	github.com/kubeflow/pytorch-operator => ./dependency/pytorch-operator
	github.com/kubeflow/tf-operator => ./dependency/tf-operator
	github.com/volcano.sh/volcano => ./dependency/volcano

	k8s.io/api => k8s.io/api v0.16.9
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.9
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.10-beta.0
	k8s.io/apiserver => k8s.io/apiserver v0.16.9
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.9
	k8s.io/client-go => k8s.io/client-go v0.16.9
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.9
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.9
	k8s.io/code-generator => k8s.io/code-generator v0.16.10-beta.0
	k8s.io/component-base => k8s.io/component-base v0.16.9
	k8s.io/cri-api => k8s.io/cri-api v0.16.10-beta.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.9
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.9
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.9
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.9
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.9
	k8s.io/kubectl => k8s.io/kubectl v0.16.9
	k8s.io/kubelet => k8s.io/kubelet v0.16.9
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.9
	k8s.io/metrics => k8s.io/metrics v0.16.9
	k8s.io/node-api => k8s.io/node-api v0.16.9
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.9
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.16.9
	k8s.io/sample-controller => k8s.io/sample-controller v0.16.9
)

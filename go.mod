module github.com/kubeflow/arena

go 1.12

require (
	github.com/PuerkitoBio/purell v1.1.0 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/cpuguy83/go-md2man v1.0.8 // indirect
	github.com/emicklei/go-restful v2.8.0+incompatible // indirect
	github.com/go-openapi/jsonpointer v0.0.0-20180322222829-3a0015ad55fa // indirect
	github.com/go-openapi/jsonreference v0.0.0-20180322222742-3fb327e6747d // indirect
	github.com/go-openapi/spec v0.0.0-20180825180323-f1468acb3b29
	github.com/go-openapi/swag v0.0.0-20180703152219-2b0bd4f193d0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/inconshreveable/mousetrap v0.0.0-20141017200713-76626ae9c91c // indirect
	github.com/mailru/easyjson v0.0.0-20180723221831-d5012789d665 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/onsi/gomega v1.10.3 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/russross/blackfriday v1.5.1 // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.3.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
	istio.io/api v0.0.0-20180824201241-76349c53b87f
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	k8s.io/kubernetes v1.16.9 // indirect
	sigs.k8s.io/controller-runtime v0.1.9
)

replace (
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

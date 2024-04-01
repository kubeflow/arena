module github.com/kubeflow/arena

go 1.20

require (
	github.com/docker/docker v23.0.5+incompatible
	github.com/go-resty/resty/v2 v2.12.0
	github.com/golang/glog v1.1.0
	github.com/google/uuid v1.3.0
	github.com/kserve/kserve v0.11.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/prometheus/client_golang v1.15.1
	github.com/prometheus/common v0.43.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/crypto v0.21.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/yaml.v2 v2.4.0
	istio.io/api v0.0.0-20200715212100-dbf5277541ef
	k8s.io/api v0.26.4
	k8s.io/apiextensions-apiserver v0.26.4
	k8s.io/apimachinery v0.26.4
	k8s.io/cli-runtime v0.26.4
	k8s.io/client-go v0.26.4
	k8s.io/kubectl v0.26.4
	sigs.k8s.io/controller-runtime v0.14.6
)

require (
	cloud.google.com/go v0.110.2 // indirect
	cloud.google.com/go/compute v1.19.3 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v1.0.1 // indirect
	cloud.google.com/go/storage v1.30.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.44.264 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.10.2 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-containerregistry v0.15.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.3 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.8.0 // indirect
	github.com/googleapis/google-cloud-go-testing v0.0.0-20210719221736-1c9a4c676720 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.15 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/api v0.122.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/grpc v1.56.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	istio.io/gogo-genproto v0.0.0-20190930162913-45029607206a // indirect
	k8s.io/component-base v0.26.4 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230515203736-54b630e78af5 // indirect
	k8s.io/utils v0.0.0-20230505201702-9f6742963106 // indirect
	knative.dev/networking v0.0.0-20230511122402-33636d99d870 // indirect
	knative.dev/pkg v0.0.0-20230502134655-db8a35330281 // indirect
	knative.dev/serving v0.37.1 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.12.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.9 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0

replace k8s.io/api => k8s.io/api v0.26.4

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.26.4

replace k8s.io/apimachinery => k8s.io/apimachinery v0.26.4

replace k8s.io/apiserver => k8s.io/apiserver v0.26.4

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.26.4

replace k8s.io/client-go => k8s.io/client-go v0.26.4

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.26.4

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.26.4

replace k8s.io/code-generator => k8s.io/code-generator v0.26.4

replace k8s.io/component-base => k8s.io/component-base v0.26.4

replace k8s.io/component-helpers => k8s.io/component-helpers v0.26.4

replace k8s.io/controller-manager => k8s.io/controller-manager v0.26.4

replace k8s.io/cri-api => k8s.io/cri-api v0.26.4

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.26.4

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.26.4

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.26.4

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.26.4

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.26.4

replace k8s.io/kubectl => k8s.io/kubectl v0.26.4

replace k8s.io/kubelet => k8s.io/kubelet v0.26.4

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.26.4

replace k8s.io/metrics => k8s.io/metrics v0.26.4

replace k8s.io/mount-utils => k8s.io/mount-utils v0.26.4

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.26.4

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.26.4

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.26.4

replace k8s.io/sample-controller => k8s.io/sample-controller v0.26.4

replace k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.26.4

replace k8s.io/kms => k8s.io/kms v0.26.4

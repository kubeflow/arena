module github.com/kubeflow/arena

go 1.18

require (
	github.com/docker/docker v20.10.7+incompatible
	github.com/golang/glog v1.0.0
	github.com/golang/protobuf v1.5.3
	github.com/google/uuid v1.3.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.39.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.6.0
	github.com/spf13/viper v1.10.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	gopkg.in/yaml.v2 v2.4.0
	istio.io/api v0.0.0-20200715212100-dbf5277541ef
	k8s.io/api v0.23.9
	k8s.io/apiextensions-apiserver v0.23.9
	k8s.io/apimachinery v0.23.9
	k8s.io/cli-runtime v0.23.9
	k8s.io/client-go v0.23.9
	k8s.io/kubectl v0.23.9
	sigs.k8s.io/controller-runtime v0.11.2
)

require (
	cloud.google.com/go v0.99.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd // indirect
	github.com/chai2010/gettext-go v0.0.0-20160711120539-c6fed771bfd5 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/term v0.0.0-20210610120745-9d4ed1856297 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday v1.5.2 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca // indirect
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/oauth2 v0.3.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/term v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.66.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	istio.io/gogo-genproto v0.0.0-20190930162913-45029607206a // indirect
	k8s.io/component-base v0.23.9 // indirect
	k8s.io/klog/v2 v2.30.0 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/kustomize/api v0.10.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace k8s.io/api => k8s.io/api v0.23.9

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.9

replace k8s.io/apimachinery => k8s.io/apimachinery v0.23.11-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.23.9

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.9

replace k8s.io/client-go => k8s.io/client-go v0.23.9

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.9

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.9

replace k8s.io/code-generator => k8s.io/code-generator v0.23.14-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.23.9

replace k8s.io/component-helpers => k8s.io/component-helpers v0.23.9

replace k8s.io/controller-manager => k8s.io/controller-manager v0.23.9

replace k8s.io/cri-api => k8s.io/cri-api v0.23.14-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.9

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.9

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.9

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.9

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.9

replace k8s.io/kubectl => k8s.io/kubectl v0.23.9

replace k8s.io/kubelet => k8s.io/kubelet v0.23.9

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.9

replace k8s.io/metrics => k8s.io/metrics v0.23.9

replace k8s.io/mount-utils => k8s.io/mount-utils v0.23.14-rc.0

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.23.9

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.9

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.23.9

replace k8s.io/sample-controller => k8s.io/sample-controller v0.23.9

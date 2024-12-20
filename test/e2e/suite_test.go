/*
Copyright 2024 The Kubeflow authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kubeflow/arena/pkg/util"
)

const (
	ReleaseName      = "arena-artifacts"
	ReleaseNamespace = "arena-system"

	WaitTimeout = 5 * time.Minute
)

var (
	cfg       *rest.Config
	testEnv   *envtest.Environment
	k8sClient client.Client
	clientset *kubernetes.Clientset
)

func TestArena(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Arena Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	var err error

	By("Bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "arena-artifacts", "all_crds", "v1", "pytorch-operator"),
			filepath.Join("..", "..", "arena-artifacts", "all_crds", "v1", "tf-operator"),
			filepath.Join("..", "..", "arena-artifacts", "all_crds", "v1", "mpi-operator"),
			filepath.Join("..", "..", "arena-artifacts", "all_crds", "v1", "et-operator"),
			filepath.Join("..", "..", "arena-artifacts", "all_crds", "v1", "cron-operator"),
		},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.29.3-%s-%s", runtime.GOOS, runtime.GOARCH)),
		UseExistingCluster: util.BoolPtr(true),
	}
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	clientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(clientset).NotTo(BeNil())

	By("Installing the arena artifacts")
	envSettings := cli.New()
	envSettings.SetNamespace(ReleaseNamespace)
	actionConfig := &action.Configuration{}
	Expect(actionConfig.Init(
		envSettings.RESTClientGetter(),
		envSettings.Namespace(),
		os.Getenv("HELM_DRIVER"),
		func(format string, v ...interface{}) {
			logf.Log.Info(fmt.Sprintf(format, v...))
		})).NotTo(HaveOccurred())
	installAction := action.NewInstall(actionConfig)
	Expect(installAction).NotTo(BeNil())
	installAction.ReleaseName = ReleaseName
	installAction.Namespace = envSettings.Namespace()
	installAction.CreateNamespace = true
	installAction.Wait = true
	installAction.Timeout = WaitTimeout
	chartPath := filepath.Join("..", "..", "arena-artifacts")
	chart, err := loader.Load(chartPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(chart).NotTo(BeNil())
	values, err := chartutil.ReadValuesFile(filepath.Join(chartPath, "ci", "values.yaml"))
	Expect(err).NotTo(HaveOccurred())
	Expect(values).NotTo(BeNil())
	release, err := installAction.Run(chart, values)
	Expect(err).NotTo(HaveOccurred())
	Expect(release).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("Tearing down the test environment")
	Expect(testEnv.Stop()).NotTo(HaveOccurred())
})

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var arenaV2Bin string

func TestArenaV2(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Arena v2 E2E Suite")
}

var _ = BeforeSuite(func() {
	crdDir := "crds"
	if _, err := os.Stat(crdDir); os.IsNotExist(err) {
		Fail(fmt.Sprintf("CRD directory not found: %s — run `make v2-e2e-setup` first", crdDir))
	}

	localBin := filepath.Join("..", "..", "bin", "arena-v2")
	if p, err := exec.LookPath(localBin); err == nil {
		arenaV2Bin = p
	} else if p, err := exec.LookPath("arena-v2"); err == nil {
		arenaV2Bin = p
	} else {
		Fail("arena-v2 binary not found in bin/ or PATH — run `make arena-v2` first")
	}
})

func createTempYAML(content string) (string, error) {
	tmpDir := os.TempDir()
	path := filepath.Join(tmpDir, fmt.Sprintf("arena-test-%d.yaml", GinkgoRandomSeed()))
	err := os.WriteFile(path, []byte(content), 0644)
	return path, err
}

func mpiJobStorageVersion() string {
	cmd := exec.Command("kubectl", "get", "crd", "mpijobs.kubeflow.org",
		"-o", "jsonpath={.spec.versions[?(@.storage==true)].name}")
	out, err := cmd.Output()
	if err != nil {
		Fail(fmt.Sprintf("failed to query MPIJob CRD storage version: %v — ensure mpijobs.kubeflow.org CRD is installed", err))
	}
	return "kubeflow.org/" + string(out)
}

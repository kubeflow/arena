// Package e2e_test contains end-to-end tests for the arena v2 CLI.
//
// The lifecycle tests in this suite are CRUD smoke tests: they verify that
// job resources can be submitted, listed, retrieved, and deleted through the
// arena CLI. They use placeholder images and do NOT wait for pod readiness
// or validate training outcomes.
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
	f, err := os.CreateTemp("", "arena-test-*.yaml")
	if err != nil {
		return "", err
	}
	name := f.Name()
	if _, err := f.Write([]byte(content)); err != nil {
		f.Close()
		os.Remove(name)
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return name, nil
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

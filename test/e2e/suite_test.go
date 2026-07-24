// Package e2e_test contains end-to-end tests for the arena v2 CLI.
//
// The lifecycle tests in this suite are CRUD smoke tests: they verify that
// job resources can be submitted, listed, retrieved, and deleted through the
// arena CLI. They use placeholder images and do NOT wait for pod readiness
// or validate training outcomes.
package e2e_test

import (
	"bytes"
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
	localBin := filepath.Join("..", "..", "bin", "arena-v2")
	if p, err := exec.LookPath(localBin); err == nil {
		arenaV2Bin = p
	} else if p, err := exec.LookPath("arena-v2"); err == nil {
		arenaV2Bin = p
	} else {
		Fail("arena-v2 binary not found in bin/ or PATH — run `make arena-v2` first")
	}
})

var _ = AfterSuite(func() {
	// Sweep any leaked arena-v2 jobs and test resources.
	// Individual AfterEach blocks should handle cleanup, but this
	// catches anything that leaked due to mid-test failures.
	var out bytes.Buffer

	// Delete all PyTorchJobs, TFJobs, and MPIJobs with arena.io/framework label
	for _, resource := range []string{"pytorchjob", "tfjob", "mpijob"} {
		cmd := exec.Command("kubectl", "delete", resource,
			"-l", "arena.io/framework",
			"-n", "default", "--ignore-not-found")
		cmd.Stdout = &out
		cmd.Stderr = &out
		_ = cmd.Run()
	}

	// Clean up storage test prerequisites
	for _, resource := range []string{
		"configmap/app-config",
		"secret/db-credentials",
		"secret/ssh-keys",
	} {
		cmd := exec.Command("kubectl", "delete", resource,
			"-n", "default", "--ignore-not-found")
		cmd.Stdout = &out
		cmd.Stderr = &out
		_ = cmd.Run()
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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

var (
	forceDelete    = flag.Bool("force", false, "force delete the Custom Resource Instances")
	manifest       = flag.String("manifest-dir", "", "specify the kubernetes-artifacts directory")
	arenaNamespace = flag.String("namespace", "arena-system", "specify the namespace")
	allCRDNames    = []string{
		"crons.apps.kubedl.io",
		"mpijobs.kubeflow.org",
		"pytorchjobs.kubeflow.org",
		"scaleins.kai.alibabacloud.com",
		"scaleouts.kai.alibabacloud.com",
		"tfjobs.kubeflow.org",
		"trainingjobs.kai.alibabacloud.com",
	}
)

func main() {
	flag.Parse()
	force := *forceDelete
	if err := deleteArenaArtifacts(force); err != nil {
		fmt.Printf("Error: failed to delete arena artifacts,reason: %v\n", err)
		os.Exit(1)
	}
	deleteClientFiles()
}

func deleteArenaArtifacts(force bool) error {
	if !force {
		crdNames, err := getInstalledCRDs(allCRDNames)
		if err != nil {
			return err
		}
		if err := CheckRunningJobs(crdNames); err != nil {
			return err
		}
	}
	execCommand([]string{"helm", "del", "arena-artifacts", "-n", *arenaNamespace})
	manifest, err := detectManifests(manifest)
	if err != nil {
		return fmt.Errorf("failed to detect manifest directory,reason: %v", err)
	}
	_, fileNames := detectK8SResources(manifest, []string{"CustomResourceDefinition"})
	execCommand([]string{"arena-kubectl", "delete", "crd", strings.Join(allCRDNames, " ")})
	deleteK8sResources(fileNames)
	_, stderr, err := execCommand([]string{"arena-kubectl", "delete", "ns", *arenaNamespace})
	if err != nil && !strings.Contains(stderr, fmt.Sprintf(`namespaces "%v" not found`, *arenaNamespace)) {
		return fmt.Errorf("failed to delete namespace %v,reason: %v,%v", *arenaNamespace, err, stderr)
	}
	fmt.Printf("Debug: succeed to delete namespace %v\n", *arenaNamespace)
	return nil
}

func deleteK8sResources(files []string) {
	for _, f := range files {
		stdout, stderr, err := execCommand([]string{"arena-kubectl", "delete", "-f", f})
		if err != nil {
			if strings.Contains(stderr, "Error from server (NotFound)") {
				continue
			}
			fmt.Printf("Error: %v\n", stderr)
			continue
		}
		fmt.Println(stdout)
	}
}

func CheckRunningJobs(crdNames []string) error {
	// check some job are existed managed by CRD
	jobs := []string{}
	for _, crdName := range crdNames {
		crs, err := getCRs(crdName)
		if err != nil {
			return err
		}
		jobs = append(jobs, crs...)
	}
	if len(jobs) != 0 {
		return fmt.Errorf("failed to delete arena,because some jobs are existed,please delete them and retry:\n%v", strings.Join(jobs, "\n"))
	}
	return nil
}

func getCRs(crdName string) ([]string, error) {
	stdout, stderr, err := execCommand([]string{"arena-kubectl", "get", crdName, "--all-namespaces"})
	if err != nil {
		return nil, fmt.Errorf("failed to exec command: [arena-kubectl get %v --all-namespaces],reason: %v,%v", crdName, err, stderr)
	}
	crs := []string{}
	for _, line := range strings.Split(stdout, "\n") {
		if len(line) == 0 {
			continue
		}
		items := strings.Split(line, " ")
		if strings.Contains(items[0], "NAME") || strings.Contains(items[0], "No resources found.") {
			continue
		}
		t := []string{}
		for _, item := range items {
			item = strings.Trim(item, " ")
			if item == "" {
				continue
			}
			t = append(t, item)
		}
		if len(t) < 2 {
			continue
		}
		jobName := fmt.Sprintf("%v/%v/%v", crdName, t[0], t[1])
		crs = append(crs, jobName)
	}
	return crs, nil
}

func detectManifests(manifestDir *string) (string, error) {
	customDir := *manifestDir
	if customDir != "" {
		return customDir, nil
	}
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if CheckFileExist(path.Join(currentPath, "kubernetes-artifacts")) {
		return path.Join(currentPath, "kubernetes-artifacts"), nil
	}
	homeDir := os.Getenv("HOME")
	if CheckFileExist(path.Join(homeDir, "kubernetes-artifacts")) {
		return path.Join(homeDir, "kubernetes-artifacts"), nil
	}
	if CheckFileExist("/charts/kubernetes-artifacts") {
		return "/charts/kubernetes-artifacts", nil
	}
	return "", fmt.Errorf("not found kubernetes-artifacts directory,please download it from https://github.com/kubeflow/arena/tree/master/kubernetes-artifacts")
}

func detectK8SResources(dir string, kinds []string) (map[string][]string, []string) {
	resourceFiles := map[string][]string{}
	fileNames := []string{}
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		realPathFile := path.Join(dir, f.Name())
		if f.IsDir() {
			if CheckFileExist(path.Join(realPathFile, "Chart.yaml")) {
				continue
			}
			resources, names := detectK8SResources(realPathFile, kinds)
			for key, contents := range resources {
				if resourceFiles[key] == nil {
					resourceFiles[key] = []string{}
				}
				resourceFiles[key] = append(resourceFiles[key], contents...)
			}
			fileNames = append(fileNames, names...)
			continue
		}
		suffix := path.Ext(realPathFile)
		if suffix != ".json" && suffix != ".yaml" && suffix != ".yml" {
			continue
		}
		fileNames = append(fileNames, realPathFile)
		contentBytes, err := ioutil.ReadFile(realPathFile)
		if err != nil {
			fmt.Printf("Error: failed to read file %v,reason: %v\n", realPathFile, err)
			continue
		}
		for _, c := range strings.Split(string(contentBytes), "---\n") {
			for _, kind := range kinds {
				if resourceFiles[kind] == nil {
					resourceFiles[kind] = []string{}
				}
				if strings.Contains(c, fmt.Sprintf("kind: %v", kind)) {
					resourceFiles[kind] = append(resourceFiles[kind], strings.Trim(c, " "))
					break
				}
			}
		}

	}
	return resourceFiles, fileNames
}

func getInstalledCRDs(crdNames []string) ([]string, error) {
	allCrdsInK8s, err := getAllCRDsInK8s()
	if err != nil {
		return nil, err
	}
	// check crd is installed or not
	installedCrds := []string{}
	for _, k8sCrdName := range allCrdsInK8s {
		for _, crdName := range crdNames {
			if crdName == k8sCrdName {
				installedCrds = append(installedCrds, crdName)
			}
		}
	}
	return installedCrds, nil
}

func getAllCRDsInK8s() ([]string, error) {
	stdout, stderr, err := execCommand([]string{"arena-kubectl", "get", "crd"})
	if err != nil {
		return nil, fmt.Errorf("exit status: %v,readon: %v", err, stderr)
	}
	crds := []string{}
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.Trim(line, " ")
		if len(line) == 0 {
			continue
		}
		items := strings.Split(line, " ")
		if strings.Contains(items[0], "NAME") || strings.Contains(items[0], "No resources found.") {
			continue
		}
		crds = append(crds, items[0])
	}
	return crds, nil
}

func deleteClientFiles() {
	execCommand([]string{"rm", "-rf", "/charts"})
	execCommand([]string{"rm", "-rf", "~/charts"})
	execCommand([]string{"rm", "-rf", "/usr/local/bin/arena"})
	execCommand([]string{"rm", "-rf", "/usr/local/bin/arena-kubectl"})
	execCommand([]string{"rm", "-rf", "/usr/local/bin/arena-helm"})
	if err := removeLines([]string{"source <(arena completion bash)"}); err != nil {
		fmt.Printf("Error: failed to remove line 'source <(arena completion bash)' from ~/bashrc or ~/.zshrc\n")
		os.Exit(4)
	}
}

func removeLines(lines []string) error {
	homeDir := os.Getenv("HOME")
	bashFile := path.Join(homeDir, ".bashrc")
	zshFile := path.Join(homeDir, ".zshrc")
	updateFile := func(f string) error {
		contentBytes, err := ioutil.ReadFile(f)
		if err != nil {
			return err
		}
		content := string(contentBytes)
		for _, line := range lines {
			content = strings.ReplaceAll(content, line, "")
		}
		return ioutil.WriteFile(f, []byte(content), 0744)
	}
	if CheckFileExist(zshFile) {
		if err := updateFile(zshFile); err != nil {
			return err
		}
	}
	if CheckFileExist(bashFile) {
		if err := updateFile(bashFile); err != nil {
			return err
		}
	}
	return nil
}

func execCommand(args []string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := strings.Join(args, " ")
	fmt.Printf("Debug: exec command: [%v]\n", command)
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func CheckFileExist(filename string) bool {
	var exist = true
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		exist = false
	}
	return exist
}

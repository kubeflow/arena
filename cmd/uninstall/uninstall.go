package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	arenaChartName = "arena-artifacts"
	forceDelete    = flag.Bool("force", false, "force delete the Custom Resource Instances")
	manifest       = flag.String("manifest-dir", "", "specify the kubernetes-artifacts directory")
	arenaNamespace = flag.String("namespace", "arena-system", "specify the namespace")
)

var deleteCommandString = `
$ rm -rf /charts
$ rm -rf ~/charts
$ rm -rf /usr/local/bin/arena
$ rm -rf /usr/local/bin/arena-kubectl
$ rm -rf /usr/local/bin/arena-helm

and delete the line 'source <(arena completion bash)' in ~/.bashrc or ~/.zshrc
`

func main() {
	flag.Parse()
	force := *forceDelete
	if err := deleteArenaArtifacts(force); err != nil {
		fmt.Printf("Error: failed to delete arena artifacts,reason: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Debug: skip to remove the arena binaries,if you want to delete them,please execute following commands:\n%v", deleteCommandString)
	//deleteClientFiles()
}

func deleteArenaArtifacts(force bool) error {
	tmpFile := "/tmp/arena-artifacts.yaml"
	manifestDir, err := detectManifests(manifest)
	if err != nil {
		return err
	}
	fmt.Printf("Debug: get chart directory: %v\n", manifestDir)
	fields, err := parseFields(path.Join(manifestDir, "values.yaml"))
	if err != nil {
		return err
	}
	stdout, stderr, err := execCommand([]string{"arena-helm", "template", arenaChartName, "-n", *arenaNamespace, manifestDir, strings.Join(fields, " "), ">", tmpFile})
	if err != nil {
		return fmt.Errorf("failed to template yaml,reason: %v,%v,%v", stdout, stderr, err)
	}
	if !force {
		allCRDNames, err := detectCRDs(tmpFile)
		if err != nil {
			return err
		}
		crdNames, err := getInstalledCRDs(allCRDNames)
		if err != nil {
			return err
		}
		if err := CheckRunningJobs(crdNames); err != nil {
			return err
		}
	}
	stdout, stderr, _ = execCommand([]string{"arena-helm", "del", "arena-artifacts", "-n", *arenaNamespace})
	fmt.Printf("%v,%v\n", stdout, stderr)
	stdout, stderr, _ = execCommand([]string{"arena-kubectl", "delete", "-f", tmpFile})
	fmt.Printf("%v,%v\n", stdout, stderr)
	stdout, stderr, err = execCommand([]string{"arena-kubectl", "delete", "ns", *arenaNamespace})
	if err != nil {
		return fmt.Errorf("failed to delete namespace %v,reason: %v,%v,%v", *arenaNamespace, stdout, stderr, err)
	}
	fmt.Printf("Debug: succeed to delete namespace %v\n", *arenaNamespace)
	return nil
}

func parseFields(fileName string) ([]string, error) {
	contentBytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %v,reason: %v", fileName, err)
	}
	fields := map[string]interface{}{}
	err = yaml.Unmarshal(contentBytes, &fields)
	if err != nil {
		return nil, fmt.Errorf("failed to parse fields from file %v,reason: %v", fileName, err)
	}
	options := []string{}
	for f := range fields {
		options = append(options, fmt.Sprintf("--set %v.enabled=true", f))
	}
	return options, nil
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
		jobName := fmt.Sprintf("%v %v/%v", t[0], crdName, t[1])
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
	if CheckFileExist(path.Join(currentPath, arenaChartName)) {
		return path.Join(currentPath, arenaChartName), nil
	}
	homeDir := os.Getenv("HOME")
	if CheckFileExist(path.Join(homeDir, "charts", arenaChartName)) {
		return path.Join(homeDir, "charts", arenaChartName), nil
	}
	if CheckFileExist(path.Join("/charts", arenaChartName)) {
		return path.Join("/charts", arenaChartName), nil
	}
	return "", fmt.Errorf("not found arena-artifacts directory,please download it from https://github.com/kubeflow/arena/tree/master/arena-artifacts")
}

func detectCRDs(fileName string) ([]string, error) {
	contentBytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %v,reason: %v", fileName, err)
	}
	type MetaInfo struct {
		Name *string `yaml:"name"`
	}
	crdNames := []string{}
	for _, c := range strings.Split(string(contentBytes), "---\n") {
		if !strings.Contains(c, "kind: CustomResourceDefinition") {
			continue
		}
		crdName := ""

		err := yaml.Unmarshal([]byte(c), &struct {
			Metadata *MetaInfo `yaml:"metadata"`
		}{
			Metadata: &MetaInfo{Name: &crdName},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to parse CRD name for yaml file,reason: %v", err)
		}
		fmt.Printf("Debug: succeed to parse CRD name %v from yaml file %v\n", crdName, fileName)
		crdNames = append(crdNames, crdName)
	}
	return crdNames, nil
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
		return nil, fmt.Errorf("exit status: %v,reason: %v", err, stderr)
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

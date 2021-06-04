package main

import (
	"bufio"
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
	forceDelete = flag.Bool("force", false, "force delete the Custom Resource Instances")
	quiet       = flag.Bool("quiet", false, "quiet for all choices")
	manifest    = flag.String("manifest-dir", "", "specify the kubernetes-artifacts directory")
)

func main() {
	flag.Parse()
	force := *forceDelete
	if !force && !*quiet {
		force = getAnswer()
	}
	manifest, err := detectManifests(manifest)
	if err != nil {
		fmt.Printf("Error: failed to detect manifest directory,reason: %v\n", err)
		os.Exit(2)
	}
	resources, fileNames := detectK8SResources(manifest, []string{"CustomResourceDefinition"})
	if err := deleteCRDs(resources["CustomResourceDefinition"], force); err != nil {
		fmt.Printf("Error: failed to delete CRDs,reason: %v\n", err)
		os.Exit(3)
	}
	deleteK8sResources(fileNames)
	_, stderr, err := execCommand([]string{"arena-kubectl", "delete", "ns", "arena-system"})
	if err != nil && !strings.Contains(stderr, `namespaces "arena-system" not found`) {
		fmt.Printf("Error: failed to delete namespace arena-system,reason: %v,%v", err, stderr)
		os.Exit(1)
	} else {
		fmt.Printf("Debug: succeed to delete namespace arena-system\n")
	}
	execCommand([]string{"rm", "-rf", "/charts"})
	execCommand([]string{"rm", "-rf", "~/charts"})
	execCommand([]string{"rm", "-rf", "/usr/local/bin/arena"})
	if err := removeLines([]string{"source <(arena completion bash)"}); err != nil {
		fmt.Printf("Error: failed to remove line 'source <(arena completion bash)' from ~/bashrc or ~/.zshrc\n")
		os.Exit(4)
	}
	return
}

func getAnswer() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please confirm whether to delete the running Custom Resource Instances(eg: tfjob,mpijob)[Y/N]: ")
	text, _ := reader.ReadString('\n')
	text = strings.Trim(strings.Trim(text, "\n"), " ")
	t := strings.ToLower(text)
	if t != "y" && t != "n" {
		fmt.Printf("Error: Unknown option %v,please input [Y/N]\n", text)
		os.Exit(1)
	}
	if t == "n" {
		return false
	}
	return true
}

func deleteCRDs(crdContents []string, force bool) error {
	readyToDelelte := []string{}
	crdNames, err := getInstallCRDs()
	if err != nil {
		return err
	}
	for _, crdName := range crdNames {
		found := false
		for _, content := range crdContents {
			if strings.Contains(content, fmt.Sprintf("name: %v", crdName)) {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		if force {
			readyToDelelte = append(readyToDelelte, crdName)
			continue
		}
		crs, err := getCR(crdName)
		if err != nil {
			return err
		}
		if len(crs) != 0 {
			return fmt.Errorf("there is some custom resource instances for %v,please use 'arena-kubectl get %v --all-namespaces' to check it", crdName, crdName)
		}
		readyToDelelte = append(readyToDelelte, crdName)
	}
	if len(readyToDelelte) == 0 {
		fmt.Printf("Debug: not found CRDs,skip to delete them\n")
		return nil
	}
	stdout, stderr, err := execCommand([]string{"arena-kubectl", "delete", "crd", strings.Join(readyToDelelte, " ")})
	if err != nil {
		return fmt.Errorf("%v,%v\n", stderr, err)
	}
	fmt.Println(stdout)
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

func getCR(crdName string) ([]string, error) {
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
		crs = append(crs, items[0])
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
				for _, content := range contents {
					resourceFiles[key] = append(resourceFiles[key], content)
				}
			}
			for _, name := range names {
				fileNames = append(fileNames, name)
			}
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

func getInstallCRDs() ([]string, error) {
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

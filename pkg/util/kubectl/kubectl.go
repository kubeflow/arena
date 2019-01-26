package kubectl

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/kubeflow/arena/pkg/types"
	log "github.com/sirupsen/logrus"
)

var kubectlCmd = []string{"kubectl"}

/**
* dry-run creating kubernetes App Info for delete in future
* Exec /usr/local/bin/kubectl, [create --dry-run -f /tmp/values313606961 --namespace default]
**/

func SaveAppInfo(fileName, namespace string) (configFileName string, err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return "", err
	}

	args := []string{"create", "--dry-run", "--namespace", namespace, "-f", fileName}
	out, err := kubectl(args)
	output := string(out)
	result := []string{}

	// fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
		log.Errorf("The output is %s\n", output)
		return "", err
	}

	// 1. generate the config file
	configFile, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", configFile.Name(), err)
		return "", err
	}

	configFileName = configFile.Name()
	log.Debugf("Save the config file %s", configFileName)

	// 2. save app types to config file
	lines := strings.Split(output, "\n")
	log.Debugf("dry run result: %v", lines)

	for _, line := range lines {
		line := strings.TrimSpace(line)
		cols := strings.Fields(line)
		log.Debugf("cols: %s, %d", cols, len(cols))
		if len(cols) == 0 {
			continue
		}
		result = append(result, cols[0])
	}

	data := []byte(strings.Join(result, "\n"))
	defer configFile.Close()
	_, err = configFile.Write(data)
	if err != nil {
		log.Errorf("Failed to write %v to %s due to %v", data, configFileName, err)
		return configFileName, err
	}

	return configFileName, nil
}

/**
* Delete kubernetes config to uninstall app
* Exec /usr/local/bin/kubectl, [delete -f /tmp/values313606961 --namespace default]
**/
func UninstallApps(fileName, namespace string) (err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return err
	}

	args := []string{"delete", "--namespace", namespace, "-f", fileName}
	out, err := kubectl(args)

	fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

/**
* Delete kubernetes config to uninstall app
* Exec /usr/local/bin/kubectl, [delete -f /tmp/values313606961 --namespace default]
**/
func UninstallAppsWithAppInfoFile(appInfoFile, namespace string) (err error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return err
	}

	if _, err = os.Stat(appInfoFile); err != nil {
		return err
	}

	args := []string{"cat", appInfoFile, "|", "xargs",
		binary, "delete"}

	log.Debugf("Exec bash -c %v", args)

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}
	out, err := cmd.Output()
	fmt.Printf("%s", string(out))

	if err != nil {
		log.Errorf("Failed to execute %s, %v with %v", "bash -c", args, err)
	}

	return err
}

/**
* Apply kubernetes config to install app
* Exec /usr/local/bin/kubectl, [apply -f /tmp/values313606961 --namespace default]
**/
func InstallApps(fileName, namespace string) (err error) {
	if _, err = os.Stat(fileName); os.IsNotExist(err) {
		return err
	}

	args := []string{"apply", "--namespace", namespace, "-f", fileName}
	out, err := kubectl(args)

	fmt.Printf("%s", string(out))
	if err != nil {
		log.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

/**
* This name should be <job-type>-<job-name>
* create configMap by using name, namespace and configFile
**/
func CreateAppConfigmap(name, trainingType, namespace, configFileName, appInfoFileName, chartName, chartVersion string) (err error) {
	if _, err = os.Stat(configFileName); os.IsNotExist(err) {
		return err
	}

	if _, err = os.Stat(appInfoFileName); os.IsNotExist(err) {
		return err
	}

	args := []string{"create", "configmap", fmt.Sprintf("%s-%s", name, trainingType),
		"--namespace", namespace,
		fmt.Sprintf("--from-file=%s=%s", "values", configFileName),
		fmt.Sprintf("--from-file=%s=%s", "app", appInfoFileName),
		fmt.Sprintf("--from-literal=%s=%s", chartName, chartVersion)}
	out, err := kubectl(args)

	fmt.Printf("%s", string(out))
	if err != nil {
		log.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

/**
*
* delete configMap by using name, namespace
**/
func DeleteAppConfigMap(name, namespace string) (err error) {
	args := []string{"delete", "configmap", name, "--namespace", namespace}
	out, err := kubectl(args)

	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
		log.Debugf("%s", string(out))
	} else {
		fmt.Printf("%s", string(out))
	}

	return err
}

/**
*
* get configMap by using name, namespace
**/
func CheckAppConfigMap(name, namespace string) (found bool) {
	args := []string{"get", "configmap", name, "--namespace", namespace}
	out, err := kubectl(args)

	if err != nil {
		log.Debugf("Failed to execute %s, %v with %v", "kubectl", args, err)
		log.Debugf("%s", string(out))
	} else {
		log.Debugf("%s", string(out))
		found = true
	}

	return found
}

/**
*
* save the key of configMap into a file
**/
func SaveAppConfigMapToFile(name, key, namespace string) (fileName string, err error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return "", err
	}

	file, err := ioutil.TempFile(os.TempDir(), name)
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", file.Name(), err)
		return fileName, err
	}
	fileName = file.Name()

	args := []string{binary, "get", "configmap", name,
		"--namespace", namespace,
		fmt.Sprintf("-o=jsonpath='{.data.%s}'", key),
		">", fileName}
	log.Debugf("Exec bash -c %s", strings.Join(args, " "))

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}
	out, err := cmd.Output()
	fmt.Printf("%s", string(out))

	if err != nil {
		return fileName, fmt.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}
	return fileName, err
}

func kubectl(args []string) ([]byte, error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return nil, err
	}

	// 1. prepare the arguments
	// args := []string{"create", "configmap", name, "--namespace", namespace, fmt.Sprintf("--from-file=%s=%s", name, configFileName)}
	log.Debugf("Exec %s, %v", binary, args)

	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}

	// return syscall.Exec(cmd, args, env)
	// 2. execute the command
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	return cmd.CombinedOutput()
}

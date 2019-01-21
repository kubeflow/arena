package kubectl

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/kubeflow/arena/types"
	log "github.com/sirupsen/logrus"
)

var kubectlCmd = []string{"kubectl"}

/**
* Apply kubernetes config
* Exec /usr/local/bin/kubectl, [apply -f /tmp/values313606961 --namespace default]
**/
func InstallApps(namespace, fileName string) error {
	args := []string{"apply", "--namespace", namespace, "-f", fileName}
	out, err := kubectl(args, fileName)

	fmt.Println("")
	fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

/**
* This name should be <job-type>-<job-name>
* create configMap by using name, namespace and configFile
**/
func CreateConfigmap(name, namespace, configFileName string) error {
	args := []string{"create", "configmap", name, "--namespace", namespace, fmt.Sprintf("--from-file=%s=%s", name, configFileName)}
	out, err := kubectl(args, configFileName)

	fmt.Println("")
	fmt.Printf("%s\n", string(out))
	if err != nil {
		log.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	return err
}

/**
*
* save configMap into a file
**/
func SaveConfigMapToFile(name, namespace string) (fileName string, err error) {
	args := []string{"get", "configmap", name, "--namespace", namespace, fmt.Sprintf("jsonpath='{.%s.test}'", name)}
	data, err := kubectl(args, "")

	if err != nil {
		return "", fmt.Errorf("Failed to execute %s, %v with %v", "kubectl", args, err)
	}

	file, err := ioutil.TempFile(os.TempDir(), name)
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", file.Name(), err)
		return fileName, err
	}

	fileName = file.Name()
	log.Debugf("Save the values file %s", fileName)

	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		log.Errorf("Failed to write %v to %s due to %v", data, file.Name(), err)
	}
	return fileName, err
}

func kubectl(args []string, fileName string) ([]byte, error) {
	binary, err := exec.LookPath(kubectlCmd[0])
	if err != nil {
		return nil, err
	}

	// 1. check if the template file exists
	if len(fileName) > 0 {
		if _, err = os.Stat(fileName); err != nil {
			return nil, err
		}
	}

	// 2. prepare the arguments
	// args := []string{"create", "configmap", name, "--namespace", namespace, fmt.Sprintf("--from-file=%s=%s", name, configFileName)}
	log.Debugf("Exec %s, %v", binary, args)

	env := os.Environ()
	if types.KubeConfig != "" {
		env = append(env, fmt.Sprintf("KUBECONFIG=%s", types.KubeConfig))
	}

	// return syscall.Exec(cmd, args, env)
	// 3. execute the command
	cmd := exec.Command(binary, args...)
	cmd.Env = env
	return cmd.CombinedOutput()
}

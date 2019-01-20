package helm

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

/*
* Generate value file
 */
func GenerateValueFile(value interface{}) (fileName string, err error) {
	// 1. generate the template file
	valueFile, err = ioutil.TempFile(os.TempDir(), "values")
	if err != nil {
		log.Errorf("Failed to create tmp file %v due to %v", valueFile.Name(), err)
		return templateFileName, err
	}

	valueFileName = valueFile.Name()
	log.Debugf("Save the values file %s", valueFileName)

	// 2. dump the object into the template file
	err = toYaml(values, valueFile)
	return valueFileName, err
}

/**
* generate helm template without tiller: helm template -f values.yaml chart_name
* Exec /usr/local/bin/helm, [template -f /tmp/values313606961 --namespace default --name hj /charts/tf-horovod]
* returns generated template file: templateFileName
 */
func GenerateHelmTemplate(name string, namespace string, valueFileName string, chartName string) (templateFileName string, err error) {
	tempName := fmt.Sprintf("/tmp/%s.yaml", name)
	templateFile, err := ioutil.TempFile(os.TempDir(), tempName)
	if err != nil {
		return templateFileName, err
	}
	templateFileName = templateFile.Name()

	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return templateFileName, err
	}

	// 3. check if the chart file exists
	// if strings.HasPrefix(chartName, "/") {
	if _, err = os.Stat(chartName); err != nil {
		return templateFileName, err
	}
	// }

	// 4. prepare the arguments
	args := []string{binary, "template", "-f", valueFile.Name(),
		"--namespace", namespace,
		"--name", name, chartName, ">", templateFileName}

	log.Debugf("Exec bash -c %v", args)

	// return syscall.Exec(cmd, args, env)
	// 5. execute the command
	log.Debugf("Generating template  %v", args)
	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	// cmd.Env = env
	out, err := cmd.CombinedOutput()
	fmt.Println("")
	fmt.Printf("%s\n", string(out))
	if err != nil {
		return templateFileName, fmt.Errorf("Failed to execute %s, %v with %v", binary, args, err)
	}

	// // 6. clean up the value file if needed
	// if log.GetLevel() != log.DebugLevel {
	// 	err = os.Remove(valueFileName)
	// 	if err != nil {
	// 		log.Warnf("Failed to delete %s due to %v", valueFileName, err)
	// 	}
	// }

	return templateFileName, nil
}

/**
* Check the chart version by given the chart directory
* helm inspect chart /charts/tf-horovod
 */

func GetChartVersion(chart string) (version string, err error) {
	binary, err := exec.LookPath(helmCmd[0])
	if err != nil {
		return "", err
	}

	// 1. check if the chart file exists, if it's it's unix path, then check if it's exist
	// if strings.HasPrefix(chart, "/") {
	if _, err = os.Stat(chart); err != nil {
		return "", err
	}
	// }

	// 2. prepare the arguments
	args := []string{binary, "inspect", "chart", chart,
		"|", "grep", "version:"}
	log.Debugf("Exec bash -c %v", args)

	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("Failed to find version when executing %s, result is %s", args, out)
	}
	fields := strings.Split(lines[0], ":")
	if len(fields) != 2 {
		return "", fmt.Errorf("Failed to find version when executing %s, result is %s", args, out)
	}

	version = strings.TrimSpace(fields[1])
	return version, nil
}

// Copyright 2018 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/kubeflow/arena/util"
	"github.com/kubeflow/arena/util/helm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"io/ioutil"
	"bytes"
)

var (
	tfservingChart        = "./charts/tfserving"
	defaultTfServingImage = "tensorflow/serving:latest"
)

func NewServingTensorFlowCommand() *cobra.Command {
	var (
		serveTensorFlowArgs ServeTensorFlowArgs
	)

	var command = &cobra.Command{
		Use:     "tensorflow",
		Short:   "Submit tensorflow serving job to deploy and serve machine learning models.",
		Aliases: []string{"tf"},
		Run: func(cmd *cobra.Command, args []string) {
			/*if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}*/

			util.SetLogLevel(logLevel)
			setupKubeconfig()
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = ensureNamespace(client, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = serveTensorFlow(args, &serveTensorFlowArgs, client)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	serveTensorFlowArgs.addServeCommonFlags(command)
	//submitArgs.addSyncFlags(command)

	// TFServingJob
	command.Flags().StringVar(&serveTensorFlowArgs.ModelConfigFile, "modelConfigFile", "", "Corresponding with --model_config_file in tensorflow serving")
	command.Flags().StringVar(&serveTensorFlowArgs.VersionPolicy, "versionPolicy", "", "support latest, latest:N, specific:N, all")

	return command
}

type ServeTensorFlowArgs struct {
	VersionPolicy          string `yaml:"versionPolicy"`   // --versionPolicy
	ModelConfigFile        string `yaml:"modelConfigFile"` // --modelConfigFile
	ModelConfigFileContent string `yaml:"modelConfigFileContent"`

	ServeArgs `yaml:",inline"`

	ModelServiceExists bool `yaml:"modelServiceExists"` // --modelServiceExists
}

func (serveTensorFlowArgs *ServeTensorFlowArgs) preprocess(client *kubernetes.Clientset, args []string) (err error) {
	serveTensorFlowArgs.Command = strings.Join(args, " ")

	if serveTensorFlowArgs.ModelConfigFile == "" {
		// need to validate modelName, modelPath and versionPolicy if not specify modelConfigFile
		// 1. validate modelName
		err := serveTensorFlowArgs.ServeArgs.validateModelName()
		if err != nil {
			return err
		}
		//2. validate modelPath
		dataDir, err := ParseBasePath(serveTensorFlowArgs.ModelPath)
		if err != nil {
			return fmt.Errorf("modelPath[%s] has wrong content: %s", serveTensorFlowArgs.ModelPath, err)
		}
		serveTensorFlowArgs.DataDirs = append(serveTensorFlowArgs.DataDirs, dataDir)
		//3. validate versionPolicy
		err = serveTensorFlowArgs.validateVersionPolicy()
		if err != nil {
			return err
		}
		//populate content according to CLI parameters
		serveTensorFlowArgs.ModelConfigFileContent = generateModelConfigFileContent(*serveTensorFlowArgs)

	} else {
		//populate content from modelConfigFile
		log.Infof("modelConfigFile is specified, so ignore --modelName and --modelPath", serveTensorFlowArgs.ModelConfigFile)
		modelConfigFileContentBytes, err := ioutil.ReadFile(serveTensorFlowArgs.ModelConfigFile)
		if err != nil {
			return fmt.Errorf("cannot read the modelConfigFile[%s]: %s", serveTensorFlowArgs.ModelConfigFile, err)
		}
		modelconfigString := string(modelConfigFileContentBytes)
		log.Debugf("The content of modelConfigFile[%s] is: %s", serveTensorFlowArgs.ModelConfigFile, modelconfigString)

		newModelconfigString, dataDirs, err := populateModelConfig(modelconfigString)
		if err != nil {
			return fmt.Errorf("modelConfigFile[%s] has wrong content: %s", serveTensorFlowArgs.ModelConfigFile, err)
		}

		serveTensorFlowArgs.ModelConfigFileContent = newModelconfigString
		serveTensorFlowArgs.DataDirs = dataDirs

		log.Debugf("modelConfigFileContent:%s", serveTensorFlowArgs.ModelConfigFileContent)
		log.Debugf("dataDirs:%s", serveTensorFlowArgs.DataDirs)
	}
	//validate Istio enablement
	err = serveTensorFlowArgs.ServeArgs.validateIstioEnablement()
	if err != nil {
		return err
	}

	// populate environment variables
	if len(envs) > 0 {
		serveTensorFlowArgs.Envs = transformSliceToMap(envs, "=")
	}

	modelServiceExists, err := checkServiceExists(client, namespace, serveTensorFlowArgs.ServiceName)
	serveTensorFlowArgs.ModelServiceExists = modelServiceExists

	return nil
}

func populateModelConfig(originalContent string) (string, []dataDirVolume, error) {
	basePathFiled := "base_path:"
	lengthOfBasePathField := len(basePathFiled)
	count := strings.Count(originalContent, basePathFiled)
	log.Debugf("model config count: %d", count)
	dataDirs := []dataDirVolume{}

	if count == 0 {
		return originalContent, dataDirs, nil
	}
	index := strings.Index(originalContent, basePathFiled) + lengthOfBasePathField
	configString := originalContent[0:index]
	tempString := originalContent[index:]
	log.Debugf("tempString:%s", tempString)
	tempString = strings.Trim(tempString, " ")
	index2 := strings.Index(tempString, "\"")
	if index2 != 0 {
		return "", dataDirs, fmt.Errorf("no available model config is provided: %s", originalContent)
	}
	tempString = tempString[1:]
	index2 = strings.Index(tempString, "\"")
	if index2 <= 0 {
		return "", dataDirs, fmt.Errorf("no available model config is provided: %s", originalContent)
	}
	originalBasePath := tempString[0:index2]
	log.Debugf("originalBasePath: %s", originalBasePath)
	dataDir, err := ParseBasePath(originalBasePath)
	if err == nil {
		updatedBasePath := dataDir.ContainerPath + dataDir.HostPath
		log.Debugf("updatedBasePath: %s", updatedBasePath)
		configString += "\"" + updatedBasePath + "\""
		dataDirs = append(dataDirs, dataDir)
	} else {
		return "", dataDirs, fmt.Errorf("modelConfigFile has wrong content for filed basepath: %s", err)
	}
	tempString = tempString[index2+1:]
	log.Debugf("tempString:%s", tempString)

	configString_, dataDirs_, err := populateModelConfig(tempString)
	if err == nil && configString_ != "" {
		configString += configString_
		for i := 0; i < len(dataDirs_); i++ {
			dataDirs = append(dataDirs, dataDirs_[i])
		}
	}
	log.Debugf("configString:%s", configString)
	return configString, dataDirs, err
}

func checkServiceExists(client *kubernetes.Clientset, namespace string, name string) (bool, error) {
	service, err := client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if service == nil {
		return false, nil
	}
	return true, nil
}

func (serveTensorFlowArgs *ServeTensorFlowArgs) validateVersionPolicy() error {
	// validate version policy
	if serveTensorFlowArgs.VersionPolicy == "" {
		serveTensorFlowArgs.VersionPolicy = "latest"
	}
	versionPolicyName := strings.Split(serveTensorFlowArgs.VersionPolicy, ":")
	switch versionPolicyName[0] {
	case "latest", "specific", "all":
		log.Debug("Support TensorFlow Serving Version Policy: latest, specific, all.")
		//serveTensorFlowArgs.ServeArgs.ModelVersion = strings.Replace(serveTensorFlowArgs.VersionPolicy, ":", "-", -1)
	default:
		return fmt.Errorf("UnSupport TensorFlow Serving Version Policy: %s", versionPolicyName[0])
	}

	return nil
}

func serveTensorFlow(args []string, serveTensorFlowArgs *ServeTensorFlowArgs, client *kubernetes.Clientset) (err error) {
	err = serveTensorFlowArgs.preprocess(client, args)
	if err != nil {
		return err
	}

	exist, err := helm.CheckRelease(name)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the job %s is already exist, please delete it firstly. use 'arena delete %s'", name, name)
	}

	//log.Debugf("ModelVersion:%s", serveTensorFlowArgs.ModelVersion)
	name = serveTensorFlowArgs.ServiceName
	if serveTensorFlowArgs.ServiceVersion != "" {
		name += "-" + serveTensorFlowArgs.ServiceVersion
	}

	return helm.InstallRelease(name, namespace, serveTensorFlowArgs, tfservingChart)
}

func generateModelConfigFileContent(serveTensorFlowArgs ServeTensorFlowArgs) string {
	modelName := serveTensorFlowArgs.ModelName
	versionPolicy := serveTensorFlowArgs.VersionPolicy
	mountPath := serveTensorFlowArgs.DataDirs[0].ContainerPath
	modelPathInPvc := serveTensorFlowArgs.DataDirs[0].HostPath
	versionPolicyName := strings.Split(versionPolicy, ":")

	var buffer bytes.Buffer
	buffer.WriteString("model_config_list: { config: { name: ")
	buffer.WriteString("\"" + modelName + "\" base_path: \"")
	buffer.WriteString(mountPath + modelPathInPvc + "\" model_platform: \"")
	buffer.WriteString("tensorflow\" model_version_policy: { ")
	switch versionPolicyName[0] {
	case "all":
		buffer.WriteString(versionPolicyName[0] + ": {} } } }")
	case "specific":
		if len(versionPolicyName) > 1 {
			buffer.WriteString(versionPolicyName[0] + ": { " + "versions: " + versionPolicyName[1] + " } } } }")
		} else {
			log.Errorf("[specific] version policy scheme should be specific:N")
		}
	case "latest":
		if len(versionPolicyName) > 1 {
			buffer.WriteString(versionPolicyName[0] + ": { " + "num_versions: " + versionPolicyName[1] + " } } } }")
		} else {
			buffer.WriteString(versionPolicyName[0] + ": { " + "num_versions: 1 } } } }")
		}
	default:
		log.Errorf("UnSupport TensorFlow Serving Version Policy: %s", versionPolicyName[0])
		buffer.Reset()
	}

	result := buffer.String()
	log.Debugf("generateModelConfigFileContent: \n%s", result)

	return result
}

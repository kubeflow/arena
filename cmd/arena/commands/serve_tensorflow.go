package commands

import (
	"fmt"
	"os"
	"strings"

	"bytes"
	"github.com/kubeflow/arena/util"
	"github.com/kubeflow/arena/util/helm"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var (
	tfserving_chart       = "/charts/tfserving"
	defaultTfServingImage = "tensorflow/serving:1.8.0-devel-gpu"
)

func NewServeTensorFlowCommand() *cobra.Command {
	var (
		serveTensorFlowArgs ServeTensorFlowArgs
	)

	var command = &cobra.Command{
		Use:     "tensorflow",
		Short:   "Submit tensorflow serving job to deploy online model.",
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

			err = serverTensorFlow(args, &serveTensorFlowArgs)
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
	command.Flags().StringVar(&serveTensorFlowArgs.VersionPolicy, "versionPolicy", "latest", "support latest, latest:N, specific:N, all")

	return command
}

type ServeTensorFlowArgs struct {
	VersionPolicy          string `yaml:"versionPolicy"`   // --versionPolicy
	ModelConfigFile        string `yaml:"modelConfigFile"` // --modelConfigFile
	ModelConfigFileContent string `yaml:"modelConfigFileContent"`

	ServeArgs `yaml:",inline"`
}

func (serveTensorFlowArgs *ServeTensorFlowArgs) prepare(args []string) (err error) {
	serveTensorFlowArgs.Command = strings.Join(args, " ")

	err = serveTensorFlowArgs.transform()
	if err != nil {
		return err
	}

	err = serveTensorFlowArgs.check()
	if err != nil {
		return err
	}

	if len(envs) > 0 {
		serveTensorFlowArgs.Envs = transformSliceToMap(envs, "=")
	}

	// read --model-config-file content, write to values.yaml in tfserving chart
	if serveTensorFlowArgs.ModelConfigFile != "" {
		modelConfigFileContentBytes, err := ioutil.ReadFile(serveTensorFlowArgs.ModelConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("The content of %s is: %s", serveTensorFlowArgs.ModelConfigFile, string(modelConfigFileContentBytes))
		tmpstr := strings.Replace(string(modelConfigFileContentBytes), "\n", " ", -1)
		serveTensorFlowArgs.ModelConfigFileContent = strings.Replace(tmpstr, "\t", " ", -1)
		log.Debugf("The content of ModelConfigFileContent is: %s", serveTensorFlowArgs.ModelConfigFileContent)
	}

	// generate model-config-file content according modelName, modelPath, versionPolicy
	if serveTensorFlowArgs.VersionPolicy != "" {
		serveTensorFlowArgs.ModelConfigFileContent = generateModelConfigFileContent(serveTensorFlowArgs.ModelName, serveTensorFlowArgs.MountPath, serveTensorFlowArgs.VersionPolicy)
	}

	// valid modelConfigFileContent
	// TODO

	// if modePath value is hdfs path
	// TODO

	return nil
}

func (serveTensorFlowArgs ServeTensorFlowArgs) check() error {
	// check name
	err := serveTensorFlowArgs.ServeArgs.check()
	if err != nil {
		return err
	}

	// check version policy
	versionPolicyName := strings.Split(serveTensorFlowArgs.VersionPolicy, ":")
	switch versionPolicyName[0] {
	case "latest", "specific", "all":
		log.Debug("Support TensorFlow Serving Version Policy: latest, specific, all.")
	default:
		log.Errorf("UnSupport TensorFlow Serving Version Policy: %s", versionPolicyName[0])
	}

	if serveTensorFlowArgs.VersionPolicy != "" {
		if serveTensorFlowArgs.ModelName == "" {
			log.Error("versionPolicy has been set %s, modelName cannt be none.")
		}
	}

	// check model-name
	if serveTensorFlowArgs.ModelName != "" {
		if serveTensorFlowArgs.ModelPath == "" {
			return fmt.Errorf("If modelName: %s has been set, the modelPath must be set too.", serveTensorFlowArgs.ModelName)
		}
		if serveTensorFlowArgs.ModelConfigFile != "" {
			return fmt.Errorf("If modelName: %s has been set, modelConfigFile connt be set.", serveTensorFlowArgs.ModelName)
		}
	}

	// check model-path
	if serveTensorFlowArgs.ModelPath != "" {
		if serveTensorFlowArgs.ModelName == "" {
			return fmt.Errorf("If modelPath: %s has been set, the modelName must be set too.", serveTensorFlowArgs.ModelPath)
		}
		if serveTensorFlowArgs.ModelConfigFile != "" {
			return fmt.Errorf("If modelPath: %s has been set, modelConfigFile cannt be set.", serveTensorFlowArgs.ModelPath)
		}
	}

	// check model-config-file
	if serveTensorFlowArgs.ModelConfigFile != "" {
		if serveTensorFlowArgs.ModelName != "" || serveTensorFlowArgs.ModelPath != "" {
			return fmt.Errorf("If modelConfigFile: %s has been set, modelName or modelPath cannt be set.", serveTensorFlowArgs.ModelConfigFile)
		}
	}

	return nil
}

func (serveTensorFlowArgs *ServeTensorFlowArgs) transform() error {

	return serveTensorFlowArgs.ServeArgs.transform()

}

func serverTensorFlow(args []string, serveTensorFlowArgs *ServeTensorFlowArgs) (err error) {
	err = serveTensorFlowArgs.prepare(args)
	if err != nil {
		return err
	}

	exist, err := helm.CheckRelease(name)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the job %s is already exist, please delete it first. use 'arena delete %s'", name, name)
	}

	return helm.InstallRelease(name, namespace, serveTensorFlowArgs, tfserving_chart)
}

func generateModelConfigFileContent(modelName, mountPath, versionPolicy string) string {
	mountPath = strings.Trim(mountPath, " ")
	mountPath = strings.TrimRight(mountPath, "/")
	versionPolicyName := strings.Split(versionPolicy, ":")
	var buffer bytes.Buffer
	buffer.WriteString("model_config_list: { config: {name: \"")
	buffer.WriteString(modelName + "\" base_path: \"")
	buffer.WriteString(mountPath + "/" + modelName + "\" model_platform: \"")
	buffer.WriteString("tensorflow" + "\" model_version_policy: { ")
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
	log.Debugf("generateModelConfigFileContent: \n%s", buffer.String())

	return fmt.Sprintf(buffer.String())
}

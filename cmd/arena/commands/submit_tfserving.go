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
	defaultTfServingImage = "tensorflow/serving:1.5.0-devel-gpu"
)

func NewSubmitTFServingJobCommand() *cobra.Command {
	var (
		submitArgs submitTFServingJobArgs
	)

	var command = &cobra.Command{
		Use:     "tfserving",
		Short:   "Submit tfserving job to deploy a online model.",
		Aliases: []string{"tfserving"},
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

			err = submitTFServingJob(args, &submitArgs)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	submitArgs.addCommonFlags(command)
	//submitArgs.addSyncFlags(command)

	// TFServingJob
	command.Flags().IntVar(&submitArgs.Replicas, "replicas", 1, "TensorFLow serving replicas")
	command.Flags().StringVar(&submitArgs.ModelConfigFile, "modelConfigFile", "", "Corresponding with --model_config_file in tensorflow serving")
	command.Flags().StringVar(&submitArgs.ModelName, "modelName", "", "Corresponding with --model_name in tensorflow serving")
	command.Flags().StringVar(&submitArgs.ModelPath, "modelPath", "", "Corresponding with --model_path in tensorflow serving")
	command.Flags().IntVar(&submitArgs.Port, "port", 9000, "Corresponding with --port in tensorflow serving")
	command.Flags().StringVar(&submitArgs.VersionPolicy, "versionPolicy", "latest", "support latest, latest:N, specific:N, all")
	//command.Flags().StringVar(&submitArgs.Vhost, "vhost", "", "")
	command.Flags().StringVar(&submitArgs.Cpu, "cpu", "", "the cpu resource to request for the tensorflow serving container")
	command.Flags().StringVar(&submitArgs.Memory, "memory", "", "the memory resource to request for the tensorflow serving container")

	return command
}

type submitTFServingJobArgs struct {
	Replicas               int    `yaml:"replicas"`        // --replicas
	ModelName              string `yaml:"modelName"`       // --modelName
	ModelPath              string `yaml:"modelPath"`       // --modelPath
	Port                   int    `yaml:"port"`            // --pot
	VersionPolicy          string `yaml:"versionPolicy"`   // --versionPolicy
	ModelConfigFile        string `yaml:"modelConfigFile"` // --modelConfigFile
	ModelConfigFileContent string `yaml:"modelConfigFileContent"`
	// Vhost   		string 	`yaml:"vhost"`   		// --vhost
	Cpu    string `yaml:"cpu"`    // --cpu
	Memory string `yaml:"memory"` // --memory
	// for common args
	submitArgs `yaml:",inline"`
}

func (submitTFServingArgs *submitTFServingJobArgs) prepare(args []string) (err error) {
	submitTFServingArgs.Command = strings.Join(args, " ")

	err = submitTFServingArgs.transform()
	if err != nil {
		return err
	}

	err = submitTFServingArgs.check()
	if err != nil {
		return err
	}

	if len(envs) > 0 {
		submitTFServingArgs.Envs = transformSliceToMap(envs, "=")
	}

	// read --model-config-file content, write to values.yaml in chart
	if submitTFServingArgs.ModelConfigFile != "" {
		modelConfigFileContentBytes, err := ioutil.ReadFile(submitTFServingArgs.ModelConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugf("The content of %s is: %s", submitTFServingArgs.ModelConfigFile, string(modelConfigFileContentBytes))
		tmpstr := strings.Replace(string(modelConfigFileContentBytes), "\n", " ", -1)
		submitTFServingArgs.ModelConfigFileContent = strings.Replace(tmpstr, "\t", " ", -1)
		log.Debugf("The content of ModelConfigFileContent is: %s", submitTFServingArgs.ModelConfigFileContent)
	}

	// generate model-config-file content according modelName, modelPath, versionPolicy
	if submitTFServingArgs.VersionPolicy != "" {
		submitTFServingArgs.ModelConfigFileContent = generateModelConfigFileContent(submitTFServingArgs.ModelName, submitTFServingArgs.ModelPath, submitTFServingArgs.VersionPolicy)
	}

	// check modelConfigFileContent whether a valid json object
	/*if !json.Valid([]byte(submitTFServingArgs.ModelConfigFileContent)) {
		return fmt.Errorf("modelConfigFileContent is not a valid json object")
	}*/

	return nil
}

func (submitTFServingArgs submitTFServingJobArgs) check() error {
	// check name
	err := submitTFServingArgs.submitArgs.check()
	if err != nil {
		return err
	}

	// check version policy
	versionPolicyName := strings.Split(submitTFServingArgs.VersionPolicy, ":")
	switch versionPolicyName[0] {
	case "latest", "specific", "all":
		log.Debug("Support TensorFlow Serving Version Policy: latest, specific, all.")
	default:
		log.Errorf("UnSupport TensorFlow Serving Version Policy: %s", versionPolicyName[0])
	}

	if submitTFServingArgs.VersionPolicy != "" {
		if submitTFServingArgs.ModelName == "" {
			log.Error("versionPolicy has been set %s, modelName cannt be none.")
		}
	}

	// check model-name
	if submitTFServingArgs.ModelName != "" {
		if submitTFServingArgs.ModelPath == "" {
			return fmt.Errorf("If modelName: %s has been set, the modelPath must be set too.", submitTFServingArgs.ModelName)
		}
		if submitTFServingArgs.ModelConfigFile != "" {
			return fmt.Errorf("If modelName: %s has been set, modelConfigFile connt be set.", submitTFServingArgs.ModelName)
		}
	}

	// check model-path
	if submitTFServingArgs.ModelPath != "" {
		if submitTFServingArgs.ModelName == "" {
			return fmt.Errorf("If modelPath: %s has been set, the modelName must be set too.", submitTFServingArgs.ModelPath)
		}
		if submitTFServingArgs.ModelConfigFile != "" {
			return fmt.Errorf("If modelPath: %s has been set, modelConfigFile cannt be set.", submitTFServingArgs.ModelPath)
		}
	}

	// check model-config-file
	if submitTFServingArgs.ModelConfigFile != "" {
		if submitTFServingArgs.ModelName != "" || submitTFServingArgs.ModelPath != "" {
			return fmt.Errorf("If modelConfigFile: %s has been set, modelName or modelPath cannt be set.", submitTFServingArgs.ModelConfigFile)
		}
	}

	return nil
}

func (submitTFServingArgs *submitTFServingJobArgs) transform() error {
	if submitTFServingArgs.ModelName == "" {
		submitTFServingArgs.VersionPolicy = ""
	}
	return nil
}

func submitTFServingJob(args []string, submitArgs *submitTFServingJobArgs) (err error) {
	err = submitArgs.prepare(args)
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

	return helm.InstallRelease(name, namespace, submitArgs, tfserving_chart)
}

/*func generateModelConfigFileContent(modelName, modelPath, versionPolicy string) string {
	versionPolicyName := strings.Split(versionPolicy, ":")
	var buffer bytes.Buffer
	buffer.WriteString("model_config_list: {\n\tconfig: {\n\t\tname: \"")
	buffer.WriteString(modelName + "\",\n\t\tbase_path: \"")
	buffer.WriteString(modelPath + "\",\n\t\tmodel_platform: \"")
	buffer.WriteString("tensorflow" + "\",\n\t\tmodel_version_policy: {\n\t\t\t")
	switch versionPolicyName[0] {
	case "all":
		buffer.WriteString(versionPolicyName[0] + ": {}\n\t\t}\n\t}\n}")
	case "specific":
		if len(versionPolicyName) > 1 {
			buffer.WriteString(versionPolicyName[0] + ": {\n\t\t\t\t" + "versions: " + versionPolicyName[1] + "\n\t\t\t}\n\t\t}\n\t}\n}")
		} else {
			log.Errorf("[specific] version policy scheme should be specific:N")
		}
	case "latest":
		if len(versionPolicyName) > 1 {
			buffer.WriteString(versionPolicyName[0] + ": {\n\t\t\t\t" + "num_versions: " + versionPolicyName[1] + "\n\t\t\t}\n\t\t}\n\t}\n}")
		} else {
			buffer.WriteString(versionPolicyName[0] + ": {\n\t\t\t\t" + "num_versions: 1\n\t\t\t}\n\t\t}\n\t}\n}")
		}
	default:
		log.Errorf("UnSupport TensorFlow Serving Version Policy: %s", versionPolicyName[0])
		buffer.Reset()
	}
	log.Debugf("generateModelConfigFileContent: \n%s", buffer.String())

	return fmt.Sprintf(buffer.String())
}*/

func generateModelConfigFileContent(modelName, modelPath, versionPolicy string) string {
	versionPolicyName := strings.Split(versionPolicy, ":")
	var buffer bytes.Buffer
	buffer.WriteString("model_config_list: { config: {name: \"")
	buffer.WriteString(modelName + "\" base_path: \"")
	buffer.WriteString(modelPath + "\" model_platform: \"")
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

//model_config_list: { config: { name: "mnist", base_path: "/tmp/monitored/_model", model_platform: "tensorflow",	model_version_policy: { all: {}	} } }

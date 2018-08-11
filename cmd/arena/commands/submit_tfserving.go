package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/kubeflow/arena/util"
	"github.com/kubeflow/arena/util/helm"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

var (
	tfserving_chart = "/charts/tfserving"
	defaultTfServingImage = "tensorflow/serving:1.8.0-devel-gpu"
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
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}

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
	command.Flags().IntVar(&submitArgs.Replicas, "replicas", 1, "")
	command.Flags().StringVar(&submitArgs.ModelConfigFile, "modelConfigFile", "", "")
	command.Flags().StringVar(&submitArgs.ModelName, "modelName", "", "")
	command.Flags().StringVar(&submitArgs.ModelPath, "modelPath", "", "")
	command.Flags().IntVar(&submitArgs.Port, "port", 9000, "")
	command.Flags().StringVar(&submitArgs.VersionPolicy, "versionPolicy", "latest", "")
	command.Flags().StringVar(&submitArgs.Vhost, "vhost", "", "")
	command.Flags().StringVar(&submitArgs.Cpu, "cpu", "", "")
	command.Flags().StringVar(&submitArgs.Memory, "memory", "", "")

	return command
}

type submitTFServingJobArgs struct {
	Replicas		int		`yaml:"replicas"`		// --replicas
	ModelName		string	`yaml:"modelName"`    	// --modelName
	ModelPath		string	`yaml:"modelPath"`		// --modelPath
	Port     		int		`yaml:"port"`     		// --pot
	VersionPolicy 	string	`yaml:"versionPolicy"`	// --versionPolicy
	ModelConfigFile	string	`yaml:"modelConfigFile"`// --modelConfigFile
	ModelConfigFileContent string `yaml:"modelConfigFileConteng"`
	Vhost   		string 	`yaml:"vhost"`   		// --vhost
	Cpu				string	`yaml:"cpu"`			// --cpu
	Memory			string	`yaml:""`				// --memory
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
		submitTFServingArgs.ModelConfigFileContent = string(modelConfigFileContentBytes)
	}

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

	// check model-name
	if submitTFServingArgs.ModelName != "" {
		if submitTFServingArgs.ModelPath == "" {
			return fmt.Errorf("If model-name: %s has been set, the model-path must be set too.", submitTFServingArgs.ModelName)
		}
		if  submitTFServingArgs.ModelConfigFile != "" {
			return fmt.Errorf("If model-name: %s has been set, model-config-file connt be set.", submitTFServingArgs.ModelName)
		}
	}

	// check model-path
	if submitTFServingArgs.ModelPath != "" {
		if submitTFServingArgs.ModelName == "" {
			return fmt.Errorf("If model-path: %s has been set, the model-name must be set too.", submitTFServingArgs.ModelPath)
		}
		if  submitTFServingArgs.ModelConfigFile != "" {
			return fmt.Errorf("If model-path: %s has been set, model-config-file cannt be set.", submitTFServingArgs.ModelPath)
		}
	}

	// check model-config-file
	if submitTFServingArgs.ModelConfigFile != "" {
		if submitTFServingArgs.ModelName != "" || submitTFServingArgs.ModelPath != "" {
			return fmt.Errorf("If model-config-file: %s has been set, model-name or model-path cannt be set.", submitTFServingArgs.ModelConfigFile)
		}
	}

	return nil
}

func (submitTFServingArgs *submitTFServingJobArgs) transform() error {
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

package training

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/kubeflow/arena/pkg/apis/arenaclient"
	"github.com/kubeflow/arena/pkg/apis/training"
	"github.com/kubeflow/arena/pkg/apis/types"
)

// Model registration
func createRegisteredModelAndModelVersion(client *arenaclient.ArenaClient, job *training.Job, versionDescription string) (*types.RegisteredModel, *types.ModelVersion, error) {
	var (
		name        string
		description string
		tags        []*types.RegisteredModelTag

		versionTags []*types.ModelVersionTag
		source      string
	)
	tags = append(tags, &types.RegisteredModelTag{
		Key:   "createdBy",
		Value: "arena",
	})
	versionTags = append(versionTags, &types.ModelVersionTag{
		Key:   "createdBy",
		Value: "arena",
	})

	switch job.Type() {
	case types.TFTrainingJob:
		args := job.Args().(*types.SubmitTFJobArgs)
		name = args.ModelName
		if name == "" {
			return nil, nil, nil
		}
		for key, value := range args.Labels {
			versionTags = append(versionTags, &types.ModelVersionTag{
				Key:   key,
				Value: value,
			})
		}
		source = args.ModelSource
	case types.PytorchTrainingJob:
		args := job.Args().(*types.SubmitPyTorchJobArgs)
		name = args.ModelName
		if name == "" {
			return nil, nil, nil
		}
		for key, value := range args.Labels {
			versionTags = append(versionTags, &types.ModelVersionTag{
				Key:   key,
				Value: value,
			})
		}
		source = args.ModelSource
	case types.MPITrainingJob:
		args := job.Args().(*types.SubmitMPIJobArgs)
		name = args.ModelName
		if name == "" {
			return nil, nil, nil
		}
		for key, value := range args.Labels {
			versionTags = append(versionTags, &types.ModelVersionTag{
				Key:   key,
				Value: value,
			})
		}
		source = args.ModelSource
	case types.HorovodTrainingJob:
		args := job.Args().(*types.SubmitHorovodJobArgs)
		name = args.ModelName
		if name == "" {
			return nil, nil, nil
		}
		for key, value := range args.Labels {
			versionTags = append(versionTags, &types.ModelVersionTag{
				Key:   key,
				Value: value,
			})
		}
		source = args.ModelSource
	case types.VolcanoTrainingJob:
		return nil, nil, nil
	case types.ETTrainingJob:
		args := job.Args().(*types.SubmitETJobArgs)
		name = args.ModelName
		if name == "" {
			return nil, nil, nil
		}
		for key, value := range args.Labels {
			versionTags = append(versionTags, &types.ModelVersionTag{
				Key:   key,
				Value: value,
			})
		}
		source = args.ModelSource
	case types.SparkTrainingJob:
		return nil, nil, nil
	case types.DeepSpeedTrainingJob:
		args := job.Args().(*types.SubmitDeepSpeedJobArgs)
		name = args.ModelName
		if name == "" {
			return nil, nil, nil
		}
		for key, value := range args.Labels {
			versionTags = append(versionTags, &types.ModelVersionTag{
				Key:   key,
				Value: value,
			})
		}
		source = args.ModelSource
	}
	modelClient, err := client.Model()
	if err != nil {
		return nil, nil, err
	}
	return modelClient.CreateRegisteredModelAndModelVersion(name, description, tags, "auto", versionDescription, versionTags, source)
}

func getFullSubmitCommand(cmd *cobra.Command, args []string) string {
	var lines []string

	// Construct a line from current command and all parent commands
	var elems []string
	for curr := cmd; curr != nil; curr = curr.Parent() {
		elems = append([]string{curr.Name()}, elems...)
	}
	lines = append(lines, strings.Join(elems, " "))

	// Append every flag as a line
	cmd.Flags().Visit(func(f *pflag.Flag) {
		if !f.Changed {
			return
		}
		sliceValue, ok := f.Value.(pflag.SliceValue)
		if ok {
			for _, v := range sliceValue.GetSlice() {
				lines = append(lines, fmt.Sprintf("--%s %v", f.Name, v))
			}
		} else {
			lines = append(lines, fmt.Sprintf("--%s %v", f.Name, f.Value))
		}
	})

	// Append all args as a line
	lines = append(lines, fmt.Sprintf("\"%s\"", strings.Join(args, " ")))

	s := strings.Join(lines, " \\\n    ")
	log.Debugf("full submit command:\n%s", s)
	return s
}

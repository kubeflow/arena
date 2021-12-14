package argsbuilder

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"reflect"
	"strings"
)

type ModelBenchmarkArgsBuilder struct {
	args        *types.ModelBenchmarkArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewModelBenchmarkArgsBuilder(args *types.ModelBenchmarkArgs) ArgsBuilder {
	args.Type = types.ModelBenchmarkJob
	m := &ModelBenchmarkArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	m.AddSubBuilder(
		NewModelArgsBuilder(&m.args.CommonModelArgs),
	)
	m.AddArgValue("default-image", DefaultModelJobImage)
	return m
}

func (m *ModelBenchmarkArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*m)), ".")
	return items[len(items)-1]
}

func (m *ModelBenchmarkArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		m.subBuilders[b.GetName()] = b
	}
	return m
}

func (m *ModelBenchmarkArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range m.subBuilders {
		m.subBuilders[name].AddArgValue(key, value)
	}
	m.argValues[key] = value
	return m
}

func (m *ModelBenchmarkArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range m.subBuilders {
		m.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().IntVar(&m.args.Concurrency, "concurrency", 0, "number of benchmark concurrently")
	command.Flags().IntVar(&m.args.Requests, "requests", 0, "number of requests to run")
	command.Flags().IntVar(&m.args.Duration, "duration", 0, "benchmark duration")
	command.Flags().StringVar(&m.args.ReportPath, "report-path", "", "benchmark result saved path")
}

func (m *ModelBenchmarkArgsBuilder) PreBuild() error {
	for name := range m.subBuilders {
		if err := m.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModelBenchmarkArgsBuilder) Build() error {
	for name := range m.subBuilders {
		if err := m.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := m.preprocess(); err != nil {
		return err
	}
	return nil
}

func (m *ModelBenchmarkArgsBuilder) preprocess() (err error) {
	log.Debugf("command: %s", m.args.Command)
	if m.args.Image == "" {
		return fmt.Errorf("image must be specified")
	}
	if m.args.ModelConfigFile == "" {
		// need to validate modelName, modelPath if not specify modelConfigFile
		if m.args.ModelName == "" {
			return fmt.Errorf("model name must be specified")
		}
		if m.args.ModelPath == "" {
			return fmt.Errorf("model path must be specified")
		}
		if m.args.Inputs == "" {
			return fmt.Errorf("model inputs must be specified")
		}
		if m.args.Outputs == "" {
			return fmt.Errorf("model outputs must be specified")
		}
	} else {
		//populate content from modelConfigFile
		if m.args.ModelName != "" {
			log.Infof("modelConfigFile=%s is specified, so --model-name will be ingored", m.args.ModelConfigFile)
		}
		if m.args.ModelPath != "" {
			log.Infof("modelConfigFile=%s is specified, so --model-path will be ignored", m.args.ModelConfigFile)
		}
		if m.args.Inputs != "" {
			log.Infof("modelConfigFile=%s is specified, so --inputs will be ignored", m.args.ModelConfigFile)
		}
		if m.args.Inputs != "" {
			log.Infof("modelConfigFile=%s is specified, so --outputs will be ignored", m.args.ModelConfigFile)
		}
	}
	return nil
}

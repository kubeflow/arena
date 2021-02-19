package argsbuilder

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/spf13/cobra"
	"reflect"
	"strings"
)

type SeldonServingArgsBuilder struct {
	args        *types.SeldonServingArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSeldonServingArgsBuilder(args *types.SeldonServingArgs) ArgsBuilder {
	args.Type = types.SeldonServingJob
	s := &SeldonServingArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
	s.AddSubBuilder(
		NewServingArgsBuilder(&s.args.CommonServingArgs),
	)
	return s
}

func (s *SeldonServingArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SeldonServingArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SeldonServingArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SeldonServingArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().StringVar(&s.args.Implementation, "implementation", "TENSORFLOW_SERVER", "the type of serving implementation, default to TENSORFLOW_SERVER")
	command.Flags().StringVar(&s.args.ModelUri, "model-uri", "", "the uri direct to the model file")
}

func (s *SeldonServingArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SeldonServingArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.validate(); err != nil {
		return err
	}
	return nil
}

func (s *SeldonServingArgsBuilder) validate() (err error) {
	if s.args.ModelUri == "" && s.args.Image == "" {
		return fmt.Errorf("model uri and image can not be empty at the same time.")
	}
	return nil
}

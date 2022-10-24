package argsbuilder

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/kubeflow/arena/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type SubmitTensorboardArgsBuilder struct {
	args        *types.SubmitTensorboardArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitTensorboardArgsBuilder(args *types.SubmitTensorboardArgs) ArgsBuilder {
	return &SubmitTensorboardArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
}

func (s *SubmitTensorboardArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitTensorboardArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitTensorboardArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}
func (s *SubmitTensorboardArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SubmitTensorboardArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	if err := s.processTensorboard(); err != nil {
		return err
	}
	return nil
}

func (s *SubmitTensorboardArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().BoolVar(&s.args.UseTensorboard, "tensorboard", false, "enable tensorboard")
	command.Flags().StringVar(&s.args.TensorboardImage, "tensorboardImage", "registry.cn-zhangjiakou.aliyuncs.com/acs/tensorflow:1.12.0-devel", "the docker image for tensorboard")
	command.Flags().MarkDeprecated("tensorboardImage", "please use --tensorboard-image instead")
	command.Flags().StringVar(&s.args.TensorboardImage, "tensorboard-image", "registry.cn-zhangjiakou.aliyuncs.com/acs/tensorflow:1.12.0-devel", "the docker image for tensorboard")
	command.Flags().StringVar(&s.args.TrainingLogdir, "logdir", "/training_logs", "the training logs dir, default is /training_logs")
}

func (s *SubmitTensorboardArgsBuilder) processTensorboard() error {
	if !s.args.UseTensorboard {
		return nil
	}
	key := ShareDataPrefix + "dataset"
	if s.argValues[key] == nil {
		return fmt.Errorf("not found dataset which is passed by parent builder")
	}
	dataSet := s.argValues[key].(map[string]string)
	log.Debugf("dataMap %v", dataSet)
	if path.IsAbs(s.args.TrainingLogdir) && !s.isLoggingInPVC(dataSet) {
		// Need to consider pvc
		s.args.HostLogPath = fmt.Sprintf("/arena_logs/training%s", util.RandomInt32())
		s.args.IsLocalLogging = true
	} else {
		// doing nothing for hdfs path
		log.Debugf("Doing nothing for logging Path %s", s.args.TrainingLogdir)
		s.args.IsLocalLogging = false
	}
	return nil
}

// check if the path in the pvc
func (s *SubmitTensorboardArgsBuilder) isLoggingInPVC(dataMap map[string]string) (inPVC bool) {
	for pvc, path := range dataMap {
		if strings.HasPrefix(s.args.TrainingLogdir, path) {
			log.Debugf("Log path %s is contained by pvc %s's path %s", s.args.TrainingLogdir, pvc, path)
			inPVC = true
			break
		} else {
			log.Debugf("Log path %s is not contained by pvc %s's path %s", s.args.TrainingLogdir, pvc, path)
		}
	}
	return inPVC
}

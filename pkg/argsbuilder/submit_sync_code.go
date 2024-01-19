package argsbuilder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type SubmitSyncCodeArgsBuilder struct {
	args        *types.SubmitSyncCodeArgs
	argValues   map[string]interface{}
	subBuilders map[string]ArgsBuilder
}

func NewSubmitSyncCodeArgsBuilder(args *types.SubmitSyncCodeArgs) ArgsBuilder {
	return &SubmitSyncCodeArgsBuilder{
		args:        args,
		argValues:   map[string]interface{}{},
		subBuilders: map[string]ArgsBuilder{},
	}
}

func (s *SubmitSyncCodeArgsBuilder) GetName() string {
	items := strings.Split(fmt.Sprintf("%v", reflect.TypeOf(*s)), ".")
	return items[len(items)-1]
}

func (s *SubmitSyncCodeArgsBuilder) AddSubBuilder(builders ...ArgsBuilder) ArgsBuilder {
	for _, b := range builders {
		s.subBuilders[b.GetName()] = b
	}
	return s
}

func (s *SubmitSyncCodeArgsBuilder) AddArgValue(key string, value interface{}) ArgsBuilder {
	for name := range s.subBuilders {
		s.subBuilders[name].AddArgValue(key, value)
	}
	s.argValues[key] = value
	return s
}

func (s *SubmitSyncCodeArgsBuilder) PreBuild() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].PreBuild(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SubmitSyncCodeArgsBuilder) Build() error {
	for name := range s.subBuilders {
		if err := s.subBuilders[name].Build(); err != nil {
			return err
		}
	}
	return s.handleSyncCode()
}

func (s *SubmitSyncCodeArgsBuilder) handleSyncCode() error {
	switch s.args.SyncMode {
	case "":
		log.Debugf("No action for sync Code")
	case "git":
		log.Debugf("Check and prepare sync code with git")
		if s.args.SyncSource == "" {
			return fmt.Errorf("--syncSource should be set when syncMode is set")
		}
		// split test.git to test
		parts := strings.Split(strings.Trim(s.args.SyncSource, "/"), "/")
		s.args.SyncGitProjectName = strings.Split(parts[len(parts)-1], ".git")[0]
		log.Debugf("Try to split %s to get project name %s", s.args.SyncSource, s.args.SyncGitProjectName)
	case "rsync":
		log.Debugf("Check and prepare sync code with rsync")
		if s.args.SyncSource == "" {
			return fmt.Errorf("--syncSource should be set when syncMode is set")
		}

	default:
		log.Fatalf("Unknown sync mode: %s", s.args.SyncMode)
		return fmt.Errorf("Unknown sync mode: %s, it should be git or rsync", s.args.SyncMode)
	}
	return nil
}

func (s *SubmitSyncCodeArgsBuilder) AddCommandFlags(command *cobra.Command) {
	for name := range s.subBuilders {
		s.subBuilders[name].AddCommandFlags(command)
	}
	command.Flags().StringVar(&s.args.SyncMode, "syncMode", "", "syncMode: support rsync, hdfs, git")
	_ = command.Flags().MarkDeprecated("syncMode", "please use --sync-mode instead")
	command.Flags().StringVar(&s.args.SyncMode, "sync-mode", "", "syncMode: support rsync, hdfs, git")

	// command.MarkFlagRequired("syncMode")
	command.Flags().StringVar(&s.args.SyncSource, "syncSource", "", "syncSource: for rsync, it's like 10.88.29.56::backup/data/logoRecoTrain.zip; for git, it's like https://github.com/kubeflow/tf-operator.git")
	_ = command.Flags().MarkDeprecated("syncSource", "please use --sync-source instead")
	command.Flags().StringVar(&s.args.SyncSource, "sync-source", "", "sync-source: for rsync, it's like 10.88.29.56::backup/data/logoRecoTrain.zip; for git, it's like https://github.com/kubeflow/tf-operator.git")

	command.Flags().StringVar(&s.args.SyncImage, "syncImage", "", "the docker image of syncImage")
	_ = command.Flags().MarkDeprecated("syncImage", "please use --sync-image instead")
	command.Flags().StringVar(&s.args.SyncImage, "sync-image", "", "the docker image of syncImage")
}

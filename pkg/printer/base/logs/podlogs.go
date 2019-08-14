package logs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kubeflow/arena/pkg/podlogs"
)

var (
	ErrPodNotFound      = errors.New(`no logs return,because not found instance.`)
	ErrTooManyPodsFound = errors.New(`too many pods found in the job,`)
)

type PodLogPrinter struct {
	PodNames []string
	PodLog   *podlogs.PodLog
	Pipe
}

type Pipe struct {
	Reader io.ReadCloser
	Writer io.WriteCloser
}

func NewPodLogPrinter(podNames []string, logArgs *podlogs.OuterRequestArgs) (*PodLogPrinter, error) {
	podLog, err := podlogs.NewPodLog(logArgs)
	if err != nil {
		return nil, err
	}
	piper, pipew := io.Pipe()
	pipe := Pipe{Reader: piper, Writer: pipew}
	return &PodLogPrinter{
		PodNames: podNames,
		PodLog:   podLog,
		Pipe:     pipe,
	}, nil
}

func (slp *PodLogPrinter) CheckPodIsInJob() error {
	podName := slp.PodLog.Args.PodName
	names := slp.PodNames
	switch length := len(names); {
	case length == 0:
		return ErrPodNotFound
	case length == 1:
		if podName == "" {
			slp.PodLog.Args.PodName = names[0]
			return nil
		}
		return ErrPodNotFound
	default:
		if podName == "" {
			return ErrTooManyPodsFound
		}
		for _, name := range names {
			if name == podName {
				return nil
			}
		}
		return ErrPodNotFound
	}
}
func (slp *PodLogPrinter) Print() (int, error) {
	defer slp.Pipe.Reader.Close()
	if err := slp.CheckPodIsInJob(); err != nil {
		if err == ErrTooManyPodsFound {
			slp.PrintMultiPodsHelp()
			return 1, nil
		}
		return 1, err
	}
	if err := slp.PodLog.GetPodLogEntry(func(reader io.ReadCloser) {
		defer slp.Pipe.Writer.Close()
		defer reader.Close()
		io.Copy(slp.Pipe.Writer, reader)
	}); err != nil {
		return 1, err
	}
	io.Copy(os.Stdout, slp.Pipe.Reader)
	return 0, nil
}
func (slp *PodLogPrinter) PrintMultiPodsHelp() {
	header := "There is more than one instance in the job:"
	footer := "please use --instance to filter"
	lines := []string{header}
	for _, name := range slp.PodNames {
		lines = append(lines, "    "+name)
	}
	lines = append(lines, footer)
	fmt.Println(strings.Join(lines, "\n"))
}

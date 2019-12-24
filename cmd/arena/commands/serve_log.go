package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/kubeflow/arena/pkg/podlogs"
	printserve "github.com/kubeflow/arena/pkg/printer/serving/logs"
	"github.com/kubeflow/arena/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

var (
	version      string
	instance     string
	isFollow     bool
	sinceTime    string
	sinceSeconds time.Duration
	tailLines    int
)

func NewServingLogCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "logs ServingJobName",
		Short: "display logs of a serving job",
		Run: func(cmd *cobra.Command, args []string) {
			// no serving name is an error
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			// set loglevel
			util.SetLogLevel(logLevel)
			// initate kubenetes client
			setupKubeconfig()
			conf, err := clientConfig.ClientConfig()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			client, err := initKubeClient()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			servingName := args[0]
			kubeClient := kubernetes.NewForConfigOrDie(conf)
			outReqArgs := &podlogs.OuterRequestArgs{
				PodName:      instance,
				Namespace:    namespace,
				Follow:       isFollow,
				RetryCount:   5,
				SinceSeconds: sinceSeconds,
				SinceTime:    sinceTime,
				Tail:         tailLines,
				RetryTimeout: time.Millisecond,
				KubeClient:   kubeClient,
			}
			code := printserve.LogPrint(client, namespace, servingName, stype, version, outReqArgs)
			if code != 0 {
				os.Exit(code)

			}
		},
	}
	//command.Flags().BoolVar(&allNamespaces, "all-namespaces", false, "all namespace")
	command.Flags().StringVar(&version, "version", "", "assign the serving job version")
	command.Flags().StringVar(&stype, "type", "", `assign the serving job type,type can be "tf"("tensorflow"),"trt"("tensorrt"),"custom"`)
	command.Flags().StringVarP(&instance, "instance", "i", "", `assign the instance name of the job`)
	command.Flags().StringVar(&sinceTime, "since-time", "", `Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used`)
	command.Flags().DurationVar(&sinceSeconds, "since-seconds", 0, `Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.`)
	command.Flags().BoolVarP(&isFollow, "follow", "f", false, `Specify if the logs should be streamed.`)
	command.Flags().IntVarP(&tailLines, "tail", "t", 0, `assign how many log lines from the end we want to read`)
	return command

}

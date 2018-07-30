package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/kubeflow/arena/util"
	"github.com/spf13/cobra"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type logEntry struct {
	displayName string
	pod         string
	time        time.Time
	line        string
}

func NewLogsCommand() *cobra.Command {
	var (
		printer   logPrinter
		since     string
		sinceTime string
		tail      int64
	)
	var command = &cobra.Command{
		Use:   "logs training job",
		Short: "print the logs for a task of the training job",
		Run: func(cmd *cobra.Command, args []string) {
			util.SetLogLevel(logLevel)
			if len(args) == 0 {
				cmd.HelpFunc()(cmd, args)
				os.Exit(1)
			}
			name = args[0]
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

			printer.kubeClient = kubernetes.NewForConfigOrDie(conf)
			if tail > 0 {
				printer.tail = &tail
			}
			if sinceTime != "" {
				parsedTime, err := time.Parse(time.RFC3339, sinceTime)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				meta1Time := metav1.NewTime(parsedTime)
				printer.sinceTime = &meta1Time
			} else if since != "" {
				parsedSince, err := strconv.ParseInt(since, 10, 64)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				printer.sinceSeconds = &parsedSince
			}
			// podName, err := getPodNameFromJob(printer.kubeClient, namespace, name)
			job, err := getTrainingJob(client, name, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			var pod v1.Pod

			if printer.pod != "" {
				for _, p := range job.AllPods() {
					if p.Name == printer.pod {
						pod = p
						break
					}
				}
				if pod.Name == "" {
					fmt.Printf("The instance %s is not belong to the job %s, so no log can be found.", printer.pod, name)
					os.Exit(1)
				}
			} else {
				pod = job.ChiefPod()
			}

			err = printer.PrintPodLogs(pod.Name, namespace)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	command.Flags().BoolVarP(&printer.follow, "follow", "f", false, "Specify if the logs should be streamed.")
	command.Flags().StringVar(&since, "since", "", "Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVar(&sinceTime, "since-time", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().Int64Var(&tail, "tail", -1, "Lines of recent log file to display. Defaults to -1 with no selector, showing all log lines otherwise 10, if a selector is provided.")
	command.Flags().BoolVar(&printer.timestamps, "timestamps", false, "Include timestamps on each line in the log output")

	// command.Flags().StringVar(&printer.pod, "instance", "", "Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.")
	command.Flags().StringVarP(&printer.pod, "instance", "i", "", "Specify the task instance to get log")
	return command
}

type logPrinter struct {
	pod          string
	follow       bool
	sinceSeconds *int64
	sinceTime    *metav1.Time
	tail         *int64
	timestamps   bool
	kubeClient   kubernetes.Interface
}

// PrintPodLogs prints logs for a single pod
func (p *logPrinter) PrintPodLogs(podName, namespace string) error {

	var logs []logEntry

	err := p.getPodLogs("", podName, namespace, p.follow, p.tail, p.sinceSeconds, p.sinceTime, func(entry logEntry) {
		logs = append(logs, entry)
	})
	if err != nil {
		return err
	}
	for _, entry := range logs {
		p.printLogEntry(entry)
	}
	return nil
}

func (p *logPrinter) getPodLogs(displayName string, podName string, podNamespace string, follow bool, tail *int64, sinceSeconds *int64, sinceTime *metav1.Time, callback func(entry logEntry)) error {
	err := p.ensureContainerStarted(podName, podNamespace, 5, time.Millisecond)
	if err != nil {
		return err
	}
	readCloser, err := p.kubeClient.CoreV1().Pods(podNamespace).GetLogs(podName, &v1.PodLogOptions{
		// Container:    p.container,
		Follow:       follow,
		Timestamps:   true,
		SinceSeconds: sinceSeconds,
		SinceTime:    sinceTime,
		TailLines:    tail,
	}).Stream()
	// if err == nil {
	// 	scanner := bufio.NewScanner(stream)
	// 	for scanner.Scan() {
	// 		line := scanner.Text()
	// 		parts := strings.Split(line, " ")
	// 		logTime, err := time.Parse(time.RFC3339, parts[0])
	// 		if err == nil {
	// 			lines := strings.Join(parts[1:], " ")
	// 			for _, line := range strings.Split(lines, "\r") {
	// 				if line != "" {
	// 					callback(logEntry{
	// 						pod:         podName,
	// 						displayName: displayName,
	// 						time:        logTime,
	// 						line:        line,
	// 					})
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	if err != nil {
		return err
	}
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()
	_, err = io.Copy(writer, readCloser)

	defer readCloser.Close()
	return err
}

func (p *logPrinter) printLogEntry(entry logEntry) {
	line := entry.line
	if p.timestamps {
		line = entry.time.Format(time.RFC3339) + "	" + line
	}
	fmt.Println(line)
}

func (p *logPrinter) ensureContainerStarted(podName string, podNamespace string, retryCnt int, retryTimeout time.Duration) error {
	for retryCnt > 0 {
		pod, err := p.kubeClient.CoreV1().Pods(podNamespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if len(pod.Status.ContainerStatuses) == 0 {
			time.Sleep(retryTimeout)
			retryCnt--
			continue
		}

		var containerStatus *v1.ContainerStatus = &pod.Status.ContainerStatuses[0]
		// for _, status := range pod.Status.ContainerStatuses {
		// 	if status.Name == container {
		// 		containerStatus = &status
		// 		break
		// 	}
		// }
		if containerStatus == nil || containerStatus.State.Waiting != nil {
			time.Sleep(retryTimeout)
			retryCnt--
		} else {
			return nil
		}
	}
	return fmt.Errorf("pod '%s' has not been started.", podName)
}

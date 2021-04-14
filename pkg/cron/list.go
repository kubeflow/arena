package cron

import (
	"encoding/json"
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/config"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

var gvr = schema.GroupVersionResource{
	Group:    "apps.kubedl.io",
	Version:  "v1alpha1",
	Resource: "crons",
}

func ListCrons(namespace string, allNamespaces bool) ([]*types.CronInfo, error) {
	config := config.GetArenaConfiger().GetRestConfig()

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	list, err := dynamicClient.Resource(gvr).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var cronInfos []*types.CronInfo

	for _, item := range list.Items {
		b, err := item.MarshalJSON()
		if err != nil {
			continue
		}

		r := gjson.ParseBytes(b)

		creationTimestamp := r.Get("metadata").Get("creationTimestamp").String()
		createTime, err := formatTime(creationTimestamp)

		c := &types.CronInfo{
			Name: r.Get("metadata").Get("name").String(),
			Namespace: r.Get("metadata").Get("namespace").String(),
			Type: r.Get("spec").Get("template").Get("kind").String(),

			Schedule: r.Get("spec").Get("schedule").String(),
			ConcurrencyPolicy: r.Get("spec").Get("concurrencyPolicy").String(),
			HistoryLimit: r.Get("spec").Get("historyLimit").Int(),
			Deadline: r.Get("spec").Get("deadline").String(),
			Suspend: r.Get("spec").Get("suspend").Bool(),
			CreationTimestamp: createTime.Unix(),
		}

		cronInfos = append(cronInfos, c)
	}
	return cronInfos, nil
}

func DisplayAllCrons(crons []*types.CronInfo, allNamespaces bool, format types.FormatStyle) {
	switch format {
	case "json":
		for _, cron := range crons {
			data, _ := json.MarshalIndent(cron, "", "    ")
			fmt.Printf("%v", string(data))
		}
		return
	case "yaml":
		for _, cron := range crons {
			data, _ := yaml.Marshal(cron)
			fmt.Printf("%v", string(data))
		}
		return
	case "", "wide":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		var header []string
		if allNamespaces {
			header = append(header, "NAMESPACE")
		}
		header = append(header, []string{"NAME", "TYPE", "SCHEDULE", "SUSPEND", "DEADLINE", "CONCURRENCYPOLICY"}...)
		PrintLine(w, header...)

		for _, cron := range crons {
			var items []string
			if allNamespaces {
				items = append(items, cron.Namespace)
			}

			items = append(items, []string{
				cron.Name,
				cron.Type,
				cron.Schedule,
				strconv.FormatBool(cron.Suspend),
				cron.Deadline,
				cron.ConcurrencyPolicy,
			}...)
			PrintLine(w, items...)
		}
		_ = w.Flush()
		return
	}
}

func PrintLine(w io.Writer, fields ...string) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}

func formatTime(creationTimestamp string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05.000Z", creationTimestamp)
}
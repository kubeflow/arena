package cron

import (
	"fmt"
	"github.com/kubeflow/arena/pkg/apis/types"
	"github.com/tidwall/gjson"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
	"time"
)

var gvr = schema.GroupVersionResource{
	Group:    "apps.kubedl.io",
	Version:  "v1alpha1",
	Resource: "crons",
}

func printLine(w io.Writer, fields ...string) {
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}

func parseTime(strTime string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", strTime)
}

func buildCronInfo(bytes []byte) (*types.CronInfo, error) {
	r := gjson.ParseBytes(bytes)

	/*
		creationTimestamp := r.Get("metadata").Get("creationTimestamp").String()
		createTime, err := parseTime(creationTimestamp)
		if err != nil {
			return nil, err
		}
	*/

	c := &types.CronInfo{
		Name:              r.Get("metadata").Get("name").String(),
		Namespace:         r.Get("metadata").Get("namespace").String(),
		CreationTimestamp: r.Get("metadata").Get("creationTimestamp").String(),
		Type:              r.Get("spec").Get("template").Get("kind").String(),
		Schedule:          r.Get("spec").Get("schedule").String(),
		ConcurrencyPolicy: r.Get("spec").Get("concurrencyPolicy").String(),
		HistoryLimit:      r.Get("spec").Get("historyLimit").Int(),
		Deadline:          r.Get("spec").Get("deadline").String(),
		Suspend:           r.Get("spec").Get("suspend").Bool(),
		LastScheduleTime:  r.Get("status").Get("lastScheduleTime").String(),
	}

	var histories []types.CronHistoryInfo
	historyList := r.Get("status").Get("history").Array()
	for _, item := range historyList {
		history := types.CronHistoryInfo{
			Name:       item.Get("object").Get("name").String(),
			Group:      item.Get("object").Get("apiGroup").String(),
			Kind:       item.Get("object").Get("kind").String(),
			Status:     item.Get("status").String(),
			CreateTime: item.Get("created").String(),
			FinishTime: item.Get("finished").String(),
		}

		histories = append(histories, history)
	}

	c.History = histories

	return c, nil
}

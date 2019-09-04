package podlogs

import (
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type OuterRequestArgs struct {
	PodName      string
	Namespace    string
	Follow       bool
	Tail         int
	RetryCount   int
	RetryTimeout time.Duration
	SinceSeconds string
	SinceTime    string
	Timestamps   bool
	KubeClient   kubernetes.Interface
}

func ParseSinceTime(sinceTime string) (*metav1.Time, error) {
	if sinceTime == "" {
		return nil, nil
	}
	parsedTime, err := time.Parse(time.RFC3339, sinceTime)
	if err != nil {
		return nil, err
	}
	meta1Time := metav1.NewTime(parsedTime)
	return &meta1Time, nil

}

func ParseSinceSeconds(since string) (*int64, error) {
	invalidReturn := int64(0)
	parsedSince, err := strconv.ParseInt(since, 10, 64)
	if err != nil {
		return &invalidReturn, err
	}
	return &parsedSince, nil
}

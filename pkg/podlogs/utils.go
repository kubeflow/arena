package podlogs

import (
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
	SinceSeconds time.Duration
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

func RoundSeconds(since time.Duration) int64 {
	return int64(since.Round(time.Second).Seconds())
}

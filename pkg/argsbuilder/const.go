package argsbuilder

const (
	// ShareDataPrefix is used to defines sharing data from parent builder to children builder
	ShareDataPrefix = "share-"

	gangSchdName = "kube-batch"

	aliyunENIAnnotation = "k8s.aliyun.com/eni"

	jobSuspend = "scheduling.x-k8s.io/suspend"

	spotInstanceAnnotation = "job-supervisor.kube-ai.io/spot-instance"

	maxWaitTimeAnnotation = "job-supervisor.kube-ai.io/max-wait-time"
)

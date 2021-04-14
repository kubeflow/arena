package cron

type Cron interface {
	// GetName returns the job name
	Schedule() string
	// GetNamespace returns the namespace
	Deadline() string
}

type Processer interface {

	ListCrons(namespace string, allNamespace bool) ([]Cron, error)

}

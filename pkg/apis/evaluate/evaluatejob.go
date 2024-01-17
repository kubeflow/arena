package evaluate

type baseJob struct {
	name string
	args interface{}
}

type EvaluateJob struct {
	baseJob
}

func newBaseJob(name string, args interface{}) baseJob {
	return baseJob{
		name: name,
		args: args,
	}
}

func (b *baseJob) Name() string {
	return b.name
}

func (b *baseJob) Args() interface{} {
	return b.args
}

func NewEvaluateJob(name string, args interface{}) *EvaluateJob {
	return &EvaluateJob{
		baseJob: newBaseJob(name, args),
	}
}

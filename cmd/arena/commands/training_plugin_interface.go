package commands

type Runtime interface {
	// get the chart
	getChartName() string
}

var tfRuntimes = make(map[string]tfRuntime)

// Customized runtime for tf training training
type tfRuntime interface {
	// check the tfjob args
	check(tf *submitTFJobArgs) (err error)
	// transform the tfjob
	transform(tf *submitTFJobArgs) (err error)

	Runtime
}

func init() {

}

type defaultTFRuntime struct {
	name string
}

func (d *defaultTFRuntime) check(tf *submitTFJobArgs) (err error) {
	return
}

func (d *defaultTFRuntime) transform(tf *submitTFJobArgs) (err error) {
	return
}

func (d *defaultTFRuntime) getChartName() string {
	return d.name
}

// Get the TF runtime
func getTFRuntime(name string) (runtime tfRuntime) {
	found := false
	if runtime, found = tfRuntimes[name]; !found {
		runtime = &defaultTFRuntime{name: name}
	}

	return runtime
}

// Get the runtime name
func getRuntimeName() string {
	name := ""
	if len(annotations) > 0 {
		annotationsMap := transformSliceToMap(annotations, "=")
		name, _ = annotationsMap["runtime"]
	}

	return name
}

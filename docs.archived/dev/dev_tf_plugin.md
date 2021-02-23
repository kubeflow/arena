## The TFJob plugin framework

If you'd like to customize or enhance the TFJob with your own chart or code.


## Developer Workflow

### Step 1: Implement the following function (optional)

```
// Customized runtime for tf training training
type tfRuntime interface {
	// check the tfjob args
	check(tf *submitTFJobArgs) (err error)
	// transform the tfjob
	transform(tf *submitTFJobArgs) (err error)
	
	getChartName() string
}
```

You can refer the implmentation of default tf runtime [../../cmd/arena/commands/training_plugin_interface.go](training_plugin_interface.go)


### Step 2. Create your own chart

If you don't need to create your code for `check` or `transform`, you can create the chart in the same directory of tfjob, mpijob. For example, the chart name is `mock`.

```
cd /charts
cp -r tfjob mock
```

## User Workflow

Just run with the command by specifying annotation `runtime={your runtime}`

```
arena submit tf \
--name=test \
--annotation="runtime=mock" \
--workers=1 \
--chief \
--chief-cpu=4 \
--evaluator \
--evaluator-cpu=4 \
--worker-cpu=2 \
"python test.py"
```


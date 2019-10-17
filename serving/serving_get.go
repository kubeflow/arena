package serving
import(
	servejob "github.com/kubeflow/arena/pkg/jobs/serving"
)
func ServingGetExecute(client *kubernetes.Clientset, servingName, namespace, stype, version, format string) {
	// check some conditions are ok
	if err := PrepareCheck(stype); err != nil {
		log.Errorf(err.Error())
		os.Exit(1)
	}
	manager,err := servejob.NewServingJobManager(client , namespace)
	if err != nil {
		log.Errorf("%v",err)
		os.Exit(1)
	}
	args := servejob.ServingJobFilterArgs{
		Namespace: namespace,
		Type: stype,
		Version: version, 
		Name: servingName,
	}
	jobs := manager.GetTargetServingJob(args)
	if len(jobs) == 0 || len(jobs) > 1{
		manager.PrintHelp(jobs)
		os.Exit(1)
	}
	job := jobs[0]
	printString := ""
	switch format {
	case "json":
		printString,err = job.GetJsonPrintString()
		if err != nil {
			log.Errorf("print job %v failed,reason: %v",servingName,err)
			os.Exit(1)
		}
	case "yaml":
		printString,err = job.GetYamlPrintString()
		if err != nil {
			log.Errorf("print job %v failed,reason: %v",servingName,err)
			os.Exit(1)
		}
	default:
		printString,_ = job.GetWidePrintString()
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, printString)
	w.Flush()
}
// make some checks firstly
func PrepareCheck(stype string) error {
	err := servejob.CheckServingTypeIsOk(stype)
	if err != nil {
		return err
	}
	return nil
}
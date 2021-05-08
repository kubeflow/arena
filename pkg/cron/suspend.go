package cron

import "fmt"

func SuspendCron(name string, namespace string, suspend bool) error {
	err := GetCronHandler().UpdateCron(namespace, name, suspend)
	if err != nil {
		return err
	}

	var out string
	if suspend {
		out = fmt.Sprintf("cron %s suspend success", name)
	} else {
		out = fmt.Sprintf("cron %s resume success", name)
	}
	fmt.Println(out)
	return nil
}

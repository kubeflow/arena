package util

import (
	"os"
)

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

var chartFolder = ""

func GetChartsFolder() string {
	if chartFolder != "" {
		return chartFolder
	}
	homeChartsFolder := os.Getenv("HOME") + "/charts"
	if !pathExists(homeChartsFolder) {
		chartFolder = "/charts"

	} else {
		chartFolder = homeChartsFolder
	}
	return chartFolder
}

func StringInSlice(x string, list []string) bool {
	for _, y := range list {
		if y == x {
			return true
		}
	}
	return false
}

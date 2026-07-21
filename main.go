package main

import (
	"fmt"
	"helm.sh/helm/v3/pkg/strvals"
)

func main() {
	values := make(map[string]interface{})
	values["a"] = "scalar"
	err := strvals.ParseIntoFile("a.b=./main.go", values, func(rs []rune) (interface{}, error) {
		return string(rs), nil
	})
	fmt.Printf("values: %+v, err: %v\n", values, err)
}

package argsbuilder

import (
	"errors"
	"github.com/kubeflow/arena/pkg/apis/types"
	"strings"
)

func transformSliceToMap(sets []string, split string) (valuesMap map[string]string) {
	valuesMap = map[string]string{}
	for _, member := range sets {
		splits := strings.SplitN(member, split, 2)
		if len(splits) == 2 {
			valuesMap[splits[0]] = splits[1]
		}
	}

	return valuesMap
}

func parseTolerationString(toleration string) (*types.TolerationArgs, error) {
	if strings.Contains(toleration, "=") && strings.Contains(toleration, ":") && strings.Contains(toleration, ",") {
		key, toleration := split(toleration, "=")
		value, toleration := split(toleration, ":")
		effect, operator := split(toleration, ",")
		return &types.TolerationArgs{
			Key:      key,
			Value:    value,
			Effect:   effect,
			Operator: operator,
		}, nil
	} else {
		return nil, errors.New("invalid toleration format")
	}
}

func split(value, sep string) (string, string) {
	index := strings.Index(value, sep)
	return value[:index], value[index+1:]
}

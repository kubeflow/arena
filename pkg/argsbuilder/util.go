package argsbuilder
import(
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
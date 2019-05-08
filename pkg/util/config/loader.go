package env

import (
	"bufio"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const doubleQuoteSpecialChars = "\\\n\r\"!$`"

// ReadEnvFile returns configs map
func ReadConfigFile(filename string) (configs map[string]string) {
	configs = make(map[string]string)
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		log.Debugf("load config file: %s due to error %v", filename, err)
		return
	}

	for _, line := range lines {
		if !canIgnore(line) {
			var key, value string
			splitString := strings.SplitN(line, "=", 2)
			if len(splitString) != 2 {
				continue
			}
			key = splitString[0]
			value = splitString[1]

			configs[key] = value
		}
	}

	return
}

func canIgnore(line string) bool {
	trimmedLine := strings.Trim(line, " \n\t")
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}

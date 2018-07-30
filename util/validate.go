package util

import (
	"fmt"
	"regexp"
)

const dns1123SubdomainFmt string = dns1123LabelFmt + "(\\." + dns1123LabelFmt + ")*"
const dns1123SubdomainErrorMsg string = "a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character"
const DNS1123SubdomainMaxLength int = 253

const dns1123LabelFmt string = "[a-z0-9]([-a-z0-9]*[a-z0-9])?"
const dns1123LabelErrMsg string = "a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character"
const DNS1123LabelMaxLength int = 63

var dns1123LabelRegexp = regexp.MustCompile("^" + dns1123LabelFmt + "$")

var dns1123SubdomainRegexp = regexp.MustCompile("^" + dns1123SubdomainFmt + "$")

func ValidateJobName(value string) error {
	if len(value) > DNS1123LabelMaxLength {
		return fmt.Errorf("The len of name %d is too long, it should be less than %d",
			len(value),
			DNS1123SubdomainMaxLength)
	}
	if !dns1123LabelRegexp.MatchString(value) {
		return fmt.Errorf("The job name must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character.")
	}
	return nil
}

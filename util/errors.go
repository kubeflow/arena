package util

import "strings"

func IsNeedWaitError(err error) bool {
	// estr := err.Error()
	// return strings.Contains(estr, "Need waited.")
	return checkError(err, "Need waited.")
}

func checkError(err error, content string) bool {
	estr := err.Error()
	return strings.Contains(estr, content)
}

func IsConnectionRefusedError(err error) bool {
	return checkError(err, "connection refused")
}

func IsUnexpectedEOFError(err error) bool {
	return checkError(err, "unexpected EOF")
}

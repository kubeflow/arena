package util

func AddNamespaceToArgs(args []string, namespace string) []string {
	if namespace == "" {
		return args
	}

	return append(args, "--namespace", namespace)
}

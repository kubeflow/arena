package main

import (
	"log"

	"github.com/kubeflow/arena/cmd/arena/commands"
	"github.com/spf13/cobra/doc"
)

func main() {
	arenaCLI := commands.NewCommand()
	err := doc.GenMarkdownTree(arenaCLI, "./docs/cli")
	if err != nil {
		log.Fatal(err)
	}
}

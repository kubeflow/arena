package serving

import (
	"fmt"
	"io"
	"strings"

	"github.com/kubeflow/arena/pkg/apis/types"
)

func PrintLine(w io.Writer, fields ...string) {
	//w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	buffer := strings.Join(fields, "\t")
	fmt.Fprintln(w, buffer)
}

// print the help info  when jobs more than one
func moreThanOneJobHelpInfo(infos []ServingJob) string {
	header := fmt.Sprintf("There is %d jobs have been found:", len(infos))
	tableHeader := "NAME\tTYPE\tVERSION"
	lines := []string{tableHeader}
	footer := fmt.Sprintf("please use '--type' or '--version' to filter.")
	for _, info := range infos {
		line := fmt.Sprintf("%s\t%s\t%s",
			info.Name(),
			info.Type(),
			info.Version(),
		)
		lines = append(lines, line)
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(lines, "\n"), footer)
}

func moreThanOneInstanceHelpInfo(instances []types.ServingInstance) string {
	header := fmt.Sprintf("There is %d instances have been found:", len(instances))
	lines := []string{}
	footer := fmt.Sprintf("please use '-i' or '--instance' to filter.")
	for _, i := range instances {
		lines = append(lines, fmt.Sprintf("%v", i.Name))
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n", header, strings.Join(lines, "\n"), footer)

}

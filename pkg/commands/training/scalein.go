// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package training

import "github.com/spf13/cobra"

var (
	scaleinLong = `scalein a job.

Available Commands:
  etjob,et           scalein a ETJob.
    `
)

func NewScaleInCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:   "scalein",
		Short: "scalein a job.",
		Long:  scaleinLong,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	command.AddCommand(NewScaleInETJobCommand())

	return command
}

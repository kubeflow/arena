/*
Copyright 2024 The Kubeflow authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_test

import (
	"bytes"
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kubeflow/arena/pkg/apis/types"
)

var _ = Describe("TFJob", func() {
	Context("Basic lifecycle management", func() {
		It("Should be able to manage tfjob", func() {
			jobName := "tfjob-test"
			jobNamespace := "default"
			jobType := string(types.TFTrainingJob)

			var err error
			var output string

			By("Use arena to submit a tfjob")
			var out bytes.Buffer
			submitCmd := exec.Command(
				"arena",
				"submit",
				jobType,
				fmt.Sprintf("--name=%s", jobName),
				fmt.Sprintf("--namespace=%s", jobNamespace),
				"--image=tensorflow:test",
				"python main.py",
			)
			submitCmd.Stdout = &out
			submitCmd.Stderr = &out
			err = submitCmd.Run()
			output = out.String()
			fmt.Print(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).Should(ContainSubstring(fmt.Sprintf("tfjob.kubeflow.org/%s created", jobName)))
			out.Reset()

			By("Use arena to get the status of a tfjob")
			getCmd := exec.Command(
				"arena",
				"get",
				fmt.Sprintf("--namespace=%s", jobNamespace),
				jobName,
			)
			getCmd.Stdout = &out
			getCmd.Stderr = &out
			err = getCmd.Run()
			output = out.String()
			fmt.Print(output)
			Expect(err).NotTo(HaveOccurred())

			By("Use arena to list all tfjobs")
			listCmd := exec.Command(
				"arena",
				"list",
				fmt.Sprintf("--namespace=%s", jobNamespace),
			)
			listCmd.Stdout = &out
			listCmd.Stderr = &out
			err = listCmd.Run()
			output = out.String()
			fmt.Print(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).Should(ContainSubstring(jobName))
			out.Reset()

			By("Use arena to delete a tfjob")
			deleteCmd := exec.Command(
				"arena",
				"delete",
				fmt.Sprintf("--namespace=%s", jobNamespace),
				jobName,
			)
			deleteCmd.Stdout = &out
			deleteCmd.Stderr = &out
			err = deleteCmd.Run()
			output = out.String()
			fmt.Print(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).Should(ContainSubstring(fmt.Sprintf("The training job %s has been deleted successfully", jobName)))
			out.Reset()
		})
	})
})

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

var _ = Describe("Cron", func() {
	Context("Basic", func() {
		It("Should be able to manage the lifecycle of a cron tfjob", func() {
			jobName := "cron-test"
			jobNamespace := "default"
			schedule := "* 0 * * *"
			jobType := string(types.CronTFTrainingJob)

			var err error
			var output string

			By("Use arena to submit a cron tfjob")
			var out bytes.Buffer
			cronCmd := exec.Command(
				"arena",
				"cron",
				jobType,
				fmt.Sprintf("--schedule=%s", schedule),
				fmt.Sprintf("--name=%s", jobName),
				fmt.Sprintf("--namespace=%s", jobNamespace),
				"--image=tensorflow:latest",
				"python main.py",
			)
			cronCmd.Stdout = &out
			cronCmd.Stderr = &out
			err = cronCmd.Run()
			output = out.String()
			fmt.Print(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).Should(ContainSubstring(fmt.Sprintf("cron.apps.kubedl.io/%s created", jobName)))
			out.Reset()

			By("Use arena to get the status of a cron tfjob")
			getCmd := exec.Command(
				"arena",
				"cron",
				"get",
				fmt.Sprintf("--namespace=%s", jobNamespace),
				jobName,
			)
			getCmd.Stdout = &out
			getCmd.Stderr = &out
			err = getCmd.Run()
			output = out.String()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).Should(ContainSubstring(jobName))
			Expect(output).Should(ContainSubstring(jobNamespace))

			By("Use arena to list cron tfjobs")
			listCmd := exec.Command(
				"arena",
				"cron",
				"list",
				fmt.Sprintf("--namespace=%s", jobNamespace),
			)
			listCmd.Stdout = &out
			listCmd.Stderr = &out
			err = listCmd.Run()
			output = out.String()
			fmt.Print(output)
			Expect(err).NotTo(HaveOccurred())

			By("Use arena to suspend a cron tfjob")
			suspendCmd := exec.Command(
				"arena",
				"cron",
				"suspend",
				fmt.Sprintf("--namespace=%s", jobNamespace),
				jobName,
			)
			suspendCmd.Stdout = &out
			suspendCmd.Stderr = &out
			err = suspendCmd.Run()
			output = out.String()
			fmt.Print(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring(fmt.Sprintf("cron %s suspend success", jobName)))

			By("Use arena to resume a cron tfjob")
			resumeCmd := exec.Command(
				"arena",
				"cron",
				"resume",
				fmt.Sprintf("--namespace=%s", jobNamespace),
				jobName,
			)
			resumeCmd.Stdout = &out
			resumeCmd.Stderr = &out
			err = resumeCmd.Run()
			output = out.String()
			fmt.Print(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring(fmt.Sprintf("cron %s resume success", jobName)))

			By("Use arena to delete a cron tfjob")
			deleteCmd := exec.Command(
				"arena",
				"cron",
				"delete",
				fmt.Sprintf("--namespace=%s", jobNamespace),
				jobName,
			)
			deleteCmd.Stdout = &out
			deleteCmd.Stderr = &out
			err = deleteCmd.Run()
			output = out.String()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).Should(ContainSubstring(fmt.Sprintf("cron %s has deleted", jobName)))
		})
	})
})

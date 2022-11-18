package test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
)

var _ = Describe("kots apps", func() {
	var (
		err error
	)

	Context("replicated app --help", func() {
		It("should print usage", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"app", "--help"})

			err = cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).ToNot(HaveOccurred())

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring("list apps and create new apps"))
		})
	})

})

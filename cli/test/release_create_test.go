package test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/replicatedhq/replicated/cli/cmd"
)

var _ = Describe("release create", func() {
	Context("error case using --yaml flag with yaml filename", func() {
		It("should return an error telling user to use --yaml-file flag", func() {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			var expectedError = "use the --yaml-file flag when passing a yaml filename"

			rootCmd := cmd.GetRootCmd()
			rootCmd.SetArgs([]string{"release", "create", "--yaml", "installer.yaml", "--app", app.Slug})

			err := cmd.Execute(rootCmd, nil, &stdout, &stderr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedError))

			Expect(stderr.String()).To(BeEmpty())
			Expect(stdout.String()).ToNot(BeEmpty())

			Expect(stdout.String()).To(ContainSubstring(expectedError))
		})
	})
})

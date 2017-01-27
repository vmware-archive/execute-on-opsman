package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("The CLI", func() {

	Context("Params", func() {
		It("validates the presence of target", func() {
			cmd := exec.Command(binaryPath)
			session, err := gexec.Start(cmd, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			session.Wait()

			Expect(session.ExitCode()).To(Equal(1))
			Expect(session.Out).To(gbytes.Say(`target flag is required`))
		})

		It("validates the presence of username", func() {
			cmd := exec.Command(binaryPath, "--target", "example.com")
			session, err := gexec.Start(cmd, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			session.Wait()

			Expect(session.ExitCode()).To(Equal(1))
			Expect(session.Out).To(gbytes.Say(`username flag is required`))
		})

		It("validates the presence of password", func() {
			cmd := exec.Command(binaryPath, "--target", "example.com", "--username", "admin")
			session, err := gexec.Start(cmd, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			session.Wait()

			Expect(session.ExitCode()).To(Equal(1))
			Expect(session.Out).To(gbytes.Say(`password flag is required`))
		})
	})

})

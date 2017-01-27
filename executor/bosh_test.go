package executor_test

import (
	"fmt"

	. "github.com/pivotal-cf-experimental/execute-on-opsman/executor"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("The Bosh Command", func() {
	var (
		bosh  *BoshExecutor
		creds BoshCredentials
		cmd   string
	)

	BeforeEach(func() {
		creds = BoshCredentials{
			ClientId:     "ops_manager",
			ClientSecret: "client_secret",
		}

		bosh = NewBoshCommand(creds)
		cmd = bosh.Command("pid")
	})

	It("includes the bosh client id", func() {
		Ω(cmd).Should(ContainSubstring(fmt.Sprintf("BOSH_CLIENT=%s", creds.ClientId)))
	})

	It("includes the bosh client secret", func() {
		Ω(cmd).Should(ContainSubstring(fmt.Sprintf("BOSH_CLIENT_SECRET=%s", creds.ClientSecret)))
	})

	It("includes the bundle gemfile", func() {
		Ω(cmd).Should(ContainSubstring("BUNDLE_GEMFILE=/home/tempest-web/tempest/web/vendor/bosh/Gemfile"))
	})

	It("includes the bosh command", func() {
		Ω(cmd).Should(ContainSubstring("bundle exec bosh -n --ca-cert /var/tempest/workspaces/default/root_ca_certificate"))
	})
})

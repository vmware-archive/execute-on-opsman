package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

func TestExecuteOnOpsman(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ExecuteOnOpsman Suite")
}

var binaryPath string

var _ = BeforeSuite(func() {
	var err error
	binaryPath, err = gexec.Build("github.com/pivotal-cf-experimental/execute-on-opsman")
	Expect(err).ToNot(HaveOccurred())
	Expect(binaryPath).ToNot(BeEmpty())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

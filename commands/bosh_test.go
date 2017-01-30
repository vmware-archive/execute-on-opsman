package commands_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pivotal-cf-experimental/execute-on-opsman/commands"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bosh", func() {
	Describe("Execute", func() {
		var (
			command        commands.Bosh
			requestService *fakes.RequestService
			stdout         *fakes.Logger
			stderr         *fakes.Logger
		)

		BeforeEach(func() {
			requestService = &fakes.RequestService{}
			stdout = &fakes.Logger{}
			stderr = &fakes.Logger{}
			requestService.InvokeStub = func(input api.RequestServiceInvokeInput) (api.RequestServiceInvokeOutput, error) {
				if input.Path == "/api/v0/deployed/products/" {
					return api.RequestServiceInvokeOutput{
						StatusCode: http.StatusOK,
						Headers: http.Header{
							"Content-Type": []string{"application/json"},
							"Accept":       []string{"text/plain"},
						},
						Body: strings.NewReader(`[
							{
								"installation_name": "p-bosh-62a54920334b1f91fcb3",
								"guid": "p-bosh-62a54920334b1f91fcb3",
								"type": "p-bosh"
							},
							{
								"installation_name": "cf-88b75fe421f5630ad6b4",
								"guid": "cf-88b75fe421f5630ad6b4",
								"type": "cf"
							}
						]`),
					}, nil
				} else if input.Path == "/api/v0/deployed/director/manifest" {
					return api.RequestServiceInvokeOutput{
						StatusCode: http.StatusOK,
						Headers: http.Header{
							"Content-Type": []string{"application/json"},
							"Accept":       []string{"text/plain"},
						},
						Body: strings.NewReader(`{
							"jobs": [{
								"properties": {
									"uaa": {
										"clients": {
											"ops_manager": {
												"secret": "opsman_secret"
											}
										}
									},
									"director": {
										"address": "10.0.4.2"
									}
								}
							}]
						}`),
					}, nil
				}
				return api.RequestServiceInvokeOutput{}, fmt.Errorf("not supported")
			}
			command = commands.NewBoshCommand(requestService, "pcf.jitterbug.gcp.london.cf-app.com.com", stdout, stderr)
		})

		It("executes the bosh command", func() {
			err := command.Execute([]string{
				"--ssh-key", "/Users/pivotal/workspace/london-meta/gcp-environments/jitterbug/jitterbug-pcf.pem",
				"--product-name", "cf",
				"--command", "status",
			})
			Ω(err).ToNot(HaveOccurred())
		})
	})
})

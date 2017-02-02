/**
 * Copyright 2017 Pivotal Software, Inc.

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 * http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package commands_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pivotal-cf/execute-on-opsman/commands"
	"github.com/pivotal-cf/execute-on-opsman/commands/fakes"
	"github.com/pivotal-cf/om/api"
	omfakes "github.com/pivotal-cf/om/commands/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bosh", func() {
	Describe("Execute", func() {
		var (
			command        commands.Bosh
			requestService *omfakes.RequestService
			sshClient      *fakes.SSHClient
			stdout         *omfakes.Logger
			stderr         *omfakes.Logger
		)

		BeforeEach(func() {
			requestService = &omfakes.RequestService{}
			stdout = &omfakes.Logger{}
			stderr = &omfakes.Logger{}
			sshClient = &fakes.SSHClient{}
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
								"installation_name": "p-bosh-guid",
								"guid": "p-bosh-guid",
								"type": "p-bosh"
							},
							{
								"installation_name": "cf-guid",
								"guid": "cf-guid",
								"type": "cf"
							}
						]`),
					}, nil
				} else if input.Path == "/api/v0/deployed/director/manifest/" {
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
			command = commands.NewBoshCommand(requestService, sshClient, "pcf.example.com", stdout, stderr)
		})

		It("executes the bosh command", func() {
			err := command.Execute([]string{
				"--ssh-key", "/path/to/key.pem",
				"--command", "stop",
				"--product-name", "cf",
			})
			Expect(err).ToNot(HaveOccurred())

			input := requestService.InvokeArgsForCall(0)
			Expect(input.Path).To(Equal("/api/v0/deployed/director/manifest/"))
			Expect(input.Method).To(Equal("GET"))

			input = requestService.InvokeArgsForCall(1)
			Expect(input.Path).To(Equal("/api/v0/deployed/products/"))
			Expect(input.Method).To(Equal("GET"))

			Expect(sshClient.ExecuteOnRemoteCallCount()).To(Equal(1))
			sshInput := sshClient.ExecuteOnRemoteArgsForCall(0)
			Expect(sshInput.SSHKeyPath).To(Equal("/path/to/key.pem"))
			Expect(sshInput.Host).To(Equal("pcf.example.com"))
			Expect(sshInput.Env).To(ContainElement(`BOSH_CLIENT="ops_manager"`))
			Expect(sshInput.Env).To(ContainElement(`BOSH_CLIENT_SECRET="opsman_secret"`))
			Expect(sshInput.Env).To(ContainElement(`BUNDLE_GEMFILE=/home/tempest-web/tempest/web/vendor/bosh/Gemfile`))

			Expect(sshInput.Command).To(ContainElement(`bundle exec bosh`))
			Expect(sshInput.Command).To(ContainElement(`-n`))
			Expect(sshInput.Command).To(ContainElement(`--ca-cert /var/tempest/workspaces/default/root_ca_certificate`))
			Expect(sshInput.Command).To(ContainElement(`-t 10.0.4.2`))
			Expect(sshInput.Command).To(ContainElement(`-d /var/tempest/workspaces/default/deployments/cf-guid.yml`))
			Expect(sshInput.Command).To(ContainElement(`stop`))
		})

		Context("when no product name is specified", func() {
			It("doesn't include deployment manifest", func() {
				err := command.Execute([]string{
					"--ssh-key", "/path/to/key.pem",
					"--command", "stop",
				})
				Ω(err).ToNot(HaveOccurred())

				Expect(requestService.InvokeCallCount()).To(Equal(1))
				input := requestService.InvokeArgsForCall(0)
				Expect(input.Path).To(Equal("/api/v0/deployed/director/manifest/"))
				Expect(input.Method).To(Equal("GET"))
			})
		})

		Context("Validation", func() {
			It("fails when no ssh is provided", func() {
				err := command.Execute([]string{
					"--command", "stop",
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("ssh key path cannot be empty"))
			})
		})
	})
})

package executor

import "fmt"

type BoshExecutor struct {
	Credentials BoshCredentials
}

type BoshCredentials struct {
	ClientId     string
	ClientSecret string
}

func NewBoshCommand(creds BoshCredentials) *BoshExecutor {
	return &BoshExecutor{Credentials: creds}
}

func (e *BoshExecutor) RunOnce() {}
func (e *BoshExecutor) Command(product_id string) string {
	return fmt.Sprintf(
		"%s %s bundle exec bosh -n --ca-cert /var/tempest/workspaces/default/root_ca_certificate ",
		e.credentials(),
		e.bundleGemfilePath(),
	)
}

func (e *BoshExecutor) credentials() string {
	return fmt.Sprintf(
		"BOSH_CLIENT=%s BOSH_CLIENT_SECRET=%s",
		e.Credentials.ClientId,
		e.Credentials.ClientSecret,
	)
}

func (e *BoshExecutor) bundleGemfilePath() string {
	return "BUNDLE_GEMFILE=/home/tempest-web/tempest/web/vendor/bosh/Gemfile"
}

/**

    #   "BOSH_CLIENT=ops_manager BOSH_CLIENT_SECRET=#{uaa_secret} \
    #   BUNDLE_GEMFILE=/home/tempest-web/tempest/web/vendor/bosh/Gemfile \
    #   bundle exec bosh -n \
    #   --ca-cert /var/tempest/workspaces/default/root_ca_certificate \
    #   -d /var/tempest/workspaces/default/deployments/#{product_id}.yml \
    #   -t #{director_ip}"
**/

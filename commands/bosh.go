package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/flags"
)

type requestService interface {
	Invoke(api.RequestServiceInvokeInput) (api.RequestServiceInvokeOutput, error)
}

type logger interface {
	Printf(format string, v ...interface{})
}

type Bosh struct {
	requestService requestService
	ssh            SSHClient
	stdout         logger
	stderr         logger
	host           string
	Options        struct {
		SSHKeyPath  string `short:"i" long:"ssh-key-path" description:"path to ssh key"`
		ProductName string `short:"p" long:"product-name" description:"Product name"`
		Command     string `short:"c" long:"command"      description:"bosh command to execute"`
	}
}

type Products struct {
	Name string `json:"installation_name"`
	Guid string `json:"guid"`
	Type string `json:"type"`
}

type DirectorManifest struct {
	Jobs []Job `json:"jobs"`
}

type Job struct {
	Properties struct {
		Uaa struct {
			Clients struct {
				OpsManager struct {
					Secret string `json:"secret"`
				} `json:"ops_manager"`
			} `json:"clients"`
		} `json:"uaa"`
		Director struct {
			Address string `json:"address"`
		} `json:"director"`
	} `json:"properties"`
}

func NewBoshCommand(rs requestService, ssh SSHClient, host string, stdout, stderr logger) Bosh {
	return Bosh{requestService: rs, ssh: ssh, host: host, stdout: stdout, stderr: stderr}
}

func (b Bosh) Usage() commands.Usage {
	return commands.Usage{
		Description:      "Runs a bosh command from the OpsManager VM",
		ShortDescription: "Runs a bosh command from the OpsManager VM",
		Flags:            b.Options,
	}
}

func (b Bosh) Execute(args []string) error {
	_, err := flags.Parse(&b.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse curl flags: %s", err)
	}

	if b.Options.SSHKeyPath == "" {
		return fmt.Errorf("ssh key path cannot be empty")
	}

	manifest, err := b.getDirectorManifest()
	if err != nil {
		return err
	}

	boshCmd := []string{
		"bundle exec bosh", "-n",
		"--ca-cert /var/tempest/workspaces/default/root_ca_certificate",
		fmt.Sprintf("-t %s", manifest.Jobs[0].Properties.Director.Address),
	}

	var productId string
	if b.Options.ProductName != "" {
		productId, err = b.getProductId()
		if err != nil {
			return err
		}
		boshCmd = append(boshCmd, fmt.Sprintf("-d /var/tempest/workspaces/default/deployments/%s.yml", productId))
	}

	boshEnv := []string{
		`BOSH_CLIENT="ops_manager"`,
		fmt.Sprintf(`BOSH_CLIENT_SECRET="%s"`, manifest.Jobs[0].Properties.Uaa.Clients.OpsManager.Secret),
		"BUNDLE_GEMFILE=/home/tempest-web/tempest/web/vendor/bosh/Gemfile",
	}

	boshCmd = append(boshCmd, b.Options.Command)

	return b.ssh.ExecuteOnRemote(ExecuteOnRemoteInput{
		Host:       b.host,
		SSHKeyPath: b.Options.SSHKeyPath,
		Env:        boshEnv,
		Command:    boshCmd,
	})
}

func (b Bosh) getProductId() (string, error) {
	input := api.RequestServiceInvokeInput{
		Path:   "/api/v0/deployed/products/",
		Method: "GET",
	}

	output, err := b.requestService.Invoke(input)
	if err != nil {
		return "", fmt.Errorf("failed to get deployed product: %s", err)
	}

	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read api response body: %s", err)
	}

	var products []Products
	if err = json.Unmarshal([]byte(body), &products); err != nil {
		return "", fmt.Errorf("Could not unmarshal deployed products: %s", err)
	}

	for _, p := range products {
		if p.Type == b.Options.ProductName {
			return p.Guid, nil
		}
	}

	return "", fmt.Errorf("Could not find product: %s", b.Options.ProductName)
}

func (b Bosh) getDirectorManifest() (DirectorManifest, error) {
	var manifest DirectorManifest
	input := api.RequestServiceInvokeInput{
		Path:   "/api/v0/deployed/director/manifest/",
		Method: "GET",
	}

	output, err := b.requestService.Invoke(input)
	if err != nil {
		return manifest, fmt.Errorf("failed to get director manifest: %s", err)
	}

	body, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return manifest, fmt.Errorf("failed to read api response body: %s", err)
	}

	if err = json.Unmarshal([]byte(body), &manifest); err != nil {
		return manifest, fmt.Errorf("Could not unmarshal director manifest: %s", err)
	}

	return manifest, nil
}

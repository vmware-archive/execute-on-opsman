package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"golang.org/x/crypto/ssh"

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
	stdout         logger
	stderr         logger
	host           string
	Options        struct {
		SSHKeyPath  string `short:"i" long:"ssh-key"      description:"path to ssh key"`
		ProductName string `short:"p" long:"product-name" description:"Product name"`
		Command     string `short:"c" long:"command"      description:"bosh command to execute"`
	}
}

func NewBoshCommand(rs requestService, host string, stdout, stderr logger) Bosh {
	return Bosh{requestService: rs, host: host, stdout: stdout, stderr: stderr}
}

func (b Bosh) Usage() commands.Usage {
	return commands.Usage{
		Description:      "TODO",
		ShortDescription: "TODO",
		Flags:            b.Options,
	}
}

func (b Bosh) Execute(args []string) error {
	_, err := flags.Parse(&b.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse curl flags: %s", err)
	}

	var productId string
	if b.Options.ProductName != "" {
		productId, err = b.getProductId()
		if err != nil {
			return err
		}
	}

	manifest, err := b.getDirectorManifest()
	if err != nil {
		return err
	}

	boshEnv := []string{
		"BOSH_CLIENT=ops_manager",
		fmt.Sprintf(`BOSH_CLIENT_SECRET="%s"`, manifest.Jobs[0].Properties.Uaa.Clients.OpsManager.Secret),
	}

	boshCmd := []string{
		"BUNDLE_GEMFILE=/home/tempest-web/tempest/web/vendor/bosh/Gemfile bundle exec bosh", "-n",
		"--ca-cert /var/tempest/workspaces/default/root_ca_certificate",
		"-d", fmt.Sprintf("/var/tempest/workspaces/default/deployments/%s.yml", productId),
		"-t", manifest.Jobs[0].Properties.Director.Address,
		b.Options.Command,
	}

	return b.executeOnRemote(boshEnv, boshCmd)
}

func (b Bosh) executeOnRemote(env, cmd []string) error {

	pemBytes, err := ioutil.ReadFile(b.Options.SSHKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		log.Fatal(err)
	}

	auths := []ssh.AuthMethod{ssh.PublicKeys(signer)}

	cfg := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: auths,
	}
	cfg.SetDefaults()

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", b.host), cfg)
	for connErr := err; connErr != nil; {
		if strings.Contains(connErr.Error(), "unexpected message type 3") {
			log.Println("retrying")
			client, connErr = ssh.Dial("tcp", fmt.Sprintf("%s:22", b.host), cfg)
		} else {
			log.Fatal(connErr)
		}
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer

	fullcmd := strings.Join(append(env, strings.Join(cmd, " ")), " ")
	fmt.Println(fullcmd)

	log.Println("we have a session!")
	session.Stdout = &stdoutBuf
	err = session.Run(fullcmd)
	if err != nil {
		log.Fatalf("Run failed:%v", err)
	}
	log.Printf(">%s", stdoutBuf)

	return err
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
		Path:   "/api/v0/deployed/director/manifest",
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

// func (e *BoshExecutor) Environment() []string {
// 	env := []string{
// 		"BUNDLE_GEMFILE=/home/tempest-web/tempest/web/vendor/bosh/Gemfile",
// 	}
// 	env = append(env, e.credentials()...)
// 	return env
// }

// func (e *BoshExecutor) CommandArguments() []string {
// 	return []string{
// 		"-n",
// 		"--ca-cert /var/tempest/workspaces/default/root_ca_certificate",
// 		fmt.Sprintf("-d /var/tempest/workspaces/default/deployments/%s.yml", e.ProductId),
// 		fmt.Sprintf("-t %s", e.DirectorIp),
// 	}
// }

// func (e *BoshExecutor) credentials() []string {
// 	return []string{
// 		fmt.Sprintf("BOSH_CLIENT=%s", e.Credentials.ClientId),
// 		fmt.Sprintf("BOSH_CLIENT_SECRET=%s", e.Credentials.ClientSecret),
// 	}
// }

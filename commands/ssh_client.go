package commands

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

// go:generate counterfeiter -o ./fakes/ssh_client.go --fake-name SSHClient . SSHClient
type SSHClient interface {
	ExecuteOnRemote(input ExecuteOnRemoteInput) error
}

type ExecuteOnRemoteInput struct {
	Host       string
	SSHKeyPath string
	Env        []string
	Command    []string
}

type sshClient struct {
	stderr logger
	stdout logger
}

func NewSSHClient(stdout, stderr logger) SSHClient {
	return &sshClient{stdout: stdout, stderr: stderr}
}

func (s *sshClient) ExecuteOnRemote(input ExecuteOnRemoteInput) error {
	pemBytes, err := ioutil.ReadFile(input.SSHKeyPath)
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

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", input.Host), cfg)
	for connErr := err; connErr != nil; {
		if strings.Contains(connErr.Error(), "unexpected message type 3") {
			s.stderr.Printf("Failed to establish connection; retrying\n")
			client, connErr = ssh.Dial("tcp", fmt.Sprintf("%s:22", input.Host), cfg)
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

	fullcmd := strings.Join(append(input.Env, strings.Join(input.Command, " ")), " ")

	session.Stdout = os.Stdout
	err = session.Run(fullcmd)
	if err != nil {
		log.Fatalf("Run failed:%v", err)
	}

	return nil
}

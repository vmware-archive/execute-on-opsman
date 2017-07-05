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
	Host        string
	SSHKeyPath  string
	SSHPassword string
	Env         []string
	Command     []string
}

type sshClient struct {
	stderr logger
	stdout logger
}

func NewSSHClient(stdout, stderr logger) SSHClient {
	return &sshClient{stdout: stdout, stderr: stderr}
}

func (s *sshClient) ExecuteOnRemote(input ExecuteOnRemoteInput) error {
	var auths []ssh.AuthMethod

	if input.SSHPassword != "" {
		auths = []ssh.AuthMethod{ssh.Password(input.SSHPassword)}
	} else {
		pemBytes, err := ioutil.ReadFile(input.SSHKeyPath)
		if err != nil {
			log.Fatal(err)
		}

		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			log.Fatal(err)
		}

		auths = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	}

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

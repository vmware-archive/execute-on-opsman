package main

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/pivotal-cf-experimental/execute-on-opsman/commands"
	"github.com/pivotal-cf/om/api"
	omcommands "github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/flags"

	"github.com/pivotal-cf/om/network"
)

func main() {
	log.SetOutput(os.Stdout)

	stdout := log.New(os.Stdout, "", 0)
	stderr := log.New(os.Stderr, "", 0)

	var global struct {
		Target            string `short:"t" long:"target"              description:"location of the Ops Manager VM"`
		Username          string `short:"u" long:"username"            description:"admin username for the Ops Manager VM (not required for unauthenticated commands)"`
		Password          string `short:"p" long:"password"            description:"admin password for the Ops Manager VM (not required for unauthenticated commands)"`
		SkipSSLValidation bool   `short:"k" long:"skip-ssl-validation" description:"skip ssl certificate validation during http requests" default:"false"`
	}

	args, err := flags.Parse(&global, os.Args[1:])

	if err != nil {
		stdout.Fatal(err)
	}

	requestTimeout := time.Duration(1800) * time.Second
	authedClient, err := network.NewOAuthClient(global.Target, global.Username, global.Password, global.SkipSSLValidation, false, requestTimeout)
	if err != nil {
		stdout.Fatal(err)
	}
	requestService := api.NewRequestService(authedClient)

	var command string
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}

	uri, err := url.Parse(global.Target)
	if err != nil {
		stdout.Fatal(err)
	}

	commandSet := omcommands.Set{}
	commandSet["bosh"] = commands.NewBoshCommand(requestService, uri.Host, stdout, stderr)
	err = commandSet.Execute(command, args)
	if err != nil {
		stdout.Fatal(err)
	}
}

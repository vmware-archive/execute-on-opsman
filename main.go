package main

import (
	"flag"
	"log"
	"os"
)

var target = flag.String("target", "", "Ops Manager URL")
var username = flag.String("username", "", " Admin username for Ops Manager")
var password = flag.String("password", "", " Admin password for Ops Manager")
var sshKeyPath = flag.String("ssh-key-path", "", " Path to private key used to ssh into Ops Manager")

func parseParams() {
	flag.Parse()

	if *target == "" {
		log.Fatal("target flag is required")
	}
	if *username == "" {
		log.Fatal("username flag is required")
	}
	if *password == "" {
		log.Fatal("password flag is required")
	}
}

func main() {
	log.SetOutput(os.Stdout)

	parseParams()

	log.Println("Yeah")
}

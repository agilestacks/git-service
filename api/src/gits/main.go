package main

import (
	"log"

	"gits/api"
	"gits/config"
	"gits/flags"
	"gits/s3"
	"gits/ssh"
)

func main() {
	flags.Parse()
	s3.Init()
	ssh.Listen("0.0.0.0", config.SshPort)
	api.Listen("0.0.0.0", config.HttpPort)
	if config.Verbose {
		log.Print("Git Service started")
	}
	select {}
}

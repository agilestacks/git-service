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
	api.Init()
	s3.Init()
	ssh.Listen("0.0.0.0", config.SshPort)
	api.Listen("0.0.0.0", config.HttpPort)
	if config.Verbose {
		log.Printf("Git Service started on HTTP port %d, SSH port %d", config.HttpPort, config.SshPort)
	}
	select {}
}

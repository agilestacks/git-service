package main

import (
	"log"

	"github.com/agilestacks/git-service/cmd/gits/api"
	"github.com/agilestacks/git-service/cmd/gits/config"
	"github.com/agilestacks/git-service/cmd/gits/flags"
	"github.com/agilestacks/git-service/cmd/gits/s3"
	"github.com/agilestacks/git-service/cmd/gits/ssh"
	"github.com/agilestacks/git-service/cmd/gits/util"
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
	util.Maintenance()
	select {}
}

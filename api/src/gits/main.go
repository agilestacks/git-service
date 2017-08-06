package main

import (
	"gits/api"
	"gits/config"
	"gits/s3"
	"gits/ssh"
)

func main() {
	parseFlags()
	s3.Init()
	ssh.Listen("0.0.0.0", config.SshPort)
	api.Listen("0.0.0.0", config.HttpPort)
	select {}
}

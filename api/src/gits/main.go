package main

import (
	"gits/api"
	"gits/config"
	"gits/ssh"
)

func main() {
	parseFlags()
	ssh.Listen("0.0.0.0", config.SshPort)
	api.Listen("0.0.0.0", config.HttpPort)
	select {}
}

package main

import (
	"gits/config"
	"gits/ssh"
)

func main() {
	parseFlags()
	ssh.Listen("0.0.0.0", config.SshPort)
	select {}
}

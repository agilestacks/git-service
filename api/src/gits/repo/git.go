package repo

import (
	"log"
	"os"
	"os/exec"

	"gits/config"
)

var dirMode = os.FileMode(0755)

func gitBinPath() string {
	return gitSubCommandBinPath("git")
}

func gitSubCommandBinPath(command string) string {
	path, err := exec.LookPath(command)
	if err != nil {
		if config.Trace {
			log.Printf("Git binary `%s` lookup: %v; using %s", command, err, config.GitBinDefault)
		}
		path = config.GitBinDefault
	}
	return path
}

func gitDebug(cmd *exec.Cmd) {
	if config.Debug {
		cmd.Stderr = os.Stdout
		if config.Trace {
			cmd.Stdout = os.Stdout
		}
	}
}

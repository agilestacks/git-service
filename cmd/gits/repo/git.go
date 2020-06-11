package repo

import (
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/agilestacks/git-service/cmd/gits/config"
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

func printGitArgs(cmd *exec.Cmd) {
	log.Printf("%s (%s)", strings.Join(cmd.Args, " "), cmd.Dir)
}

func gitDebug(cmd *exec.Cmd) {
	if config.Debug {
		cmd.Stderr = os.Stdout
		if config.Trace {
			cmd.Stdout = os.Stdout
			printGitArgs(cmd)
		}
	}
}

func gitDebug2(cmd *exec.Cmd, stdoutCopy io.Writer) {
	if config.Trace {
		stdoutCopy = io.MultiWriter(stdoutCopy, os.Stdout)
		printGitArgs(cmd)
	}
	cmd.Stdout = stdoutCopy
	if config.Debug {
		cmd.Stderr = os.Stdout
	}
}

func gitDebug4(cmd *exec.Cmd, stdoutCopy io.Writer) {
	if config.Trace {
		printGitArgs(cmd)
	}
	cmd.Stdout = stdoutCopy
	if config.Debug {
		cmd.Stderr = os.Stdout
	}
}

func gitDebug3(cmd *exec.Cmd, stdoutCopy io.Writer, stderrCopy io.Writer) {
	if config.Debug {
		stderrCopy = io.MultiWriter(stderrCopy, os.Stdout)
	}
	if config.Trace {
		stdoutCopy = io.MultiWriter(stdoutCopy, os.Stdout)
		printGitArgs(cmd)
	}
	cmd.Stdout = stdoutCopy
	cmd.Stderr = stderrCopy
}

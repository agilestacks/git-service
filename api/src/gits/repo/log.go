package repo

import (
	"os"
	"os/exec"
	"path/filepath"

	"gits/config"
)

func Log(repoId string) ([]byte, error) {
	dir := filepath.Join(config.RepoDir, repoId)
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  dir,
		Args: []string{"git", "log"},
	}
	if config.Debug {
		cmd.Stderr = os.Stdout
	}
	return cmd.Output()
}

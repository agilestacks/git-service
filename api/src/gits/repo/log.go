package repo

import (
	"os"
	"os/exec"
	"path/filepath"

	"gits/config"
)

func Log(repoId, ref string) ([]byte, error) {
	if ref == "" {
		ref = "master"
	}
	dir := filepath.Join(config.RepoDir, repoId)
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  dir,
		Args: []string{"git", "log", ref},
	}
	if config.Debug {
		cmd.Stderr = os.Stdout
	}
	return cmd.Output()
}

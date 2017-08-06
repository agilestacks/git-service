package repo

import (
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
	return cmd.Output()
}

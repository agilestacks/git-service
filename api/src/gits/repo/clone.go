package repo

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"gits/config"
)

func InfoPack(repoId string, out io.Writer) error {
	dir := filepath.Join(config.RepoDir, repoId)
	command := "git-upload-pack"
	cmd := exec.Cmd{
		Path:   gitSubCommandBinPath(command),
		Dir:    dir,
		Args:   []string{command, "--stateless-rpc", "--advertise-refs", "."},
		Stdout: out,
	}
	if config.Debug {
		cmd.Stderr = os.Stdout
	}
	return cmd.Run()
}

func RefsPack(repoId string, out io.Writer, in io.Reader) error {
	dir := filepath.Join(config.RepoDir, repoId)
	command := "git-upload-pack"
	cmd := exec.Cmd{
		Path:   gitSubCommandBinPath(command),
		Dir:    dir,
		Args:   []string{command, "--stateless-rpc", "."},
		Stdout: out,
		Stdin:  in,
	}
	if config.Debug {
		cmd.Stderr = os.Stdout
	}
	return cmd.Run()
}

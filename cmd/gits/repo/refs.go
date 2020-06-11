package repo

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/agilestacks/git-service/cmd/gits/config"
)

func Exist(repoId string) bool {
	dir := filepath.Join(config.RepoDir, repoId)
	_, err := os.Stat(dir)
	return err == nil
}

// `service` parameter is validated by HTTP handling layer and
// is verified to be one of git-upload-pack, git-receive-pack

func RefsInfo(repoId string, service string, out io.Writer) error {
	dir := filepath.Join(config.RepoDir, repoId)
	cmd := exec.Cmd{
		Path:   gitSubCommandBinPath(service),
		Dir:    dir,
		Args:   []string{service, "--stateless-rpc", "--advertise-refs", "."},
		Stdout: out,
	}
	if config.Trace {
		log.Printf("Git refs info: %s %v %s", cmd.Path, cmd.Args, cmd.Dir)
	}
	gitDebug2(&cmd, out)
	return cmd.Run()
}

func Pack(repoId string, service string, out io.Writer, in io.Reader) error {
	dir := filepath.Join(config.RepoDir, repoId)
	cmd := exec.Cmd{
		Path:   gitSubCommandBinPath(service),
		Dir:    dir,
		Args:   []string{service, "--stateless-rpc", "."},
		Stdout: out,
		Stdin:  in,
	}
	if config.Trace {
		log.Printf("Git pack: %s %v %s", cmd.Path, cmd.Args, cmd.Dir)
	}
	gitDebug4(&cmd, out)
	return cmd.Run()
}

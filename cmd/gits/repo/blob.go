package repo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/agilestacks/git-service/cmd/gits/config"
)

func Blob(repoId, ref, path string) (io.ReadCloser, error) {
	dir := filepath.Join(config.RepoDir, repoId)
	if ref == "" {
		ref = "master"
	}

	// temp dir for work tree
	worktree, err := ioutil.TempDir("", "gits-")
	if err != nil {
		return nil, fmt.Errorf("Unable to create temporary directory: %v", err)
	}
	defer deleteDir(worktree)

	// checkout
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  dir,
		Args: []string{"git", "--work-tree", worktree, "checkout", ref, "--", path},
	}
	var stdoutBuffer bytes.Buffer
	gitDebug2(&cmd, &stdoutBuffer)
	err = cmd.Run()
	if err != nil {
		if strings.Contains(stdoutBuffer.String(), "dit not match any") {
			return nil, errors.New("not found")
		}
		return nil, fmt.Errorf("Unable to checkout `%s` path `%s` at ref `%s`: %v", repoId, path, ref, err)
	}
	fullPath := filepath.Join(worktree, path)
	if !strings.HasPrefix(fullPath, worktree) {
		return nil, fmt.Errorf("Path `%s` is not under work tree `%s` while processing download `%s`",
			path, worktree, path)
	}

	info, err := os.Lstat(fullPath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("is a directory")
	}
	if info.Mode()&os.ModeType != 0 {
		return nil, errors.New("not a regular file")
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

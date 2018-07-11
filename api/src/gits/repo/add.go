package repo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gits/config"
)

type AddFile struct {
	Path    string
	Content io.Reader
	Mode    os.FileMode
}

func Add(repoId string, files []AddFile, commitMessage string) error {
	dir := filepath.Join(config.RepoDir, repoId)

	// temp dir for clone
	clone, err := ioutil.TempDir("", "gits-")
	if err != nil {
		return fmt.Errorf("Unable to create temporary directory: %v", err)
	}
	defer deleteDir(clone)

	// clone
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  "/",
		Args: []string{"git", "clone", dir, clone},
	}
	gitDebug(&cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Unable to clone `%s` into `%s`: %v", repoId, clone, err)
	}

	// add files
	filesArgs := make([]string, 0, len(files))
	for _, file := range files {
		fileDir := filepath.Dir(file.Path)
		if fileDir != "." {
			err := os.MkdirAll(filepath.Join(clone, fileDir), dirMode)
			if err != nil {
				return err
			}
		}
		fullPath := filepath.Join(clone, file.Path)
		mode := file.Mode
		if mode == 0 {
			mode = 0644
		}
		out, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, mode)
		if err != nil {
			return err
		}
		io.Copy(out, file.Content)
		out.Close()

		filesArgs = append(filesArgs, file.Path)
	}
	cmd = exec.Cmd{
		Path: gitBinPath(),
		Dir:  clone,
		Args: append([]string{"git", "add"}, filesArgs...),
	}
	gitDebug(&cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Unable to add `%s` to `%s` clone `%s`: %v",
			strings.Join(filesArgs, ","), repoId, clone, err)
	}

	// commit
	if commitMessage == "" {
		commitMessage = "Add files"
	}
	cmd = exec.Cmd{
		Path: gitBinPath(),
		Dir:  clone,
		Args: []string{"git", "commit", "-m", commitMessage},
	}
	var stdoutBuffer bytes.Buffer
	gitDebug2(&cmd, &stdoutBuffer)
	err = cmd.Run()
	if err != nil {
		if strings.Contains(stdoutBuffer.String(), "nothing to commit, working tree clean") {
			return nil
		}
		return fmt.Errorf("Unable to commit in `%s` clone `%s`: %v", repoId, clone, err)
	}

	// push
	cmd = exec.Cmd{
		Path: gitBinPath(),
		Dir:  clone,
		Args: []string{"git", "push"},
	}
	gitDebug(&cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Unable to push into `%s` from clone `%s`: %v", repoId, clone, err)
	}

	if config.Verbose {
		added := make([]string, 0, len(files))
		for _, file := range files {
			mode := ""
			if file.Mode > 0 {
				mode = fmt.Sprintf(" (%04o)", file.Mode)
			}
			added = append(added, file.Path+mode)
		}
		log.Printf("Added `%s` to `%s`", strings.Join(added, ", "), repoId)
	}

	return nil
}

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

func Add(repoId, branch string, files []AddFile, commitMessage string) error {
	dir := filepath.Join(config.RepoDir, repoId)
	if branch == "" {
		branch = "master"
	}

	// temp dir for work tree
	worktree, err := ioutil.TempDir("", "gits-")
	if err != nil {
		return fmt.Errorf("Unable to create temporary directory: %v", err)
	}
	defer deleteDir(worktree)

	gitBin := gitBinPath()
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer

	// show refs to determine if the repo is empty (has the branch) or not
	cmd := exec.Cmd{
		Path: gitBin,
		Dir:  dir,
		Args: []string{"git", "show-ref"},
	}
	gitDebug3(&cmd, &stdoutBuffer, &stderrBuffer)
	err = cmd.Run()
	if err != nil {
		if stdoutBuffer.Len() > 0 || stderrBuffer.Len() > 0 {
			return fmt.Errorf("Unable to retrieve `%s` Git refs: %v", repoId, err)
		}
	}
	ref := branch
	if !strings.HasPrefix(ref, "refs/") {
		ref = "refs/heads/" + ref
	}
	repoEmpty := !strings.Contains(stdoutBuffer.String(), " "+ref)

	// clone or init work tree
	if repoEmpty {
		cmd = exec.Cmd{
			Path: gitBin,
			Dir:  worktree,
			Args: []string{"git", "init"},
		}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Unable to init Git repository `%s` work tree: %v", worktree, err)
		}
		cmd = exec.Cmd{
			Path: gitBin,
			Dir:  worktree,
			Args: []string{"git", "checkout", "-b", branch},
		}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Unable to create Git branch `%s` in `%s` work tree: %v", branch, worktree, err)
		}
	} else {
		cmd := exec.Cmd{
			Path: gitBin,
			Dir:  "/",
			Args: []string{"git", "clone", "--branch", branch, "--single-branch", "--no-tags", dir, worktree},
		}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Unable to Git clone `%s` into `%s` work tree: %v", repoId, worktree, err)
		}
	}

	// add files
	filesArgs := make([]string, 0, len(files))
	for _, file := range files {
		fileDir := filepath.Dir(file.Path)
		if fileDir != "." {
			err := os.MkdirAll(filepath.Join(worktree, fileDir), dirMode)
			if err != nil {
				return err
			}
		}
		fullPath := filepath.Join(worktree, file.Path)
		if !strings.HasPrefix(fullPath, worktree) {
			return fmt.Errorf("Path `%s` is not under work tree `%s` while processing upload `%s`",
				fullPath, worktree, file.Path)
		}
		mode := file.Mode
		if mode == 0 {
			mode = 0644
		}
		out, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, file.Content)
		out.Close()
		if err != nil {
			return err
		}
		filesArgs = append(filesArgs, file.Path)
	}
	cmd = exec.Cmd{
		Path: gitBin,
		Dir:  worktree,
		Args: append([]string{"git", "add"}, filesArgs...),
	}
	gitDebug(&cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Unable to add `%s` to `%s` work tree `%s`: %v",
			strings.Join(filesArgs, ","), repoId, worktree, err)
	}

	// commit
	if commitMessage == "" {
		commitMessage = "Add files"
	}
	cmd = exec.Cmd{
		Path: gitBin,
		Dir:  worktree,
		Args: []string{"git", "commit", "-m", commitMessage},
	}
	stdoutBuffer.Reset()
	gitDebug2(&cmd, &stdoutBuffer)
	err = cmd.Run()
	if err != nil {
		if strings.Contains(stdoutBuffer.String(), "nothing to commit, working tree clean") {
			return nil
		}
		return fmt.Errorf("Unable to commit in `%s` work tree `%s`: %v", repoId, worktree, err)
	}

	// push
	if repoEmpty {
		cmd = exec.Cmd{
			Path: gitBin,
			Dir:  worktree,
			Args: []string{"git", "remote", "add", "origin", dir},
		}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Unable to add origin remote `%s` to `%s` work tree: %v", dir, worktree, err)
		}
	}
	cmd = exec.Cmd{
		Path: gitBin,
		Dir:  worktree,
		Args: []string{"git", "push", "origin", branch},
	}
	gitDebug(&cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Unable to push `%s` into `%s` from `%s` work tree: %v", branch, repoId, worktree, err)
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

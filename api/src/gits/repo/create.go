package repo

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gits/config"
	"gits/s3"
)

type CreateRequest struct {
	Remote  string
	Ref     string
	Squash  bool
	Message string
	Archive string
}

func Create(repoId string, req *CreateRequest) error {
	dir := filepath.Join(config.RepoDir, repoId)
	_, err := os.Stat(dir)
	if err == nil {
		return fmt.Errorf("Directory already exists: %s", dir)
	}
	err = os.MkdirAll(dir, dirMode)
	if err != nil {
		return err
	}
	if req != nil && req.Archive != "" {
		err = initWithArchive(dir, req.Archive)
	} else if req != nil && req.Remote != "" {
		if req.Squash {
			err = errors.New("Squash not implemented")
		} else {
			err = initWithRemote(dir, req.Remote, req.Ref)
		}
	} else {
		err = initBare(dir)
	}
	if err != nil {
		deleteDir(dir)
	} else {
		if config.Verbose {
			log.Printf("Repo `%s` created", repoId)
		}
	}
	return err
}

func initBare(dir string) error {
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  dir,
		Args: []string{"git", "init", "--bare"},
	}
	gitDebug(&cmd)
	err := cmd.Run()
	if err != nil {
		log.Printf("`git init` failed: %v", err)
		return err
	}
	return nil
}

func initWithArchive(dir string, archive string) error {
	if !strings.HasPrefix(archive, "s3://") {
		return fmt.Errorf("Archive `%s` scheme not supported, only `s3://` scheme is supported", archive)
	}
	return s3.Unarchive(dir, archive)
}

func initWithRemote(dir, remote, ref string) error {
	if ref == "" {
		ref = "master"
	}
	err := initBare(dir)
	if err != nil {
		return err
	}
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  dir,
		Args: []string{"git", "fetch", "-n", remote, ref + ":master"},
	}
	gitDebug(&cmd)
	err = cmd.Run()
	if err != nil {
		log.Printf("`git fetch` failed: %v", err)
		return err
	}
	return nil
}

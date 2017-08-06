package repo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gits/config"
	"gits/s3"
)

func Create(repoId string, archive string) error {
	dir := filepath.Join(config.RepoDir, repoId)
	_, err := os.Stat(dir)
	if err == nil {
		return fmt.Errorf("Directory already exists: %s", dir)
	}
	err = os.MkdirAll(dir, dirMode)
	if err != nil {
		return err
	}
	if archive == "" {
		err = initBare(dir)
	} else {
		err = initWithArchive(dir, archive)
	}
	if err != nil {
		deleteDir(dir)
	}
	return err
}

func initBare(dir string) error {
	cmd := exec.Cmd{
		Path: gitBinPath(),
		Dir:  dir,
		Args: []string{"git", "init", "--bare", "."},
	}
	_, err := cmd.Output()
	if err != nil {
		log.Printf("`git init %s` failed: %v", err)
		return err
	}
	return nil
}

func initWithArchive(dir string, archive string) error {
	if !strings.HasPrefix(archive, "s3://") {
		return fmt.Errorf("Archive `%s` scheme not supported, only `s3://`", archive)
	}
	return s3.Unarchive(dir, archive)
}

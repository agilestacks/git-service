package repo

import (
	"log"
	"os"
	"path/filepath"

	"github.com/agilestacks/git-service/cmd/gits/config"
)

func Delete(repoId string) error {
	dir := filepath.Join(config.RepoDir, repoId)
	return deleteDir(dir)
}

func deleteDir(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Printf("Unable to remove %s: %v", dir, err)
	}
	return err
}

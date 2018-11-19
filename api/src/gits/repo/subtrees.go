package repo

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gits/config"
)

type AddSubtree struct {
	Prefix     string
	Repository string
	Remote     string
	Ref        string
	Squash     bool
}

func AddSubtrees(repoId string, subtrees []AddSubtree) error {
	dir := filepath.Join(config.RepoDir, repoId)

	// validate, set defaults
	for i, subtree := range subtrees {
		if subtree.Remote == "" && subtree.Repository != "" { // compat
			subtrees[i].Remote = subtree.Repository
			subtree.Remote = subtree.Repository
		}
		if subtree.Prefix == "" || subtree.Remote == "" ||
			!(strings.HasPrefix(subtree.Remote, "http:") || strings.HasPrefix(subtree.Remote, "https:") ||
				strings.HasPrefix(subtree.Remote, "git:") || strings.HasPrefix(subtree.Remote, "git@")) {
			return fmt.Errorf("Invalid subtree spec at index %d: %+v ", i, subtree)
		}
		if subtree.Ref == "" {
			subtrees[i].Ref = "master"
		}
	}

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

	// check path does not exist
	for _, subtree := range subtrees {
		subtreeDir := filepath.Join(clone, subtree.Prefix)
		_, err = os.Stat(subtreeDir)
		if err == nil {
			return fmt.Errorf("Path `%s` already exist", subtree.Prefix)
		}
		if !noSuchFile(err) {
			return fmt.Errorf("Unable to stat path `%s` in repo clone `%s`: %v",
				subtree.Prefix, subtreeDir, err)
		}
	}

	// add subtrees
	for _, subtree := range subtrees {
		args := []string{"git", "subtree", "add", "--prefix=" + subtree.Prefix, subtree.Remote, subtree.Ref}
		if subtree.Squash {
			args = append(args, "--squash")
		}
		cmd = exec.Cmd{
			Path: gitBinPath(),
			Dir:  clone,
			Args: args,
		}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Unable to add subtree `%s` ref `%s`: %v", maskAuth(subtree.Remote), subtree.Ref, err)
		}
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
		return fmt.Errorf("Unable to push repo clone `%s`: %v", clone, err)
	}

	if config.Verbose {
		if config.Trace {
			log.Printf("Subtrees added to `%s` repo:", repoId)
			for _, subtree := range subtrees {
				log.Printf("\t%s => %s @ %s", subtree.Prefix, maskAuth(subtree.Remote), subtree.Ref)
			}
		} else {
			added := make([]string, 0, len(subtrees))
			for _, subtree := range subtrees {
				added = append(added, subtree.Prefix)
			}
			log.Printf("Subtrees added to `%s` repo: %v", repoId, added)
		}
	}

	return nil
}

func maskAuth(maybeUrl string) string {
	remote, err := url.Parse(maybeUrl)
	if err != nil {
		return maybeUrl
	}
	remote.User = url.UserPassword("masked", "")
	return remote.String()
}

func noSuchFile(err error) bool {
	str := err.Error()
	return str == "file does not exist" ||
		strings.Contains(str, "no such file or directory")
}

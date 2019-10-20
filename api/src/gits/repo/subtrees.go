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
	Prefix      string
	Remote      string
	Ref         string
	SplitPrefix string `json:"splitPrefix"`
	splitBranch string
	Squash      bool
}

type RemoteWithRef struct {
	Remote string
	Ref    string
}

// TODO add subtree to an empty branch
func AddSubtrees(repoId, branch string, subtrees []AddSubtree) error {
	dir := filepath.Join(config.RepoDir, repoId)
	if branch == "" {
		branch = "master"
	}

	// validate, set defaults
	for i, subtree := range subtrees {
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

	gitBin := gitBinPath()
	// clone
	cmd := exec.Cmd{
		Path: gitBin,
		Dir:  "/",
		Args: []string{"git", "clone", "--branch", branch, "--single-branch", "--no-tags", dir, clone},
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

	// see if there are subtrees that requires splitting a prefix
	subtreeRemotes := make(map[string]RemoteWithRef)
	for _, subtree := range subtrees {
		if subtree.SplitPrefix != "" {
			key := fmt.Sprintf("%s|%s", subtree.Remote, subtree.Ref)
			subtreeRemotes[key] = RemoteWithRef{subtree.Remote, subtree.Ref}
		}
	}
	// for each unique combination of such remote/ref, fetch the ref and checkout as local branch
	remoteIndex := 0
	splitBranchIndex := 0
	for _, remote := range subtreeRemotes {
		remoteName := fmt.Sprintf("remote-%d", remoteIndex)
		remoteIndex++
		// git remote add remote-0 git@github.com:agilestacks/components.git
		args := []string{"git", "remote", "add", remoteName, remote.Remote}
		cmd = exec.Cmd{Path: gitBin, Dir: clone, Args: args}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Unable to add remote `%s`: %v", maskAuth(remote.Remote), err)
		}
		// git fetch remote-0 distribution:_remote-0/distribution
		args = []string{"git", "fetch", remoteName, fmt.Sprintf("%s:_%s/%[1]s", remote.Ref, remoteName)}
		cmd = exec.Cmd{Path: gitBin, Dir: clone, Args: args}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Unable to fetch `%s` ref `%s`: %v", maskAuth(remote.Remote), remote.Ref, err)
		}
		// git checkout -b _remote-0-distribution _remote-0/distribution
		args = []string{"git", "checkout", "-b", fmt.Sprintf("_%s-%s", remoteName, remote.Ref),
			fmt.Sprintf("_%s/%s", remoteName, remote.Ref)}
		cmd = exec.Cmd{Path: gitBin, Dir: clone, Args: args}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Unable to checkout `%s` ref `%s` as local branch: %v", maskAuth(remote.Remote), remote.Ref, err)
		}
		// now, for each subtree from remote/ref, split prefix into a branch
		for i, subtree := range subtrees {
			if subtree.SplitPrefix == "" || !(remote.Remote == subtree.Remote && remote.Ref == subtree.Ref) {
				continue
			}
			splitBranchName := fmt.Sprintf("_split-%d", splitBranchIndex)
			splitBranchIndex++
			// git subtree split --prefix=pgweb -b _split-0
			// TODO -q is too quiet
			args = []string{"git", "subtree", "split", "-q", "--prefix=" + subtree.SplitPrefix, "-b", splitBranchName}
			if subtree.Squash {
				args = append(args, "--squash")
			}
			cmd = exec.Cmd{Path: gitBin, Dir: clone, Args: args}
			gitDebug(&cmd)
			err = cmd.Run()
			if err != nil {
				return fmt.Errorf("Unable to split `%s` ref `%s` prefix `%s` as local branch: %v",
					maskAuth(remote.Remote), remote.Ref, subtree.SplitPrefix, err)
			}
			subtrees[i].splitBranch = splitBranchName
		}
	}

	// return to the branch
	if len(subtreeRemotes) > 0 {
		cmd := exec.Cmd{
			Path: gitBin,
			Dir:  clone,
			Args: []string{"git", "checkout", branch},
		}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("Unable to checkout `%s`: %v", branch, err)
		}
	}

	// add subtrees
	for _, subtree := range subtrees {
		args := []string{"git", "subtree", "add", "--prefix=" + subtree.Prefix}
		if subtree.SplitPrefix == "" {
			args = append(args, subtree.Remote, subtree.Ref)
		} else {
			args = append(args, subtree.splitBranch)
		}
		if subtree.Squash {
			args = append(args, "--squash")
		}
		cmd = exec.Cmd{Path: gitBin, Dir: clone, Args: args}
		gitDebug(&cmd)
		err = cmd.Run()
		if err != nil {
			prefix := ""
			if subtree.SplitPrefix != "" {
				prefix = fmt.Sprintf(" prefix `%s`", subtree.SplitPrefix)
			}
			return fmt.Errorf("Unable to add subtree `%s` ref `%s`%s: %v", maskAuth(subtree.Remote), subtree.Ref, prefix, err)
		}
	}

	// push
	cmd = exec.Cmd{
		Path: gitBin,
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
				prefix := ""
				if subtree.SplitPrefix != "" {
					prefix = " /" + subtree.SplitPrefix
				}
				log.Printf("\t%s => %s%s @ %s", subtree.Prefix, maskAuth(subtree.Remote), prefix, subtree.Ref)
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

package repo

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"gits/config"
)

const (
	gitBinDefault = "/usr/bin/git"
)

func GitServer(command string, stdin io.Reader, stdout io.Writer, stderr io.Writer, users []string) (*exec.Cmd, error) {
	if config.Debug {
		log.Printf("Git command requested: %q", command)
	}
	parts := strings.SplitN(command, " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("Unknown repo name in %q", command)
	}

	verb := parts[0]
	if !allowedVerb(verb) {
		return nil, fmt.Errorf("%q is not allowed Git sub-command", verb)
	}

	repo := parts[1]
	repo = strings.TrimLeft(repo, "'/")
	repo = strings.TrimRight(repo, "'")
	repo = strings.TrimSuffix(repo, ".git")
	if config.Debug {
		log.Printf("Git command parsed: %s %s", verb, repo)
	}

	hasAccess, err := access(repo, verb, users)
	if err != nil {
		log.Printf("Checking `%s` repo permissions for %v: %v", repo, users, err)
	}
	if !hasAccess {
		err := fmt.Errorf("%v have no access to `%s`", users, repo)
		if config.Verbose {
			log.Printf("%v", err)
		}
		return nil, err
	} else {
		if config.Debug {
			log.Printf("%v have access to `%s`", users, repo)
		}
	}

	repoPath := filepath.Join(config.RepoDir, repo)
	gitBinPath, err := exec.LookPath(verb)
	if err != nil {
		if config.Trace {
			log.Printf("Git binary lookup for %s: %v; using %s", verb, err, gitBinDefault)
		}
		gitBinPath = gitBinDefault
	}

	cmd := exec.Cmd{
		Path: gitBinPath,
		// Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Dir:    repoPath,
		Args:   []string{verb, "."},
	}
	// a workaround for bizarre Wait() lockup
	inputPipe, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	go io.Copy(inputPipe, stdin)
	if config.Trace {
		log.Printf("Starting Git:\n\t%+v", cmd)
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	return &cmd, nil
}

var allowedVerbs = []string{"git-receive-pack", "git-upload-archive", "git-upload-pack"}

func allowedVerb(verb string) bool {
	for _, allowed := range allowedVerbs {
		if verb == allowed {
			return true
		}
	}
	return false
}

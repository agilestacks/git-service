package config

import (
	"log"
	"path/filepath"
	"strings"
)

var (
	Verbose  bool
	Debug    bool
	Trace    bool
	ZipTrace bool

	RepoDir     string
	HttpPort    int
	SshPort     int
	HostKeyFile string
	BlobsFrom   []string
	ApiSecret   string
)

func Update() {
	if Trace {
		Debug = true
	}
	if Debug {
		Verbose = true
	}
	for _, from := range BlobsFrom {
		if !strings.HasPrefix(from, "s3://") {
			log.Fatalf("Blob prefix `%s` not supported", from)
		}
	}
	if RepoDir == "" {
		RepoDir = "."
	} else if strings.HasPrefix(RepoDir, string(filepath.Separator)) {
		if len(RepoDir) == 1 {
			log.Fatalf("Repo directory `%s` is root directory?", RepoDir)
		}
	}
}

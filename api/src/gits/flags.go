package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gits/config"
)

func parseFlags() {
	var blobsFrom, apiSecretEnvVar string

	flag.BoolVar(&config.Verbose, "verbose", true, "Print progress if set")
	flag.BoolVar(&config.Debug, "debug", false, "Print debug information if set")
	flag.BoolVar(&config.Trace, "trace", false, "Print detailed trace if set")
	flag.BoolVar(&config.ZipTrace, "zip_trace", false, "Print detailed debug of S3 ZIP streaming if set")

	flag.StringVar(&config.RepoDir, "repo_dir", "/git", "Base directory for Git repositories")
	flag.StringVar(&blobsFrom, "blobs", "", "Allowed URL prefixes to fetch repo sources from, empty for no restrictions")
	flag.IntVar(&config.HttpPort, "http_port", 8005, "HTTP API port to listen")
	flag.IntVar(&config.SshPort, "ssh_port", 2022, "SSH port to listen")
	flag.StringVar(&config.HostKeyFile, "host_key", "gits-key", "Path to SSH server host private key file")
	flag.StringVar(&apiSecretEnvVar, "secret_env", "GIT_API_SECRET", "Environment variable to get secret from to protect HTTP API")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr,
			`Usage:
  gits -verbose -repo_dir /git -http_port 80 -ssh_port 22 -blobsFrom s3://agilestacks/

  If no -blobsFrom is set then any supported URL is allowed, currently s3://
  If -secret_env is empty or env variable exist but is empty then HTTP API is open - no access control.

Flags:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	if apiSecretEnvVar != "" {
		var exist bool
		config.ApiSecret, exist = os.LookupEnv(apiSecretEnvVar)
		if !exist {
			log.Fatalf("`-secret_env %s` is set but variable not found in process environment, try `gits -h`",
				apiSecretEnvVar)
		}
	}
	if blobsFrom != "" {
		config.BlobsFrom = strings.Split(blobsFrom, ",")
	}

	config.Update()
}

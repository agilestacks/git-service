package repo

import (
    "log"
    "os/exec"

    "gits/config"
)

func gitBinPath() string {
    path, err := exec.LookPath("git")
    if err != nil {
        if config.Trace {
            log.Printf("Git binary lookup: %v; using %s", err, config.GitBinDefault)
        }
        path = config.GitBinDefault
    }
    return path
}

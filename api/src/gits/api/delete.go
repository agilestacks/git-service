package api

import (
    "fmt"
	"net/http"
    "log"

    "gits/config"
    "gits/repo"
)

func deleteRepo(repoId string, w http.ResponseWriter) {
    err := repo.Delete(repoId)
    if err != nil {
        message := fmt.Sprintf("Unable to delete Git repo `%s`: %v", repoId, err)
        log.Print(message)
        writeError(w, http.StatusInternalServerError, message)
    } else {
        if config.Verbose {
            log.Printf("Repo `%s` deleted", repoId)
        }
        w.WriteHeader(http.StatusNoContent)
    }
}

package api

import (
	"fmt"
	"log"
	"net/http"

	"gits/config"
	"gits/repo"
)

func sendRepoLog(repoId string, w http.ResponseWriter) {
	logBytes, err := repo.Log(repoId)
	if err != nil {
		message := fmt.Sprintf("Unable to obtain Git repo `%s` log: %v", repoId, err)
		log.Print(message)
		writeError(w, http.StatusInternalServerError, message)
	} else {
		if config.Verbose {
			log.Printf("Sending repo `%s` log", repoId)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(logBytes)
	}
}

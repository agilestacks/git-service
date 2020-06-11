package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/agilestacks/git-service/cmd/gits/config"
	"github.com/agilestacks/git-service/cmd/gits/repo"
)

func sendRepoLog(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

	ref := req.URL.Query().Get("ref")

	logBytes, err := repo.Log(repoId, ref)
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

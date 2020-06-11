package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/agilestacks/git-service/cmd/gits/config"
	"github.com/agilestacks/git-service/cmd/gits/repo"
)

func deleteRepo(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

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

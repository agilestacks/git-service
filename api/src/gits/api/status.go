package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"gits/config"
	"gits/repo"
)

func sendRepoStatus(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

	ref := req.URL.Query().Get("ref")

	status, err := repo.Status(repoId, ref)
	if err != nil {
		// TODO return partial status if any? (status != nil)
		message := fmt.Sprintf("Unable to obtain Git repo `%s` status: %v", repoId, err)
		log.Print(message)
		httpStatus := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			httpStatus = http.StatusNotFound
		} else if strings.Contains(err.Error(), "multiple refs") {
			httpStatus = http.StatusConflict
		}
		writeError(w, httpStatus, message)
	} else {
		statusBytes, err := json.Marshal(status)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("Unable to marshall JSON: %v", err))
			return
		}
		if config.Verbose {
			log.Printf("Sending repo `%s` status", repoId)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(statusBytes)
	}
}

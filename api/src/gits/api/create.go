package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"gits/repo"
)

func createRepo(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error reading request body: %v", err))
		return
	}
	var createReq *repo.CreateRequest
	if len(body) > 4 {
		var reqData repo.CreateRequest
		err = json.Unmarshal(body, &reqData)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Error unmarshalling JSON request: %v", err))
			return
		}
		createReq = &reqData
	}

	err = repo.Create(repoId, createReq)
	if err != nil {
		message := fmt.Sprintf("Unable to create Git repo `%s`: %v", repoId, err)
		log.Print(message)
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "exist") {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "not supported") {
			status = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "not implemented") {
			status = http.StatusNotImplemented
		} else if strings.Contains(err.Error(), "S3") {
			status = http.StatusBadGateway
		}
		writeError(w, status, message)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

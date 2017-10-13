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

type CreateRequest struct {
	Archive string
}

func createRepo(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error reading request body: %v", err))
		return
	}
	archive := ""
	if len(body) > 4 {
		var reqData CreateRequest
		err = json.Unmarshal(body, &reqData)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("Error unmarshalling JSON request: %v", err))
			return
		}
		archive = reqData.Archive
	}

	err = repo.Create(repoId, archive)
	if err != nil {
		message := fmt.Sprintf("Unable to create Git repo `%s`: %v", repoId, err)
		log.Print(message)
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "exist") {
			status = http.StatusConflict
		} else if strings.Contains(err.Error(), "not supported") {
			status = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "S3") {
			status = http.StatusGatewayTimeout
		}
		writeError(w, status, message)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

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

type SubtreesRequest struct {
	Subtrees []repo.AddSubtree
}

func addSubtrees(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])
	branch := req.URL.Query().Get("ref")

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("Error reading request body: %v", err))
		return
	}

	var reqData SubtreesRequest
	err = json.Unmarshal(body, &reqData)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Error unmarshalling JSON request: %v", err))
		return
	}

	if len(reqData.Subtrees) == 0 {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Request `subtrees` is empty"))
		return
	}

	err = repo.AddSubtrees(repoId, branch, reqData.Subtrees)
	if err != nil {
		message := fmt.Sprintf("Unable to add subtrees to Git repo `%s`: %v", repoId, err)
		log.Print(message)
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "exist") {
			status = http.StatusConflict
		}
		writeError(w, status, message)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

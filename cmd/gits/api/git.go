package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/agilestacks/git-service/cmd/gits/config"
	"github.com/agilestacks/git-service/cmd/gits/repo"
)

// used by api/ test
var InfoPack = repo.RefsInfo

/* https://github.com/git/git/blob/master/Documentation/technical/http-protocol.txt
   https://git-scm.com/book/en/v2/Git-Internals-Transfer-Protocols
   https://github.com/go-gitea/gitea/blob/HEAD/routers/repo/http.go */

func checkGitService(req *http.Request) bool {
	vars := mux.Vars(req)
	service := vars["service"]

	return service == "git-upload-pack" || service == "git-receive-pack"
}

func checkRepoExist(req *http.Request) bool {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

	return repo.Exist(repoId)
}

func checkUserRepoAccess(req *http.Request) bool {
	username, password, ok := req.BasicAuth()
	if !ok {
		return false
	}
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])
	service := vars["service"]

	deploymentKey := ""
	if len(username) >= deploymentKeyMinHexLen {
		deploymentKey = username
	} else if len(password) >= deploymentKeyMinHexLen {
		deploymentKey = password
	}

	hasAccess := false

	if deploymentKey != "" {
		decodedUsername, decodedSubject, decodeErr := decodeDeploymentKey(deploymentKey)
		var accessErr error
		hasAccess, accessErr = repo.Access(repoId, service, []string{decodedUsername})
		if decodeErr != nil || (accessErr != nil && !hasAccess) {
			log.Printf("No %s access to `%s` for token `%s...` user `%s`: %v",
				service, repoId, deploymentKey[0:8], decodedUsername, seeErrors2(decodeErr, accessErr))
			return false
		}
		if hasAccess && decodedSubject != "" {
			hasAccess = false
			subjectPrefix := "git:"
			if len(decodedSubject) > len(subjectPrefix) {
				i := strings.Index(decodedSubject, subjectPrefix)
				if i == 0 {
					subjectId := decodedSubject[len(subjectPrefix):]
					templateId, err := repo.TemplateId(repoId)
					if err == nil && templateId == subjectId {
						hasAccess = true
					}
				}
			}
		}
	} else {
		var err error
		hasAccess, err = repo.AccessWithLogin(vars["organization"], repoId, service, username, password)
		if err != nil {
			log.Printf("No %s access to `%s` for user `%s`: %v", service, repoId, username, err)
			return false
		}
	}

	return hasAccess
}

func refsInfo(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])
	service := vars["service"]

	if config.Verbose {
		log.Printf("Repo `%s` %s refs", repoId, service)
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(gitRpcPacket(fmt.Sprintf("# service=%s\n", service))))
	w.Write([]byte("0000"))

	err := InfoPack(repoId, service, w)
	if err != nil {
		log.Printf("Got error from Git while %s repo `%s` refs: %v", service, repoId, err)
	}
}

func gitRpcPacket(str string) string {
	s := strconv.FormatInt(int64(len(str)+4), 16)
	off := len(s) % 4
	if off != 0 {
		s = strings.Repeat("0", 4-off) + s
	}
	return s + str
}

func pack(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])
	service := vars["service"]

	if config.Verbose {
		log.Printf("Repo `%s` %s pack", repoId, service)
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-result", service))
	w.WriteHeader(http.StatusOK)

	err := repo.Pack(repoId, service, w, req.Body)
	if err != nil {
		log.Printf("Got error from Git while %s repo `%s` pack: %v", service, repoId, err)
	}
}

package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gits/config"
	"gits/repo"
)

/* https://github.com/git/git/blob/master/Documentation/technical/http-protocol.txt
   https://git-scm.com/book/en/v2/Git-Internals-Transfer-Protocols
   https://github.com/go-gitea/gitea/blob/HEAD/routers/repo/http.go */

func sendRefsInfo(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

	if config.Verbose {
		log.Printf("Sending repo `%s` refs", repoId)
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "application/x-git-upload-pack-advertisement")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(gitRpcPacket("# service=git-upload-pack\n")))
	w.Write([]byte("0000"))

	err := repo.InfoPack(repoId, w)
	if err != nil {
		message := fmt.Sprintf("Unable to send Git repo `%s` refs: %v", repoId, err)
		log.Print(message)
		strings.NewReader(message).WriteTo(w)
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

func sendRefsPack(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	repoId := getRepositoryId(vars["organization"], vars["repository"])

	if config.Verbose {
		log.Printf("Sending repo `%s` pack", repoId)
	}

	w.Header().Set("Content-Type", "application/x-git-upload-pack-result")
	w.WriteHeader(http.StatusOK)

	err := repo.RefsPack(repoId, w, req.Body)
	if err != nil {
		message := fmt.Sprintf("Unable to send Git repo `%s` pack: %v", repoId, err)
		log.Print(message)
		strings.NewReader(message).WriteTo(w)
	}
}

package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMuxCapturesEverythingAfterSlash(t *testing.T) {
	r := mux.NewRouter()
	r.PathPrefix("/api/{org}/{repo}").
		Subrouter().
		HandleFunc("/commit/{file:.*}", func(rw http.ResponseWriter, req *http.Request) {
			vars := mux.Vars(req)
			res, _ := json.Marshal(map[string]interface{}{
				"path":  vars,
				"query": req.URL.Query(),
			})
			rw.Write(res)
		}).
		Methods("GET")

	const url = "/api/org1/repo1/commit/src/test.js?message=lalal"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	const expected = `{"path":{"file":"src/test.js","org":"org1","repo":"repo1"},"query":{"message":["lalal"]}}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

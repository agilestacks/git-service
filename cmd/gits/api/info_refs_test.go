package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agilestacks/git-service/cmd/gits/config"
)

func init() {
	config.GitApiSecret = "secret1213"
	// Mock
	InfoPack = func(repoId, service string, out io.Writer) error {
		return nil
	}
}

func testBasicAuth(username, password string, t *testing.T) *httptest.ResponseRecorder {
	r := getRouter()

	const url = "/repo/X/Y/info/refs?service=git-upload-pack"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.SetBasicAuth(username, password)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	return rr
}

func TestBasicAuthUsernameValidates(t *testing.T) {
	rr := testBasicAuth("secret1213", "random", t)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "001e# service=git-upload-pack\n0000"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestBasicAuthPasswordValidates(t *testing.T) {
	rr := testBasicAuth("random", "secret1213", t)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "001e# service=git-upload-pack\n0000"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestBasicAuthFailsForBadCredentials(t *testing.T) {
	rr := testBasicAuth("random", "radnom", t)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

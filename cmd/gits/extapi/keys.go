package extapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/agilestacks/git-service/cmd/gits/config"
)

type HubUserKey struct {
	UserId    string `json:"userId"`
	PublicKey string `json:"publicKey"`
}

func UsersBySshKey(keyBase64 string, keyFingerprintSHA256 string) ([]string, error) {
	if config.NoExtApiCalls {
		return agileUsers, nil
	}

	hubUserKeys := fmt.Sprintf("%s/user/keys?fingerprint=%s", config.HubApiEndpoint, url.QueryEscape(keyFingerprintSHA256))
	req, err := http.NewRequest("GET", hubUserKeys, nil)
	if config.HubApiSecret != "" {
		req.Header.Add("X-API-Secret", config.HubApiSecret)
	}

	resp, err := hubApi.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error querying Hub user SSH keys: %v", err)
	}
	if config.Trace {
		log.Printf("%s %s: %s", req.Method, req.URL.String(), resp.Status)
	}
	if resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got %d HTTP querying Hub user SSH keys", resp.StatusCode)
	}
	var body bytes.Buffer
	read, err := body.ReadFrom(resp.Body)
	if read < 2 || err != nil {
		return nil, fmt.Errorf("Error reading Hub response (read %d bytes): %v", read, err)
	}
	var usersKeys []HubUserKey
	err = json.Unmarshal(body.Bytes(), &usersKeys)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling Hub user SSH keys response: %v", err)
	}
	users := make([]string, 0, len(usersKeys))
	for _, uk := range usersKeys {
		for _, maybeKey := range strings.Split(uk.PublicKey, " ") {
			if len(maybeKey) > 100 && keyBase64 == maybeKey {
				users = append(users, uk.UserId)
			}
		}
	}
	return users, nil
}

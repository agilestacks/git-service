package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"gits/config"
)

type UserKey struct {
	UserId    string
	PublicKey string
}

var agileUsers = []string{"anton", "arkadi", "igor", "igorlysak", "nikolay", "oleg", "rick"}

var hubApi = &http.Client{Timeout: 10 * time.Second}

func UsersBySshKey(keyBase64 string, keyFingerprintSHA256 string) ([]string, error) {
	if config.NoExtApiCalls {
		return agileUsers, nil
	}

	hubUserKeys := fmt.Sprintf("%s/api/v1/user/keys?fingerprint=%s", config.HubApiEndpoint, url.QueryEscape(keyFingerprintSHA256))
	req, err := http.NewRequest("GET", hubUserKeys, nil)
	if config.HubApiSecret != "" {
		req.Header.Add("X-API-Secret", config.HubApiSecret)
	}

	resp, err := hubApi.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error querying Hub user SSH keys: %v", err)
	}
	if resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got %d HTTP querying Hub user SSH keys", resp.StatusCode)
	}
	var body bytes.Buffer
	read, err := body.ReadFrom(resp.Body)
	if read < 4 || err != nil {
		return nil, fmt.Errorf("Error reading Hub response (read %d bytes): %v", read, err)
	}
	var usersKeys []UserKey
	err = json.Unmarshal(body.Bytes(), &usersKeys)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling Hub response: %v", err)
	}
	users := make([]string, 0, len(usersKeys))
	for _, uk := range usersKeys {
		if keyBase64 == uk.PublicKey {
			users = append(users, uk.UserId)
		}
	}
	return users, nil
}

package extapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"gits/config"
)

type AuthUserPass struct {
	Username string
	Password string
}

type AuthUser struct {
	Uid          string
	Organization string
	Groups       []string
}

func Login(username string, password string) (*AuthUser, error) {
	if config.NoExtApiCalls {
		for _, agileUser := range agileUsers {
			if username == agileUser {
				return &AuthUser{Uid: username, Organization: "ASI", Groups: []string{"ASI.Dev"}}, nil
			}
		}
		return nil, fmt.Errorf("No `%s` user found", username)
	}

	userPass := AuthUserPass{Username: username, Password: password}
	reqBody, err := json.Marshal(userPass)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling signin request: %v", err)
	}

	req, err := http.NewRequest("POST", signin, bytes.NewBuffer(reqBody))
	signin := fmt.Sprintf("%s/signin", config.AuthApiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("Error creating Auth Service request: %v", err)
	}
	if config.AuthApiSecret != "" {
		req.Header.Add("X-API-Secret", config.AuthApiSecret)
	}

	resp, err := authApi.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error during Auth Service signin: %v", err)
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("No `%s` user found", username)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got %d HTTP from Auth Service signin", resp.StatusCode)
	}
	var body bytes.Buffer
	read, err := body.ReadFrom(resp.Body)
	if read < 4 || err != nil {
		return nil, fmt.Errorf("Error reading Auth Service response (read %d bytes): %v", read, err)
	}
	var user AuthUser
	err = json.Unmarshal(body.Bytes(), &user)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling Auth Service signin response: %v", err)
	}
	return &user, nil
}

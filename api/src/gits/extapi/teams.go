package extapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"gits/config"
)

type AuthTeamMember struct {
	Id string
}

type AuthTeam struct {
	Members []AuthTeamMember
}

func UsersByTeam(teamId string) ([]string, error) {
	if config.NoExtApiCalls {
		if teamId == "1" {
			return agileUsers, nil
		}
		return nil, fmt.Errorf("No `%s` team found", teamId)
	}

	authTeams := fmt.Sprintf("%s/api/v1/teams/%s", config.AuthApiEndpoint, url.QueryEscape(teamId))
	req, err := http.NewRequest("GET", authTeams, nil)
	if config.AuthApiSecret != "" {
		req.Header.Add("X-API-Secret", config.AuthApiSecret)
	}

	resp, err := authApi.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error querying Auth Service team: %v", err)
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("No `%s` team found", teamId)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got %d HTTP querying Auth Service team", resp.StatusCode)
	}
	var body bytes.Buffer
	read, err := body.ReadFrom(resp.Body)
	if read < 4 || err != nil {
		return nil, fmt.Errorf("Error reading Auth Service response (read %d bytes): %v", read, err)
	}
	var team AuthTeam
	err = json.Unmarshal(body.Bytes(), &team)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling Auth Service team response: %v", err)
	}
	members := make([]string, 0, len(team.Members))
	for _, user := range team.Members {
		members = append(members, user.Id)
	}
	return members, nil
}

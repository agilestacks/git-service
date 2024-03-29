package extapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/agilestacks/git-service/cmd/gits/config"
)

type HubTeam struct {
	Id   string
	Name string
	Role string
}

type HubTemplate struct {
	OwnerId          string    `json:"ownerId"`
	TeamsPermissions []HubTeam `json:"teamsPermissions"`
}

type TeamAccess struct {
	TeamId   string
	TeamName string
	CanWrite bool
}

type Template struct {
	OwnerUserId string
	Teams       []TeamAccess
	// maybe we should have OwnerOrg here too
}

func TemplateById(templateId string) (*Template, error) {
	if config.NoExtApiCalls {
		return &Template{OwnerUserId: "arkadi", Teams: []TeamAccess{
			TeamAccess{TeamId: "1", CanWrite: true},
			TeamAccess{TeamId: "2" /*non-existing team*/, CanWrite: false}}}, nil
	}

	hubTemplates := fmt.Sprintf("%s/templates/%s", config.HubApiEndpoint, url.QueryEscape(templateId))
	req, err := http.NewRequest("GET", hubTemplates, nil)
	if config.HubApiSecret != "" {
		req.Header.Add("X-API-Secret", config.HubApiSecret)
	}

	resp, err := hubApi.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error querying Hub templates: %v", err)
	}
	if config.Trace {
		log.Printf("%s %s: %s", req.Method, req.URL.String(), resp.Status)
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("No `%s` template found", templateId)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got %d HTTP querying Hub templates", resp.StatusCode)
	}
	var body bytes.Buffer
	read, err := body.ReadFrom(resp.Body)
	if read < 2 || err != nil {
		return nil, fmt.Errorf("Error reading Hub response (read %d bytes): %v", read, err)
	}
	var template HubTemplate
	err = json.Unmarshal(body.Bytes(), &template)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling Hub templates response: %v", err)
	}
	teams := make([]TeamAccess, 0, len(template.TeamsPermissions))
	for _, team := range template.TeamsPermissions {
		canWrite := false
		if team.Role == "admin" || team.Role == "write" {
			canWrite = true
		}
		teams = append(teams, TeamAccess{TeamId: team.Id, TeamName: team.Name, CanWrite: canWrite})
	}
	return &Template{OwnerUserId: template.OwnerId, Teams: teams}, nil
}

package extapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"gits/config"
)

type HubTeam struct {
	Id   string
	Name string
	Role string
}

type HubTemplate struct {
	OwnerId string
	Teams   []HubTeam
}

type TeamAccess struct {
	TeamId   string
	CanWrite bool
}

type Template struct {
	OwnerUserId string
	Teams       []TeamAccess
}

func TemplateById(templateId string) (*Template, error) {
	if config.NoExtApiCalls {
		return &Template{OwnerUserId: "arkadi", Teams: []TeamAccess{
			TeamAccess{TeamId: "1", CanWrite: true},
			TeamAccess{TeamId: "2" /*non-existing team*/, CanWrite: false}}}, nil
	}

	hubTemplates := fmt.Sprintf("%s/api/v1/templates/%s", config.HubApiEndpoint, url.QueryEscape(templateId))
	req, err := http.NewRequest("GET", hubTemplates, nil)
	if config.HubApiSecret != "" {
		req.Header.Add("X-API-Secret", config.HubApiSecret)
	}

	resp, err := hubApi.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error querying Hub user SSH keys: %v", err)
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("No `%s` template found", templateId)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got %d HTTP querying Hub templates", resp.StatusCode)
	}
	var body bytes.Buffer
	read, err := body.ReadFrom(resp.Body)
	if read < 4 || err != nil {
		return nil, fmt.Errorf("Error reading Hub response (read %d bytes): %v", read, err)
	}
	var template HubTemplate
	err = json.Unmarshal(body.Bytes(), &template)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling Hub templates response: %v", err)
	}
	teams := make([]TeamAccess, 0, len(template.Teams))
	for _, team := range template.Teams {
		canWrite := false
		if team.Role == "ADMIN" || team.Role == "RW" {
			canWrite = true
		}
		teams = append(teams, TeamAccess{TeamId: team.Id, CanWrite: canWrite})
	}
	return &Template{OwnerUserId: template.OwnerId, Teams: teams}, nil
}

package repo

import (
	"fmt"
	"strings"

	"gits/api"
)

func access(repo string, users []string) (bool, error) {
	dash := strings.LastIndex(repo, "-")
	if dash <= 1 || dash >= len(repo)-1 {
		return false, fmt.Errorf("Unable to determine stack template id from repo name `%s`", repo)
	}
	templateId := repo[dash+1:]

	template, err := api.TemplateById(templateId)
	if err != nil {
		return false, fmt.Errorf("Unable to fetch template `%s` info: %v", templateId, err)
	}

	granted := make([]string, 0, 1)
	granted = append(granted, template.OwnerUserId)

	var teamErr error
	for _, teamId := range template.Teams {
		team, err := api.UsersByTeam(teamId)
		if err != nil {
			if teamErr == nil {
				teamErr = err
			}
			continue
		}
		granted = append(granted, team...)
	}

	for _, user := range users {
		for _, teamMember := range granted {
			if user == teamMember {
				return true, teamErr
			}
		}
	}

	return false, teamErr
}

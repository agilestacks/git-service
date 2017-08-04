package repo

import (
	"fmt"
	"strings"

	"gits/api"
)

type UserAccess struct {
	UserId   string
	CanWrite bool
}

func access(repo string, verb string, users []string) (bool, error) {
	dash := strings.LastIndex(repo, "-")
	if dash <= 1 || dash >= len(repo)-1 {
		return false, fmt.Errorf("Unable to determine stack template id from repo name `%s`", repo)
	}
	templateId := repo[dash+1:]

	template, err := api.TemplateById(templateId)
	if err != nil {
		return false, fmt.Errorf("Unable to fetch template `%s` info: %v", templateId, err)
	}

	granted := make([]UserAccess, 0, 1)
	granted = append(granted, UserAccess{UserId: template.OwnerUserId, CanWrite: true})

	var teamErr error
	for _, team := range template.Teams {
		teamUsers, err := api.UsersByTeam(team.TeamId)
		if err != nil {
			if teamErr == nil {
				teamErr = err
			}
			continue
		} else {
			for _, userId := range teamUsers {
				granted = append(granted, UserAccess{UserId: userId, CanWrite: team.CanWrite})
			}
		}
	}

	writeRequested := verb == "git-receive-pack"

	for _, userId := range users {
		for _, grantedTo := range granted {
			if userId == grantedTo.UserId && (!writeRequested || grantedTo.CanWrite) {
				return true, teamErr
			}
		}
	}

	return false, teamErr
}

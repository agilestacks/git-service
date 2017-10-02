package repo

import (
	"fmt"
	"strings"

	"gits/extapi"
)

type UserAccess struct {
	UserId   string
	CanWrite bool
}

func templateId(repo string) (string, error) {
	dash := strings.LastIndex(repo, "-")
	if dash <= 1 || dash >= len(repo)-1 {
		return "", fmt.Errorf("Unable to determine stack template id from repo name `%s`", repo)
	}
	return repo[dash+1:], nil
}

func Access(repo string, verb string, users []string) (bool, error) {
	templateId, err := templateId(repo)
	if err != nil {
		return false, err
	}

	template, err := extapi.TemplateById(templateId)
	if err != nil {
		return false, fmt.Errorf("Unable to fetch template `%s` info: %v", templateId, err)
	}

	granted := make([]UserAccess, 0, 1)
	granted = append(granted, UserAccess{UserId: template.OwnerUserId, CanWrite: true})

	var teamErr error
	for _, team := range template.Teams {
		teamUsers, err := extapi.UsersByTeam(team.TeamId)
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

func AccessWithLogin(org, repo, verb, username, password string) (bool, error) {
	templateId, err := templateId(repo)
	if err != nil {
		return false, err
	}

	user, err := extapi.Login(username, password)
	if err != nil {
		return false, fmt.Errorf("Unable to signin user `%s`: %v", username, err)
	}

	if strings.ToLower(org) != strings.ToLower(user.Organization) {
		return false, fmt.Errorf("User org `%s` does not match repo org `%s`", user.Organization, org)
	}

	template, err := extapi.TemplateById(templateId)
	if err != nil {
		return false, fmt.Errorf("Unable to fetch template `%s` info: %v", templateId, err)
	}

	if user.Uid == template.OwnerUserId {
		return true, nil
	}

	writeRequested := verb == "git-receive-pack"

	for _, userTeam := range user.Groups {
		for _, templateTeam := range template.Teams {
			if userTeam == templateTeam.TeamName && (!writeRequested || templateTeam.CanWrite) {
				return true, nil
			}
		}
	}

	return false, nil
}

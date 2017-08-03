package api

import (
	"fmt"
)

func UsersByTeam(teamId string) ([]string, error) {
	if teamId == "1" {
		return []string{"arkadi", "anton", "igor", "igorlysak", "nikolay", "oleg", "rick"}, nil
	}
	return nil, fmt.Errorf("No `%s` team found", teamId)
}

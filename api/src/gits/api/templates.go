package api

import (
	"fmt"
	"strings"
)

type Template struct {
	OwnerUserId string
	Teams       []string
}

func TemplateById(templateId string) (*Template, error) {
	if strings.HasPrefix(templateId, "1") {
		return &Template{OwnerUserId: "arkadi", Teams: []string{"1", "2" /*non-existing team*/}}, nil
	}
	return nil, fmt.Errorf("No `%s` template found", templateId)
}

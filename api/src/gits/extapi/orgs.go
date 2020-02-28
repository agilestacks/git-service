package extapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"gits/config"
)

type Org struct {
	Id         string
	ShowSource bool
}

func OrgById(orgId string) (*Org, error) {
	if config.NoExtApiCalls {
		return &Org{Id: "ASI", ShowSource: true}, nil
	}

	orgId = strings.ToUpper(orgId)

	orgs := fmt.Sprintf("%s/organizations/%s", config.SubsApiEndpoint, url.QueryEscape(orgId))
	req, err := http.NewRequest("GET", orgs, nil)
	if config.HubApiSecret != "" {
		req.Header.Add("X-API-Secret", config.SubsApiSecret)
	}

	resp, err := subsApi.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error querying Hub organizations: %v", err)
	}
	if config.Trace {
		log.Printf("%s %s: %s", req.Method, req.URL.String(), resp.Status)
	}
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("No `%s` organization found", orgId)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Got %d HTTP querying Hub organizations", resp.StatusCode)
	}
	var body bytes.Buffer
	read, err := body.ReadFrom(resp.Body)
	if read < 2 || err != nil {
		return nil, fmt.Errorf("Error reading Hub response (read %d bytes): %v", read, err)
	}
	var org Org
	err = json.Unmarshal(body.Bytes(), &org)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling Hub organizations response: %v", err)
	}
	return &org, nil
}

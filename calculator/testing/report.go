package testing_demo

import "encoding/json"

type UserReport struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

func GenerateReportJSON(id int, name string, roles []string) ([]byte, error) {
	report := UserReport{
		ID:    id,
		Name:  name,
		Roles: roles,
	}
	return json.MarshalIndent(report, "", "  ")
}

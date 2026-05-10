package structscomposition1

import "testing"

type ParsedUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestParseStruct(t *testing.T) {

	jsonData := []byte(`{"id": "u-100", "name": "Khizer"}`)
	var user ParsedUser

	result := ParseStruct(jsonData, &user)
	parsedUser, ok := result.(*ParsedUser)
	if !ok {
		t.Fatal("expected result to be of type *ParsedUser")
	}

	if parsedUser.ID != "u-100" {
		t.Fatalf("expected ID u-100, got %s", parsedUser.ID)
	}
	if parsedUser.Name != "Khizer" {
		t.Fatalf("expected Name Khizer, got %s", parsedUser.Name)
	}
}

func TestParseStructMalformedJSON(t *testing.T) {
	badJSON := []byte(`{"id": "u-100", "name": "Khizer"`) // truncated

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for malformed JSON, got none")
		}
	}()

	var user ParsedUser
	ParseStruct(badJSON, &user)
}

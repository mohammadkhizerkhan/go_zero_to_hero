package testing_demo

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// Declare command-line flag for updating golden files: go test -update
var update = flag.Bool("update", false, "update golden files")

// go test -v -run TestGenerateReportJSON_Golden ./testing -update
func TestGenerateReportJSON_Golden(t *testing.T) {
	actual, err := GenerateReportJSON(101, "Alice Smith", []string{"Admin", "Developer"})
	if err != nil {
		t.Fatalf("Failed to generate JSON report: %v", err)
	}

	goldenPath := filepath.Join("testdata", "user_report.json.golden")

	// If -update flag is passed, write output to golden file and finish
	if *update {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("Failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, actual, 0644); err != nil {
			t.Fatalf("Failed to update golden file: %v", err)
		}
		t.Log("Updated golden file successfully!")
		return
	}

	// Read saved golden file
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file %s: %v (run with -update to generate)", goldenPath, err)
	}

	// Compare generated output with golden file
	if !bytes.Equal(actual, expected) {
		t.Errorf("Generated JSON does not match golden file!\nGot:\n%s\n\nWant:\n%s", actual, expected)
	}
}

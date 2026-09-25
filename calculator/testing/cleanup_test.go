package testing_demo

import (
	"os"
	"testing"
)

// Helper setup function
func createTempConfig(t *testing.T, content string) string {
	t.Helper() // Marks function as a test helper so file line numbers reported on failure point to caller

	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Register teardown action. Runs automatically when the calling test completes.
	t.Cleanup(func() {
		t.Errorf("removing the file from %s",tmpFile.Name())
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}


// go test -v -run TestConfigLoader ./testing/...
func TestConfigLoader(t *testing.T) {
	filePath := createTempConfig(t, `{"env": "test"}`)

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("Expected temp file %s to exist", filePath)
	}

	// Read and test logic here...
	data, err := os.ReadFile(filePath)
	if err != nil || string(data) != `{"env": "test"}` {
		t.Errorf("Unexpected file content: %s", string(data))
	}

	// After this test function returns, t.Cleanup automatically removes filePath!
}

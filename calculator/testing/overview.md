# Topic J: Unit Testing & Mocking

Writing production-grade Go requires testing that is maintainable, deterministic, and isolated from external dependencies. The standard `testing` package provides minimal primitives (`*testing.T`) rather than large assertion libraries, encouraging clean Go code patterns.

---

## J.1 Theory & Mental Model

- **Table-Driven Tests**: Instead of duplicating test code for every scenario, you define a slice of test cases (input vs. expected output) and iterate over them in a single loop.
- **Subtests (`t.Run`)**: Creates isolated test execution contexts with unique names. Subtests allow running specific test cases directly from the CLI (e.g., `go test -run TestIsPalindrome/Empty_String`).
- **Test Isolation (`t.Cleanup`)**: Registers teardown functions that run automatically when a test or subtest completes—even if the test panics or fails early.
- **Interface Mocking**: Relying on Go's implicit interfaces to inject fake implementations into unit tests without calling real network endpoints or databases.
- **Golden Files**: Storing large, complex test expectations (JSON responses, rendered templates) in external `.golden` files under a `testdata/` directory rather than hardcoding long strings in `.go` files.

---

## Assignments & Implementation Guide

### J.1 & J.2 The Table Test & The Subtest

**Goal**: Implement `IsPalindrome(s string) bool` and test it using a table of struct test cases executed via `t.Run`.

#### Code Implementation

##### `palindrome.go`

```go
package testing_demo

import "unicode"

// IsPalindrome checks if a string reads the same forwards and backwards,
// ignoring spaces and case.
func IsPalindrome(s string) bool {
	var runes []rune
	for _, r := range []rune(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			runes = append(runes, unicode.ToLower(r))
		}
	}

	for i := 0; i < len(runes)/2; i++ {
		if runes[i] != runes[len(runes)-1-i] {
			return false
		}
	}
	return true
}
```

##### `palindrome_test.go`

```go
package testing_demo

import "testing"

func TestIsPalindrome(t *testing.T) {
	// Table of test cases
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "Simple palindrome", input: "racecar", expected: true},
		{name: "Mixed case and spaces", input: "A man a plan a canal Panama", expected: true},
		{name: "Non-palindrome", input: "golang", expected: false},
		{name: "Empty string", input: "", expected: true},
		{name: "Single character", input: "a", expected: true},
	}

	for _, tt := range tests {
		// Subtest runner: t.Run isolates each table row
		t.Run(tt.name, func(t *testing.T) {
			actual := IsPalindrome(tt.input)
			if actual != tt.expected {
				t.Errorf("IsPalindrome(%q) = %v; want %v", tt.input, actual, tt.expected)
			}
		})
	}
}
```

---

### J.3 The Cleanup (`t.Cleanup`)

**Goal**: Create a temporary file during a test and register a cleanup function using `t.Cleanup()` to guarantee file deletion after execution.

> [!TIP]
> **Why `t.Cleanup` over `defer`?**  
> `defer` runs when the surrounding test function exits. `t.Cleanup` runs when the test (and all its subtests) finish executing, making it safer for parallel tests (`t.Parallel()`) and modular helper setup functions.

```go
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
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}

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
```

---

### J.4 The Interface Mock

**Goal**: Define a `WeatherAPI` interface and inject a mock instance into a `WeatherAdvisor` service to test recommendation logic without making network calls.

#### Code Implementation

##### `weather.go`

```go
package testing_demo

import "fmt"

// Interface contract
type WeatherAPI interface {
	GetTemperature(city string) (float64, error)
}

// Service depending on the interface
type WeatherAdvisor struct {
	api WeatherAPI
}

func NewWeatherAdvisor(api WeatherAPI) *WeatherAdvisor {
	return &WeatherAdvisor{api: api}
}

func (w *WeatherAdvisor) RecommendOutfit(city string) (string, error) {
	temp, err := w.api.GetTemperature(city)
	if err != nil {
		return "", fmt.Errorf("could not fetch weather: %w", err)
	}

	if temp < 15.0 {
		return "Wear a heavy jacket", nil
	}
	return "Wear a t-shirt", nil
}
```

##### `weather_test.go`

```go
package testing_demo

import (
	"errors"
	"testing"
)

// Mock implementation satisfying WeatherAPI
type MockWeatherAPI struct {
	GetTemperatureFunc func(city string) (float64, error)
}

func (m *MockWeatherAPI) GetTemperature(city string) (float64, error) {
	return m.GetTemperatureFunc(city)
}

func TestRecommendOutfit(t *testing.T) {
	t.Run("Cold weather recommendation", func(t *testing.T) {
		mockAPI := &MockWeatherAPI{
			GetTemperatureFunc: func(city string) (float64, error) {
				return 10.0, nil // Return 10°C fake temperature
			},
		}

		advisor := NewWeatherAdvisor(mockAPI)
		recommendation, err := advisor.RecommendOutfit("London")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if recommendation != "Wear a heavy jacket" {
			t.Errorf("Expected 'Wear a heavy jacket', got %q", recommendation)
		}
	})

	t.Run("API failure handling", func(t *testing.T) {
		mockAPI := &MockWeatherAPI{
			GetTemperatureFunc: func(city string) (float64, error) {
				return 0, errors.New("network timeout")
			},
		}

		advisor := NewWeatherAdvisor(mockAPI)
		_, err := advisor.RecommendOutfit("London")

		if err == nil {
			t.Fatal("Expected error when API fails, got nil")
		}
	})
}
```

---

### J.5 The Golden File

**Goal**: Test a function that outputs a complex JSON string by comparing it against a saved `.golden` file on disk. Add `-update` flag support to generate or update the golden file automatically.

#### Code Implementation

##### `report.go`

```go
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
```

##### `report_test.go`

```go
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
```

---

## Running Your Tests

Run all unit tests with detailed subtest logging:

```bash
go test -v ./...
```

Generate or update golden files:

```bash
go test -v -update
```

Run a specific subtest by path:

```bash
go test -v -run TestIsPalindrome/Mixed_case_and_spaces
```

---

### J.6 Automated Mock Generation & Dependency Injection (GoMock / mockgen)

**Goal**: Automatically generate mock types from interface definitions using `mockgen` and `//go:generate`, and inject them into service instances to test complex scenarios.

#### Why GoMock?
- **Compile-Time Safety**: Auto-generated mocks match interface changes without manually updating struct mocks.
- **Rich Matchers & Assertions**: Assert exact method call counts (`.Times(1)`), argument matching (`gomock.Any()`, `gomock.Eq()`), and return sequences.

#### Code Generation Directive (`weather.go`)

```go
//go:generate mockgen -source=weather.go -destination=mocks/mock_weather.go -package=mocks
type WeatherAPI interface {
	GetTemperature(city string) (float64, error)
}
```

Run code generator:
```bash
go generate ./...
```

#### GoMock Test Suite (`weather_test.go`)

```go
func TestRecommendOutfit_GoMock(t *testing.T) {
	t.Run("Cold weather recommendation with GoMock", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockAPI := mocks.NewMockWeatherAPI(ctrl)
		mockAPI.EXPECT().
			GetTemperature("London").
			Return(10.0, nil).
			Times(1)

		// Inject mock into service via constructor
		advisor := NewWeatherAdvisor(mockAPI)
		recommendation, err := advisor.RecommendOutfit("London")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if recommendation != "Wear a heavy jacket" {
			t.Errorf("Expected 'Wear a heavy jacket', got %q", recommendation)
		}
	})
}
```

Run GoMock tests:
```bash
go test -v -run TestRecommendOutfit_GoMock ./testing/...
```
package testing_demo

import "testing"

//  go test -v run TestIsPalindrome ./testing/...
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
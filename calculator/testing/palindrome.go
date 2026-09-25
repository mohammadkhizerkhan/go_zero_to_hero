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
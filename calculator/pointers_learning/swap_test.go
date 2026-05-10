package pointerslearning

import (
	"testing"
)

func TestSwap(t *testing.T) {
	a := 5
	b := 10

	Swap(&a, &b)

	if a != 10 || b != 5 {
		t.Errorf("Expected a to be 10 and b to be 5, but got a: %d, b: %d", a, b)
	}

	if a != 5 || b != 10 {
		t.Logf("a and b are swapped. a: %d, b: %d", a, b)
	}
}

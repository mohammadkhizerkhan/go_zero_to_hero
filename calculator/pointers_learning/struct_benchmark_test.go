package pointerslearning

import "testing"

func BenchmarkProcessByValue(b *testing.B) {
	h := HeavyStruct{}
	for i := 0; i < 10000; i++ {
		h.field[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ProcessByValue(h)
	}
}

func BenchmarkProcessByPointer(b *testing.B) {
	h := HeavyStruct{}
	for i := 0; i < 10000; i++ {
		h.field[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CallProcessByPointer(&h)
	}
}

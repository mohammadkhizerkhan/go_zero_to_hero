package pointerslearning

import "testing"

func TestHeapEscapeValue(t *testing.T) {
	p := HeapEscape()
	if *p != 42 {
		t.Fatalf("expected 42, got %d", *p)
	}
}

func BenchmarkStackOnly(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SinkInt = StackOnly()
	}
}

func BenchmarkHeapEscape(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SinkPtr = HeapEscape()
	}
}

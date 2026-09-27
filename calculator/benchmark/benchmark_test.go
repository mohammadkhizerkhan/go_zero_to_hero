package benchmark_demo

import (
	"bytes"
	"strings"
	"testing"
)

    var words = []string{"go", "zero", "to", "hero", "performance", "benchmarking", "tutorial"}
    var globalSink string

    // go test -bench=. -benchmem ./benchmark
    func BenchmarkConcatPlus(b *testing.B) {
        var s string
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            s = ""                                                                                                      
            for _, w := range words {                                                                                   
                s += w                                                                                                  
            }                                                                                                           
        }
        globalSink = s
    }

    func BenchmarkConcatBuffer(b *testing.B) {
        var s string
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            var buf bytes.Buffer
            for _, w := range words {
                buf.WriteString(w)
            }
            s = buf.String()
        }
        globalSink = s
    }

    func BenchmarkConcatBuilder(b *testing.B) {
        var s string
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            var sb strings.Builder
            sb.Grow(64)
            for _, w := range words {
                sb.WriteString(w)
            }
            s = sb.String()
        }
        globalSink = s
    }
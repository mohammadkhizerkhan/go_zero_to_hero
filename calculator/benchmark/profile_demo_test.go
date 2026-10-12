package benchmark_demo

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// ============================================================================
// Assignment 3: CPU Profiling Demo
// ============================================================================

// slowCPUFunction simulates CPU-intensive computation (nested cryptographic hashing).
func slowCPUFunction(iterations int) [32]byte {
	var hash [32]byte
	data := []byte("go_zero_to_hero_benchmark_seed_data")
	for i := 0; i < iterations; i++ {
		h := sha256.New()
		h.Write(data)
		h.Write(hash[:])
		copy(hash[:], h.Sum(nil))
	}
	return hash
}

var globalHashSink [32]byte

// BenchmarkCPUSlowFunction demonstrates CPU profiling.
// Run: go test -bench=BenchmarkCPUSlowFunction -cpuprofile=cpu.pprof ./benchmark
// View: go tool pprof -http=:8080 cpu.pprof
func BenchmarkCPUSlowFunction(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalHashSink = slowCPUFunction(100)
	}
}

// ============================================================================
// Assignment 4: Memory Profiling Demo (Garbage Generation vs Optimized)
// ============================================================================

// generateGarbage creates excessive short-lived heap allocations inside a loop
// through string formatting, heap escaping slices, and intermediate string conversions.
func generateGarbage(n int) string {
	var result string
	for i := 0; i < n; i++ {
		// fmt.Sprintf and string concatenation create multiple allocations per iteration
		temp := fmt.Sprintf("item-%d-val:%s", i, strconv.Itoa(i*2))
		result += temp + ","
	}
	return result
}

// generateOptimized minimizes garbage by pre-allocating strings.Builder buffer
// and avoiding fmt.Sprintf reflection overhead.
func generateOptimized(n int) string {
	var sb strings.Builder
	// Pre-allocate estimated buffer capacity: ~20 bytes per item
	sb.Grow(n * 20)
	for i := 0; i < n; i++ {
		sb.WriteString("item-")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString("-val:")
		sb.WriteString(strconv.Itoa(i * 2))
		sb.WriteString(",")
	}
	return sb.String()
}

var globalMemSink string

// BenchmarkMemoryGarbage demonstrates high heap allocation / GC garbage pressure.
// Run: go test -bench=BenchmarkMemoryGarbage -benchmem -memprofile=mem.pprof ./benchmark
// Analyze: go tool pprof -alloc_space mem.pprof
func BenchmarkMemoryGarbage(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalMemSink = generateGarbage(100)
	}
}

// BenchmarkMemoryOptimized demonstrates low-allocation / memory-efficient code.
func BenchmarkMemoryOptimized(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalMemSink = generateOptimized(100)
	}
}

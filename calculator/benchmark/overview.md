# Topic K: Benchmarking & Performance Profiling

Writing high-performance Go applications requires measuring latency, throughput, and memory allocations with statistical precision. Go provides first-class benchmarking primitives built directly into the standard `testing` package via `*testing.B`.

---

## 🧭 K.1 Theory & Mental Model

### 1. How `testing.B` Works

Unlike standard tests (`*testing.T`), a benchmark function is	 executed multiple times with dynamically increasing values of `b.N` until the benchmark runner achieves a statistically stable duration (default target: 1 second per benchmark).

```text
Iter 1: b.N = 1       (Duration: 120ns  -> Too short, scale up)
Iter 2: b.N = 100     (Duration: 11μs   -> Scale up)
Iter 3: b.N = 10000   (Duration: 1.1ms  -> Scale up)
Iter 4: b.N = 1000000 (Duration: 0.11s  -> Statistically significant!)
```

### 2. Anatomy of Benchmark Metrics

Running `go test -bench=. -benchmem` produces standard output metrics:

```text
BenchmarkConcatBuilder-8    15234120    78.4 ns/op    32 B/op    1 allocs/op
│                           │           │             │          │
│                           │           │             │          └── Allocations per operation
│                           │           │             └───────────── Bytes allocated per operation
│                           │           └─────────────────────────── Nanoseconds per operation (speed)
│                           └─────────────────────────────────────── Total iterations run (b.N)
└─────────────────────────────────────────────────────────────────── Benchmark Name - GOMAXPROCS
```

### 3. Golden Rules of Go Benchmarking

- **Always loop up to `b.N`**: `for i := 0; i < b.N; i++` must wrap the code under test. Never use a fixed loop count inside the benchmark loop.
- **Reset Expensive Setup Time (`b.ResetTimer()`)**: If setting up test data or fixtures takes time, call `b.ResetTimer()` right before the `b.N` loop to exclude setup overhead.
- **Prevent Compiler Dead-Code Elimination (The Sink Pattern)**: If the Go compiler determines that a computed result inside a loop is unused, it may optimize away the entire loop. Assign outputs to a package-level global sink variable (`var globalSink string`).
- **Always Track Allocations**: Use `b.ReportAllocs()` inside the benchmark or pass the `-benchmem` flag via the CLI to measure GC pressure.

---

## 🛠️ Assignments & Implementation Guide

### K.1 String Concatenation: `+` vs `strings.Builder` vs `bytes.Buffer`

**Goal**: Benchmark the performance difference between naive string concatenation (`+`), `bytes.Buffer`, and pre-allocated `strings.Builder`.

```go
package benchmark_demo

import (
	"bytes"
	"strings"
	"testing"
)

var words = []string{"go", "zero", "to", "hero", "performance", "benchmarking", "tutorial"}

// Global sink prevents compiler dead-code elimination
var globalStringSink string

// 1. Naive string concatenation (Creates new allocations on every +=)
func BenchmarkConcatPlus(b *testing.B) {
	var s string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s = ""
		for _, w := range words {
			s += w
		}
	}
	globalStringSink = s
}

// 2. bytes.Buffer
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
	globalStringSink = s
}

// 3. strings.Builder with Grow (Zero extra heap allocations)
func BenchmarkConcatBuilder(b *testing.B) {
	var s string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var sb strings.Builder
		sb.Grow(64) // Pre-allocate buffer capacity
		for _, w := range words {
			sb.WriteString(w)
		}
		s = sb.String()
	}
	globalStringSink = s
}
```

---

### K.2 Slice Pre-Allocation: Dynamic `append` vs `make([]T, 0, cap)`

**Goal**: Demonstrate how avoiding slice reallocations and memory copies improves throughput.

```go
package benchmark_demo

import "testing"

const sliceSize = 10000

var globalSliceSink []int

// 1. Dynamic slice growth (triggers repeated doubling & GC copies)
func BenchmarkSliceDynamic(b *testing.B) {
	b.ReportAllocs()
	var res []int
	for i := 0; i < b.N; i++ {
		res = []int{}
		for j := 0; j < sliceSize; j++ {
			res = append(res, j)
		}
	}
	globalSliceSink = res
}

// 2. Pre-allocated capacity (1 heap allocation total)
func BenchmarkSlicePreAllocated(b *testing.B) {
	b.ReportAllocs()
	var res []int
	for i := 0; i < b.N; i++ {
		res = make([]int, 0, sliceSize)
		for j := 0; j < sliceSize; j++ {
			res = append(res, j)
		}
	}
	globalSliceSink = res
}
```

---

### K.3 Table-Driven Parameterized Benchmarks (`b.Run`)

**Goal**: Run sub-benchmarks across varying input sizes (N = 10, 100, 1000, 10000) to evaluate algorithmic scaling.

```go
package benchmark_demo

import (
	"fmt"
	"testing"
)

func linearSearch(data []int, target int) bool {
	for _, v := range data {
		if v == target {
			return true
		}
	}
	return false
}

func BenchmarkLinearSearch_Table(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			data := make([]int, size)
			for i := 0; i < size; i++ {
				data[i] = i
			}
			target := size - 1 // Worst-case search at end of slice

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = linearSearch(data, target)
			}
		})
	}
}
```

---

### K.4 Parallel Benchmarking (`b.RunParallel`)

**Goal**: Measure throughput of concurrent access under multiple CPU cores using `b.RunParallel` and `pb.Next()`.

```go
package benchmark_demo

import (
	"sync"
	"testing"
)

type SafeCounter struct {
	mu  sync.RWMutex
	val int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	c.val++
	c.mu.Unlock()
}

func (c *SafeCounter) Read() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.val
}

func BenchmarkParallelCounter(b *testing.B) {
	counter := &SafeCounter{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter.Inc()
		}
	})
}
```

---

## ⚡ CLI Commands Cheatsheet

### 1. Run All Benchmarks

```bash
go test -bench=. ./benchmark/...
```

### 2. Include Memory Allocation Statistics

```bash
go test -bench=. -benchmem ./benchmark/...
```

### 3. Run a Specific Benchmark by Regex Pattern

```bash
go test -bench=BenchmarkConcatBuilder -benchmem ./benchmark/...
```

### 4. Increase Benchmark Time & Iteration Count

Increase accuracy by running each benchmark for 3 seconds across 5 rounds:

```bash
go test -bench=. -benchtime=3s -count=5 ./benchmark/...
```

### 5. Generate CPU and Memory Profiles

Generate pprof profiling artifacts during benchmark runs:

```bash
go test -bench=. -cpuprofile=cpu.pprof -memprofile=mem.pprof ./benchmark/...
```

Analyze with Go pprof tool:

```bash
go tool pprof -http=:8080 cpu.pprof
```

---

## 📊 Summary: Optimization Matrix

| Optimization Technique | Typical Speedup | Memory Benefit | Best Used For |
| :--- | :--- | :--- | :--- |
| **`strings.Builder` with `Grow(n)`** | ~5x – 10x faster | 0–1 allocations vs O(N) allocations | Repeated string building & formatting |
| **Slice Pre-allocation (`make([]T, 0, n)`)** | ~3x – 6x faster | Eliminates slice capacity doublings | Slices with known/predictable bounds |
| **Passing Pointers for Large Structs (>128 B)** | ~2x – 4x faster | Eliminates stack/heap copy overhead | Heavy config/data payload structs |
| **`sync/atomic` vs `sync.Mutex`** | ~3x – 5x faster | Lock-free CPU instruction level synchronization | Numeric counters, state flags, atomic pointers |
| **Object Pooling with `sync.Pool`** | ~2x – 4x faster | Drastically reduces GC overhead | High-frequency short-lived byte buffers / objects |
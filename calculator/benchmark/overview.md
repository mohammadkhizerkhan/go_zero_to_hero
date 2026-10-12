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

### K.5 CPU Profiling: Profiling Slow Functions with `pprof`

**Goal**: Pinpoint CPU bottlenecks and hot functions using Go's sampling CPU profiler and visualize them with `go tool pprof`.

#### How It Works:
The Go CPU profiler interrupts program execution 100 times per second (100 Hz). On every interrupt, it records the call stack of the currently running goroutine. Functions that appear most frequently across samples consume the most CPU cycles.

```go
// Code reference: benchmark/profile_demo_test.go
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

func BenchmarkCPUSlowFunction(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		globalHashSink = slowCPUFunction(100)
	}
}
```

#### Step-by-Step Usage:

1. **Generate the CPU Profile**:
   ```bash
   go test -run=^$ -bench=BenchmarkCPUSlowFunction -cpuprofile=cpu.pprof ./benchmark
   ```

2. **Interactive CLI Analysis**:
   ```bash
   go tool pprof cpu.pprof
   ```
   Key interactive commands inside `pprof`:
   - `top10`: Lists the top 10 functions by CPU time consumption.
   - `top -cum`: Lists functions sorted by cumulative time (including functions they call).
   - `list slowCPUFunction`: Shows line-by-line CPU time breakdown inside `slowCPUFunction`.
   - `web`: Generates an SVG call graph and opens it in your browser (requires `graphviz`).
   - `exit` or `quit`: Exits the interactive pprof shell.

3. **Visual Browser & Flame Graph Interface**:
   ```bash
   go tool pprof -http=:8080 cpu.pprof
   ```
   Open `http://localhost:8080` in your browser. Use the **VIEW** dropdown menu to switch between:
   - **Top**: Tabular overview of hot functions.
   - **Graph**: Directed acyclic call graph with edge weights representing execution time.
   - **Flame Graph**: Interactive hierarchical visualization where bar width represents execution time.
   - **Source**: Annotated Go source code showing line-by-line time expenditure.

---

### K.6 Memory Profiling: Identifying Garbage & Heap Allocations

**Goal**: Trace where short-lived temporary objects (garbage) and high heap allocations occur to eliminate GC pauses.

#### How It Works:
Go's memory profiler records heap allocations by sampling 1 in every 512 KB allocated. It lets you inspect memory in four distinct modes:
- **`alloc_space`**: Total number of bytes allocated (including garbage collected memory). Ideal for reducing GC pressure.
- **`alloc_objects`**: Total count of allocated heap objects.
- **`inuse_space`**: Number of bytes currently alive on the heap. Ideal for tracking memory leaks.
- **`inuse_objects`**: Number of objects currently alive on the heap.

```go
// Code reference: benchmark/profile_demo_test.go
// 1. Garbage generator (350 allocs/op, 82 KB/op)
func generateGarbage(n int) string {
	var result string
	for i := 0; i < n; i++ {
		temp := fmt.Sprintf("item-%d-val:%s", i, strconv.Itoa(i*2))
		result += temp + ","
	}
	return result
}

// 2. Low-allocation optimized (51 allocs/op, 2.2 KB/op)
func generateOptimized(n int) string {
	var sb strings.Builder
	sb.Grow(n * 20) // Pre-allocate buffer
	for i := 0; i < n; i++ {
		sb.WriteString("item-")
		sb.WriteString(strconv.Itoa(i))
		sb.WriteString("-val:")
		sb.WriteString(strconv.Itoa(i * 2))
		sb.WriteString(",")
	}
	return sb.String()
}
```

#### Step-by-Step Usage:

1. **Generate the Memory Profile**:
   ```bash
   go test -run=^$ -bench=BenchmarkMemoryGarbage -benchmem -memprofile=mem.pprof ./benchmark
   ```

2. **Inspect Cumulative Allocated Space (Garbage)**:
   ```bash
   go tool pprof -alloc_space mem.pprof
   ```
   Inside pprof:
   ```text
   (pprof) top
   (pprof) list generateGarbage
   ```
   The `list` command reveals the exact lines responsible for heap escaping:
   ```text
   ROUTINE ======================== generateGarbage in .../profile_demo_test.go
     80.75MB    80.75MB (flat, cum)   100% of Total
           .          .     43: func generateGarbage(n int) string {
           .          .     44: 	var result string
           .          .     45: 	for i := 0; i < n; i++ {
     38.25MB    38.25MB     46: 		temp := fmt.Sprintf("item-%d-val:%s", i, strconv.Itoa(i*2))
     42.50MB    42.50MB     47: 		result += temp + ","
           .          .     48: 	}
           .          .     49: 	return result
   ```

3. **Launch Web UI**:
   ```bash
   go tool pprof -http=:8080 -alloc_space mem.pprof
   ```

---

#### explaination
 ### 1. The Header: Function Summary

    ROUTINE ======================== github.com/.../benchmark.generateGarbage in .../profile_demo_test.go
        2.75GB     2.81GB (flat, cum) 99.91% of Total

   Metric              │ Value              │ Meaning
  ─────────────────────┼────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────
   flat                │ 2.75GB             │ Memory allocated directly by this function itself (not counting calls to external functions).
   cum (cumulative)    │ 2.81GB             │ Total memory allocated by this function PLUS all functions called by it (e.g., fmt.Sprintf, strconv.Itoa).
   % of Total          │ 99.91%             │ This single function is responsible for virtually all allocations during the entire benchmark run.
  ──────
  ### 2. Line-by-Line Breakdown

             .          .     46:func generateGarbage(n int) string {
             .          .     47:   var result string
             .          .     48:   for i := 0; i < n; i++ {
             .          .     49:           // fmt.Sprintf and string concatenation create multiple allocations per iteration
       49.50MB   110.50MB     50:           temp := fmt.Sprintf("item-%d-val:%s", i, strconv.Itoa(i*2))
        2.70GB     2.70GB     51:           result += temp + ","
             .          .     52:   }
             .          .     53:   return result
             .          .     54:}

  The two columns on the left correspond to:

  • Column 1 (flat): Memory allocated on this exact line.
  • Column 2 (cum): Memory allocated on this line including inside any helper functions called on this line.
  • Dots (.): Zero heap allocations occurred on that line.
  ──────
  ### 3. Analyzing the Two Culprit Lines

  #### Line 50: temp := fmt.Sprintf(...)

       49.50MB   110.50MB     50:           temp := fmt.Sprintf("item-%d-val:%s", i, strconv.Itoa(i*2))

  • flat: 49.50MB: generateGarbage allocated 49.50 MB just setting up the arguments for fmt.Sprintf (boxing the integer and string interfaces into
  interface{} / any).
  • cum: 110.50MB: Total allocations grew to 110.50 MB once fmt.Sprintf and strconv.Itoa created internal reflection buffers and intermediate return
  strings.
  ──────
  #### Line 51: result += temp + "," (The Primary Bottleneck!)

        2.70GB     2.70GB     51:           result += temp + ","

  • flat: 2.70GB & cum: 2.70GB: Consumed over 96% of all memory allocations!
  • Why this happened: In Go, strings are immutable byte arrays. Each time result += ... executes inside a loop:
      1. Go cannot modify result in place.
      2. It must allocate a brand new slice on the heap.
      3. It copies all previous characters into the new slice, appends temp and ",", and discards the old slice for the Garbage Collector.
      4. With N = 100 across thousands of benchmark iterations, this causes quadratic (O(N²)) allocation pressure totaling 2.70 GB of discarded garbage.

  ──────
  ### 4. Summary: What Action to Take

  When you see a profile like this:

  1. Target the largest flat line first: Line 51 (+= string concatenation) is responsible for 2.70 GB of garbage.
  2. Fix: Replace result += ... with a pre-allocated strings.Builder via sb.Grow(size) to drop allocations from 2.81 GB down to just a few kilobytes.

### K.7 Race Detection & Continuous Integration (`-race` & Makefile)

**Goal**: Prevent data races from reaching production by integrating `-race` enabled test suites into Makefiles and CI pipelines.

#### How It Works:
The Go Race Detector instruments memory reads and writes during compilation (via ThreadSanitizer v2). It tracks synchronization events and detects when two goroutines access the same variable concurrently where at least one access is a write without synchronization.

#### Performance Trade-offs:
- **CPU Overhead**: 2x – 10x slower execution.
- **Memory Overhead**: 5x – 20x higher RAM consumption.
- **Best Practice**: Always enable `-race` in local development and automated CI pipelines, but **never** compile production releases with `-race`.

#### Makefile Automation:
Located at [`calculator/Makefile`](file:///home/khizerkhan/go_zero_to_hero/calculator/Makefile):

```makefile
.PHONY: help test test-race bench profile-cpu profile-mem pprof-cpu-cli pprof-mem-cli pprof-cpu-web pprof-mem-web clean ci

test-race: ## Run tests with the Go race detector enabled (-race)
	@echo "==> Running tests with Race Detector (-race)..."
	go test -race ./sync/... ./pointers_learning/... ./structs_composition_1/...

bench: ## Run all benchmarks with memory allocation metrics
	@echo "==> Running benchmarks..."
	go test -run=^$$ -bench=. -benchmem ./benchmark/...

profile-cpu: ## Profile a slow CPU function and generate cpu.pprof
	go test -run=^$$ -bench=BenchmarkCPUSlowFunction -cpuprofile=cpu.pprof ./benchmark/...

profile-mem: ## Profile memory allocations and generate mem.pprof
	go test -run=^$$ -bench=BenchmarkMemory -benchmem -memprofile=mem.pprof ./benchmark/...

ci: test-race bench ## CI target for automated pipelines
	@echo "==> CI pipeline checks passed successfully!"
```

#### GitHub Actions CI Example (`.github/workflows/ci.yml`):
```yaml
name: Go CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Run Tests with Race Detector
        run: |
          cd calculator
          make test-race
      - name: Run Benchmarks
        run: |
          cd calculator
          make bench
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
```bash
go test -bench=. -cpuprofile=cpu.pprof -memprofile=mem.pprof ./benchmark/...
```

### 6. Interactive & Web UI Profiler
```bash
# Web UI
go tool pprof -http=:8080 cpu.pprof
go tool pprof -http=:8080 -alloc_space mem.pprof

# CLI
go tool pprof cpu.pprof
go tool pprof -alloc_space mem.pprof
```

### 7. Run Race Detection
```bash
go test -race ./...
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

---

## 🎯 Hands-On Assignments & Solutions

1. **Concat Benchmark**: Benchmark `+` vs `bytes.Buffer` vs `strings.Builder`.
   - File: [`calculator/benchmark/benchmark_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/benchmark/benchmark_test.go)
   - Command: `go test -bench=BenchmarkConcat -benchmem ./benchmark`

2. **Slice Pre-allocation**: Benchmark dynamic `append` vs `make([]T, 0, cap)`.
   - File: [`calculator/benchmark/pre_allocation_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/benchmark/pre_allocation_test.go)
   - Command: `go test -bench=BenchmarkSlice -benchmem ./benchmark`

3. **CPU Profile**: Profile a slow function and visualize with `go tool pprof`.
   - File: [`calculator/benchmark/profile_demo_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/benchmark/profile_demo_test.go)
   - Command: `make profile-cpu` or `go test -run=^$ -bench=BenchmarkCPUSlowFunction -cpuprofile=cpu.pprof ./benchmark`
   - Visualize: `make pprof-cpu-web` or `go tool pprof -http=:8080 cpu.pprof`

4. **Memory Profile**: Identify where a function is generating garbage allocations.
   - File: [`calculator/benchmark/profile_demo_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/benchmark/profile_demo_test.go)
   - Command: `make profile-mem` or `go test -run=^$ -bench=BenchmarkMemoryGarbage -benchmem -memprofile=mem.pprof ./benchmark`
   - Analyze: `make pprof-mem-cli` or `go tool pprof -alloc_space mem.pprof` (use `top` and `list generateGarbage`)

5. **Race CI**: Makefile running tests with `-race` enabled.
   - File: [`calculator/Makefile`](file:///home/khizerkhan/go_zero_to_hero/calculator/Makefile)
   - Command: `make test-race` or `make ci`
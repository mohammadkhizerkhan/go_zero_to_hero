# Benchmarking in Go: Theory & Practical Guide

This guide explains how to run benchmarks and breaks down the theory behind [`calculator/benchmark/benchmark_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/benchmark/benchmark_test.go).

---

## 1. How to Run the Benchmark Tests

Run any of the following commands from the root directory `/home/khizerkhan/go_zero_to_hero/calculator`:

### Standard Benchmark Command (Recommended)
```bash
go test -bench=. -benchmem ./benchmark
```

### Useful CLI Variations

| Goal | Command |
| :--- | :--- |
| **Run only benchmarks (skip unit tests)** | `go test -run=^$ -bench=. -benchmem ./benchmark` |
| **Run a specific benchmark** | `go test -bench=BenchmarkConcatBuilder -benchmem ./benchmark` |
| **Run multiple iterations for statistical consistency** | `go test -bench=. -benchmem -count=5 ./benchmark` |
| **Increase benchmark duration (e.g., 3s per test)** | `go test -bench=. -benchmem -benchtime=3s ./benchmark` |
| **Run directly from inside `benchmark/` folder** | `cd benchmark && go test -bench=. -benchmem` |

---

## 2. Benchmark Output Breakdown

When you run `go test -bench=. -benchmem ./benchmark`, you see output similar to:

```text
BenchmarkConcatPlus-8        2657089       382.8 ns/op     152 B/op       6 allocs/op
BenchmarkConcatBuffer-8      6728018       163.9 ns/op     112 B/op       2 allocs/op
BenchmarkConcatBuilder-8    10182306       110.9 ns/op      64 B/op       1 allocs/op
```

### Understanding the Columns
1. **`BenchmarkConcatPlus-8`**: The benchmark name followed by `GOMAXPROCS` (number of CPU threads used).
2. **`2657089` (Iterations)**: The number of times the loop ran (`b.N`) to get a statistically valid measurement.
3. **`382.8 ns/op` (Speed)**: Nanoseconds taken per single operation (lower is faster).
4. **`152 B/op` (Memory)**: Bytes of heap memory allocated per operation (lower is better).
5. **`6 allocs/op` (Allocations)**: Number of distinct heap allocations per operation (lower reduces Garbage Collection pressure).

---

## 3. Deep Dive: Theory of `benchmark_test.go`

### A. How `testing.B` and `b.N` Work
Go benchmarks do not run for a fixed number of loops. Instead, the testing framework starts with `b.N = 1` and measures the execution time. If the elapsed time is too short (less than the default target of 1 second), Go increases `b.N` exponentially (e.g., 1, 100, 10000, 1000000...) until the benchmark runs long enough to calculate a stable average.

```go
func BenchmarkConcatPlus(b *testing.B) {
    // ...
    for i := 0; i < b.N; i++ {
        // Code executed b.N times
    }
}
```

### B. `b.ResetTimer()`
If your benchmark requires setup work (initializing slices, reading files, creating connections) before running the loop, this setup shouldn't skew the benchmark results. Calling `b.ResetTimer()` zeroes out the elapsed time and memory counters immediately before entering the `b.N` loop.

### C. The Compiler Optimization Trap & `globalSink`
Modern Go compilers use optimizations like **Dead Code Elimination (DCE)**. If the compiler detects that a variable computed inside a benchmark loop is never used outside the loop, it might optimize the loop away completely, producing false "0 ns/op" results.

To prevent DCE, we assign the loop output to an exported/package-level variable:
```go
var globalSink string

func BenchmarkConcatPlus(b *testing.B) {
    var s string
    for i := 0; i < b.N; i++ {
        // compute s...
    }
    globalSink = s // Guarantees the compiler will actually execute the loop
}
```

---

## 4. Comparing the Three String Concatenation Approaches

We concatenate 7 words: `[]string{"go", "zero", "to", "hero", "performance", "benchmarking", "tutorial"}`.

### 1. `BenchmarkConcatPlus` (`+` Operator) — 🐌 Slowest
- **Mechanism**: In Go, strings are **immutable byte slices**. Every time you do `s += w`, Go cannot modify `s` in place. It must allocate a brand-new memory block on the heap, copy old bytes, append the new word bytes, and point `s` to the new memory block.
- **Result**: 6 intermediate string allocations + 152 bytes allocated per op.
- **When to use**: Only when combining 2-3 static strings or literals where readability is paramount.

### 2. `BenchmarkConcatBuffer` (`bytes.Buffer`) — ⚡ Moderate
- **Mechanism**: `bytes.Buffer` manages a resizable `[]byte` slice. As words are written with `buf.WriteString(w)`, it grows the slice dynamically. However, converting the buffer back to a string via `buf.String()` requires allocating and copying memory to guarantee string immutability.
- **Result**: 2 allocations + 112 bytes per op.
- **When to use**: When working across general I/O interfaces (`io.Writer`, `io.Reader`).

### 3. `BenchmarkConcatBuilder` (`strings.Builder`) — 🚀 Fastest & Most Efficient
- **Mechanism**: `strings.Builder` was introduced in Go 1.10 specifically for high-efficiency string construction:
  - **`sb.Grow(64)`**: Pre-allocates enough buffer capacity upfront, avoiding slice reallocations during appending.
  - **Zero-Copy Conversion**: Unlike `bytes.Buffer.String()`, `sb.String()` uses `unsafe.Pointer` internally to return the string directly from the builder's internal byte slice without making an extra copy.
- **Result**: Exactly **1 allocation** (the pre-allocated 64 bytes) and **64 B/op**, taking ~1/4th the time of `+`.
- **When to use**: High-performance string concatenation and dynamic text generation in Go.

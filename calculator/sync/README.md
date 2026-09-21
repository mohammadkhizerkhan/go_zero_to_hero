# Concurrency & Synchronization: `sync.RWMutex` vs `sync.Mutex`

This package demonstrates how to synchronize shared memory across multiple goroutines using Go's `sync` package primitives: **`sync.RWMutex`** and **`sync.Mutex`**.

---

## 1. Core Concepts

### What is `sync.RWMutex`?
A Reader/Writer Mutual Exclusion Lock (`sync.RWMutex`) allows:
- **Multiple Readers simultaneously**: Any number of goroutines can hold an `RLock()` at the same time without blocking each other.
- **Single Exclusive Writer**: A call to `Lock()` waits until all active readers and writers finish, gaining exclusive access to the resource.

### Comparison Table

| Feature | No Lock (`NoLockCache`) | `sync.Mutex` (`MutexCache`) | `sync.RWMutex` (`Cache`) |
| :--- | :--- | :--- | :--- |
| **Concurrent Reads** | ❌ Data race if writing | ⚠️ Serialized (1 reader at a time) |  **Parallel (N readers simultaneously)** |
| **Concurrent Writes** | ❌ Data race / crash |  Serialized (1 writer at a time) |  Serialized (1 writer at a time) |
| **Lock Overhead** | None (`~0ns`) | Low for single-thread, High contention under concurrency | Slightly higher CPU cycles per lock, but scales well for read-heavy workloads |
| **Best Used When** | Read-only static data | Equal mix of reads and writes | **Read-heavy workloads** (e.g., 90%+ reads, caches, config lookups) |

---

## 2. File Overview

- [`reader.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/reader.go): Contains the cache implementations:
  - `Cache`: Uses `sync.RWMutex` (`RLock`/`RUnlock` for `Get`, `Lock`/`Unlock` for `Set`).
  - `MutexCache`: Uses standard `sync.Mutex` (`Lock`/`Unlock` for both `Get` and `Set`).
  - `NoLockCache`: Baseline implementation with no locking mechanism.
- [`reader_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/reader_test.go): Unit tests verifying concurrency safety and benchmarks comparing execution speeds.

---

## 3. How to Run Tests and Benchmarks

### 1. Run Concurrency Unit Test
```bash
go test -v ./sync
```

### 2. Run Race Detector (Ensure Thread Safety)
```bash
go test -race ./sync
```

### 3. Run Benchmarks
To run the benchmarks and inspect memory allocations:
```bash
go test -bench=. -benchmem ./sync
```

To run benchmarks with specific CPU core counts (e.g., 4 or 8 parallel workers):
```bash
go test -bench=. -benchmem -cpu=1,2,4,8 ./sync
```

---

## 4. Understanding Benchmark Results

When you run `go test -bench=. -benchmem ./sync`, the output will look similar to:

```text
BenchmarkCache_WithRLock_Parallel-8      25000000        45.2 ns/op        0 B/op        0 allocs/op
BenchmarkCache_WithMutex_Parallel-8       5000000       240.1 ns/op        0 B/op        0 allocs/op
BenchmarkCache_WithoutLock_Single-8      85000000        13.5 ns/op        0 B/op        0 allocs/op
```

### Metrics Explained:
- **`ns/op` (Nanoseconds per Operation)**: Lower is faster. Measures how long a single read/write operation took.
- **`B/op` (Bytes per Operation)**: Heap memory allocated per operation.
- **`allocs/op` (Allocations per Operation)**: Number of distinct heap allocations made per operation.

### Key Takeaways:
1. **Parallel Reads**: Under multi-core parallel execution (`b.RunParallel`), `BenchmarkCache_WithRLock_Parallel` outperforms `BenchmarkCache_WithMutex_Parallel` because goroutines do not block waiting for exclusive locks.
2. **Lock Overhead**: `BenchmarkCache_WithoutLock_Single` shows the raw map read speed. Locking adds a small fixed overhead to provide thread safety.

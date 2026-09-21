# Synchronization: `sync.Mutex` vs `sync/atomic` (Atomic Counter)

This document explains the differences, underlying theory, benchmarks, and trade-offs between using **`sync.Mutex`** and hardware-level **`sync/atomic`** primitives in Go.

---

## 1. Core Problem: Race Conditions in `count++`

In Go, incrementing an integer (`count++`) looks like a single atomic step, but at the CPU assembly level it compiles down to a **3-step Read-Modify-Write (RMW)** sequence:

1. **`MOV`**: Read the variable from main memory/cache into a CPU register.
2. **`ADD`**: Increment the register value by 1.
3. **`MOV`**: Write the modified value back to memory.

### The Race Condition
When multiple goroutines execute these steps concurrently on separate CPU cores, their memory operations interleave:

```text
Goroutine 1 (Core 1): Reads count (5) ───> Adds 1 (6) ─────────────> Writes (6)
Goroutine 2 (Core 2): Reads count (5) ────────────> Adds 1 (6) ────> Writes (6) [Lost Update!]
```
Even though two increments occurred, the final counter value is `6` instead of `7`.

---

## 2. Implementations

### A. Mutex Approach (`sync.Mutex`)
Uses runtime software locks to enforce mutual exclusion across the critical section.

```go
type MutexCounter struct {
	mu    sync.Mutex
	count int64
}

func (c *MutexCounter) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}

func (c *MutexCounter) Value() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}
```

### B. Atomic Approach (`sync/atomic`)
Uses hardware-level CPU instructions to update the integer without software locks.

```go
type AtomicCounter struct {
	count int64
}

func (c *AtomicCounter) Inc() {
	atomic.AddInt64(&c.count, 1)
}

func (c *AtomicCounter) Value() int64 {
	return atomic.LoadInt64(&c.count)
}
```

---

## 3. Underlying Mechanics & Theory

```
┌────────────────────────────────────────────────────────┐
│                   Go Runtime Layer                     │
│  [Mutex]: Goroutine Queue -> Scheduler -> Spin -> Park │
└──────────────────────────┬─────────────────────────────┘
                           │ (Bypassed by Atomic)
┌──────────────────────────▼─────────────────────────────┐
│                    CPU Hardware                        │
│  [Atomic]: Direct CPU Instruction (LOCK XADDQ)         │
└────────────────────────────────────────────────────────┘
```

### How `sync.Mutex` Works
1. **Fast-path**: Attempts a fast atomic Compare-And-Swap (CAS) to acquire the lock.
2. **Slow-path (Contention)**: If locked, waiting goroutines may spin briefly on the CPU. If the lock is still unavailable, the Go runtime **parks (suspends)** the goroutine and performs a context switch.
3. **Unlock**: The Go runtime must wake up (unpark) sleeping goroutines and reschedule them on available threads (`M`s).
> **Overhead**: Goroutine context switching, scheduler coordination, and queue management consume significant CPU cycles under high contention.

### How `sync/atomic` Works
`atomic.AddInt64` bypasses the Go runtime scheduler completely and maps directly to single CPU hardware instructions:
- **x86-64**: `LOCK XADDQ` (locks the cache line at the hardware bus level during the operation).
- **ARM64**: `LDADD` or `LDREX`/`STREX` (Load-Linked / Store-Conditional).

The CPU's hardware cache-coherence protocol (MESI/MOESI) synchronizes the value directly across CPU cores in silicon without putting goroutines to sleep.

---

## 4. Benchmark Comparison

### Run Command:
```bash
cd calculator
go test -bench="BenchmarkCounter" -benchmem ./sync
```

### Benchmark Results:
```text
BenchmarkCounter_Mutex_Single-8       90107972        14.44 ns/op        0 B/op        0 allocs/op
BenchmarkCounter_Atomic_Single-8     181538274         6.45 ns/op        0 B/op        0 allocs/op

BenchmarkCounter_Mutex_Parallel-8     19733916        76.17 ns/op        0 B/op        0 allocs/op
BenchmarkCounter_Atomic_Parallel-8    65265970        18.93 ns/op        0 B/op        0 allocs/op
```

### Summary of Performance:
- **Single-Core / Low Contention:** `atomic` is **~2.2x faster** than `sync.Mutex`.
- **Multi-Core / High Parallel Contention:** `atomic` is **~4.0x faster** than `sync.Mutex`.

---

## 5. Pros and Cons Comparison

| Feature | `sync.Mutex` | `sync/atomic` (`atomic.AddInt64`) |
| :--- | :--- | :--- |
| **Throughput / Speed** | Slower under high contention (~76 ns/op) | **Extremely fast** (~18 ns/op) |
| **Mechanism** | Go runtime scheduler, goroutine parking/unparking | Hardware CPU instructions |
| **Complexity Handled** | Can protect **multi-line code blocks**, maps, slices, structs | Protects **only single primitive values** (int, pointer, bool) |
| **Deadlock Risk** | Possible if locks are acquired out of order or forgotten | **Impossible** (lock-free) |
| **Readability & Safety** | Standard, clear critical section boundaries | Requires discipline (must consistently use `atomic.Load`/`Store`) |

---

## 6. When to Use Which?

### Use `sync/atomic` when:
- Tracking simple numeric counters (e.g., request counters, active connection counts, metrics).
- Managing boolean state flags (e.g., `isShutdown int32` or `atomic.Bool`).
- Hot-reloading configurations via atomic pointers (`atomic.Pointer[T]` or `atomic.Value`).

### Use `sync.Mutex` when:
- Updating multiple fields together that must remain synchronized (e.g., debiting Account A and crediting Account B).
- Protecting non-primitive data structures such as `map`, `slice`, or channels.
- Guarding critical sections that perform I/O, database queries, or external network requests.

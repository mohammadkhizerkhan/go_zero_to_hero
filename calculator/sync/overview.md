# Go `sync` Package: Complete Visual Overview & Cheatsheet

A single-glance reference guide covering every synchronization primitive implemented in the `calculator/sync` package.

---

## 🧭 1. Decision Matrix: Which Primitive Should I Use?

```
What are you trying to synchronize?
│
├── Simple numeric counter / flag / pointer?
│   └── ➔ `sync/atomic` (Lock-free, ~4x faster than Mutex)
│
├── One-time setup / Singleton / Lazy initialization?
│   └── ➔ `sync.Once` (Guarantees execution exactly once)
│
├── Wait for N goroutines to finish?
│   └── ➔ `sync.WaitGroup` (`Add`, `Done`, `Wait`)
│
├── Coordinate workers waiting for an event/condition?
│   └── ➔ `sync.Cond` (`Wait`, `Signal`, `Broadcast`)
│
└── Protecting shared state / complex data structures?
    ├── Mostly Reads (>80-90% reads, few writes)?
    │   └── ➔ `sync.RWMutex` (`RLock`/`RUnlock` for readers, `Lock`/`Unlock` for writers)
    └── Balanced Reads and Writes or simple critical sections?
        └── ➔ `sync.Mutex` (`Lock`/`Unlock`)
```

---

## ⚡ 2. Quick Reference Cards

### 1. `sync.Mutex` vs `sync.RWMutex`
> **Files:** [`reader.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/reader.go) | [`reader_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/reader_test.go)

- **`sync.Mutex`**: Mutual exclusion lock. Only **1 goroutine** (reader or writer) can access at a time.
- **`sync.RWMutex`**: Reader/Writer lock. **$N$ readers can read in parallel** simultaneously; writers get exclusive access.

```go
// Reader-heavy cache pattern:
var mu sync.RWMutex
var cache = make(map[string]string)

// Read (Parallel)
mu.RLock()
val := cache["key"]
mu.RUnlock()

// Write (Exclusive)
mu.Lock()
cache["key"] = "new_value"
mu.Unlock()
```

---

### 2. `sync/atomic` (Hardware-Level Primitives)
> **Files:** [`atomic.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/atomic.go) | [`atomic_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/atomic_test.go) | [`atomic.md`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/atomic.md)

- Bypasses the Go runtime scheduler by executing direct CPU instructions (`LOCK XADDQ` on x86, `LDADD` on ARM).
- Lock-free, immune to deadlocks, and ~4x faster than a Mutex under multi-core contention.

```go
var count int64

// Atomic increment
atomic.AddInt64(&count, 1)

// Atomic read
val := atomic.LoadInt64(&count)
```

---

### 3. `sync.Once` (Singleton & Lazy Initialization)
> **Files:** [`once.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/once.go) | [`once_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/once_test.go)

- Ensures a function runs **only once across the entire application lifetime**, even if called concurrently by thousands of goroutines.
- Ideal for database connections, heavy configuration loading, or singleton instances.

```go
var (
    instance *Database
    once     sync.Once
)

func GetDB() *Database {
    once.Do(func() {
        instance = initExpensiveDatabase()
    })
    return instance
}
```

---

### 4. `sync.Cond` (Condition Variables & Broadcast)
> **Files:** [`condition.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/condition.go) | [`condition_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/condition_test.go)

- Allows goroutines to pause execution (`Wait()`) until a specific state condition becomes true.
- **`Signal()`**: Wakes up **one** waiting goroutine.
- **`Broadcast()`**: Wakes up **all** waiting goroutines simultaneously (e.g., service ready gate, pub/sub).
- **Rule:** Always wrap `cond.Wait()` inside a `for !condition` loop to guard against spurious wakeups.

```go
var mu sync.Mutex
cond := sync.NewCond(&mu)
ready := false

// Waiting Goroutine
mu.Lock()
for !ready {
    cond.Wait() // Releases lock, suspends goroutine, re-acquires lock upon wake
}
mu.Unlock()

// Broadcasting Goroutine
mu.Lock()
ready = true
mu.Unlock()
cond.Broadcast() // Wakes all waiting workers
```

---

### 5. `sync.WaitGroup` (Goroutine Barrier)
> Used across all test files

- Coordinates multiple concurrent goroutines by waiting until a counter decrements to 0.

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        doWork(id)
    }(i)
}

wg.Wait() // Blocks until all 5 goroutines call wg.Done()
```

---

## 📊 3. Performance & Characteristic Comparison

| Primitive | Synchronization Level | Goroutine Blocking? | Deadlock Risk? | Best Use Case |
| :--- | :--- | :--- | :--- | :--- |
| **`sync.Mutex`** | Software / Runtime | Yes (Puts goroutines to sleep) | Yes | Complex multi-variable updates, slices, maps |
| **`sync.RWMutex`** | Software / Runtime | Yes (Only for writers) | Yes | Read-heavy caches (>80% reads) |
| **`sync/atomic`** | Hardware / CPU Cache | **No** (Lock-free) | **No** | Numeric counters, state flags, atomic pointers |
| **`sync.Once`** | Fast-path atomic + Mutex | Only during initial run | Yes (if nested) | Database pools, singletons, init logic |
| **`sync.Cond`** | Software + Mutex | Yes (Until signaled) | Yes | Multi-worker barrier, reusable event queues |
| **`sync.WaitGroup`** | Atomic counter + Semaphore | Yes (At `Wait()`) | Yes (if counter != 0)| Fan-out / Fan-in task completion |

---

## 🛠️ 4. Quick Command Reference

Run all commands from the `calculator` folder (`cd calculator`):

### Run All Tests
```bash
go test -v ./sync
```

### Run Race Detector (Find data races automatically)
```bash
go test -race ./sync
```

### Run Specific Test
```bash
go test -v -count=1 -run TestCondBroadcast ./sync
go test -v -count=1 -run TestAtomicCounter ./sync
go test -v -count=1 -run TestSingleton ./sync
```

### Run All Benchmarks with Memory Stats
```bash
go test -bench=. -benchmem ./sync
```

### Run Multi-Core Scaled Benchmarks
```bash
go test -bench="BenchmarkCounter" -benchmem -cpu=1,2,4,8 ./sync
```

---

## ⚠️ 5. Golden Rules to Avoid Bugs in Go Concurrency

1. **Never copy synchronization primitives**: `sync.Mutex`, `sync.RWMutex`, `sync.Cond`, and `sync.WaitGroup` must never be passed by value. Always pass by pointer or embed in a pointer receiver struct.
2. **Always `defer mu.Unlock()` / `defer wg.Done()`**: Ensures cleanup occurs even if panics or early returns happen.
3. **Always wrap `cond.Wait()` in a `for` loop**: Never use an `if` check; OS thread schedulers can issue spurious wakeups.
4. **Never call `once.Do` recursively**: Calling `once.Do` inside another `once.Do` with the same `sync.Once` variable causes an instant deadlock.
5. **Always use `-race` during CI/testing**: Go’s built-in race detector catches unsynchronized memory accesses that unit tests might miss.

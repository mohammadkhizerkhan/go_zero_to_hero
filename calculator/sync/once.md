# Synchronization: `sync.Once` (Safe One-Time Initialization)

This document explains the internal mechanics, architecture, execution lifecycle, and design patterns of Go's **`sync.Once`** primitive.

---

## 1. What is `sync.Once`?

`sync.Once` is a concurrency primitive in Go that guarantees a specific action (`f func()`) is executed **strictly once** across the entire lifecycle of an application, regardless of how many goroutines invoke it concurrently.

### Typical Use Cases:
- **Singleton Database Connection Pools** (e.g., initializing PostgreSQL/Redis pools).
- **Heavy Configuration Loading** (parsing JSON/YAML files once on demand).
- **Compiling Complex Regexes** (pre-compiling `regexp.MustCompile` at runtime).
- **One-time Teardown/Cleanup** (closing channels or flushing buffers safely).

---

## 2. Code Example: Singleton Pattern

> **Files:** [`once.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/once.go) | [`once_test.go`](file:///home/khizerkhan/go_zero_to_hero/calculator/sync/once_test.go)

```go
package sync

import (
	"fmt"
	"sync"
)

type DatabaseConnection struct {
	URL string
}

var (
	dbInstance *DatabaseConnection // 1. Holds the actual instance
	once       sync.Once           // 2. The boolean execution gate (done: 0 or 1)
)

func GetDBInstance() *DatabaseConnection {
	// 3. Gatekeeper: executes the closure ONLY if done == 0
	once.Do(func() {
		fmt.Println("Initializing expensive database connection pool...")
		dbInstance = &DatabaseConnection{URL: "postgres://localhost:5432/db"}
	})

	// 4. Parent function: always returns the created instance
	return dbInstance
}
```

---

## 3. How `sync.Once` Works Internally

`sync.Once` is the Go standard library's optimized, standardized implementation of the **Double-Checked Locking (DCL)** pattern.

### Go Standard Library Source Code (`sync/once.go`):

```go
type Once struct {
    done uint32     // Atomic flag: 0 = not executed, 1 = executed
    m    Mutex      // Protects the slow path during initial execution
}

func (o *Once) Do(f func()) {
    // ⚡ Fast Path: Lock-free atomic read (Check #1)
    if atomic.LoadUint32(&o.done) == 0 {
        o.doSlow(f)
    }
}

func (o *Once) doSlow(f func()) {
    o.m.Lock()
    defer o.m.Unlock()

    // 🔒 Slow Path: Double-check under lock (Check #2)
    if o.done == 0 {
        defer atomic.StoreUint32(&o.done, 1)
        f()
    }
}
```

---

## 4. Execution Flow Diagram

```
                       Caller invokes once.Do(f)
                                  │
                                  ▼
                  ┌───────────────────────────────┐
                  │ atomic.LoadUint32(&done) == 1 │
                  └───────────────┬───────────────┘
                                  │
                 ┌────────────────┴────────────────┐
          YES (done == 1)                    NO (done == 0)
                 │                                 │
                 ▼                                 ▼
      ⚡ Return immediately               Acquire o.m.Lock()
    (Sub-nanosecond, No Lock)                      │
                                                   ▼
                                         Double-check: done == 0?
                                                   │
                                  ┌────────────────┴────────────────┐
                           NO (done == 1)                    YES (done == 0)
                                  │                                 │
                                  ▼                                 ▼
                         Unlock & Return                     Execute f()
                       (Another goroutine finished)                  │
                                                                    ▼
                                                         atomic.StoreUint32(&done, 1)
                                                                    │
                                                                    ▼
                                                             Unlock & Return
```

---

## 5. Key Architecture Insights

### A. It is a Boolean Gatekeeper, Not an Object Cache
`sync.Once` **does not store, cache, or return** your object instance.
- `once.Do(f)` accepts a `func()` with no arguments and no return values.
- It acts purely as a synchronized boolean switch (`done == 0` or `1`).
- The outer caller function (`GetDBInstance`) is responsible for holding the variable (`dbInstance`) and returning it.

### B. Blocking Behavior on First Execution
If 50 goroutines call `once.Do` simultaneously when `done == 0`:
- **Goroutine 1** acquires the mutex and runs `f()`.
- **Goroutines 2–50** block on `o.m.Lock()`.
- They will **not return** until Goroutine 1 completely finishes executing `f()`.
- Once `f()` completes, `done` becomes `1`, the lock is released, and all waiting goroutines wake up and return.
- **Guarantee:** By the time `once.Do()` returns on *any* goroutine, the resource is guaranteed to be 100% initialized.

---

## 6. Comparison: Go `sync.Once` vs Java Double-Checked Locking

In languages like Java, developers must manually construct the Double-Checked Locking pattern:

#### Java (Manual DCL):
```java
public class Database {
    private static volatile Database instance; // Requires 'volatile' memory barrier

    public static Database getInstance() {
        if (instance == null) {                 // Check 1 (No lock)
            synchronized (Database.class) {     // Lock
                if (instance == null) {         // Check 2 (Under lock)
                    instance = new Database();
                }
            }
        }
        return instance;
    }
}
```

#### Go (`sync.Once`):
```go
var (
    instance *Database
    once     sync.Once
)

func GetInstance() *Database {
    once.Do(func() {
        instance = &Database{}
    })
    return instance
}
```

**Advantage in Go:** The synchronization complexity (atomic checks, memory barriers, mutexes) is completely encapsulated in `sync.Once`.

---

## 7. ⚠️ Gotchas and Best Practices

1. **Deadlock on Recursive Calls:**
   Never call `once.Do` with the same `sync.Once` instance inside the function `f()` itself.
   ```go
   // ❌ DEADLOCK:
   var once sync.Once
   once.Do(func() {
       once.Do(func() { // Blocks forever waiting for the outer lock!
           fmt.Println("Will never run")
       })
   })
   ```

2. **Panics Mark `Once` as Done:**
   Because `defer atomic.StoreUint32(&o.done, 1)` runs even if `f()` panics, if initialization fails with a panic, `sync.Once` will consider itself completed and will **not** retry on subsequent calls.

3. **Pass by Pointer:**
   Never copy a struct that embeds `sync.Once` by value. Always pass pointers to avoid duplicating the internal mutex and `done` state.

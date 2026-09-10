Here is a comprehensive documentation reference covering all five Go error handling patterns, their mechanical behaviors, and standard code usage.

---

**1. The Divider (Explicit Error Returns)**

- **Concept**: Go does not use exception handling (`try/catch`). Functions return errors as explicit secondary return values (`(T, error)`).
- **Mechanism**: If an invalid operation occurs (e.g., division by zero), return the zero-value for the result type alongside an error initialized via `errors.New()`. On success, return the calculated result and `nil`.
- **Code Pattern**:

```go
func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("cannot divide by zero")
    }
    return a / b, nil
}

```

---

**2. The Sentinel (Package-Level Errors & `errors.Is`)**

- **Concept**: Sentinel errors are pre-declared, immutable error instances used to represent specific expected domain conditions (e.g., `io.EOF`, `sql.ErrNoRows`).
- **Mechanism**: Declare a package-level variable using `var ErrName = errors.New(...)`. Check if an incoming error matches this target using `errors.Is(err, ErrName)`.
- **Code Pattern**:

```go
var ErrNotFound = errors.New("resource not found")

func GetUser(id int) (string, error) {
    if id <= 0 {
        return "", ErrNotFound
    }
    return "Alice", nil
}

// Caller side:
if errors.Is(err, ErrNotFound) {
    // Handle 404 condition safely
}

```

---

**3. The Rich Error (Custom Error Structs & `errors.As`)**

- **Concept**: When simple error strings are insufficient, custom structs attach rich metadata (e.g., HTTP status codes, field validation rules, error codes) directly to the error object.
- **Mechanism**: Implement the `Error() string` method on a struct pointer. Callers extract the concrete struct fields from a generic `error` interface using `errors.As(err, &target)`.
- **Code Pattern**:

```go
type AppError struct {
    Code    int
    Message string
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unpacking fields on caller side:
var appErr *AppError
if errors.As(err, &appErr) {
    fmt.Println("Extracted HTTP Code:", appErr.Code)
}

```

---

**4. The Wrapper (Context Annotations & Error Chains)**

- **Concept**: Allows higher-level callers to add operational context ("where" and "why") to low-level errors while preserving the root cause.
- **Mechanism**: Use `fmt.Errorf("context: %w", err)` with the `%w` verb. This constructs a singly linked list inside Go where each error stores a pointer to the inner cause (accessible via `Unwrap() error`). `errors.Is` and `errors.As` automatically traverse this chain.
- **Code Pattern**:

```go
func Connect() error {
    return errors.New("connection refused")
}

func Query() error {
    if err := Connect(); err != nil {
        return fmt.Errorf("db failed: %w", err) // Chain: "db failed: connection refused"
    }
    return nil
}

```

---

**5. The Safe Recovery (Panic, Defer & Recover)**

- **Concept**: Panics are unrecoverable runtime exceptions (e.g., array index out of bounds, nil pointer dereferences). Services use `recover()` to prevent a single panicking request from crashing the entire application process.
- **Mechanism**: Place a deferred function call at the entry point of a handler/goroutine. Inside `defer`, execute `recover()`. If `recover()` returns non-nil, a panic was intercepted.
- **Code Pattern**:

```go
func SafeWebHandler() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered from panic gracefully: %v", r)
            // Optionally return HTTP 500 Internal Server Error
        }
    }()

    // Simulated unexpected crash
    var ptr *string
    _ = *ptr // Triggers nil-pointer dereference panic
}

```

---

Which of these 5 assignments would you like to construct in code first to verify your understanding?

Go interfaces are implicit, satisfied automatically whenever a type implements the required methods without needing explicit `implements` keywords.

## Core Concepts

- **Implicit Interfaces:** A type satisfies an interface simply by defining matching method signatures.
- **Duck Typing at Compile-Time:** If a type has the methods an interface requires, Go treats it as satisfying that interface with full compile-time type safety.
- **The `any` Type:** `type any = interface{}` is the empty interface. Because it specifies zero methods, every Go type satisfies it.

---

## Assignment Solutions

### 1. The Shape Solver

Demonstrates basic implicit interface satisfaction and polymorphism across different concrete structs.

```go
package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func PrintArea(s Shape) {
	fmt.Printf("Area: %.2f\n", s.Area())
}

func main() {
	c := Circle{Radius: 5}
	r := Rectangle{Width: 4, Height: 6}

	PrintArea(c)
	PrintArea(r)
}

```

### 2. The Writer Adapter

Shows how satisfying standard library interfaces like `io.Writer` allows seamless integration with functions such as `fmt.Fprintf`.

```go
package main

import (
	"fmt"
	"os"
)

type ConsoleWriter struct{}

func (cw ConsoleWriter) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func main() {
	cw := ConsoleWriter{}

	// fmt.Fprintf accepts any io.Writer (anything with Write([]byte) (int, error))
	n, err := fmt.Fprintf(cw, "Formatted output sent directly through ConsoleWriter!\n")
	if err != nil {
		fmt.Println("Write error:", err)
		return
	}
	fmt.Printf("Wrote %d bytes.\n", n)
}

```

### 3. The Type Switch

Demonstrates safely unpacking dynamic types stored inside an `any` (`interface{}`) value.

```go
package main

import "fmt"

type User struct {
	Name string
}

func main() {
	data := map[string]any{
		"age":   28,
		"role":  "Developer",
		"user":  User{Name: "Sam"},
		"valid": true,
	}

	for key, val := range data {
		switch v := val.(type) {
		case int:
			fmt.Printf("Key %q is an int: %d (doubled = %d)\n", key, v, v*2)
		case string:
			fmt.Printf("Key %q is a string: %q\n", key, v)
		case User:
			fmt.Printf("Key %q is a User struct with Name: %s\n", key, v.Name)
		default:
			fmt.Printf("Key %q is an unhandled type (%T)\n", key, v)
		}
	}
}

```

### 4. The Mock Interface

Enables dependency injection by swapping real and mock implementations for clean testing without altering caller code.

```go
package main

import "fmt"

type PaymentProcessor interface {
	Process(amount float64) error
}

type StripeProcessor struct {
	APIKey string
}

func (s StripeProcessor) Process(amount float64) error {
	fmt.Printf("[Stripe API] Processing real payment of $%.2f...\n", amount)
	return nil
}

type MockProcessor struct {
	LastProcessed float64
}

func (m *MockProcessor) Process(amount float64) error {
	m.LastProcessed = amount
	fmt.Printf("[Mock] Intercepted payment of $%.2f (no network call)\n", amount)
	return nil
}

func Checkout(p PaymentProcessor, amount float64) {
	if err := p.Process(amount); err != nil {
		fmt.Println("Payment error:", err)
		return
	}
	fmt.Println("Checkout successful.")
}

func main() {
	// Production usage
	stripe := StripeProcessor{APIKey: "sk_live_abc123"}
	Checkout(stripe, 99.95)

	// Unit test usage
	mock := &MockProcessor{}
	Checkout(mock, 19.99)
	fmt.Printf("Mock verified amount: $%.2f\n", mock.LastProcessed)
}

```

### 5. Interface Segregation & Composition

Implements the **Interface Segregation Principle** by breaking large monolithic interfaces into smaller, focused behaviors that can be composed on demand.

```go
package main

import "fmt"

// Small, single-purpose interfaces
type Reader interface {
	Read() string
}

type Writer interface {
	Write(data string)
}

type Closer interface {
	Close() error
}

// Composition: Embedding Reader and Writer into a combined interface
type ReadWriter interface {
	Reader
	Writer
}

// Concrete implementation satisfying ReadWriter and Closer
type DataBuffer struct {
	content string
}

func (b *DataBuffer) Read() string {
	return b.content
}

func (b *DataBuffer) Write(data string) {
	b.content = data
}

func (b *DataBuffer) Close() error {
	fmt.Println("Buffer closed.")
	return nil
}

func Transfer(rw ReadWriter, payload string) {
	rw.Write(payload)
	fmt.Println("Transferred data:", rw.Read())
}

func main() {
	buf := &DataBuffer{}
	Transfer(buf, "Hello from composed interface!")
	buf.Close()
}

```

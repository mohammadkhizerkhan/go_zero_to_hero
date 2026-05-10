# Topic C: Structs & Composition in Go (Not Inheritance)
A complete quick-reference guide with theory, examples, and practice tasks.

## 1. Big Picture
Go does not use class inheritance like Java/C++.  
Go uses:
1. Structs for data.
2. Methods on structs for behavior.
3. Embedding for composition (reusing behavior).
4. Interfaces for abstraction.

Core mental model:
1. Build larger things by combining smaller structs.
2. Prefer clear, concrete return types.
3. Accept behavior contracts (interfaces) as inputs.

---

## 2. Embedding: Composition, Not Inheritance

### What it means
When you embed one struct into another, the outer struct *contains* the embedded struct.

That is a **has-a** relationship, not an **is-a** inheritance chain.

### Example
```go
package main

import (
	"fmt"
	"time"
)

type BaseEntity struct {
	ID        string
	CreatedAt time.Time
}

type User struct {
	BaseEntity // embedded
	Name       string
}

func main() {
	u := User{
		BaseEntity: BaseEntity{
			ID:        "u-100",
			CreatedAt: time.Now(),
		},
		Name: "Khizer",
	}

	// Promoted fields
	fmt.Println(u.ID)
	fmt.Println(u.CreatedAt)

	// Explicit path still works
	fmt.Println(u.BaseEntity.ID)
}
```

### Key rule
Embedding gives convenient field/method access, but does not change type identity.

### Common pitfall
Thinking `User` is a subtype of `BaseEntity`. It is not.

---

## 3. Promoted Fields and Promoted Methods

### What gets promoted
If `Child` embeds `Base`, then exported fields/methods of `Base` are accessible directly on `Child` if there is no ambiguity.

### Example
```go
package main

import "fmt"

type Base struct{}

func (Base) Describe() string {
	return "base description"
}

type Child struct {
	Base
}

// Child has its own Describe
func (Child) Describe() string {
	return "child description"
}

func main() {
	c := Child{}

	// Uses Child's method
	fmt.Println(c.Describe()) // child description

	// Explicitly call embedded Base method
	fmt.Println(c.Base.Describe()) // base description
}
```

### Key rule
If both outer and embedded type define same method name, outer method wins for direct call.

### Common pitfall
Calling this “override” exactly like OOP inheritance. In Go this is method shadowing + explicit access path.

---

## 4. Constructor Pattern and Invariants

### Why constructor pattern
Constructors centralize validation and prevent invalid object states.

### Pattern
1. Keep struct unexported (`server`).
2. Expose constructor (`NewServer`).
3. Validate inputs inside constructor.
4. Return `nil` or error on invalid input.

### Example
```go
package main

import "fmt"

type server struct {
	port int
}

func NewServer(port int) *server {
	if port <= 0 || port > 65535 {
		return nil
	}
	return &server{port: port}
}

func (s *server) Port() int {
	if s == nil {
		return 0
	}
	return s.port
}

func main() {
	ok := NewServer(8080)
	bad := NewServer(70000)

	fmt.Println(ok != nil, ok.Port()) // true 8080
	fmt.Println(bad == nil)           // true
}
```

### Key rule
Constructor protects invariants: “a valid object must satisfy rules from birth.”

### Common pitfall
Exporting fields and letting callers create invalid values directly.

---

## 5. “Accept Interfaces, Return Structs”

### What this means
1. **Accept interfaces** as function parameters where behavior is needed.
2. **Return concrete structs** (or pointers to structs) from constructors/factories for clarity and capability.

### Example
```go
package main

import "fmt"

type Logger interface {
	Log(msg string)
}

type ConsoleLogger struct{}

func (ConsoleLogger) Log(msg string) {
	fmt.Println("LOG:", msg)
}

type Service struct {
	name   string
	logger Logger
}

func NewService(name string, logger Logger) *Service {
	return &Service{name: name, logger: logger}
}

func (s *Service) Run() {
	s.logger.Log("service " + s.name + " running")
}

func main() {
	svc := NewService("payments", ConsoleLogger{})
	svc.Run()
}
```

### Why it works
1. Input is flexible (any type implementing `Logger` works).
2. Output is explicit and useful (`*Service` exposes full behavior).

### Common pitfall
Returning an interface from constructor too early and hiding useful methods unnecessarily.

---

## 6. Mixin-Style Composition in Go

### Idea
Compose behavior by embedding multiple small structs.

### Example
```go
package main

import "fmt"

type Drivable struct{}

func (Drivable) Drive() string {
	return "driving on road"
}

type Flyable struct{}

func (Flyable) Fly() string {
	return "flying in sky"
}

type FlyingCar struct {
	Drivable
	Flyable
	Model string
}

func main() {
	car := FlyingCar{Model: "SkyRunner"}
	fmt.Println(car.Drive())
	fmt.Println(car.Fly())
}
```

### Key rule
Small reusable behavior units compose cleanly.

### Common pitfall
Embedding too many types and creating unclear APIs or name conflicts.

---

## 7. JSON Modeling with Struct Tags

### Goal
Map nested JSON into structs with `json` tags and validate decoding via tests.

### Example model + parser
```go
package main

import "encoding/json"

type APIResponse struct {
	Meta MetaInfo `json:"meta"`
	Data Payload  `json:"data"`
}

type MetaInfo struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

type Payload struct {
	User   APIUser `json:"user"`
	Orders []Order `json:"orders"`
}

type APIUser struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Address Address  `json:"address"`
	Tags    []string `json:"tags,omitempty"`
}

type Address struct {
	City    string `json:"city"`
	ZipCode string `json:"zip_code"`
}

type Order struct {
	OrderID string `json:"order_id"`
	Amount  int    `json:"amount"`
	Status  string `json:"status"`
}

func ParseAPIResponse(raw []byte) (APIResponse, error) {
	var out APIResponse
	err := json.Unmarshal(raw, &out)
	return out, err
}
```

### Test pattern
```go
package main

import "testing"

func TestParseAPIResponse(t *testing.T) {
	raw := []byte(`{
		"meta": {"request_id":"req-123","timestamp":"2026-05-10T12:00:00Z"},
		"data": {
			"user": {
				"id": 10,
				"name": "Khizer",
				"email":"k@example.com",
				"address":{"city":"Lahore","zip_code":"54000"}
			},
			"orders":[
				{"order_id":"o-1","amount":120,"status":"paid"},
				{"order_id":"o-2","amount":80,"status":"pending"}
			]
		}
	}`)

	got, err := ParseAPIResponse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Meta.RequestID != "req-123" {
		t.Fatalf("wrong request_id: %s", got.Meta.RequestID)
	}
	if got.Data.User.Address.City != "Lahore" {
		t.Fatalf("wrong city: %s", got.Data.User.Address.City)
	}
	if got.Data.Orders[1].Status != "pending" {
		t.Fatalf("wrong status: %s", got.Data.Orders[1].Status)
	}
}

func TestParseAPIResponseMalformed(t *testing.T) {
	_, err := ParseAPIResponse([]byte(`{"meta":`))
	if err == nil {
		t.Fatal("expected error for malformed json")
	}
}
```

---

## 8. Practice Set (Do This in Order)

### Practice A: Promoted Field
1. Create `BaseEntity { ID, CreatedAt }`.
2. Embed in `User`.
3. Assert `user.ID` and `user.BaseEntity.ID` are equal.

### Practice B: Override Trap
1. Create `Base.Describe()`.
2. Embed `Base` in `Child`.
3. Add `Child.Describe()`.
4. Verify output of both call paths.

### Practice C: Constructor Pattern
1. Build unexported `server`.
2. Add `NewServer(port int) *server`.
3. Validate invalid ports.
4. Write tests for valid and invalid cases.

### Practice D: Mixin
1. Create `Drivable`, `Flyable`.
2. Embed in `FlyingCar`.
3. Verify `Drive()` and `Fly()` both work.

### Practice E: JSON Modeler
1. Define nested structs and tags.
2. Parse JSON.
3. Add malformed input test.
4. Add missing optional field case.

---

## 9. Self-Check (Memory Recap)
Answer these without looking:

1. Why embedding is composition, not inheritance?
2. What does “promoted field” mean?
3. Why constructor exists even when struct creation is simple?
4. Why accept interfaces but return structs?
5. In method name conflict, what does direct call choose?

If you can answer all 5 confidently and implement each mini exercise from scratch, Topic C is learned well.

---

## 10. Quick One-Line Memory Hooks
1. Embed to compose, not to inherit.
2. Promoted means shortcut access.
3. Constructors keep invalid states out.
4. Inputs abstract, outputs concrete.
5. Small structs + interfaces = scalable Go design.
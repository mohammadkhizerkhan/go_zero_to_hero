

## Plan: Topic C Fast-Track (Right Now)

Topic C: Structs & Composition (Not Inheritance) 
Theory: Embedding fields. Promoted methods. "Accept interfaces, return structs." 
Assignments: 
1. The JSON Modeler: Map a complex nested JSON response to Go structs using struct 
tags `json:"field"`. Unmarshal it. 
2. The Promoted Field: Embed BaseEntity (ID, CreatedAt) into User. Access user.ID directly. 
3. The Constructor Pattern: Create a private struct server. Create a public NewServer(port 
int) *server. Prevent direct initialization. 
4. The Override Trap: Embed Base in Child. Give both a Describe() method. Call 
child.Describe(). Call child.Base.Describe(). 
5. The Mixin: Create Drivable and Flyable structs. Embed both in FlyingCar. Use methods 
from both.

This is a same-session structure: short theory exploration, then immediate hands-on.

### 1. Theory To Explore First (30-40 min)
Explore in this exact order:
1. Struct basics and field tags (`json:"..."`)
2. Embedding fields (composition)
3. Promoted methods/fields
4. Method shadowing (not true inheritance override)
5. Constructor pattern with unexported type + exported constructor
6. Design rule: accept interfaces, return structs

What to do while exploring:
1. Write 1 rule per concept
2. Write 1 common pitfall per concept

### 2. Hands-on Problems You Should Do Immediately

1. **JSON Modeler (45-60 min)**  
Map complex nested JSON into structs with tags.  
Unmarshal and assert nested values in tests.  
Add one missing/optional field case.

2. **Promoted Field (25-35 min)**  
Create `BaseEntity { ID, CreatedAt }`, embed in `User`.  
Access `user.ID` directly in code and test it.

3. **Override Trap (25-35 min)**  
Embed `Base` in `Child`.  
Both implement `Describe()`.  
Call `child.Describe()` and `child.Base.Describe()`; verify difference.

4. **Constructor Pattern (30-45 min)**  
Create unexported `server` struct and exported `NewServer(port int) *server`.  
Add validation (invalid port).  
Test valid + invalid construction.

5. **Mixin Composition (30-40 min)**  
Create `Drivable` and `Flyable` structs, embed in `FlyingCar`.  
Use methods from both through one `FlyingCar` value.

### 3. How To Verify Learning (20-30 min)
1. Run topic tests only: `go test ./<topic_c_folder> -v`
2. Run all module tests: `go test ./... -v`
3. Explain from memory:
- Why embedding is composition, not inheritance
- How promoted methods work
- Why constructor protects invariants
- What “accept interfaces, return structs” means in API design

### 4. Scope Boundaries
Included:
1. Struct tags
2. Embedding and promotion
3. Shadowing behavior
4. Constructor privacy
5. Mixin composition

Excluded for now:
1. Reflection-heavy mapping
2. Frameworks/libraries
3. Advanced interface mocking

If you want, next I can give you a **single-session checklist** in execution order with exact file names for each assignment so you can start immediately.
# Go Backend Deep Dive: Gin, Postgres & Booking-System Concurrency

Oct 3, 2026 · @khan

## Why this stack

Every choice in this module trades a convenience for a guarantee, and understanding the trade is more valuable than memorizing the API.

**Gin over raw `net/http`:** raw `net/http` gives you nothing but routing primitives — you'd hand-roll middleware chaining, context propagation, and JSON binding on every project. Gin standardizes that plumbing without hiding what's happening underneath; its `Context` is a thin wrapper, not a framework-managed black box, so you can always drop to the underlying `*http.Request`/`http.ResponseWriter` when you need to.

**`gin.New()` over `gin.Default()`:** `gin.Default()` silently attaches `Logger()` and `Recovery()` in a fixed order. That's fine for a demo, wrong for production, for two concrete reasons: (1) you often want your own structured logger instead of Gin's plain-text one, and attaching both means double-logging; (2) middleware order matters — `Recovery()` must wrap everything that could panic, including your auth and logging middleware, or a panic in your own logger crashes the process instead of returning a 500. `gin.New()` forces you to decide that order consciously instead of inheriting Gin's defaults.

**sqlc over GORM:** GORM is a runtime ORM — it builds SQL from Go structs via reflection, which means a typo in a struct tag or a renamed column is a *runtime* error, often one that only surfaces under a specific query path in production. sqlc flips the direction: you write real SQL, and it generates Go types *from* that SQL at build time. A schema/query mismatch becomes a compile failure. The cost is you lose GORM's auto-migrations and dynamic query building — a fair trade for correctness in a system where "the query silently returned the wrong column" is a worse failure mode than "the build failed."

**Why pair them at all:** Gin owns the HTTP boundary (parsing, validating, authenticating requests); sqlc owns the data boundary (typed, reviewable SQL). Neither leaks into the other — your handlers never embed raw SQL, and your queries never know about HTTP status codes. That separation is what makes the transaction and concurrency patterns in this doc possible to write and test in isolation.

## Gin web layer, in depth

**Context lifecycle.** Gin pools `*gin.Context` objects to avoid an allocation per request. When your handler returns, that `Context` goes back into the pool and may be reused by the *next* request on a different goroutine almost immediately. This is the root cause behind the async-handler rule: if you hand `c` itself to a goroutine and the original request finishes first, your goroutine is now reading fields on a `Context` that belongs to someone else's request — a data race that won't show up in local testing but will under load. `c.Copy()` allocates a real, detached copy safe to hold past the handler's return.

**Middleware chain ordering.** Middleware in Gin is a stack of closures around `c.Next()`. Code before `c.Next()` runs on the way in, code after runs on the way out, in reverse order of registration. This is why order matters:

1. `Recovery()` first — it must wrap every other middleware so a panic anywhere downstream (including your own code) is caught and turned into a 500, not a crashed process.
2. Your structured logger next — so it can measure total latency (including auth, rate limiting, etc.) and log the final status code, which it only knows *after* `c.Next()` returns.
3. Auth / rate-limit guards after that, scoped to specific route groups, not globally, when they shouldn't apply to every route (e.g. `/register` shouldn't require an API key).

**Why `AbortWithStatusJSON` and not just `return`.** A bare `return` inside a middleware stops *that* function but Gin still calls the remaining handlers in the chain unless you explicitly abort. `c.Abort()` (or `AbortWithStatusJSON`) sets a flag Gin checks between each handler in the chain — skipping everything after it, including your actual route handler. Forgetting this is a classic bug: an auth middleware "rejects" a request with a 401 but the handler runs anyway because the write wasn't paired with an abort.

**Structured logging done right.** Log one JSON line per request with at minimum: method, path, status, latency, and a request ID (generate a UUID in the first middleware, inject it into `c.Request.Context()`, and thread it through to your DB layer so you can correlate a slow query with the request that caused it). This is non-negotiable at any real scale — grepping plain-text logs for a specific failing request is how you lose an afternoon.

**Request-scoped timeouts.** Wrap `c.Request.Context()` with `context.WithTimeout` in a middleware, replace `c.Request` with the derived request (`c.Request = c.Request.WithContext(ctx)`), and defer `cancel()`. Every downstream call — your DB queries included, if you pass the context through — now respects that deadline. Without this, a slow downstream (DB lock, external API) can hold a goroutine open indefinitely, and under enough concurrent slow requests that's how you exhaust your server's goroutines or your DB pool.

**Graceful shutdown.** Run Gin's engine inside an `http.Server` you control (not `r.Run()`, which blocks forever with no shutdown hook). On `SIGTERM`, call `server.Shutdown(ctx)` with a bounded deadline — it stops accepting new connections but lets in-flight ones finish. Skipping this means every deploy or pod restart silently drops whatever requests were mid-flight, which is a direct source of the kind of "random 502s during deploys" that's painful to debug after the fact.

## Connection pooling (pgxpool)

**Why pool at all.** Opening a TCP connection to Postgres, doing TLS/auth negotiation, and spinning up a backend process on the Postgres side costs single-digit milliseconds to tens of milliseconds — paid on *every* request if you don't reuse connections. A pool keeps a set of live connections warm and hands them out/back, turning that per-request cost into near-zero after warm-up.

**Key `pgxpool.Config` fields and how to reason about them:**

- `MaxConns`: the ceiling. Set this with Postgres's own `max_connections` in mind — if you run 5 service instances each with `MaxConns=20`, you need Postgres configured for 100+ connections plus headroom for admin/migration tools, or you'll see `too many clients` errors under load. A common mistake is sizing this per-service without accounting for how many service instances exist.
- `MinConns`: connections kept warm even when idle, so a traffic spike doesn't pay the connection-setup cost for the first N requests. Usually a small fraction of `MaxConns`.
- `MaxConnLifetime`: forces connections to be recycled periodically (e.g. 30-60 min) — useful for load-balanced Postgres setups (like read replicas behind a proxy) where you want connections to eventually redistribute, and as a safety net against subtle connection-state leaks.
- `MaxConnIdleTime`: closes connections that have been idle too long, so the pool shrinks back toward `MinConns` during quiet periods instead of holding `MaxConns` open forever.
- `HealthCheckPeriod`: how often pgx proactively pings idle connections to catch a dead connection before a real request hits it.

**Sizing guidance.** There's no universal number — it depends on your DB's CPU/connection overhead and your workload — but the common failure mode to avoid is setting `MaxConns` far higher than Postgres can handle well. Postgres connections aren't free even when idle (each is a backend process with its own memory). A frequently cited starting point is `connections ≈ (core_count × 2) + effective_spindle_count`, but the only way to know your real number is to load-test and watch `pg_stat_activity` and connection wait time.

**What happens when the pool is exhausted.** If every connection is checked out and a new request needs one, `pgxpool.Acquire` blocks until one frees up or the context deadline (you *are* passing a context with a timeout, per the Gin section above) expires. This is a feature, not a bug — it applies backpressure instead of letting unbounded goroutines pile up against the database. Watching "time spent waiting to acquire a connection" as a metric is one of the earliest signals that your pool is undersized for current load, well before you see outright errors.

## Schema design and indexing

**Why indexes matter here specifically.** Without an index, Postgres finds matching rows with a sequential scan — reading every row in the table. Fine at 1,000 rows, a measurable problem at 1,000,000. The default index type, B-tree, is right for the vast majority of cases: equality lookups, range queries, and sorting all use it efficiently.

**What to index in this schema:**

- Every foreign key column (`books.author_id`) — Postgres does **not** auto-index foreign keys (a common surprise coming from other databases). Without it, deleting an author forces a sequential scan of `books` to check for orphaned rows, and your `WHERE author_id = $1` queries do the same.
- Columns you filter or sort by in hot-path queries — if `ListAuthors` paginates with `ORDER BY name`, an index on `name` turns that sort into an index scan instead of sort-then-scan.
- **Composite indexes** when you consistently filter on two columns together (e.g. in the booking system below, `(event_id, status)` — "find available seats for this event" is the query that runs constantly). Column order in a composite index matters: Postgres can use a leading-column prefix of the index, so `(event_id, status)` serves `WHERE event_id = ?` alone but not `WHERE status = ?` alone.

\**The standard of "never SELECT *" connects directly to indexing**: if a query only needs `id, title`, selecting those explicitly lets Postgres potentially satisfy it from the index alone (an *index-only scan*, no heap lookup needed) if those columns are in the index. `SELECT *` forces a heap fetch every time.

**`EXPLAIN ANALYZE` is how you stop guessing.** Run it on any query you're unsure about: `EXPLAIN ANALYZE SELECT ...`. Look for `Seq Scan` on a table you expected to hit an index — that's your signal either the index is missing, or Postgres's query planner decided the table is small enough that a sequential scan is actually cheaper (which it genuinely can be on small tables — don't index reflexively).

**Foreign keys and cascade behavior.** Decide `ON DELETE` behavior explicitly per relationship instead of leaving the default (`NO ACTION`, which just errors if orphans would result). For `books.author_id → authors.id`, `ON DELETE CASCADE` silently deletes all of an author's books when the author is deleted — convenient, but dangerous if triggered accidentally; `ON DELETE RESTRICT` forces you to handle orphaned rows explicitly before deleting. For the booking system later in this doc, you'll almost always want `RESTRICT` on `bookings → events` — you never want a careless event deletion to silently wipe booking history.

## sqlc query patterns

**How sqlc actually works.** It parses your `schema.sql` and `query.sql`, type-checks the queries against the schema, and generates a `Queries` struct with one typed Go method per named query (`-- name: GetAuthor :one`). The generated code has zero reflection and zero runtime SQL building — what you wrote is what runs.

**The `WithTx` pattern, concretely:**

```go
tx, err := pool.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx) // no-op if Commit already ran

qtx := queries.WithTx(tx)
author, err := qtx.CreateAuthor(ctx, ...)
if err != nil { return err } // deferred Rollback fires

_, err = qtx.CreateBook(ctx, ...)
if err != nil { return err } // deferred Rollback fires

return tx.Commit(ctx)
```

The `defer tx.Rollback(ctx)` immediately after `Begin` is the pattern to memorize: it's always safe to call — a `Rollback` after a successful `Commit` is a documented no-op — and it guarantees that *any* early return, including ones you add later while editing this function, rolls back instead of leaving a dangling open transaction.

**Batch inserts with `UNNEST`.** A loop of individual `INSERT`s inside one transaction is correct but pays a network round-trip per row. `UNNEST` turns an array parameter into a set of rows in one statement:

```sql
-- name: CreateBooksBatch :many
INSERT INTO books (author_id, title)
SELECT $1, unnest($2::text[])
RETURNING *;
```

One call, one round-trip, N rows inserted — this is the pattern to reach for whenever you're inserting a batch that's known up front, not just in this assignment.

**Common gotchas:**

- **Nullable columns** generate `pgtype.Text`/`sql.NullString`-style wrapper types, not plain `string`. Forgetting to check `.Valid` before reading `.String` is a frequent source of bugs — decide early whether a column should genuinely be nullable or just default to an empty string, because the Go-side ergonomics differ a lot.
- **`:one` vs `:many` vs `:exec`** — a `:one` query that can legitimately return zero rows (e.g. `GetAuthor` for a nonexistent id) returns `pgx.ErrNoRows`, which you must check for explicitly and translate into your own "not found" error — it is not a panic or a nil struct.
- **Regenerate after every schema change** and let the compiler find every call site that needs updating, rather than hand-patching generated code (never edit generated files directly — they're overwritten on the next `sqlc generate`).

## Transactions, in depth

This is the section that matters most for the booking-system case study next, so it's worth being precise.

**ACID, in terms of what actually breaks without them:**

- **Atomicity** — the author+book insert either both happen or neither does. Without it, a crash mid-write leaves a book pointing at an author that doesn't exist.
- **Consistency** — the database never violates its own constraints (foreign keys, unique constraints, check constraints), even mid-transaction from another session's point of view.
- **Isolation** — concurrent transactions don't see each other's uncommitted changes. This is the one with the most nuance (below).
- **Durability** — once committed, a transaction survives a crash. Postgres guarantees this via write-ahead logging (WAL).

**Isolation levels — what each one actually permits:**

- **Read Committed** (Postgres's default): a transaction only sees data committed *before each individual statement* started. Two `SELECT`s in the same transaction can see different data if another transaction committed in between — this is the subtle trap: Read Committed does **not** give you a consistent snapshot across your whole transaction, only per-statement.
- **Repeatable Read**: the entire transaction sees a single consistent snapshot taken at its start. Solves the above, but introduces **serialization failures** — if two Repeatable Read transactions both try to update the same row, one will fail with a `could not serialize access` error and must be retried by your application code.
- **Serializable**: the strictest — guarantees the result is *as if* transactions ran one at a time, even if they ran concurrently. Postgres implements this with extra bookkeeping (predicate locking) and a higher chance of serialization failures to retry. Reach for this only when Repeatable Read's guarantees genuinely aren't enough — it has real throughput cost under contention.

**Row locking — the tool you'll use constantly in the booking system:**

- `SELECT ... FOR UPDATE`: locks the selected rows until the transaction commits or rolls back. Any other transaction trying to `SELECT ... FOR UPDATE` the same row **blocks** until the first one finishes. This is how you prevent two requests from both reading "seat is available" before either has committed a booking.
- `SELECT ... FOR UPDATE SKIP LOCKED`: same locking, but instead of blocking, immediately skips any row already locked by another transaction. This is the pattern for "grab *any* available seat from a pool" — competing requests each grab a *different* free row instead of queueing up behind each other, which matters enormously for throughput under high concurrency (detailed in the case study below).

**Deadlocks.** If transaction A locks row 1 then waits for row 2, while transaction B locks row 2 then waits for row 1, neither can proceed. Postgres detects this automatically and kills one transaction with a `deadlock detected` error so the other can continue — but *your code* must catch that error and retry, it isn't resolved for you. The practical defense is simple and important: **always acquire locks in a consistent order** across your whole codebase (e.g. always lock the lower `id` first) — most real-world deadlocks are eliminated by this one discipline rather than by retry logic.

**When rollback happens.** Beyond an explicit `tx.Rollback()`: any statement that violates a constraint (unique, foreign key, check) aborts the *entire* transaction, not just that statement — every subsequent statement in that transaction will fail with `current transaction is aborted` until you roll back. This surprises people coming from databases with more forgiving semantics; in Postgres, catch the error and roll back immediately rather than trying to "continue" the transaction.

## Case study: ticket booking under concurrent users

**The core problem.** 500 users click "Book" on the same seat within the same second. Without protection, the naive flow (SELECT status, check it's available in application code, then UPDATE status) has a race window between the read and the write: two requests can both read "available" before either writes "booked," and both proceed to confirm the same seat. This is the overselling bug, and it's the single most important failure mode this whole module exists to prevent.

| Strategy | How it works | Trade-offs | Best for |
| --- | --- | --- | --- |
| Pessimistic locking (SELECT FOR UPDATE) | Lock the seat row the instant you read it; a second request blocks until the first transaction commits or rolls back, then sees the updated status and fails cleanly | Simple to reason about; under heavy contention on one row, requests queue and latency climbs | Booking a specific named seat where contention is on one exact row |
| SELECT FOR UPDATE SKIP LOCKED | Query a pool of candidate rows (any available seat in a section); each transaction grabs the first row not already locked, skipping locked ones instead of waiting | Much higher throughput for "any available seat" booking; doesn't help if the user wants one specific seat | General-admission booking at high concurrency |
| Optimistic locking (version column) | Add a version column; read it, then update WHERE id = ? AND version = ?; zero rows updated means someone else won, so retry or fail | No locks held during think-time between read and write; wasted work on conflict, needs explicit retry logic | Lower-contention resources, or flows with a review-before-confirm step |
| Unique DB constraint as last line of defense | A unique constraint on seat\_id in the bookings table means the database physically cannot store two confirmed bookings for the same seat, even if app logic has a bug | Doesn't prevent the race, only guarantees you catch it | Always, regardless of which locking strategy above you use, as a correctness backstop |
| Idempotency keys for retries | Client sends a unique key with the request; server checks if that key was already processed before doing any work | Protects against the client's own retry causing a double booking, a different problem from concurrent different users | Any booking or payment flow where the client might retry a request that actually succeeded |
| Queue-based serialization | Push booking requests for one event into a queue processed one at a time for that event | Eliminates the race by removing concurrency at the point of contention; adds latency and infra complexity | Extremely high-demand single events, flash sales |
| Distributed lock (Redis SET NX) | Acquire a short-lived lock keyed by seat or event ID in Redis before touching Postgres | Keeps lock contention out of DB connections; adds a new failure mode around lock expiry and Redis availability | Multi-service architectures that want to fail fast before acquiring a DB connection |

**What this module's assignments should add, concretely, to actually exercise this:** build the schema with a seats table (status enum: available, held, booked) and a unique-constrained bookings table; implement the booking handler three ways (plain FOR UPDATE, SKIP LOCKED against a pool, optimistic with a version column); then load-test each with concurrent requests (the hey tool, or a small Go script firing 100 goroutines at one seat) and observe the difference in response time and error patterns directly, instead of trusting it in theory. A hold state (seat reserved for 5-10 minutes during checkout, released if not confirmed) is also worth building; it's how real booking systems handle the gap between starting checkout and actually paying, implemented as a held\_until timestamp plus a background job that releases expired holds back to available.

## Caching layer

**Decide what to cache by how fast it goes stale, not by what's slow.** Seat-level availability for a specific seat changes the instant it's booked — caching it risks showing a seat as available after it's gone, which is worse than a slightly slower query. A seat *map* (the static layout of an event: sections, rows, seat numbers) rarely changes and is an excellent caching candidate. The rule of thumb: cache read-heavy, write-rarely data; never cache the exact field your concurrency-control logic depends on for correctness.

**Cache-aside pattern.** On read: check cache, on miss read from Postgres and populate the cache, return the value. On write: update Postgres, then invalidate (delete) the cache key rather than trying to update it in place — invalidation is simpler to get right than keeping a cached value in sync with every possible write path.

**Avoiding cache stampede.** When a popular cache key expires, many concurrent requests can simultaneously miss and all hammer Postgres at once to repopulate it — a self-inflicted spike at exactly the moment you were trying to protect the DB. Mitigate with a short-lived lock (Redis `SET NX` again) around the recompute step so only one request repopulates the cache while others wait briefly or serve a slightly stale value.

**Redis doing double duty.** In this stack, Redis is a reasonable home for both the cache *and* the distributed lock from the case study above — just keep the two uses conceptually separate (different key prefixes, different TTLs) so a caching bug doesn't accidentally corrupt your locking logic or vice versa.

**What not to cache in a booking system:** real-time seat status, a user's current booking-in-progress state, or anything a `SELECT FOR UPDATE` depends on being current. Caching here is for reducing load on genuinely slow, rarely-changing reads (event details, venue layout, pricing tiers) — not for the hot path where correctness under concurrency is the whole point of this module.

## Testing and observability

**Why `testcontainers-go` over a manually managed Docker container.** A test that assumes "Postgres is already running on localhost:5432" works on your machine and breaks in CI, or silently reuses stale state between runs. `testcontainers-go` starts a real, disposable Postgres container per test run, applies migrations against it, and tears it down after — the test is fully self-contained and reproducible anywhere, including a teammate's machine or a CI pipeline with nothing pre-provisioned.

**What to actually test with it, beyond basic CRUD:**

- The transaction rollback case from Module 5 (force a failure mid-transaction, assert *both* tables are unaffected).
- The concurrency strategies from the case study: spin up N goroutines hitting the same seat concurrently, assert exactly one booking succeeds and the rest get a clean, expected error (not a panic, not a partial write).
- Idempotency: send the same booking request twice with the same idempotency key, assert only one booking exists.

**Metrics worth exposing (Prometheus-style) once this is running as a real service:**

- **Pool saturation** — connections in use vs. `MaxConns`, and time spent waiting to acquire one. This is your earliest warning that the DB layer is becoming a bottleneck, well before user-facing errors appear.
- **Lock wait time** on booking transactions specifically — a rising trend here, even without outright failures, tells you contention is building (e.g. a popular event about to sell out) before it becomes a problem.
- **Request latency and status code counts**, broken down by route — the structured logging from the Gin section gives you the raw data; aggregate it into histograms rather than relying on grepping logs after the fact.
- **Deadlock and serialization-failure counts** — these should be rare; a sudden rise means your lock-ordering discipline broke somewhere or load increased past what your current strategy handles well.

## Capstone: concurrent-safe ticket booking API

One project that exercises every section above, sequenced so each piece builds on the last.

**Schema** — `events` (id, name, venue, starts\_at), `seats` (id, event\_id FK, label, status enum: available/held/booked, version int), `bookings` (id, seat\_id FK unique, user\_id, idempotency\_key unique, created\_at). Index `(event_id, status)` on seats; unique constraints on `bookings.seat_id` and `bookings.idempotency_key`.

**API surface (Gin)**

- `POST /events/:id/hold` — body: seat preference (specific seat\_id, or "any" for general admission). Holds a seat for 5 minutes using `SELECT FOR UPDATE` (specific seat) or `SKIP LOCKED` (any-seat pool). Returns a hold token.
- `POST /bookings` — body: hold token, idempotency key. Confirms the booking inside a transaction, checking the idempotency key first. Returns 409 cleanly on conflict (seat already booked, hold expired).
- `GET /events/:id/availability` — cached seat-map-level read (not per-seat status) per the caching section.
- `/admin/events` group, API-key guarded, for creating events and seat inventory.

**Background job** — releases expired holds (`held_until < now()`) back to `available`, running on a ticker.

**Build order and exit criteria:**

1. Schema + migrations + sqlc queries, basic CRUD working — exit: `go test` passes CRUD against a real `testcontainers-go` Postgres.
2. Gin handlers wired to real queries, no concurrency protection yet — exit: happy path works end to end via curl.
3. Add `SELECT FOR UPDATE` to the hold endpoint, add the unique constraint backstop — exit: a hand-written load test (100 concurrent goroutines booking the same seat) shows exactly one success, 99 clean 409s, zero panics.
4. Swap in `SKIP LOCKED` for the any-seat pool path, compare load-test latency against step 3's plain `FOR UPDATE` on a 50-seat pool — exit: you can explain, with your own numbers, why one is faster under this load shape.
5. Add idempotency keys, structured logging, graceful shutdown, request timeouts — exit: killing the service mid-request during a load test doesn't leave inconsistent state, and a retried request with the same idempotency key never double-books.
6. Add the cache-aside layer for `GET /events/:id/availability`, with stampede protection — exit: cache invalidates correctly on every booking, verified by a test that books a seat and immediately checks the availability endpoint reflects it.
7. Metrics + a 10-minute load test against the full system, screenshot the pool-saturation and lock-wait-time graphs — exit: you can point at the graph and explain what's happening physically underneath each line.

By the end of step 7 you have a project that directly answers "tell me about a time you handled a concurrency problem" with real numbers, not a description of the theory.

# Escape Analysis and Benchmark Guide (Go)

This guide explains:
1. How to run escape analysis
2. How to run benchmarks with memory stats
3. How to read the output

## 1) Escape Analysis Commands

Run in pointers_learning:

go test . -gcflags="-m"

For more detailed reasoning:

go test . -gcflags="-m -m"

To focus only on your escape example test:

go test . -run '^TestHeapEscapeValue$' -gcflags="-m -m"

### How to read escape output

Common lines:
1. moved to heap: x
Meaning: variable x is heap-allocated.

2. x does not escape
Meaning: variable stays on stack.

3. can inline FunctionName
Meaning: compiler may inline this function (important for benchmark interpretation).

In your example files:
- escape_example.go
- escape_example_test.go

HeapEscape returns the address of a local variable, so you should see moved to heap for that local variable.

## 2) Benchmark Commands

Run only benchmarks (skip normal tests):

go test . -run=^$ -bench=. -benchmem

Run only the two escape benchmarks:

go test . -run=^$ -bench='Benchmark(StackOnly|HeapEscape)$' -benchmem

Optional: run longer for more stable numbers:

go test . -run=^$ -bench='Benchmark(StackOnly|HeapEscape)$' -benchmem -benchtime=3s

## 3) How to read benchmark output

Example shape:

BenchmarkStackOnly-2      1000000000    0.65 ns/op    0 B/op    0 allocs/op
BenchmarkHeapEscape-2       87689420   13.7 ns/op     8 B/op    1 allocs/op

Meaning of columns:
1. Benchmark name
2. Iteration count chosen by Go benchmark runner
3. ns/op: average nanoseconds per operation (lower is faster)
4. B/op: bytes allocated per operation
5. allocs/op: heap allocations per operation

What to conclude:
1. StackOnly with 0 B/op and 0 allocs/op means no heap allocation.
2. HeapEscape with 8 B/op and 1 allocs/op means one heap allocation each call.
3. This matches escape analysis output showing moved to heap.

## 4) Why both tools are needed

1. Escape analysis tells what the compiler decided.
2. benchmem shows runtime allocation cost per operation.
3. Together they give complete proof: cause plus measured impact.

If you want, I can next provide a second example showing interface-based escaping and how it changes both escape output and benchmem numbers.

## 5) inline
Inline means the compiler may replace a function call with the function body directly at the call site.

So instead of:
- call function
- jump to another place
- return back

the compiler may just paste the function logic into the caller. That removes call overhead and can make the code faster.

In the Go report you will see things like:
- can inline StackOnly
- inlining call to StackOnly
- cannot inline ... cost exceeds budget

What it means:
- can inline: the function is small/simple enough that Go may inline it
- inlining call to X: the compiler actually did inline it at that call site
- cannot inline: the function is too complex for the compiler to inline

Important: you do not write inline as normal code to make this happen. It is a compiler optimization. The thing you can write in code is a directive called noinline to stop inlining, but that is different from inline itself.

Why it matters for your learning:
- If a function gets inlined, benchmark results may look faster because function call overhead disappears.
- Inlining can also change escape analysis, because once the function body is inside the caller, the compiler can often reason more precisely about whether values escape to the heap.

A simple way to think about it:
- inline = compiler optimization
- escape analysis = compiler deciding stack vs heap
- benchmem = runtime measurement of allocations

If you want, I can add a short section to escape_analysis_doc.md explaining inline and noinline in beginner language.
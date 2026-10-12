# ==============================================================================
# Go Zero to Hero: Performance, Benchmarking & Race CI Makefile
# ==============================================================================

.PHONY: help test test-race bench profile-cpu profile-mem pprof-cpu-cli pprof-mem-cli pprof-cpu-web pprof-mem-web clean ci

help: ## Display available targets
	@echo "Available Makefile commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

test: ## Run unit tests across all packages
	go test -v ./sync/... ./pointers_learning/... ./structs_composition_1/...

test-race: ## Run tests with the Go race detector enabled (-race)
	@echo "==> Running tests with Race Detector (-race)..."
	go test -race ./sync/... ./pointers_learning/... ./structs_composition_1/...

bench: ## Run all benchmarks with memory allocation metrics
	@echo "==> Running benchmarks..."
	go test -run=^$$ -bench=. -benchmem ./benchmark/...

profile-cpu: ## Profile a slow CPU function and generate cpu.pprof
	@echo "==> Generating CPU profile (cpu.pprof)..."
	go test -run=^$$ -bench=BenchmarkCPUSlowFunction -cpuprofile=cpu.pprof ./benchmark/...
	@echo "Generated cpu.pprof. Inspect with 'make pprof-cpu-cli' or 'make pprof-cpu-web'"

profile-mem: ## Profile memory allocations and generate mem.pprof
	@echo "==> Generating Memory profile (mem.pprof)..."
	go test -run=^$$ -bench=BenchmarkMemory -benchmem -memprofile=mem.pprof ./benchmark/...
	@echo "Generated mem.pprof. Inspect with 'make pprof-mem-cli' or 'make pprof-mem-web'"

pprof-cpu-cli: ## Open interactive CLI for CPU profile
	go tool pprof cpu.pprof

pprof-mem-cli: ## Open interactive CLI for Memory profile (alloc_space mode)
	go tool pprof -alloc_space mem.pprof

pprof-cpu-web: ## Launch browser UI visualization on port 8080 for CPU profile
	@echo "Starting pprof web server on http://localhost:8080 ..."
	go tool pprof -http=:8080 cpu.pprof

pprof-mem-web: ## Launch browser UI visualization on port 8080 for Memory profile
	@echo "Starting pprof web server on http://localhost:8080 ..."
	go tool pprof -http=:8080 mem.pprof

clean: ## Remove generated profiling artifacts and test binaries
	rm -f cpu.pprof mem.pprof benchmark.test

ci: test-race bench ## Continuous Integration (CI) target: runs race detector and benchmarks
	@echo "==> CI pipeline checks passed successfully!"

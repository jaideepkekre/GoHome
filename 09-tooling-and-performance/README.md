# Tooling & Performance — At a Glance

> Quick-reference for profiling, race detection, and benchmarking.
> For `cgo` overhead and `pprof` deep dive → [Advanced](advanced.md)

---

## `pprof` — Where to Start

| Problem | Profile | Endpoint |
|---------|---------|----------|
| "My program is slow" | CPU | `/debug/pprof/profile` |
| "Too much memory" | Heap | `/debug/pprof/heap` |
| "It's stuck / goroutine leak" | Goroutine | `/debug/pprof/goroutine` |
| "Mutex contention" | Mutex | `/debug/pprof/mutex` |
| "Blocking on channels" | Block | `/debug/pprof/block` |

### Setup

```go
import _ "net/http/pprof"
go http.ListenAndServe("localhost:6060", nil)
```

```bash
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof -http=:8080 profile.pb.gz   # flame graph web UI
```

### Heap Profile Modes

| Mode | Shows |
|------|-------|
| `inuse_space` | What's live NOW (memory leaks) |
| `alloc_space` | Total allocated since start (allocation rate) |

---

## Race Detector

```bash
go test -race ./...
go run -race main.go
```

| Aspect | Detail |
|--------|--------|
| Detects | Actual data races that occur during execution |
| Misses | Races in untested code paths |
| Cost | 2-20x slower, 5-10x memory |
| Production? | Not recommended (some teams canary with `-race`) |

---

## Benchmarking

```go
func BenchmarkMyFunc(b *testing.B) {
    for i := 0; i < b.N; i++ {
        myFunc()
    }
}
```

```bash
go test -bench=. -benchmem ./...
go test -bench=BenchmarkMyFunc -count=5 -benchtime=3s
```

`-benchmem` shows allocations per operation. Use `benchstat` to compare runs.

---

## `GOMAXPROCS`

Sets the number of **Ps** (logical processors). Default: number of logical CPUs. Does NOT limit OS threads (blocked syscalls can create more Ms).

```
  GOMAXPROCS = 1  → single P, useful for deterministic race repro
  GOMAXPROCS = N  → N Ps, N goroutines run simultaneously
  GOMAXPROCS > CPUs → rarely beneficial, wastes per-P cache memory
```

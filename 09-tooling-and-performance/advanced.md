# Tooling & Performance — Advanced

> `cgo` internals, race detector details, and `pprof` deep dive.
> Back to → [At a Glance](README.md)

---

## `cgo` — The Cost of Crossing

```
  Go function call:    ~2 ns
  cgo call:            ~50-200 ns (25-100x overhead)

  What happens on each cgo call:
  ┌──────────────────────────────────────────────────────────┐
  │  1. Save Go stack state                                  │
  │  2. Switch to C-compatible stack                         │
  │     (goroutine stacks are too small and may move)        │
  │  3. Mark goroutine as "in syscall"                       │
  │     (P can be handed to another M)                       │
  │  4. Call C function                                      │
  │  5. Restore Go state                                     │
  │  6. Reacquire a P (may go to global queue if none free)  │
  └──────────────────────────────────────────────────────────┘
```

### Mitigation Strategies

- **Batch work** across the boundary (pass a slice, not individual items)
- **Minimize crossing frequency** (accumulate work in Go, cross once)
- **Rewrite critical C code in Go** if feasible

### Other `cgo` Costs

| Cost | Impact |
|------|--------|
| C memory invisible to GC | Manual management required |
| Slower builds | C compiler + linking step |
| Cross-compilation | Breaks easily |
| Race detector blind to C code | Races in C are undetected |
| C can corrupt Go memory | Mystery crashes |

---

## Race Detector Internals

The race detector is based on **ThreadSanitizer v2** (TSan). It instruments every memory access and synchronization operation at compile time.

```
  Instrumented code records:
  ┌────────────────────────────────────────┐
  │  For each memory access:               │
  │  • goroutine ID                        │
  │  • memory address                      │
  │  • read or write                       │
  │  • logical timestamp (vector clock)    │
  └────────────────────────────────────────┘

  Detection:
  Two accesses to same address
  + at least one is a write
  + vector clocks are NOT ordered
  = DATA RACE REPORTED
```

### What Crashes Without the Detector

| Race Type | Behavior |
|-----------|---------|
| Concurrent map writes | Fatal crash (`recover` can't catch) |
| Interface assignment race | Possible crash (mismatched type/value) |
| Slice header race | Possible crash or corruption |
| Other data races | Silent corruption until observed |

---

## `pprof` Deep Dive

### Heap Profile: `alloc_space` vs `inuse_space`

```
  alloc_space:                    inuse_space:
  ┌─────────────────────┐        ┌─────────────────────┐
  │ Everything ever      │        │ What's live NOW      │
  │ allocated (incl.     │        │ (currently on heap)  │
  │ already freed)       │        │                      │
  └─────────────────────┘        └─────────────────────┘
  Use for: allocation             Use for: memory leaks
  rate hotspots                   and current usage
```

### Mutex Profile

Must be enabled explicitly:

```go
runtime.SetMutexProfileFraction(5)  // sample 1/5 of contention events
```

Shows which mutexes have the most contention (goroutines waiting to acquire).

### Block Profile

```go
runtime.SetBlockProfileRate(1000)  // sample blocking events lasting > 1μs
```

Shows where goroutines block: channels, select, mutex, etc.

---

## Build Tags

```go
//go:build linux && amd64
```

| Suffix | Implicit Tag |
|--------|-------------|
| `_linux.go` | `linux` |
| `_amd64.go` | `amd64` |
| `_test.go` | Test-only |

---

## `replace` Directive Dangers

```
replace github.com/foo/bar => ../local-bar
```

| Danger | Why |
|--------|-----|
| CI/CD doesn't have local path | Build fails |
| Replacement diverges from original | Silent bugs |
| Easy to forget it's there | Stale overrides |
| Non-main modules | `replace` is **ignored** by consumers |

**Always grep for `replace` in `go.mod` when debugging.**

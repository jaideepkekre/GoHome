# Memory & GC — At a Glance

> Quick-reference for the memory model, escape analysis, and GC tuning.
> For tri-color marking, write barrier details, and alignment → [Advanced](advanced.md)

---

## Happens-Before (What You Need to Know)

If there's no happens-before edge between a write and a read, the read may see **stale data, partial writes, or nothing at all**.

| Operation | Guarantee |
|-----------|----------|
| Package `init()` | All complete before `main()` |
| `go f()` | Values visible before `go` are safe in `f()` |
| `mu.Unlock()` | Happens-before next `mu.Lock()` |
| Channel send | Happens-before receive completes |
| `wg.Done()` | Happens-before `wg.Wait()` returns |
| `once.Do(f)` | `f` completes before any `Do()` returns |
| `atomic` ops | Total order on single variable |

---

## Data Races

Two goroutines access the same memory + at least one writes + no synchronization = **undefined behavior**.

```
  Goroutine A:  x = 42       Goroutine B:  print(x)
                    │                           │
                    └─── no happens-before ──────┘
                         DATA RACE
```

x86 has strong memory ordering (TSO) — races hide on Intel. They surface on ARM64, with more cores, or with compiler changes. Always run tests with `-race`.

---

## Escape Analysis

```
Stack allocation:  fast (~1 ns, freed with function frame)
Heap allocation:   slow (~25 ns + GC tracking)
```

Variables "escape" when the compiler can't prove no reference outlives the stack frame.

### Common Escape Triggers

- Returning a pointer to a local variable
- Storing in an interface
- Sending a pointer over a channel
- Closures capturing variables
- Slices growing beyond compiler prediction

### Inspect Decisions

```bash
go build -gcflags="-m" ./...       # basic
go build -gcflags="-m -m" ./...    # detailed
```

---

## GC Tuning

| Setting | What It Does | Default |
|---------|-------------|---------|
| `GOGC` | Trigger GC when heap grows by this % | 100 |
| `GOMEMLIMIT` (Go 1.19+) | Soft memory cap — GC works harder to stay under | No limit |

```
 Lower GOGC → more GC → less memory, more CPU
Higher GOGC → less GC → more memory, less CPU
```

`GOMEMLIMIT` is especially useful in containers with fixed memory budgets.

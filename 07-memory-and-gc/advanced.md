# Memory & GC — Advanced

> Tri-color GC, write barrier, memory alignment, and false sharing.
> Back to → [At a Glance](README.md)

---

## Tri-Color Mark-and-Sweep

Go's GC is **concurrent** — it runs alongside your application, with two brief stop-the-world pauses.

```
Object states during marking:

  ●  WHITE  — not visited yet. At end of marking → garbage.
  ◐  GRAY   — visited, children NOT yet scanned. In work queue.
  ●  BLACK  — visited, all children scanned. Reachable / alive.


Phase 1:  Mark roots as GRAY
          ┌─────┐
          │roots│ → goroutine stacks, globals
          └──┬──┘
             │ mark gray
             ▼
          ◐ ◐ ◐  (gray set)

Phase 2:  Process gray set (CONCURRENT with application)
          For each gray object:
            • Mark it BLACK
            • Mark all objects it points to GRAY
          Repeat until gray set is empty.

          ● ● ●  (black = alive)
          ○ ○ ○  (white = garbage → sweep)

Phase 3:  Sweep white objects (concurrent, lazy)
```

### Stop-the-World Phases

```
  Application running ─────────────────────────────────────────►
                         │                        │
                      STW 1                    STW 2
                     (< 100μs)               (< 100μs)
                         │                        │
                    Enable write           Drain remaining
                    barrier, flip          work, disable
                    GC phase               write barrier
```

---

## The Write Barrier

During concurrent marking, the mutator (your code) might move a live pointer from a black object to a white object. Without protection, the GC would miss the white object → use-after-free.

```
  Without write barrier:
  ● Black obj ──ptr──► ○ White obj     (GC thinks white is garbage!)
                         ▲
                  pointer moved here
                  from a gray object

  With write barrier:
  Every pointer write notifies the GC:
    "I'm writing a pointer here — shade the target GRAY"

  Cost: a few extra instructions per pointer write.
  Only active during GC marking phase.
```

---

## Escape Analysis Details

### `-gcflags` Output Explained

```bash
go build -gcflags="-m -m" ./...
```

```
./main.go:20:10: &Bar{} escapes to heap
  ↳ The address of Bar is taken and stored somewhere
    the compiler can't prove stays on the stack.

./main.go:25:10: &Baz{} does not escape
  ↳ Compiler proved the pointer doesn't outlive the function.

./main.go:30:15: leaking param: x
  ↳ Parameter x (or something derived from it) escapes
    through the return value or is stored externally.
```

### Interface Boxing Escape

```go
func log(v any) { /* ... */ }  // v escapes because interface stores on heap

// In a tight loop, this causes allocation on every iteration:
for _, item := range items {
    log(item)  // item escapes to heap for interface boxing
}
```

---

## Memory Alignment

Go lays out struct fields in **declaration order** with padding for alignment:

```
Bad layout (24 bytes with padding):
  Offset 0:  bool   (1 byte)
  Offset 1:  ████   (7 bytes padding)
  Offset 8:  int64  (8 bytes)
  Offset 16: bool   (1 byte)
  Offset 17: ████   (7 bytes padding)
  Total: 24 bytes

Good layout (16 bytes):
  Offset 0:  int64  (8 bytes)
  Offset 8:  bool   (1 byte)
  Offset 9:  bool   (1 byte)
  Offset 10: ████   (6 bytes padding)
  Total: 16 bytes
```

**Rule:** Order fields from largest to smallest alignment.

---

## False Sharing

When two goroutines write to different fields on the **same CPU cache line** (64 bytes), the line bounces between cores — destroying parallelism.

```
  Cache line (64 bytes):
  ┌──────────────────────────────────────────────────────────────────┐
  │ counter_A (8 bytes)  │  counter_B (8 bytes)  │  ...padding...   │
  └──────────────────────────────────────────────────────────────────┘
        ▲                          ▲
   Goroutine 1 writes        Goroutine 2 writes
        │                          │
        └── Cache line BOUNCES ────┘
            between CPU cores
            (destroys parallelism)

  Fix: Pad to separate cache lines:
  ┌────────────────────────────────────────────────────────────────┐
  │ counter_A (8 bytes)  │  [56 bytes padding]                    │
  ├────────────────────────────────────────────────────────────────┤
  │ counter_B (8 bytes)  │  [56 bytes padding]                    │
  └────────────────────────────────────────────────────────────────┘
```

```go
type Counters struct {
    read  atomic.Int64
    _     [56]byte         // pad to separate cache line
    write atomic.Int64
}
```

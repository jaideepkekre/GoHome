# Concurrency — At a Glance

> Quick-reference for goroutines, channels, `select`, `context`, and sync primitives.
> For scheduler internals, `hchan` structs, and `sync.Once`/`sync.Pool` deep dives → [Advanced](advanced.md)

---

## Goroutines

Lightweight, user-space threads. Start with ~2-8 KB stack, grow dynamically.

```go
go myFunction()
go func() { /* ... */ }()
```

### What Blocks a Goroutine?

- Channel operations
- Network I/O
- `time.Sleep` / timers
- `sync.Mutex` / `sync.WaitGroup` / etc.

If a goroutine doesn't block, post-Go 1.14 it gets preempted via OS signals. Pre-1.14, it could starve its thread.

---

## Channels

```go
ch := make(chan int)      // unbuffered
ch := make(chan int, 10)  // buffered, cap 10
```

### Unbuffered vs Buffered

```
Unbuffered (synchronous handoff):
  Sender ──────────► Receiver
         blocks until      blocks until
         receiver ready    sender ready

Buffered (async, up to capacity):
  Sender ──► [  |  |  |  ] ──► Receiver
             └── buffer ──┘
             blocks when full   blocks when empty
```

### Channel Operations Cheat Sheet

| Operation | Nil Channel | Closed Channel | Open Channel |
|-----------|------------|---------------|-------------|
| Send | **Blocks forever** | **Panic** | Sends or blocks |
| Receive | **Blocks forever** | Returns zero value | Receives or blocks |
| Close | **Panic** | **Panic** | Closes |

### Closing Rules

- Only the **sender** should close
- Closing twice → panic
- Use `v, ok := <-ch` — `ok == false` means closed and drained
- `range ch` terminates when closed

---

## `select`

Waits on multiple channel operations. Multiple ready → **pseudo-random** pick (not source order).

```go
select {
case v := <-ch1:       // received
case ch2 <- val:       // sent
case <-time.After(5*time.Second):  // timeout
default:               // non-blocking (if present)
}
```

### Nil Channel in Select = Disabled Case

```go
var ch <-chan int = nil
select {
case v := <-ch:   // ← skipped (nil channel)
case <-done:      // ← still active
}
```

---

## Context

Carries cancellation, deadlines, and request-scoped values.

```go
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()  // ALWAYS call cancel

select {
case result := <-doWork(ctx):
    return result
case <-ctx.Done():
    return ctx.Err()  // Canceled or DeadlineExceeded
}
```

Cancellation propagates **top-down** (parent → children). Never bottom-up.

---

## sync Primitives — Quick Reference

| Primitive | Use When |
|-----------|---------|
| `sync.Mutex` | Protect shared state (read-modify-write) |
| `sync.RWMutex` | Reads >> writes |
| `sync.WaitGroup` | Wait for N goroutines to finish |
| `sync.Once` | One-time initialization |
| `sync.Pool` | Reuse expensive allocations |
| `atomic.Int64` / `atomic.Pointer[T]` | Lock-free read-heavy counters/config |

### `sync.WaitGroup` Pattern

```go
var wg sync.WaitGroup
for i := 0; i < 5; i++ {
    wg.Add(1)              // MUST happen before `go`
    go func() {
        defer wg.Done()
        // work
    }()
}
wg.Wait()
```

---

## Concurrency Patterns

| Pattern | What It Does |
|---------|-------------|
| **Fan-in** | N producers → 1 channel |
| **Fan-out** | 1 producer → N workers |
| **Pipeline** | Stages chained by channels |
| **Done channel** | Signal stop via `close(ch)` |
| **Worker pool** | Fixed goroutines + job channel |
| **Rate limiting** | `time.Tick` or `x/time/rate` |

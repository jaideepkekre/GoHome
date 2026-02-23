# Concurrency — Advanced

> Scheduler internals, channel implementation, and sync primitive deep dives.
> Back to → [At a Glance](README.md)

---

## The G, M, P Scheduler

Go uses an **M:N scheduler** — M goroutines on N OS threads.

```
                ┌─────────────────────────────────────────┐
                │           Go Runtime Scheduler          │
                └─────────────────────────────────────────┘

  ┌──────┐   ┌──────┐   ┌──────┐
  │  P0  │   │  P1  │   │  P2  │    Logical processors (GOMAXPROCS)
  │      │   │      │   │      │
  │ LRQ: │   │ LRQ: │   │ LRQ: │    LRQ = Local Run Queue
  │[G G G]│   │[G G] │   │[G]  │    (lock-free ring buffer, max 256)
  └──┬───┘   └──┬───┘   └──┬───┘
     │          │          │
     ▼          ▼          ▼
  ┌──────┐   ┌──────┐   ┌──────┐
  │  M0  │   │  M1  │   │  M2  │    OS Threads
  └──────┘   └──────┘   └──────┘

  ┌────────────────────────────────────────┐
  │  Global Run Queue (mutex-protected)    │
  │  [G G G G G ...]                       │    Overflow + work stealing source
  └────────────────────────────────────────┘
```

### Key Rules

| Entity | What It Is | Count |
|--------|-----------|-------|
| **G** | Goroutine (stack + instruction pointer + state) | Unlimited |
| **M** | OS thread | Dynamic (grows when Gs block in syscalls) |
| **P** | Scheduling context + local queue + caches | `GOMAXPROCS` (default: CPU count) |

**A G needs a P to run. A P needs an M to execute.**

### Work Stealing

```
P0 (empty) looks for work:
  1. Check global queue (every 61 ticks, to prevent starvation)
  2. Steal HALF of another random P's local queue
  3. Check network poller
  4. Check timer heap
  5. If nothing → park the M
```

### Syscall Handling

```
Goroutine enters blocking syscall:

  Before:                      After:
  ┌──────┐                     ┌──────┐
  │  P0  │── attached to ──►   │  P0  │── handed to new/idle M
  │      │      M0              │      │
  └──────┘                     └──────┘

  M0 stays blocked in kernel with the G.
  When syscall returns, G tries to reacquire a P.
  If none available → G goes to global queue.
```

### Preemption

| Era | Mechanism | Limitation |
|-----|-----------|-----------|
| Pre-Go 1.14 | Cooperative (function call boundaries) | Tight loops starve other Gs |
| Go 1.14+ | Async signals (SIGURG on Unix) | runtime internals + cgo may resist |

---

## Channel Internals (`hchan`)

```
  hchan struct (runtime/chan.go):
  ┌──────────────────────────────────────────┐
  │  qcount   uint     ← elements in buffer │
  │  dataqsiz uint     ← buffer capacity    │
  │  buf      *byte    ← circular buffer    │
  │  elemsize uint16                        │
  │  sendx    uint     ← send index         │
  │  recvx    uint     ← receive index      │
  │  recvq    waitq    ← blocked receivers  │
  │  sendq    waitq    ← blocked senders    │
  │  lock     mutex                          │
  └──────────────────────────────────────────┘

  Circular buffer (buffered channel, cap=4):
  ┌─────┬─────┬─────┬─────┐
  │  A  │  B  │     │     │
  └─────┴─────┴─────┴─────┘
    ▲                  ▲
   recvx             sendx

  Wait queues (sudog linked list):
  sendq: G3 → G7 → nil     (blocked senders)
  recvq: G5 → nil           (blocked receivers)
```

### Unbuffered: Direct Stack-to-Stack Copy

```
  Sender G                         Receiver G
  ┌──────────┐                     ┌──────────┐
  │ stack:   │   direct copy       │ stack:   │
  │ val = 42 │ ──────────────────► │ val = 42 │
  └──────────┘   (no buffer)       └──────────┘
```

No intermediate buffer. Full bidirectional synchronization (happens-before in both directions).

### Blocking and Waking

```
Sender on full channel:
  1. Create sudog { G: current, elem: &value }
  2. Enqueue sudog on channel's sendq
  3. Park goroutine (gopark)

Space becomes available (receiver reads):
  1. Dequeue sudog from sendq
  2. Copy data from sudog.elem to buffer (or directly to receiver)
  3. Mark G as runnable (goready)
  4. G gets placed on a P's run queue
```

---

## `select` Internals

```
  select { case <-ch1:  case ch2 <- v:  case <-ch3: }

  Evaluation steps:
  ┌──────────────────────────────────────────────────────┐
  │ 1. Evaluate all channel expressions (source order)   │
  │ 2. Lock ALL involved channels                        │
  │ 3. Check which cases are ready                       │
  │ 4. Multiple ready → pseudo-random pick               │
  │ 5. None ready + default → run default                │
  │ 6. None ready, no default →                          │
  │    • Create sudog on ALL channels' wait queues       │
  │    • Park goroutine                                  │
  │ 7. When woken → remove sudog from all OTHER channels │
  │    (O(N) per operation)                              │
  └──────────────────────────────────────────────────────┘
```

---

## `sync.Once` Internals

```
  Fast path (after initialization):
  ┌──────────────────────────────────┐
  │  atomic.Load(&done) == 1         │  ← single instruction
  │  return immediately              │
  └──────────────────────────────────┘

  Slow path (first call):
  ┌──────────────────────────────────┐
  │  Lock mutex                      │
  │  Double-check: done == 0?        │
  │  Call f()                        │
  │  atomic.Store(&done, 1)          │  ← AFTER f completes (happens-before!)
  │  Unlock mutex                    │
  └──────────────────────────────────┘
```

**Deadlock trap:** `once.Do(func() { once.Do(...) })` — inner `Do` tries to lock the same mutex.

### `sync.OnceValue` (Go 1.21+)

```go
var getConfig = sync.OnceValue(func() *Config {
    return loadConfig()
})
cfg := getConfig()  // computed once, returned every subsequent call
```

---

## `sync.Pool` Internals

```
  Per-P structure:
  ┌────────────────────────────────────┐
  │  P0: private (lock-free, 1 item)  │
  │      shared  (lock-free list)      │
  ├────────────────────────────────────┤
  │  P1: private                       │
  │      shared                        │
  └────────────────────────────────────┘

  Get():
    1. Check own P's private slot
    2. Pop from own P's shared list
    3. Steal from other P's shared list
    4. Check victim cache (last GC cycle's pool)
    5. Call New() function

  GC interaction:
    GC start → current pool becomes "victim"
             → previous victim discarded
    Objects survive at MOST 2 GC cycles in pool
```

**It is NOT** a connection pool or durable cache. Objects can be GC'd at any time. Pool objects that are expensive to allocate but cheap to reset (byte buffers, encoders). Always `Reset()` before `Put()`.

---

## `atomic` vs Mutex vs Channels

```
                    Read Latency    Write Latency    Multi-Var Invariants
  ┌─────────────┬──────────────┬──────────────────┬──────────────────────┐
  │ atomic      │  ~1 ns       │  ~1 ns           │  ✗ No               │
  │ sync.Mutex  │  ~25 ns      │  ~25 ns          │  ✓ Yes              │
  │ sync.RWMutex│  ~5 ns (read)│  ~25 ns (write)  │  ✓ Yes              │
  │ Channel     │  ~50 ns      │  ~50 ns          │  ✓ (via ownership)  │
  └─────────────┴──────────────┴──────────────────┴──────────────────────┘
```

Rule of thumb: atomics for single-variable, read-heavy state. Mutex for multi-variable invariants. Channels for communication and workflow coordination.

# Gotchas & Pitfalls — At a Glance

> The most common Go traps, collected in one place.
> For subtle race conditions and advanced traps → [Advanced](advanced.md)

---

## Top 10 Gotchas

| # | Gotcha | What Happens |
|---|--------|-------------|
| 1 | [Nil interface ≠ nil value in interface](#1-nil-interface-trap) | `w == nil` is `false` even though the value inside is nil |
| 2 | [Shared slice backing array](#2-shared-backing-array) | `append` overwrites data in the original slice |
| 3 | [Loop variable capture](#3-loop-variable-capture) | All goroutines see the same (final) value |
| 4 | [Writing to nil map](#4-nil-map-write) | Panics |
| 5 | [Concurrent map access](#5-concurrent-map-access) | Fatal crash (unrecoverable) |
| 6 | [Defer in loop](#6-defer-in-loop) | Resources pile up until function returns |
| 7 | [Send/close on closed channel](#7-closed-channel-operations) | Panics |
| 8 | [Data races that "work" on x86](#8-hidden-data-races) | Break on ARM64 or with more cores |
| 9 | [`recover()` not in defer body](#9-recover-placement) | Recovery silently fails |
| 10 | [http.Client per request](#10-http-client-per-request) | Connection pool leak |

---

### 1. Nil Interface Trap

```go
var buf *bytes.Buffer  // nil
var w io.Writer = buf  // NOT nil!
w == nil               // false
```

**Fix:** Return `nil` directly, not a typed nil variable.

### 2. Shared Backing Array

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]
b = append(b, 99)  // overwrites a[3]!
```

**Fix:** `b := a[1:3:3]` — cap the capacity.

### 3. Loop Variable Capture

```go
for i := 0; i < 5; i++ {
    go func() { fmt.Println(i) }()  // all print 5
}
```

**Fix (pre-1.22):** `i := i` inside loop. **Go 1.22+:** Fixed at language level.

### 4. Nil Map Write

```go
var m map[string]int
m["key"] = 1  // PANIC
```

**Fix:** `m := make(map[string]int)` or `m := map[string]int{}`.

### 5. Concurrent Map Access

Concurrent read+write or write+write → **fatal crash** (not a panic, `recover` can't help).

### 6. Defer in Loop

All defers accumulate until the enclosing function returns. 1M iterations = 1M deferred calls + 1M held resources.

### 7. Closed Channel Operations

| Operation | Result |
|-----------|--------|
| Close a closed channel | **Panic** |
| Send on closed channel | **Panic** |
| Receive from closed channel | Zero value (ok=false) |

### 8. Hidden Data Races

x86 has strong memory ordering. Races hide. They surface on ARM64, with more cores, or compiler changes.

### 9. `recover()` Placement

```go
defer recover()              // ✗ does nothing
defer func() { recover() }() // ✓ works
```

### 10. `http.Client` Per Request

Each `http.Transport` has its own connection pool. Creating per request = wasted TLS handshakes + connection leaks.

---

## Resource Cleanup Checklist

| Resource | Close With | Notes |
|----------|-----------|-------|
| HTTP response body | `defer resp.Body.Close()` | Also drain with `io.Copy(io.Discard, ...)` |
| Files | `defer f.Close()` | Check `Close()` error on writes |
| DB connections | `defer db.Close()` | Usually in `main()` |
| Network connections | `defer conn.Close()` | Set deadlines |
| Channels | `close(ch)` | Only sender closes |
| Mutexes | `defer mu.Unlock()` | Always pair |
| Goroutines | `sync.WaitGroup` | Ensure completion |
| Context cancel | `defer cancel()` | Releases timer goroutines |

---

## Nil Slice vs Empty Slice

```go
var s1 []int          // nil:   s1 == nil, JSON → null
s2 := []int{}         // empty: s2 != nil, JSON → []
s3 := make([]int, 0)  // empty: s3 != nil, JSON → []
```

All three work with `len()`, `cap()`, `range`, and `append`.

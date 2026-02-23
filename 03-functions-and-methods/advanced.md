# Functions & Methods — Advanced

> `defer` internals and method set edge cases. Back to → [At a Glance](README.md)

---

## `defer` Implementation Evolution

```
Go Version     Implementation                   Cost per defer
─────────────────────────────────────────────────────────────
Pre-1.13       Heap-allocated _defer records     ~35 ns
               linked into goroutine chain

Go 1.13+       Open-coded defers (≤8 defers,    ~6 ns
               no loops): inlined at each
               return point with bitmask

Go 1.14+       Stack-allocated defer records      ~6 ns
               when heap alloc can be avoided
```

### Why Defer in Loops Is Dangerous

```go
for i := 0; i < 1_000_000; i++ {
    f, _ := os.Open(name)
    defer f.Close()  // accumulates until function returns
}
```

All 1M defers pile up:

```
Function stack frame:
┌──────────────────────────────────────────┐
│ _defer → _defer → _defer → ... (1M)     │
│                                          │
│ 1M file descriptors held open            │
│ 1M _defer records allocated              │
│                                          │
│ At function return: all execute LIFO     │
└──────────────────────────────────────────┘
```

**Fix:** Extract loop body into a function, or close explicitly.

---

## Method Sets & Interface Satisfaction — Why?

The compiler enforces value-receiver-only method sets on `T` because of **non-addressable values**:

```
Map values:        m["key"].PointerMethod()  ← can't take address
Function returns:  getUser().PointerMethod() ← can't take address
```

If the compiler allowed it, it would silently operate on a copy that gets discarded — a silent bug worse than a compile error.

---

## `recover()` Rules (Summary)

```
✓ Works:     defer func() { recover() }()
✗ No effect: defer recover()                   ← not called inside defer body
✗ No effect: defer fmt.Println(recover())      ← evaluated before defer runs
✗ No effect: recover() in different goroutine   ← cross-goroutine recovery impossible
```

A panic in one goroutine cannot be recovered by another. Unrecovered panics crash the **entire program**.

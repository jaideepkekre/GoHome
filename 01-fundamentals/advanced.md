# Fundamentals — Advanced

> Deeper internals behind the basics. Back to → [At a Glance](README.md)

---

## How Zero Values Actually Work

When Go allocates memory (stack or heap), it **always** zeroes it. There's no uninitialized memory. The runtime uses `memclr` (optimized per-architecture) to zero allocations. This is why `new(T)` returns a pointer to a zeroed `T`, and `var x T` is always safe.

This has performance implications: large struct allocations pay the zeroing cost. For hot paths allocating large buffers, `sync.Pool` can avoid repeated zeroing. See [Concurrency — sync.Pool](../06-concurrency/advanced.md#syncpool).

---

## `init()` Execution Order

```
┌─────────────────────────────────────────────────────┐
│  1. Package-level variables (dependency order)      │
│  2. init() functions (source file order, per pkg)   │
│  3. Repeat for each importing package (DAG order)   │
│  4. main.main() runs LAST                           │
└─────────────────────────────────────────────────────┘
```

A package can have **multiple** `init()` functions, even in the same file. All run before the package is considered "ready."

**Traps:** `init()` runs before `main()` — no goroutines, no flags parsed, no config loaded. Side effects in `init()` make testing hard.

**Prefer:** Explicit init in `main()` or lazy init with `sync.Once`. See [Concurrency — sync.Once](../06-concurrency/advanced.md#synconce-internals).

---

## Build Tags

```go
//go:build linux && amd64
```

Files with `_linux.go`, `_amd64.go` suffixes have implicit tags. `_test.go` files compile only during `go test`.

---

## The `replace` Directive Trap

```
replace github.com/foo/bar => ../local-bar
```

`replace` in `go.mod` silently overrides dependencies. **Always grep for it** when debugging unexpected behavior. `replace` in non-main modules is **ignored** by consumers.

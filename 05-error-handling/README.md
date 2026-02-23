# Error Handling — At a Glance

> Quick-reference for errors, wrapping, `errors.Is`/`As`, and `panic`/`recover`.
> For advanced patterns → [Advanced](advanced.md)

---

## The Basics

```go
type error interface {
    Error() string
}
```

Functions return errors as the last return value. No exceptions.

```go
f, err := os.Open("file.txt")
if err != nil {
    return fmt.Errorf("opening file: %w", err)
}
defer f.Close()
```

---

## Wrapping & Inspecting

### Wrapping

```go
return fmt.Errorf("reading config: %w", err)   // %w wraps (chain preserved)
return fmt.Errorf("reading config: %v", err)    // %v formats (chain BROKEN)
```

### Inspecting

```go
// Check for a specific error value (walks the chain)
if errors.Is(err, os.ErrNotExist) { ... }

// Check for a specific error type (walks the chain)
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println(pathErr.Path)
}
```

```
Error Chain Visualization:

  fmt.Errorf("A: %w", err1)
       │
       ▼
  fmt.Errorf("B: %w", err2)
       │
       ▼
  os.ErrNotExist   ← errors.Is walks down to find this
```

---

## Sentinel Errors

```go
var ErrNotFound = errors.New("not found")
```

Exported, documented, stable. Check with `errors.Is`. Don't overuse — if callers don't need to distinguish the kind, a descriptive message is enough.

---

## `panic` / `recover`

### When Panics Happen

- `nil` pointer dereference, index out of bounds
- Send on closed channel
- Explicit `panic()` calls

### Recovery Pattern

```go
defer func() {
    if r := recover(); r != nil {
        log.Println("recovered:", r)
    }
}()
```

### What Does NOT Work

```go
defer recover()                // ✗  not inside a defer function body
defer fmt.Println(recover())   // ✗  evaluated before defer runs
// recover in goroutine B      // ✗  can't catch panic from goroutine A
```

### Unrecoverable Crashes

| Situation | Why |
|-----------|-----|
| Concurrent map writes | Fatal runtime error, not a panic |
| Stack overflow | Runtime limit exceeded |
| Out of memory | OS-level failure |

# Error Handling — Advanced

> Custom error types, advanced patterns, and `panic`/`recover` internals. Back to → [At a Glance](README.md)

---

## Custom Error Types

Implement `Error() string` and optionally `Unwrap`, `Is`, or `As`:

```go
type ValidationError struct {
    Field   string
    Message string
    Wrapped error
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s — %s", e.Field, e.Message)
}

func (e *ValidationError) Unwrap() error { return e.Wrapped }
```

### Custom `Is()` and `As()`

```go
// Custom Is: match on field rather than identity
func (e *ValidationError) Is(target error) bool {
    t, ok := target.(*ValidationError)
    return ok && t.Field == e.Field
}
```

---

## Error Wrapping with `defer`

The standard pattern for adding context to all error paths in a function:

```go
func loadConfig(path string) (cfg *Config, err error) {
    defer func() {
        if err != nil {
            err = fmt.Errorf("loadConfig(%q): %w", path, err)
        }
    }()

    f, err := os.Open(path)
    if err != nil { return nil, err }    // wrapped by defer
    defer f.Close()

    err = json.NewDecoder(f).Decode(&cfg)
    return cfg, err                       // also wrapped by defer
}
```

---

## `panic` / `recover` Internals

### Panic Unwinding

```
Goroutine stack during panic:

  main()
    ├── handler()
    │     ├── processRequest()     ← panic here
    │     │     └── PANIC unwinds
    │     ├── defer func() {       ← recover() catches it
    │     │       recover()
    │     │   }
    │     └── (continues after recover)
    └── ...

Without recover → unwind reaches top → program crashes
```

### Cross-Goroutine Panics

```
  Goroutine A               Goroutine B
  ┌──────────────┐          ┌──────────────┐
  │ defer recover │          │ panic("boom")│
  │ ...          │          │              │
  └──────────────┘          └──────┬───────┘
                                   │
                     Cannot cross  │  goroutine boundary
                                   ▼
                            PROGRAM CRASHES
```

A panic in goroutine B **cannot** be recovered by goroutine A. This is why `net/http` adds recovery in each handler, and why you should add `recover` in goroutines launched by libraries.

---

## The Nil Interface Error Trap

```go
func getError() error {
    var err *MyError   // typed nil
    return err         // returns non-nil error!
}

getError() != nil  →  true (ALWAYS)
```

```
  What you think:            What Go sees:
  ┌──────────┐               ┌──────────────────┬──────────┐
  │   nil    │               │ type: *MyError    │ val: nil │
  └──────────┘               └──────────────────┴──────────┘
  err == nil → true           err == nil → FALSE
```

**Fix:** Return `nil` directly, never a typed nil variable.

# Interfaces & Generics — At a Glance

> Quick-reference for interfaces, type assertions, embedding, and generics.
> For `iface`/`eface` internals and `itab` caching → [Advanced](advanced.md)

---

## Interfaces

An interface defines method signatures. Satisfaction is **implicit** — no `implements` keyword.

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
// Any type with a Write method satisfies io.Writer automatically.
```

### Type Assertions

```go
s := i.(string)         // panics if wrong type
s, ok := i.(string)     // safe — ok is false if wrong type
```

### Type Switches

```go
switch v := i.(type) {
case string:  fmt.Println("string:", v)
case int:     fmt.Println("int:", v)
default:      fmt.Println("unknown")
}
```

### Interface Equality

Two interface values are equal if they have **equal concrete values** AND **identical dynamic types**, or both are `nil`.

**Trap:** Comparing interfaces holding non-comparable types (slices, maps) → **runtime panic**.

---

## The Nil Interface Trap

```
var buf *bytes.Buffer  ← nil pointer
var w io.Writer = buf  ← NOT nil interface!

  Interface value:
  ┌────────────────────┬──────────┐
  │ type: *bytes.Buffer│ value: nil│  ← type is set!
  └────────────────────┴──────────┘

  w == nil  →  false
```

An interface is nil **only** when both type AND value are nil. See [Gotchas](../10-gotchas-and-pitfalls/README.md) for the full trap.

---

## Embedding Interfaces

Compose larger interfaces from smaller ones:

```go
type ReadWriter interface {
    io.Reader
    io.Writer
}
```

Embedding an interface in a struct: the struct satisfies the interface but methods **panic** if you don't provide an implementation (embedded interface is nil).

---

## Generics (Go 1.18+)

```go
func Map[S ~[]E, E any](s S, f func(E) E) S {
    result := make(S, len(s))
    for i, v := range s { result[i] = f(v) }
    return result
}
```

`~T` means "any type whose underlying type is `T`" — allows named types.

### Limitations

- No method type parameters (function and type level only)
- No specialization or variadic type parameters
- Constraint inference has limits — sometimes you must specify type args explicitly
- For ordered comparison (`<`, `>`), use `cmp.Ordered` from the standard library

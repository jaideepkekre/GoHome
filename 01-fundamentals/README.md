# Fundamentals — At a Glance

> Quick-reference for types, zero values, loops, `make`/`new`, and common patterns.
> For deeper coverage → [Advanced](advanced.md)

---

## Types & Zero Values

| Type | Zero Value | Notes |
|------|-----------|-------|
| `bool` | `false` | |
| `int`, `int8`…`int64` | `0` | `int` is platform-sized (32 or 64 bit) |
| `float32`, `float64` | `0.0` | |
| `string` | `""` | Immutable byte sequence |
| `pointer` | `nil` | |
| `slice` | `nil` | nil slice has len=0, cap=0 |
| `map` | `nil` | Read OK; **write panics** |
| `channel` | `nil` | Send/receive blocks forever |
| `func` | `nil` | |
| `interface` | `nil` | Both type AND value must be nil |
| `struct` | all fields zeroed | Each field gets its own zero value |

**Key insight:** Zero values are a guarantee, not undefined. `var mu sync.Mutex` is ready to use — no constructor needed.

---

## `make` vs `new`

| | `new(T)` | `make(T, args...)` |
|---|---------|-------------------|
| Returns | `*T` (pointer) | `T` (value) |
| Works on | Any type | Slices, maps, channels only |
| Does | Zeros memory | Allocates + **initializes** internal structures |

```go
s := make([]int, 0, 10)    // slice: len=0, cap=10
m := make(map[string]int)   // map: ready to use
ch := make(chan int, 5)      // buffered channel, cap 5
```

---

## Loops

Go has **one loop keyword**: `for`.

```go
for i := 0; i < 10; i++ { }        // C-style
for condition { }                    // while-style
for { }                              // infinite
for i, v := range slice { }         // slice (index, value)
for k, v := range myMap { }         // map — order is RANDOM
for v := range ch { }               // channel — blocks until closed
for i, r := range "hello 世界" { }  // string — iterates RUNES, not bytes
```

---

## Variadic Functions

```go
func sum(nums ...int) int { /* nums is a []int */ }
sum(1, 2, 3)       // individual values
sum(mySlice...)     // spread a slice
```

---

## Functional Options Pattern

```go
type Option func(*Server)

func WithPort(p int) Option {
    return func(s *Server) { s.port = p }
}

func NewServer(opts ...Option) *Server {
    s := &Server{port: 8080} // sensible defaults
    for _, opt := range opts { opt(s) }
    return s
}

srv := NewServer(WithPort(9090))
```

---

## `iota` and Constants

```go
type Weekday int
const (
    Sunday Weekday = iota  // 0
    Monday                 // 1
    Tuesday                // 2
)

// Bitmask pattern
const (
    Read    = 1 << iota // 1
    Write               // 2
    Execute             // 4
)
```

---

## Multiple Returns & Named Returns

```go
func divide(a, b float64) (float64, error) {
    if b == 0 { return 0, fmt.Errorf("division by zero") }
    return a / b, nil
}
```

Named returns are useful for godoc, error-wrapping with `defer`, and short functions. Avoid in long functions — naked returns become confusing.

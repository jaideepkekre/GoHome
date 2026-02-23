# Functions & Methods — At a Glance

> Quick-reference for receivers, `defer`, closures, and function types.
> For implementation internals → [Advanced](advanced.md)

---

## Pointer vs Value Receivers

```go
func (t T) ValueMethod()    { }  // gets a COPY
func (t *T) PointerMethod() { }  // gets a pointer, can mutate
```

### Method Set Rules

```
  Value (T):        can call value receiver methods ONLY
  Pointer (*T):     can call BOTH value and pointer receiver methods

  ┌─────────────────────────────────────────────────┐
  │  If ANY method is pointer receiver →            │
  │  use pointer receivers for ALL methods on type  │
  └─────────────────────────────────────────────────┘
```

This affects interface satisfaction:

```go
type Stringer interface { String() string }
type Foo struct{}
func (f *Foo) String() string { return "foo" }

var _ Stringer = Foo{}   // ✗ COMPILE ERROR
var _ Stringer = &Foo{}  // ✓ OK
```

### When to Use Which

| Pointer Receiver | Value Receiver |
|-----------------|----------------|
| Method mutates the receiver | Type is small + immutable |
| Struct is large | Type is map/slice/chan/func |
| Consistency with other methods | |

---

## `defer`

Defers execute in **LIFO** order when the enclosing function returns.

```go
defer f.Close()         // runs on function exit
defer mu.Unlock()       // pair with Lock()
defer cancel()          // always cancel contexts
```

### Arguments are evaluated immediately

```go
x := 1
defer fmt.Println(x)   // prints 1, not 2
x = 2
```

To capture the later value, use a closure:

```go
defer func() { fmt.Println(x) }()  // prints 2
```

### Named returns + defer = error wrapping

```go
func doWork() (err error) {
    defer func() {
        if err != nil {
            err = fmt.Errorf("doWork: %w", err)
        }
    }()
    // ...
}
```

---

## Closures

A function that captures variables from its surrounding scope — **by reference**, not by value.

```go
func counter() func() int {
    n := 0
    return func() int { n++; return n }
}
c := counter()
c() // 1
c() // 2
```

### The Loop Variable Trap (pre-Go 1.22)

```go
for i := 0; i < 5; i++ {
    go func() { fmt.Println(i) }()  // BAD: all print 5
}
```

**Fix:** `i := i` inside the loop (shadows with copy). **Go 1.22+:** Fixed at the language level — loop vars are per-iteration.

---

## Function Types

Functions are first-class. This is how `http.HandlerFunc` works:

```go
type HandlerFunc func(ResponseWriter, *Request)
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }
```

A function type that satisfies an interface by calling itself.

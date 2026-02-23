# Data Structures — At a Glance

> Quick-reference for slices, maps, structs, strings, and comparison rules.
> For internals and implementation details → [Advanced](advanced.md)

---

## Slices

### Memory Layout

```
Slice Header (24 bytes on 64-bit)
┌──────────┬──────────┬──────────┐
│ pointer  │  length  │ capacity │
│ (8 bytes)│ (8 bytes)│ (8 bytes)│
└────┬─────┴──────────┴──────────┘
     │
     ▼
┌────┬────┬────┬────┬────┬────┬────┐
│ 10 │ 20 │ 30 │  0 │  0 │  0 │  0 │  ← backing array
└────┴────┴────┴────┴────┴────┴────┘
 len=3                    cap=7
```

### Essential Operations

```go
s := make([]int, 0, 10)    // len=0, cap=10
s = append(s, 1, 2, 3)     // grows as needed
b := s[1:3]                 // sub-slice — SHARES backing array
b := s[1:3:3]               // full slice expr — caps capacity (safe)
```

### Growth Strategy

| Current Cap | Growth |
|-------------|--------|
| < 256 | 2x |
| >= 256 | ~1.25x |

### The Shared Backing Array Trap

```go
a := []int{1, 2, 3, 4, 5}
b := a[1:3]                 //  b = [2, 3], shares array with a
b = append(b, 99)           //  OVERWRITES a[3]!
```

**Fix:** `b := a[1:3:3]` — caps capacity so `append` allocates fresh.

---

## Maps

### Quick Reference

```go
m := make(map[string]int)       // initialize
m := map[string]int{"a": 1}     // literal
v := m["key"]                    // read (zero value if missing)
v, ok := m["key"]                // check existence
delete(m, "key")                 // remove
```

### Concurrency Rules

| Operation | Safe? |
|-----------|-------|
| Concurrent reads | Yes |
| Read + write | **Fatal crash** (not a panic) |
| Concurrent writes | **Fatal crash** |

Use `sync.Mutex`, `sync.RWMutex`, or `sync.Map`.

### Nil Map Behavior

Reading → returns zero value (no panic). **Writing → panics.** Always `make()` before writing.

---

## Structs

```go
type User struct {
    Name  string
    Email string
    Age   int
}

var u User  // zero value: Name="", Email="", Age=0
```

### Embedding (Composition, Not Inheritance)

```
┌─────────────────────────┐
│  Derived                │
│  ┌───────────────────┐  │
│  │  Base              │  │
│  │  ID: 42            │  │  ← promoted: d.ID works
│  │  String() string   │  │  ← promoted: d.String() works
│  └───────────────────┘  │
│  Name: "example"        │
└─────────────────────────┘
```

---

## Strings

A string is `(pointer, length)`. Immutable. **Indexing gives bytes, not characters.**

```go
s := "Hello, 世界"
s[7]                // byte, NOT the character '世'
for i, r := range s // rune iteration — correct for Unicode
```

---

## Comparison Rules

| Type | `==` ? | Notes |
|------|--------|-------|
| `bool`, `int`, `float`, `string` | Yes | Direct comparison |
| `pointer`, `channel` | Yes | Same address / same instance |
| `struct` (all fields comparable) | Yes | Field-by-field |
| `struct` (has map/slice/func field) | **No** | Compile error |
| `array` | Yes | Element-by-element |
| `slice`, `map`, `func` | **No** | Only `== nil` allowed |
| `interface` | Conditional | Panics if non-comparable type inside |

For non-comparable types → `reflect.DeepEqual(x, y)` (slow, avoid in hot paths).

# Gotchas & Pitfalls — Advanced

> Subtle race conditions, interface traps, and deep comparison rules.
> Back to → [At a Glance](README.md)

---

## Subtle Race: Interface Assignment

Interface assignment writes **two words** (type pointer + data pointer). This is NOT atomic.

```
  Goroutine A writes interface:         Goroutine B reads interface:
  ┌────────────────┬──────────┐         ┌────────────────┬──────────┐
  │ type: *Foo     │ data: p1 │   ──►   │ type: *Foo     │ data: p2 │
  └────────────────┴──────────┘         └────────────────┴──────────┘
                                         ↑ old type + new data = CRASH
                                           or wrong method dispatch
```

A concurrent read can see a **mismatched pair** (old type, new data or vice versa) → bizarre crashes or wrong method dispatch.

---

## Subtle Race: `WaitGroup.Add` Inside Goroutine

```go
// BAD: race between Add and Wait
for i := 0; i < n; i++ {
    go func() {
        wg.Add(1)          // ← RACE: Wait() might run before Add()
        defer wg.Done()
        // work
    }()
}
wg.Wait()
```

`Add()` must happen-before the `go` statement. The goroutine launch is not guaranteed to execute `Add()` before `Wait()` on the main goroutine.

---

## Map Value Mutation Trap

```go
type Point struct{ X, Y int }

m := map[string]Point{"a": {1, 2}}
m["a"].X = 10  // COMPILE ERROR: cannot assign to struct field in map
```

Map values are not addressable. You can't modify a field in-place.

**Fix:** Assign whole value back, or use `map[string]*Point`.

---

## Comparison Rules (Complete)

```
  Rule 1: Types are compared
          (for structs, field types are compared individually)

  Rule 2: Values must be the same
          (except for unexported fields in some contexts)

  Rule 3: Maps, slices, and functions are reference types
          You can ONLY compare them to nil

  Rule 4: Interfaces go one level down
          Check dynamic type, then dynamic value
          ⚠️  Panics if dynamic type is non-comparable (slice inside interface)
```

---

## String Indexing Trap

```go
s := "Hello, 世界"

s[7]          // byte: 0xe4 (NOT the character '世')
len(s)        // 13 (bytes, not characters)
len([]rune(s))  // 9 (characters/runes)

// Correct character iteration:
for i, r := range s {
    // i = byte position, r = rune (Unicode code point)
}
```

---

## JSON Number Precision Loss

```go
var data interface{}
json.Unmarshal([]byte(`{"id": 9007199254740993}`), &data)

m := data.(map[string]interface{})
fmt.Println(m["id"])  // 9.007199254740992e+18  ← WRONG! Lost precision
```

JSON numbers become `float64` in `interface{}`. Values > 2^53 lose precision.

**Fix:** `json.Decoder` with `UseNumber()`, or unmarshal into a struct with `int64` field.

---

## Pass-by-Value: The Complete Picture

```
  Type           Copied                 Shared
  ─────────────────────────────────────────────
  int/bool       Full value             Nothing
  string         Header (16 B)          Byte array (immutable, safe)
  struct         All fields             Internal pointers (shallow copy)
  slice          Header (24 B)          Backing array
  map            Pointer (8 B)          ENTIRE map
  channel        Pointer (8 B)          ENTIRE channel
  pointer        Address (8 B)          Target value
  interface      Type+data (16 B)       Value may be shared
```

There is no pass-by-reference in Go. Everything is copied. But some types contain internal pointers, making them behave like references.

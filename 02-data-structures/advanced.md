# Data Structures — Advanced

> Internals of slices, maps, and strings. Back to → [At a Glance](README.md)

---

## Slice Internals

### `append` Under the Hood

```
Before append (at capacity):
┌──────────┬─────┬─────┐
│ ptr ──┐  │ len │ cap │   Slice Header
└───────┼──┴─────┴─────┘
        ▼
        ┌───┬───┬───┐
        │ 1 │ 2 │ 3 │         Backing Array (cap=3)
        └───┴───┴───┘

append(s, 4):
  1. cap exceeded → allocate new array (cap=6)
  2. copy [1,2,3] to new array
  3. write 4 at index 3
  4. return NEW slice header pointing to new array

After append:
┌──────────┬─────┬─────┐
│ ptr ──┐  │  4  │  6  │   New Slice Header
└───────┼──┴─────┴─────┘
        ▼
        ┌───┬───┬───┬───┬───┬───┐
        │ 1 │ 2 │ 3 │ 4 │ 0 │ 0 │   New Backing Array
        └───┴───┴───┴───┴───┴───┘

Old array becomes garbage (if no other slices reference it)
```

The growth factor changed in **Go 1.18**: below 256 elements it doubles, above 256 it grows by ~25% + a small constant. This smooths the transition (older versions jumped from 2x → 1.25x at 1024).

### Passing Slices to Functions

```
Caller:                          Function:
┌─────┬─────┬─────┐             ┌─────┬─────┬─────┐
│ ptr │ len │ cap │  ──copy──►  │ ptr │ len │ cap │
└──┬──┴─────┴─────┘             └──┬──┴─────┴─────┘
   │                                │
   └──────────┬─────────────────────┘
              ▼
        ┌───┬───┬───┬───┐
        │ 1 │ 2 │ 3 │ 4 │   SAME backing array
        └───┴───┴───┴───┘
```

Modifications to existing elements → visible to caller.
Appends beyond capacity → new array, caller sees nothing.

---

## Map Internals

### Hash Table Structure

```
                       ┌─ tophash [8]byte   ← top 8 bits of hash
                       │                      (fast reject before full key compare)
map header ──►  Bucket │─ keys [8]K
  │                    │─ values [8]V
  │                    └─ overflow *bucket  ← chain for > 8 entries
  │
  ├─ bucket array (2^B buckets)
  ├─ count (number of elements)
  ├─ flags (concurrent write detection)
  └─ hash seed (random, per-map)
```

Each bucket holds **8 key-value pairs**. The hash of a key selects the bucket, and the tophash byte speeds up comparisons inside the bucket.

### Map Growth (Incremental Rehashing)

```
Load factor > 6.5  OR  too many overflow buckets
                │
                ▼
   ┌─────────────────────────────────────────┐
   │  Allocate new bucket array (2x size)    │
   │  Keep old array alive                   │
   │  Set "growing" flag                     │
   └─────────────────────────────────────────┘
                │
                ▼
   Each subsequent read/write evacuates
   1-2 old buckets lazily
                │
                ▼
   Eventually all old buckets evacuated,
   old array becomes garbage
```

Growth is **incremental** — no single operation pays the full rehash cost.

### Concurrent Access Detection

The map header has a `flags` field. On write, the runtime sets a "writing" flag. On read, it checks the flag. If set → `fatal("concurrent map read and map write")`. This is a **fatal error**, not a panic — `recover()` cannot catch it.

---

## String Internals

### Memory Layout

```
String Header (16 bytes on 64-bit)
┌──────────┬──────────┐
│ pointer  │  length  │
└────┬─────┴──────────┘
     │
     ▼
     ┌────┬────┬────┬────┬────┐
     │ 0x48│0x65│0x6C│0x6C│0x6F│  ← UTF-8 bytes ("Hello")
     └────┴────┴────┴────┴────┘
```

Strings are **immutable**. The compiler and runtime share string memory safely.

### `[]byte(s)` / `string(b)` Conversion Cost

Normally allocates + copies (because strings are immutable, byte slices are mutable).

**Compiler-optimized zero-copy cases:**

| Pattern | Copy? |
|---------|-------|
| `string(b)` as map key in lookup | No |
| `[]byte(s)` in immediate comparison | No |
| `string(b)` in concatenation | No |
| `range []byte(s)` | No |
| Everything else | **Yes** |

For hot paths: `unsafe.String` / `unsafe.Slice` create zero-copy conversions. You own the safety guarantee.

---

## Pass-by-Value: What Actually Gets Copied

```
Primitive (int, bool, float):
  ┌─────┐      ┌─────┐
  │  42 │ ──►  │  42 │   Full copy. Independent.
  └─────┘      └─────┘

Slice:
  ┌───┬───┬───┐      ┌───┬───┬───┐
  │ptr│len│cap│ ──►  │ptr│len│cap│   Header copied.
  └─┬─┴───┴───┘      └─┬─┴───┴───┘
    │                    │
    └────────┬───────────┘            Backing array SHARED.
             ▼

Map:
  ┌─────┐       ┌─────┐
  │ ptr │  ──►  │ ptr │   Pointer copied.
  └──┬──┘       └──┬──┘
     │              │
     └──────┬───────┘                 SAME underlying map.
            ▼
```

There is **no pass-by-reference in Go.** Everything is copied. But some types contain internal pointers, so they act like references.

Reference: [Dave Cheney — There is no pass-by-reference in Go](https://dave.cheney.net/2017/04/29/there-is-no-pass-by-reference-in-go)

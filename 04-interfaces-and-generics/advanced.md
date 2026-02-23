# Interfaces & Generics — Advanced

> Interface memory layout, `itab` caching, and generics compilation. Back to → [At a Glance](README.md)

---

## Interface Memory Representation

### Non-Empty Interface (`iface`)

```
  Interface value (e.g., io.Writer holding *os.File)
  ┌─────────────────────────────────────────────────────────┐
  │  itab*                           │  data*               │
  │  ┌───────────────────────────┐   │  ┌────────────────┐  │
  │  │ interface type descriptor │   │  │ pointer to the │  │
  │  │ concrete type descriptor  │   │  │ actual value   │  │
  │  │ method table:             │   │  │ (or heap copy) │  │
  │  │   Write → os.(*File).Write│   │  └────────────────┘  │
  │  └───────────────────────────┘   │                      │
  └─────────────────────────────────────────────────────────┘
```

### Empty Interface (`eface` / `any`)

```
  ┌────────────────────┬──────────┐
  │ type*              │  data*   │   No method table needed.
  └────────────────────┴──────────┘
```

---

## `itab` Caching

```
First assignment:  concrete type → interface type
                       │
                       ▼
         ┌──────────────────────────────────┐
         │  Compute itab:                   │
         │  • match method names            │
         │  • resolve function pointers     │
         │  • store in global hash table    │
         │    key: (interface type,          │
         │          concrete type)           │
         └──────────────────────────────────┘
                       │
                       ▼
         Cached. All subsequent assignments of
         same type pair reuse the cached itab.
```

First interface assignment is slightly more expensive. After that, it's a hash table lookup.

---

## Type Assertion Mechanics

| Assertion Type | How It Works |
|---------------|-------------|
| Non-empty interface → concrete | Compare `itab`'s concrete type pointer against target |
| Empty interface → concrete | Compare `type` pointer against target |
| Interface → interface | `itab` lookup for the new (interface, concrete) pair |

All are fast — pointer comparison for exact type matching.

**Type switches** compile to a chain of type comparisons. For many cases, the compiler may emit a hash-based jump table.

---

## Interface Cost Analysis

```
  Call path:
  ┌──────────┐    indirect call     ┌──────────────┐
  │ caller   │ ──────────────────►  │ method impl  │
  │          │    via itab.fun[0]   │              │
  └──────────┘                      └──────────────┘

  Dispatch cost:  ~2-3ns (indirect function call)
  Real cost:      HEAP ALLOCATION of the data
```

The dispatch is cheap. The expensive part is **heap escape**: storing a value in an interface often causes an allocation because the compiler can't prove the interface won't outlive the stack frame.

**Optimization:** Small values (≤ pointer size) can be stored directly in the `data` pointer — no allocation.

---

## Generics Compilation: GC-Shape Stenciling

```
Generic function:   Map[S ~[]E, E any](...)

Instantiated with *User (pointer):
  ┌─────────────────────────────────┐
  │  Shared "pointer shape" code    │  ← one copy for ALL pointer types
  │  + dictionary with type info    │
  └─────────────────────────────────┘

Instantiated with int (value):
  ┌─────────────────────────────────┐
  │  Separate "int shape" code      │  ← distinct instantiation
  └─────────────────────────────────┘

Instantiated with string (value):
  ┌─────────────────────────────────┐
  │  Separate "string shape" code   │  ← distinct instantiation
  └─────────────────────────────────┘
```

Go uses **GC-shape stenciling** — a middle ground between full monomorphization (C++) and pure dictionary dispatch (Java). Types with the same GC shape (all pointer types, for example) share one code instantiation + a runtime dictionary.

# GoHome 🏠

A personal Go knowledge base and design templates. For experienced developers who get rusty between projects.

Each topic has **two tiers**: an **At a Glance** quick-reference and an **Advanced** deep-dive into internals. Pick what you need.

---

## 📋 At a Glance — Quick Reference Index

Jump straight to the practical Q&A for any topic.

| # | Topic | Key Questions |
|---|-------|--------------|
| 01 | [Fundamentals](01-fundamentals/) | Types, zero values, `make` vs `new`, loops, variadic & functional options |
| 02 | [Data Structures](02-data-structures/) | Slices, maps, structs, strings, comparison rules, copying behavior |
| 03 | [Functions & Methods](03-functions-and-methods/) | Value vs pointer receivers, `defer`, closures, function types |
| 04 | [Interfaces & Generics](04-interfaces-and-generics/) | Implicit satisfaction, type assertions, embedding, generics basics |
| 05 | [Error Handling](05-error-handling/) | `errors.Is`/`As`, wrapping with `%w`, `panic`/`recover`, sentinel errors |
| 06 | [Concurrency](06-concurrency/) | Goroutines, channels, `select`, `context`, `sync.WaitGroup`, patterns |
| 07 | [Memory & GC](07-memory-and-gc/) | Happens-before rules, escape analysis, `GOGC` / `GOMEMLIMIT` |
| 08 | [Networking & HTTP](08-networking-and-http/) | `net/http` handlers, timeouts, REST calls, JSON quick-reference |
| 09 | [Tooling & Performance](09-tooling-and-performance/) | `pprof`, race detector, benchmarks, `GOMAXPROCS` |
| 10 | [Gotchas & Pitfalls](10-gotchas-and-pitfalls/) | Top 10 traps, resource cleanup checklist, nil slice vs empty slice |

---

## 🔬 Advanced — Deep Dive Index

For when you need to understand what's happening under the hood.

| # | Topic | What's Inside |
|---|-------|--------------|
| 01 | [Fundamentals](01-fundamentals/advanced.md) | `init()` order, build tags, `replace` directive traps |
| 02 | [Data Structures](02-data-structures/advanced.md) | Slice growth internals, map bucket/tophash structure, string conversion costs, pass-by-value diagrams |
| 03 | [Functions & Methods](03-functions-and-methods/advanced.md) | `defer` implementation evolution (heap → open-coded), non-addressable values, `recover()` rules |
| 04 | [Interfaces & Generics](04-interfaces-and-generics/advanced.md) | `iface`/`eface` memory layout, `itab` caching, interface cost analysis, GC-shape stenciling |
| 05 | [Error Handling](05-error-handling/advanced.md) | Custom `Is()`/`As()`, error wrapping with `defer`, panic unwinding diagram, nil interface error trap |
| 06 | [Concurrency](06-concurrency/advanced.md) | G/M/P scheduler, `hchan` struct, `select` locking mechanics, `sync.Once`/`sync.Pool` internals, atomic vs mutex latency |
| 07 | [Memory & GC](07-memory-and-gc/advanced.md) | Tri-color mark-and-sweep, write barrier, STW phases, struct alignment, false sharing with cache lines |
| 08 | [Networking & HTTP](08-networking-and-http/advanced.md) | Transport connection pooling, `io.Reader` composition, JSON performance (easyjson, sonic), `sync.Pool` for encoders |
| 09 | [Tooling & Performance](09-tooling-and-performance/advanced.md) | `cgo` call overhead (50-200ns), race detector (TSan v2), `pprof` heap modes, mutex/block profiles |
| 10 | [Gotchas & Pitfalls](10-gotchas-and-pitfalls/advanced.md) | Interface assignment race, `WaitGroup.Add` race, map value mutation, JSON precision loss, complete pass-by-value table |

---

## 🛠️ Code Examples

| Folder | Contents |
|--------|---------|
| [REST/SERVER](REST/SERVER/) | HTTP server patterns, JSON handlers, RabbitMQ integration |

---

## How to Use

**Rusty on a topic?** Start with the At a Glance README in the topic folder.

**Need internals?** Open the `advanced.md` in the same folder.

**Prepping for an interview?** Read through the "What happens when..." questions in both tiers.

**Quick lookup?** Each At a Glance doc has tables and cheat sheets near the top.

---

## Resources

- [Go Spec](https://go.dev/ref/spec) · [Effective Go](https://go.dev/doc/effective_go) · [Go by Example](https://gobyexample.com/)
- [Go Blog](https://go.dev/blog/) · [Ardan Labs Blog](https://www.ardanlabs.com/blog/) · [Dave Cheney](https://dave.cheney.net/)
- [yourbasic.org/golang](https://yourbasic.org/golang/) · [Go Wiki](https://go.dev/wiki/)

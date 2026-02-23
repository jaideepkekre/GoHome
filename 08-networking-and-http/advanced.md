# Networking & HTTP — Advanced

> Transport internals, connection pooling, `io.Reader` patterns, and JSON performance.
> Back to → [At a Glance](README.md)

---

## Transport & Connection Pooling

```
  http.Client
  ┌────────────────────────────────────────────┐
  │  Transport (http.RoundTripper)             │
  │  ┌──────────────────────────────────────┐  │
  │  │  Connection Pool                     │  │
  │  │  ┌────────────────────────────────┐  │  │
  │  │  │  idle conns per host (default 2)│  │  │
  │  │  │  max idle total (default 100)  │  │  │
  │  │  │  idle timeout (default 90s)    │  │  │
  │  │  └────────────────────────────────┘  │  │
  │  │  TLS Session Cache                   │  │
  │  │  Dialer (with DNS resolution)        │  │
  │  └──────────────────────────────────────┘  │
  │  Timeout (end-to-end)                      │
  │  Redirect policy                           │
  └────────────────────────────────────────────┘
```

### Tuning Defaults

| Setting | Default | Often Needed |
|---------|---------|-------------|
| `MaxIdleConnsPerHost` | **2** | 10-100 for high-throughput services |
| `MaxIdleConns` | 100 | Increase if calling many hosts |
| `MaxConnsPerHost` | 0 (unlimited) | Set to prevent stampede |

### Why Per-Request Clients Are a Bug

```
  BAD:  each call creates a new Transport
  ┌──────────────────────────────────────┐
  │ Request 1 → Transport A → Pool A    │  ← TLS handshake
  │ Request 2 → Transport B → Pool B    │  ← TLS handshake again!
  │ Request 3 → Transport C → Pool C    │  ← and again!
  │ ...                                  │
  │ Idle connections from A, B, C pile up│
  └──────────────────────────────────────┘

  GOOD: shared Client + Transport
  ┌──────────────────────────────────────┐
  │ Request 1 ─┐                         │
  │ Request 2 ─┤─► Single Transport      │  ← connections reused
  │ Request 3 ─┘   Single Pool           │  ← TLS sessions cached
  └──────────────────────────────────────┘
```

### Response Body: Read and Close

```
  If you don't drain + close resp.Body:
  ┌──────────────────────────────────────────────┐
  │  Connection stays "in use"                   │
  │  Cannot return to idle pool                  │
  │  Eventually: connection leak + pool exhaustion│
  └──────────────────────────────────────────────┘

  Always:
    defer resp.Body.Close()
    io.Copy(io.Discard, resp.Body)  // drain remaining bytes
```

---

## `io.Reader` Composition Patterns

`io.Reader` is the composable building block for all I/O:

```
  io.Reader: Read(p []byte) (n int, err error)

  Composition:
  ┌──────────┐    ┌──────────┐    ┌──────────┐
  │ os.File  │───►│ TeeReader│───►│ gzip.    │───► output
  │ (source) │    │ (splits) │    │ Reader   │
  └──────────┘    └────┬─────┘    └──────────┘
                       │
                       ▼
                  ┌──────────┐
                  │ sha256   │  (side channel: hash as you read)
                  │ .Write() │
                  └──────────┘
```

| Pattern | Function | Use Case |
|---------|----------|---------|
| PassThrough | `io.TeeReader` | Progress tracking during download |
| Mutating | Custom `Read()` | Transform bytes in-flight (rot13, etc.) |
| Splitting | `io.TeeReader` | Hash + process simultaneously |
| Concat | `io.MultiReader` | Stitch multiple sources into one stream |
| Limiting | `io.LimitReader` | Cap bytes read (prevent DoS) |

---

## JSON Performance Deep Dive

### Why `encoding/json` Is Slow

`encoding/json` uses reflection for every marshal/unmarshal. This means:
- Type info is resolved at runtime
- Allocation on every operation
- No code paths optimized for specific struct layouts

### Alternatives

| Library | Approach | Speed vs stdlib |
|---------|----------|----------------|
| `easyjson` | Code generation | 4-5x faster |
| `goccy/go-json` | Runtime optimization | 2-3x faster |
| `bytedance/sonic` | JIT + SIMD | 5-10x faster |
| `encoding/json/v2` | New stdlib (experimental) | 2-3x faster |

### Reducing Allocations

Pool your encoders/decoders with `sync.Pool`:

```go
var bufPool = sync.Pool{
    New: func() any { return new(bytes.Buffer) },
}

func marshal(v any) ([]byte, error) {
    buf := bufPool.Get().(*bytes.Buffer)
    defer func() { buf.Reset(); bufPool.Put(buf) }()
    err := json.NewEncoder(buf).Encode(v)
    return buf.Bytes(), err
}
```

### Number Precision Fix

```go
dec := json.NewDecoder(reader)
dec.UseNumber()  // numbers become json.Number (string), not float64
```

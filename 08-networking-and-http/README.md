# Networking & HTTP — At a Glance

> Quick-reference for `net/http`, REST patterns, and JSON.
> For transport internals and JSON performance → [Advanced](advanced.md)

---

## `net/http` Handlers

```
  ┌──────────────────────────────────────────────────────────────┐
  │  Handle(pattern, handler)     ← registers a Handler interface│
  │  HandleFunc(pattern, func)    ← registers a plain function  │
  │  HandlerFunc                  ← type adapter (func → Handler)│
  └──────────────────────────────────────────────────────────────┘

  HandlerFunc is the bridge:
    type HandlerFunc func(ResponseWriter, *Request)
    func (f HandlerFunc) ServeHTTP(w, r) { f(w, r) }
```

### ServeMux (Router)

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /api/users/{id}", handler)  // Go 1.22+ patterns
http.ListenAndServe(":8080", mux)
```

---

## Timeout Architecture

```
  Client.Timeout ─── end-to-end (including redirects + body read)
  │
  ├── Transport.TLSHandshakeTimeout
  ├── Transport.ResponseHeaderTimeout
  ├── Transport.ExpectContinueTimeout
  └── net.Conn.SetDeadline (read/write)

  ⚠️  net/http has NO DEFAULT TIMEOUTS.
  Always set them explicitly on both client and server.
```

---

## Making REST Calls

```go
client := &http.Client{Timeout: 10 * time.Second}  // reuse this!

req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
req.Header.Set("Content-Type", "application/json")

resp, err := client.Do(req)
if err != nil { return err }
defer resp.Body.Close()
io.Copy(io.Discard, resp.Body)  // drain for connection reuse

var result MyStruct
json.NewDecoder(resp.Body).Decode(&result)
```

---

## Serving a REST API

```go
func handleUsers(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(users)
}
```

**Note:** `io/ioutil` is deprecated since Go 1.16. Use `io.ReadAll`.

---

## JSON Quick Reference

### Unmarshal into `interface{}` produces:

| JSON | Go Type |
|------|---------|
| number | `float64` (precision loss for large int64!) |
| string | `string` |
| boolean | `bool` |
| null | `nil` |
| object | `map[string]interface{}` |
| array | `[]interface{}` |

### `omitempty` Gotchas

- `*time.Time` with `omitempty` → omits nil. Good.
- `time.Time` with `omitempty` → zero value is `"0001-01-01T00:00:00Z"`, **not omitted**. Bad.

---

## Key Standard Library Packages

| Package | Purpose |
|---------|---------|
| `net/http` | HTTP client and server |
| `encoding/json` | JSON marshal/unmarshal |
| `io` | Core interfaces: `Reader`, `Writer`, `Closer` |
| `bytes` | Byte buffer operations |
| `time` | Durations, timers, tickers |
| `context` | Cancellation, deadlines, request-scoped values |

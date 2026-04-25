# HTTP Timeout Tests Design

**Date:** 2026-04-25
**Branch:** config-http-timeout
**Scope:** Add structural and behavioral unit tests to verify HTTP timeouts introduced in the `config-http-timeout` branch.

---

## Background

The `config-http-timeout` branch adds 60-second HTTP timeouts to several clients that previously had none:

- `appvalidate/coastal_secrets.go` — new `http.Client` per request with `coastalHTTPTimeout`
- `appvalidate/palm_tree_secrets.go` — existing `http.Client` in constructor gains `palmTreeHTTPTimeout`
- `client/config.go` — new `DefaultHTTPTimeout` constant + `Timeout` field on `Config`
- `platform/client.go` — uses `cfg.Config.Timeout` instead of `http.DefaultClient`
- `oauth/token/source.go` — sets `client.DefaultHTTPTimeout` on oauth2-managed HTTP client

---

## Production Code Changes

To make behavioral tests practical (avoiding 60-second sleeps), add an injectable `HTTPTimeout` field to both appvalidate config structs.

### `appvalidate/CoastalSecretsConfig`

```go
type CoastalSecretsConfig struct {
    // ... existing fields ...
    HTTPTimeout time.Duration `envconfig:"TIDEPOOL_COASTAL_HTTP_TIMEOUT"`
}
```

In `GetSecret`, resolve the timeout:
```go
timeout := c.Config.HTTPTimeout
if timeout == 0 {
    timeout = coastalHTTPTimeout
}
res, err := (&http.Client{Timeout: timeout}).Do(req)
```

### `appvalidate/PalmTreeSecretsConfig`

```go
type PalmTreeSecretsConfig struct {
    // ... existing fields ...
    HTTPTimeout time.Duration `envconfig:"TIDEPOOL_PALM_TREE_HTTP_TIMEOUT"`
}
```

In `NewPalmTreeSecrets`, resolve the timeout:
```go
timeout := cfg.HTTPTimeout
if timeout == 0 {
    timeout = palmTreeHTTPTimeout
}
return &PalmTreeSecrets{
    Config: cfg,
    client: &http.Client{Transport: tr, Timeout: timeout},
}, nil
```

**No other production code changes.** `platform/client.go` already uses `cfg.Config.Timeout` which is configurable.

---

## Test Structure

All tests use **Ginkgo v2 + Gomega**, matching the codebase convention. Behavioral tests use `httptest.NewServer` with a handler that sleeps longer than the test timeout (e.g. sleep 50ms, timeout 5ms).

### 1. `client/config_test.go` — extend existing file

**Structural tests only:**
- `DefaultHTTPTimeout` equals `60 * time.Second`
- `NewConfig().Timeout` equals `DefaultHTTPTimeout`

### 2. `platform/client_test.go` — extend existing file

**Structural:**
- `NewClientWithErrorResponseParser` with a config whose `Timeout` is set to a known value produces an `http.Client` with that timeout (verified by attempting a request against a slow server and observing the timeout fires at the right duration)

**Behavioral:**
- Start `httptest.Server` with handler that sleeps 50ms
- Create `platform.Config` with `Config.Timeout = 5ms`
- Make a request via the client
- Assert the error contains a timeout/deadline exceeded message

### 3. `appvalidate/coastal_secrets_test.go` — new file

Needs a new suite file `appvalidate/appvalidate_suite_test.go` if one doesn't exist.

**Structural:**
- `coastalHTTPTimeout` constant equals `60 * time.Second` (tested via `CoastalSecretsConfig{}.HTTPTimeout` fallback behaviour — constant is unexported so tested indirectly through a real request with zero config timeout)

**Behavioral:**
- Start `httptest.Server` with handler that sleeps 50ms
- Construct `CoastalSecrets` with `CoastalSecretsConfig{HTTPTimeout: 5ms, ...}`
- Call `GetSecret` with valid-shaped (but fake) partner data pointing at the test server
- Assert error wraps a timeout/deadline exceeded error

### 4. `appvalidate/palm_tree_secrets_test.go` — new file

Same pattern as coastal.

**Structural:**
- `palmTreeHTTPTimeout` equals `60 * time.Second` (indirect, via zero-config default)

**Behavioral:**
- Start `httptest.Server` with handler that sleeps 50ms
- Construct `PalmTreeSecrets` with `PalmTreeSecretsConfig{HTTPTimeout: 5ms, ...}`
- Call a method that issues an HTTP request
- Assert timeout error

### 5. `oauth/token/source_test.go` — new file

**Structural only** (behavioral requires deep mocking of the oauth2 token exchange internals — not worth the complexity):
- Assert `client.DefaultHTTPTimeout == 60 * time.Second`
- The assignment `httpClient.Timeout = client.DefaultHTTPTimeout` in `source.go` is covered by the constant test + code review

---

## Test Helpers

Slow-server helper (inline in test files, following `alerts/client_test.go` pattern):

```go
func slowServer(t GinkgoTInterface, delay time.Duration) *httptest.Server {
    s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(delay)
        w.WriteHeader(http.StatusOK)
    }))
    t.Cleanup(s.Close)
    return s
}
```

---

## Error Assertion Pattern

Go HTTP timeout errors satisfy `os.IsTimeout` or contain "context deadline exceeded" / "i/o timeout". Use:

```go
Expect(err).To(MatchError(ContainSubstring("deadline exceeded")))
// or
var netErr net.Error
Expect(errors.As(err, &netErr)).To(BeTrue())
Expect(netErr.Timeout()).To(BeTrue())
```

---

## Out of Scope

- `services/tools/tapi/api/api.go` — no existing tests, adding a Ginkgo suite for this CLI tool package is disproportionate effort for one timeout constant
- Changing the 60-second default value
- Adding timeouts to any other HTTP clients not touched by this branch

# deterministic

Package `deterministic` provides deterministic implementations of functions commonly replaced during testing, enabling reliable and reproducible test execution.

It offers drop-in replacements for non-deterministic sources like `crypto/rand.Read`, `time.Now`, `slog` timestamps, and OpenTelemetry trace/span IDs — making tests fully reproducible.

## Documentation

Full API documentation and examples are available on [pkg.go.dev](https://pkg.go.dev/github.com/selesy/deterministic).

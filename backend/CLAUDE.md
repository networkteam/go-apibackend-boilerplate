# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code Search Preferences

IMPORTANT: Always use mcp-gopls for code search instead of grep/glob:
- For finding symbol references: use `mcp__mcp-gopls__references`
- For finding definitions: use `mcp__mcp-gopls__definition`
- DO NOT use grep or glob tools for Go code search if mcp-gopls is available and can be used

Tips for using mcp-gopls:

1. Use fully qualified names for symbols (e.g., myvendor.mytld/myproject/backend/domain/model.User)
2. For methods, use the format Type.Method (e.g., User.IsActive)
3. For packages, use the full import path
4. Line and column numbers are 1-indexed when using hover or rename

## Go Code Style

### Error Handling

#### Multi-line error handling
We prefer explicit, multi-line error handling over inline error assignment. This makes function calls more obvious and improves readability.

**Preferred:**
```go
err := foo()
if err != nil {
    return err
}
```

**Avoid:**
```go
if err := foo(); err != nil {
    return err
}
```

The multi-line approach makes the `foo()` call more prominent and easier to spot when reading code.

#### Don't log and return errors
NEVER log an error and then return it. This causes duplicate logging as the error will be logged again by the caller. Either handle the error (log it and recover) OR return it for the caller to handle.

**Wrong:**
```go
if err != nil {
    logger.ErrorContext(ctx, "Something failed", "error", err)
    return err  // BAD: Error will be logged again by caller
}
```

**Correct - Return with context:**
```go
if err != nil {
    return errors.Wrap(err, "something failed")  // Add context and return
}
```

**Correct - Handle locally (don't return):**
```go
if err != nil {
    logger.ErrorContext(ctx, "Non-critical operation failed, continuing", "error", err)
    // Continue execution without returning error
}
```

There's only one exception: for security relevant errors that are only presented to the user we also want to log a warning.

### Database Conventions

**Primary keys:** Use UUID v7 for all primary keys (time-sortable, globally unique).

**Timestamps:** All tables should have `created_at` and `updated_at` columns, managed automatically.

**Filtering:** Always filter at the database level, not in Go code. When implementing finders, pass filter criteria to the repository layer instead of fetching all records and filtering in memory.

**Multi-tenancy:** Data is typically scoped to an organisation/tenant. Always include organisation filtering in queries - never return data across tenants.

### Writing tests

Prefer table driven tests and structure tests to use `t.Run` for subtests.

**When to use test snapshots with cupaloy.SnapshotT?**

We don't overuse this feature. We only add roughly one snapshot per test for all cases and use individual assertions for the rest of the cases.

**How to write assertion messages?**

Never repeat expected values in assertion messages, but rather describe the expected behaviour (e.g. "values should be changed" instead of "value should be 200").

## Development Setup

This is a Go backend application that uses Devbox (Nix-based) for dependency management. No Docker is required.

### Common Commands

IMPORTANT!: NEVER build the project. Run tests instead.
You can run `go vet ./...` for a quick compile check.

*Testing*

Always run all tests at least once before finishing your work:

```bash
go test ./... -failfast # Run all tests, stop on first failure
```

Depending on the changed files we can run only specific tests during development:

```bash
go test ./api/graph/... # GraphQL API tests
go test ./domain/... # Domain tests
```

NEVER run individual test files directly, always run a test package (or recursive packages) and use filters.

To update cupaloy snapshots in tests (if certain):

```bash
UPDATE_SNAPSHOTS=true go test ./...
```

**Linting:**

Always lint all files:

```bash
devbox run backend:lint
```

**Development Server:**

The user starts all required services with process-compose using `devbox services up`.

## Code Generation

IMPORTANT: Never edit generated files directly. They will be overwritten.

**GraphQL (gqlgen):**

```bash
go generate gqlgen.go
```

- Edit `.graphqls` schema files in `/api/graph/`
- Generated files: `generated.go`, `models_gen.go`
- Resolver stubs are generated once, then manually implemented

**Persistence mappings (construct):**

```bash
go generate ./persistence/repository/mappings.go
```

- Edit model structs with `construct:` tags in `/persistence/repository/`
- Generated file: `mappings_gen.go`

### Architecture Overview

**Main Components:**
- `/api/` - GraphQL and HTTP API layer with resolvers, handlers and middlewares
- `/domain/` - Core business logic with commands + handlers, queries + finders, models and other types (CQRS architecture)
- `/persistence/` - Database layer with repositories and migrations
- `/security/` - Authentication and authorization
- `/cli/ctl/` - CLI commands for running the server or administrative tasks

**Key Patterns:**
- Command/Query separation in domain layer
- Repository pattern for data access with qrb as a query builder
- GraphQL with gqlgen for API
- Structured logging with slog

**Dependencies:**
- PostgreSQL

**Testing:**
- Uses standard Go testing with `go test ./...`
- Test fixtures in `/test/fixtures/`
- GraphQL testing utilities in `/test/graphql/`
- Database test utilities in `/test/db/`

### CLI Tool

The main CLI tool is located at `./cli/ctl/main.go` and provides commands for:
- Database migrations and setup
- Fixture import
- Server operations

## Testable Time

Always use `types.TimeSource` interface instead of `time.Now()` directly. This allows for deterministic testing with fixed timestamps.

**In production/application code:**
- Accept `types.TimeSource` as a dependency
- Use `timeSource.Now()` instead of `time.Now()`

**In tests:**
- Use `test.FixedTime()` to get a fixed timestamp
- Use `test.MustFixedTimeSource(isoTime)` for specific timestamps
- The `FixedTimeSource` type supports `Add()` and `AddDate()` for time manipulation

## Go version and dependencies

The project uses Go 1.25 and the dependencies are managed with Devbox (Nix-based). The `devbox.json` file contains all required dependencies.
Be sure to make use of the latest go packages like slices and maps.

For UUID handling we use `github.com/gofrs/uuid` package. Do not introduce other UUID packages.

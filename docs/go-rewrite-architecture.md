# Go Rewrite Architecture Decision

Date: 2026-04-15

## Decision

Implement the rewrite with the Go standard library plus a thin internal command
layer:

- A small manual parser for top-level flags and the `search` subcommand so the
  CLI can preserve the bash script's interleaved `source ... --flag ...`
  behavior exactly
- `os/exec` for `git clone`
- `net/http` and `encoding/json` for Serper, Scrapedo, and playbooks requests
- `path/filepath`, `os`, and `io` for skill discovery and file copying

Do not use Cobra or another external CLI framework.

## Why this is the right fit

### Documented facts

1. Go's official `flag` package says `FlagSet` exists specifically so callers
   can define independent sets of flags "such as to implement subcommands in a
   command-line interface."
   Source: https://pkg.go.dev/flag

2. The same official `flag` docs also say parsing stops just before the first
   non-flag argument, which means it does not preserve the current bash
   script's interleaved `url --dry-run other-url` behavior by default.
   Source: https://pkg.go.dev/flag

3. Go's official `os/exec` package intentionally does not invoke the system
   shell or expand globs, which makes `git clone --depth 1 ...` execution more
   predictable and safer than reconstructing shell strings.
   Source: https://pkg.go.dev/os/exec

4. Cobra's official README says it provides automatic help generation, help-flag
   recognition, nested subcommands, suggestions, shell completion, and other
   command-framework behavior through Cobra and `pflag`.
   Source: https://github.com/spf13/cobra

### Inference from those sources plus this repo's constraints

- The parity target has one real subcommand (`search`) and otherwise custom,
  hand-written help text already frozen in `testdata/parity/`.
- Because the help text, argument ordering, and stderr/stdout split are part of
  the compatibility target, the right move is a tiny hand-written parser plus
  standard-library building blocks, not a framework or the default `flag`
  parsing behavior.
- The standard library keeps the rewrite dependency-light, easier to audit, and
  simpler to install from source with `go install` or `go build`.

## Architecture shape

- `cmd/skill-dl/main.go`: process entrypoint
- `internal/cli`: argument parsing, dispatch, output writers, exit codes
- `internal/download`: source ingestion, repo grouping, clone orchestration,
  skill discovery, copy logic
- `internal/search`: search keyword validation, HTTP calls, aggregation, table
  rendering
- `internal/install` is not needed as a Go package; installation remains a repo
  concern via build/install script plus README instructions

## Rejected option

### Cobra

Rejected because its documented feature set is optimized for richer command
trees than this tool needs. For this repo, those defaults increase the risk of
behavioral mismatch with the captured bash contract.

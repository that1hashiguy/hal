# HAL Health Command Spec

## Base Command
- Command: `hal health`
- Purpose: Manage the `hal-health` runtime container that serves live HAL ecosystem state
- Default behavior: prints help when no subcommand is provided

## Subcommands
- `hal health create`
  - Create (or recreate) the `hal-health` container
  - Inspects the current ecosystem state and serves it as a JSON + HTML health API on `hal-net`

- `hal health update`
  - Refresh the `hal-health` snapshot
  - Re-inspects the ecosystem and replaces the running container with fresh state

- `hal health delete`
  - Remove the `hal-health` container

## Notes
- The `hal-health` container is a lightweight API server consumed by `hal plus` and other internal consumers to display live ecosystem state.
- The API is reachable from within `hal-net` at `http://hal-health:<port>/api/status`.
- `_serve` is a hidden internal subcommand used by the container entrypoint — it is not intended for direct use.
- Supports `--dry-run`: logs what would happen without mutating any state.

## Example
```bash
# Create the hal-health container
hal health create

# Refresh the snapshot after deploying new services
hal health update

# Remove the container
hal health delete
```

## Sources
- Namespace and subcommands: `cmd/health/health.go`

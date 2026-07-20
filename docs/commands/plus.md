# HAL Plus Command Spec

## Base Command
- Command: `hal plus`
- Purpose: Manage the HAL Plus web UI runtime (chat + RAG frontend backed by a local Ollama LLM)
- Default behavior: runs `hal plus status`

## Subcommands
- `hal plus create`
  - Create (or recreate) the HAL Plus container runtime
  - Starts `hal-plus`, `hal-mcp`, and optionally `hal-qdrant` containers

- `hal plus status`
  - Show runtime status for HAL Plus, the embedded MCP sidecar, and the Qdrant container

- `hal plus delete`
  - Remove HAL Plus runtime containers (`hal-plus`, `hal-mcp`, `hal-qdrant`)

## Flags (`hal plus create`)
```text
    --plus-image string         HAL Plus container image name
    --plus-tag string           HAL Plus container image tag
    --plus-mcp-image string     HAL MCP container image name (must exist locally)
    --plus-mcp-tag string       HAL MCP container image tag
    --pull                      Force pull latest GHCR images before starting (no-op for localhost/ images)
    --port int                  Host port for HAL Plus UI
    --model string              Ollama model to use for HAL Plus chat
    --model-config string       Optional Modelfile path to build a HAL-managed Ollama model on the host
    --ollama-host-url string    Host-side Ollama URL used for preflight checks
    --ollama-base-url string    Container-side OLLAMA_BASE_URL override (defaults by engine)
    --keep-alive string         Ollama model keep-alive duration (e.g. 10m or 0)
    --rag string                Retrieval backend: 'qdrant' (pre-seeded vector container) or 'local' (in-process)
    --embed-model string        Ollama embedding model used for Qdrant query vectors
    --qdrant-image string       Pre-seeded HAL Plus Qdrant container image name
    --qdrant-tag string         Pre-seeded HAL Plus Qdrant container image tag
-h, --help                      help for create
```
- Global flags: `--debug`, `--dry-run`, `--verbose`

## Notes
- HAL Plus requires a running Ollama instance on the host. Install from [ollama.com](https://ollama.com) and pull a model before starting.
- The MCP sidecar (`hal-mcp`) is started alongside HAL Plus and provides tool-call capabilities inside the chat.
- `hal delete` and `hal daisy` also remove HAL Plus containers as part of global teardown.

## Example
```bash
# Start HAL Plus with default settings
hal plus create

# Start with a specific model
hal plus create --model llama3.2

# Start with Qdrant RAG backend
hal plus create --rag qdrant

# Check status
hal plus status

# Remove containers
hal plus delete
```

## Sources
- Namespace and lifecycle commands: `cmd/plus/plus.go`, `cmd/plus/create.go`, `cmd/plus/status.go`, `cmd/plus/delete.go`

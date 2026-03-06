# CLAUDE.md

This file provides guidance to Kscc (claudexxxxxx.ai/code) when working with code in this repository.

## Project Overview

A task-driven agent tool consisting of:
1. Main agent: receives and analyzes tasks, generates fixed-type tasks
2. Built-in sub-agents: execute tasks based on task type

## Tech Stack

- **Framework**: GoFrame v2 (gf v2.7.1)
- **Go Version**: 1.26
- **Database**: SQLite
- **LLM Gateway**: Maxim Bifrost (multi-provider LLM abstraction)
- **Data Formats**: JSONL, YAML
- **Logging/Config**: glog, gcfg (GoFrame built-in)

## Common Commands

```bash
# Build the project
make build

# Run the server (default port :8000)
./temp/linux_amd64/main

# Or run directly with go
go run main.go

# Generate controller code from API definitions
make ctrl

# Generate DAO/DO/Entity for database
make dao

# Generate service interfaces
make service

# Install/update GoFrame CLI
make cli
```

## Architecture

This project follows GoFrame's standard project layout:

```
api/                        # API definitions (request/response structs)
  hello/v1/                 # Hello module API v1
internal/                   # Private application code
  cmd/                      # CLI entry point and command definitions
  config/                   # Configuration loader and types
  consts/                   # Global constants
  controller/               # HTTP handlers (auto-generated stubs)
  llm/                      # LLM client abstraction
    bifrost/                # Bifrost LLM gateway implementation
  middleware/               # HTTP middleware
  service/                  # Business logic interfaces
manifest/                   # Configuration and deployment manifests
  config/config.yaml        # Application config
hack/                       # Build scripts and CLI config
```

### Key Architectural Patterns

**Configuration Loading** (`internal/config/`):
- Config loaded from `manifest/config.yaml` at startup via `FileLoader`
- Supports environment variable overrides (prefix `AGENT_`)
- LLM config can reference separate Bifrost config file

**LLM Integration** (`internal/llm/`):
- Abstract `Client` interface for chat completion
- Bifrost implementation provides multi-provider support (OpenAI, Anthropic, etc.)
- Supports tool calling and fallback routing
- API keys loaded from environment variables (e.g., `OPENAI_API_KEY`)

**Routing** (`internal/cmd/cmd.go`):
- Routes registered using router groups with middleware
- `HandlerResponse` middleware wraps responses uniformly

## Configuration

Main config: `manifest/config/config.yaml`
- Server: `:8000`, Swagger at `/swagger`, OpenAPI at `/api.json`
- Database: MySQL (configurable)
- LLM: Provider/model settings, Bifrost config path

Bifrost config (referenced by `llm.bifrost_config_path`):
- Defines providers, models, API keys, routing strategy
- Supports fallback chains and load balancing

## API Development

1. Define request/response structs in `api/<module>/v1/`
2. Run `make ctrl` to generate controller stubs
3. Implement business logic in `internal/controller/<module>/`
4. Run `make service` to generate service interfaces if needed
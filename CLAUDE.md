# CLAUDE.md
## Project Overview

A task-driven agent tool consisting of:
1. Main agent: receives and analyzes tasks, generates fixed-type tasks
2. Built-in sub-agents: execute tasks based on task type

## Tech Stack

- **Framework**: GoFrame v2 (gf v2.7.1)
- **Go Version**: 1.26
- **Database**: SQLite
- **Data Formats**: JSONL, YAML
- **Logging/Config**: glog, gcfg (GoFrame built-in)

## Architecture

This project follows GoFrame's standard project layout:

```
api/                    # API definitions (request/response structs)
  hello/                # Module: hello
    v1/                 # API version 1
internal/               # Private application code
  cmd/                  # CLI entry point and command definitions
  consts/               # Global constants
  controller/           # HTTP handlers (auto-generated stubs)
    hello/              # Controller for hello module
  packed/               # Packed resources (generated)
  service/              # Business logic interfaces
manifest/               # Configuration and deployment manifests
  config/               # Application config (config.yaml)
utility/                # Utility functions
hack/                   # Build scripts and CLI config
  hack.mk               # Core makefile targets
  hack-cli.mk           # CLI installation
  config.yaml           # gf CLI configuration
```

### Routing
Routes are registered in `internal/cmd/cmd.go` using router groups with middleware.

## Configuration

- Application config: `manifest/config/config.yaml`
- Build/CLI config: `hack/config.yaml`
- Server runs on `:8000` by default
- Swagger UI available at `/swagger`
- OpenAPI spec at `/api.json`
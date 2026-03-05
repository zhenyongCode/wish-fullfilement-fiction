# File Structure Tree

This document maintains the project's file structure as single source of truth.

```
/                               – Project root
  main.go                       – Application entry point
  go.mod                        – Go module definition
  go.sum                        – Go dependencies checksum
  Makefile                      – Build targets entry
  README.MD                     – Project readme
  CLAUDE.md                     – AI coding assistant guidance

/api                            – API definitions (request/response)
  /hello                        – Hello module API
    hello.go                    – Interface definition (auto-generated)
    /v1                         – API version 1
      hello.go                  – HelloReq/HelloRes structs

/internal                       – Private application code
  /cmd                          – CLI commands
    cmd.go                      – Main command and router setup
  /consts                       – Global constants
    consts.go                   – Constant definitions
  /controller                   – HTTP handlers
    /hello                      – Hello module controller
      hello.go                  – Package placeholder
      hello_new.go              – Controller constructor (auto-generated)
      hello_v1_hello.go         – Hello endpoint implementation
  /dao                          – Data access objects
  /logic                        – Business logic implementations
    example.go                  – Example service functions (echo, hello, add)
  /model                        – Data models
  /service                      – Business logic interfaces
    api.go                      – Service definitions
  /servicefunc                  – Service function registry and executor
    servicefunc.go              – RegisterFunc, RegisterMethod, ServiceFuncExe

/manifest                       – Configuration and deployment
  /config                       – Application config
    config.yaml                 – Server, logger, database config

/utility                        – Utility functions
  .gitkeep                      – Directory placeholder

/hack                           – Build scripts
  hack.mk                       – Core makefile targets
  hack-cli.mk                   – GoFrame CLI installation
  config.yaml                   – gf CLI configuration

/resource                       – Static resources (placeholder)

/docs                           – Documentation
  progress.md                   – Project history (append-only)
  files.md                      – This file (file structure tree)
```
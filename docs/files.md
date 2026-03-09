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
  bifrost.yaml                  – Bifrost LLM gateway config

/api                            – API definitions (request/response)
  /hello                        – Hello module API
    hello.go                    – Interface definition (auto-generated)
    /v1                         – API version 1
      hello.go                  – HelloReq/HelloRes structs

/internal                       – Private application code
  /cmd                          – CLI commands
    cmd.go                      – Main command and router setup
  /config                       – Configuration (using gcfg)
    config.go                   – Config struct, Load, Get helpers
  /consts                       – Global constants
    consts.go                   – Constant definitions
  /controller                   – HTTP handlers
    /hello                      – Hello module controller
      hello.go                  – Package placeholder
      hello_new.go              – Controller constructor (auto-generated)
      hello_v1_hello.go         – Hello endpoint implementation
  /llm                          – LLM client abstraction layer
    client.go                   – ChatRequest/Response types, Client interface
    /bifrost                    – Bifrost LLM gateway implementation
      client.go                 – Bifrost client with multi-provider support
      config.go                 – Bifrost config types and LoadConfig
    /prompt                     – Prompt management system
      prompt.go                 – Prompt structures (待实现)
    /skills                     – Skills management system
      skills.go                 – Skill structures (待实现)
    /tools                      – Tool execution system
      tools.go                  – Tool interface definition
      hub.go                    – Tool registration and execution center
      read_file_tool.go         – File read tool implementation
      bash_tool.go              – Bash command tool implementation
  /logic                        – Business logic implementations
    example.go                  – Service functions (echo, hello, add, chat)
  /middleware                   – HTTP middleware
    resp.go                     – HandlerResponse middleware
  /service                      – Business logic interfaces
    api.go                      – Service definitions
    chat.go                     – ChatService implementation
  /servicefunc                  – Service function registry and executor
    servicefunc.go              – RegisterFunc, RegisterMethod, ServiceFuncExe

/manifest                       – Configuration and deployment
  /config                       – Application config
    config.yaml                 – Server, logger, database, LLM path config

/utility                        – Utility functions
  .gitkeep                      – Directory placeholder

/hack                           – Build scripts
  hack.mk                       – Core makefile targets
  hack-cli.mk                   – GoFrame CLI installation
  config.yaml                   – gf CLI configuration

/resource                       – Static resources (placeholder)

/docs                           – Documentation
  progress.md                   – Project history (append-only)
  files.md                      – 文件结构树
  prompt-skills-design.md       – Prompt & Skills 加载实现方案
```
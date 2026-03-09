# Project History Log

This file maintains an append-only history of what this AI+human team has done.

---

#### `[2026-03-05 16:30] Initialize project documentation`
- **Context**: User requested CLAUDE.md creation for future AI instances.
- **Changes**:
  - `CLAUDE.md`: created with project overview, build commands, and architecture documentation.
  - `docs/progress.md`: created for project history tracking.
  - `docs/files.md`: created for file structure documentation.
- **Decisions**:
  - Follow GoFrame standard project layout conventions.
- **Next steps**:
  - Implement main agent logic for task analysis.
  - Implement sub-agents for task execution.
- **Actor**: Claude Code

---

#### `[2026-03-05 17:00] Implement ServiceFunc registration and execution`
- **Context**: User requested ServiceFuncExe implementation for dynamic service function registration and execution in hello API.
- **Changes**:
  - `internal/servicefunc/servicefunc.go`: created service function registry with RegisterFunc, RegisterMethod, and ServiceFuncExe.
  - `internal/logic/example.go`: created example service functions (echo, hello, add) demonstrating registration patterns.
  - `internal/controller/hello/hello_v1_hello.go`: modified to call ServiceFuncExe with func name and params from request.
  - `main.go`: added import for logic package to trigger init() registration.
- **Decisions**:
  - Use map-based registry for dynamic function lookup.
  - Support both function and method registration via reflection.
  - Use g.Map as universal parameter/result type.
- **Next steps**:
  - Add more service functions as needed.
  - Implement task-type based agent dispatch.
- **Actor**: Claude Code

---

#### `[2026-03-06 10:00] Update CLAUDE.md with comprehensive project documentation`
- **Context**: User requested /init to analyze codebase and improve CLAUDE.md for future AI instances.
- **Changes**:
  - `CLAUDE.md`: enhanced with common commands, LLM integration details, configuration patterns.
  - `docs/files.md`: updated to include new directories (config/, llm/, middleware/).
- **Decisions**:
  - Added Maxim Bifrost as key LLM gateway dependency.
  - Documented environment variable override pattern (AGENT_* prefix).
- **Next steps**:
  - Complete Chat service implementation in internal/service/chat.go.
  - Add bifrost.yaml configuration file for LLM providers.
- **Actor**: Claude Code, 全能的ai统治者

---

#### `[2026-03-06 10:30] Refactor config management using GoFrame gcfg`
- **Context**: User requested to simplify config management with gcfg and make bifrost config independent.
- **Changes**:
  - `internal/config/config.go`: simplified to use gcfg, removed manual YAML parsing and Bifrost types.
  - `internal/llm/bifrost/config.go`: created independent config types and LoadConfig function.
  - `internal/llm/bifrost/client.go`: updated to use internal config, added Option pattern for customization.
  - `main.go`: simplified, removed manual config loading (gcfg auto-loads from manifest/config/config.yaml).
  - `bifrost.yaml`: created example bifrost config file.
  - `manifest/config/config.yaml`: simplified, only keeps llm.bifrost_config_path reference.
  - `internal/service/chat.go`: completed ChatService implementation.
- **Decisions**:
  - Use GoFrame gcfg for auto-loading config from manifest/config/config.yaml.
  - Bifrost config in separate file (bifrost.yaml), loaded independently by bifrost module.
  - Option pattern for bifrost client customization (WithProvider, WithModel, WithTimeout).
- **Next steps**:
  - Add unit tests for bifrost config loading.
  - Implement main agent logic.
- **Actor**: Claude Code, 全能的ai统治者

---

#### `[2026-03-06 14:00] Register ChatService with servicefunc pattern`
- **Context**: User requested to register chat service methods using servicefunc.RegisterMethod pattern.
- **Changes**:
  - `internal/logic/chat.go`: created ChatLogic wrapper with Chat/ChatStream methods matching servicefunc signature.
  - Lazy initialization pattern for ChatService to avoid init-time errors.
- **Decisions**:
  - Use adapter pattern to wrap ChatService.Chat(ctx, *ChatRequest) to ChatLogic.Chat(ctx, g.Map).
  - Register methods lazily on first call to ensure bifrost config is loaded.
- **Next steps**:
  - Implement streaming chat support.
  - Add unit tests for chat logic.
- **Actor**: Claude Code

---

#### `[2026-03-06 14:30] Simplify ChatService registration`
- **Context**: User clarified to register service methods directly, no separate ChatLogic needed.
- **Changes**:
  - `internal/service/chat.go`: added Exec method with servicefunc signature.
  - `internal/logic/example.go`: added chat registration using RegisterFunc with lazy init.
  - Deleted `internal/logic/chat.go` (no longer needed).
- **Decisions**:
  - ChatService.Exec method handles g.Map params conversion.
  - Lazy initialization via wrapper function to avoid config loading issues.
- **Next steps**:
  - Test chat endpoint with curl.
- **Actor**: Claude Code

---

#### `[2026-03-09 14:00] Prompt & Skills 加载实现方案设计`
- **Context**: 用户要求分析当前 agent、llm 模块实现，设计后续 prompt 和 skills 加载方案。
- **Changes**:
  - `docs/prompt-skills-design.md`: 创建完整设计方案文档。
  - `docs/files.md`: 更新文件结构树，补充 llm/tools 和 docs 目录详情。
- **Decisions**:
  - Prompt 系统作为 Agent 基础行为定义，支持模板变量和运行时热更新。
  - Skills 系统作为任务级动态能力增强，支持触发器匹配和优先级排序。
  - 两系统独立但协同工作：Prompt 提供基础，Skills 按需追加增强。
- **Next steps**:
  - Phase 1: 实现 Prompt 系统（prompt.go, manager.go, loader.go, render.go）。
  - Phase 2: 实现 Skills 系统（skills.go, hub.go, matcher.go, loader.go）。
  - Phase 3: 增强功能（热更新、API 接口、遥测指标）。
- **Actor**: Claude Code, 全能的ai统治者
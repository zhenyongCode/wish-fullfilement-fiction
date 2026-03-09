package consts

import "time"

const (
	DefaultConfigPath = "./manifest/config.yaml"
	DefaultWorkspace  = "."
	DefaultOutput     = "text"
	OutputJSON        = "json"
	OutputText        = "text"
)

const (
	DefaultLogLevel = "info"
	LogLevelDebug   = "debug"
	LogLevelInfo    = "info"
	LogLevelWarn    = "warn"
	LogLevelError   = "error"

	DefaultLogFormat = "json"
	LogFormatJSON    = "json"
	LogFormatConsole = "console"

	DefaultLogFile = "logs/agent.log"
)

const (
	DefaultCommandTimeout = 30 * time.Second
	DefaultLLMTimeout     = 30 * time.Second
	DefaultMaxRounds      = 30
	DefaultToolTimeout    = 120 * time.Second
)

const (
	ModuleConfig  = "config"
	ModuleLogging = "logging"
	ModuleState   = "state"
	ModuleCLI     = "cli"
	ModuleAgent   = "agent"
)

const (
	OperationAgentLoop = "agent.loop"

	StepRoundStart   = "round_start"
	StepLLMRequest   = "llm_request"
	StepToolDispatch = "tool_dispatch"
	StepRoundEnd     = "round_end"
)

const (
	StopReasonToolUse = "tool_use"
	ToolNameBash      = "bash"
	ToolNameReadFile  = "read_file"
)

const (
	StatusStart  = "start"
	StatusStep   = "step"
	StatusFinish = "finish"
)

const (
	ContextTraceIDKey   = "trace_id"
	ContextSessionIDKey = "session_id"
	ContextTaskIDKey    = "task_id"
	ContextAgentNameKey = "agent_name"
)

const (
	BuildInAgentTranslation = "translation"
)

var BuildInAgents = []string{
	BuildInAgentTranslation,
}

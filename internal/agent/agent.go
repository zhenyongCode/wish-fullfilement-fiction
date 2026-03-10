package agent

import (
	"context"
	"github.com/gogf/gf/v2/util/guid"
	"wish-fullfilement-fiction/internal/consts"
	"wish-fullfilement-fiction/internal/llm"
	"wish-fullfilement-fiction/internal/llm/prompt"
	"wish-fullfilement-fiction/internal/llm/skills"
	"wish-fullfilement-fiction/internal/llm/tools"
)

type Agent struct {
	name      string // Agent 的名称，用于区分不同的 Agent 实例
	loop      *Loop
	skillHub  *skills.Hub // 技能中心，管理可用的技能实例
	workspace string
	messages  []llm.ChatMessage // Agent 当前的对话消息列表，可以包含系统提示、用户输入、助手回复等
	sessionId string
}

func NewAgent(name string, llmClient llm.Client) *Agent {
	if name == "" {
		name = "default-agent"
	}
	if agent, exists := agents[name]; exists {
		return agent
	}
	agent := &Agent{name: name}
	agent.loop = NewLoop(llmClient, 10) // TODO: maxRounds 可以根据需要调整
	agent.loop.toolHub = tools.NewHub(consts.DefaultCommandTimeout)
	agent.skillHub = skills.NewHub()
	agents[name] = agent

	return agent
}

func (a *Agent) Name() string {
	return a.name
}

// RegisterTool agent tools 注册
func (a *Agent) RegisterTool(tool tools.Tool) {
	err := a.loop.toolHub.Register(tool)
	if err != nil {
		return
	}
}

// agent skills 注册
// func (a *Agent) RegisterSkill(skill Skill) {
// 	a.loop.skillHub.Register(skill)
// }

// agent system prompt 注册
// func (a *Agent) SetSystemPrompt(prompt string) {
// 	a.loop.systemPrompt = prompt
// }

// ResetSession session
func (a *Agent) ResetSession(sessionId string) {
	if sessionId == "" {
		sessionId = a.generateSessionId()
	}
	a.sessionId = sessionId
}
func (a *Agent) generateSessionId() string {
	return guid.S([]byte(a.name))
}

func (a *Agent) Run(ctx context.Context, task string) (string, error) {
	// TODO: 根据 task 构建初始的 messages，可以包含系统提示、用户输入等
	if a.sessionId == "" {
		a.ResetSession("")
	}
	if len(a.messages) == 0 {
		p := prompt.NewPrompt()
		systemPrompt := p.GetSystemPrompt(a.name)
		if a.workspace != "" {
			sm := a.skillHub.BuildSummary(a.workspace)
			systemPrompt += "The following skills extend your capabilities. To use a skill, read its SKILL.md file using the read_file tool.\n\n" + sm
		}
		a.messages = []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
		}
	}

	a.messages = append(a.messages, llm.ChatMessage{Role: "user", Content: task})

	return a.loop.Execute(ctx, a.messages)
}

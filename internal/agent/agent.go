package agent

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
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

	sessionMessages map[string][]llm.ChatMessage // 对应会话的消息列表，key 是 sessionId
	sessionId       string
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
	a.sessionMessages = make(map[string][]llm.ChatMessage)
	a.sessionMessages[sessionId] = []llm.ChatMessage{}
}
func (a *Agent) generateSessionId() string {
	return guid.S([]byte(a.name))
}

func (a *Agent) GetSessionMessages() []llm.ChatMessage {
	return a.sessionMessages[a.sessionId]
}
func (a *Agent) SetSessionMessages(messages []llm.ChatMessage) {
	a.sessionMessages[a.sessionId] = messages
}
func (a *Agent) AppendSessionMessages(messages []llm.ChatMessage) {
	a.sessionMessages[a.sessionId] = append(a.sessionMessages[a.sessionId], messages...)
}

func (a *Agent) Run(ctx context.Context, task string) (string, error) {
	if a.sessionId == "" {
		a.ResetSession("")
	}
	if len(a.GetSessionMessages()) == 0 {
		p := prompt.NewPrompt()
		systemPrompt := p.GetSystemPrompt(a.name)
		if a.workspace != "" {
			sm := a.skillHub.BuildSummary(a.workspace)
			systemPrompt += "The following skills extend your capabilities. To use a skill, read its SKILL.md file using the read_file tool.\n\n" + sm
			g.Log().Debugf(ctx, systemPrompt)
		}
		a.SetSessionMessages([]llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
		})
	}
	a.AppendSessionMessages([]llm.ChatMessage{
		{Role: "user", Content: task},
	})

	result, updatedMessages, err := a.loop.Execute(ctx, a.GetSessionMessages())
	a.sessionMessages[a.sessionId] = updatedMessages
	return result, err
}

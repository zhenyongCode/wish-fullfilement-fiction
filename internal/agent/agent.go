package agent

import (
	"context"
	"wish-fullfilement-fiction/internal/llm"
)

var agents = make(map[string]*Agent)

func GetAgent(name string) *Agent {
	if agent, exists := agents[name]; exists {
		return agent
	}
	return NewAgent(name, nil) // TODO: llmClient 需要从外部传入
}

func ListAgents() []*Agent {
	result := make([]*Agent, 0, len(agents))
	for _, agent := range agents {
		result = append(result, agent)
	}
	return result
}

type Agent struct {
	name string // Agent 的名称，用于区分不同的 Agent 实例
	loop *Loop
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
	agents[name] = agent

	return agent
}

func (a *Agent) Run(ctx context.Context, task string) (string, error) {
	return a.loop.Execute(ctx, task)
}

package agent

import (
	"context"
	"wish-fullfilement-fiction/internal/consts"
	"wish-fullfilement-fiction/internal/llm/bifrost"
)

// 这是一个内置的 Agent 实现，提供了一个简单的示例，展示如何创建和使用 Agent。

// prompt tools skills 预定义

var agents = make(map[string]*Agent)

func init() {
	initBuiltinAgents()
}
func GetAgent(name string) *Agent {
	if agent, exists := agents[name]; exists {
		return agent
	}
	llmClient, err := bifrost.New(context.Background())
	if err != nil {
		return nil
	}
	return NewAgent(name, llmClient)
}

func ListAgents() []*Agent {
	result := make([]*Agent, 0, len(agents))
	for _, agent := range agents {
		result = append(result, agent)
	}
	return result
}

func initBuiltinAgents() {
	for _, agent := range consts.BuildInAgents {
		initTranslationAgent(agent)
	}
}
func initTranslationAgent(name string) {
	agent := GetAgent(name)

	agent.ResetSession("")

}

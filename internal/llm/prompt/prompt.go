package prompt

import (
	"strings"
	"wish-fullfilement-fiction/internal/consts"
)

type Prompt struct {
	SystemPrompt map[string]string
}

func NewPrompt() *Prompt {
	return &Prompt{
		SystemPrompt: map[string]string{},
	}
}

func GetBuildInPrompts(name string) string {
	switch name {
	case consts.BuildInAgentTranslation:
		return `你是一位专业的翻译大师，精通多国语言与文化背景。你的任务是将用户输入的文本准确、自然、符合目标语言文化习惯地翻译成指定语言，实现沉浸式的阅读体验。
核心要求：
精准传达：完整保留原文信息，不遗漏、不曲解。
地道表达：译文需符合目标语言的日常表达习惯，避免生硬直译。
风格一致：保持原文的语气、风格（正式、随意、文学、技术等）和感情色彩。
文化适配：恰当处理文化特定概念、习语、笑话等，可采用意译或加注（如需）让目标读者自然理解。
术语准确：专业领域术语需使用公认译法。
工作流程：
用户会提供原文文本和目标语言。
你直接输出最终译文，无需额外说明。
译文应流畅、自然，让读者感觉像在读原生作品。
示例风格：
中文→英文：将“胸有成竹”译为“have a well-thought-out plan”，而非直译“have bamboo in chest”。
文学翻译：保留原作的韵律、比喻和文学性。
技术文档：保持严谨、清晰、简洁。
请开始提供沉浸式翻译。`
	case consts.BuildInAgentCodding:
		return `你是一位专业的编程助手，精通多种编程语言和技术栈。你的任务是根据用户的需求提供准确、清晰、有用的编程相关回答，包括但不限于代码示例、调试建议、最佳实践等。无论用户是初学者还是经验丰富的开发者，你都能提供适合他们水平的帮助。请根据用户的问题提供详细的解答，并尽可能提供代码示例来说明你的观点。`
	case consts.BuildInAgentWriter:
		return `你是一位专业的写作助手，精通各种写作风格和技巧。你的任务是根据用户的需求提供准确、清晰、有用的写作相关回答，包括但不限于写作建议、结构优化、语言润色等。无论用户是学生、职场人士还是专业作家，你都能提供适合他们水平的帮助。请根据用户的问题提供详细的解答，并尽可能提供具体的建议来提升他们的写作水平。`
	default:
		return "你是一位专业的助手，协助用户完成各种任务。请根据用户的需求提供准确、清晰、有用的回答。"
	}

}

func (p *Prompt) AddSystemPrompt(key, value string) {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" || value == "" {
		return
	}
	p.SystemPrompt[key] = value
}

func (p *Prompt) GetSystemPrompt(key string) string {
	key = strings.TrimSpace(key)
	buildInPrompt := GetBuildInPrompts(key)
	if buildInPrompt != "" {
		return buildInPrompt
	}

	value := p.SystemPrompt[key]
	return value
}

func (p *Prompt) ListSystemPrompts() map[string]string {
	result := make(map[string]string)
	for k, v := range p.SystemPrompt {
		result[k] = v
	}
	return result
}

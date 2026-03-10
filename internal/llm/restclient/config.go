package restclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// ─────────────────────────────────────────────
//  Config structures
// ─────────────────────────────────────────────

// ModelEntry 单个模型条目，支持按权重随机选择
type ModelEntry struct {
	// Name 模型名称，例如 gpt-4o、glm-4.7-flash
	Name string `yaml:"name"`
	// Weight 权重，越大被选中概率越高（默认 1）
	Weight int `yaml:"weight"`
}

// ProviderConfig 单个 Provider（一个 base_url + 一批模型）
type ProviderConfig struct {
	// Name provider 标识，仅用于日志
	Name string `yaml:"name"`
	// BaseURL OpenAI 兼容接口地址
	BaseURL string `yaml:"base_url"`
	// APIKey 明文 key（优先级低于 APIKeyEnv）
	APIKey string `yaml:"api_key"`
	// APIKeyEnv 从该环境变量读取 API Key
	APIKeyEnv string `yaml:"api_key_env"`
	// Models 该 provider 支持的模型列表；若为空则使用 DefaultModel
	Models []ModelEntry `yaml:"models"`
	// DefaultModel 当 Models 为空时使用的默认模型
	DefaultModel string `yaml:"default_model"`
	// TimeoutStr 超时字符串，例如 "60s"
	TimeoutStr string `yaml:"timeout"`
	// MaxRetries 单 provider 内部重试次数
	MaxRetries int `yaml:"max_retries"`

	timeout time.Duration
}

// FallbackEntry 回退链条目，指向另一个 provider（按 Name 引用）
type FallbackEntry struct {
	// Provider 回退的 provider name（对应 Providers[i].Name）
	Provider string `yaml:"provider"`
	// Model 指定回退使用的模型；为空则由该 provider 自动选择
	Model string `yaml:"model"`
}

// Config REST 客户端顶层配置
type Config struct {
	// Providers 所有可用 provider 列表，第一个为主 provider
	Providers []ProviderConfig `yaml:"providers"`
	// Fallback 全局回退链，主 provider 所有模型均失败后依次尝试
	Fallback []FallbackEntry `yaml:"fallback"`

	// ── 以下为向下兼容的扁平字段（仍支持旧的单 provider 配置）──
	BaseURL      string `yaml:"base_url"`
	APIKey       string `yaml:"api_key"`
	APIKeyEnv    string `yaml:"api_key_env"`
	DefaultModel string `yaml:"default_model"`
	TimeoutStr   string `yaml:"timeout"`
	MaxRetries   int    `yaml:"max_retries"`
}

var (
	ErrBaseURLRequired = errors.New("restclient: base_url is required")
	ErrAPIKeyMissing   = errors.New("restclient: api_key or api_key_env is required")
	ErrModelRequired   = errors.New("restclient: model is required (set default_model or pass via ChatRequest)")
	ErrNoProviders     = errors.New("restclient: no providers configured")
)

// LoadConfig 从 gcfg (manifest/config/config.yaml) 读取 restclient 配置节
//
// 完整配置示例（支持多模型 + 回退链）：
//
//	llm:
//	  restclient:
//	    providers:
//	      - name: glm
//	        base_url: https://open.bigmodel.cn/api/paas/v4
//	        api_key_env: GLM_API_KEY
//	        timeout: 120s
//	        max_retries: 2
//	        models:
//	          - name: glm-4.7-flash
//	            weight: 3
//	          - name: glm-4.6v-flash
//	            weight: 1
//	      - name: openai
//	        base_url: https://api.openai.com/v1
//	        api_key_env: OPENAI_API_KEY
//	        models:
//	          - name: gpt-4o-mini
//	            weight: 1
//	    fallback:
//	      - provider: openai
//	        model: gpt-4o-mini
//
// 兼容旧版单 provider 扁平配置：
//
//	llm:
//	  restclient:
//	    base_url: https://open.bigmodel.cn/api/paas/v4
//	    api_key_env: GLM_API_KEY
//	    default_model: glm-4.7-flash
//	    timeout: 120s
//	    max_retries: 2
func LoadConfig(ctx context.Context) (*Config, error) {
	v := g.Cfg().MustGet(ctx, "llm.restclient")
	if v.IsNil() || v.IsEmpty() {
		return nil, fmt.Errorf("restclient: config section 'llm.restclient' not found")
	}
	cfg := &Config{}
	if err := v.Scan(cfg); err != nil {
		return nil, fmt.Errorf("restclient: parse config failed: %w", err)
	}
	// 兼容：将旧的扁平字段转换为 Providers[0]
	cfg.normalize()
	return cfg, nil
}

// normalize 将旧版扁平配置升级为 Providers 列表
func (c *Config) normalize() {
	if len(c.Providers) > 0 {
		return
	}
	if strings.TrimSpace(c.BaseURL) == "" {
		return
	}
	p := ProviderConfig{
		Name:         "default",
		BaseURL:      c.BaseURL,
		APIKey:       c.APIKey,
		APIKeyEnv:    c.APIKeyEnv,
		DefaultModel: c.DefaultModel,
		TimeoutStr:   c.TimeoutStr,
		MaxRetries:   c.MaxRetries,
	}
	if p.DefaultModel != "" {
		p.Models = []ModelEntry{{Name: p.DefaultModel, Weight: 1}}
	}
	c.Providers = []ProviderConfig{p}
}

// validate 校验配置合法性
func (c *Config) validate() error {
	if len(c.Providers) == 0 {
		return ErrNoProviders
	}
	for i, p := range c.Providers {
		if strings.TrimSpace(p.BaseURL) == "" {
			return fmt.Errorf("restclient: providers[%d].base_url is required", i)
		}
		if _, err := p.resolveAPIKey(); err != nil {
			return fmt.Errorf("restclient: providers[%d] (%s): %w", i, p.Name, err)
		}
		if len(p.Models) == 0 && strings.TrimSpace(p.DefaultModel) == "" {
			return fmt.Errorf("restclient: providers[%d] (%s): models or default_model is required", i, p.Name)
		}
	}
	return nil
}

// resolveAPIKey 解析 provider 最终使用的 API Key
func (p *ProviderConfig) resolveAPIKey() (string, error) {
	if env := strings.TrimSpace(p.APIKeyEnv); env != "" {
		if key := strings.TrimSpace(os.Getenv(env)); key != "" {
			return key, nil
		}
	}
	if key := strings.TrimSpace(p.APIKey); key != "" {
		return key, nil
	}
	return "", ErrAPIKeyMissing
}

// resolveTimeout 解析 provider 超时
func (p *ProviderConfig) resolveTimeout() time.Duration {
	if p.timeout > 0 {
		return p.timeout
	}
	if p.TimeoutStr != "" {
		if d, err := time.ParseDuration(p.TimeoutStr); err == nil && d > 0 {
			p.timeout = d
			return d
		}
	}
	return 60 * time.Second
}

// effectiveModels 返回该 provider 的有效模型列表（至少一个）
func (p *ProviderConfig) effectiveModels() []ModelEntry {
	if len(p.Models) > 0 {
		return p.Models
	}
	if m := strings.TrimSpace(p.DefaultModel); m != "" {
		return []ModelEntry{{Name: m, Weight: 1}}
	}
	return nil
}

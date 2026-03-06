package bifrost

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wish-fullfilement-fiction/internal/consts"

	"github.com/gogf/gf/v2/frame/g"
	"gopkg.in/yaml.v3"
)

// Config Bifrost 配置结构体
type Config struct {
	Global    GlobalConfig              `yaml:"global"`
	Routing   RoutingConfig             `yaml:"routing"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// GlobalConfig 全局配置
type GlobalConfig struct {
	Timeout             string `yaml:"timeout"`
	MaxRetries          int    `yaml:"max_retries"`
	RetryBackoffInitial string `yaml:"retry_backoff_initial"`
	RetryBackoffMax     string `yaml:"retry_backoff_max"`
	LogLevel            string `yaml:"log_level"`
}

// RoutingConfig 路由配置
type RoutingConfig struct {
	Strategy string              `yaml:"strategy"`
	Default  RouteTargetConfig   `yaml:"default"`
	Fallback []RouteTargetConfig `yaml:"fallback"`
}

// RouteTargetConfig 路由目标配置
type RouteTargetConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	Enabled   bool     `yaml:"enabled"`
	BaseURL   string   `yaml:"base_url"`
	APIKeyEnv string   `yaml:"api_key_env"`
	Models    []string `yaml:"models"`
	Weight    int      `yaml:"weight"`
}

var (
	ErrConfigNotFound = errors.New("bifrost config not found")
	ErrNoProviders    = errors.New("no providers configured")
)

// LoadConfig 从指定路径加载 Bifrost 配置
func LoadConfig(ctx context.Context, configPath string) (*Config, error) {
	// 如果未指定路径，尝试从 gcfg 获取
	if configPath == "" {
		configPath = g.Cfg().MustGet(ctx, "llm.bifrost_config_path").String()
	}

	// 如果仍未指定，使用默认路径
	if configPath == "" {
		configPath = consts.DefaultConfigPath
	}

	// 解析相对路径
	if !filepath.IsAbs(configPath) {
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, fmt.Errorf("resolve config path failed: %w", err)
		}
		configPath = absPath
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, configPath)
		}
		return nil, fmt.Errorf("read bifrost config failed: %w", err)
	}

	// 解析 YAML
	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse bifrost yaml failed: %w", err)
	}

	// 验证配置
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if len(cfg.Providers) == 0 {
		return ErrNoProviders
	}
	if strings.TrimSpace(cfg.Routing.Default.Provider) == "" {
		return errors.New("routing.default.provider is required")
	}
	if strings.TrimSpace(cfg.Routing.Default.Model) == "" {
		return errors.New("routing.default.model is required")
	}
	return nil
}

// GetDefaultProvider 获取默认提供商
func (c *Config) GetDefaultProvider() string {
	return strings.TrimSpace(c.Routing.Default.Provider)
}

// GetDefaultModel 获取默认模型
func (c *Config) GetDefaultModel() string {
	return strings.TrimSpace(c.Routing.Default.Model)
}

// GetTimeout 获取超时时间
func (c *Config) GetTimeout() time.Duration {
	if c.Global.Timeout == "" {
		return 30 * time.Second
	}
	d, err := time.ParseDuration(c.Global.Timeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// GetEnabledProviders 获取已启用的提供商配置
func (c *Config) GetEnabledProviders() map[string]ProviderConfig {
	result := make(map[string]ProviderConfig)
	for name, cfg := range c.Providers {
		if cfg.Enabled {
			result[strings.ToLower(name)] = cfg
		}
	}
	return result
}

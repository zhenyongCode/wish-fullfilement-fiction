package config

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// Config 应用配置结构体
type Config struct {
	Workspace string        `json:"workspace"`
	Output    string        `json:"output"`
	Timeout   time.Duration `json:"timeout"`
	Logging   LoggingConfig `json:"logging"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	File   string `json:"file"`
}

// Load 从 gcfg 加载配置
func Load(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	// 使用 gcfg 自动从 manifest/config/config.yaml 加载
	if err := g.Cfg().MustGet(ctx, ".").Scan(cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	applyDefaults(cfg)

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Workspace == "" {
		cfg.Workspace = "."
	}
	if cfg.Output == "" {
		cfg.Output = "text"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Logging.File == "" {
		cfg.Logging.File = "logs/agent.log"
	}
}

// Get 获取配置值（便捷方法）
func Get(ctx context.Context, key string, def ...interface{}) interface{} {
	return g.Cfg().MustGet(ctx, key, def...).Val()
}

// GetString 获取字符串配置
func GetString(ctx context.Context, key string, def ...string) string {
	if len(def) > 0 {
		return g.Cfg().MustGet(ctx, key, def[0]).String()
	}
	return g.Cfg().MustGet(ctx, key).String()
}

// GetInt 获取整数配置
func GetInt(ctx context.Context, key string, def ...int) int {
	if len(def) > 0 {
		return g.Cfg().MustGet(ctx, key, def[0]).Int()
	}
	return g.Cfg().MustGet(ctx, key).Int()
}

// GetDuration 获取时间间隔配置
func GetDuration(ctx context.Context, key string, def ...time.Duration) time.Duration {
	if len(def) > 0 {
		return g.Cfg().MustGet(ctx, key, def[0]).Duration()
	}
	return g.Cfg().MustGet(ctx, key).Duration()
}
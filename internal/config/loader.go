package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 表示 MCP 服务器的完整配置
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Permissions PermissionsConfig `yaml:"permissions"`
}

// ServerConfig 表示服务器设置
type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// PermissionsConfig 表示工具权限配置
type PermissionsConfig struct {
	Default DefaultPermission          `yaml:"default"` // 默认权限
	Agents  map[string]AgentPermission `yaml:"agents"`  // 特定 agent 的权限
}

// DefaultPermission 表示默认权限设置
type DefaultPermission struct {
	Tools []string `yaml:"tools"` // 允许调用的工具列表
}

// AgentPermission 表示特定 agent 的权限设置
type AgentPermission struct {
	Tools []string `yaml:"tools"` // 允许调用的工具列表
}

// Load 从 YAML 文件加载配置
func Load() (*Config, error) {
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 应用默认值
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = "44444"
	}

	// 验证配置
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	return &cfg, nil
}

// validateConfig 验证必需的配置字段
func validateConfig(cfg *Config) error {
	// 验证权限配置
	// if len(cfg.Permissions.Default.Tools) == 0 {
	// 	return fmt.Errorf("permissions.default.tools 至少需要一个工具")
	// }

	return nil
}

// GetAllowedTools 获取指定 agent_id 允许调用的工具列表
func (c *Config) GetAllowedTools(agentID string) []string {
	if agentID == "" {
		// 如果没有提供 agent_id，使用默认权限
		return c.Permissions.Default.Tools
	}

	// 查找特定 agent 的权限
	if agentPerm, exists := c.Permissions.Agents[agentID]; exists {
		return agentPerm.Tools
	}

	// 如果 agent_id 未配置，使用默认权限
	return c.Permissions.Default.Tools
}

// IsToolAllowed 检查指定 agent_id 是否允许调用指定工具
func (c *Config) IsToolAllowed(agentID, toolName string) bool {
	allowedTools := c.GetAllowedTools(agentID)
	for _, allowedTool := range allowedTools {
		if allowedTool == toolName {
			return true
		}
	}
	return false
}

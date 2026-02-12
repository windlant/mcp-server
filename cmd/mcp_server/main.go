package main

import (
	"log"
	"os"

	"github.com/windlant/mcp-server/internal/config"
	"github.com/windlant/mcp-server/internal/tools/builtin"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建并初始化 server
	server := NewMCPServer(cfg)

	// 注册内置工具
	server.RegisterTool(builtin.GetTimeToolDef)

	// 启动服务器
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting MCP server on http://%s", addr)
	if err := server.Start(addr); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}

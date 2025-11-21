package server

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	mcpServer "github.com/tx7do/kratos-transport/transport/mcp"

	"go-kratos-mcp-demo/internal/service"
)

func NewMcpServer(_ log.Logger, recommendService *service.RecommendService) *mcpServer.Server {
	srv := mcpServer.NewServer(
		mcpServer.WithServerName("Recommend MCP Server"),
		mcpServer.WithServerVersion("1.0.0"),
		mcpServer.WithMCPServeType(mcpServer.ServerTypeHTTP),
		mcpServer.WithMCPServeAddress(":8080"),
		mcpServer.WithMCPServerOptions(
			server.WithToolCapabilities(false),
			server.WithRecovery(),
		),
	)

	var err error

	// 注册推荐工具
	if err = srv.RegisterHandler(
		mcp.Tool{
			Name:        "recommend",
			Description: "获取个性化推荐结果",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"userFeature": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"userId": map[string]interface{}{
								"type":        "string",
								"description": "用户ID",
							},
						},
					},
					"triggerItemId": map[string]interface{}{
						"type":        "integer",
						"description": "触发推荐的商品ID",
					},
					"scene": map[string]interface{}{
						"type":        "string",
						"description": "推荐场景",
					},
				},
				Required: []string{"userFeature"},
			},
		},
		recommendService.HandleRecommend,
	); err != nil {
		log.Errorf("failed to register recommend tool: %v", err)
	}

	return srv
}

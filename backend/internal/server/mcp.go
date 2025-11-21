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
		mcpServer.WithMCPServerOptions(
			server.WithToolCapabilities(false),
			server.WithRecovery(),
		),
	)

	recommendTool := mcp.NewTool("recommend",
		mcp.WithDescription("Provide product recommendations based on user behavior"),
		mcp.WithString("user_id",
			mcp.Required(),
			mcp.Description("The ID of the user"),
		),
		mcp.WithNumber("item_id",
			mcp.Required(),
			mcp.Description("The ID of the item the user interacted with"),
		),
		mcp.WithString("scene",
			mcp.Description("The recommendation scene or context"),
		),
	)

	_ = srv.RegisterHandler(recommendTool, recommendService.HandleRecommend)

	return srv
}

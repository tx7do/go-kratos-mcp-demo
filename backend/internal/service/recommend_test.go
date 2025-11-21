package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
)

func TestRecommendService_HandleRecommend(t *testing.T) {
	// 准备依赖服务
	logger := log.DefaultLogger
	recallService := NewRecallService(logger, nil)
	rankService := NewRankService(logger, nil)
	filterService := NewFilterService(logger, nil)

	// 创建推荐服务
	service := NewRecommendService(logger, recallService, rankService, filterService)

	// 构造 MCP 请求
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "recommend",
			Arguments: map[string]interface{}{
				"userFeature": map[string]interface{}{
					"userId": "user123",
				},
				"triggerItemId": int64(1001),
				"scene":         "homepage",
				"actionType":    "click",
			},
		},
	}

	// 执行测试
	ctx := context.Background()
	result, err := service.HandleRecommend(ctx, request)

	// 断言结果
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)
}

func TestMcpClient(t *testing.T) {
	ctx := context.Background()

	httpTransport, err := transport.NewStreamableHTTP("http://localhost:8080/mcp")
	assert.NoError(t, err)
	assert.NotNil(t, httpTransport)

	mcpClient := client.NewClient(
		httpTransport,
	)
	assert.NotNil(t, mcpClient)
	defer mcpClient.Close()

	err = mcpClient.Start(ctx)
	assert.NoError(t, err)

	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "recommend-http-client",
				Version: "1.0.0",
			},
		},
	}

	_, err = mcpClient.Initialize(ctx, initRequest)
	assert.NoError(t, err)

	// 调用推荐工具
	result, err := mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "recommend",
			Arguments: map[string]interface{}{
				"userFeature": map[string]interface{}{
					"userId": "user123",
				},
				"triggerItemId": int64(1001),
				"scene":         "homepage",
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Content)

	if len(result.Content) == 0 {
		t.Fatalf("empty result.Content")
	}
	for i, c := range result.Content {
		switch ct := c.(type) {
		case mcp.TextContent:
			var data map[string]interface{}
			if err = json.Unmarshal([]byte(ct.Text), &data); err != nil {
				t.Fatalf("unmarshal text content[%d] failed: %v", i, err)
			}
			assert.NotEmpty(t, data)
			t.Logf("text content[%d]=%#v", i, data)

			root := data
			assert.Equal(t, "CONTEXT_STAGE_OUTPUT", fmt.Sprintf("%v", root["stage"]))
			assert.NotEmpty(t, root["requestId"])

			inputCtx, ok := root["inputCtx"].(map[string]interface{})
			if !ok {
				t.Fatalf("missing inputCtx")
			}

			if bl, ok := inputCtx["blacklistItemIds"].([]interface{}); ok {
				assert.True(t, len(bl) >= 0)
				assert.Equal(t, bl[0], "999")
				assert.Equal(t, bl[1], "888")
			} else {
				t.Fatalf("blacklistItemIds missing or invalid")
			}

		default:
			t.Logf("content[%d] unexpected type=%T value=%#v", i, c, c)
		}
	}

	t.Logf("完整结果: %#v", result)
}

package service

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
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

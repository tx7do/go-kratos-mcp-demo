package service

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go-kratos-mcp-demo/api/gen/go/conf"
	recommendV1 "go-kratos-mcp-demo/api/gen/go/recommend/service/v1"
)

// FilterService 过滤服务（过滤黑名单、已购买商品等）
type FilterService struct {
	logger log.Logger
	rc     *conf.RecommendConfig
}

func NewFilterService(logger log.Logger, rc *conf.RecommendConfig) *FilterService {
	return &FilterService{
		logger: logger,
		rc:     rc,
	}
}

// Filter 执行过滤逻辑
func (s *FilterService) Filter(ctx context.Context, input *recommendV1.FilterInputContext) (*recommendV1.RecommendOutput, error) {
	requestId := input.GetRequestId()
	rankedItems := input.GetRankCtx().GetRankedItems()

	log.Infow(ctx, "start filter", "request_id", requestId, "ranked_count", len(rankedItems))

	// 过滤黑名单商品
	var finalItems []*recommendV1.RankedItem
	blacklistMap := make(map[int64]struct{})

	// 从配置获取黑名单
	for _, itemID := range s.rc.GetService().GetFilter().GetBlacklist() {
		blacklistMap[itemID] = struct{}{}
	}

	// 从输入上下文获取黑名单
	for _, itemID := range input.GetBlacklistItemIds() {
		blacklistMap[itemID] = struct{}{}
	}

	// 构建已购商品映射
	purchasedMap := make(map[int64]struct{})
	for _, itemID := range input.GetPurchasedItemIds() {
		purchasedMap[itemID] = struct{}{}
	}

	// 过滤黑名单和已购商品
	for _, item := range rankedItems {
		itemID := item.GetItem().GetItemId()
		if _, inBlacklist := blacklistMap[itemID]; inBlacklist {
			continue
		}
		if _, purchased := purchasedMap[itemID]; purchased {
			continue
		}
		finalItems = append(finalItems, item)
	}

	log.Infow(ctx, "filter finished", "request_id", requestId, "final_count", len(finalItems))

	return &recommendV1.RecommendOutput{
		Stage:          recommendV1.ContextStage_CONTEXT_STAGE_OUTPUT,
		RequestId:      requestId,
		InputCtx:       input,
		RecommendItems: finalItems,
		TotalCount:     int32(len(finalItems)),
		ResponseTime:   timestamppb.New(time.Now()),
	}, nil
}

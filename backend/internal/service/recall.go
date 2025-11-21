package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"go-kratos-mcp-demo/api/gen/go/conf"
	recommendV1 "go-kratos-mcp-demo/api/gen/go/recommend/service/v1"
)

// RecallService 召回服务(模拟基于历史行为的协同过滤召回)
type RecallService struct {
	logger log.Logger

	rc *conf.RecommendConfig

	// 模拟商品库:key=商品ID,value=关联商品ID列表
	itemCorrelation map[int64][]int64
}

func NewRecallService(logger log.Logger, rc *conf.RecommendConfig) *RecallService {
	// 初始化模拟商品关联数据
	itemCorrelation := map[int64][]int64{
		101: {102, 103, 104, 105, 106}, // 商品101关联商品
		102: {101, 103, 107, 108, 109},
		103: {101, 102, 110, 111, 112},
		201: {202, 203, 204, 205}, // 通勤场景商品
		202: {201, 203, 206, 207},
	}

	return &RecallService{
		logger:          logger,
		rc:              rc,
		itemCorrelation: itemCorrelation,
	}
}

// Recall 执行召回逻辑
func (s *RecallService) Recall(ctx context.Context, input *recommendV1.RecallInputContext) (*recommendV1.RecallOutputContext, error) {
	requestId := input.GetRequestId()
	scene := input.GetScene()
	historyItemIds := input.GetHistoryItemIds()

	log.Infow(ctx, "start recall", "request_id", requestId, "scene", scene, "history_items", historyItemIds)

	// 基于场景和历史行为召回候选商品
	var candidateIds []int64

	// 场景适配:通勤场景优先推荐通勤类商品(200+ID)
	if scene == recommendV1.SceneType_SCENE_TYPE_COMMUTE {
		commuteItems := []int64{201, 202, 203, 204, 205, 206, 207, 208}
		candidateIds = append(candidateIds, commuteItems...)
	}

	// 基于历史行为召回关联商品
	for _, itemID := range historyItemIds {
		if relatedItems, ok := s.itemCorrelation[itemID]; ok {
			candidateIds = append(candidateIds, relatedItems...)
		}
	}

	// 去重+随机筛选topK
	uniqueCandidates := make(map[int64]struct{})
	for _, item := range candidateIds {
		uniqueCandidates[item] = struct{}{}
		if int32(len(uniqueCandidates)) >= s.rc.GetService().GetRecall().GetTopK() {
			break
		}
	}

	// 构建候选商品列表
	var candidateItems []*recommendV1.Item
	for itemID := range uniqueCandidates {
		candidateItems = append(candidateItems, &recommendV1.Item{
			ItemId:   itemID,
			Title:    "商品标题", // 实际应从数据库查询
			Price:    99.99,
			ImageUrl: "",
			Sales:    1000,
		})
	}

	log.Infow(ctx, "recall finished", "request_id", requestId, "candidate_count", len(candidateItems))

	return &recommendV1.RecallOutputContext{
		Stage:          recommendV1.ContextStage_CONTEXT_STAGE_RECALL,
		RequestId:      requestId,
		InputCtx:       input,
		CandidateItems: candidateItems,
		RecallStrategy: "collaborative_filtering_v1",
	}, nil
}

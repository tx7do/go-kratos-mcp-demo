package service

import (
	"context"
	"math/rand"

	"github.com/go-kratos/kratos/v2/log"

	"go-kratos-mcp-demo/api/gen/go/conf"
	recommendV1 "go-kratos-mcp-demo/api/gen/go/recommend/service/v1"
)

// RankService 排序服务(模拟基于用户特征和商品相关性的打分排序)
type RankService struct {
	logger log.Logger
	rc     *conf.RecommendConfig
}

func NewRankService(logger log.Logger, rc *conf.RecommendConfig) *RankService {
	return &RankService{
		logger: logger,
		rc:     rc,
	}
}

// Rank 执行排序逻辑
func (s *RankService) Rank(ctx context.Context, input *recommendV1.RankInputContext) (*recommendV1.RankOutputContext, error) {
	recallCtx := input.GetRecallCtx()
	candidateItems := recallCtx.GetCandidateItems()
	requestId := input.GetRequestId()

	log.Infow(ctx, "start rank", "request_id", requestId, "candidate_count", len(candidateItems))

	// 从 RecallInputContext 获取场景信息
	scene := recallCtx.GetInputCtx().GetScene()

	// 模拟打分:结合用户特征(简化为随机分数+场景权重)
	var rankedItems []*recommendV1.RankedItem

	for _, item := range candidateItems {
		// 基础分数(50-100)
		baseScore := 50.0 + rand.Float64()*50.0

		// 场景权重:通勤场景短内容商品加分
		if scene == recommendV1.SceneType_SCENE_TYPE_COMMUTE && item.GetItemId() >= 200 {
			baseScore *= 1.2
		}

		// 热门度加权
		if item.GetSales() > 1000 {
			baseScore *= 1.1
		}

		rankedItems = append(rankedItems, &recommendV1.RankedItem{
			Item:   item,
			Score:  baseScore,
			Reason: "基于个性化推荐算法",
		})
	}

	// 按分数降序排序
	for i := 0; i < len(rankedItems); i++ {
		for j := i + 1; j < len(rankedItems); j++ {
			if rankedItems[j].GetScore() > rankedItems[i].GetScore() {
				rankedItems[i], rankedItems[j] = rankedItems[j], rankedItems[i]
			}
		}
	}

	// 取topK
	topK := input.GetRankTopK()
	if topK <= 0 {
		topK = s.rc.GetService().GetRank().GetTopK()
	}
	if int32(len(rankedItems)) > topK {
		rankedItems = rankedItems[:topK]
	}

	log.Infow(ctx, "rank finished", "request_id", requestId, "ranked_count", len(rankedItems))

	return &recommendV1.RankOutputContext{
		Stage:       recommendV1.ContextStage_CONTEXT_STAGE_RANK,
		RequestId:   requestId,
		InputCtx:    input,
		RankedItems: rankedItems,
		RankModel:   "simple_score_v1",
	}, nil
}

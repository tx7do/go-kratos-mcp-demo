package service

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/protobuf/types/known/timestamppb"

	recommendV1 "go-kratos-mcp-demo/api/gen/go/recommend/service/v1"
)

type RecommendService struct {
	recommendV1.RecommendServiceHTTPServer

	recallService *RecallService
	rankService   *RankService
	filterService *FilterService

	userHistoryCache map[string][]int64
	maxHistoryLen    int

	log *log.Helper

	sync.RWMutex
}

func NewRecommendService(
	logger log.Logger,
	recallService *RecallService,
	rankService *RankService,
	filterService *FilterService,
) *RecommendService {
	l := log.NewHelper(log.With(logger, "module", "file/service/mcp-service"))

	return &RecommendService{
		recallService: recallService,
		rankService:   rankService,
		filterService: filterService,

		userHistoryCache: make(map[string][]int64),
		maxHistoryLen:    10,

		log: l,
	}
}

// HandleRecommend 处理推荐请求(MCP核心调度逻辑)
func (s *RecommendService) HandleRecommend(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.log.Info("received recommend request", request)

	var err error

	codec := encoding.GetCodec("json")

	// 解析用户行为上下文(MCP协议标准化输入)
	var actionCtx recommendV1.UserActionContext
	var argsData []byte
	if argsData, err = codec.Marshal(request.Params.Arguments); err != nil {
		log.Errorw(ctx, "failed to marshal arguments", "err", err)
		return nil, err
	}
	if err = codec.Unmarshal(argsData, &actionCtx); err != nil {
		log.Errorw(ctx, "failed to unmarshal user action context", "err", err)
		return nil, err
	}

	// 生成请求ID(用于链路追踪)
	requestId := uuid.New().String()
	actionCtx.RequestId = requestId
	actionCtx.Stage = recommendV1.ContextStage_CONTEXT_STAGE_INPUT

	// 更新用户历史行为缓存
	userId := actionCtx.GetUserFeature().GetUserId()
	triggerItemId := actionCtx.GetTriggerItemId()

	s.Lock()
	history := s.userHistoryCache[userId]
	history = append([]int64{triggerItemId}, history...) // 最新行为放前面
	if len(history) > s.maxHistoryLen {
		history = history[:s.maxHistoryLen]
	}
	s.userHistoryCache[userId] = history
	s.Unlock()

	// 多模块协同: MCP上下文传递

	// 召回模块
	recallInput := &recommendV1.RecallInputContext{
		Stage:          recommendV1.ContextStage_CONTEXT_STAGE_RECALL,
		RequestId:      requestId,
		UserFeature:    actionCtx.GetUserFeature(),
		Scene:          actionCtx.GetScene(),
		HistoryItemIds: history,
		RecallTopK:     s.recallService.rc.GetService().GetRecall().GetTopK(),
	}
	var recallOutput *recommendV1.RecallOutputContext
	if recallOutput, err = s.recallService.Recall(ctx, recallInput); err != nil {
		s.log.Errorw(ctx, "recall failed", "err", err)
		return nil, err
	}

	// 排序模块
	rankInput := &recommendV1.RankInputContext{
		Stage:     recommendV1.ContextStage_CONTEXT_STAGE_RANK,
		RequestId: requestId,
		RecallCtx: recallOutput,
		RankTopK:  s.rankService.rc.GetService().GetRank().GetTopK(),
	}
	var rankOutput *recommendV1.RankOutputContext
	if rankOutput, err = s.rankService.Rank(ctx, rankInput); err != nil {
		s.log.Errorw(ctx, "rank failed", "err", err)
		return nil, err
	}

	// 过滤模块
	filterInput := &recommendV1.FilterInputContext{
		Stage:            recommendV1.ContextStage_CONTEXT_STAGE_FILTER,
		RequestId:        requestId,
		RankCtx:          rankOutput,
		BlacklistItemIds: s.filterService.rc.GetService().GetFilter().GetBlacklist(),
		PurchasedItemIds: []int64{}, // 实际应从用户购买记录查询
	}
	var finalOutput *recommendV1.RecommendOutput
	if finalOutput, err = s.filterService.Filter(ctx, filterInput); err != nil {
		s.log.Errorw(ctx, "filter failed", "err", err)
		return nil, err
	}

	// 标准化输出(MCP协议格式)
	var finalOutputJson []byte
	if finalOutputJson, err = codec.Marshal(finalOutput); err != nil {
		s.log.Errorw(ctx, "failed to marshal final output", "err", err)
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(finalOutputJson),
			},
		},
	}, nil
}

// TriggerRecommend 推荐接口
func (s *RecommendService) TriggerRecommend(ctx context.Context, req *recommendV1.RecommendRequest) (*recommendV1.RecommendResponse, error) {
	actionCtx := req.GetActionCtx()

	// 生成请求ID
	requestId := uuid.New().String()
	actionCtx.RequestId = requestId
	actionCtx.Stage = recommendV1.ContextStage_CONTEXT_STAGE_INPUT
	actionCtx.ActionTime = timestamppb.Now()

	// 更新用户历史
	userId := actionCtx.GetUserFeature().GetUserId()
	triggerItemId := actionCtx.GetTriggerItemId()

	s.Lock()
	history := s.userHistoryCache[userId]
	history = append([]int64{triggerItemId}, history...)
	if len(history) > s.maxHistoryLen {
		history = history[:s.maxHistoryLen]
	}
	s.userHistoryCache[userId] = history
	s.Unlock()

	// 召回
	recallInput := &recommendV1.RecallInputContext{
		Stage:          recommendV1.ContextStage_CONTEXT_STAGE_RECALL,
		RequestId:      requestId,
		UserFeature:    actionCtx.GetUserFeature(),
		Scene:          actionCtx.GetScene(),
		HistoryItemIds: history,
		RecallTopK:     s.recallService.rc.GetService().GetRecall().GetTopK(),
	}
	recallOutput, err := s.recallService.Recall(ctx, recallInput)
	if err != nil {
		return nil, recommendV1.ErrorInternalServerError("召回失败")
	}

	// 排序
	rankInput := &recommendV1.RankInputContext{
		Stage:     recommendV1.ContextStage_CONTEXT_STAGE_RANK,
		RequestId: requestId,
		RecallCtx: recallOutput,
		RankTopK:  s.rankService.rc.GetService().GetRank().GetTopK(),
	}
	rankOutput, err := s.rankService.Rank(ctx, rankInput)
	if err != nil {
		return nil, recommendV1.ErrorInternalServerError("排序失败")
	}

	// 过滤
	filterInput := &recommendV1.FilterInputContext{
		Stage:            recommendV1.ContextStage_CONTEXT_STAGE_FILTER,
		RequestId:        requestId,
		RankCtx:          rankOutput,
		BlacklistItemIds: s.filterService.rc.GetService().GetFilter().GetBlacklist(),
		PurchasedItemIds: []int64{},
	}
	finalOutput, err := s.filterService.Filter(ctx, filterInput)
	if err != nil {
		return nil, recommendV1.ErrorInternalServerError("过滤失败")
	}

	return &recommendV1.RecommendResponse{
		Output:    finalOutput,
		RequestId: requestId,
	}, nil
}

// GetRecommendHistory 获取推荐历史(待实现)
func (s *RecommendService) GetRecommendHistory(ctx context.Context, req *recommendV1.RecommendHistoryRequest) (*recommendV1.RecommendHistoryResponse, error) {
	return &recommendV1.RecommendHistoryResponse{
		Total:    0,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

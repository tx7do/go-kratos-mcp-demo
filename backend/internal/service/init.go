//go:build wireinject
// +build wireinject

package service

import (
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(
	NewFilterService,
	NewRankService,
	NewRecallService,
	NewRecommendService,
)

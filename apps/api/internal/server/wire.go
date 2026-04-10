//go:build wireinject
// +build wireinject

package server

import (
	"context"

	"github.com/google/wire"

	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/config"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/database"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/handler"
	"github.com/chaosplane-hq/chaosplane-platform/apps/api/internal/service"
)

var ProviderSet = wire.NewSet(
	config.New,
	database.NewPool,
	handler.NewHealthHandler,
	handler.NewAuthHandler,
	handler.NewHierarchyHandler,
	handler.NewOnboardingHandler,
	handler.NewInvitationHandler,
	handler.NewAPIKeyHandler,
	handler.NewOAuthHandler,
	handler.NewAccountHandler,
	handler.NewAgentHandler,
	handler.NewExperimentHandler,
	handler.NewPolicyHandler,
	handler.NewWebSocketHandler,
	service.NewAuthService,
	service.NewHierarchyService,
	service.NewOnboardingService,
	service.NewInvitationService,
	service.NewAPIKeyService,
	service.NewOAuthService,
	service.NewAccountService,
	service.NewAgentService,
	service.NewK8sClient,
	service.NewExperimentService,
	service.NewPolicyService,
	NewRedisClient,
	New,
)

func InitializeServer(ctx context.Context) (*Server, error) {
	wire.Build(ProviderSet)
	return nil, nil
}

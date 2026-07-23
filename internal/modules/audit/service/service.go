package service

import (
	"context"
	"encoding/json"

	"github.com/fredele20/nimbus/internal/modules/audit/model"
	"github.com/fredele20/nimbus/internal/modules/audit/repository"
)

type service struct {
	repo repository.Repository
}

type Service interface {
	Log(ctx context.Context, actorID string, userID string, action string, entity string, entityID string, metadata any) error
}

func New(repo repository.Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) Log(ctx context.Context, actorID string, userID string, action string, entity string, entityID string, metadata any) error {
	metadataBytes, _ := json.Marshal(metadata)

	audit := model.AuditLog{
		UserID:   userID,
		ActorID:  actorID,
		Action:   action,
		Entity:   entity,
		EntityID: entityID,
		Metadata: metadataBytes,
	}

	return s.repo.Create(ctx, &audit)
}

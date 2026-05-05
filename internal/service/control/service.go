// Package control owns the read+attach/detach API for the canonical
// catalogue of controls (А, FW, HP, …) and the asset_controls bridge.
package control

import (
	"context"
	"fmt"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

type Service interface {
	List(ctx context.Context) ([]domain.Control, error)
	ListForAsset(ctx context.Context, assetID int64) ([]domain.Control, error)
	Attach(ctx context.Context, assetID, controlID int64) error
	Detach(ctx context.Context, assetID, controlID int64) error
}

type service struct {
	repo repository.ControlRepository
}

func NewService(repo repository.ControlRepository) Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) ([]domain.Control, error) {
	return s.repo.List(ctx)
}

func (s *service) ListForAsset(ctx context.Context, assetID int64) ([]domain.Control, error) {
	if assetID <= 0 {
		return nil, fmt.Errorf("invalid assetID")
	}
	return s.repo.ListAttached(ctx, assetID)
}

func (s *service) Attach(ctx context.Context, assetID, controlID int64) error {
	if assetID <= 0 || controlID <= 0 {
		return fmt.Errorf("invalid assetID/controlID")
	}
	return s.repo.Attach(ctx, assetID, controlID)
}

func (s *service) Detach(ctx context.Context, assetID, controlID int64) error {
	if assetID <= 0 || controlID <= 0 {
		return fmt.Errorf("invalid assetID/controlID")
	}
	return s.repo.Detach(ctx, assetID, controlID)
}

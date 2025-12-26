package service

import (
	"context"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/Cxiyuan/NTA/services/asset-service/internal/repository"
	"github.com/sirupsen/logrus"
)

type AssetService struct {
	repo   *repository.AssetRepository
	logger *logrus.Logger
}

func NewAssetService(repo *repository.AssetRepository, logger *logrus.Logger) *AssetService {
	return &AssetService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AssetService) GetAssetByIP(ctx context.Context, ip string) (*models.Asset, error) {
	return s.repo.GetAssetByIP(ctx, ip)
}

func (s *AssetService) ListAssets(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]models.Asset, int64, error) {
	return s.repo.ListAssets(ctx, limit, offset, filters)
}

func (s *AssetService) CreateAsset(ctx context.Context, asset *models.Asset) error {
	return s.repo.CreateAsset(ctx, asset)
}

func (s *AssetService) UpdateAsset(ctx context.Context, ip string, updates map[string]interface{}) error {
	return s.repo.UpdateAsset(ctx, ip, updates)
}

func (s *AssetService) DeleteAsset(ctx context.Context, ip string) error {
	return s.repo.DeleteAsset(ctx, ip)
}

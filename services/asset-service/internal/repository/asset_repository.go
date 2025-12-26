package repository

import (
	"context"

	"github.com/Cxiyuan/NTA/pkg/models"
	"gorm.io/gorm"
)

type AssetRepository struct {
	db *gorm.DB
}

func NewAssetRepository(db *gorm.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

func (r *AssetRepository) GetAssetByIP(ctx context.Context, ip string) (*models.Asset, error) {
	var asset models.Asset
	err := r.db.WithContext(ctx).Where("ip = ?", ip).First(&asset).Error
	return &asset, err
}

func (r *AssetRepository) ListAssets(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]models.Asset, int64, error) {
	var assets []models.Asset
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Asset{})

	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Limit(limit).Offset(offset).Find(&assets).Error
	return assets, total, err
}

func (r *AssetRepository) CreateAsset(ctx context.Context, asset *models.Asset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

func (r *AssetRepository) UpdateAsset(ctx context.Context, ip string, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Asset{}).Where("ip = ?", ip).Updates(updates).Error
}

func (r *AssetRepository) DeleteAsset(ctx context.Context, ip string) error {
	return r.db.WithContext(ctx).Where("ip = ?", ip).Delete(&models.Asset{}).Error
}

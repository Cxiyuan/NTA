package repository

import (
	"context"

	"github.com/Cxiyuan/NTA/pkg/models"
	"gorm.io/gorm"
)

type AlertRepository struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) GetAlertByID(ctx context.Context, id uint) (*models.Alert, error) {
	var alert models.Alert
	err := r.db.WithContext(ctx).First(&alert, id).Error
	return &alert, err
}

func (r *AlertRepository) ListAlerts(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]models.Alert, int64, error) {
	var alerts []models.Alert
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Alert{})
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Order("timestamp DESC").Limit(limit).Offset(offset).Find(&alerts).Error
	return alerts, total, err
}

func (r *AlertRepository) CreateAlert(ctx context.Context, alert *models.Alert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *AlertRepository) UpdateAlert(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Updates(updates).Error
}

func (r *AlertRepository) DeleteAlert(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.Alert{}, id).Error
}

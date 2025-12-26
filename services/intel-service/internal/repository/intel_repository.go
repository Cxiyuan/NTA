package repository

import (
	"gorm.io/gorm"
)

type IntelRepository struct {
	db *gorm.DB
}

func NewIntelRepository(db *gorm.DB) *IntelRepository {
	return &IntelRepository{db: db}
}

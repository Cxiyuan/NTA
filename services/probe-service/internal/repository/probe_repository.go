package repository

import (
	"gorm.io/gorm"
)

type ProbeRepository struct {
	db *gorm.DB
}

func NewProbeRepository(db *gorm.DB) *ProbeRepository {
	return &ProbeRepository{db: db}
}

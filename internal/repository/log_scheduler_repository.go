package repository

import (
	"ipl-be-svc/internal/models"

	"gorm.io/gorm"
)

// LogSchedulerRepository defines the interface for log scheduler data operations
type LogSchedulerRepository interface {
	CreateLogScheduler(log *models.LogSchedullers) error
}

// logSchedulerRepository implements LogSchedulerRepository
type logSchedulerRepository struct {
	db *gorm.DB
}

// NewLogSchedulerRepository creates a new instance of LogSchedulerRepository
func NewLogSchedulerRepository(db *gorm.DB) LogSchedulerRepository {
	return &logSchedulerRepository{
		db: db,
	}
}

// CreateLogScheduler creates a new log scheduler record
func (r *logSchedulerRepository) CreateLogScheduler(log *models.LogSchedullers) error {
	return r.db.Create(log).Error
}

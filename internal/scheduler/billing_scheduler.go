package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"ipl-be-svc/internal/models"
	"ipl-be-svc/internal/repository"
	"ipl-be-svc/internal/service"
	"ipl-be-svc/pkg/logger"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// BillingScheduler handles scheduled billing operations
type BillingScheduler struct {
	billingService   service.BillingService
	logSchedulerRepo repository.LogSchedulerRepository
	logger           *logger.Logger
	cron             *cron.Cron
	cronExpression   string
}

// NewBillingScheduler creates a new billing scheduler
func NewBillingScheduler(billingService service.BillingService, logSchedulerRepo repository.LogSchedulerRepository, logger *logger.Logger, cronExpression string) *BillingScheduler {
	// Create cron with seconds precision
	c := cron.New(cron.WithSeconds())

	return &BillingScheduler{
		billingService:   billingService,
		logSchedulerRepo: logSchedulerRepo,
		logger:           logger,
		cron:             c,
		cronExpression:   cronExpression,
	}
}

// Start initializes and starts all scheduled jobs
func (s *BillingScheduler) Start() error {
	s.logger.Info("Starting billing scheduler...")

	// Schedule job using cron expression from configuration
	// Cron format: "seconds minutes hours day-of-month month day-of-week"
	s.logger.WithField("cron_expression", s.cronExpression).Info("Scheduling billing job")
	_, err := s.cron.AddFunc(s.cronExpression, s.createMonthlyBillings)
	if err != nil {
		return fmt.Errorf("failed to schedule monthly billings job: %w", err)
	}

	s.logger.WithField("cron_expression", s.cronExpression).Info("Billing job scheduled successfully")

	// Start the cron scheduler
	s.cron.Start()
	s.logger.Info("Billing scheduler started successfully")

	return nil
}

// Stop gracefully stops the scheduler
func (s *BillingScheduler) Stop() {
	s.logger.Info("Stopping billing scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("Billing scheduler stopped successfully")
}

// createMonthlyBillings is the scheduled job that creates billings for all users
func (s *BillingScheduler) createMonthlyBillings() {
	schedullerCode := "MONTHLY_BILLING_CREATION"
	adminID := 1
	now := time.Now()
	docID := uuid.New().String()

	// Log START status
	startMessage := "Starting scheduled monthly billing creation"
	s.logScheduler(schedullerCode, docID, startMessage, "START", adminID, &now)

	s.logger.Info("Starting scheduled monthly billing creation...")

	month := int(now.Month())
	year := now.Year()

	s.logger.WithField("month", month).WithField("year", year).Info("Creating monthly billings for all users")

	// Log RUNNING status
	runningMessage := fmt.Sprintf("Creating monthly billings for month %d year %d", month, year)
	s.logScheduler(schedullerCode, docID, runningMessage, "RUNNING", adminID, &now)

	// Create monthly billings
	monthlyResponse, err := s.billingService.CreateBulkMonthlyBillingsForAllUsers(month, year)

	if err != nil {
		// Log FAILED status
		failedMessage := fmt.Sprintf("Failed to create monthly billings: %v", err)
		s.logScheduler(schedullerCode, docID, failedMessage, "FAILED", adminID, &now)
		s.logger.WithField("error", err).Error("Failed to create monthly billings")
		return
	}

	// Log SUCCESS status with response
	responseJSON, _ := json.Marshal(monthlyResponse)
	successMessage := fmt.Sprintf("Monthly billings created successfully: %s", string(responseJSON))
	s.logScheduler(schedullerCode, docID, successMessage, "SUCCESS", adminID, &now)

	s.logger.WithField("response", monthlyResponse).Info("Monthly billings created successfully")
	s.logger.Info("Scheduled monthly billing creation completed")
}

// logScheduler creates a new log entry in the database
func (s *BillingScheduler) logScheduler(schedullerCode, documentID, message, status string, createdByID int, createdAt *time.Time) {
	logEntry := &models.LogSchedullers{
		DocumentID:       &documentID,
		SchedullerCode:   &schedullerCode,
		Message:          &message,
		StatusScheduller: &status,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
		PublishedAt:      createdAt,
		CreatedByID:      &createdByID,
		UpdatedByID:      &createdByID,
		Locale:           stringPtr("en"),
	}

	if err := s.logSchedulerRepo.CreateLogScheduler(logEntry); err != nil {
		s.logger.WithField("error", err).WithField("status", status).Error("Failed to create scheduler log entry")
	} else {
		s.logger.WithField("status", status).WithField("document_id", documentID).Info("Scheduler log entry created")
	}
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

package service

import (
	"fmt"
	"ipl-be-svc/internal/models"
	"ipl-be-svc/internal/models/response"
	"ipl-be-svc/internal/repository"
	"ipl-be-svc/pkg/logger"
)

// UserService interface defines user service methods
type UserService interface {
	GetUserDetailByProfileID(profileID uint) (*models.UserDetail, error)
	GetPenghuniUsers() ([]*response.PenghuniUserResponse, error)
}

// userService implements UserService interface
type userService struct {
	userRepo repository.UserRepository
	logger   *logger.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, logger *logger.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetUserDetailByProfileID gets user detail by profile ID
func (s *userService) GetUserDetailByProfileID(profileID uint) (*models.UserDetail, error) {
	if profileID == 0 {
		s.logger.WithField("profile_id", profileID).Error("Invalid profile ID")
		return nil, fmt.Errorf("invalid profile ID")
	}

	userDetail, err := s.userRepo.GetUserDetailByProfileID(profileID)
	if err != nil {
		s.logger.WithError(err).WithField("profile_id", profileID).Error("Failed to get user detail")
		return nil, err
	}

	s.logger.WithFields(map[string]interface{}{
		"profile_id": profileID,
		"user_id":    userDetail.UserID,
		"email":      userDetail.Email,
	}).Info("User detail retrieved successfully")

	return userDetail, nil
}

// GetPenghuniUsers gets all users with role type "penghuni"
func (s *userService) GetPenghuniUsers() ([]*response.PenghuniUserResponse, error) {
	// Get users with penghuni role from repository
	users, err := s.userRepo.GetUsersWithPenghuniRole()
	if err != nil {
		s.logger.WithError(err).Error("Failed to get penghuni users from repository")
		return nil, err
	}

	// Convert to service response format
	var penghuniUsers []*response.PenghuniUserResponse
	for _, user := range users {
		penghuniUser := &response.PenghuniUserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			RoleName: user.RoleName,
			RoleID:   user.RoleID,
			RoleType: user.RoleType,
		}

		// Get profile information if available
		if user.UserID > 0 {
			penghuniUser.NamaPenghuni = user.NamaPenghuni
			penghuniUser.NamaPemilik = user.NamaPemilik
			penghuniUser.Blok = user.Blok
			penghuniUser.Rt = user.Rt
			penghuniUser.NoHP = user.NoHP
			penghuniUser.NoTelp = user.NoTelp
			penghuniUser.DocumentID = user.DocumentID
		}

		penghuniUsers = append(penghuniUsers, penghuniUser)
	}

	s.logger.WithField("count", len(penghuniUsers)).Info("Penghuni users retrieved successfully")

	return penghuniUsers, nil
}

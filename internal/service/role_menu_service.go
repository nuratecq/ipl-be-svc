package service

import (
	"fmt"
	"ipl-be-svc/internal/models"
	"ipl-be-svc/internal/repository"
	"ipl-be-svc/pkg/logger"
)

// RoleMenuService interface defines role menu service methods
type RoleMenuService interface {
	CreateRoleMenu(req *CreateRoleMenuRequest) (*models.RoleMenu, error)
	GetRoleMenuByID(id uint) (*models.RoleMenu, error)
	GetAllRoleMenus(limit, offset int) ([]models.RoleMenu, int64, error)
	UpdateRoleMenu(id uint, req *UpdateRoleMenuRequest) (*models.RoleMenu, error)
	DeleteRoleMenu(id uint) error
	GetRoleMenusByRoleID(roleID uint) ([]models.RoleMenu, error)
	AttachMasterMenuToRoleMenu(roleMenuID, masterMenuID uint, order *float64) error
	DetachMasterMenuFromRoleMenu(roleMenuID, masterMenuID uint) error
	AttachRoleToRoleMenu(roleMenuID, roleID uint, order *float64) error
	DetachRoleFromRoleMenu(roleMenuID, roleID uint) error
}

// CreateRoleMenuRequest represents the request to create a role menu
type CreateRoleMenuRequest struct {
	DocumentID  *string  `json:"document_id" example:"role_menu001"`
	RoleMenuOrd *float64 `json:"role_menu_ord" example:"1.0"`
	IsActive    *bool    `json:"is_active" example:"true"`
	MasterMenus []uint   `json:"master_menu_ids,omitempty" example:"1,2,3"`
	Roles       []uint   `json:"role_ids,omitempty" example:"1,2"`
}

// UpdateRoleMenuRequest represents the request to update a role menu
type UpdateRoleMenuRequest struct {
	DocumentID  *string  `json:"document_id" example:"role_menu001"`
	RoleMenuOrd *float64 `json:"role_menu_ord" example:"1.0"`
	IsActive    *bool    `json:"is_active" example:"true"`
}

// AttachMasterMenuRequest represents the request to attach a master menu to role menu
type AttachMasterMenuRequest struct {
	MasterMenuID uint     `json:"master_menu_id" binding:"required" example:"1"`
	Order        *float64 `json:"order" example:"1.0"`
}

// AttachRoleRequest represents the request to attach a role to role menu
type AttachRoleRequest struct {
	RoleID uint     `json:"role_id" binding:"required" example:"1"`
	Order  *float64 `json:"order" example:"1.0"`
}

// roleMenuService implements RoleMenuService interface
type roleMenuService struct {
	roleMenuRepo   repository.RoleMenuRepository
	masterMenuRepo repository.MasterMenuRepository
	logger         *logger.Logger
}

// NewRoleMenuService creates a new role menu service
func NewRoleMenuService(
	roleMenuRepo repository.RoleMenuRepository,
	masterMenuRepo repository.MasterMenuRepository,
	logger *logger.Logger,
) RoleMenuService {
	return &roleMenuService{
		roleMenuRepo:   roleMenuRepo,
		masterMenuRepo: masterMenuRepo,
		logger:         logger,
	}
}

// CreateRoleMenu creates a new role menu
func (s *roleMenuService) CreateRoleMenu(req *CreateRoleMenuRequest) (*models.RoleMenu, error) {
	// Create role menu
	roleMenu := &models.RoleMenu{
		RoleMenuOrd: req.RoleMenuOrd,
		IsActive:    req.IsActive,
	}

	if req.DocumentID != nil {
		roleMenu.DocumentID = req.DocumentID
	}

	err := s.roleMenuRepo.Create(roleMenu)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create role menu")
		return nil, err
	}

	// Attach master menus if provided
	if len(req.MasterMenus) > 0 {
		for i, masterMenuID := range req.MasterMenus {
			order := float64(i + 1)
			err := s.roleMenuRepo.AttachMasterMenu(roleMenu.ID, masterMenuID, &order)
			if err != nil {
				s.logger.WithError(err).WithFields(map[string]interface{}{
					"role_menu_id":   roleMenu.ID,
					"master_menu_id": masterMenuID,
				}).Error("Failed to attach master menu to role menu")
			}
		}
	}

	// Attach roles if provided
	if len(req.Roles) > 0 {
		for i, roleID := range req.Roles {
			order := float64(i + 1)
			err := s.roleMenuRepo.AttachRole(roleMenu.ID, roleID, &order)
			if err != nil {
				s.logger.WithError(err).WithFields(map[string]interface{}{
					"role_menu_id": roleMenu.ID,
					"role_id":      roleID,
				}).Error("Failed to attach role to role menu")
			}
		}
	}

	s.logger.WithField("id", roleMenu.ID).Info("Role menu created successfully")

	// Return with relations
	return s.roleMenuRepo.GetWithRelations(roleMenu.ID)
}

// GetRoleMenuByID retrieves a role menu by ID
func (s *roleMenuService) GetRoleMenuByID(id uint) (*models.RoleMenu, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid role menu ID")
	}

	roleMenu, err := s.roleMenuRepo.GetWithRelations(id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get role menu")
		return nil, err
	}

	return roleMenu, nil
}

// GetAllRoleMenus retrieves all role menus with pagination
func (s *roleMenuService) GetAllRoleMenus(limit, offset int) ([]models.RoleMenu, int64, error) {
	roleMenus, total, err := s.roleMenuRepo.GetAll(limit, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get role menus")
		return nil, 0, err
	}

	return roleMenus, total, nil
}

// UpdateRoleMenu updates a role menu
func (s *roleMenuService) UpdateRoleMenu(id uint, req *UpdateRoleMenuRequest) (*models.RoleMenu, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid role menu ID")
	}

	// Get existing role menu
	roleMenu, err := s.roleMenuRepo.GetByID(id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to get role menu for update")
		return nil, err
	}

	// Update fields if provided
	if req.DocumentID != nil {
		roleMenu.DocumentID = req.DocumentID
	}
	if req.RoleMenuOrd != nil {
		roleMenu.RoleMenuOrd = req.RoleMenuOrd
	}
	if req.IsActive != nil {
		roleMenu.IsActive = req.IsActive
	}

	err = s.roleMenuRepo.Update(roleMenu)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to update role menu")
		return nil, err
	}

	s.logger.WithField("id", id).Info("Role menu updated successfully")

	// Return with relations
	return s.roleMenuRepo.GetWithRelations(id)
}

// DeleteRoleMenu deletes a role menu
func (s *roleMenuService) DeleteRoleMenu(id uint) error {
	if id == 0 {
		return fmt.Errorf("invalid role menu ID")
	}

	// Check if role menu exists
	_, err := s.roleMenuRepo.GetByID(id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Role menu not found for deletion")
		return err
	}

	err = s.roleMenuRepo.Delete(id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to delete role menu")
		return err
	}

	s.logger.WithField("id", id).Info("Role menu deleted successfully")
	return nil
}

// GetRoleMenusByRoleID retrieves role menus by role ID
func (s *roleMenuService) GetRoleMenusByRoleID(roleID uint) ([]models.RoleMenu, error) {
	if roleID == 0 {
		return nil, fmt.Errorf("invalid role ID")
	}

	roleMenus, err := s.roleMenuRepo.GetByRoleID(roleID)
	if err != nil {
		s.logger.WithError(err).WithField("role_id", roleID).Error("Failed to get role menus by role ID")
		return nil, err
	}

	return roleMenus, nil
}

// AttachMasterMenuToRoleMenu attaches a master menu to a role menu
func (s *roleMenuService) AttachMasterMenuToRoleMenu(roleMenuID, masterMenuID uint, order *float64) error {
	if roleMenuID == 0 || masterMenuID == 0 {
		return fmt.Errorf("invalid role menu ID or master menu ID")
	}

	// Verify that both role menu and master menu exist
	_, err := s.roleMenuRepo.GetByID(roleMenuID)
	if err != nil {
		return fmt.Errorf("role menu not found")
	}

	_, err = s.masterMenuRepo.GetByID(masterMenuID)
	if err != nil {
		return fmt.Errorf("master menu not found")
	}

	err = s.roleMenuRepo.AttachMasterMenu(roleMenuID, masterMenuID, order)
	if err != nil {
		s.logger.WithError(err).WithFields(map[string]interface{}{
			"role_menu_id":   roleMenuID,
			"master_menu_id": masterMenuID,
		}).Error("Failed to attach master menu to role menu")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"role_menu_id":   roleMenuID,
		"master_menu_id": masterMenuID,
	}).Info("Master menu attached to role menu successfully")

	return nil
}

// DetachMasterMenuFromRoleMenu detaches a master menu from a role menu
func (s *roleMenuService) DetachMasterMenuFromRoleMenu(roleMenuID, masterMenuID uint) error {
	if roleMenuID == 0 || masterMenuID == 0 {
		return fmt.Errorf("invalid role menu ID or master menu ID")
	}

	err := s.roleMenuRepo.DetachMasterMenu(roleMenuID, masterMenuID)
	if err != nil {
		s.logger.WithError(err).WithFields(map[string]interface{}{
			"role_menu_id":   roleMenuID,
			"master_menu_id": masterMenuID,
		}).Error("Failed to detach master menu from role menu")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"role_menu_id":   roleMenuID,
		"master_menu_id": masterMenuID,
	}).Info("Master menu detached from role menu successfully")

	return nil
}

// AttachRoleToRoleMenu attaches a role to a role menu
func (s *roleMenuService) AttachRoleToRoleMenu(roleMenuID, roleID uint, order *float64) error {
	if roleMenuID == 0 || roleID == 0 {
		return fmt.Errorf("invalid role menu ID or role ID")
	}

	// Verify that role menu exists
	_, err := s.roleMenuRepo.GetByID(roleMenuID)
	if err != nil {
		return fmt.Errorf("role menu not found")
	}

	err = s.roleMenuRepo.AttachRole(roleMenuID, roleID, order)
	if err != nil {
		s.logger.WithError(err).WithFields(map[string]interface{}{
			"role_menu_id": roleMenuID,
			"role_id":      roleID,
		}).Error("Failed to attach role to role menu")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"role_menu_id": roleMenuID,
		"role_id":      roleID,
	}).Info("Role attached to role menu successfully")

	return nil
}

// DetachRoleFromRoleMenu detaches a role from a role menu
func (s *roleMenuService) DetachRoleFromRoleMenu(roleMenuID, roleID uint) error {
	if roleMenuID == 0 || roleID == 0 {
		return fmt.Errorf("invalid role menu ID or role ID")
	}

	err := s.roleMenuRepo.DetachRole(roleMenuID, roleID)
	if err != nil {
		s.logger.WithError(err).WithFields(map[string]interface{}{
			"role_menu_id": roleMenuID,
			"role_id":      roleID,
		}).Error("Failed to detach role from role menu")
		return err
	}

	s.logger.WithFields(map[string]interface{}{
		"role_menu_id": roleMenuID,
		"role_id":      roleID,
	}).Info("Role detached from role menu successfully")

	return nil
}

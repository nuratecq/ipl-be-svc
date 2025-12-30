package service

import (
	"errors"

	"ipl-be-svc/internal/models"
	"ipl-be-svc/internal/repository"
)

// MenuService interface defines menu service methods
type MenuService interface {
	GetMenusByUserID(userID uint) ([]*models.MasterMenu, error)
}

// menuService implements MenuService interface
type menuService struct {
	menuRepo repository.MenuRepository
}

// NewMenuService creates a new menu service
func NewMenuService(menuRepo repository.MenuRepository) MenuService {
	return &menuService{
		menuRepo: menuRepo,
	}
}

// GetMenusByUserID gets menus by user ID with business logic validation
func (s *menuService) GetMenusByUserID(userID uint) ([]*models.MasterMenu, error) {
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}

	menus, err := s.menuRepo.GetMenusByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Business logic: filter only active and published menus
	var activeMenus []*models.MasterMenu
	for _, menu := range menus {
		// Check if menu is active (handle nullable boolean)
		isActive := menu.IsActive != nil && *menu.IsActive
		if isActive {
			activeMenus = append(activeMenus, menu)
		}
	}

	return activeMenus, nil
}

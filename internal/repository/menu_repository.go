package repository

import (
	"gorm.io/gorm"

	"ipl-be-svc/internal/models"
)

// MenuRepository interface defines menu repository methods
type MenuRepository interface {
	GetMenusByUserID(userID uint) ([]*models.MasterMenu, error)
}

// menuRepository implements MenuRepository interface
type menuRepository struct {
	db *gorm.DB
}

// NewMenuRepository creates a new menu repository
func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

// GetMenusByUserID gets distinct menus by user ID using the provided SQL query
func (r *menuRepository) GetMenusByUserID(userID uint) ([]*models.MasterMenu, error) {
	var menus []*models.MasterMenu

	query := `
		SELECT DISTINCT ON (mm.document_id) mm.*
		FROM up_users_role_lnk uurl
		INNER JOIN role_menus_role_lnk rmrl ON rmrl.role_id = uurl.role_id
		INNER JOIN role_menus_master_menu_lnk rmmml ON rmrl.role_menu_id = rmmml.role_menu_id
		INNER JOIN master_menus mm ON rmmml.master_menu_id = mm.id
		WHERE uurl.user_id = ?
		ORDER BY mm.document_id, mm.id
	`

	err := r.db.Raw(query, userID).Scan(&menus).Error
	return menus, err
}

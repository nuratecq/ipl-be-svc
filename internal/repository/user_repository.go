package repository

import (
	"ipl-be-svc/internal/models"

	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetUserDetailByProfileID(profileID uint) (*models.UserDetail, error)
	GetUsersWithPenghuniRole() ([]*models.UserDetail, error)
}

// userRepository implements UserRepository
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// GetUserDetailByProfileID retrieves user detail by profile ID
func (r *userRepository) GetUserDetailByProfileID(profileID uint) (*models.UserDetail, error) {
	var userDetail models.UserDetail

	query := `
		select p.id, p.nama_penghuni, p.no_hp, p.no_telp, p.document_id,
			   uu.email, uu.id as user_id,
			   ur."name", ur.id as role_id, ur."type" as role_type
		from profiles p
		inner join up_users_profile_lnk pul on p.id = pul.profile_id
		inner join up_users uu on uu.id = pul.user_id 
		inner join up_users_role_lnk uurl on uurl.user_id = uu.id
		inner join up_roles ur on ur.id = uurl.role_id
		where uu.id = ?
		limit 1
	`

	err := r.db.Raw(query, profileID).Scan(&userDetail).Error
	if err != nil {
		return nil, err
	}

	return &userDetail, nil
}

// GetUsersWithPenghuniRole retrieves all users with role type "penghuni"
func (r *userRepository) GetUsersWithPenghuniRole() ([]*models.UserDetail, error) {
	var users []*models.UserDetail

	query := `
		select uu.id, uu.username, uu.email,
			   p.nama_penghuni, p.no_hp, p.no_telp, p.document_id,
			   ur."name" as role_name, ur.id as role_id, ur."type" as role_type,
			   uu.id as user_id
		from up_users uu
		inner join up_users_role_lnk uurl on uurl.user_id = uu.id
		inner join up_roles ur on ur.id = uurl.role_id
		left join up_users_profile_lnk pul on pul.user_id = uu.id
		left join profiles p on p.id = pul.profile_id 
		where ur."type" = 'penghuni' AND p.published_at IS NOT NULL
		order by uu.id
	`

	err := r.db.Raw(query).Scan(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

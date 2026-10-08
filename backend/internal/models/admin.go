package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Admin roles
type AdminRole string

const (
	RoleSuperAdmin AdminRole = "super_admin"
	RoleAdmin      AdminRole = "admin"
	RoleModerator  AdminRole = "moderator"
	RoleSales      AdminRole = "sales"
)

// Admin represents a platform administrator
type Admin struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email     string         `gorm:"type:text;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"type:text;not null" json:"-"`
	Name      string         `gorm:"type:text;not null" json:"name"`
	Role      AdminRole      `gorm:"type:text;not null;default:'admin'" json:"role"`
	IsActive  bool           `gorm:"default:true" json:"is_active"`
	LastLogin *time.Time     `json:"last_login,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Admin) TableName() string {
	return "admins"
}

// AdminLoginInput for login request
type AdminLoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// CreateAdminInput for creating new admin
type CreateAdminInput struct {
	Email    string    `json:"email" binding:"required,email"`
	Password string    `json:"password" binding:"required,min=6"`
	Name     string    `json:"name" binding:"required"`
	Role     AdminRole `json:"role" binding:"required"`
}

// UpdateAdminInput for updating admin
type UpdateAdminInput struct {
	Email    string    `json:"email" binding:"omitempty,email"`
	Password string    `json:"password" binding:"omitempty,min=6"`
	Name     string    `json:"name"`
	Role     AdminRole `json:"role"`
	IsActive *bool     `json:"is_active"`
}

// Permission checks
func (a *Admin) CanManageAdmins() bool {
	return a.Role == RoleSuperAdmin
}

func (a *Admin) CanManageSettings() bool {
	return a.Role == RoleSuperAdmin
}

func (a *Admin) CanEditUsers() bool {
	return a.Role == RoleSuperAdmin || a.Role == RoleAdmin
}

func (a *Admin) CanSuspendStores() bool {
	return a.Role == RoleSuperAdmin || a.Role == RoleAdmin || a.Role == RoleModerator
}

func (a *Admin) CanViewAnalytics() bool {
	return true // All roles can view
}

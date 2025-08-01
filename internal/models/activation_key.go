package models

import (
	"time"

	"gorm.io/gorm"
)

// ActivationKey 激活码模型
type ActivationKey struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Code      string         `json:"code" gorm:"uniqueIndex;not null"`
	Type      string         `json:"type" gorm:"not null"`           // premium, standard, trial
	Status    string         `json:"status" gorm:"default:'unused'"` // unused, used, expired, revoked
	MaxUses   int            `json:"max_uses" gorm:"default:1"`
	UsedCount int            `json:"used_count" gorm:"default:0"`
	ExpiresAt *time.Time     `json:"expires_at"`
	CreatedBy string         `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Activation 激活记录模型
type Activation struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	KeyID       uint           `json:"key_id" gorm:"not null"`
	Key         ActivationKey  `json:"key" gorm:"foreignKey:KeyID"`
	UserEmail   string         `json:"user_email" gorm:"not null"`
	UserID      string         `json:"user_id"`
	DeviceID    string         `json:"device_id"`
	IPAddress   string         `json:"ip_address"`
	UserAgent   string         `json:"user_agent"`
	ActivatedAt time.Time      `json:"activated_at"`
	ExpiresAt   *time.Time     `json:"expires_at"`
	Status      string         `json:"status" gorm:"default:'active'"` // active, expired, revoked
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Role      string         `json:"role" gorm:"default:'user'"`     // admin, user
	Status    string         `json:"status" gorm:"default:'active'"` // active, inactive
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (ActivationKey) TableName() string {
	return "activation_keys"
}

func (Activation) TableName() string {
	return "activations"
}

func (User) TableName() string {
	return "users"
}

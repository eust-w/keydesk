package activation

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/eust-w/keydesk/internal/models"
	"gorm.io/gorm"
)

// Manager 激活码管理器
type Manager struct {
	db *gorm.DB
}

// NewManager 创建激活码管理器
func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// GenerateKeyRequest 生成激活码请求
type GenerateKeyRequest struct {
	Type      string     `json:"type"`       // premium, standard, trial
	MaxUses   int        `json:"max_uses"`   // 最大使用次数
	ExpiresAt *time.Time `json:"expires_at"` // 过期时间
	CreatedBy string     `json:"created_by"` // 创建者
	KeyLength int        `json:"key_length"` // 激活码长度
	KeyPrefix string     `json:"key_prefix"` // 激活码前缀
}

// ActivateRequest 激活请求
type ActivateRequest struct {
	Code      string `json:"code" binding:"required"`
	UserEmail string `json:"user_email" binding:"required,email"`
	UserID    string `json:"user_id"`
	DeviceID  string `json:"device_id"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

// GenerateKey 生成激活码
func (m *Manager) GenerateKey(req *GenerateKeyRequest) (*models.ActivationKey, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if req.MaxUses < 0 {
		return nil, fmt.Errorf("max_uses must be positive")
	}
	// 设置默认值
	if req.KeyLength == 0 {
		req.KeyLength = 16
	}
	if req.KeyPrefix == "" {
		req.KeyPrefix = "KD"
	}
	if req.MaxUses == 0 {
		req.MaxUses = 1
	}

	// 生成唯一激活码
	code, err := m.generateUniqueCode(req.KeyPrefix, req.KeyLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique code: %w", err)
	}

	// 设置过期时间
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		expiresAt = req.ExpiresAt
	} else {
		// 默认一年后过期
		defaultExpiry := time.Now().AddDate(1, 0, 0)
		expiresAt = &defaultExpiry
	}

	key := &models.ActivationKey{
		Code:      code,
		Type:      req.Type,
		Status:    "unused",
		MaxUses:   req.MaxUses,
		UsedCount: 0,
		ExpiresAt: expiresAt,
		CreatedBy: req.CreatedBy,
	}

	if err := m.db.Create(key).Error; err != nil {
		return nil, fmt.Errorf("failed to create activation key: %w", err)
	}

	return key, nil
}

// ValidateKey 验证激活码
func (m *Manager) ValidateKey(code string) (bool, error) {
	var key models.ActivationKey
	err := m.db.Where("code = ?", code).First(&key).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to query activation key: %w", err)
	}

	// 检查状态
	if key.Status != "unused" {
		return false, nil
	}

	// 检查使用次数
	if key.UsedCount >= key.MaxUses {
		return false, nil
	}

	// 检查过期时间
	if key.ExpiresAt != nil && time.Now().After(*key.ExpiresAt) {
		return false, nil
	}

	return true, nil
}

// ActivateKey 激活激活码
func (m *Manager) ActivateKey(req *ActivateRequest) (*models.Activation, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	var activation *models.Activation
	err := m.db.Transaction(func(tx *gorm.DB) error {
		// The conditional write serializes capacity reservation on this key.
		now := time.Now()
		result := tx.Model(&models.ActivationKey{}).
			Where("code = ? AND status = ? AND used_count < max_uses AND (expires_at IS NULL OR expires_at > ?)", req.Code, "unused", now).
			Updates(map[string]interface{}{"used_count": gorm.Expr("used_count + 1"), "status": gorm.Expr("CASE WHEN used_count + 1 >= max_uses THEN 'used' ELSE 'unused' END")})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("invalid or exhausted activation key")
		}
		var key models.ActivationKey
		if err := tx.Where("code = ?", req.Code).First(&key).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&models.Activation{}).Where("key_id = ? AND user_email = ?", key.ID, req.UserEmail).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("activation key already used by this email")
		}
		activation = &models.Activation{KeyID: key.ID, UserEmail: req.UserEmail, UserID: req.UserID, DeviceID: req.DeviceID, IPAddress: req.IPAddress, UserAgent: req.UserAgent, ActivatedAt: now, ExpiresAt: key.ExpiresAt, Status: "active"}
		return tx.Create(activation).Error
	})
	if err != nil {
		return nil, err
	}
	return activation, nil
}

// GetKeyInfo 获取激活码信息
func (m *Manager) GetKeyInfo(code string) (*models.ActivationKey, error) {
	var key models.ActivationKey
	err := m.db.Where("code = ?", code).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// GetActivations 获取激活记录
func (m *Manager) GetActivations(keyID uint) ([]models.Activation, error) {
	var activations []models.Activation
	err := m.db.Where("key_id = ?", keyID).Find(&activations).Error
	return activations, err
}

// RevokeKey 撤销激活码
func (m *Manager) RevokeKey(code string) error {
	return m.db.Model(&models.ActivationKey{}).Where("code = ?", code).Update("status", "revoked").Error
}

// generateUniqueCode 生成唯一激活码
func (m *Manager) generateUniqueCode(prefix string, length int) (string, error) {
	if length <= len(prefix) || length > 256 {
		return "", fmt.Errorf("key length must exceed prefix length and be at most 256")
	}
	randomLength := length - len(prefix)
	maxAttempts := 100
	for i := 0; i < maxAttempts; i++ {
		// 生成随机字节
		bytes := make([]byte, (randomLength+1)/2)
		if _, err := rand.Read(bytes); err != nil {
			return "", err
		}

		// 转换为十六进制字符串
		code := prefix + hex.EncodeToString(bytes)[:length-len(prefix)]

		// 检查是否已存在
		var count int64
		err := m.db.Model(&models.ActivationKey{}).Where("code = ?", code).Count(&count).Error
		if err != nil {
			return "", err
		}

		if count == 0 {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

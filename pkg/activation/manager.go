package activation

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/keydesk/activation/internal/models"
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
	// 开始事务
	tx := m.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 查找激活码
	var key models.ActivationKey
	err := tx.Where("code = ?", req.Code).First(&key).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("activation key not found")
		}
		return nil, fmt.Errorf("failed to query activation key: %w", err)
	}

	// 验证激活码
	valid, err := m.ValidateKey(req.Code)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if !valid {
		tx.Rollback()
		return nil, fmt.Errorf("invalid activation key")
	}

	// 检查是否已经激活过
	var existingActivation models.Activation
	err = tx.Where("key_id = ? AND user_email = ?", key.ID, req.UserEmail).First(&existingActivation).Error
	if err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("activation key already used by this email")
	}

	// 创建激活记录
	activation := &models.Activation{
		KeyID:       key.ID,
		UserEmail:   req.UserEmail,
		UserID:      req.UserID,
		DeviceID:    req.DeviceID,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
		ActivatedAt: time.Now(),
		ExpiresAt:   key.ExpiresAt,
		Status:      "active",
	}

	if err := tx.Create(activation).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create activation record: %w", err)
	}

	// 更新激活码使用次数
	key.UsedCount++
	if key.UsedCount >= key.MaxUses {
		key.Status = "used"
	}
	if err := tx.Save(&key).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update activation key: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
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
	maxAttempts := 100
	for i := 0; i < maxAttempts; i++ {
		// 生成随机字节
		bytes := make([]byte, length/2)
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

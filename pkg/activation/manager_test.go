package activation

import (
	"testing"

	"github.com/keydesk/activation/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&models.ActivationKey{}, &models.Activation{}, &models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// 验证表是否创建成功
	if !db.Migrator().HasTable(&models.ActivationKey{}) {
		t.Fatal("ActivationKey table was not created")
	}

	return db
}

func TestManager_GenerateKey(t *testing.T) {
	db := setupTestDB(t)
	manager := NewManager(db)

	req := &GenerateKeyRequest{
		Type:      "premium",
		MaxUses:   1,
		CreatedBy: "test",
	}

	key, err := manager.GenerateKey(req)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if key.Code == "" {
		t.Error("Generated key code is empty")
	}

	if key.Type != "premium" {
		t.Errorf("Expected key type 'premium', got '%s'", key.Type)
	}

	if key.Status != "unused" {
		t.Errorf("Expected key status 'unused', got '%s'", key.Status)
	}

	if key.MaxUses != 1 {
		t.Errorf("Expected max uses 1, got %d", key.MaxUses)
	}
}

func TestManager_ValidateKey(t *testing.T) {
	db := setupTestDB(t)
	manager := NewManager(db)

	// 生成一个激活码
	req := &GenerateKeyRequest{
		Type:      "standard",
		MaxUses:   1,
		CreatedBy: "test",
	}

	key, err := manager.GenerateKey(req)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// 测试有效激活码
	valid, err := manager.ValidateKey(key.Code)
	if err != nil {
		t.Fatalf("Failed to validate key: %v", err)
	}

	if !valid {
		t.Error("Valid key should return true")
	}

	// 测试无效激活码
	valid, err = manager.ValidateKey("INVALID-CODE")
	if err != nil {
		t.Fatalf("Failed to validate invalid key: %v", err)
	}

	if valid {
		t.Error("Invalid key should return false")
	}
}

func TestManager_ActivateKey(t *testing.T) {
	db := setupTestDB(t)
	manager := NewManager(db)

	// 生成一个激活码
	req := &GenerateKeyRequest{
		Type:      "trial",
		MaxUses:   1,
		CreatedBy: "test",
	}

	key, err := manager.GenerateKey(req)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	// 激活激活码
	activateReq := &ActivateRequest{
		Code:      key.Code,
		UserEmail: "test@example.com",
		UserID:    "user123",
		DeviceID:  "device456",
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
	}

	activation, err := manager.ActivateKey(activateReq)
	if err != nil {
		t.Fatalf("Failed to activate key: %v", err)
	}

	if activation.UserEmail != "test@example.com" {
		t.Errorf("Expected user email 'test@example.com', got '%s'", activation.UserEmail)
	}

	if activation.Status != "active" {
		t.Errorf("Expected activation status 'active', got '%s'", activation.Status)
	}
}

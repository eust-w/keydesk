package main

import (
	"fmt"
	"log"
	"time"

	"github.com/keydesk/activation/pkg/activation"
	"github.com/keydesk/activation/pkg/database"
)

func main() {
	// 连接数据库
	db, err := database.NewConnectionFromDSN("postgres://user:password@localhost/keydesk?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建激活码管理器
	manager := activation.NewManager(db)

	// 示例1: 生成激活码
	fmt.Println("=== 生成激活码 ===")
	key, err := manager.GenerateKey(&activation.GenerateKeyRequest{
		Type:      "premium",
		MaxUses:   1,
		CreatedBy: "admin",
	})
	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}
	fmt.Printf("生成的激活码: %s\n", key.Code)
	fmt.Printf("类型: %s\n", key.Type)
	fmt.Printf("状态: %s\n", key.Status)

	// 示例2: 验证激活码
	fmt.Println("\n=== 验证激活码 ===")
	valid, err := manager.ValidateKey(key.Code)
	if err != nil {
		log.Fatalf("Failed to validate key: %v", err)
	}
	fmt.Printf("激活码是否有效: %t\n", valid)

	// 示例3: 激活激活码
	fmt.Println("\n=== 激活激活码 ===")
	activation, err := manager.ActivateKey(&activation.ActivateRequest{
		Code:      key.Code,
		UserEmail: "user@example.com",
		UserID:    "user123",
		DeviceID:  "device456",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0...",
	})
	if err != nil {
		log.Fatalf("Failed to activate key: %v", err)
	}
	fmt.Printf("激活成功! 激活ID: %d\n", activation.ID)
	fmt.Printf("用户邮箱: %s\n", activation.UserEmail)
	fmt.Printf("激活时间: %s\n", activation.ActivatedAt.Format(time.RFC3339))

	// 示例4: 再次验证已使用的激活码
	fmt.Println("\n=== 验证已使用的激活码 ===")
	valid, err = manager.ValidateKey(key.Code)
	if err != nil {
		log.Fatalf("Failed to validate key: %v", err)
	}
	fmt.Printf("激活码是否有效: %t\n", valid)

	// 示例5: 获取激活码信息
	fmt.Println("\n=== 获取激活码信息 ===")
	keyInfo, err := manager.GetKeyInfo(key.Code)
	if err != nil {
		log.Fatalf("Failed to get key info: %v", err)
	}
	fmt.Printf("激活码: %s\n", keyInfo.Code)
	fmt.Printf("使用次数: %d/%d\n", keyInfo.UsedCount, keyInfo.MaxUses)
	fmt.Printf("状态: %s\n", keyInfo.Status)

	// 示例6: 获取激活记录
	fmt.Println("\n=== 获取激活记录 ===")
	activations, err := manager.GetActivations(keyInfo.ID)
	if err != nil {
		log.Fatalf("Failed to get activations: %v", err)
	}
	for _, act := range activations {
		fmt.Printf("激活记录ID: %d, 用户: %s, 时间: %s\n",
			act.ID, act.UserEmail, act.ActivatedAt.Format(time.RFC3339))
	}

	fmt.Println("\n=== 示例完成 ===")
}

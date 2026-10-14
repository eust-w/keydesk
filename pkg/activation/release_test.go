package activation

import (
	"fmt"
	"github.com/eust-w/keydesk/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"sync"
	"sync/atomic"
	"testing"
)

func TestKeyLengthBoundsAndRepeatedActivation(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db)
	for _, length := range []int{-1, 1, 2, 257} {
		if _, err := m.generateUniqueCode("KD", length); err == nil {
			t.Fatalf("accepted length %d", length)
		}
	}
	for _, length := range []int{3, 15, 17} {
		s, err := m.generateUniqueCode("KD", length)
		if err != nil || len(s) != length {
			t.Fatalf("length %d: %s %v", length, s, err)
		}
	}
	key, err := m.GenerateKey(&GenerateKeyRequest{Type: "trial", MaxUses: 2})
	if err != nil {
		t.Fatal(err)
	}
	req := &ActivateRequest{Code: key.Code, UserEmail: "first@example.com"}
	if _, err = m.ActivateKey(req); err != nil {
		t.Fatal(err)
	}
	if _, err = m.ActivateKey(req); err == nil {
		t.Fatal("duplicate accepted")
	}
	if _, err = m.ActivateKey(&ActivateRequest{Code: key.Code, UserEmail: "second@example.com"}); err != nil {
		t.Fatal(err)
	}
	if _, err = m.ActivateKey(&ActivateRequest{Code: key.Code, UserEmail: "third@example.com"}); err == nil {
		t.Fatal("exhausted key accepted")
	}
}
func TestPostgresConcurrentFinalSlot(t *testing.T) {
	dsn := os.Getenv("KEYDESK_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set KEYDESK_TEST_POSTGRES_DSN for PostgreSQL acceptance")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&models.ActivationKey{}, &models.Activation{}, &models.User{}); err != nil {
		t.Fatal(err)
	}
	m := NewManager(db)
	key, err := m.GenerateKey(&GenerateKeyRequest{Type: "test", MaxUses: 1})
	if err != nil {
		t.Fatal(err)
	}
	var wins atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, e := m.ActivateKey(&ActivateRequest{Code: key.Code, UserEmail: fmt.Sprintf("race-%d@example.com", i)}); e == nil {
				wins.Add(1)
			}
		}(i)
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("%d activations succeeded", wins.Load())
	}
	var count int64
	db.Model(&models.Activation{}).Where("key_id = ?", key.ID).Count(&count)
	if count != 1 {
		t.Fatalf("%d records", count)
	}
}

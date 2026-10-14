package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eust-w/keydesk/internal/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestPostgresHTTPPermissionsAndActivation(t *testing.T) {
	old, _ := os.Getwd()
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	dsn := os.Getenv("KEYDESK_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PostgreSQL acceptance fixture not configured")
	}
	db, e := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&models.User{}, &models.ActivationKey{}, &models.Activation{}); e != nil {
		t.Fatal(e)
	}
	cfg := &Config{}
	cfg.Server.Mode = "test"
	cfg.JWT.Secret = "local-test-secret-not-for-production"
	cfg.JWT.ExpiresIn = "1h"
	r := setupRouter(cfg, db, logrus.New())
	stamp := time.Now().UnixNano()
	email := fmt.Sprintf("member%d@example.com", stamp)
	admin := fmt.Sprintf("admin%d@example.com", stamp)
	call := func(method, path string, body interface{}, token string) (int, map[string]interface{}) {
		b, _ := json.Marshal(body)
		q := httptest.NewRequest(method, path, bytes.NewReader(b))
		q.Header.Set("Content-Type", "application/json")
		if token != "" {
			q.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, q)
		v := map[string]interface{}{}
		_ = json.Unmarshal(w.Body.Bytes(), &v)
		return w.Code, v
	}
	if status, _ := call("GET", "/api/v1/health", nil, ""); status != 200 {
		t.Fatalf("health %d", status)
	}
	if status, _ := call("POST", "/api/v1/users/register", map[string]string{"email": email, "password": "valid-test-password"}, ""); status != 201 {
		t.Fatalf("register %d", status)
	}
	login := func(email string) string {
		s, v := call("POST", "/api/v1/users/login", map[string]string{"email": email, "password": "valid-test-password"}, "")
		if s != 200 {
			t.Fatalf("login %d", s)
		}
		return v["data"].(map[string]interface{})["token"].(string)
	}
	token := login(email)
	req := map[string]interface{}{"type": "trial", "max_uses": 1}
	if s, _ := call("POST", "/api/v1/keys/generate", req, token); s != 403 {
		t.Fatalf("ordinary generation %d", s)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("valid-test-password"), bcrypt.DefaultCost)
	if e = db.Create(&models.User{Email: admin, Password: string(hash), Role: "admin", Status: "active"}).Error; e != nil {
		t.Fatal(e)
	}
	adminToken := login(admin)
	if status, _ := call("POST", "/api/v1/keys/generate", map[string]interface{}{"type": "trial", "key_length": -1}, adminToken); status != 400 {
		t.Fatalf("invalid length status %d", status)
	}
	s, v := call("POST", "/api/v1/keys/generate", req, adminToken)
	if s != 201 {
		t.Fatalf("admin generation %d: %v", s, v)
	}
	code := v["data"].(map[string]interface{})["code"].(string)
	s, _ = call("POST", "/api/v1/activate", map[string]string{"code": code, "user_email": email}, "")
	if s != 200 && s != 201 {
		t.Fatalf("activate %d", s)
	}
	s, _ = call("POST", "/api/v1/activate", map[string]string{"code": code, "user_email": email}, "")
	if s < 400 {
		t.Fatalf("duplicate status %d", s)
	}
}

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthenticationRequiresExpirationAndHS256(t *testing.T) {
	secret := "unit-test-signing-secret"
	cases := []struct {
		name   string
		method jwt.SigningMethod
		expiry *jwt.NumericDate
		want   int
	}{
		{"valid", jwt.SigningMethodHS256, jwt.NewNumericDate(time.Now().Add(time.Hour)), 200},
		{"expired", jwt.SigningMethodHS256, jwt.NewNumericDate(time.Now().Add(-time.Hour)), 401},
		{"legacy-no-expiry", jwt.SigningMethodHS256, nil, 401},
		{"wrong-algorithm", jwt.SigningMethodHS384, jwt.NewNumericDate(time.Now().Add(time.Hour)), 401},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwt.NewWithClaims(tt.method, &Claims{UserID: "1", RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: tt.expiry}}).SignedString([]byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			r := gin.New()
			r.GET("/", AuthMiddleware(&JWTConfig{Secret: secret}, logrus.New()), func(c *gin.Context) { c.Status(200) })
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("status %d, want %d", w.Code, tt.want)
			}
		})
	}
	token, err := GenerateToken("1", "test@example.com", "user", secret, 30*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	claims := &Claims{}
	_, err = jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) { return []byte(secret), nil })
	if err != nil || claims.ExpiresAt == nil {
		t.Fatalf("missing expiration: %v", err)
	}
}

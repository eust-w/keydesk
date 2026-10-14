package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eust-w/keydesk/internal/middleware"
	"github.com/eust-w/keydesk/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserHandler 用户处理器
type UserHandler struct {
	db        *gorm.DB
	logger    *logrus.Logger
	jwtSecret string
	jwtTTL    time.Duration
}

// NewUserHandler 创建用户处理器
func NewUserHandler(db *gorm.DB, logger *logrus.Logger, jwtSecret string, durations ...time.Duration) *UserHandler {
	ttl := 24 * time.Hour
	if len(durations) > 0 {
		ttl = durations[0]
	}
	return &UserHandler{
		db:        db,
		logger:    logger,
		jwtSecret: jwtSecret,
		jwtTTL:    ttl,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// 检查邮箱是否已存在
	var existingUser models.User
	err := h.db.Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already exists",
		})
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process registration",
		})
		return
	}

	// 创建用户
	user := &models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user",
		Status:   "active",
	}

	if err := h.db.Create(user).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"data": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// 查找用户
	var user models.User
	err := h.db.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Account is not active",
		})
		return
	}

	// 生成JWT token
	token, err := middleware.GenerateToken(
		fmt.Sprintf("%d", user.ID),
		user.Email,
		user.Role,
		h.jwtSecret,
		h.jwtTTL,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data": gin.H{
			"token": token,
			"user": gin.H{
				"id":    user.ID,
				"email": user.Email,
				"role":  user.Role,
			},
		},
	})
}

// GetProfile 获取用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var user models.User
	err := h.db.First(&user, userID).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"role":       user.Role,
			"status":     user.Status,
			"created_at": user.CreatedAt,
		},
	})
}

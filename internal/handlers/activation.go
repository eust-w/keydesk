package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/keydesk/activation/internal/models"
	"github.com/keydesk/activation/pkg/activation"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ActivationHandler 激活码处理器
type ActivationHandler struct {
	manager *activation.Manager
	db      *gorm.DB
	logger  *logrus.Logger
}

// NewActivationHandler 创建激活码处理器
func NewActivationHandler(manager *activation.Manager, db *gorm.DB, logger *logrus.Logger) *ActivationHandler {
	return &ActivationHandler{
		manager: manager,
		db:      db,
		logger:  logger,
	}
}

// GenerateKey 生成激活码
func (h *ActivationHandler) GenerateKey(c *gin.Context) {
	var req activation.GenerateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// 从JWT中获取用户信息
	userID, exists := c.Get("user_id")
	if exists {
		req.CreatedBy = userID.(string)
	}

	key, err := h.manager.GenerateKey(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate activation key")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate activation key",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    key,
	})
}

// ActivateKey 激活激活码
func (h *ActivationHandler) ActivateKey(c *gin.Context) {
	var req activation.ActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// 获取客户端信息
	req.IPAddress = c.ClientIP()
	req.UserAgent = c.GetHeader("User-Agent")

	activation, err := h.manager.ActivateKey(&req)
	if err != nil {
		h.logger.WithError(err).WithField("code", req.Code).Error("Failed to activate key")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to activate key",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activation,
		"message": "Activation successful",
	})
}

// ValidateKey 验证激活码
func (h *ActivationHandler) ValidateKey(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Activation code is required",
		})
		return
	}

	valid, err := h.manager.ValidateKey(code)
	if err != nil {
		h.logger.WithError(err).WithField("code", code).Error("Failed to validate key")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to validate key",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   valid,
	})
}

// GetKeyInfo 获取激活码信息
func (h *ActivationHandler) GetKeyInfo(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Activation code is required",
		})
		return
	}

	key, err := h.manager.GetKeyInfo(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Activation key not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    key,
	})
}

// ListKeys 获取激活码列表
func (h *ActivationHandler) ListKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	keyType := c.Query("type")

	offset := (page - 1) * limit

	// 构建查询条件
	query := h.db.Model(&models.ActivationKey{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if keyType != "" {
		query = query.Where("type = ?", keyType)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取分页数据
	var keys []models.ActivationKey
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&keys).Error
	if err != nil {
		h.logger.WithError(err).Error("Failed to get activation keys")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get activation keys",
			"message": err.Error(),
		})
		return
	}

	// 计算总页数
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"keys": keys,
			"pagination": gin.H{
				"page":     page,
				"limit":    limit,
				"total":    total,
				"pages":    totalPages,
				"has_next": page < totalPages,
				"has_prev": page > 1,
			},
		},
	})
}

// RevokeKey 撤销激活码
func (h *ActivationHandler) RevokeKey(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Activation code is required",
		})
		return
	}

	err := h.manager.RevokeKey(code)
	if err != nil {
		h.logger.WithError(err).WithField("code", code).Error("Failed to revoke key")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to revoke key",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Key revoked successfully",
	})
}

// GetActivations 获取激活记录
func (h *ActivationHandler) GetActivations(c *gin.Context) {
	keyIDStr := c.Param("key_id")
	keyID, err := strconv.ParseUint(keyIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid key ID",
		})
		return
	}

	activations, err := h.manager.GetActivations(uint(keyID))
	if err != nil {
		h.logger.WithError(err).WithField("key_id", keyID).Error("Failed to get activations")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get activations",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activations,
	})
}

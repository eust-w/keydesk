package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eust-w/keydesk/internal/handlers"
	"github.com/eust-w/keydesk/internal/middleware"
	"github.com/eust-w/keydesk/pkg/activation"
	"github.com/eust-w/keydesk/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// Config 应用配置
type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`

	Database database.Config `yaml:"database"`

	JWT struct {
		Secret    string `yaml:"secret"`
		ExpiresIn string `yaml:"expires_in"`
	} `yaml:"jwt"`

	Activation struct {
		KeyLength         int    `yaml:"key_length"`
		KeyPrefix         string `yaml:"key_prefix"`
		DefaultExpiryDays int    `yaml:"default_expiry_days"`
		MaxAttempts       int    `yaml:"max_attempts"`
	} `yaml:"activation"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		Output string `yaml:"output"`
	} `yaml:"logging"`
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if config.JWT.ExpiresIn == "" {
		config.JWT.ExpiresIn = "24h"
	}
	ttl, err := time.ParseDuration(config.JWT.ExpiresIn)
	if err != nil || ttl <= 0 {
		return nil, fmt.Errorf("jwt.expires_in must be a positive duration")
	}
	return &config, nil
}

// setupLogger 设置日志
func setupLogger(config *Config) *logrus.Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if config.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return logger
}

// setupRouter 设置路由
func setupRouter(config *Config, db *gorm.DB, logger *logrus.Logger) *gin.Engine {
	// 设置Gin模式
	gin.SetMode(config.Server.Mode)

	router := gin.New()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 创建激活码管理器
	manager := activation.NewManager(db)

	// 创建处理器
	activationHandler := handlers.NewActivationHandler(manager, db, logger)
	ttl, _ := time.ParseDuration(config.JWT.ExpiresIn)
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	userHandler := handlers.NewUserHandler(db, logger, config.JWT.Secret, ttl)

	// JWT中间件
	jwtConfig := &middleware.JWTConfig{Secret: config.JWT.Secret}
	authMiddleware := middleware.AuthMiddleware(jwtConfig, logger)
	adminMiddleware := middleware.AdminMiddleware()

	// API路由组
	api := router.Group("/api/v1")
	{
		// 用户认证路由
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)
			users.GET("/profile", authMiddleware, userHandler.GetProfile)
		}

		// 激活码相关路由
		keys := api.Group("/keys")
		{
			keys.POST("/generate", authMiddleware, adminMiddleware, activationHandler.GenerateKey)
			keys.GET("", authMiddleware, activationHandler.ListKeys)
			keys.GET("/:code", activationHandler.GetKeyInfo)
			keys.DELETE("/:code", authMiddleware, adminMiddleware, activationHandler.RevokeKey)
		}

		// 激活记录路由
		activations := api.Group("/activations")
		{
			activations.GET("/key/:key_id", authMiddleware, activationHandler.GetActivations)
		}

		// 激活相关路由
		activate := api.Group("/activate")
		{
			activate.POST("", activationHandler.ActivateKey)
			activate.GET("/:code", activationHandler.ValidateKey)
		}

		// 健康检查
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"time":   time.Now().Format(time.RFC3339),
			})
		})
	}

	// 静态文件服务
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// 前端页面路由
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "KeyDesk - 激活码管理系统",
		})
	})

	return router
}

func main() {
	// 加载配置
	configPath := "config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置日志
	logger := setupLogger(config)
	logger.Info("Starting KeyDesk activation server...")

	// 连接数据库
	db, err := database.NewConnection(&config.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	logger.Info("Database connected successfully")

	// 设置路由
	router := setupRouter(config, db, logger)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	logger.Infof("Server starting on %s", addr)

	if err := router.Run(addr); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}

# KeyDesk - 激活码管理系统

一个功能完整的Go语言激活码管理包，提供激活码生成、验证、管理和API接口。

## 功能特性

- 🔑 激活码生成和管理
- ✅ 激活码验证和激活
- 📊 用户激活记录追踪
- 🗄️ 数据库支持 (PostgreSQL)
- 🌐 RESTful API接口
- 🎨 现代化Web界面
- 🔒 JWT身份验证
- 📝 详细日志记录

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 配置数据库

1. 确保PostgreSQL已安装并运行
2. 创建数据库：
```sql
CREATE DATABASE keydesk;
```

### 环境配置

复制并修改配置文件：
```bash
cp config/config.example.yaml config/config.yaml
```

### 运行服务

```bash
go run cmd/server/main.go
```

## API接口

### 激活码管理

- `POST /api/v1/keys/generate` - 生成激活码
- `GET /api/v1/keys` - 获取激活码列表
- `DELETE /api/v1/keys/:id` - 删除激活码

### 激活验证

- `POST /api/v1/activate` - 验证并激活
- `GET /api/v1/activation/:key` - 查询激活状态

### 用户管理

- `POST /api/v1/users/register` - 用户注册
- `POST /api/v1/users/login` - 用户登录

## 使用示例

### 作为Go包使用

```go
package main

import (
    "github.com/keydesk/activation/pkg/activation"
    "github.com/keydesk/activation/pkg/database"
)

func main() {
    // 初始化数据库连接
    db, err := database.NewConnection("postgres://user:pass@localhost/keydesk")
    if err != nil {
        panic(err)
    }

    // 创建激活码管理器
    manager := activation.NewManager(db)

    // 生成激活码
    key, err := manager.GenerateKey("premium", 30) // 30天有效期
    if err != nil {
        panic(err)
    }

    // 验证激活码
    valid, err := manager.ValidateKey(key.Code)
    if err != nil {
        panic(err)
    }

    if valid {
        // 激活码有效，执行激活逻辑
        err = manager.ActivateKey(key.Code, "user@example.com")
    }
}
```

## 项目结构

```
keydesk/
├── cmd/                    # 命令行工具
│   └── server/            # 主服务器
├── config/                # 配置文件
├── internal/              # 内部包
│   ├── models/           # 数据模型
│   ├── handlers/         # HTTP处理器
│   └── middleware/       # 中间件
├── pkg/                   # 公共包
│   ├── activation/       # 激活码核心逻辑
│   ├── database/         # 数据库操作
│   └── utils/            # 工具函数
├── web/                   # 前端文件
├── migrations/            # 数据库迁移
└── docs/                  # 文档
```

## 许可证

MIT License
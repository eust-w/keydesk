# KeyDesk API 文档

## 概述

KeyDesk 提供完整的激活码管理 API，支持激活码生成、验证、激活和管理功能。

## 基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **内容类型**: `application/json`
- **认证方式**: JWT Bearer Token (部分接口需要)

## 响应格式

所有API响应都遵循以下格式：

```json
{
  "success": true,
  "data": {},
  "message": "操作成功"
}
```

错误响应：

```json
{
  "error": "错误类型",
  "message": "详细错误信息"
}
```

## 激活码管理

### 生成激活码

**POST** `/keys/generate`

生成新的激活码。

**请求体**:
```json
{
  "type": "premium",
  "max_uses": 1,
  "expires_at": "2024-12-31T23:59:59Z",
  "key_length": 16,
  "key_prefix": "KD"
}
```

**参数说明**:
- `type`: 激活码类型 (premium, standard, trial)
- `max_uses`: 最大使用次数 (默认: 1)
- `expires_at`: 过期时间 (可选，默认一年后)
- `key_length`: 激活码长度 (默认: 16)
- `key_prefix`: 激活码前缀 (默认: "KD")

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "code": "KD1234567890ABCDEF",
    "type": "premium",
    "status": "unused",
    "max_uses": 1,
    "used_count": 0,
    "expires_at": "2024-12-31T23:59:59Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 获取激活码列表

**GET** `/keys`

获取激活码列表，支持分页和筛选。

**查询参数**:
- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 10)
- `status`: 状态筛选 (unused, used, expired, revoked)
- `type`: 类型筛选 (premium, standard, trial)

**响应示例**:
```json
{
  "success": true,
  "data": {
    "keys": [
      {
        "id": 1,
        "code": "KD1234567890ABCDEF",
        "type": "premium",
        "status": "unused",
        "max_uses": 1,
        "used_count": 0,
        "expires_at": "2024-12-31T23:59:59Z",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "pages": 1
    }
  }
}
```

### 获取激活码信息

**GET** `/keys/{code}`

获取指定激活码的详细信息。

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "code": "KD1234567890ABCDEF",
    "type": "premium",
    "status": "unused",
    "max_uses": 1,
    "used_count": 0,
    "expires_at": "2024-12-31T23:59:59Z",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 撤销激活码

**DELETE** `/keys/{code}`

撤销指定的激活码。

**响应示例**:
```json
{
  "success": true,
  "message": "Key revoked successfully"
}
```

### 获取激活记录

**GET** `/keys/{key_id}/activations`

获取指定激活码的所有激活记录。

**响应示例**:
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "key_id": 1,
      "user_email": "user@example.com",
      "user_id": "user123",
      "device_id": "device456",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "activated_at": "2024-01-01T12:00:00Z",
      "expires_at": "2024-12-31T23:59:59Z",
      "status": "active",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

## 激活验证

### 验证激活码

**GET** `/activate/{code}`

验证激活码是否有效。

**响应示例**:
```json
{
  "success": true,
  "valid": true
}
```

### 激活激活码

**POST** `/activate`

使用激活码进行激活。

**请求体**:
```json
{
  "code": "KD1234567890ABCDEF",
  "user_email": "user@example.com",
  "user_id": "user123",
  "device_id": "device456"
}
```

**参数说明**:
- `code`: 激活码 (必需)
- `user_email`: 用户邮箱 (必需)
- `user_id`: 用户ID (可选)
- `device_id`: 设备ID (可选)

**响应示例**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "key_id": 1,
    "user_email": "user@example.com",
    "user_id": "user123",
    "device_id": "device456",
    "ip_address": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
    "activated_at": "2024-01-01T12:00:00Z",
    "expires_at": "2024-12-31T23:59:59Z",
    "status": "active"
  },
  "message": "Activation successful"
}
```

## 系统状态

### 健康检查

**GET** `/health`

检查系统健康状态。

**响应示例**:
```json
{
  "status": "ok",
  "time": "2024-01-01T12:00:00Z"
}
```

## 错误码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权 |
| 403 | 禁止访问 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 常见错误

### 激活码不存在
```json
{
  "error": "Activation key not found",
  "message": "activation key not found"
}
```

### 激活码已过期
```json
{
  "error": "Failed to activate key",
  "message": "invalid activation key"
}
```

### 激活码已被使用
```json
{
  "error": "Failed to activate key",
  "message": "activation key already used by this email"
}
```

### 达到最大使用次数
```json
{
  "error": "Failed to activate key",
  "message": "invalid activation key"
}
```

## 使用示例

### cURL 示例

生成激活码：
```bash
curl -X POST http://localhost:8080/api/v1/keys/generate \
  -H "Content-Type: application/json" \
  -d '{
    "type": "premium",
    "max_uses": 1
  }'
```

激活激活码：
```bash
curl -X POST http://localhost:8080/api/v1/activate \
  -H "Content-Type: application/json" \
  -d '{
    "code": "KD1234567890ABCDEF",
    "user_email": "user@example.com"
  }'
```

验证激活码：
```bash
curl http://localhost:8080/api/v1/activate/KD1234567890ABCDEF
```

### JavaScript 示例

```javascript
// 生成激活码
const generateKey = async () => {
  const response = await fetch('/api/v1/keys/generate', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      type: 'premium',
      max_uses: 1
    })
  });
  
  const data = await response.json();
  console.log('生成的激活码:', data.data.code);
};

// 激活激活码
const activateKey = async (code, email) => {
  const response = await fetch('/api/v1/activate', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      code: code,
      user_email: email
    })
  });
  
  const data = await response.json();
  if (data.success) {
    console.log('激活成功!');
  } else {
    console.error('激活失败:', data.message);
  }
};
``` 
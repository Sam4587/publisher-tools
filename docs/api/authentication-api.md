# API 认证系统文档

## 概述

账号管理系统现已集成完整的认证和授权功能，支持 JWT 和 API Key 两种认证方式。

## 认证方式

### 1. JWT 认证

JWT (JSON Web Token) 适用于用户登录场景，通过用户名密码获取 Token。

**获取 Token**:
```bash
POST /api/v1/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-02-22T16:00:00Z",
    "user": {
      "user_id": "user_abc123",
      "username": "your_username",
      "role": "user"
    }
  }
}
```

**使用 Token**:
```bash
# 方式 1: Authorization 头
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/accounts

# 方式 2: 查询参数
curl http://localhost:8080/api/v1/accounts?token=YOUR_TOKEN
```

### 2. API Key 认证

API Key 适用于服务间调用或自动化脚本。

**获取 API Key**:
1. 注册用户后自动生成
2. 登录后调用 `/api/v1/auth/me/api-key` 重新生成

**使用 API Key**:
```bash
curl -H "X-API-Key: pk_your_api_key_here" \
  http://localhost:8080/api/v1/accounts
```

## 用户管理 API

### 公开接口（无需认证）

#### 1. 用户注册

```http
POST /api/v1/register
Content-Type: application/json

{
  "username": "newuser",
  "email": "user@example.com",
  "password": "secure_password_123"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "user_id": "user_xyz789",
    "username": "newuser",
    "api_key": "pk_abc123...",
    "message": "User registered successfully"
  }
}
```

**要求**:
- 用户名：必填，唯一
- 密码：必填，至少 8 个字符
- 邮箱：可选，唯一

---

#### 2. 用户登录

```http
POST /api/v1/login
Content-Type: application/json

{
  "username": "your_username",
  "password": "your_password"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-02-22T16:00:00Z",
    "user": {
      "user_id": "user_abc123",
      "username": "your_username",
      "email": "user@example.com",
      "role": "user",
      "api_key": "pk_xyz...",
      "is_active": true
    }
  }
}
```

---

### 认证接口（需要认证）

#### 3. 获取当前用户信息

```http
GET /api/v1/auth/me
Authorization: Bearer YOUR_TOKEN
```

**响应**:
```json
{
  "success": true,
  "data": {
    "user_id": "user_abc123",
    "username": "your_username",
    "email": "user@example.com",
    "role": "user",
    "api_key": "pk_xyz...",
    "is_active": true,
    "last_login_at": "2026-02-21T16:00:00Z"
  }
}
```

---

#### 4. 更新用户信息

```http
PUT /api/v1/auth/me
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "email": "newemail@example.com",
  "password": "new_secure_password"
}
```

---

#### 5. 重新生成 API Key

```http
POST /api/v1/auth/me/api-key
Authorization: Bearer YOUR_TOKEN
```

**响应**:
```json
{
  "success": true,
  "data": {
    "api_key": "pk_new_key_here...",
    "message": "API Key regenerated successfully"
  }
}
```

**注意**: 重新生成后，旧的 API Key 将立即失效。

---

#### 6. 修改密码

```http
POST /api/v1/auth/change-password
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "old_password": "current_password",
  "new_password": "new_secure_password"
}
```

---

### 管理员接口（需要管理员权限）

#### 7. 列出用户

```http
GET /api/v1/admin/users?limit=50&offset=0
Authorization: Bearer ADMIN_TOKEN
```

---

#### 8. 获取用户信息

```http
GET /api/v1/admin/users/{user_id}
Authorization: Bearer ADMIN_TOKEN
```

---

#### 9. 修改用户角色

```http
PUT /api/v1/admin/users/{user_id}/role
Authorization: Bearer ADMIN_TOKEN
Content-Type: application/json

{
  "role": "admin"
}
```

**可用角色**:
- `user`: 普通用户
- `admin`: 管理员

---

#### 10. 禁用用户

```http
POST /api/v1/admin/users/{user_id}/disable
Authorization: Bearer ADMIN_TOKEN
```

---

#### 11. 启用用户

```http
POST /api/v1/admin/users/{user_id}/enable
Authorization: Bearer ADMIN_TOKEN
```

---

## 权限控制

### 用户角色

| 角色 | 权限 |
|------|------|
| user | 访问自己的资源 |
| admin | 访问所有资源，管理用户 |

### 资源访问规则

1. **账号管理**:
   - 用户只能访问自己创建的账号
   - 管理员可以访问所有账号

2. **账号池管理**:
   - 用户只能管理自己的账号池
   - 管理员可以管理所有账号池

3. **敏感操作**:
   - 获取解密后的 Cookie 需要资源所有者或管理员权限

---

## 安全最佳实践

### 1. 密码安全

- ✅ 密码使用 bcrypt 哈希存储
- ✅ 最小长度 8 个字符
- ✅ 不在响应中返回密码哈希

### 2. Token 安全

- ✅ JWT 使用 HS256 签名
- ✅ Token 有效期 24 小时
- ✅ Token 包含用户 ID、用户名、角色

### 3. API Key 安全

- ✅ API Key 使用 `pk_` 前缀
- ✅ 32 字节随机生成
- ✅ 支持重新生成

### 4. 传输安全

- ⚠️ **生产环境必须使用 HTTPS**
- ⚠️ 不要在 URL 中传递敏感信息
- ⚠️ 使用安全的 HTTP 头

---

## 使用示例

### Python 示例

```python
import requests

BASE_URL = "http://localhost:8080/api/v1"

# 1. 注册
response = requests.post(f"{BASE_URL}/register", json={
    "username": "testuser",
    "email": "test@example.com",
    "password": "secure_password_123"
})
print("注册:", response.json())

# 2. 登录
response = requests.post(f"{BASE_URL}/login", json={
    "username": "testuser",
    "password": "secure_password_123"
})
token = response.json()["data"]["token"]
print("Token:", token)

# 3. 使用 Token 访问 API
headers = {"Authorization": f"Bearer {token}"}
response = requests.get(f"{BASE_URL}/accounts", headers=headers)
print("账号列表:", response.json())

# 4. 使用 API Key
api_key = response.json()["data"]["user"]["api_key"]
headers = {"X-API-Key": api_key}
response = requests.get(f"{BASE_URL}/accounts", headers=headers)
print("账号列表 (API Key):", response.json())
```

### JavaScript 示例

```javascript
const axios = require('axios');

const BASE_URL = 'http://localhost:8080/api/v1';

// 1. 注册
async function register() {
  const response = await axios.post(`${BASE_URL}/register`, {
    username: 'testuser',
    email: 'test@example.com',
    password: 'secure_password_123'
  });
  console.log('注册:', response.data);
}

// 2. 登录
async function login() {
  const response = await axios.post(`${BASE_URL}/login`, {
    username: 'testuser',
    password: 'secure_password_123'
  });
  return response.data.data.token;
}

// 3. 使用 Token
async function useToken(token) {
  const response = await axios.get(`${BASE_URL}/accounts`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  console.log('账号列表:', response.data);
}

// 4. 使用 API Key
async function useAPIKey(apiKey) {
  const response = await axios.get(`${BASE_URL}/accounts`, {
    headers: { 'X-API-Key': apiKey }
  });
  console.log('账号列表:', response.data);
}

// 执行
(async () => {
  await register();
  const token = await login();
  await useToken(token);
})();
```

---

## 错误处理

### 认证错误

```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required"
  }
}
```

**HTTP 状态码**: 401

### 授权错误

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Insufficient permissions"
  }
}
```

**HTTP 状态码**: 403

### Token 过期

```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Token expired"
  }
}
```

**解决方案**: 重新登录获取新 Token

---

## 配置说明

### 环境变量

```bash
# JWT 密钥（生产环境必须设置）
JWT_SECRET=your-secret-key-here

# JWT 过期时间（小时）
JWT_EXPIRATION=24

# API Key 请求头名称
API_KEY_HEADER=X-API-Key
```

### 跳过认证的路径

以下路径不需要认证：
- `/api/v1/health` - 健康检查
- `/api/v1/login` - 登录
- `/api/v1/register` - 注册

---

## 常见问题

### Q: Token 过期了怎么办？

A: 重新调用 `/api/v1/login` 获取新 Token。

### Q: API Key 泄露了怎么办？

A: 调用 `/api/v1/auth/me/api-key` 重新生成新的 API Key。

### Q: 忘记密码怎么办？

A: 目前需要联系管理员重置密码。后续会添加密码找回功能。

### Q: 如何创建管理员账号？

A: 第一个注册的用户默认为普通用户，需要通过数据库手动修改角色为 `admin`，或使用管理员接口修改。

---

## 更新日志

### v1.1.0 (2026-02-21)

- ✅ 添加 JWT 认证
- ✅ 添加 API Key 认证
- ✅ 实现用户注册和登录
- ✅ 实现基于角色的权限控制
- ✅ 添加管理员功能
- ✅ 密码使用 bcrypt 哈希存储

---

## 下一步

1. 添加密码找回功能
2. 实现双因素认证（2FA）
3. 添加 OAuth2 集成
4. 实现会话管理
5. 添加审计日志

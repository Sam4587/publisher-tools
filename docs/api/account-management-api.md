# 账号管理系统 API 文档

## 概述

账号管理系统提供多平台账号的统一管理、认证、健康检查和负载均衡功能。支持抖音、今日头条、小红书、B站等主流内容平台。

## 核心功能

- **登录认证**: 支持二维码扫码登录，自动获取和存储 Cookie
- **Cookie 加密存储**: 使用 AES-256-GCM 加密算法保护敏感数据
- **健康检查**: 定期检查账号状态，自动标记失效账号
- **账号池管理**: 支持多账号池，提供多种负载均衡策略
- **使用统计**: 记录账号使用情况，提供成功率分析

## API 端点

### 1. 登录相关 API

#### 1.1 检查登录状态

检查指定平台的登录状态。

**请求**

```http
GET /api/v1/login/status?platform={platform}
```

**参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台名称：douyin, toutiao, xiaohongshu, bilibili |

**响应示例**

```json
{
  "success": true,
  "data": {
    "is_logged_in": true,
    "platform": "xiaohongshu",
    "account_id": "acc_123456",
    "account_name": "我的小红书账号",
    "last_check": "2026-02-21T16:00:00Z",
    "status": "healthy",
    "message": "Account is healthy and logged in"
  }
}
```

**状态码**

- `200 OK`: 请求成功
- `400 Bad Request`: 缺少 platform 参数
- `500 Internal Server Error`: 服务器内部错误

---

#### 1.2 获取登录二维码

获取指定平台的登录二维码，用户使用手机 APP 扫码登录。

**请求**

```http
GET /api/v1/login/qrcode?platform={platform}
```

**参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台名称 |

**响应示例**

```json
{
  "success": true,
  "data": {
    "platform": "xiaohongshu",
    "qrcode_url": "https://www.xiaohongshu.com/explore",
    "qrcode_image": "data:image/png;base64,iVBORw0KG...",
    "expires_in": 300,
    "session_id": "sess_abc123",
    "message": "Please scan the QR code with your mobile app to login"
  }
}
```

**字段说明**

| 字段 | 类型 | 说明 |
|------|------|------|
| qrcode_url | string | 登录页面 URL |
| qrcode_image | string | Base64 编码的二维码图片（可选） |
| expires_in | int | 二维码有效期（秒） |
| session_id | string | 会话 ID，用于后续轮询登录状态 |

**使用流程**

1. 调用此 API 获取二维码
2. 展示二维码给用户
3. 用户使用手机 APP 扫码
4. 轮询 `/api/v1/login/status` 检查登录状态
5. 登录成功后，系统自动保存 Cookie

---

#### 1.3 登录回调

登录成功后的回调接口，用于保存账号信息。

**请求**

```http
POST /api/v1/login/callback
Content-Type: application/json
```

**请求体**

```json
{
  "platform": "xiaohongshu",
  "session_id": "sess_abc123",
  "cookie_data": "a1=xxx; a2=yyy; ...",
  "account_name": "我的小红书账号",
  "user_id": "user_001"
}
```

**字段说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台名称 |
| session_id | string | 否 | 会话 ID |
| cookie_data | string | 是 | Cookie 数据 |
| account_name | string | 否 | 账号名称 |
| user_id | string | 否 | 所属用户 ID |

**响应示例**

```json
{
  "success": true,
  "data": {
    "account_id": "acc_789xyz",
    "platform": "xiaohongshu",
    "account_name": "我的小红书账号",
    "status": "pending",
    "message": "Account created and logged in successfully"
  }
}
```

---

### 2. 账号管理 API

#### 2.1 创建账号

手动创建新账号（通常用于导入已有 Cookie）。

**请求**

```http
POST /api/v1/accounts
Content-Type: application/json
```

**请求体**

```json
{
  "platform": "douyin",
  "account_name": "抖音主账号",
  "account_type": "personal",
  "cookie_data": "sessionid=xxx; ttwid=yyy; ...",
  "priority": 8,
  "user_id": "user_001",
  "project_id": "proj_001",
  "tags": {
    "type": "main",
    "region": "beijing"
  },
  "metadata": {
    "follower_count": 10000,
    "verified": true
  }
}
```

**字段说明**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台名称 |
| account_name | string | 否 | 账号名称 |
| account_type | string | 否 | 账号类型：personal, business |
| cookie_data | string | 是 | Cookie 数据（将被加密存储） |
| priority | int | 否 | 优先级（1-10，默认 5） |
| user_id | string | 否 | 所属用户 ID |
| project_id | string | 否 | 所属项目 ID |
| tags | object | 否 | 标签键值对 |
| metadata | object | 否 | 元数据 |

**响应示例**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "account_id": "acc_abc123",
    "platform": "douyin",
    "account_name": "抖音主账号",
    "status": "pending",
    "priority": 8,
    "created_at": "2026-02-21T16:00:00Z"
  }
}
```

---

#### 2.2 列出账号

查询账号列表，支持多条件筛选。

**请求**

```http
GET /api/v1/accounts?platform={platform}&status={status}&limit={limit}&offset={offset}
```

**参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 否 | 平台筛选 |
| status | string | 否 | 状态筛选：pending, active, inactive, expired |
| user_id | string | 否 | 用户 ID 筛选 |
| project_id | string | 否 | 项目 ID 筛选 |
| limit | int | 否 | 每页数量（默认 50） |
| offset | int | 否 | 偏移量（默认 0） |

**响应示例**

```json
{
  "success": true,
  "data": {
    "accounts": [
      {
        "account_id": "acc_001",
        "platform": "xiaohongshu",
        "account_name": "小红书主号",
        "status": "active",
        "priority": 8,
        "use_count": 150,
        "success_count": 145,
        "fail_count": 5,
        "last_used_at": "2026-02-21T15:30:00Z"
      }
    ],
    "total": 25,
    "limit": 50,
    "offset": 0
  }
}
```

---

#### 2.3 获取账号详情

获取单个账号的详细信息。

**请求**

```http
GET /api/v1/accounts/{id}
```

**响应示例**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "account_id": "acc_001",
    "platform": "xiaohongshu",
    "account_name": "小红书主号",
    "account_type": "personal",
    "status": "active",
    "priority": 8,
    "use_count": 150,
    "success_count": 145,
    "fail_count": 5,
    "last_used_at": "2026-02-21T15:30:00Z",
    "last_check_at": "2026-02-21T16:00:00Z",
    "created_at": "2026-02-01T10:00:00Z",
    "updated_at": "2026-02-21T16:00:00Z"
  }
}
```

---

#### 2.4 更新账号

更新账号信息。

**请求**

```http
PUT /api/v1/accounts/{id}
Content-Type: application/json
```

**请求体**

```json
{
  "account_name": "新账号名称",
  "priority": 9,
  "status": "active",
  "tags": {
    "type": "backup"
  }
}
```

---

#### 2.5 删除账号

删除指定账号。

**请求**

```http
DELETE /api/v1/accounts/{id}
```

---

#### 2.6 获取账号统计

获取账号的使用统计信息。

**请求**

```http
GET /api/v1/accounts/{id}/stats
```

**响应示例**

```json
{
  "success": true,
  "data": {
    "account_id": "acc_001",
    "platform": "xiaohongshu",
    "status": "active",
    "use_count": 150,
    "success_count": 145,
    "fail_count": 5,
    "success_rate": 96.67,
    "recent_logs": [
      {
        "action": "publish",
        "success": true,
        "duration_ms": 2340,
        "created_at": "2026-02-21T15:30:00Z"
      }
    ]
  }
}
```

---

#### 2.7 健康检查

对单个账号执行健康检查。

**请求**

```http
POST /api/v1/accounts/{id}/health-check
```

**响应示例**

```json
{
  "success": true,
  "data": {
    "account_id": "acc_001",
    "healthy": true,
    "message": "Account is healthy"
  }
}
```

---

#### 2.8 批量健康检查

对所有账号或指定平台的账号执行批量健康检查。

**请求**

```http
POST /api/v1/accounts/batch-health-check?platform={platform}
```

**响应示例**

```json
{
  "success": true,
  "data": {
    "success_count": 8,
    "fail_count": 2,
    "total": 10
  }
}
```

---

### 3. 账号池管理 API

#### 3.1 创建账号池

创建新的账号池。

**请求**

```http
POST /api/v1/pools
Content-Type: application/json
```

**请求体**

```json
{
  "name": "小红书发布池",
  "platform": "xiaohongshu",
  "description": "用于自动发布的小红书账号池",
  "strategy": "round_robin",
  "max_size": 10
}
```

**策略说明**

| 策略 | 说明 |
|------|------|
| round_robin | 轮询选择，均匀分配 |
| random | 随机选择 |
| priority | 按优先级选择 |
| least_used | 选择使用次数最少的账号 |

---

#### 3.2 列出账号池

查询账号池列表。

**请求**

```http
GET /api/v1/pools?platform={platform}
```

---

#### 3.3 添加账号到池

将账号添加到账号池。

**请求**

```http
POST /api/v1/pools/{pool_id}/members
Content-Type: application/json
```

**请求体**

```json
{
  "account_id": "acc_001",
  "priority": 8
}
```

---

#### 3.4 从池中移除账号

从账号池中移除账号。

**请求**

```http
DELETE /api/v1/pools/{pool_id}/members/{account_id}
```

---

#### 3.5 从池中选择账号

根据负载均衡策略从池中选择一个账号。

**请求**

```http
POST /api/v1/pools/{pool_id}/select
```

**响应示例**

```json
{
  "success": true,
  "data": {
    "account_id": "acc_001",
    "platform": "xiaohongshu",
    "account_name": "小红书主号",
    "status": "active"
  }
}
```

---

## 数据模型

### AccountStatus 账号状态

| 状态 | 说明 |
|------|------|
| pending | 待验证（新创建） |
| active | 活跃（健康） |
| inactive | 不活跃（失效） |
| expired | 已过期 |

### PlatformAccount 平台账号

```go
type PlatformAccount struct {
    ID           uint          // 主键
    AccountID    string        // 唯一账号ID
    Platform     string        // 平台名称
    AccountName  string        // 账号名称
    AccountType  string        // 账号类型
    CookieData   string        // 加密的Cookie数据
    CookieHash   string        // Cookie哈希
    Status       AccountStatus // 状态
    Priority     int           // 优先级（1-10）
    LastUsedAt   *time.Time    // 最后使用时间
    LastCheckAt  *time.Time    // 最后检查时间
    ExpiresAt    *time.Time    // 过期时间
    UseCount     int           // 使用次数
    SuccessCount int           // 成功次数
    FailCount    int           // 失败次数
    LastError    string        // 最后错误信息
    Tags         string        // 标签（JSON）
    Metadata     string        // 元数据（JSON）
    UserID       string        // 所属用户
    ProjectID    string        // 所属项目
    CreatedAt    time.Time     // 创建时间
    UpdatedAt    time.Time     // 更新时间
}
```

---

## 安全机制

### Cookie 加密

- 使用 AES-256-GCM 加密算法
- 密钥通过 SHA-256 派生
- 每次加密生成随机 nonce
- 加密数据以 Base64 编码存储

### 访问控制

- Cookie 数据不通过 API 直接返回
- 需要专门的接口获取解密后的 Cookie
- 建议在生产环境启用 API 认证

---

## 使用示例

### 完整登录流程

```bash
# 1. 获取登录二维码
curl "http://localhost:8080/api/v1/login/qrcode?platform=xiaohongshu"

# 2. 展示二维码给用户扫码

# 3. 轮询检查登录状态
curl "http://localhost:8080/api/v1/login/status?platform=xiaohongshu"

# 4. 登录成功后，系统自动保存账号
```

### 创建账号池并使用

```bash
# 1. 创建账号池
curl -X POST http://localhost:8080/api/v1/pools \
  -H "Content-Type: application/json" \
  -d '{
    "name": "小红书发布池",
    "platform": "xiaohongshu",
    "strategy": "round_robin"
  }'

# 2. 添加账号到池
curl -X POST http://localhost:8080/api/v1/pools/pool_001/members \
  -H "Content-Type: application/json" \
  -d '{"account_id": "acc_001", "priority": 8}'

# 3. 从池中选择账号
curl -X POST http://localhost:8080/api/v1/pools/pool_001/select
```

---

## 最佳实践

1. **定期健康检查**: 建议每天执行一次批量健康检查
2. **多账号备份**: 为每个平台配置多个账号，避免单点故障
3. **优先级设置**: 为重要账号设置更高优先级
4. **监控告警**: 监控账号失败率，及时发现问题
5. **Cookie 更新**: 定期更新 Cookie，避免过期

---

## 错误处理

所有错误响应遵循统一格式：

```json
{
  "success": false,
  "error": {
    "code": "ACCOUNT_NOT_FOUND",
    "message": "Account not found: acc_999",
    "details": null
  }
}
```

常见错误码：

| 错误码 | 说明 |
|--------|------|
| INVALID_REQUEST | 请求参数错误 |
| ACCOUNT_NOT_FOUND | 账号不存在 |
| CREATE_ACCOUNT_FAILED | 创建账号失败 |
| POOL_NOT_FOUND | 账号池不存在 |
| NO_AVAILABLE_ACCOUNT | 没有可用账号 |

---

## 更新日志

### v1.0.0 (2026-02-21)

- 初始版本发布
- 支持多平台账号管理
- 实现 Cookie 加密存储
- 支持账号池和负载均衡
- 提供健康检查功能

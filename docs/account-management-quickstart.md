# 账号管理系统快速开始指南

## 概述

账号管理系统已成功实现，提供了完整的多平台账号管理、登录认证、健康检查和负载均衡功能。本指南将帮助你快速上手使用。

## 已实现的功能

### ✅ 核心功能

1. **登录认证 API**
   - `GET /api/v1/login/status` - 检查登录状态
   - `GET /api/v1/login/qrcode` - 获取登录二维码
   - `POST /api/v1/login/callback` - 登录回调

2. **账号管理 API**
   - 创建、查询、更新、删除账号
   - 账号统计和使用记录
   - 健康检查（单个和批量）

3. **账号池管理**
   - 创建和管理账号池
   - 多种负载均衡策略
   - 智能账号选择

4. **安全机制**
   - AES-256-GCM Cookie 加密
   - Cookie 哈希验证
   - 安全的数据存储

## 快速开始

### 1. 检查登录状态

```bash
# 检查小红书平台的登录状态
curl "http://localhost:8080/api/v1/login/status?platform=xiaohongshu"
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "is_logged_in": true,
    "platform": "xiaohongshu",
    "account_id": "acc_123456",
    "status": "healthy"
  }
}
```

### 2. 获取登录二维码

```bash
# 获取小红书登录二维码
curl "http://localhost:8080/api/v1/login/qrcode?platform=xiaohongshu"
```

**响应示例：**
```json
{
  "success": true,
  "data": {
    "platform": "xiaohongshu",
    "qrcode_url": "https://www.xiaohongshu.com/explore",
    "expires_in": 300,
    "session_id": "sess_abc123",
    "message": "Please scan the QR code with your mobile app to login"
  }
}
```

### 3. 创建账号池

```bash
# 创建一个小红书账号池
curl -X POST http://localhost:8080/api/v1/pools \
  -H "Content-Type: application/json" \
  -d '{
    "name": "小红书发布池",
    "platform": "xiaohongshu",
    "strategy": "round_robin",
    "max_size": 10
  }'
```

### 4. 添加账号到池

```bash
# 将账号添加到账号池
curl -X POST http://localhost:8080/api/v1/pools/pool_001/members \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "acc_001",
    "priority": 8
  }'
```

### 5. 从池中选择账号

```bash
# 根据负载均衡策略选择账号
curl -X POST http://localhost:8080/api/v1/pools/pool_001/select
```

## 完整登录流程示例

### Python 示例

```python
import requests
import time
import json

BASE_URL = "http://localhost:8080/api/v1"

def login_platform(platform):
    """完整的登录流程"""
    
    # 1. 获取登录二维码
    print(f"正在获取 {platform} 登录二维码...")
    response = requests.get(f"{BASE_URL}/login/qrcode", params={"platform": platform})
    qrcode_data = response.json()["data"]
    
    print(f"二维码URL: {qrcode_data['qrcode_url']}")
    print(f"会话ID: {qrcode_data['session_id']}")
    print(f"有效期: {qrcode_data['expires_in']}秒")
    print("\n请使用手机APP扫描二维码登录...")
    
    # 2. 轮询检查登录状态
    session_id = qrcode_data["session_id"]
    max_attempts = 60  # 最多等待5分钟
    
    for i in range(max_attempts):
        time.sleep(5)
        
        response = requests.get(
            f"{BASE_URL}/login/status",
            params={"platform": platform}
        )
        status_data = response.json()["data"]
        
        if status_data["is_logged_in"]:
            print(f"\n登录成功！")
            print(f"账号ID: {status_data['account_id']}")
            print(f"账号名称: {status_data['account_name']}")
            return status_data["account_id"]
        
        print(f"等待登录... ({i+1}/{max_attempts})")
    
    print("\n登录超时")
    return None

# 使用示例
if __name__ == "__main__":
    account_id = login_platform("xiaohongshu")
    if account_id:
        print(f"账号 {account_id} 已成功登录并保存")
```

### JavaScript 示例

```javascript
const axios = require('axios');

const BASE_URL = 'http://localhost:8080/api/v1';

async function loginPlatform(platform) {
  try {
    // 1. 获取登录二维码
    console.log(`正在获取 ${platform} 登录二维码...`);
    const qrcodeResponse = await axios.get(`${BASE_URL}/login/qrcode`, {
      params: { platform }
    });
    
    const qrcodeData = qrcodeResponse.data.data;
    console.log(`二维码URL: ${qrcodeData.qrcode_url}`);
    console.log(`会话ID: ${qrcodeData.session_id}`);
    console.log('请使用手机APP扫描二维码登录...\n');
    
    // 2. 轮询检查登录状态
    const maxAttempts = 60;
    for (let i = 0; i < maxAttempts; i++) {
      await new Promise(resolve => setTimeout(resolve, 5000));
      
      const statusResponse = await axios.get(`${BASE_URL}/login/status`, {
        params: { platform }
      });
      
      const statusData = statusResponse.data.data;
      if (statusData.is_logged_in) {
        console.log('\n登录成功！');
        console.log(`账号ID: ${statusData.account_id}`);
        console.log(`账号名称: ${statusData.account_name}`);
        return statusData.account_id;
      }
      
      console.log(`等待登录... (${i + 1}/${maxAttempts})`);
    }
    
    console.log('\n登录超时');
    return null;
  } catch (error) {
    console.error('登录失败:', error.message);
    return null;
  }
}

// 使用示例
loginPlatform('xiaohongshu').then(accountId => {
  if (accountId) {
    console.log(`账号 ${accountId} 已成功登录并保存`);
  }
});
```

## 账号管理最佳实践

### 1. 定期健康检查

```bash
# 每天执行一次批量健康检查
curl -X POST "http://localhost:8080/api/v1/accounts/batch-health-check?platform=xiaohongshu"
```

### 2. 监控账号状态

```bash
# 查看账号统计信息
curl "http://localhost:8080/api/v1/accounts/acc_001/stats"
```

### 3. 使用账号池

```bash
# 创建账号池
curl -X POST http://localhost:8080/api/v1/pools \
  -H "Content-Type: application/json" \
  -d '{
    "name": "生产环境账号池",
    "platform": "xiaohongshu",
    "strategy": "least_used"
  }'

# 添加多个账号
for account_id in acc_001 acc_002 acc_003; do
  curl -X POST http://localhost:8080/api/v1/pools/pool_001/members \
    -H "Content-Type: application/json" \
    -d "{\"account_id\": \"$account_id\", \"priority\": 5}"
done

# 使用时自动选择最优账号
curl -X POST http://localhost:8080/api/v1/pools/pool_001/select
```

## 支持的平台

| 平台 | 平台代码 | 登录URL |
|------|---------|---------|
| 小红书 | xiaohongshu | https://www.xiaohongshu.com/explore |
| 抖音 | douyin | https://www.douyin.com/ |
| 今日头条 | toutiao | https://www.toutiao.com/ |
| B站 | bilibili | https://www.bilibili.com/ |

## 负载均衡策略

| 策略 | 代码 | 说明 |
|------|------|------|
| 轮询 | round_robin | 按顺序轮流使用账号 |
| 随机 | random | 随机选择账号 |
| 优先级 | priority | 优先使用高优先级账号 |
| 最少使用 | least_used | 选择使用次数最少的账号 |

## 安全建议

1. **定期更新 Cookie**: 建议每 30 天更新一次账号 Cookie
2. **多账号备份**: 每个平台至少配置 2-3 个账号
3. **监控失败率**: 当账号失败率超过 20% 时及时处理
4. **使用 HTTPS**: 生产环境务必使用 HTTPS
5. **限制 API 访问**: 添加认证机制保护 API

## 故障排查

### 问题：登录状态检查失败

**解决方案：**
1. 检查账号 Cookie 是否过期
2. 执行健康检查：`POST /api/v1/accounts/{id}/health-check`
3. 重新登录获取新 Cookie

### 问题：账号池选择失败

**解决方案：**
1. 检查账号池是否有活跃账号
2. 检查账号状态是否为 active
3. 添加更多账号到池中

## 下一步

1. 集成浏览器自动化服务，实现真正的二维码获取
2. 开发前端管理界面
3. 添加定时健康检查任务
4. 实现账号使用告警机制

## 相关文档

- [完整 API 文档](./api/account-management-api.md)
- [数据库模型设计](../publisher-core/database/models.go)
- [账号服务实现](../publisher-core/account/service.go)

## 技术支持

如有问题，请查看：
- API 文档：`docs/api/account-management-api.md`
- 源代码：`publisher-core/account/` 和 `publisher-core/api/account_handlers.go`
- 数据库模型：`publisher-core/database/models.go`

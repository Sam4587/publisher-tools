# 日志安全最佳实践

## 概述

本文档说明了如何安全地记录日志,避免泄露敏感信息。

## 敏感信息类型

以下类型的信息应该被视为敏感信息,不应记录在日志中:

1. **认证信息**
   - 密码、口令
   - JWT token
   - API密钥
   - 访问令牌
   - 刷新令牌

2. **个人信息**
   - 身份证号
   - 信用卡号
   - 社会安全号
   - 银行账户信息

3. **保密信息**
   - 商业机密
   - 内部系统信息
   - 配置密钥

## 日志安全工具

项目提供了 `logger/security.go` 包,包含以下安全日志函数:

### SafeError
```go
logger.SafeError(logrus.Fields{
    "user_id": userID,
    "action": "login",
}, "User login failed")
```

### SafeErrorf
```go
logger.SafeErrorf("Request failed: %s", err.Error())
```

### SafeWarn
```go
logger.SafeWarn(logrus.Fields{
    "user_id": userID,
    "path": r.URL.Path,
}, "Unauthorized access attempt")
```

### SanitizeLog
```go
data := map[string]interface{}{
    "username": "john",
    "password": "secret123",  // 会被自动清理
    "email": "john@example.com",
}
cleanData := logger.SanitizeLog(data)
// 结果: {"username": "john", "password": "***REDACTED***", "email": "john@example.com"}
```

### SanitizeString
```go
url := "https://api.example.com?token=abc123&user=foo"
cleanURL := logger.SanitizeString(url)
// 结果: "https://api.example.com?token=***REDACTED***&user=foo"
```

## 日志记录原则

### DO (应该做)

#### 1. 记录关键操作
```go
logger.SafeInfo(logrus.Fields{
    "user_id": user.UserID,
    "action": "create_task",
    "task_id": taskID,
}, "Task created successfully")
```

#### 2. 记录错误信息(不包含敏感数据)
```go
logger.SafeError(logrus.Fields{
    "user_id": user.UserID,
    "endpoint": r.URL.Path,
    "method": r.Method,
}, "Request processing failed")
```

#### 3. 使用请求ID追踪
```go
requestID := r.Header.Get("X-Request-ID")
logger.SafeInfo(logrus.Fields{
    "request_id": requestID,
    "user_id": user.UserID,
}, "Processing request")
```

#### 4. 记录性能指标
```go
logger.SafeInfo(logrus.Fields{
    "endpoint": r.URL.Path,
    "duration": time.Since(start).Milliseconds(),
    "status_code": statusCode,
}, "Request completed")
```

### DON'T (不应该做)

#### 1. 不要记录密码
```go
// ❌ 错误
logrus.Infof("User login: username=%s, password=%s", username, password)

// ✅ 正确
logger.SafeInfo(logrus.Fields{
    "username": username,
}, "User login attempt")
```

#### 2. 不要记录完整的token
```go
// ❌ 错误
logrus.Infof("Authorization: %s", r.Header.Get("Authorization"))

// ✅ 正确
authHeader := r.Header.Get("Authorization")
if strings.HasPrefix(authHeader, "Bearer ") {
    logger.SafeInfo(logrus.Fields{
        "auth_type": "Bearer",
    }, "Request authenticated")
}
```

#### 3. 不要记录完整的错误信息(可能包含敏感数据)
```go
// ❌ 错误
logrus.Errorf("Database error: %v", err)

// ✅ 正确
logger.SafeError(logrus.Fields{
    "query": query,  // 确保查询中不包含敏感数据
    "error_type": "database",
}, "Database query failed")
```

#### 4. 不要记录完整的请求参数
```go
// ❌ 错误
logrus.Infof("Request params: %v", r.URL.Query())

// ✅ 正确
logger.SafeInfo(logrus.Fields{
    "endpoint": r.URL.Path,
    "method": r.Method,
    "param_count": len(r.URL.Query()),
}, "Request received")
```

## 认证相关日志

### 登录失败
```go
// ✅ 正确
logger.SafeWarn(logrus.Fields{
    "username": username,
    "ip": r.RemoteAddr,
}, "Login failed")

// ❌ 错误
logrus.Warnf("Login failed for user %s with password %s", username, password)
```

### 认证失败
```go
// ✅ 正确
logger.SafeWarn(logrus.Fields{
    "endpoint": r.URL.Path,
    "method": r.Method,
    "ip": r.RemoteAddr,
}, "Authentication failed")

// ❌ 错误
logrus.Warnf("Authentication failed: %s", err.Error())
```

## API请求日志

### 记录请求信息
```go
// ✅ 正确
logger.SafeInfo(logrus.Fields{
    "method": r.Method,
    "path": r.URL.Path,
    "user_id": userID,
    "request_id": requestID,
}, "API request")

// ❌ 错误
logrus.Infof("API request: %s %s %s", r.Method, r.URL.Path, r.URL.Query())
```

### 记录响应信息
```go
// ✅ 正确
logger.SafeInfo(logrus.Fields{
    "request_id": requestID,
    "status_code": statusCode,
    "duration": duration,
}, "API response")

// ❌ 错误
logrus.Infof("API response: %s", string(responseBody))
```

## 错误处理日志

### 数据库错误
```go
// ✅ 正确
logger.SafeError(logrus.Fields{
    "operation": "create_user",
    "error_type": "database",
}, "Database operation failed")

// ❌ 错误
logrus.Errorf("Database error: %v", err.Error())
```

### API调用错误
```go
// ✅ 正确
logger.SafeError(logrus.Fields{
    "endpoint": endpoint,
    "method": method,
    "status_code": statusCode,
}, "External API call failed")

// ❌ 错误
logrus.Errorf("API call failed: %s %s %s", method, endpoint, string(responseBody))
```

## 日志级别使用指南

- **DEBUG**: 详细的调试信息,仅在开发环境使用
- **INFO**: 一般信息性消息,记录重要操作
- **WARN**: 警告信息,可能的问题但不影响运行
- **ERROR**: 错误信息,需要关注但不影响系统运行
- **FATAL**: 致命错误,导致系统无法运行

## 配置建议

### 开发环境
```env
LOG_LEVEL=debug
LOG_FILE_PATH=./logs/app.log
LOG_MAX_SIZE=100
LOG_MAX_BACKUPS=3
LOG_MAX_AGE=28
LOG_COMPRESS=true
```

### 生产环境
```env
LOG_LEVEL=warn
LOG_FILE_PATH=/var/log/app/app.log
LOG_MAX_SIZE=500
LOG_MAX_BACKUPS=10
LOG_MAX_AGE=90
LOG_COMPRESS=true
```

## 日志轮转

使用 `lumberjack` 实现日志轮转:

```go
import "gopkg.in/natefinch/lumberjack.v2"

logrus.SetOutput(&lumberjack.Logger{
    Filename:   "./logs/app.log",
    MaxSize:    100, // MB
    MaxBackups: 3,
    MaxAge:     28, // days
    Compress:   true,
})
```

## 审计日志

对于需要审计的操作,使用单独的审计日志:

```go
logger.SafeInfo(logrus.Fields{
    "user_id": user.UserID,
    "action": "delete_account",
    "target_id": targetID,
    "timestamp": time.Now().Unix(),
    "ip": r.RemoteAddr,
}, "AUDIT: Account deletion")
```

## 监控和告警

### 关键指标
- 错误率
- 响应时间
- 请求量
- 认证失败次数

### 告警规则
- 错误率超过5%
- 响应时间超过1秒
- 认证失败次数超过10次/分钟
- 5XX错误超过1%

## 合规性考虑

### GDPR
- 不记录个人身份信息(PII)
- 提供数据删除机制
- 实现数据访问控制

### PCI DSS
- 不记录完整的信用卡号
- 加密存储敏感数据
- 定期审计日志

## 总结

1. 始终使用安全日志函数
2. 不要记录敏感信息
3. 使用日志级别区分重要性
4. 实现日志轮转和归档
5. 定期审查日志内容
6. 保护日志文件访问权限

---

**文档版本**: 1.0  
**最后更新**: 2026-02-23  
**维护者**: 开发团队

# 配置参数说明

## 概述

本文档详细说明Publisher Tools的各项配置参数，包括环境变量、配置文件和系统设置。

## 目录

- [环境变量配置](#环境变量配置)
- [配置文件说明](#配置文件说明)
- [系统参数详解](#系统参数详解)
- [平台特定配置](#平台特定配置)
- [安全配置](#安全配置)

## 环境变量配置

### 核心服务配置

| 变量名 | 默认值 | 说明 | 必需 |
|--------|--------|------|------|
| `PORT` | `8080` | API服务监听端口 | 否 |
| `HOST` | `localhost` | 服务绑定主机地址 | 否 |
| `DEBUG` | `false` | 调试模式开关 | 否 |
| `LOG_LEVEL` | `info` | 日志级别（debug/info/warn/error） | 否 |
| `HEADLESS` | `true` | 浏览器无头模式 | 否 |

### 存储配置

| 变量名 | 默认值 | 说明 | 必需 |
|--------|--------|------|------|
| `STORAGE_DIR` | `./uploads` | 文件存储目录 | 否 |
| `COOKIE_DIR` | `./cookies` | Cookie存储目录 | 否 |
| `DATA_DIR` | `./data` | 数据存储目录 | 否 |
| `TEMP_DIR` | `./temp` | 临时文件目录 | 否 |

### AI服务配置

| 变量名 | 默认值 | 说明 | 必需 |
|--------|--------|------|------|
| `OPENROUTER_API_KEY` | `` | OpenRouter API密钥 | 否 |
| `DEEPSEEK_API_KEY` | `` | DeepSeek API密钥 | 否 |
| `GOOGLE_API_KEY` | `` | Google AI API密钥 | 否 |
| `GROQ_API_KEY` | `` | Groq API密钥 | 否 |
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama服务地址 | 否 |
| `DEFAULT_AI_PROVIDER` | `openrouter` | 默认AI提供商 | 否 |

### 数据库配置（预留）

| 变量名 | 默认值 | 说明 | 必需 |
|--------|--------|------|------|
| `DATABASE_URL` | `` | 数据库连接字符串 | 否 |
| `REDIS_URL` | `` | Redis连接地址 | 否 |

## 配置文件说明

### 主配置文件 (config.yaml)

```yaml
# 服务配置
server:
  port: 8080
  host: "0.0.0.0"
  debug: false
  log_level: "info"

# 存储配置
storage:
  upload_dir: "./uploads"
  cookie_dir: "./cookies"
  data_dir: "./data"
  temp_dir: "./temp"

# 浏览器配置
browser:
  headless: true
  timeout: 30
  max_instances: 5
  user_agent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

# 任务配置
task:
  max_concurrent: 10
  queue_size: 100
  timeout: 1800  # 30分钟
  retry_count: 3

# AI配置
ai:
  default_provider: "openrouter"
  timeout: 120
  max_tokens: 2000
  temperature: 0.7

# 平台配置
platforms:
  douyin:
    enabled: true
    rate_limit: 60  # 每小时请求数
    max_daily_posts: 50
    
  toutiao:
    enabled: true
    rate_limit: 120
    max_daily_posts: 100
    
  xiaohongshu:
    enabled: true
    rate_limit: 30
    max_daily_posts: 30
```

### 平台特定配置

#### 抖音配置 (douyin.yaml)
```yaml
douyin:
  # 登录配置
  login:
    timeout: 300
    retry_count: 3
    headless: false  # 登录时显示浏览器
  
  # 发布配置
  publish:
    title_max_length: 30
    content_max_length: 2000
    max_images: 12
    max_video_size: "4GB"
    min_interval: 300  # 5分钟最小间隔
  
  # 浏览器配置
  browser:
    viewport:
      width: 1920
      height: 1080
    wait_timeout: 30
```

#### 今日头条配置 (toutiao.yaml)
```yaml
toutiao:
  # 登录配置
  login:
    timeout: 300
    methods: ["phone", "wechat", "qq"]
  
  # 发布配置
  publish:
    title_max_length: 30
    content_max_length: 2000
    content_types: ["images", "video", "article"]
    max_tags: 10
  
  # 内容审核
  moderation:
    auto_check: true
    sensitive_words: ["违禁词1", "违禁词2"]
```

## 系统参数详解

### 性能调优参数

#### 并发控制
```yaml
# 任务并发配置
concurrency:
  max_workers: 10          # 最大工作协程数
  task_queue_size: 1000    # 任务队列大小
  browser_pool_size: 5     # 浏览器实例池大小
  api_rate_limit: 100      # API请求速率限制（每分钟）
```

#### 超时设置
```yaml
# 各种超时配置（秒）
timeouts:
  browser_navigation: 30    # 页面导航超时
  element_wait: 10          # 元素等待超时
  api_request: 30          # API请求超时
  task_execution: 1800     # 任务执行超时（30分钟）
  login_process: 300       # 登录流程超时
```

#### 内存管理
```yaml
# 内存相关配置
memory:
  max_cache_size: "100MB"   # 最大缓存大小
  cleanup_interval: 3600    # 清理间隔（秒）
  browser_memory_limit: "2GB" # 单个浏览器实例内存限制
```

### 安全参数

#### 访问控制
```yaml
security:
  # API访问控制
  api:
    rate_limit:
      requests_per_minute: 100
      burst_limit: 20
    cors_origins: ["http://localhost:5173", "https://yourdomain.com"]
    require_auth: true
  
  # 文件安全
  file:
    allowed_extensions: [".jpg", ".png", ".mp4", ".mov"]
    max_upload_size: "100MB"
    scan_virus: true  # 是否扫描病毒（需要集成杀毒软件）
  
  # Cookie安全
  cookie:
    encrypt_storage: true
    auto_refresh: true
    refresh_threshold: 86400  # 24小时前过期时刷新
```

#### 日志配置
```yaml
logging:
  level: "info"
  format: "json"  # json/text
  output: "file"  # file/stdout/both
  file:
    path: "./logs/app.log"
    max_size: "100MB"
    max_age: 30
    compress: true
  sensitive_data_masking: true  # 是否遮蔽敏感信息
```

## 平台特定配置

### 抖音平台参数
```yaml
douyin_specific:
  # 内容规范
  content_limits:
    title_min_length: 5
    title_max_length: 30
    content_max_length: 2000
    hashtag_max_count: 5
    link_max_count: 3
  
  # 发布时机
  optimal_times:
    - "08:00"  # 早上活跃时段
    - "12:00"  # 午休时段
    - "18:00"  # 下班高峰
    - "21:00"  # 晚间娱乐时段
  
  # 内容审核规则
  moderation_rules:
    - pattern: "\\b(敏感词)\\b"
      action: "reject"
      message: "包含敏感词汇"
    - pattern: "^(.{0,4})$"
      action: "warn"
      message: "标题过短"
```

### 今日头条参数
```yaml
toutiao_specific:
  # 分类配置
  categories:
    - name: "科技"
      weight: 0.8
      recommended_tags: ["AI", "技术", "创新"]
    - name: "生活"
      weight: 0.6
      recommended_tags: ["日常", "分享", "生活技巧"]
  
  # 推荐算法参数
  recommendation_boost:
    fresh_content_bonus: 1.2    # 新内容加权
    engagement_bonus: 1.5       # 高互动加权
    time_decay_factor: 0.95     # 时间衰减因子
```

### 小红书参数
```yaml
xiaohongshu_specific:
  # 内容质量评分
  quality_scoring:
    image_quality_weight: 0.4   # 图片质量权重
    content_relevance_weight: 0.3 # 内容相关性权重
    engagement_potential_weight: 0.3 # 互动潜力权重
  
  # 社区规范
  community_guidelines:
    forbidden_topics: ["虚假宣传", "恶意营销"]
    required_elements: ["真实体验", "实用价值"]
    hashtag_requirements: 
      minimum: 3
      maximum: 10
      must_include: ["#笔记", "#分享"]
```

## 安全配置

### 加密配置
```yaml
encryption:
  # Cookie加密
  cookie:
    algorithm: "AES-256-GCM"
    key_rotation_interval: "30d"  # 30天轮换一次密钥
  
  # 敏感数据加密
  sensitive_data:
    enabled: true
    fields: ["api_keys", "passwords", "tokens"]
    algorithm: "RSA-4096"
  
  # TLS配置
  tls:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    min_version: "TLS1.2"
```

### 访问审计
```yaml
audit:
  # 操作日志
  operation_logging:
    enabled: true
    level: "info"
    retention_days: 90
  
  # 安全事件监控
  security_monitoring:
    failed_login_threshold: 5
    suspicious_activity_detection: true
    alert_channels: ["email", "slack"]
  
  # 数据访问控制
  data_access_control:
    role_based_access: true
    field_level_security: true
    audit_trail: true
```

## 配置最佳实践

### 1. 环境分离
```bash
# 开发环境
export DEBUG=true
export LOG_LEVEL=debug
export HEADLESS=false

# 生产环境
export DEBUG=false
export LOG_LEVEL=warn
export HEADLESS=true
```

### 2. 敏感信息管理
```bash
# 使用.env文件管理敏感配置
cat > .env << EOF
OPENROUTER_API_KEY=sk-xxxxxxxxxxxxxxxx
DEEPSEEK_API_KEY=sk-xxxxxxxxxxxxxxxx
DATABASE_URL=postgresql://user:pass@host:5432/db
EOF

# 在代码中安全加载
source .env
```

### 3. 配置验证
```bash
# 启动前验证配置
make validate-config

# 或使用配置检查工具
go run tools/config-validator/main.go --config config.yaml
```

## 故障排除

### 配置相关问题

#### 1. 配置文件加载失败
```bash
# 检查配置文件语法
yamllint config.yaml

# 验证必需参数
go run tools/config-checker/main.go --check-required
```

#### 2. 环境变量覆盖问题
```bash
# 查看当前生效的配置
curl http://localhost:8080/api/v1/config/dump

# 检查环境变量
printenv | grep PUBLISHER_
```

#### 3. 权限问题
```bash
# 检查目录权限
ls -la ./cookies/
ls -la ./uploads/

# 设置正确权限
chmod 755 ./cookies/
chmod 755 ./uploads/
chown appuser:appgroup ./cookies/
```

## 相关文档

- [部署指南](../../guides/deployment/)
- [安全配置指南](../../guides/security/)
- [API参考](../../api/rest-api.md)

## 维护信息

- 最后更新：2026-02-19
- 维护者：MonkeyCode Team
- 版本：v1.0
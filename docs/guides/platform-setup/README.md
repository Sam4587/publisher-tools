# 平台配置指南

## 概述

本文档详细介绍如何配置和管理抖音、今日头条、小红书等平台的账号连接和发布设置。

## 目录

- [准备工作](#准备工作)
- [抖音平台配置](#抖音平台配置)
- [今日头条配置](#今日头条配置)
- [小红书配置](#小红书配置)
- [批量配置](#批量配置)
- [故障排除](#故障排除)

## 准备工作

### 环境要求
- 确保已安装Chrome/Chromium浏览器
- 网络环境稳定，能够正常访问目标平台
- 准备好各平台的有效账号

### 目录结构
```
./cookies/                 # Cookie存储目录
├── douyin_cookies.json    # 抖音Cookie
├── toutiao_cookies.json   # 今日头条Cookie
└── xiaohongshu_cookies.json # 小红书Cookie

./uploads/                 # 文件上传目录
├── images/               # 图片文件
├── videos/               # 视频文件
└── documents/            # 文档文件
```

## 抖音平台配置

### 1. 账号准备
- 确保账号已完成实名认证
- 账号状态正常，无违规记录
- 准备手机号用于扫码登录

### 2. 登录配置
```bash
# 通过API登录
curl -X POST http://localhost:8080/api/v1/platforms/douyin/login

# 响应示例
{
  "success": true,
  "data": {
    "qrcode_url": "https://qr.douyin.com/xxx",
    "login_timeout": 300
  }
}
```

### 3. 扫码登录
1. 打开返回的二维码URL
2. 使用抖音APP扫描二维码
3. 在手机上确认登录授权
4. 等待系统自动保存Cookie

### 4. 验证登录状态
```bash
curl http://localhost:8080/api/v1/platforms/douyin/check

# 成功响应
{
  "success": true,
  "data": {
    "logged_in": true,
    "username": "你的昵称",
    "expires_at": "2026-03-19T10:00:00Z"
  }
}
```

### 5. 发布限制说明
- **标题长度**: 最多30字
- **正文长度**: 最多2000字
- **图片数量**: 最多12张
- **视频大小**: 最大4GB
- **发布间隔**: 建议≥5分钟

### 6. 注意事项
- 首次发布前建议先手动发布测试
- 避免频繁操作触发风控
- 定期检查账号状态和Cookie有效期

## 今日头条配置

### 1. 账号准备
- 需要头条号或西瓜视频账号
- 账号等级建议≥Lv2
- 确保有内容发布权限

### 2. 登录配置
```bash
# 登录今日头条
curl -X POST http://localhost:8080/api/v1/platforms/toutiao/login
```

### 3. 登录方式
支持两种登录方式：
1. **手机号验证码登录**
2. **第三方账号登录**（微信/QQ）

### 4. 内容类型支持
- **图文**: 支持多图+文字
- **视频**: 支持横竖屏视频
- **文章**: 支持富文本编辑

### 5. 发布限制
- **标题长度**: 最多30字
- **正文长度**: 最多2000字
- **图片数量**: 无明确上限
- **视频格式**: MP4格式推荐

## 小红书配置

### 1. 账号准备
- 确保账号已完成专业认证
- 账号粉丝数建议≥100
- 内容垂直度较高有利于推荐

### 2. 登录配置
```bash
# 登录小红书
curl -X POST http://localhost:8080/api/v1/platforms/xiaohongshu/login
```

### 3. 登录特点
- 小红书登录相对复杂
- 可能需要多次验证
- 建议在非高峰时段登录

### 4. 内容规范
- **标题长度**: 最多20字
- **正文长度**: 最多1000字
- **图片数量**: 最多18张
- **视频大小**: 最大500MB

### 5. 内容建议
- 图片质量要求较高
- 标题要有吸引力
- 正文要有实用价值
- 适当添加话题标签

## 批量配置

### 1. 批量登录脚本
```bash
#!/bin/bash
# batch_login.sh

PLATFORMS=("douyin" "toutiao" "xiaohongshu")

for platform in "${PLATFORMS[@]}"; do
    echo "正在登录 $platform..."
    response=$(curl -s -X POST "http://localhost:8080/api/v1/platforms/$platform/login")
    
    if echo "$response" | grep -q "qrcode_url"; then
        qrcode_url=$(echo "$response" | jq -r '.data.qrcode_url')
        echo "请扫描二维码登录: $qrcode_url"
        
        # 等待登录完成
        sleep 30
        
        # 检查登录状态
        status=$(curl -s "http://localhost:8080/api/v1/platforms/$platform/check")
        if echo "$status" | grep -q '"logged_in":true'; then
            echo "$platform 登录成功"
        else
            echo "$platform 登录失败"
        fi
    fi
done
```

### 2. 配置状态检查
```bash
# 检查所有平台登录状态
curl "http://localhost:8080/api/v1/platforms" | jq '.data[] | {name, logged_in, username}'
```

### 3. 自动化配置
```yaml
# config/platforms.yaml
platforms:
  douyin:
    enabled: true
    auto_login: true
    check_interval: 3600  # 1小时检查一次
    retry_attempts: 3
  
  toutiao:
    enabled: true
    auto_login: true
    check_interval: 7200  # 2小时检查一次
    
  xiaohongshu:
    enabled: false  # 暂时禁用
    reason: "需要人工审核"
```

## 故障排除

### 常见问题

#### 1. 登录失败
**现象**: 二维码生成后无法登录
**解决方案**:
```bash
# 检查浏览器是否正常启动
ps aux | grep chrome

# 清理Cookie重新登录
rm ./cookies/douyin_cookies.json
curl -X POST http://localhost:8080/api/v1/platforms/douyin/login
```

#### 2. Cookie过期
**现象**: 显示已登录但发布失败
**解决方案**:
```bash
# 手动刷新Cookie
curl -X POST http://localhost:8080/api/v1/platforms/douyin/login

# 设置自动刷新
# 在配置文件中启用自动刷新功能
```

#### 3. 网络连接问题
**现象**: 无法访问平台网站
**解决方案**:
```bash
# 检查网络连通性
ping www.douyin.com
curl -I https://www.douyin.com

# 如果有网络限制，配置代理
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080
```

#### 4. 浏览器兼容性
**现象**: 自动化操作异常
**解决方案**:
```bash
# 检查Chrome版本
google-chrome --version

# 更新Chrome浏览器
sudo apt update && sudo apt upgrade google-chrome-stable

# 或使用Chromium
sudo apt install chromium-browser
```

### 日志分析
```bash
# 查看详细日志
tail -f ./logs/server.log | grep -i "douyin\|login"

# 关键日志关键词
# "login success" - 登录成功
# "cookie expired" - Cookie过期
# "element not found" - 元素定位失败
# "navigation timeout" - 页面加载超时
```

### 联系支持
如遇到无法解决的问题，请提供以下信息：
1. 完整的错误日志
2. 操作步骤重现过程
3. 系统环境信息
4. 相关配置文件

## 相关文档

- [API接口文档](../../api/rest-api.md)
- [Cookie管理文档](../../modules/cookies/)
- [浏览器自动化文档](../../modules/browser/)

## 维护信息

- 最后更新：2026-02-19
- 维护者：MonkeyCode Team
- 版本：v1.0
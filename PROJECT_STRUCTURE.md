# 项目结构

## 核心目录

```
publisher-tools/
├── hotspot-server/      # Go热点数据服务
│   └── main.go          # 服务入口
├── server/              # Node.js后端服务
│   └── simple-server.js # 服务入口
├── publisher-web/       # React前端
│   ├── src/            # 源代码
│   ├── package.json    # 依赖配置
│   └── vite.config.ts  # Vite配置
├── logs/               # 日志目录(自动创建)
└── .archive/           # 归档的废弃文件
```

## 启动脚本

```
publisher-tools/
├── start-all.bat       # ✅ 启动所有服务(推荐)
├── stop.bat            # ✅ 停止所有服务
├── health-check.bat    # ✅ 健康检查工具
└── run_hidden.js       # 隐藏窗口运行工具
```

## 文档

```
publisher-tools/
├── README.md           # 快速入门指南
├── OPERATIONS.md       # 运维文档
└── PROJECT_STRUCTURE.md # 本文件
```

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Hotspot Server | 8080 | Go热点数据服务 |
| Node Backend | 3001 | Node.js业务后端 |
| Frontend | 5173 | React前端(可能使用5174/5175) |

## 日志文件

所有日志存储在 `logs/` 目录:
- `hotspot-server.log` - Go服务日志
- `node-backend.log` - Node后端日志
- `frontend.log` - 前端开发服务器日志

## 归档文件

`.archive/` 目录包含已废弃的文件:
- `start.bat` - 旧的启动脚本
- `start-stable.bat` - 服务管理器启动脚本
- `service-manager.js` - 服务管理器
- `test-server.js` - 测试服务器

这些文件仅作备份,不再维护。

## 开发指南

### 添加新服务

1. 在项目根目录创建服务目录
2. 更新 `start-all.bat` 添加启动逻辑
3. 更新 `stop.bat` 添加停止逻辑
4. 更新 `health-check.bat` 添加健康检查
5. 更新本文档

### 修改端口

1. 修改服务配置
2. 更新 `start-all.bat` 中的端口检查
3. 更新 `stop.bat` 中的端口清理
4. 更新 `vite.config.ts` 中的代理配置
5. 更新本文档

## 维护说明

- 定期清理 `logs/` 目录中的旧日志
- 不要删除 `.archive/` 目录,保留历史记录
- 修改启动逻辑时,同步更新文档

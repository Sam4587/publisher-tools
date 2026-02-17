# 重要说明：本项目不使用 MongoDB

## 存储方案

本项目采用**轻量级存储方案**，不依赖 MongoDB 或其他重量级数据库。

### 当前使用的存储方式

1. **JSON 文件存储**
   - 热点数据：`data/hot-topics.json`
   - 分析数据：`data/analytics/`
   - 用户数据：内存 + JSON 文件

2. **文件系统存储**
   - 上传文件：`uploads/`
   - Cookie：`cookies/`
   - 日志：`logs/`

3. **Go 后端存储**
   - 任务管理：内存存储（`publisher-core/task`）
   - 文件存储：本地文件系统（`publisher-core/storage`）

## 为什么不用 MongoDB？

### 优势

1. **零配置** - 无需安装和配置数据库
2. **轻量级** - 适合中小型项目
3. **易于迁移** - JSON 文件可以直接复制
4. **开发便捷** - 无需管理数据库连接
5. **部署简单** - 减少依赖项

### 适用场景

- ✅ 单机部署
- ✅ 开发和测试环境
- ✅ 数据量不大的应用
- ✅ 快速原型开发

## 如果需要升级到 MongoDB？

如果未来数据量增大，需要使用 MongoDB，可以：

1. **安装 MongoDB 依赖**
   ```bash
   cd server
   npm install mongoose
   ```

2. **恢复模型文件**
   - 从 Git 历史恢复原始的 `server/models/Task.js`

3. **配置环境变量**
   ```env
   MONGODB_URI=mongodb://localhost:27017/publisher-tools
   ```

4. **更新存储实现**
   - 修改相关服务使用 MongoDB

## 当前数据存储位置

```
data/
├── hot-topics.json       # 热点数据
├── analytics/            # 分析数据
│   ├── daily/
│   └── weekly/
├── cache/                # 缓存数据
└── temp/                 # 临时文件
```

## 数据备份

由于使用 JSON 文件存储，备份非常简单：

```bash
# 备份所有数据
cp -r data/ backup/data_$(date +%Y%m%d)/

# 备份上传文件
cp -r uploads/ backup/uploads_$(date +%Y%m%d)/
```

## 总结

- ❌ **不使用** MongoDB
- ❌ **不使用** Redis
- ✅ **使用** JSON 文件存储
- ✅ **使用** 本地文件系统
- ✅ **轻量级**、**易部署**

---

**最后更新**：2026-02-17

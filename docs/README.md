# Publisher Tools 文档中心

欢迎来到 Publisher Tools 文档中心！这里是项目所有文档的统一入口和导航中心。

## 🎯 快速导航

### 新用户和 AI 助手
**👉 [项目架构统一实施方案](./project-architecture-unified-implementation.md)** - 统一规划中心（推荐首先阅读）

本文档整合了当前项目架构分析和四个借鉴项目，提供了完整的实施方案和开发路线图。

### 核心开发文档
- **[AI 服务开发指南](./ai-service-development-guide.md)** - 统一的 AI 服务开发权威文档
  - 20+ 免费 AI 资源汇总
  - LiteLLM 统一接口方案
  - 开发进度记录（AI 助手协作）
- **[热点监控借鉴文档](./hot-topics-reference.md)** - TrendRadar 项目分析
- **[热点监控路线图](./hot-topics-roadmap.md)** - 功能增强开发计划
- **[Huobao Drama 项目分析](./huobao-drama-analysis.md)** - AI 短剧生成平台深度分析
- **[智能启动系统实施报告](./SMART_LAUNCHER_IMPLEMENTATION_REPORT.md)** - 智能启动系统实施报告

---

## 📚 文档分类导航

### [📖 系统架构](./architecture/)
了解系统的整体设计、技术选型和模块关系。

- [架构概览](./architecture/README.md) - 系统整体架构设计
- [模块详细设计](./architecture/module-design/) - 各核心模块技术实现
- [数据流图](./architecture/data-flow/) - 数据流转和处理逻辑
- [部署架构](./architecture/deployment/) - 部署方案和环境配置

### [👨‍💻 开发者指南](./development/developer-guide.md)
为开发者提供的完整开发环境搭建和编码指导。

- [环境搭建](./development/development-plan.md#环境搭建) - 开发环境配置
- [编码规范](./development/development-plan.md#编码规范) - 代码质量和风格要求
- [开发流程](./development/development-plan.md#开发流程) - 标准开发工作流
- [调试技巧](./development/development-plan.md#调试技巧) - 问题排查和调试方法
- [测试指南](./development/development-plan.md#测试指南) - 单元测试和集成测试
- [提交规范](./development/development-plan.md#提交规范) - Git提交和分支管理

### [🧩 功能模块](./modules/)
各功能模块的详细技术文档和使用说明。

- [浏览器自动化](./modules/browser/) - 基于Rod框架的网页操作
- [Cookie管理](./modules/cookies/) - 登录凭证管理和安全存储
- [平台适配器](./modules/adapters/) - 多平台发布接口统一
- [任务管理](./modules/task/) - 异步任务处理和状态追踪
- [热点监控](./modules/hotspot/) - 实时热点抓取和分析
- [AI服务](./modules/ai/) - 多提供商AI集成和服务管理
  - **[AI服务开发指南](./ai-service-development-guide.md)** - 统一的AI服务开发权威文档
  - 免费AI资源汇总（20+提供商）
  - LiteLLM统一接口方案
  - 开发进度记录（AI助手协作）
- [文件存储](./modules/storage/) - 统一文件存储抽象层
- [数据分析](./modules/analytics/) - 数据采集、分析和报告

### [🔌 API接口](./api/)
完整的REST API接口文档和使用指南。

- [REST API参考](./api/rest-api.md) - 所有HTTP接口详细说明
- [WebSocket接口](./api/websocket.md) - 实时通信接口（规划中）
- [SDK使用指南](./api/sdk/) - 客户端SDK集成说明（规划中）

### [📋 操作指南](./guides/)
面向用户的操作手册和最佳实践。

- [平台配置指南](./guides/platform-setup/) - 各平台账号配置和管理
- [内容发布指南](./guides/content-publish/) - 内容创作和发布流程
- [数据分析指南](./guides/analytics/) - 数据查看和报告生成
- [系统维护指南](./guides/maintenance/) - 日常运维和故障处理

### [📚 参考资料](./reference/)
技术参考资料和配置说明文档。

- [配置参数说明](./reference/configuration/) - 环境变量和配置文件详解
- [版本变更记录](./reference/changelog/) - 版本更新和功能变更（规划中）
- [常见问题解答](./reference/faq/) - FAQ和问题解决方案（规划中）
- [热点监控借鉴](./hot-topics-reference.md) - TrendRadar项目分析和整合建议
- [热点监控路线图](./hot-topics-roadmap.md) - 功能增强开发计划
- [Huobao Drama分析](./huobao-drama-analysis.md) - AI短剧生成平台深度分析
- **[AI服务开发指南](./ai-service-development-guide.md)** - 统一的AI服务开发权威文档
- **[项目架构统一实施方案](./project-architecture-unified-implementation.md)** - 统一规划中心（推荐）
- [项目架构优化文档](./project-architecture-optimization.md) - 项目架构优化规划

### [🤖 AI任务管理](./ai-tasks/)
AI开发者专项任务跟踪和管理系统。

- [AI任务工作单](./ai-tasks/ai-task-checklist.md) - 可追溯的任务管理清单
- [任务状态跟踪](./ai-tasks/ai-task-checklist.md#任务跟踪矩阵) - 实时任务进度监控
- [质量保证机制](./ai-tasks/ai-task-checklist.md#质量保证机制) - 任务验收标准

- [废弃路由文档](./archive/deprecated-routes/) - 旧版Node.js路由实现
- [历史版本文档](./archive/v1.0/) - 版本1.0相关文档（如有）

## 🚀 快速开始

如果你是新用户，建议按以下顺序阅读：

1. **[快速入门](./getting-started.md)** - 项目概览和基本使用
2. **[平台配置指南](./guides/platform-setup/)** - 配置第一个平台账号
3. **[内容发布指南](./guides/content-publish/)** - 学习基本发布操作
4. **[开发环境搭建](./development/development-plan.md#环境搭建)** - 如需二次开发

如果你是开发者：

1. **[开发计划](./development/development-plan.md)** - 完整的开发环境搭建
2. **[架构文档](./architecture/)** - 理解系统设计原理
3. **[API文档](./api/)** - 接口集成和扩展开发
4. **[模块文档](./modules/)** - 深入了解具体功能实现

## 📊 项目状态

- **版本**: v1.0.0
- **状态**: ✅ 稳定可用
- **最近更新**: 2026-02-19
- **支持平台**: 抖音、今日头条、小红书

## 🔧 技术栈

- **后端**: Go 1.21+, Rod框架, Gorilla Mux
- **前端**: React 18, TypeScript, Vite
- **AI集成**: OpenRouter, DeepSeek, Google AI
- **部署**: Docker, Docker Compose

## 📞 获取帮助

- **GitHub Issues**: [提交问题和建议](https://github.com/monkeycode/publisher-tools/issues)
- **文档更新**: 发现文档问题请提交PR
- **技术支持**: team@monkeycode.cn

## 📝 贡献指南

我们欢迎任何形式的贡献：

1. **文档改进**: 发现错误或不够清晰的地方
2. **功能建议**: 新特性和改进想法
3. **Bug报告**: 使用过程中发现的问题
4. **代码贡献**: 提交Pull Request

## 📅 更新日志

文档会随着项目版本同步更新，主要更新记录：

- **2026-02-19**: 完成文档系统重构和结构优化
- **2026-02-17**: 添加一键发布功能文档
- **2026-01-15**: 完成多平台适配器文档
- **2025-12-01**: 项目初期文档创建

---

**最后更新**: 2026-02-19  
**维护团队**: MonkeyCode Team  
**许可证**: MIT
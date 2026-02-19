const express = require('express');
const cors = require('cors');
const dotenv = require('dotenv');
const cron = require('node-cron');
const helmet = require('helmet');
const compression = require('compression');
const path = require('path');

dotenv.config();

const app = express();
const PORT = process.env.PORT || 3001;

// 静态文件服务
app.use('/videos', express.static(path.join(__dirname, '../public/videos')));
app.use('/audio', express.static(path.join(__dirname, '../public/audio')));
app.use('/previews', express.static(path.join(__dirname, '../public/previews')));

// 导入中间件
// const logger = require('./utils/logger'); // 暂时注释掉
// const { requestLogger, errorLogger } = require('./utils/enhancedLogger'); // 暂时注释掉
// const rateLimiter = require('./utils/rateLimiter'); // 暂时注释掉

// 服务器重启标记

// 导入新模块
const { fetcherManager } = require('./fetchers');
const { liteLLMAdapter } = require('./ai');
const { notificationDispatcher } = require('./notification');
const { reportGenerator } = require('./reports');
const { storageManager, topicAnalyzer, trendAnalyzer } = require('./core');

// 内存存储 - 用于轻量级存储方案
const memoryStorage = {
  hotTopics: [],
  lastUpdate: null
};

// 安全中间件
app.use(helmet({
  contentSecurityPolicy: false,
  crossOriginEmbedderPolicy: false,
}));

// 压缩中间件
app.use(compression());

// CORS配置
app.use(cors({
  origin: process.env.CORS_ORIGIN || '*',
  credentials: true
}));

// 请求体解析
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true, limit: '10mb' }));

// 请求日志
// app.use(requestLogger); // 暂时注释掉

// 限流中间件（暂时禁用以便测试）
// app.use(rateLimiter.apiLimit);

// 健康检查端点
app.get('/api/health', (req, res) => {
  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    environment: process.env.NODE_ENV,
    modules: {
      fetcherManager: fetcherManager ? 'initialized' : 'not loaded',
      liteLLMAdapter: liteLLMAdapter ? 'initialized' : 'not loaded',
      notificationDispatcher: notificationDispatcher ? 'initialized' : 'not loaded',
      reportGenerator: reportGenerator ? 'initialized' : 'not loaded',
      storageManager: storageManager.isConnected ? 'connected' : 'disconnected'
    }
  });
});


// 初始化内存存储（统一使用轻量级存储方案）
console.log("使用轻量级存储方案（内存 + JSON 文件）");
initializeMemoryStorage();

// 初始化内存存储
async function initializeMemoryStorage() {
  try {
    console.log('正在初始化内存存储...');
    await updateMemoryHotTopics();
    console.log('内存存储初始化完成');
  } catch (error) {
    console.warn('内存存储初始化失败:', error.message);
  }
}

// 更新内存中的热点数据
async function updateMemoryHotTopics() {
  try {
    const { newsNowFetcher } = require('./fetchers/NewsNowFetcher');
    const topics = await newsNowFetcher.fetch();

    memoryStorage.hotTopics = topics.map((topic, index) => ({
      ...topic,
      _id: `mem-${Date.now()}-${index}`,
      createdAt: new Date(),
      updatedAt: new Date()
    }));
    memoryStorage.lastUpdate = new Date();
    console.log(`内存存储更新: ${memoryStorage.hotTopics.length} 条热点数据`);
    return memoryStorage.hotTopics;
  } catch (error) {
    console.error('更新内存热点数据失败:', error.message);
    return [];
  }
}

// 注释：项目不使用 MongoDB，以下注释说明已过时
// 路由 - 统一使用内存版本（不依赖 MongoDB）
// 内存版本使用 JSON 文件存储数据
app.use('/api/hot-topics', require('./routes/hotTopicsMemory'));
app.use('/api/content', require('./routes/contentRewrite'));
app.use('/api/analytics', require('./routes/analytics'));
app.use('/api/auth', require('./routes/auth'));
app.use('/api/video', require('./routes/video'));
app.use('/api/transcription', require('./routes/transcription'));
app.use('/api/llm', require('./routes/llm'));
app.use('/api/tasks', require('./routes/tasks')); // 任务管理API
app.use('/api/v1/platforms', require('./routes/platforms')); // 平台管理API代理
app.use('/api/v1/publish', require('./routes/publish')); // 发布管理API代理
app.use('/api/v1/tasks', require('./routes/tasksGo')); // 任务管理API代理（Go后端）

// 404处理
app.use('*', (req, res) => {
  res.status(404).json({
    success: false,
    message: 'API端点不存在'
  });
});

// 错误处理
// app.use(errorLogger); // 暂时注释掉

// 定时任务 - 每30分钟更新热点数据
cron.schedule('*/30 * * * *', async () => {
  try {
    console.log('执行定时热点数据更新...');

    // 使用新的 FetcherManager 获取热点
    const topics = await fetcherManager.fetchAll();
    console.log(`获取到 ${topics.length} 条热点数据`);

    // 使用 TopicAnalyzer 分析热点
    for (const topic of topics) {
      topic.category = topicAnalyzer.categorize(topic.title);
      topic.keywords = topicAnalyzer.extractKeywords(topic.title);
      topic.suitability = topicAnalyzer.calculateSuitability(topic.title, topic.description);
    }

    // 保存到数据库
    if (storageManager.isConnected) {
      const saved = await storageManager.saveTopicsBatch(topics);
      console.log(`保存了 ${saved} 条热点数据`);
    }

    console.log('热点数据更新完成');
  } catch (error) {
    console.error('定时热点更新失败:', error);
  }
});

// 定时任务 - 每小时清理缓存
cron.schedule('0 * * * *', async () => {
  try {
    console.log('执行缓存清理...');
    // 清理各种缓存
    console.log('缓存清理完成');
  } catch (error) {
    console.error('缓存清理失败:', error);
  }
});

// 启动服务器
app.listen(PORT, () => {
  console.log(`服务器运行在端口 ${PORT}`);
});

module.exports = app;
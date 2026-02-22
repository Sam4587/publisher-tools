const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const compression = require('compression');

const app = express();
const PORT = process.env.PORT || 3001;

// 错误处理 - 捕获未处理的异常
process.on('uncaughtException', (error) => {
  console.error('[ERROR] Uncaught Exception:', error);
  // 不要退出进程，继续运行
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('[ERROR] Unhandled Rejection at:', promise, 'reason:', reason);
  // 不要退出进程，继续运行
});

// 优雅关闭
process.on('SIGTERM', () => {
  console.log('[INFO] 收到 SIGTERM 信号，正在关闭...');
  server.close(() => {
    console.log('[INFO] 服务器已关闭');
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('[INFO] 收到 SIGINT 信号，正在关闭...');
  server.close(() => {
    console.log('[INFO] 服务器已关闭');
    process.exit(0);
  });
});

app.use(helmet({
  contentSecurityPolicy: false,
  crossOriginEmbedderPolicy: false,
}));

app.use(compression());

app.use(cors({
  origin: '*',
  credentials: true
}));

app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true, limit: '10mb' }));

// 请求日志中间件
app.use((req, res, next) => {
  console.log(`[${new Date().toISOString()}] ${req.method} ${req.url}`);
  next();
});

app.get('/api/health', (req, res) => {
  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    message: 'Node backend is running',
    memory: process.memoryUsage()
  });
});

app.get('/api/platforms', (req, res) => {
  res.json({
    success: true,
    platforms: [
      { id: 'douyin', name: '抖音', icon: 'douyin' },
      { id: 'xiaohongshu', name: '小红书', icon: 'xiaohongshu' },
      { id: 'toutiao', name: '今日头条', icon: 'toutiao' },
      { id: 'bilibili', name: 'B站', icon: 'bilibili' }
    ]
  });
});

// 添加 /api/v1/publisher/platforms 路由
app.get('/api/v1/publisher/platforms', (req, res) => {
  res.json({
    success: true,
    data: [
      'douyin',
      'xiaohongshu',
      'toutiao',
      'bilibili'
    ]
  });
});

app.get('/api/tasks', (req, res) => {
  res.json({ success: true, tasks: [] });
});

app.post('/api/publish', (req, res) => {
  res.json({ success: true, message: 'Publish endpoint - connect to Go backend for full functionality' });
});

// 平台登录状态检查
app.get('/api/v1/publisher/platforms/:platform/check', (req, res) => {
  const platform = req.params.platform;
  console.log(`[INFO] Checking login status for platform: ${platform}`);
  
  // 模拟返回未登录状态
  res.json({
    success: true,
    data: {
      platform: platform,
      logged_in: false,
      account_name: null,
      avatar: null,
      last_check: new Date().toISOString()
    }
  });
});

// 平台登录
app.post('/api/v1/publisher/platforms/:platform/login', (req, res) => {
  const platform = req.params.platform;
  console.log(`[INFO] Login request for platform: ${platform}`);
  
  // 生成一个简单的 SVG 二维码占位符
  const svgQRCode = `data:image/svg+xml,${encodeURIComponent(`
    <svg xmlns="http://www.w3.org/2000/svg" width="300" height="300" viewBox="0 0 300 300">
      <rect width="300" height="300" fill="#ffffff"/>
      <text x="150" y="120" font-family="Arial, sans-serif" font-size="16" text-anchor="middle" fill="#333">
        ${platform.toUpperCase()}
      </text>
      <text x="150" y="150" font-family="Arial, sans-serif" font-size="14" text-anchor="middle" fill="#666">
        扫码登录
      </text>
      <text x="150" y="180" font-family="Arial, sans-serif" font-size="12" text-anchor="middle" fill="#999">
        (模拟二维码)
      </text>
      <rect x="50" y="200" width="200" height="60" fill="#f0f0f0" rx="5"/>
      <text x="150" y="235" font-family="Arial, sans-serif" font-size="11" text-anchor="middle" fill="#888">
        请使用 ${platform} APP 扫码
      </text>
    </svg>
  `)}`;
  
  // 模拟返回二维码登录信息
  res.json({
    success: true,
    data: {
      qrcode_url: svgQRCode,
      expires_at: new Date(Date.now() + 5 * 60 * 1000).toISOString() // 5分钟后过期
    }
  });
});

// 平台登出
app.post('/api/v1/publisher/platforms/:platform/logout', (req, res) => {
  const platform = req.params.platform;
  console.log(`[INFO] Logout request for platform: ${platform}`);
  
  res.json({
    success: true,
    data: {
      platform: platform,
      message: 'Logout successful'
    }
  });
});

// =====================================================
// 热点监控 API (模拟数据)
// =====================================================

// 获取热点列表
app.get('/api/hot-topics', (req, res) => {
  const limit = parseInt(req.query.limit) || 50;
  console.log(`[INFO] Getting hot topics, limit: ${limit}`);
  
  // 模拟热点数据
  const mockTopics = [
    {
      _id: '1',
      title: 'AI 技术突破：GPT-5 即将发布',
      description: 'OpenAI 宣布下一代语言模型即将发布',
      category: '科技',
      heat: 9999,
      trend: 'hot',
      source: 'weibo',
      keywords: ['AI', 'GPT-5', 'OpenAI'],
      suitability: 85,
      publishedAt: new Date().toISOString(),
      createdAt: new Date().toISOString()
    },
    {
      _id: '2',
      title: '新能源汽车销量创新高',
      description: '2024年新能源汽车销量突破千万辆',
      category: '财经',
      heat: 8888,
      trend: 'up',
      source: 'toutiao',
      keywords: ['新能源', '汽车', '销量'],
      suitability: 90,
      publishedAt: new Date().toISOString(),
      createdAt: new Date().toISOString()
    },
    {
      _id: '3',
      title: '春节档电影票房破纪录',
      description: '2024年春节档总票房突破80亿',
      category: '娱乐',
      heat: 7777,
      trend: 'hot',
      source: 'douyin',
      keywords: ['春节档', '电影', '票房'],
      suitability: 75,
      publishedAt: new Date().toISOString(),
      createdAt: new Date().toISOString()
    }
  ];
  
  res.json({
    success: true,
    data: mockTopics.slice(0, limit),
    pagination: {
      page: 1,
      limit: limit,
      total: mockTopics.length,
      pages: 1
    }
  });
});

// 获取数据源列表
app.get('/api/hot-topics/newsnow/sources', (req, res) => {
  console.log('[INFO] Getting hot sources');
  
  res.json({
    success: true,
    data: [
      { id: 'weibo', name: '微博热搜', enabled: true },
      { id: 'douyin', name: '抖音热点', enabled: true },
      { id: 'toutiao', name: '今日头条', enabled: true },
      { id: 'zhihu', name: '知乎热榜', enabled: true },
      { id: 'bilibili', name: 'B站热门', enabled: true }
    ]
  });
});

// 抓取热点
app.post('/api/hot-topics/newsnow/fetch', (req, res) => {
  console.log('[INFO] Fetching hot topics from sources');
  
  res.json({
    success: true,
    data: {
      fetched: 15,
      saved: 12,
      topics: []
    }
  });
});

// 获取新增热点
app.get('/api/hot-topics/trends/new', (req, res) => {
  const hours = parseInt(req.query.hours) || 24;
  console.log(`[INFO] Getting new hot topics in last ${hours} hours`);
  
  res.json({
    success: true,
    data: []
  });
});

// 错误处理中间件
app.use((err, req, res, next) => {
  console.error('[ERROR] Request error:', err);
  res.status(500).json({
    success: false,
    error: 'Internal server error',
    message: err.message
  });
});

// 404 处理
app.use((req, res) => {
  res.status(404).json({
    success: false,
    error: 'Not found',
    message: `Route ${req.method} ${req.url} not found`
  });
});

const server = app.listen(PORT, () => {
  console.log(`[INFO] Node backend running on port ${PORT}`);
  console.log(`[INFO] Health check: http://localhost:${PORT}/api/health`);
  console.log(`[INFO] Process ID: ${process.pid}`);
});

// 保持进程活跃
setInterval(() => {
  // 心跳检测
}, 30000);

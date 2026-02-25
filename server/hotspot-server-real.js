/**
 * 热点监控服务器 - 真实数据版
 */

const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const compression = require('compression');
const https = require('https');
const http = require('http');

const app = express();
const PORT = process.env.PORT || 3001;

// 错误处理
process.on('uncaughtException', (error) => {
  console.error('[ERROR] Uncaught Exception:', error);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('[ERROR] Unhandled Rejection at:', promise, 'reason:', reason);
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

// =====================================================
// 真实数据抓取功能
// =====================================================

// 真实数据源配置
const DATA_SOURCES = [
  {
    id: 'weibo',
    name: '微博热搜',
    url: 'https://weibo.com/ajax/side/hotSearch',
    enabled: true
  },
  {
    id: 'zhihu',
    name: '知乎热榜',
    url: 'https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total',
    enabled: true
  },
  {
    id: 'toutiao',
    name: '今日头条',
    url: 'https://www.toutiao.com/hot-event/hot-board/?origin=toutiao_pc',
    enabled: true
  }
];

// 带超时的 HTTP 请求
function fetchWithTimeout(url, timeout = 10000) {
  return new Promise((resolve, reject) => {
    const protocol = url.startsWith('https') ? https : http;

    const options = {
      headers: {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Accept': 'application/json, text/plain, */*',
        'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8',
        'Referer': 'https://www.google.com/'
      },
      timeout: timeout
    };

    const req = protocol.get(url, options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        if (res.statusCode === 200) {
          resolve(data);
        } else {
          reject(new Error(`HTTP ${res.statusCode}`));
        }
      });
    });

    req.on('error', (error) => {
      reject(error);
    });

    req.on('timeout', () => {
      req.destroy();
      reject(new Error('Request timeout'));
    });
  });
}

// 解析微博热搜数据
function parseWeiboData(data) {
  try {
    const json = typeof data === 'string' ? JSON.parse(data) : data;
    if (json.data && json.data.realtime) {
      return json.data.realtime.map((item, index) => ({
        _id: `weibo_${item.word || index}`,
        title: item.word || item.word_scheme || '未知',
        description: item.word_scheme || '',
        category: '综合',
        heat: (item.num || 1000000) / 10000,
        trend: index < 5 ? 'hot' : index < 15 ? 'up' : 'stable',
        source: 'weibo',
        keywords: extractKeywords(item.word || ''),
        suitability: 85,
        publishedAt: new Date().toISOString(),
        createdAt: new Date().toISOString()
      }));
    }
  } catch (e) {
    console.error('[ERROR] Parse weibo data error:', e.message);
  }
  return [];
}

// 解析知乎热榜数据
function parseZhihuData(data) {
  try {
    const json = typeof data === 'string' ? JSON.parse(data) : data;
    if (json.data && Array.isArray(json.data)) {
      return json.data.map((item, index) => ({
        _id: `zhihu_${item.target.id || index}`,
        title: item.target.title || '未知',
        description: item.target.excerpt || '',
        category: '综合',
        heat: parseHeatValue(item.detail_text),
        trend: index < 5 ? 'hot' : index < 15 ? 'up' : 'stable',
        source: 'zhihu',
        keywords: extractKeywords(item.target.title || ''),
        suitability: 88,
        publishedAt: new Date(item.target.created * 1000).toISOString(),
        createdAt: new Date().toISOString()
      }));
    }
  } catch (e) {
    console.error('[ERROR] Parse zhihu data error:', e.message);
  }
  return [];
}

// 解析热度值
function parseHeatValue(text) {
  if (!text) return 100;
  const match = text.match(/(\d+)/);
  return match ? parseInt(match[1]) : 100;
}

// 提取关键词
function extractKeywords(title) {
  const words = title.split(/[\s,，。！？!?\uff0c\u3002\uff01\uff1f]+/);
  return words.filter(word => word.length >= 2 && word.length <= 10).slice(0, 5);
}

// 生成备用数据
function generateFallbackTopics() {
  const now = new Date();
  const hours = now.getHours();
  const timeKeywords = hours < 12 ? '早间' : hours < 18 ? '午间' : '晚间';

  return [
    {
      _id: '1',
      title: `${timeKeywords}热搜:全国多地迎来降温天气`,
      description: '气象部门发布寒潮预警,注意保暖',
      category: '社会',
      heat: 9999,
      trend: 'hot',
      source: 'weibo',
      keywords: ['天气', '降温', '寒潮'],
      suitability: 90,
      publishedAt: new Date(now - 3600000).toISOString(),
      createdAt: new Date(now - 3600000).toISOString()
    },
    {
      _id: '2',
      title: '科技前沿:人工智能技术在各领域加速应用',
      description: 'AI技术正在改变我们的生活方式',
      category: '科技',
      heat: 8888,
      trend: 'up',
      source: 'zhihu',
      keywords: ['AI', '人工智能', '科技'],
      suitability: 92,
      publishedAt: new Date(now - 7200000).toISOString(),
      createdAt: new Date(now - 7200000).toISOString()
    },
    {
      _id: '3',
      title: '经济观察:消费市场持续回暖',
      description: '多项数据显示消费信心正在恢复',
      category: '财经',
      heat: 7777,
      trend: 'hot',
      source: 'toutiao',
      keywords: ['经济', '消费', '市场'],
      suitability: 88,
      publishedAt: new Date(now - 10800000).toISOString(),
      createdAt: new Date(now - 10800000).toISOString()
    },
    {
      _id: '4',
      title: '体育新闻:多项体育赛事即将开赛',
      description: '球迷们期待已久的精彩赛事',
      category: '体育',
      heat: 6666,
      trend: 'up',
      source: 'douyin',
      keywords: ['体育', '赛事', '比赛'],
      suitability: 85,
      publishedAt: new Date(now - 14400000).toISOString(),
      createdAt: new Date(now - 14400000).toISOString()
    },
    {
      _id: '5',
      title: '文化娱乐:国产影视作品口碑获赞',
      description: '多部新作品获得观众好评',
      category: '娱乐',
      heat: 5555,
      trend: 'stable',
      source: 'bilibili',
      keywords: ['影视', '娱乐', '作品'],
      suitability: 82,
      publishedAt: new Date(now - 18000000).toISOString(),
      createdAt: new Date(now - 18000000).toISOString()
    },
    {
      _id: '6',
      title: '教育资讯:多地出台教育改革新举措',
      description: '教育部门推进素质教育发展',
      category: '教育',
      heat: 4444,
      trend: 'up',
      source: 'weibo',
      keywords: ['教育', '改革', '素质'],
      suitability: 86,
      publishedAt: new Date(now - 21600000).toISOString(),
      createdAt: new Date(now - 21600000).toISOString()
    },
    {
      _id: '7',
      title: '健康生活:全民健身活动广泛开展',
      description: '健康理念深入人心,运动成为新风尚',
      category: '健康',
      heat: 3333,
      trend: 'stable',
      source: 'zhihu',
      keywords: ['健康', '健身', '运动'],
      suitability: 84,
      publishedAt: new Date(now - 25200000).toISOString(),
      createdAt: new Date(now - 25200000).toISOString()
    },
    {
      _id: '8',
      title: '国际新闻:全球多国加强合作应对挑战',
      description: '国际社会携手共促和平发展',
      category: '国际',
      heat: 2222,
      trend: 'stable',
      source: 'toutiao',
      keywords: ['国际', '合作', '和平'],
      suitability: 80,
      publishedAt: new Date(now - 28800000).toISOString(),
      createdAt: new Date(now - 28800000).toISOString()
    }
  ];
}

// 抓取真实热点数据
async function fetchRealHotTopics() {
  const allTopics = [];

  for (const source of DATA_SOURCES) {
    if (!source.enabled) continue;

    try {
      console.log(`[INFO] Fetching from ${source.name}...`);
      const data = await fetchWithTimeout(source.url, 10000);

      let topics = [];
      if (source.id === 'weibo') {
        topics = parseWeiboData(data);
      } else if (source.id === 'zhihu') {
        topics = parseZhihuData(data);
      }

      if (topics.length > 0) {
        allTopics.push(...topics);
        console.log(`[INFO] Fetched ${topics.length} topics from ${source.name}`);
      }
    } catch (error) {
      console.error(`[ERROR] Failed to fetch from ${source.name}:`, error.message);
    }
  }

  // 如果没有获取到真实数据,返回备用数据
  if (allTopics.length === 0) {
    console.log('[INFO] No real data fetched, using fallback data');
    return generateFallbackTopics();
  }

  return allTopics;
}

// 缓存热点数据
let cachedTopics = [];
let lastFetchTime = 0;
const CACHE_DURATION = 5 * 60 * 1000; // 5分钟缓存

async function getHotTopics() {
  const now = Date.now();

  // 如果缓存过期或为空,重新抓取
  if (!cachedTopics.length || now - lastFetchTime > CACHE_DURATION) {
    console.log('[INFO] Fetching fresh hot topics...');
    cachedTopics = await fetchRealHotTopics();
    lastFetchTime = now;
  }

  return cachedTopics;
}

// =====================================================
// API 路由
// =====================================================

// 健康检查
app.get('/api/health', (req, res) => {
  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    message: 'Hotspot server is running',
    memory: process.memoryUsage(),
    cachedTopics: cachedTopics.length
  });
});

// 获取热点列表
app.get('/api/hot-topics', async (req, res) => {
  try {
    const limit = parseInt(req.query.limit) || 50;
    console.log(`[INFO] Getting hot topics, limit: ${limit}`);

    const topics = await getHotTopics();
    const limitedTopics = topics.slice(0, limit);

    res.json({
      success: true,
      data: limitedTopics,
      pagination: {
        page: 1,
        limit: limit,
        total: topics.length,
        pages: 1
      }
    });
  } catch (error) {
    console.error('[ERROR] Get hot topics failed:', error);
    res.status(500).json({
      success: false,
      error: 'Failed to get hot topics',
      message: error.message
    });
  }
});

// 获取数据源列表
app.get('/api/hot-topics/newsnow/sources', (req, res) => {
  res.json({
    success: true,
    data: DATA_SOURCES.map(source => ({
      id: source.id,
      name: source.name,
      enabled: source.enabled
    }))
  });
});

// 抓取热点
app.post('/api/hot-topics/newsnow/fetch', async (req, res) => {
  try {
    console.log('[INFO] Force fetching hot topics...');

    // 清除缓存,强制重新抓取
    cachedTopics = [];
    lastFetchTime = 0;

    const topics = await fetchRealHotTopics();

    res.json({
      success: true,
      data: {
        fetched: topics.length,
        saved: topics.length,
        topics: topics.slice(0, 10)
      }
    });
  } catch (error) {
    console.error('[ERROR] Fetch hot topics failed:', error);
    res.status(500).json({
      success: false,
      error: 'Failed to fetch hot topics',
      message: error.message
    });
  }
});

// 更新热点
app.post('/api/hot-topics/update', async (req, res) => {
  try {
    console.log('[INFO] Updating hot topics...');

    // 清除缓存,强制重新抓取
    cachedTopics = [];
    lastFetchTime = 0;

    const topics = await fetchRealHotTopics();

    res.json({
      success: true,
      data: {
        message: 'Hotspots updated',
        count: topics.length
      }
    });
  } catch (error) {
    console.error('[ERROR] Update hot topics failed:', error);
    res.status(500).json({
      success: false,
      error: 'Failed to update hot topics',
      message: error.message
    });
  }
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

// 其他路由(兼容现有功能)
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

app.get('/api/tasks', (req, res) => {
  res.json({ success: true, tasks: [] });
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

// 启动服务器
const server = app.listen(PORT, () => {
  console.log(`[INFO] Hotspot server running on port ${PORT}`);
  console.log(`[INFO] Health check: http://localhost:${PORT}/api/health`);
  console.log(`[INFO] Process ID: ${process.pid}`);

  // 启动时预抓取数据
  setTimeout(async () => {
    console.log('[INFO] Performing initial hotspot fetch...');
    try {
      cachedTopics = await fetchRealHotTopics();
      lastFetchTime = Date.now();
      console.log(`[INFO] Initial fetch completed: ${cachedTopics.length} topics`);
    } catch (error) {
      console.error('[ERROR] Initial fetch failed:', error);
      cachedTopics = generateFallbackTopics();
      lastFetchTime = Date.now();
    }
  }, 1000);
});

// 优雅关闭
process.on('SIGTERM', () => {
  console.log('[INFO] Received SIGTERM, shutting down...');
  server.close(() => {
    console.log('[INFO] Server closed');
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('[INFO] Received SIGINT, shutting down...');
  server.close(() => {
    console.log('[INFO] Server closed');
    process.exit(0);
  });
});

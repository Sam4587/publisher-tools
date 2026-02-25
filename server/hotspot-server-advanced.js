/**
 * 热点监控服务器 - 高级版
 * 借鉴 TrendRadar 的设计理念
 *
 * 核心特性:
 * 1. 多数据源支持 (NewsNow API + RSS 源)
 * 2. 智能反爬虫对策 (代理、重试、随机延迟)
 * 3. 数据缓存和新鲜度过滤
 * 4. 请求限流和速率控制
 * 5. 优雅降级机制
 */

const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const compression = require('compression');
const https = require('https');
const http = require('http');
const { URL } = require('url');

const app = express();
const PORT = process.env.PORT || 3002;

// =====================================================
// 配置
// =====================================================

const CONFIG = {
  // NewsNow API 配置
  newsnow: {
    apiUrl: 'https://newsnow.busiyi.world/api/s',
    timeout: 10000,
    maxRetries: 3,
    minRetryWait: 2,
    maxRetryWait: 5,
    useProxy: false,
    proxyUrl: null
  },

  // RSS 源配置
  rss: {
    enabled: true,
    timeout: 15000,
    maxItemsPerFeed: 20,
    maxAgeDays: 7,
    requestInterval: 1000, // 毫秒
    feeds: [
      {
        id: '36kr',
        name: '36氪',
        url: 'https://36kr.com/feed',
        category: '科技',
        enabled: true
      },
      {
        id: 'huxiu',
        name: '虎嗅网',
        url: 'https://www.huxiu.com/rss/0.xml',
        category: '科技',
        enabled: true
      },
      {
        id: 'ifanr',
        name: '爱范儿',
        url: 'https://www.ifanr.com/feed',
        category: '科技',
        enabled: true
      },
      {
        id: 'geekpark',
        name: '极客公园',
        url: 'https://www.geekpark.net/rss',
        category: '科技',
        enabled: true
      },
      {
        id: 'techcrunch',
        name: 'TechCrunch',
        url: 'https://techcrunch.com/feed/',
        category: '科技',
        enabled: true
      },
      {
        id: 'caixin',
        name: '财新网',
        url: 'https://www.caixin.com/rss/newswire.xml',
        category: '财经',
        enabled: true
      },
      {
        id: 'eastmoney',
        name: '东方财富',
        url: 'https://finance.eastmoney.com/a/cjjsp.xml',
        category: '财经',
        enabled: true
      },
      {
        id: 'cnbeta',
        name: 'cnBeta',
        url: 'https://www.cnbeta.com/backend.php',
        category: '科技',
        enabled: true
      },
      {
        id: 'solidot',
        name: 'Solidot',
        url: 'https://www.solidot.org/index.rss',
        category: '科技',
        enabled: true
      }
    ]
  },

  // 缓存配置
  cache: {
    duration: 5 * 60 * 1000, // 5分钟
    maxSize: 1000
  },

  // 请求限流
  rateLimit: {
    requestsPerMinute: 60,
    burst: 10
  }
};

// =====================================================
// 数据源管理器
// =====================================================

class DataSourceManager {
  constructor() {
    this.sources = new Map();
    this.initializeSources();
  }

  initializeSources() {
    // 注册 NewsNow 数据源
    this.registerSource('newsnow', new NewsNowDataSource(CONFIG.newsnow));

    // 注册 RSS 数据源
    CONFIG.rss.feeds.forEach(feed => {
      if (feed.enabled) {
        this.registerSource(`rss_${feed.id}`, new RSSDataSource(feed, CONFIG.rss));
      }
    });
  }

  registerSource(id, source) {
    this.sources.set(id, source);
  }

  getSource(id) {
    return this.sources.get(id);
  }

  getAllSources() {
    return Array.from(this.sources.values());
  }

  getEnabledSources() {
    return this.getAllSources().filter(s => s.isEnabled());
  }
}

// =====================================================
// NewsNow 数据源
// =====================================================

class NewsNowDataSource {
  constructor(config) {
    this.id = 'newsnow';
    this.name = 'NewsNow 热点';
    this.config = config;
    this.enabled = true;
  }

  isEnabled() {
    return this.enabled;
  }

  async fetch(maxItems = 50) {
    const platforms = ['weibo', 'zhihu', 'douyin', 'toutiao', 'baidu'];
    const allTopics = [];

    for (const platform of platforms) {
      try {
        const topics = await this.fetchPlatform(platform);
        allTopics.push(...topics);

        if (allTopics.length >= maxItems) {
          break;
        }

        // 随机延迟,避免被识别为爬虫
        await this.randomDelay(500, 1500);
      } catch (error) {
        console.error(`[ERROR] Failed to fetch ${platform}:`, error.message);
      }
    }

    return allTopics.slice(0, maxItems);
  }

  async fetchPlatform(platformId) {
    const url = `${this.config.apiUrl}?id=${platformId}&latest`;
    let retries = 0;

    while (retries <= this.config.maxRetries) {
      try {
        const data = await this.fetchWithRetry(url);
        return this.parseResponse(data, platformId);
      } catch (error) {
        retries++;

        if (retries > this.config.maxRetries) {
          throw error;
        }

        const waitTime = this.calculateRetryDelay(retries);
        console.log(`[INFO] Retry ${retries}/${this.config.maxRetries} for ${platformId}, waiting ${waitTime}ms`);
        await this.sleep(waitTime);
      }
    }

    return [];
  }

  async fetchWithRetry(url) {
    return new Promise((resolve, reject) => {
      const protocol = url.startsWith('https') ? https : http;

      const options = {
        headers: this.generateHeaders(),
        timeout: this.config.timeout
      };

      const req = protocol.get(url, options, (res) => {
        let data = '';

        res.on('data', chunk => data += chunk);
        res.on('end', () => {
          if (res.statusCode === 200) {
            // 检查是否被 Cloudflare 阻挡
            if (data.includes('Cloudflare') || data.includes('Attention Required')) {
              reject(new Error('Blocked by Cloudflare'));
            } else {
              resolve(data);
            }
          } else {
            reject(new Error(`HTTP ${res.statusCode}`));
          }
        });
      });

      req.on('error', reject);
      req.on('timeout', () => {
        req.destroy();
        reject(new Error('Request timeout'));
      });
    });
  }

  generateHeaders() {
    const userAgents = [
      'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
      'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
      'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0',
      'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15'
    ];

    return {
      'User-Agent': userAgents[Math.floor(Math.random() * userAgents.length)],
      'Accept': 'application/json, text/plain, */*',
      'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8',
      'Accept-Encoding': 'gzip, deflate, br',
      'Connection': 'keep-alive',
      'Cache-Control': 'no-cache',
      'Referer': 'https://www.google.com/'
    };
  }

  parseResponse(data, platform) {
    try {
      const json = typeof data === 'string' ? JSON.parse(data) : data;

      if (json.status === 'success' || json.status === 'cache') {
        const items = json.data?.items || json.items || [];

        return items.map((item, index) => ({
          _id: `${platform}_${item.id || index}`,
          title: item.title || '未知',
          description: item.description || '',
          category: this.mapCategory(platform),
          heat: this.calculateHeat(index),
          trend: this.calculateTrend(index),
          source: platform,
          keywords: this.extractKeywords(item.title || ''),
          suitability: 85,
          publishedAt: new Date().toISOString(),
          createdAt: new Date().toISOString()
        }));
      }
    } catch (error) {
      console.error('[ERROR] Parse NewsNow response error:', error.message);
    }

    return [];
  }

  mapCategory(platform) {
    const categoryMap = {
      'weibo': '综合',
      'zhihu': '综合',
      'douyin': '娱乐',
      'toutiao': '综合',
      'baidu': '综合'
    };
    return categoryMap[platform] || '综合';
  }

  calculateHeat(index) {
    if (index < 5) return 100 - index * 5;
    if (index < 15) return 80 - (index - 5) * 3;
    return 50 - (index - 15);
  }

  calculateTrend(index) {
    if (index < 5) return 'hot';
    if (index < 15) return 'up';
    return 'stable';
  }

  extractKeywords(title) {
    const words = title.split(/[\s,，。！？!?\uff0c\u3002\uff01\uff1f]+/);
    return words.filter(word => word.length >= 2 && word.length <= 10).slice(0, 5);
  }

  calculateRetryDelay(retry) {
    const { minRetryWait, maxRetryWait } = this.config;
    const baseWait = minRetryWait + Math.random() * (maxRetryWait - minRetryWait);
    const additionalWait = (retry - 1) * Math.random() * 2;
    return (baseWait + additionalWait) * 1000;
  }

  randomDelay(min, max) {
    const delay = min + Math.random() * (max - min);
    return this.sleep(delay);
  }

  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// =====================================================
// RSS 数据源
// =====================================================

class RSSDataSource {
  constructor(feedConfig, rssConfig) {
    this.id = feedConfig.id;
    this.name = feedConfig.name;
    this.url = feedConfig.url;
    this.category = feedConfig.category;
    this.enabled = true;
    this.config = rssConfig;
  }

  isEnabled() {
    return this.enabled;
  }

  async fetch(maxItems = 20) {
    try {
      const data = await this.fetchRSS();
      return this.parseRSS(data, maxItems);
    } catch (error) {
      console.error(`[ERROR] Failed to fetch RSS ${this.name}:`, error.message);
      return [];
    }
  }

  async fetchRSS() {
    return new Promise((resolve, reject) => {
      const parsedUrl = new URL(this.url);
      const protocol = parsedUrl.protocol === 'https:' ? https : http;

      const options = {
        hostname: parsedUrl.hostname,
        port: parsedUrl.port,
        path: parsedUrl.pathname + parsedUrl.search,
        method: 'GET',
        headers: {
          'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
          'Accept': 'application/rss+xml, application/xml, text/xml, */*',
          'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8'
        },
        timeout: this.config.timeout
      };

      const req = protocol.request(options, (res) => {
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

      req.on('error', reject);
      req.on('timeout', () => {
        req.destroy();
        reject(new Error('Request timeout'));
      });

      req.end();
    });
  }

  parseRSS(data, maxItems) {
    try {
      // 简单的 XML 解析
      const items = this.extractRSSItems(data);
      const filteredItems = this.filterByFreshness(items);

      return filteredItems.slice(0, maxItems).map((item, index) => ({
        _id: `rss_${this.id}_${index}`,
        title: this.cleanText(item.title),
        description: this.cleanText(item.description || ''),
        category: this.category,
        heat: this.calculateRSSHeat(index),
        trend: 'stable',
        source: 'rss',
        keywords: this.extractKeywords(item.title),
        suitability: 80,
        publishedAt: item.pubDate || new Date().toISOString(),
        createdAt: new Date().toISOString()
      }));
    } catch (error) {
      console.error('[ERROR] Parse RSS error:', error.message);
      return [];
    }
  }

  extractRSSItems(xml) {
    const items = [];
    const itemRegex = /<item>([\s\S]*?)<\/item>/g;
    let match;

    while ((match = itemRegex.exec(xml)) !== null) {
      const itemContent = match[1];

      const titleMatch = itemContent.match(/<title>(?:<!\[CDATA\[)?(.*?)(?:\]\]>)?<\/title>/);
      const linkMatch = itemContent.match(/<link>(.*?)<\/link>/);
      const descMatch = itemContent.match(/<description>(?:<!\[CDATA\[)?(.*?)(?:\]\]>)?<\/description>/);
      const pubDateMatch = itemContent.match(/<pubDate>(.*?)<\/pubDate>/);

      items.push({
        title: titleMatch ? titleMatch[1] : '未知',
        link: linkMatch ? linkMatch[1] : '',
        description: descMatch ? descMatch[1] : '',
        pubDate: pubDateMatch ? new Date(pubDateMatch[1]).toISOString() : null
      });
    }

    return items;
  }

  filterByFreshness(items) {
    if (!this.config.maxAgeDays || this.config.maxAgeDays === 0) {
      return items;
    }

    const cutoffDate = new Date(Date.now() - this.config.maxAgeDays * 24 * 60 * 60 * 1000);

    return items.filter(item => {
      if (!item.pubDate) return true;
      const itemDate = new Date(item.pubDate);
      return itemDate > cutoffDate;
    });
  }

  calculateRSSHeat(index) {
    return Math.max(50, 80 - index * 2);
  }

  cleanText(text) {
    if (!text) return '';
    return text
      .replace(/<!\[CDATA\[|\]\]>/g, '')
      .replace(/<[^>]*>/g, '')
      .replace(/&nbsp;/g, ' ')
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&amp;/g, '&')
      .replace(/&quot;/g, '"')
      .replace(/&#39;/g, "'")
      .trim()
      .substring(0, 200);
  }

  extractKeywords(title) {
    const words = title.split(/[\s,，。！？!?\uff0c\u3002\uff01\uff1f]+/);
    return words.filter(word => word.length >= 2 && word.length <= 10).slice(0, 5);
  }
}

// =====================================================
// 缓存管理器
// =====================================================

class CacheManager {
  constructor(config) {
    this.config = config;
    this.cache = new Map();
    this.timestamps = new Map();
  }

  set(key, value) {
    // 如果缓存已满,删除最旧的条目
    if (this.cache.size >= this.config.maxSize) {
      const oldestKey = this.findOldestKey();
      if (oldestKey) {
        this.cache.delete(oldestKey);
        this.timestamps.delete(oldestKey);
      }
    }

    this.cache.set(key, value);
    this.timestamps.set(key, Date.now());
  }

  get(key) {
    const timestamp = this.timestamps.get(key);
    if (!timestamp) return null;

    // 检查是否过期
    if (Date.now() - timestamp > this.config.duration) {
      this.cache.delete(key);
      this.timestamps.delete(key);
      return null;
    }

    return this.cache.get(key);
  }

  findOldestKey() {
    let oldestKey = null;
    let oldestTimestamp = Infinity;

    for (const [key, timestamp] of this.timestamps.entries()) {
      if (timestamp < oldestTimestamp) {
        oldestTimestamp = timestamp;
        oldestKey = key;
      }
    }

    return oldestKey;
  }

  clear() {
    this.cache.clear();
    this.timestamps.clear();
  }

  size() {
    return this.cache.size;
  }
}

// =====================================================
// 速率限制器
// =====================================================

class RateLimiter {
  constructor(config) {
    this.config = config;
    this.requests = [];
  }

  async checkLimit() {
    const now = Date.now();
    const oneMinuteAgo = now - 60000;

    // 清理过期的请求记录
    this.requests = this.requests.filter(timestamp => timestamp > oneMinuteAgo);

    if (this.requests.length >= this.config.requestsPerMinute) {
      const oldestRequest = this.requests[0];
      const waitTime = oldestRequest + 60000 - now;

      if (waitTime > 0) {
        console.log(`[INFO] Rate limit reached, waiting ${waitTime}ms`);
        await this.sleep(waitTime);
      }
    }

    this.requests.push(now);
  }

  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// =====================================================
// 主服务
// =====================================================

const dataSourceManager = new DataSourceManager();
const cacheManager = new CacheManager(CONFIG.cache);
const rateLimiter = new RateLimiter(CONFIG.rateLimit);

let cachedTopics = [];
let lastFetchTime = 0;

async function fetchAllTopics() {
  const allTopics = [];
  const enabledSources = dataSourceManager.getEnabledSources();

  console.log(`[INFO] Fetching from ${enabledSources.length} sources...`);

  for (const source of enabledSources) {
    try {
      await rateLimiter.checkLimit();

      const topics = await source.fetch();
      allTopics.push(...topics);

      console.log(`[INFO] Fetched ${topics.length} topics from ${source.name}`);

      // 随机延迟,避免请求过于频繁
      await sleep(500 + Math.random() * 1000);
    } catch (error) {
      console.error(`[ERROR] Failed to fetch from ${source.name}:`, error.message);
    }
  }

  // 按热度排序
  allTopics.sort((a, b) => b.heat - a.heat);

  return allTopics;
}

async function getHotTopics() {
  const now = Date.now();

  // 检查缓存
  if (cachedTopics.length && now - lastFetchTime < CONFIG.cache.duration) {
    console.log('[INFO] Using cached topics');
    return cachedTopics;
  }

  // 获取新数据
  console.log('[INFO] Fetching fresh topics...');
  cachedTopics = await fetchAllTopics();
  lastFetchTime = now;

  // 保存到缓存
  cacheManager.set('hot_topics', cachedTopics);

  return cachedTopics;
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

// =====================================================
// Express 应用设置
// =====================================================

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
// API 路由
// =====================================================

// 健康检查
app.get('/api/health', (req, res) => {
  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    message: 'Advanced hotspot server is running',
    memory: process.memoryUsage(),
    cachedTopics: cachedTopics.length,
    sources: dataSourceManager.getEnabledSources().length
  });
});

// 获取热点列表
app.get('/api/hot-topics', async (req, res) => {
  try {
    const limit = parseInt(req.query.limit) || 50;
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
  const sources = dataSourceManager.getEnabledSources().map(source => ({
    id: source.id,
    name: source.name,
    enabled: source.isEnabled()
  }));

  res.json({
    success: true,
    data: sources
  });
});

// 强制刷新热点
app.post('/api/hot-topics/newsnow/fetch', async (req, res) => {
  try {
    console.log('[INFO] Force fetching hot topics...');

    // 清除缓存
    cacheManager.clear();
    cachedTopics = [];
    lastFetchTime = 0;

    const topics = await getHotTopics();

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

    cacheManager.clear();
    cachedTopics = [];
    lastFetchTime = 0;

    const topics = await getHotTopics();

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

  // =====================================================
  // AI 分析 API
  // =====================================================

  // AI 分析热点话题
  app.post("/api/hot-topics/ai/analyze", async (req, res) => {
    try {
      const { topics, options } = req.body;

      if (!topics || !Array.isArray(topics) || topics.length === 0) {
        return res.status(400).json({
          success: false,
          error: "Invalid request",
          message: "Topics array is required"
        });
      }

      console.log(`[INFO] AI analyzing ${topics.length} topics...`);

      // 模拟AI分析 - 实际项目中可以接入真实的AI服务
      const analysis = generateMockAnalysis(topics);

      res.json({
        success: true,
        data: analysis
      });
    } catch (error) {
      console.error("[ERROR] AI analyze failed:", error);
      res.status(500).json({
        success: false,
        error: "AI analysis failed",
        message: error.message
      });
    }
  });

  // 生成热点简报
  app.post("/api/hot-topics/ai/briefing", async (req, res) => {
    try {
      const { topics, maxLength } = req.body;

      if (!topics || !Array.isArray(topics) || topics.length === 0) {
        return res.status(400).json({
          success: false,
          error: "Invalid request",
          message: "Topics array is required"
        });
      }

      console.log(`[INFO] Generating briefing for ${topics.length} topics...`);

      // 生成简报
      const brief = generateBrief(topics, maxLength || 300);

      res.json({
        success: true,
        data: {
          brief: brief
        }
      });
    } catch (error) {
      console.error("[ERROR] Generate briefing failed:", error);
      res.status(500).json({
        success: false,
        error: "Briefing generation failed",
        message: error.message
      });
    }
  });

  // 模拟AI分析结果生成
  function generateMockAnalysis(topics) {
    // 提取关键词
    const allKeywords = topics.flatMap(t => t.keywords || []);
    const topKeywords = [...new Set(allKeywords)].slice(0, 5);

    // 生成摘要
    const summary = `分析了${topics.length}个热点话题。主要涉及领域包括：${topKeywords.join("、")}。这些话题在各大平台都有较高热度，值得关注和跟进。`;

    // 生成关键点
    const keyPoints = topics.slice(0, 3).map((topic, index) => {
      return `${index + 1}. ${topic.title} - 热度${topic.heat}，来自${topic.source}`;
    });

    // 判断情感倾向
    const sentiments = ["positive", "negative", "neutral"];
    const sentiment = sentiments[Math.floor(Math.random() * sentiments.length)];

    // 生成建议
    const recommendations = [
      "建议关注这些热点话题的发展趋势",
      "可以考虑结合这些话题创作相关内容",
      "注意跟踪话题的时效性和影响力变化"
    ];

    return {
      summary: summary,
      keyPoints: keyPoints,
      sentiment: sentiment,
      recommendations: recommendations,
      topics: topics.map(t => t.title),
      analyzedAt: new Date().toISOString()
    };
  }

  // 生成简报
  function generateBrief(topics, maxLength) {
    const titles = topics.slice(0, 5).map(t => t.title);
    let brief = `【热点简报】\n\n`;
    brief += `今日共监控到${topics.length}个热点话题。\n\n`;
    brief += `TOP热点：\n`;
    titles.forEach((title, index) => {
      brief += `${index + 1}. ${title}\n`;
    });
    brief += `\n更多详情请查看完整分析。`;

    // 如果超过最大长度，截断
    if (brief.length > maxLength) {
      brief = brief.substring(0, maxLength - 3) + "...";
    }

    return brief;
  }


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

// =====================================================
// 启动服务器
// =====================================================

const server = app.listen(PORT, () => {
  console.log(`[INFO] Advanced hotspot server running on port ${PORT}`);
  console.log(`[INFO] Health check: http://localhost:${PORT}/api/health`);
  console.log(`[INFO] Process ID: ${process.pid}`);
  console.log(`[INFO] Data sources: ${dataSourceManager.getEnabledSources().length}`);

  // 启动时预抓取数据
  setTimeout(async () => {
    console.log('[INFO] Performing initial hotspot fetch...');
    try {
      cachedTopics = await getHotTopics();
      console.log(`[INFO] Initial fetch completed: ${cachedTopics.length} topics`);
    } catch (error) {
      console.error('[ERROR] Initial fetch failed:', error);
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

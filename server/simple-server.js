const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const compression = require('compression');
const https = require('https');
const http = require('http');

const app = express();
const PORT = process.env.PORT || 3001;

// 错误处理 - 捕获未处理的异常
process.on('uncaughtException', (error) => {
  console.error('[ERROR] Uncaught Exception:', error);
  // 不要退出进程,继续运行
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('[ERROR] Unhandled Rejection at:', promise, 'reason:', reason);
  // 不要退出进程,继续运行
});

// 优雅关闭
process.on('SIGTERM', () => {
  console.log('[INFO] 收到 SIGTERM 信号,正在关闭...');
  server.close(() => {
    console.log('[INFO] 服务器已关闭');
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('[INFO] 收到 SIGINT 信号,正在关闭...');
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
      title: 'AI 技术突破:GPT-5 即将发布',
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

// =====================================================
// AI内容生成 API
// =====================================================

// 模拟内容生成函数
function generateMockContent(topic, type, style, length) {
  const lengthMap = {
    'short': 200,
    'medium': 500,
    'long': 1000
  };

  const wordCount = lengthMap[length] || 500;

  const templates = {
    'professional': `关于"${topic}"的专业分析。

首先,我们需要了解这个话题的背景和重要性。从专业角度来看,${topic}在当前环境下具有重要意义。

其次,从实践角度来看,我们需要关注以下几个关键点:
1. 核心概念的理解
2. 实际应用场景
3. 未来发展趋势

最后,总结来说,${topic}是一个值得深入研究和应用的领域。希望通过本文的分析,能够为读者提供有价值的参考。

希望这篇文章能帮助您更好地理解和应用相关知识。`,
    'engaging': `你有没有想过"${topic}"到底是什么?今天就让我们一起来探索这个有趣的话题!

想象一下,如果我们能够深入了解${topic},会有什么惊人的发现呢?

首先,${topic}其实离我们并不远。在我们的日常生活中,你可能会经常遇到相关的情况。

那么问题来了,为什么${topic}如此重要?

总的来说,${topic}是一个充满魅力的话题。通过今天的分享,相信你对它有了更深入的了解。

是不是觉得很有趣呢?欢迎在评论区分享你的想法和看法!`,
    'casual': `最近"${topic}"这个话题挺火的,我也来聊聊我的看法。

说实话,一开始我对${topic}也不是很了解。但是经过一段时间的研究和观察,我发现它确实挺有意思的。

我觉得吧,${topic}之所以这么受欢迎,主要是因为它解决了一些实际问题。

反正就是这么回事,大家觉得呢?欢迎来聊聊!`,
    'humorous': `哈哈,说到"${topic}",这可真是个有趣的话题!

你知道${topic}最有趣的地方是什么吗?就是它总能给人带来意想不到的惊喜。

想象一下,如果${topic}是一个人,那它一定是个幽默风趣的家伙,总能逗得大家哈哈大笑。

不过话说回来,${topic}虽然有趣,但我们还是得认真对待它。毕竟,有趣的背后往往隐藏着深刻的道理。

所以,下次再遇到${topic},记得微笑面对,说不定会有意外收获哦!😄`
  };

  const content = templates[style] || templates.professional;

  return {
    title: `【${type === 'article' ? '文章' : type === 'video_script' ? '视频脚本' : '内容'}】${topic}`,
    content: content,
    summary: content.substring(0, 100) + '...',
    wordCount: content.length,
    type: type,
    style: style,
    length: length,
    generatedAt: new Date().toISOString()
  };
}

// 模拟内容改写函数
function rewriteMockContent(content, style, tone, length) {
  const tonePrefix = {
    'neutral': '',
    'positive': '令人欣喜的是,',
    'negative': '遗憾的是,',
    'enthusiastic': '太棒了!'
  };

  const styleTransformations = {
    'professional': content.replace(/!/g, '。').replace(/哈哈|呵呵/g, '有趣的是'),
    'engaging': content.replace(/。/g, '!') + '\n\n你觉得呢?',
    'casual': '说句实话,' + content.replace(/正式|专业/g, '普通'),
    'formal': '综上所述,' + content.replace(/我觉得|我认为/g, '据分析')
  };

  const lengthModifications = {
    'shorter': content.substring(0, Math.floor(content.length * 0.7)) + '...',
    'same': content,
    'longer': content + '\n\n此外,还值得关注的是相关的细节和补充说明。'
  };

  const transformedContent = styleTransformations[style] || content;
  const modifiedContent = lengthModifications[length] || transformedContent;
  const finalContent = tonePrefix[tone] ? tonePrefix[tone] + '' + modifiedContent : modifiedContent;

  return {
    originalContent: content,
    rewrittenContent: finalContent,
    style: style,
    tone: tone,
    length: length,
    wordCount: finalContent.length,
    rewrittenAt: new Date().toISOString()
  };
}

// AI内容生成
app.post('/api/v1/ai/content/generate', (req, res) => {
  console.log('[INFO] AI content generation request:', req.body);

  try {
    const { topic, type = 'article', style = 'professional', length = 'medium', platform } = req.body;
      
      // 兼容前端参数格式
      let normalizedStyle = style;
      if (style === '轻松幽默') {
        normalizedStyle = 'humorous';
      } else if (style === '正式专业') {
        normalizedStyle = 'professional';
      } else if (!['professional', 'engaging', 'casual', 'humorous'].includes(normalizedStyle)) {
        normalizedStyle = 'engaging';
      }
      
      const normalizedType = type || 'article';

    // 参数验证
    if (!topic || typeof topic !== 'string' || topic.trim().length === 0) {
      return res.status(400).json({
        success: false,
        error: 'Invalid parameters',
        message: 'Topic is required and must be a non-empty string'
      });
    }

    const validTypes = ['article', 'video_script', 'short_text', 'social_media'];
    if (!validTypes.includes(normalizedType)) {
      return res.status(400).json({
        success: false,
        error: 'Invalid type',
        message: `Type must be one of: ${validTypes.join(', ')}`
      });
    }

    const validStyles = ['professional', 'engaging', 'casual', 'humorous'];
    if (!validStyles.includes(normalizedStyle)) {
      return res.status(400).json({
        success: false,
        error: 'Invalid style',
        message: `Style must be one of: ${validStyles.join(', ')}`
      });
    }

    const validLengths = ['short', 'medium', 'long'];
    if (!validLengths.includes(length)) {
      return res.status(400).json({
        success: false,
        error: 'Invalid length',
        message: `Length must be one of: ${validLengths.join(', ')}`
      });
    }

    // 模拟生成内容
    const generatedContent = generateMockContent(topic, type, style, length);

    res.json({
      success: true,
      data: generatedContent
    });
  } catch (error) {
    console.error('[ERROR] AI content generation failed:', error);
    res.status(500).json({
      success: false,
      error: 'Content generation failed',
      message: error.message
    });
  }
});

// AI内容改写
app.post('/api/v1/ai/content/rewrite', (req, res) => {
  console.log('[INFO] AI content rewrite request:', req.body);

  try {
    const { content, style = 'professional', tone = 'neutral', length = 'same' } = req.body;

    // 参数验证
    if (!content || typeof content !== 'string' || content.trim().length === 0) {
      return res.status(400).json({
        success: false,
        error: 'Invalid parameters',
        message: 'Content is required and must be a non-empty string'
      });
    }

    const validStyles = ['professional', 'engaging', 'casual', 'formal'];
    if (!validStyles.includes(normalizedStyle)) {
      return res.status(400).json({
        success: false,
        error: 'Invalid style',
        message: `Style must be one of: ${validStyles.join(', ')}`
      });
    }

    const validTones = ['neutral', 'positive', 'negative', 'enthusiastic'];
    if (!validTones.includes(tone)) {
      return res.status(400).json({
        success: false,
        error: 'Invalid tone',
        message: `Tone must be one of: ${validTones.join(', ')}`
      });
    }

    const validLengths = ['shorter', 'same', 'longer'];
    if (!validLengths.includes(length)) {
      return res.status(400).json({
        success: false,
        error: 'Invalid length',
        message: `Length must be one of: ${validLengths.join(', ')}`
      });
    }

    // 模拟改写内容
    const rewrittenContent = rewriteMockContent(content, style, tone, length);

    res.json({
      success: true,
      data: rewrittenContent
    });
  } catch (error) {
    console.error('[ERROR] AI content rewrite failed:', error);
    res.status(500).json({
      success: false,
      error: 'Content rewrite failed',
      message: error.message
    });
  }
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

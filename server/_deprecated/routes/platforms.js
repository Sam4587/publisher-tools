/**
 * 平台管理 API 代理路由
 * 将请求转发到 Go 后端服务
 */

const express = require('express');
const router = express.Router();
const http = require('http');

// Go 后端服务地址
const GO_BACKEND_HOST = process.env.GO_BACKEND_HOST || 'localhost';
const GO_BACKEND_PORT = process.env.GO_BACKEND_PORT || 8080;

// 代理请求到 Go 后端
function proxyToGoBackend(req, res, path) {
  const options = {
    hostname: GO_BACKEND_HOST,
    port: GO_BACKEND_PORT,
    path: path,
    method: req.method,
    headers: {
      'Content-Type': 'application/json',
    }
  };

  const proxyReq = http.request(options, (proxyRes) => {
    let data = '';
    
    proxyRes.on('data', (chunk) => {
      data += chunk;
    });
    
    proxyRes.on('end', () => {
      try {
        // 设置相同的响应头
        res.status(proxyRes.statusCode);
        res.setHeader('Content-Type', 'application/json');
        
        // 返回数据
        if (data) {
          res.send(data);
        } else {
          res.json({ success: true, message: 'OK' });
        }
      } catch (error) {
        console.error('[Platforms Proxy] 响应处理失败:', error);
        res.status(500).json({
          success: false,
          error: '代理响应处理失败'
        });
      }
    });
  });

  proxyReq.on('error', (error) => {
    console.error('[Platforms Proxy] 请求 Go 后端失败:', error.message);
    res.status(503).json({
      success: false,
      error: 'Go 后端服务不可用',
      details: error.message
    });
  });

  // 如果有请求体，发送到 Go 后端
  if (req.body && Object.keys(req.body).length > 0) {
    proxyReq.write(JSON.stringify(req.body));
  }
  
  proxyReq.end();
}

// GET /api/v1/platforms - 获取平台列表
router.get('/', (req, res) => {
  console.log('[Platforms Proxy] 获取平台列表');
  proxyToGoBackend(req, res, '/api/v1/platforms');
});

// GET /api/v1/platforms/:platform - 获取平台信息
router.get('/:platform', (req, res) => {
  const { platform } = req.params;
  console.log(`[Platforms Proxy] 获取平台信息: ${platform}`);
  proxyToGoBackend(req, res, `/api/v1/platforms/${platform}`);
});

// GET /api/v1/platforms/:platform/check - 检查登录状态
router.get('/:platform/check', (req, res) => {
  const { platform } = req.params;
  console.log(`[Platforms Proxy] 检查登录状态: ${platform}`);
  proxyToGoBackend(req, res, `/api/v1/platforms/${platform}/check`);
});

// POST /api/v1/platforms/:platform/login - 登录
router.post('/:platform/login', (req, res) => {
  const { platform } = req.params;
  console.log(`[Platforms Proxy] 登录请求: ${platform}`);
  proxyToGoBackend(req, res, `/api/v1/platforms/${platform}/login`);
});

// POST /api/v1/platforms/:platform/logout - 登出
router.post('/:platform/logout', (req, res) => {
  const { platform } = req.params;
  console.log(`[Platforms Proxy] 登出请求: ${platform}`);
  proxyToGoBackend(req, res, `/api/v1/platforms/${platform}/logout`);
});

module.exports = router;

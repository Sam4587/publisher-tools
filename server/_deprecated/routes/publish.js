/**
 * 发布管理 API 代理路由
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
        res.status(proxyRes.statusCode);
        res.setHeader('Content-Type', 'application/json');
        
        if (data) {
          res.send(data);
        } else {
          res.json({ success: true, message: 'OK' });
        }
      } catch (error) {
        console.error('[Publish Proxy] 响应处理失败:', error);
        res.status(500).json({
          success: false,
          error: '代理响应处理失败'
        });
      }
    });
  });

  proxyReq.on('error', (error) => {
    console.error('[Publish Proxy] 请求 Go 后端失败:', error.message);
    res.status(503).json({
      success: false,
      error: 'Go 后端服务不可用',
      details: error.message
    });
  });

  if (req.body && Object.keys(req.body).length > 0) {
    proxyReq.write(JSON.stringify(req.body));
  }
  
  proxyReq.end();
}

// POST /api/v1/publish - 同步发布
router.post('/', (req, res) => {
  console.log('[Publish Proxy] 同步发布请求');
  proxyToGoBackend(req, res, '/api/v1/publish');
});

// POST /api/v1/publish/async - 异步发布
router.post('/async', (req, res) => {
  console.log('[Publish Proxy] 异步发布请求');
  proxyToGoBackend(req, res, '/api/v1/publish/async');
});

module.exports = router;

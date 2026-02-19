/**
 * 任务管理 API 代理路由
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
        console.error('[Tasks Proxy] 响应处理失败:', error);
        res.status(500).json({
          success: false,
          error: '代理响应处理失败'
        });
      }
    });
  });

  proxyReq.on('error', (error) => {
    console.error('[Tasks Proxy] 请求 Go 后端失败:', error.message);
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

// POST /api/v1/tasks - 创建任务
router.post('/', (req, res) => {
  console.log('[Tasks Proxy] 创建任务请求');
  proxyToGoBackend(req, res, '/api/v1/tasks');
});

// GET /api/v1/tasks - 获取任务列表
router.get('/', (req, res) => {
  console.log('[Tasks Proxy] 获取任务列表');
  const query = req._parsedUrl.search || '';
  proxyToGoBackend(req, res, '/api/v1/tasks' + query);
});

// GET /api/v1/tasks/:taskId - 获取任务详情
router.get('/:taskId', (req, res) => {
  const { taskId } = req.params;
  console.log(`[Tasks Proxy] 获取任务详情: ${taskId}`);
  proxyToGoBackend(req, res, `/api/v1/tasks/${taskId}`);
});

// POST /api/v1/tasks/:taskId/cancel - 取消任务
router.post('/:taskId/cancel', (req, res) => {
  const { taskId } = req.params;
  console.log(`[Tasks Proxy] 取消任务: ${taskId}`);
  proxyToGoBackend(req, res, `/api/v1/tasks/${taskId}/cancel`);
});

module.exports = router;

const express = require('express');
const cors = require('cors');
const helmet = require('helmet');
const compression = require('compression');

const app = express();
const PORT = process.env.PORT || 3001;

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

app.get('/api/health', (req, res) => {
  res.json({
    success: true,
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
    message: 'Node backend is running'
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

app.get('/api/tasks', (req, res) => {
  res.json({ success: true, tasks: [] });
});

app.post('/api/publish', (req, res) => {
  res.json({ success: true, message: 'Publish endpoint - connect to Go backend for full functionality' });
});

app.listen(PORT, () => {
  console.log(`Node backend running on port ${PORT}`);
  console.log(`Health check: http://localhost:${PORT}/api/health`);
});

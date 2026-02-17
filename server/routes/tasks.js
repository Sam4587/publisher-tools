/**
 * 任务管理API路由
 */

const express = require('express');
const router = express.Router();

// 系统统计信息（管理员接口）
router.get('/stats/system', async (req, res) => {
  try {
    // 模拟统计数据
    const stats = {
      runningTasks: 0,
      registeredWorkers: 4,
      activeTaskIds: []
    };
    
    res.json({
      success: true,
      data: stats
    });
  } catch (error) {
    console.error('[Tasks] 获取系统统计失败', { error: error.message });
    res.status(500).json({
      success: false,
      error: '获取系统统计失败'
    });
  }
});

module.exports = router;
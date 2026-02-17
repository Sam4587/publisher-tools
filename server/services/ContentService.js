/**
 * 内容管理服务 - 简化稳定版本
 * 临时修复语法错误，确保服务稳定运行
 */

const logger = require('../utils/logger');
const publishIntegration = require('./PublishIntegration');

// 模拟数据存储（内存）
const memoryStorage = {
  contents: []
};

class ContentService {
  constructor() {
    this.publishIntegration = publishIntegration;
  }

  // 简化方法，避免语法错误
  async create(contentData, userId) {
    try {
      logger.warn('[ContentService] 服务已简化，使用内存存储');
      const content = {
        _id: `content_${Date.now()}`,
        title: contentData.title || '未命名内容',
        content: contentData.content || '',
        userId,
        createdAt: new Date(),
        status: 'DRAFT'
      };
      
      memoryStorage.contents.push(content);
      
      return {
        success: true,
        content: content
      };
    } catch (error) {
      logger.error('[ContentService] 创建内容失败', { error: error.message });
      return { success: false, error: error.message };
    }
  }

  async getById(contentId) {
    try {
      const content = memoryStorage.contents.find(c => c._id === contentId);
      return content ? { success: true, content } : { success: false, error: '内容不存在' };
    } catch (error) {
      logger.error('[ContentService] 获取内容详情失败', { error: error.message });
      return { success: false, error: error.message };
    }
  }

  async update(contentId, updateData, userId) {
    try {
      const index = memoryStorage.contents.findIndex(c => c._id === contentId);
      if (index === -1) {
        return { success: false, error: '内容不存在' };
      }
      
      Object.assign(memoryStorage.contents[index], updateData);
      return { success: true, content: memoryStorage.contents[index] };
    } catch (error) {
      logger.error('[ContentService] 更新内容失败', { error: error.message });
      return { success: false, error: error.message };
    }
  }

  async delete(contentId, userId) {
    try {
      const index = memoryStorage.contents.findIndex(c => c._id === contentId);
      if (index === -1) {
        return { success: false, error: '内容不存在' };
      }
      
      memoryStorage.contents.splice(index, 1);
      return { success: true, message: '内容已删除' };
    } catch (error) {
      logger.error('[ContentService] 删除内容失败', { error: error.message });
      return { success: false, error: error.message };
    }
  }

  async list(filters = {}, page = 1, limit = 20) {
    try {
      let results = [...memoryStorage.contents];
      
      // 简单过滤
      if (filters.userId) {
        results = results.filter(c => c.userId === filters.userId);
      }
      
      const skip = (page - 1) * limit;
      const data = results.slice(skip, skip + limit);
      
      return {
        success: true,
        data,
        pagination: {
          page: parseInt(page),
          limit: parseInt(limit),
          total: results.length,
          pages: Math.ceil(results.length / limit)
        }
      };
    } catch (error) {
      logger.error('[ContentService] 查询内容列表失败', { error: error.message });
      return { success: false, error: error.message };
    }
  }

  /**
   * 生成文章内容
   */
  async generateArticle(topic, options = {}) {
    try {
      const { style = 'professional', length = 'medium' } = options;
      
      // 模拟AI生成内容
      const content = this.simulateContentGeneration(topic, 'article', style, length);
      
      const article = {
        title: `【${style}】${topic}`,
        content: content,
        summary: content.substring(0, 200) + '...',
        wordCount: content.split(' ').length,
        readingTime: Math.ceil(content.split(' ').length / 200), // 假设每分钟200词
        type: 'article',
        style: style,
        length: length
      };
      
      return article;
    } catch (error) {
      logger.error('[ContentService] 生成文章失败', { error: error.message });
      throw error;
    }
  }

  /**
   * 生成视频脚本
   */
  async generateVideoScript(topic, options = {}) {
    try {
      const { style = 'engaging', length = 'medium' } = options;
      
      // 模拟AI生成视频脚本
      const script = this.simulateContentGeneration(topic, 'video_script', style, length);
      
      const videoScript = {
        title: `【视频】${topic}`,
        content: script,
        summary: script.substring(0, 150) + '...',
        wordCount: script.split(' ').length,
        estimatedDuration: Math.ceil(script.split(' ').length / 150), // 假设每分钟150词
        type: 'video_script',
        style: style,
        length: length
      };
      
      return videoScript;
    } catch (error) {
      logger.error('[ContentService] 生成视频脚本失败', { error: error.message });
      throw error;
    }
  }

  /**
   * 通用内容生成方法
   */
  async generateContent(topic, type, options = {}) {
    try {
      const { style = 'professional' } = options;
      
      const content = this.simulateContentGeneration(topic, type, style, 'medium');
      
      return {
        title: `【${type}】${topic}`,
        content: content,
        summary: content.substring(0, 100) + '...',
        wordCount: content.split(' ').length,
        type: type,
        style: style
      };
    } catch (error) {
      logger.error('[ContentService] 生成内容失败', { error: error.message });
      throw error;
    }
  }

  /**
   * 模拟内容生成（实际项目中这里会调用AI服务）
   */
  simulateContentGeneration(topic, type, style, length) {
    const lengthMap = {
      'short': 200,
      'medium': 500,
      'long': 1000
    };
    
    const wordCount = lengthMap[length] || 500;
    
    const templates = {
      'professional': `关于"${topic}"的专业分析。本文将从多个维度深入探讨这一话题，为您提供有价值的见解和实用建议。

首先，我们需要了解...（此处省略具体内容）

其次，从实践角度来看...（此处省略具体内容）

最后，总结来说...（此处省略具体内容）

希望通过本文的分享，能够帮助您更好地理解和应用相关知识。`,
      'engaging': `你有没有想过"${topic}"到底是什么？今天就让我们一起来探索这个有趣的话题！

想象一下...（此处省略具体内容）

那么问题来了...（此处省略具体内容）

总的来说...（此处省略具体内容）

是不是觉得很有趣呢？欢迎在评论区分享你的想法！`,
      'casual': `最近"${topic}"这个话题挺火的，我也来聊聊我的看法。

说实话...（此处省略具体内容）

我觉得吧...（此处省略具体内容）

反正就是这么回事，大家觉得呢？`
    };
    
    return templates[style] || templates.professional;
  }
}

// 创建单例
const contentService = new ContentService();
module.exports = contentService;
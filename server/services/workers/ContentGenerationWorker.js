/**
 * 内容生成任务处理器
 * 处理文章、视频脚本等内容生成任务
 */

const ContentService = require('../../core/ContentService');
const logger = require('../../../utils/logger');

class ContentGenerationWorker {
  constructor() {
    this.contentService = new ContentService();
  }

  /**
   * 执行内容生成任务
   * @param {Object} task - 任务对象
   * @param {Function} progressCallback - 进度回调函数
   */
  async execute(task, progressCallback) {
    try {
      const { topic, type, style, length } = task.parameters;
      
      // 验证参数
      if (!topic) {
        throw new Error('缺少主题参数');
      }

      // 更新进度：开始处理
      await progressCallback(10, '开始生成内容...', 'initializing');

      // 生成内容
      await progressCallback(30, '正在生成内容...', 'generating');
      
      let content;
      if (type === 'article') {
        content = await this.contentService.generateArticle(topic, {
          style: style || 'professional',
          length: length || 'medium'
        });
      } else if (type === 'video_script') {
        content = await this.contentService.generateVideoScript(topic, {
          style: style || 'engaging',
          length: length || 'medium'
        });
      } else {
        content = await this.contentService.generateContent(topic, type, {
          style: style || 'professional'
        });
      }

      // 更新进度：处理完成
      await progressCallback(90, '内容生成完成，正在保存...', 'saving');

      // 保存结果
      const result = {
        content: content.content,
        title: content.title,
        summary: content.summary,
        wordCount: content.wordCount,
        readingTime: content.readingTime,
        generatedAt: new Date().toISOString()
      };

      await progressCallback(100, '任务完成', 'completed');

      return result;

    } catch (error) {
      logger.error('[ContentGenerationWorker] 任务执行失败', {
        taskId: task.taskId,
        error: error.message,
        stack: error.stack
      });
      throw error;
    }
  }
}

module.exports = ContentGenerationWorker;
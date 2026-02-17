/**
 * 任务模型
 * 用于跟踪内容生成、转录等长时间运行的任务
 * 
 * 注意：此项目使用内存存储，不依赖 MongoDB
 * 此文件为兼容性保留，实际使用 publisher-core/task 模块
 */

// Mock 模型 - 实际任务管理使用 publisher-core/task 模块
const TaskModel = {
  create: async (data) => {
    console.log('[Mock Task Model] create:', data);
    return { id: data.taskId || Date.now().toString(), ...data };
  },
  
  findById: async (id) => {
    console.log('[Mock Task Model] findById:', id);
    return null;
  },
  
  findOne: async (query) => {
    console.log('[Mock Task Model] findOne:', query);
    return null;
  },
  
  updateOne: async (id, data) => {
    console.log('[Mock Task Model] updateOne:', id, data);
    return { modifiedCount: 1 };
  },
  
  deleteOne: async (id) => {
    console.log('[Mock Task Model] deleteOne:', id);
    return { deletedCount: 1 };
  },
  
  aggregate: async () => []
};

module.exports = TaskModel;

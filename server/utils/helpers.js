/**
 * 通用辅助函数
 */

/**
 * 计算阅读时间
 * @param {string} content 内容文本
 * @param {number} wordsPerMinute 每分钟阅读字数，默认300
 * @returns {number} 阅读时间（分钟）
 */
function calculateReadingTime(content, wordsPerMinute = 300) {
  if (!content || typeof content !== 'string') {
    return 0;
  }
  
  // 移除HTML标签和多余空格
  const cleanContent = content.replace(/<[^>]*>/g, '').trim();
  const chineseChars = cleanContent.match(/[\u4e00-\u9fa5]/g);
  const englishWords = cleanContent.replace(/[\u4e00-\u9fa5]/g, '').match(/\S+/g);
  
  let totalWords = 0;
  
  // 中文字符每个算1个词
  if (chineseChars) {
    totalWords += chineseChars.length;
  }
  
  // 英文单词每个算1个词
  if (englishWords) {
    totalWords += englishWords.length;
  }
  
  // 计算阅读时间（向上取整）
  return Math.ceil(totalWords / wordsPerMinute);
}

/**
 * 清理HTML内容
 * @param {string} html HTML内容
 * @param {Object} options 清理选项
 * @returns {string} 清理后的HTML
 */
function sanitizeHtml(html, options = {}) {
  if (!html || typeof html !== 'string') {
    return '';
  }
  
  let cleanHtml = html;
  
  // 默认选项
  const defaultOptions = {
    allowTags: ['p', 'br', 'strong', 'em', 'u', 'ol', 'ul', 'li', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6'],
    allowAttributes: ['href', 'src', 'alt', 'title'],
    stripScripts: true,
    stripStyles: true
  };
  
  const opts = { ...defaultOptions, ...options };
  
  // 移除script标签
  if (opts.stripScripts) {
    cleanHtml = cleanHtml.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '');
  }
  
  // 移除style标签
  if (opts.stripStyles) {
    cleanHtml = cleanHtml.replace(/<style\b[^<]*(?:(?!<\/style>)<[^<]*)*<\/style>/gi, '');
  }
  
  // 只保留允许的标签
  if (opts.allowTags && opts.allowTags.length > 0) {
    const allowedTags = opts.allowTags.join('|');
    const tagPattern = new RegExp(`<(/?)(${allowedTags})([^>]*)>`, 'gi');
    cleanHtml = cleanHtml.replace(tagPattern, '<$1$2$3>');
    
    // 移除不允许的标签
    cleanHtml = cleanHtml.replace(/<\/?([a-z][a-z0-9]*)[^>]*>/gi, (match, tagName) => {
      if (opts.allowTags.includes(tagName.toLowerCase())) {
        return match;
      }
      return '';
    });
  }
  
  // 只保留允许的属性
  if (opts.allowAttributes && opts.allowAttributes.length > 0) {
    const attrPattern = new RegExp(`\\s*(${opts.allowAttributes.join('|')})\\s*=\\s*("[^"]*"|'[^']*'|[^\\s>]*)`, 'gi');
    cleanHtml = cleanHtml.replace(/<([a-z][a-z0-9]*)[^>]*>/gi, (match, tagName) => {
      if (!opts.allowTags || opts.allowTags.includes(tagName.toLowerCase())) {
        const allowedAttrs = match.match(attrPattern) || [];
        return `<${tagName}${allowedAttrs.join('')}>`;
      }
      return match;
    });
  }
  
  // 清理多余的空白字符
  cleanHtml = cleanHtml.replace(/\s+/g, ' ').trim();
  
  return cleanHtml;
}

/**
 * 生成随机ID
 * @param {number} length ID长度
 * @returns {string} 随机ID
 */
function generateId(length = 8) {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

/**
 * 深度合并对象
 * @param {...Object} objects 要合并的对象
 * @returns {Object} 合并后的对象
 */
function deepMerge(...objects) {
  const result = {};
  
  objects.forEach(obj => {
    if (obj && typeof obj === 'object') {
      Object.keys(obj).forEach(key => {
        if (obj[key] && typeof obj[key] === 'object' && !Array.isArray(obj[key])) {
          result[key] = deepMerge(result[key] || {}, obj[key]);
        } else {
          result[key] = obj[key];
        }
      });
    }
  });
  
  return result;
}

/**
 * 格式化日期
 * @param {Date|string|number} date 日期
 * @param {string} format 格式字符串
 * @returns {string} 格式化后的日期
 */
function formatDate(date, format = 'YYYY-MM-DD HH:mm:ss') {
  const d = new Date(date);
  
  if (isNaN(d.getTime())) {
    return '';
  }
  
  const pad = (num) => num.toString().padStart(2, '0');
  
  const replacements = {
    'YYYY': d.getFullYear(),
    'MM': pad(d.getMonth() + 1),
    'DD': pad(d.getDate()),
    'HH': pad(d.getHours()),
    'mm': pad(d.getMinutes()),
    'ss': pad(d.getSeconds())
  };
  
  let formatted = format;
  Object.keys(replacements).forEach(key => {
    formatted = formatted.replace(new RegExp(key, 'g'), replacements[key]);
  });
  
  return formatted;
}

/**
 * 防抖函数
 * @param {Function} func 要防抖的函数
 * @param {number} delay 延迟时间（毫秒）
 * @returns {Function} 防抖后的函数
 */
function debounce(func, delay) {
  let timeoutId;
  return function (...args) {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => func.apply(this, args), delay);
  };
}

/**
 * 节流函数
 * @param {Function} func 要节流的函数
 * @param {number} delay 延迟时间（毫秒）
 * @returns {Function} 节流后的函数
 */
function throttle(func, delay) {
  let lastCall = 0;
  return function (...args) {
    const now = Date.now();
    if (now - lastCall >= delay) {
      lastCall = now;
      func.apply(this, args);
    }
  };
}

module.exports = {
  calculateReadingTime,
  sanitizeHtml,
  generateId,
  deepMerge,
  formatDate,
  debounce,
  throttle
};
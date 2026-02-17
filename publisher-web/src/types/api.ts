// API 类型定义

// 平台类型
export type Platform = 'douyin' | 'toutiao' | 'xiaohongshu'

// 平台信息
export interface PlatformInfo {
  name: string
  display_name: string
  icon?: string
  limits: {
    title_max_length: number
    body_max_length: number
    max_images: number
    max_video_size: number
    allowed_video_formats: string[]
    allowed_image_formats: string[]
  }
}

// 登录结果
export interface LoginResult {
  success: boolean
  qrcode_url?: string
  error?: string
  expires_at?: string
}

// 发布状态
export type PublishStatus = 'pending' | 'processing' | 'success' | 'failed'

// 发布内容类型
export type ContentType = 'text' | 'images' | 'video' | 'article'

// 发布内容
export interface PublishContent {
  platform: Platform
  type: ContentType
  title: string
  body: string
  images?: string[]
  video?: string
  tags?: string[]
}

// 发布结果
export interface PublishResult {
  task_id: string
  status: PublishStatus
  platform: string
  post_id?: string
  post_url?: string
  error?: string
  created_at: string
  finished_at?: string
}

// 任务
export interface Task {
  id: string
  type: string
  status: PublishStatus
  platform: string
  payload: Record<string, unknown>
  result?: Record<string, unknown>
  error?: string
  progress: number
  created_at: string
  started_at?: string
  finished_at?: string
}

// 账号状态
export interface AccountStatus {
  platform: Platform
  logged_in: boolean
  account_name?: string
  avatar?: string
  last_check?: string
}

// API 响应
export interface APIResponse<T = unknown> {
  success: boolean
  data?: T
  error?: string
  error_code?: string
  timestamp: number
}

// =====================================================
// 热点监控相关类型
// =====================================================

// 热点话题分类
export type HotTopicCategory = '娱乐' | '科技' | '财经' | '体育' | '社会' | '国际' | '其他'

// 热点趋势
export type HotTopicTrend = 'up' | 'down' | 'stable' | 'new' | 'hot'

// 热点话题
export interface HotTopic {
  _id: string
  title: string
  description?: string
  category: HotTopicCategory
  heat: number
  trend: HotTopicTrend
  source: string
  sourceId?: string
  sourceUrl?: string
  originalUrl?: string
  keywords?: string[]
  suitability?: number
  publishedAt?: string
  createdAt?: string
  updatedAt?: string
  extra?: {
    hotValue?: number | null
    originTitle?: string | null
  }
}

// 数据源
export interface HotSource {
  id: string
  name: string
  enabled: boolean
}

// 分页信息
export interface Pagination {
  page: number
  limit: number
  total: number
  pages: number
}

// 热点列表响应
export interface HotTopicsResponse {
  success: boolean
  data: HotTopic[]
  pagination: Pagination
}

// 趋势数据点
export interface TrendDataPoint {
  date: string
  heat: number
  rank?: number
}

// 跨平台分析结果
export interface CrossPlatformAnalysis {
  title: string
  platforms: {
    name: string
    heat: number
    rank: number
    url?: string
  }[]
  totalHeat: number
  trend: HotTopicTrend
}

// AI 分析结果
export interface AIAnalysisResult {
  summary: string
  keyPoints: string[]
  sentiment: 'positive' | 'negative' | 'neutral'
  recommendations?: string[]
}

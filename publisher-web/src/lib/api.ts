import type { Platform, PlatformInfo, LoginResult, PublishResult, Task, AccountStatus, APIResponse, PublishContent, HotTopic, HotSource, Pagination, CrossPlatformAnalysis, AIAnalysisResult } from '@/types/api'

const API_BASE = 'http://localhost:3001/api'
const HOT_API_BASE = '/api/hot-topics'

// 平台列表响应类型
interface PlatformsResponse {
  count: number
  platforms: string[]
}

// 通用请求方法
async function request<T>(url: string, options?: RequestInit): Promise<APIResponse<T>> {
  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    ...options,
  })

  const data = await response.json()
  return data as APIResponse<T>
}

// 获取平台列表
export async function getPlatforms(): Promise<APIResponse<PlatformsResponse>> {
  return request<PlatformsResponse>(`${API_BASE}/platforms`)
}

// 获取平台信息
export async function getPlatformInfo(platform: Platform): Promise<APIResponse<PlatformInfo>> {
  return request<PlatformInfo>(`${API_BASE}/platforms/${platform}`)
}

// 登录
export async function login(platform: Platform): Promise<APIResponse<LoginResult>> {
  return request<LoginResult>(`${API_BASE}/platforms/${platform}/login`, {
    method: 'POST',
  })
}

// 检查登录状态
export async function checkLogin(platform: Platform): Promise<APIResponse<AccountStatus>> {
  return request<AccountStatus>(`${API_BASE}/platforms/${platform}/check`)
}

// 发布内容

export async function publish(content: PublishContent): Promise<APIResponse<PublishResult>> {
  return request<PublishResult>(`${API_BASE}/publish`, {
    method: 'POST',
    body: JSON.stringify(content),
  })
}

// 异步发布
export async function publishAsync(content: PublishContent): Promise<APIResponse<{ task_id: string }>> {
  return request<{ task_id: string }>(`${API_BASE}/publish/async`, {
    method: 'POST',
    body: JSON.stringify(content),
  })
}

// 获取任务列表
export async function getTasks(status?: string, platform?: string, limit = 20): Promise<APIResponse<Task[]>> {
  const params = new URLSearchParams()
  if (status) params.set('status', status)
  if (platform) params.set('platform', platform)
  params.set('limit', limit.toString())

  return request<Task[]>(`${API_BASE}/tasks?${params}`)
}

// 获取任务详情
export async function getTask(taskId: string): Promise<APIResponse<Task>> {
  return request<Task>(`${API_BASE}/tasks/${taskId}`)
}

// 取消任务
export async function cancelTask(taskId: string): Promise<APIResponse<void>> {
  return request<void>(`${API_BASE}/tasks/${taskId}/cancel`, {
    method: 'POST',
  })
}

// =====================================================
// 热点监控 API
// =====================================================

// 热点列表查询参数
export interface HotTopicsParams {
  page?: number
  limit?: number
  category?: string
  search?: string
  minHeat?: number
  maxHeat?: number
  sortBy?: 'heat' | 'createdAt' | 'publishedAt' | 'suitability'
  sortOrder?: 'asc' | 'desc'
}

// 获取热点列表
export async function getHotTopics(params: HotTopicsParams = {}): Promise<{ success: boolean; data: HotTopic[]; pagination: Pagination }> {
  const query = new URLSearchParams()
  if (params.page) query.set('page', params.page.toString())
  if (params.limit) query.set('limit', params.limit.toString())
  if (params.category) query.set('category', params.category)
  if (params.search) query.set('search', params.search)
  if (params.minHeat !== undefined) query.set('minHeat', params.minHeat.toString())
  if (params.maxHeat !== undefined) query.set('maxHeat', params.maxHeat.toString())
  if (params.sortBy) query.set('sortBy', params.sortBy)
  if (params.sortOrder) query.set('sortOrder', params.sortOrder)

  const response = await fetch(`${HOT_API_BASE}?${query}`)
  return response.json()
}

// 获取热点详情
export async function getHotTopic(id: string): Promise<{ success: boolean; data?: HotTopic; message?: string }> {
  const response = await fetch(`${HOT_API_BASE}/${id}`)
  return response.json()
}

// 获取数据源列表
export async function getHotSources(): Promise<{ success: boolean; data: HotSource[] }> {
  const response = await fetch(`${HOT_API_BASE}/newsnow/sources`)
  return response.json()
}

// 从 NewsNow 抓取热点
export async function fetchHotTopics(sources?: string[], maxItems = 20): Promise<{ success: boolean; data: { fetched: number; saved: number; topics: HotTopic[] } }> {
  const response = await fetch(`${HOT_API_BASE}/newsnow/fetch`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ sources, maxItems }),
  })
  return response.json()
}

// 从指定数据源抓取热点
export async function fetchHotTopicsFromSource(sourceId: string, maxItems = 20): Promise<{ success: boolean; data: { source: string; sourceName: string; count: number; topics: HotTopic[] } }> {
  const response = await fetch(`${HOT_API_BASE}/newsnow/fetch/${sourceId}?maxItems=${maxItems}`)
  return response.json()
}

// 刷新热点数据
export async function refreshHotTopics(): Promise<{ success: boolean; message: string; data?: { count: number; topics: HotTopic[] } }> {
  const response = await fetch(`${HOT_API_BASE}/update`, { method: 'POST' })
  return response.json()
}

// 获取新增热点
export async function getNewHotTopics(hours = 24): Promise<{ success: boolean; data: HotTopic[] }> {
  const response = await fetch(`${HOT_API_BASE}/trends/new?hours=${hours}`)
  return response.json()
}

// 获取热点趋势
export async function getHotTopicTrend(id: string, days = 7): Promise<{ success: boolean; data: { topic: HotTopic; trend: { date: string; heat: number; rank: number }[] } }> {
  const response = await fetch(`${HOT_API_BASE}/trends/timeline/${id}?days=${days}`)
  return response.json()
}

// 获取跨平台分析
export async function getCrossPlatformAnalysis(title: string): Promise<{ success: boolean; data: CrossPlatformAnalysis }> {
  const response = await fetch(`${HOT_API_BASE}/trends/cross-platform/${encodeURIComponent(title)}`)
  return response.json()
}

// AI 分析热点
export async function analyzeHotTopics(topics: HotTopic[], options?: { provider?: string; focus?: string }): Promise<{ success: boolean; data?: AIAnalysisResult; message?: string }> {
  const response = await fetch(`${HOT_API_BASE}/ai/analyze`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ topics, options }),
  })
  return response.json()
}

// 生成热点简报
export async function generateHotTopicsBrief(topics: HotTopic[], maxLength = 300): Promise<{ success: boolean; data?: { brief: string }; message?: string }> {
  const response = await fetch(`${HOT_API_BASE}/ai/briefing`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ topics, maxLength }),
  })
  return response.json()
}

// =====================================================
// 视频下载 API
// =====================================================

const VIDEO_API_BASE = '/api/video'
const TRANSCRIPTION_API_BASE = '/api/transcription'

// 视频信息类型
export interface VideoInfo {
  videoId: string
  platform: string
  title: string
  author: string
  duration: number
  cover: string
  localPath: string
  fileSize: number
  status: 'downloaded' | 'uploaded' | 'transcribing' | 'transcribed'
  transcription?: TranscriptResult
  createdAt: string
}

// 转录结果类型
export interface TranscriptResult {
  success: boolean
  engine: string
  duration: number
  language: string
  text: string
  segments: TranscriptSegment[]
  keywords: string[]
  metadata: {
    modelSize?: string
    processingTime: number
  }
}

// 登出
export async function logout(platform: Platform): Promise<APIResponse<{ platform: string; message: string }>> {
  return request<{ platform: string; message: string }>(`${API_BASE}/platforms/${platform}/logout`, {
    method: 'POST',
  })
}


// 转录片段
export interface TranscriptSegment {
  index: number
  start: number
  end: number
  text: string
  confidence: number
}

// 转录任务
export interface TranscriptionTask {
  taskId: string
  videoId: string
  status: 'pending' | 'processing' | 'completed' | 'failed'
  progress: number
  result: TranscriptResult | null
  error: string | null
  createdAt: string
  updatedAt: string
}

// 下载视频
export async function downloadVideo(url: string, removeWatermark = false): Promise<{ success: boolean; data?: { videoId: string; status: string }; message?: string }> {
  const response = await fetch(`${VIDEO_API_BASE}/download`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url, removeWatermark }),
  })
  return response.json()
}

// 获取视频状态
export async function getVideoStatus(videoId: string): Promise<{ success: boolean; data?: VideoInfo; message?: string }> {
  const response = await fetch(`${VIDEO_API_BASE}/download/${videoId}/status`)
  return response.json()
}

// 获取视频列表
export async function getVideoList(params?: { platform?: string; status?: string; page?: number; pageSize?: number }): Promise<{ success: boolean; data?: { total: number; data: VideoInfo[] }; message?: string }> {
  const query = new URLSearchParams()
  if (params?.platform) query.set('platform', params.platform)
  if (params?.status) query.set('status', params.status)
  if (params?.page) query.set('page', params.page.toString())
  if (params?.pageSize) query.set('pageSize', params.pageSize.toString())

  const response = await fetch(`${VIDEO_API_BASE}/download/list?${query}`)
  return response.json()
}

// 删除视频
export async function deleteVideo(videoId: string): Promise<{ success: boolean; message?: string }> {
  const response = await fetch(`${VIDEO_API_BASE}/download/${videoId}`, { method: 'DELETE' })
  return response.json()
}

// 识别视频平台
export async function identifyVideoPlatform(url: string): Promise<{ success: boolean; data?: { platform: string; videoId?: string }; message?: string }> {
  const response = await fetch(`${VIDEO_API_BASE}/metadata`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ url }),
  })
  return response.json()
}

// 提交转录任务
export async function submitTranscription(videoId: string, engine?: string): Promise<{ success: boolean; data?: { taskId: string; status: string }; message?: string }> {
  const response = await fetch(`${TRANSCRIPTION_API_BASE}/submit`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ videoId, engine }),
  })
  return response.json()
}

// 获取转录任务状态
export async function getTranscriptionTask(taskId: string): Promise<{ success: boolean; data?: TranscriptionTask; message?: string }> {
  const response = await fetch(`${TRANSCRIPTION_API_BASE}/${taskId}`)
  return response.json()
}

// 获取视频的转录结果
export async function getVideoTranscription(videoId: string): Promise<{ success: boolean; data?: { status: string; transcription?: TranscriptResult }; message?: string }> {
  const response = await fetch(`${TRANSCRIPTION_API_BASE}/video/${videoId}`)
  return response.json()
}

// 获取可用转录引擎
export async function getTranscriptionEngines(): Promise<{ success: boolean; data?: { name: string; enabled: boolean }[] }> {
  const response = await fetch(`${TRANSCRIPTION_API_BASE}/engines/list`)
  return response.json()
}

// =====================================================
// AI 服务 API
// =====================================================

// AI 消息类型
export interface AIMessage {
  role: 'system' | 'user' | 'assistant'
  content: string
}

// AI 生成选项
export interface AIGenerateOptions {
  messages: AIMessage[]
  model?: string
  max_tokens?: number
  temperature?: number
}

// AI 生成结果
export interface AIGenerateResult {
  content: string
  model: string
  provider: string
  input_tokens: number
  output_tokens: number
}

// 获取 AI 提供商列表
export async function getAIProviders(): Promise<APIResponse<{ providers: string[]; count: number }>> {
  return request(`${API_BASE}/ai/providers`)
}

// 获取 AI 模型列表
export async function getAIModels(): Promise<APIResponse<Record<string, string[]>>> {
  return request(`${API_BASE}/ai/models`)
}

// AI 生成（使用默认提供商）
export async function aiGenerate(options: AIGenerateOptions): Promise<APIResponse<AIGenerateResult>> {
  return request<AIGenerateResult>(`${API_BASE}/ai/generate`, {
    method: 'POST',
    body: JSON.stringify(options),
  })
}

// AI 生成（指定提供商）
export async function aiGenerateWithProvider(provider: string, options: AIGenerateOptions): Promise<APIResponse<AIGenerateResult>> {
  return request<AIGenerateResult>(`${API_BASE}/ai/generate/${provider}`, {
    method: 'POST',
    body: JSON.stringify(options),
  })
}

// 热点分析请求
export interface HotspotAnalyzeRequest {
  title: string
  content: string
}

// 热点分析结果
export interface HotspotAnalyzeResult {
  analysis: string
  provider: string
  model: string
}

// AI 分析热点
export async function aiAnalyzeHotspot(req: HotspotAnalyzeRequest): Promise<APIResponse<HotspotAnalyzeResult>> {
  return request<HotspotAnalyzeResult>(`${API_BASE}/ai/analyze/hotspot`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

// 内容生成请求
export interface ContentGenerateRequest {
  topic: string
  platform?: string
  style?: string
  length?: number
}

// 内容生成结果
export interface ContentGenerateResult {
  content: string
  provider: string
  model: string
}

// AI 生成内容
export async function aiContentGenerate(req: ContentGenerateRequest): Promise<APIResponse<ContentGenerateResult>> {
  return request<ContentGenerateResult>(`${API_BASE}/ai/content/generate`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

// 内容改写请求
export interface ContentRewriteRequest {
  content: string
  style?: string
  platform?: string
}

// AI 改写内容
export async function aiContentRewrite(req: ContentRewriteRequest): Promise<APIResponse<ContentGenerateResult>> {
  return request<ContentGenerateResult>(`${API_BASE}/ai/content/rewrite`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

// 内容审核请求
export interface ContentAuditRequest {
  content: string
}

// 内容审核结果
export interface ContentAuditResult {
  audit_result: string
  provider: string
  model: string
}

// AI 审核内容
export async function aiContentAudit(req: ContentAuditRequest): Promise<APIResponse<ContentAuditResult>> {
  return request<ContentAuditResult>(`${API_BASE}/ai/content/audit`, {
    method: 'POST',
    body: JSON.stringify(req),
  })
}



import { useState, useCallback } from 'react'
import { Mic, Upload, Settings, Play, Pause, RefreshCw, Check, AlertCircle, Cloud, Cpu, ChevronDown, ChevronUp } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'

// ASR提供商类型
type ASRProvider = 'auto' | 'bcut-asr' | 'whisper'

// 识别结果接口
interface RecognitionResult {
  provider: ASRProvider
  language: string
  text: string
  segments: RecognitionSegment[]
  duration: number
  audio_duration: number
  word_count: number
  quality_score: number
  cached: boolean
  metadata?: Record<string, unknown>
}

// 识别片段接口
interface RecognitionSegment {
  id: number
  start: number
  end: number
  text: string
  confidence?: number
  words?: Word[]
}

// 词级时间戳接口
interface Word {
  word: string
  start: number
  end: number
  confidence?: number
}

// 提供商信息接口
interface ProviderInfo {
  name: ASRProvider
  available: boolean
  priority: number
  max_file_size: number
  max_duration: number
}

// 统计信息接口
interface ASRStats {
  total_requests: number
  cache_hits: number
  cache_misses: number
  fallback_count: number
  provider_stats: Record<string, {
    requests: number
    successes: number
    failures: number
    avg_time: number
  }>
}

// ASR面板属性
interface ASRPanelProps {
  onRecognitionComplete?: (result: RecognitionResult) => void
  audioPath?: string
}

// 支持的语言
const SUPPORTED_LANGUAGES: Record<string, string> = {
  'auto': '自动检测',
  'zh': '中文',
  'en': '英语',
  'ja': '日语',
  'ko': '韩语',
  'fr': '法语',
  'de': '德语',
  'es': '西班牙语',
  'ru': '俄语',
  'pt': '葡萄牙语',
  'it': '意大利语',
  'ar': '阿拉伯语',
}

// Whisper模型
const WHISPER_MODELS = [
  { name: 'tiny', label: 'Tiny (最快)', params: '39M', speed: '~32x' },
  { name: 'base', label: 'Base (推荐)', params: '74M', speed: '~16x' },
  { name: 'small', label: 'Small', params: '244M', speed: '~6x' },
  { name: 'medium', label: 'Medium', params: '769M', speed: '~2x' },
  { name: 'large', label: 'Large (最准)', params: '1550M', speed: '~1x' },
]

export default function ASRPanel({ onRecognitionComplete, audioPath: initialAudioPath }: ASRPanelProps) {
  // 状态
  const [audioPath, setAudioPath] = useState(initialAudioPath || '')
  const [language, setLanguage] = useState('auto')
  const [model, setModel] = useState('base')
  const [provider, setProvider] = useState<ASRProvider>('auto')
  const [enableTimestamps, setEnableTimestamps] = useState(true)
  const [enableWordLevel, setEnableWordLevel] = useState(false)
  const [enableGPU, setEnableGPU] = useState(false)
  const [useChunking, setUseChunking] = useState(true)
  
  // 识别状态
  const [isRecognizing, setIsRecognizing] = useState(false)
  const [progress, setProgress] = useState(0)
  const [result, setResult] = useState<RecognitionResult | null>(null)
  const [error, setError] = useState<string | null>(null)
  
  // 提供商信息
  const [providers, setProviders] = useState<ProviderInfo[]>([])
  const [stats, setStats] = useState<ASRStats | null>(null)
  
  // 高级设置展开
  const [showAdvanced, setShowAdvanced] = useState(false)

  // 获取提供商列表
  const fetchProviders = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/asr/providers')
      const data = await response.json()
      if (data.success) {
        setProviders(data.data)
      }
    } catch (err) {
      console.error('Failed to fetch providers:', err)
    }
  }, [])

  // 获取统计信息
  const fetchStats = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/asr/stats')
      const data = await response.json()
      if (data.success) {
        setStats(data.data)
      }
    } catch (err) {
      console.error('Failed to fetch stats:', err)
    }
  }, [])

  // 执行识别
  const handleRecognize = async () => {
    if (!audioPath) {
      setError('请输入音频文件路径')
      return
    }

    setIsRecognizing(true)
    setError(null)
    setProgress(0)
    setResult(null)

    // 模拟进度更新
    const progressInterval = setInterval(() => {
      setProgress(prev => Math.min(prev + 5, 90))
    }, 500)

    try {
      const endpoint = useChunking ? '/api/v1/asr/recognize/chunked' : '/api/v1/asr/recognize'
      
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          audio_path: audioPath,
          language,
          model,
          enable_timestamps: enableTimestamps,
          enable_word_level: enableWordLevel,
          enable_gpu: enableGPU,
          timeout: 1800,
        }),
      })

      clearInterval(progressInterval)
      setProgress(100)

      const data = await response.json()

      if (data.success) {
        setResult(data.data)
        onRecognitionComplete?.(data.data)
      } else {
        setError(data.error?.message || '识别失败')
      }
    } catch (err) {
      clearInterval(progressInterval)
      setError(err instanceof Error ? err.message : '识别失败')
    } finally {
      setIsRecognizing(false)
    }
  }

  // 使用指定提供商识别
  const handleRecognizeWithProvider = async (providerName: ASRProvider) => {
    if (!audioPath) {
      setError('请输入音频文件路径')
      return
    }

    setIsRecognizing(true)
    setError(null)
    setProgress(0)

    const progressInterval = setInterval(() => {
      setProgress(prev => Math.min(prev + 5, 90))
    }, 500)

    try {
      const response = await fetch(`/api/v1/asr/providers/${providerName}/recognize`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          audio_path: audioPath,
          language,
          model,
          enable_timestamps: enableTimestamps,
          enable_word_level: enableWordLevel,
          enable_gpu: enableGPU,
        }),
      })

      clearInterval(progressInterval)
      setProgress(100)

      const data = await response.json()

      if (data.success) {
        setResult(data.data)
        onRecognitionComplete?.(data.data)
      } else {
        setError(data.error?.message || '识别失败')
      }
    } catch (err) {
      clearInterval(progressInterval)
      setError(err instanceof Error ? err.message : '识别失败')
    } finally {
      setIsRecognizing(false)
    }
  }

  // 清除缓存
  const handleClearCache = async () => {
    try {
      await fetch('/api/v1/asr/cache/clear', { method: 'POST' })
      fetchStats()
    } catch (err) {
      console.error('Failed to clear cache:', err)
    }
  }

  // 格式化时长
  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  // 获取提供商图标
  const getProviderIcon = (name: ASRProvider) => {
    switch (name) {
      case 'bcut-asr':
        return <Cloud className="h-4 w-4" />
      case 'whisper':
        return <Cpu className="h-4 w-4" />
      default:
        return <Mic className="h-4 w-4" />
    }
  }

  return (
    <div className="space-y-6">
      {/* 基本设置 */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Mic className="h-5 w-5 text-blue-600" />
            <CardTitle className="text-lg">智能语音识别</CardTitle>
          </div>
          <CardDescription>
            支持多种ASR引擎，自动选择最优方案
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 音频路径 */}
          <div className="space-y-2">
            <Label>音频文件路径</Label>
            <div className="flex gap-2">
              <Input
                value={audioPath}
                onChange={(e) => setAudioPath(e.target.value)}
                placeholder="输入音频文件路径或上传文件"
                className="flex-1"
              />
              <Button variant="outline" size="icon">
                <Upload className="h-4 w-4" />
              </Button>
            </div>
          </div>

          {/* 语言和模型选择 */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label>识别语言</Label>
              <Select value={language} onValueChange={setLanguage}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(SUPPORTED_LANGUAGES).map(([code, name]) => (
                    <SelectItem key={code} value={code}>
                      {name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Whisper模型</Label>
              <Select value={model} onValueChange={setModel}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {WHISPER_MODELS.map((m) => (
                    <SelectItem key={m.name} value={m.name}>
                      {m.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* 提供商选择 */}
          <div className="space-y-2">
            <Label>ASR引擎</Label>
            <div className="flex gap-2">
              <Button
                variant={provider === 'auto' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setProvider('auto')}
              >
                <Mic className="h-4 w-4 mr-2" />
                自动选择
              </Button>
              {providers.filter(p => p.available).map((p) => (
                <Button
                  key={p.name}
                  variant={provider === p.name ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setProvider(p.name)}
                >
                  {getProviderIcon(p.name)}
                  <span className="ml-2">{p.name === 'bcut-asr' ? '必剪云' : 'Whisper'}</span>
                </Button>
              ))}
            </div>
          </div>

          {/* 高级设置 */}
          <div className="border-t pt-4">
            <button
              className="flex items-center gap-2 text-sm text-gray-600 hover:text-gray-900"
              onClick={() => setShowAdvanced(!showAdvanced)}
            >
              <Settings className="h-4 w-4" />
              高级设置
              {showAdvanced ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
            </button>

            {showAdvanced && (
              <div className="mt-4 space-y-3">
                <div className="flex items-center gap-4">
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={enableTimestamps}
                      onChange={(e) => setEnableTimestamps(e.target.checked)}
                      className="rounded"
                    />
                    启用时间戳
                  </label>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={enableWordLevel}
                      onChange={(e) => setEnableWordLevel(e.target.checked)}
                      className="rounded"
                    />
                    词级时间戳
                  </label>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={enableGPU}
                      onChange={(e) => setEnableGPU(e.target.checked)}
                      className="rounded"
                    />
                    GPU加速
                  </label>
                  <label className="flex items-center gap-2 text-sm">
                    <input
                      type="checkbox"
                      checked={useChunking}
                      onChange={(e) => setUseChunking(e.target.checked)}
                      className="rounded"
                    />
                    大文件分片
                  </label>
                </div>
              </div>
            )}
          </div>

          {/* 识别按钮 */}
          <div className="flex gap-2">
            <Button
              onClick={provider === 'auto' ? handleRecognize : () => handleRecognizeWithProvider(provider)}
              disabled={isRecognizing || !audioPath}
              className="flex-1"
            >
              {isRecognizing ? (
                <>
                  <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                  识别中...
                </>
              ) : (
                <>
                  <Play className="h-4 w-4 mr-2" />
                  开始识别
                </>
              )}
            </Button>
          </div>

          {/* 进度条 */}
          {isRecognizing && (
            <div className="space-y-2">
              <Progress value={progress} />
              <p className="text-sm text-gray-500 text-center">
                {progress < 30 ? '正在加载音频...' :
                 progress < 60 ? '正在识别...' :
                 progress < 90 ? '正在处理结果...' : '即将完成...'}
              </p>
            </div>
          )}

          {/* 错误提示 */}
          {error && (
            <div className="flex items-center gap-2 text-red-600 bg-red-50 p-3 rounded-lg">
              <AlertCircle className="h-4 w-4" />
              <span className="text-sm">{error}</span>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 识别结果 */}
      {result && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Check className="h-5 w-5 text-green-600" />
                <CardTitle className="text-lg">识别结果</CardTitle>
              </div>
              <div className="flex items-center gap-2">
                {result.cached && (
                  <Badge variant="secondary" className="bg-green-100 text-green-700">
                    来自缓存
                  </Badge>
                )}
                <Badge variant="outline">
                  {result.provider === 'bcut-asr' ? '必剪云' : 'Whisper'}
                </Badge>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* 元信息 */}
            <div className="flex items-center gap-4 text-sm text-gray-500">
              <span>语言: {SUPPORTED_LANGUAGES[result.language] || result.language}</span>
              <span>时长: {formatDuration(result.audio_duration)}</span>
              <span>词数: {result.word_count}</span>
              <span>质量评分: {result.quality_score.toFixed(1)}</span>
            </div>

            {/* 完整文本 */}
            <div className="bg-gray-50 p-4 rounded-lg max-h-64 overflow-y-auto">
              <p className="text-sm text-gray-700 whitespace-pre-wrap">{result.text}</p>
            </div>

            {/* 分段信息 */}
            {result.segments.length > 0 && (
              <div className="space-y-2">
                <h4 className="text-sm font-medium">分段信息</h4>
                <div className="max-h-48 overflow-y-auto space-y-1">
                  {result.segments.slice(0, 20).map((seg) => (
                    <div key={seg.id} className="flex gap-2 text-sm">
                      <span className="text-gray-400 w-20 flex-shrink-0">
                        [{formatDuration(seg.start)}]
                      </span>
                      <span className="text-gray-600 flex-1">{seg.text}</span>
                      {seg.confidence !== undefined && (
                        <span className="text-gray-400">
                          {(seg.confidence * 100).toFixed(0)}%
                        </span>
                      )}
                    </div>
                  ))}
                  {result.segments.length > 20 && (
                    <p className="text-sm text-gray-400 text-center">
                      ...共 {result.segments.length} 个分段
                    </p>
                  )}
                </div>
              </div>
            )}

            {/* 元数据 */}
            {result.metadata && (
              <div className="text-xs text-gray-400">
                {result.metadata.chunked && (
                  <span>分片处理: {result.metadata.chunk_count} 个分片</span>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* 统计信息 */}
      {stats && (
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <CardTitle className="text-sm">识别统计</CardTitle>
              <Button variant="ghost" size="sm" onClick={handleClearCache}>
                清除缓存
              </Button>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-4 gap-4 text-center">
              <div>
                <p className="text-2xl font-bold">{stats.total_requests}</p>
                <p className="text-xs text-gray-500">总请求数</p>
              </div>
              <div>
                <p className="text-2xl font-bold text-green-600">{stats.cache_hits}</p>
                <p className="text-xs text-gray-500">缓存命中</p>
              </div>
              <div>
                <p className="text-2xl font-bold text-blue-600">
                  {stats.total_requests > 0 
                    ? ((stats.cache_hits / stats.total_requests) * 100).toFixed(1) 
                    : 0}%
                </p>
                <p className="text-xs text-gray-500">命中率</p>
              </div>
              <div>
                <p className="text-2xl font-bold text-orange-600">{stats.fallback_count}</p>
                <p className="text-xs text-gray-500">降级次数</p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

import { useState, useEffect } from 'react'
import { Video, Download, Loader2, CheckCircle2, XCircle, Clock, Play, Pause, RefreshCw, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { downloadVideo, getVideoStatus, submitTranscription, getVideoTranscription, type VideoInfo, type TranscriptResult } from '@/lib/api'

interface VideoProcessingProgressProps {
  videoUrl?: string
  onVideoDownloaded?: (videoId: string) => void
  onTranscriptionComplete?: (result: TranscriptResult) => void
}

interface ProcessingStep {
  id: string
  name: string
  status: 'pending' | 'processing' | 'completed' | 'failed'
  progress: number
  error?: string
}

const processingSteps: ProcessingStep[] = [
  { id: 'download', name: '下载视频', status: 'pending', progress: 0 },
  { id: 'transcribe', name: '语音转录', status: 'pending', progress: 0 },
  { id: 'optimize', name: '文本优化', status: 'pending', progress: 0 },
  { id: 'summary', name: '生成摘要', status: 'pending', progress: 0 },
]

export default function VideoProcessingProgress({ videoUrl, onVideoDownloaded, onTranscriptionComplete }: VideoProcessingProgressProps) {
  const [url, setUrl] = useState(videoUrl || '')
  const [steps, setSteps] = useState<ProcessingStep[]>(processingSteps)
  const [videoInfo, setVideoInfo] = useState<VideoInfo | null>(null)
  const [transcription, setTranscription] = useState<TranscriptResult | null>(null)
  const [isActive, setIsActive] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isActive) {
      processVideo()
    }
  }, [isActive])

  async function processVideo() {
    if (!url) {
      setError('请输入视频链接')
      return
    }

    setError(null)
    setIsActive(true)

    try {
      // 步骤 1: 下载视频
      await updateStep('download', 'processing', 0)
      const downloadResult = await downloadVideo(url, true)

      if (!downloadResult.success) {
        await updateStep('download', 'failed', 0, downloadResult.message || '下载失败')
        setIsActive(false)
        return
      }

      // 模拟下载进度
      for (let i = 10; i <= 100; i += 10) {
        await new Promise(resolve => setTimeout(resolve, 200))
        await updateStep('download', 'processing', i)
      }

      await updateStep('download', 'completed', 100)

      // 获取视频信息
      const videoId = downloadResult.data?.videoId
      if (!videoId) {
        await updateStep('download', 'failed', 0, '下载失败: 无效的视频ID')
        setIsActive(false)
        return
      }

      const statusResult = await getVideoStatus(videoId)
      if (statusResult.success && statusResult.data) {
        setVideoInfo(statusResult.data)
        onVideoDownloaded?.(videoId)
      }

      // 步骤 2: 语音转录
      await updateStep('transcribe', 'processing', 0)
      const transcribeResult = await submitTranscription(videoId)

      if (!transcribeResult.success) {
        await updateStep('transcribe', 'failed', 0, transcribeResult.message || '转录失败')
        setIsActive(false)
        return
      }

      // 轮询转录结果
      await pollTranscriptionResult(videoId)

    } catch (err) {
      setError('处理失败，请检查网络连接')
      setIsActive(false)
    }
  }

  async function pollTranscriptionResult(videoId: string) {
    const maxAttempts = 60
    let attempts = 0

    while (attempts < maxAttempts && isActive) {
      attempts++

      try {
        const result = await getVideoTranscription(videoId)

        if (result.success && result.data) {
          const progress = Math.min((attempts / maxAttempts) * 100, 90)
          await updateStep('transcribe', 'processing', progress)

          if (result.data.status === 'completed' && result.data.transcription) {
            await updateStep('transcribe', 'completed', 100)
            setTranscription(result.data.transcription)
            onTranscriptionComplete?.(result.data.transcription)

            // 继续后续步骤
            await continueOptimization(result.data.transcription)
            return
          } else if (result.data.status === 'failed') {
            await updateStep('transcribe', 'failed', 0, '转录失败')
            setIsActive(false)
            return
          }
        }

        await new Promise(resolve => setTimeout(resolve, 3000))
      } catch (err) {
        if (attempts >= maxAttempts) {
          await updateStep('transcribe', 'failed', 0, '获取转录结果超时')
          setIsActive(false)
          return
        }
      }
    }
  }

  async function continueOptimization(_transcriptResult: TranscriptResult) {
    // 步骤 3: 文本优化 (模拟)
    await updateStep('optimize', 'processing', 0)
    for (let i = 0; i <= 100; i += 20) {
      if (!isActive) return
      await new Promise(resolve => setTimeout(resolve, 500))
      await updateStep('optimize', 'processing', i)
    }
    await updateStep('optimize', 'completed', 100)

    // 步骤 4: 生成摘要 (模拟)
    await updateStep('summary', 'processing', 0)
    for (let i = 0; i <= 100; i += 25) {
      if (!isActive) return
      await new Promise(resolve => setTimeout(resolve, 600))
      await updateStep('summary', 'processing', i)
    }
    await updateStep('summary', 'completed', 100)

    setIsActive(false)
  }

  async function updateStep(stepId: string, status: ProcessingStep['status'], progress: number, error?: string) {
    setSteps(prev => prev.map(step =>
      step.id === stepId ? { ...step, status, progress, error } : step
    ))
  }

  function resetProcessing() {
    setSteps(processingSteps)
    setVideoInfo(null)
    setTranscription(null)
    setIsActive(false)
    setError(null)
  }

  function formatFileSize(bytes: number) {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
    return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`
  }

  function formatDuration(seconds: number) {
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const isCompleted = steps.every(step => step.status === 'completed')
  const hasError = steps.some(step => step.status === 'failed')
  const overallProgress = Math.round(steps.reduce((acc, step) => acc + step.progress, 0) / steps.length)

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Video className="h-5 w-5 text-blue-600" />
            <CardTitle className="text-lg">视频处理</CardTitle>
            {isActive && (
              <Badge variant="secondary" className="animate-pulse">
                处理中
              </Badge>
            )}
            {isCompleted && (
              <Badge variant="default" className="bg-green-500">
                已完成
              </Badge>
            )}
            {hasError && (
              <Badge variant="destructive">
                处理失败
              </Badge>
            )}
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={resetProcessing}
              disabled={isActive && !hasError}
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              重置
            </Button>
            {!isActive && !isCompleted && !hasError && (
              <Button
                size="sm"
                onClick={() => setIsActive(true)}
              >
                <Play className="h-4 w-4 mr-2" />
                开始处理
              </Button>
            )}
            {isActive && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsActive(false)}
              >
                <Pause className="h-4 w-4 mr-2" />
                暂停
              </Button>
            )}
          </div>
        </div>
        <CardDescription>自动下载、转录、优化和生成视频摘要</CardDescription>
      </CardHeader>

      <CardContent className="space-y-6">
        {/* URL 输入 */}
        {!videoInfo && (
          <div className="flex gap-2">
            <input
              type="text"
              placeholder="输入抖音/快手/B站视频链接..."
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              className="flex-1 px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={isActive}
            />
            <Button
              onClick={() => setIsActive(true)}
              disabled={isActive || !url}
            >
              <Download className="h-4 w-4 mr-2" />
              添加
            </Button>
          </div>
        )}

        {/* 错误提示 */}
        {error && (
          <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700">
            <AlertCircle className="h-4 w-4" />
            <span className="text-sm">{error}</span>
          </div>
        )}

        {/* 总体进度 */}
        {(isActive || isCompleted) && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">总体进度</span>
              <span className="font-medium">{overallProgress}%</span>
            </div>
            <Progress value={overallProgress} className="h-2" />
          </div>
        )}

        {/* 处理步骤 */}
        <div className="space-y-3">
          {steps.map((step) => (
            <div
              key={step.id}
              className={`p-4 border rounded-lg transition-all ${
                step.status === 'processing' ? 'border-blue-300 bg-blue-50' :
                step.status === 'completed' ? 'border-green-300 bg-green-50' :
                step.status === 'failed' ? 'border-red-300 bg-red-50' :
                'border-gray-200'
              }`}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  {step.status === 'processing' && <Loader2 className="h-4 w-4 animate-spin text-blue-600" />}
                  {step.status === 'completed' && <CheckCircle2 className="h-4 w-4 text-green-600" />}
                  {step.status === 'failed' && <XCircle className="h-4 w-4 text-red-600" />}
                  {step.status === 'pending' && <Clock className="h-4 w-4 text-gray-400" />}
                  <span className="font-medium">{step.name}</span>
                </div>
                <Badge
                  variant={step.status === 'completed' ? 'default' :
                          step.status === 'processing' ? 'secondary' :
                          step.status === 'failed' ? 'destructive' : 'outline'}
                  className="text-xs"
                >
                  {step.status === 'processing' ? `${step.progress}%` :
                   step.status === 'completed' ? '完成' :
                   step.status === 'failed' ? '失败' : '等待'}
                </Badge>
              </div>

              {step.status === 'processing' && (
                <Progress value={step.progress} className="h-1" />
              )}

              {step.error && (
                <p className="text-xs text-red-600 mt-2">{step.error}</p>
              )}
            </div>
          ))}
        </div>

        {/* 视频信息 */}
        {videoInfo && (
          <Tabs defaultValue="info" className="w-full">
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="info">视频信息</TabsTrigger>
              <TabsTrigger value="transcript">转录结果</TabsTrigger>
            </TabsList>

            <TabsContent value="info" className="space-y-3 mt-4">
              <div className="border rounded-lg p-4 space-y-3">
                <div>
                  <h4 className="font-medium line-clamp-2">{videoInfo.title}</h4>
                  <div className="flex items-center gap-2 mt-1 text-sm text-muted-foreground">
                    <span>{videoInfo.author}</span>
                    <Badge variant="secondary">{videoInfo.platform}</Badge>
                  </div>
                </div>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    时长: {formatDuration(videoInfo.duration)}
                  </span>
                  <span>大小: {formatFileSize(videoInfo.fileSize)}</span>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="transcript" className="space-y-3 mt-4">
              {transcription ? (
                <div className="border rounded-lg p-4 space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <CheckCircle2 className="h-4 w-4 text-green-600" />
                      <span className="font-medium">转录完成</span>
                    </div>
                    <Badge variant="secondary">{transcription.engine}</Badge>
                  </div>
                  <p className="text-sm text-gray-700 line-clamp-6">{transcription.text}</p>
                  {transcription.keywords.length > 0 && (
                    <div className="flex flex-wrap gap-1 pt-2">
                      {transcription.keywords.slice(0, 8).map((kw, idx) => (
                        <Badge key={idx} variant="outline" className="text-xs">
                          {kw}
                        </Badge>
                      ))}
                    </div>
                  )}
                </div>
              ) : (
                <div className="flex items-center justify-center h-32 text-muted-foreground text-sm">
                  {isActive ? '正在转录中...' : '等待转录'}
                </div>
              )}
            </TabsContent>
          </Tabs>
        )}
      </CardContent>
    </Card>
  )
}

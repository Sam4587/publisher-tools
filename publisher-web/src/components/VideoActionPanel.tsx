import { useState } from 'react'
import { Video, Download, Loader2, CheckCircle, XCircle, FileText } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { downloadVideo, getVideoStatus, submitTranscription, getVideoTranscription, type VideoInfo, type TranscriptResult } from '@/lib/api'

interface VideoActionPanelProps {
  videoUrl?: string
  onVideoDownloaded?: (videoId: string) => void
  onTranscriptionComplete?: (result: TranscriptResult) => void
}

export default function VideoActionPanel({ videoUrl, onVideoDownloaded, onTranscriptionComplete }: VideoActionPanelProps) {
  const [url, setUrl] = useState(videoUrl || '')
  const [_loading, _setLoading] = useState(false)
  const [downloading, setDownloading] = useState(false)
  const [transcribing, setTranscribing] = useState(false)
  const [videoInfo, setVideoInfo] = useState<VideoInfo | null>(null)
  const [transcription, setTranscription] = useState<TranscriptResult | null>(null)
  const [error, setError] = useState<string | null>(null)

  // 下载视频
  async function handleDownload() {
    if (!url) return

    setDownloading(true)
    setError(null)

    try {
      const result = await downloadVideo(url, true)

      if (result.success && result.data) {
        // 获取视频详情
        const statusResult = await getVideoStatus(result.data.videoId)
        if (statusResult.success && statusResult.data) {
          setVideoInfo(statusResult.data)
          onVideoDownloaded?.(result.data.videoId)
        }
      } else {
        setError(result.message || '下载失败')
      }
    } catch (err) {
      setError('下载失败，请检查链接是否正确')
    } finally {
      setDownloading(false)
    }
  }

  // 开始转录
  async function handleTranscribe() {
    if (!videoInfo) return

    setTranscribing(true)
    setError(null)

    try {
      const result = await submitTranscription(videoInfo.videoId)

      if (result.success && result.data) {
        // 轮询获取转录结果
        pollTranscriptionResult(videoInfo.videoId)
      } else {
        setError(result.message || '提交转录任务失败')
        setTranscribing(false)
      }
    } catch (err) {
      setError('提交转录任务失败')
      setTranscribing(false)
    }
  }

  // 轮询获取转录结果
  async function pollTranscriptionResult(videoId: string) {
    const maxAttempts = 60 // 最多等待 5 分钟
    let attempts = 0

    const poll = async () => {
      attempts++

      try {
        const result = await getVideoTranscription(videoId)

        if (result.success && result.data) {
          if (result.data.status === 'completed' && result.data.transcription) {
            setTranscription(result.data.transcription)
            onTranscriptionComplete?.(result.data.transcription)
            setTranscribing(false)
            return
          } else if (result.data.status === 'processing') {
            // 继续等待
            if (attempts < maxAttempts) {
              setTimeout(poll, 5000)
            } else {
              setError('转录超时')
              setTranscribing(false)
            }
          }
        } else {
          if (attempts < maxAttempts) {
            setTimeout(poll, 5000)
          } else {
            setError('获取转录结果超时')
            setTranscribing(false)
          }
        }
      } catch (err) {
        if (attempts < maxAttempts) {
          setTimeout(poll, 5000)
        } else {
          setError('获取转录结果失败')
          setTranscribing(false)
        }
      }
    }

    poll()
  }

  // 格式化文件大小
  function formatFileSize(bytes: number) {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
    return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`
  }

  // 格式化时长
  function formatDuration(seconds: number) {
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Video className="h-5 w-5 text-blue-600" />
          <CardTitle className="text-lg">视频转录</CardTitle>
        </div>
        <CardDescription>下载视频并进行 AI 转录</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* URL 输入 */}
        <div className="flex gap-2">
          <Input
            placeholder="输入抖音/快手视频链接..."
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            className="flex-1"
          />
          <Button onClick={handleDownload} disabled={downloading || !url}>
            {downloading ? (
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
            ) : (
              <Download className="h-4 w-4 mr-2" />
            )}
            下载
          </Button>
        </div>

        {/* 错误提示 */}
        {error && (
          <div className="flex items-center gap-2 text-red-600 text-sm">
            <XCircle className="h-4 w-4" />
            {error}
          </div>
        )}

        {/* 视频信息 */}
        {videoInfo && (
          <div className="border rounded-lg p-4 space-y-3">
            <div className="flex items-start justify-between">
              <div>
                <h4 className="font-medium line-clamp-2">{videoInfo.title}</h4>
                <div className="flex items-center gap-2 mt-1 text-sm text-gray-500">
                  <span>{videoInfo.author}</span>
                  <Badge variant="secondary">{videoInfo.platform}</Badge>
                </div>
              </div>
              <CheckCircle className="h-5 w-5 text-green-600 flex-shrink-0" />
            </div>
            <div className="flex items-center gap-4 text-sm text-gray-500">
              <span>时长: {formatDuration(videoInfo.duration)}</span>
              <span>大小: {formatFileSize(videoInfo.fileSize)}</span>
            </div>

            {/* 转录按钮 */}
            {!transcription && (
              <Button
                onClick={handleTranscribe}
                disabled={transcribing}
                variant="outline"
                className="w-full"
              >
                {transcribing ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-2" />
                    转录中...
                  </>
                ) : (
                  <>
                    <FileText className="h-4 w-4 mr-2" />
                    开始转录
                  </>
                )}
              </Button>
            )}
          </div>
        )}

        {/* 转录结果预览 */}
        {transcription && (
          <div className="border rounded-lg p-4 space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <CheckCircle className="h-4 w-4 text-green-600" />
                <span className="font-medium">转录完成</span>
              </div>
              <Badge variant="secondary">{transcription.engine}</Badge>
            </div>
            <p className="text-sm text-gray-600 line-clamp-3">{transcription.text.slice(0, 200)}...</p>
            {transcription.keywords.length > 0 && (
              <div className="flex flex-wrap gap-1">
                {transcription.keywords.slice(0, 5).map((kw, idx) => (
                  <Badge key={idx} variant="outline" className="text-xs">
                    {kw}
                  </Badge>
                ))}
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

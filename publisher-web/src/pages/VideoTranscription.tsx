import { useState } from 'react'
import { Video, FileText, Sparkles, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import VideoActionPanel from '@/components/VideoActionPanel'
import ContentRewritePanel from '@/components/ContentRewritePanel'
import { publishAsync } from '@/lib/api'
import type { TranscriptResult } from '@/lib/api'
import { useNavigate } from 'react-router-dom'

export default function VideoTranscription() {
  const navigate = useNavigate()
  const [transcript, setTranscript] = useState<TranscriptResult | null>(null)
  const [activeStep, setActiveStep] = useState<'download' | 'transcribe' | 'rewrite'>('download')

  // 转录完成回调
  function handleTranscriptionComplete(result: TranscriptResult) {
    setTranscript(result)
    setActiveStep('rewrite')
  }

  // 发布到平台
  async function handlePublish(platform: string, content: any) {
    try {
      const publishContent = {
        platform: platform as any,
        type: 'images' as const, // 默认图文类型
        title: content.title || '无标题',
        body: content.content || content.mainContent || '',
        tags: content.tags || [],
      }

      const response = await publishAsync(publishContent)
      
      if (response.success && response.data) {
        alert(`发布任务已创建，任务ID: ${response.data.task_id}`)
        // 跳转到历史页面查看任务状态
        navigate('/history')
      } else {
        alert(response.error || '发布失败')
      }
    } catch (error) {
      console.error('发布失败:', error)
      alert('发布失败，请重试')
    }
  }

  // 步骤配置
  const steps = [
    { id: 'download', label: '下载视频', icon: Video },
    { id: 'transcribe', label: 'AI转录', icon: FileText },
    { id: 'rewrite', label: '内容改写', icon: Sparkles },
  ]

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="sm" onClick={() => window.history.back()}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            返回
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900">视频转录</h1>
            <p className="text-gray-500 text-sm mt-1">将视频内容转化为可编辑文本，并进行二次创作</p>
          </div>
        </div>
      </div>

      {/* 步骤指示器 */}
      <div className="flex items-center gap-4">
        {steps.map((step, index) => {
          const isActive = step.id === activeStep
          const isCompleted = steps.findIndex(s => s.id === activeStep) > index
          const Icon = step.icon

          return (
            <div key={step.id} className="flex items-center">
              <button
                onClick={() => setActiveStep(step.id as typeof activeStep)}
                className={`flex items-center gap-2 px-4 py-2 rounded-full transition-colors ${
                  isActive
                    ? 'bg-blue-600 text-white'
                    : isCompleted
                    ? 'bg-green-100 text-green-700'
                    : 'bg-gray-100 text-gray-500'
                }`}
              >
                <Icon className="h-4 w-4" />
                <span className="text-sm font-medium">{step.label}</span>
                {isCompleted && (
                  <Badge variant="secondary" className="ml-1 bg-green-200 text-green-800">
                    完成
                  </Badge>
                )}
              </button>
              {index < steps.length - 1 && (
                <div className={`w-8 h-0.5 mx-2 ${isCompleted ? 'bg-green-300' : 'bg-gray-200'}`} />
              )}
            </div>
          )
        })}
      </div>

      {/* 主内容区域 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 左侧：视频下载与转录 */}
        <div className="space-y-4">
          <VideoActionPanel
            onTranscriptionComplete={handleTranscriptionComplete}
          />
        </div>

        {/* 右侧：转录结果预览 */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <FileText className="h-5 w-5 text-green-600" />
              <CardTitle className="text-lg">转录结果</CardTitle>
            </div>
            <CardDescription>视频内容的文本转录</CardDescription>
          </CardHeader>
          <CardContent>
            {transcript ? (
              <div className="space-y-4">
                {/* 元信息 */}
                <div className="flex items-center gap-4 text-sm">
                  <Badge variant="secondary">{transcript.engine}</Badge>
                  <span className="text-gray-500">时长: {Math.floor(transcript.duration)}s</span>
                  <span className="text-gray-500">语言: {transcript.language}</span>
                </div>

                {/* 完整文本 */}
                <div className="max-h-96 overflow-y-auto">
                  <p className="text-sm text-gray-700 whitespace-pre-wrap">{transcript.text}</p>
                </div>

                {/* 关键词 */}
                {transcript.keywords.length > 0 && (
                  <div className="border-t pt-4">
                    <p className="text-sm font-medium mb-2">关键词</p>
                    <div className="flex flex-wrap gap-1">
                      {transcript.keywords.map((kw, idx) => (
                        <Badge key={idx} variant="outline">{kw}</Badge>
                      ))}
                    </div>
                  </div>
                )}

                {/* 分段信息 */}
                {transcript.segments.length > 0 && (
                  <div className="border-t pt-4">
                    <p className="text-sm font-medium mb-2">分段信息</p>
                    <div className="space-y-2 max-h-48 overflow-y-auto">
                      {transcript.segments.slice(0, 10).map((seg, idx) => (
                        <div key={idx} className="flex gap-2 text-sm">
                          <span className="text-gray-400 w-16 flex-shrink-0">
                            [{Math.floor(seg.start)}s]
                          </span>
                          <span className="text-gray-600">{seg.text}</span>
                        </div>
                      ))}
                      {transcript.segments.length > 10 && (
                        <p className="text-sm text-gray-400">...共 {transcript.segments.length} 个分段</p>
                      )}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="text-center py-12 text-gray-500">
                <FileText className="h-12 w-12 mx-auto mb-4 opacity-50" />
                <p>请先下载并转录视频</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 内容改写面板 */}
      {transcript && (
        <ContentRewritePanel
          transcript={transcript}
          onPublish={handlePublish}
        />
      )}
    </div>
  )
}

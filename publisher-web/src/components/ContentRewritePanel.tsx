import { useState } from 'react'
import { Sparkles, Loader2, Copy, Send } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Textarea } from '@/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { TranscriptResult } from '@/lib/api'

interface RewriteResult {
  title?: string
  content?: string
  tags?: string[]
  hook?: string
  mainContent?: string
  cta?: string
  microContent?: string
}

interface ContentRewritePanelProps {
  transcript: TranscriptResult | null
  onPublish?: (platform: string, content: RewriteResult) => void
}

export default function ContentRewritePanel({ transcript, onPublish }: ContentRewritePanelProps) {
  const [rewriting, setRewriting] = useState(false)
  const [publishing, setPublishing] = useState<string | null>(null)
  const [results, setResults] = useState<Record<string, RewriteResult>>({})
  const [error, setError] = useState<string | null>(null)

  // 生成改写内容
  async function handleRewrite() {
    if (!transcript) return

    setRewriting(true)
    setError(null)

    try {
      const response = await fetch('/api/content/video-rewrite', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          text: transcript.text,
          platforms: ['xiaohongshu', 'douyin', 'toutiao']
        })
      })

      const result = await response.json()

      if (result.success && result.data) {
        setResults(result.data.results || {})
      } else {
        setError(result.message || '生成失败')
      }
    } catch (err) {
      setError('生成失败，请稍后重试')
    } finally {
      setRewriting(false)
    }
  }

  // 发布到平台
  async function handlePublish(platform: string) {
    const content = results[platform]
    if (!content || !onPublish) return

    setPublishing(platform)

    try {
      await onPublish(platform, content)
    } finally {
      setPublishing(null)
    }
  }

  // 复制内容
  function handleCopy(content: string) {
    navigator.clipboard.writeText(content)
  }

  if (!transcript) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">内容改写</CardTitle>
          <CardDescription>请先完成视频转录</CardDescription>
        </CardHeader>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-purple-600" />
            <CardTitle className="text-lg">内容改写</CardTitle>
          </div>
          <Button onClick={handleRewrite} disabled={rewriting}>
            {rewriting ? (
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
            ) : (
              <Sparkles className="h-4 w-4 mr-2" />
            )}
            生成内容
          </Button>
        </div>
        <CardDescription>将转录内容改写为多平台风格</CardDescription>
      </CardHeader>
      <CardContent>
        {error && (
          <div className="text-red-600 text-sm mb-4">{error}</div>
        )}

        {Object.keys(results).length > 0 ? (
          <Tabs defaultValue="xiaohongshu">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="xiaohongshu">小红书</TabsTrigger>
              <TabsTrigger value="douyin">抖音</TabsTrigger>
              <TabsTrigger value="toutiao">今日头条</TabsTrigger>
            </TabsList>

            {/* 小红书 */}
            <TabsContent value="xiaohongshu" className="space-y-4">
              {results.xiaohongshu && (
                <div className="space-y-3">
                  <div>
                    <label className="text-sm font-medium">标题</label>
                    <div className="flex items-center gap-2 mt-1">
                      <input
                        type="text"
                        value={results.xiaohongshu.title || ''}
                        readOnly
                        className="flex-1 border rounded px-3 py-2 text-sm"
                      />
                      <Button variant="ghost" size="sm" onClick={() => handleCopy(results.xiaohongshu?.title || '')}>
                        <Copy className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                  <div>
                    <label className="text-sm font-medium">正文</label>
                    <Textarea
                      value={results.xiaohongshu.content || ''}
                      readOnly
                      rows={6}
                      className="mt-1"
                    />
                  </div>
                  {results.xiaohongshu.tags && (
                    <div className="flex flex-wrap gap-1">
                      {results.xiaohongshu.tags.map((tag, idx) => (
                        <Badge key={idx} variant="secondary">#{tag}</Badge>
                      ))}
                    </div>
                  )}
                  <div className="flex justify-end">
                    <Button onClick={() => handlePublish('xiaohongshu')} disabled={publishing === 'xiaohongshu'}>
                      {publishing === 'xiaohongshu' ? (
                        <Loader2 className="h-4 w-4 animate-spin mr-2" />
                      ) : (
                        <Send className="h-4 w-4 mr-2" />
                      )}
                      发布到小红书
                    </Button>
                  </div>
                </div>
              )}
            </TabsContent>

            {/* 抖音 */}
            <TabsContent value="douyin" className="space-y-4">
              {results.douyin && (
                <div className="space-y-3">
                  <div>
                    <label className="text-sm font-medium text-red-600">开头钩子（前3秒）</label>
                    <div className="mt-1 p-3 bg-red-50 rounded text-sm">
                      {results.douyin.hook}
                    </div>
                  </div>
                  <div>
                    <label className="text-sm font-medium">主体内容</label>
                    <Textarea
                      value={results.douyin.mainContent || ''}
                      readOnly
                      rows={5}
                      className="mt-1"
                    />
                  </div>
                  <div>
                    <label className="text-sm font-medium text-blue-600">结尾引导</label>
                    <div className="mt-1 p-3 bg-blue-50 rounded text-sm">
                      {results.douyin.cta}
                    </div>
                  </div>
                  <div className="flex justify-end">
                    <Button onClick={() => handlePublish('douyin')} disabled={publishing === 'douyin'}>
                      {publishing === 'douyin' ? (
                        <Loader2 className="h-4 w-4 animate-spin mr-2" />
                      ) : (
                        <Send className="h-4 w-4 mr-2" />
                      )}
                      发布到抖音
                    </Button>
                  </div>
                </div>
              )}
            </TabsContent>

            {/* 今日头条 */}
            <TabsContent value="toutiao" className="space-y-4">
              {results.toutiao && (
                <div className="space-y-3">
                  <div>
                    <label className="text-sm font-medium">标题</label>
                    <div className="flex items-center gap-2 mt-1">
                      <input
                        type="text"
                        value={results.toutiao.title || ''}
                        readOnly
                        className="flex-1 border rounded px-3 py-2 text-sm"
                      />
                      <Button variant="ghost" size="sm" onClick={() => handleCopy(results.toutiao?.title || '')}>
                        <Copy className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                  <div>
                    <label className="text-sm font-medium">正文</label>
                    <Textarea
                      value={results.toutiao.content || ''}
                      readOnly
                      rows={8}
                      className="mt-1"
                    />
                  </div>
                  {results.toutiao.microContent && (
                    <div>
                      <label className="text-sm font-medium text-orange-600">微头条版本</label>
                      <div className="mt-1 p-3 bg-orange-50 rounded text-sm">
                        {results.toutiao.microContent}
                      </div>
                    </div>
                  )}
                  {results.toutiao.tags && (
                    <div className="flex flex-wrap gap-1">
                      {results.toutiao.tags.map((tag, idx) => (
                        <Badge key={idx} variant="outline">#{tag}</Badge>
                      ))}
                    </div>
                  )}
                  <div className="flex justify-end">
                    <Button onClick={() => handlePublish('toutiao')} disabled={publishing === 'toutiao'}>
                      {publishing === 'toutiao' ? (
                        <Loader2 className="h-4 w-4 animate-spin mr-2" />
                      ) : (
                        <Send className="h-4 w-4 mr-2" />
                      )}
                      发布到头条
                    </Button>
                  </div>
                </div>
              )}
            </TabsContent>
          </Tabs>
        ) : (
          <div className="text-center py-8 text-gray-500">
            <Sparkles className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p>点击"生成内容"开始 AI 改写</p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

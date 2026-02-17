import { useState, useEffect } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { Sparkles, Loader2, Copy, RefreshCw, CheckCircle, Upload, Send } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { aiContentGenerate, aiContentRewrite, publish } from '@/lib/api'
import type { Platform, ContentType } from '@/types/api'

const styles = [
  { value: '轻松幽默', label: '轻松幽默' },
  { value: '正式专业', label: '正式专业' },
  { value: '感性温暖', label: '感性温暖' },
  { value: '理性分析', label: '理性分析' },
  { value: '故事化', label: '故事化' },
]

const platforms = [
  { value: 'douyin', label: '抖音' },
  { value: 'toutiao', label: '今日头条' },
  { value: 'xiaohongshu', label: '小红书' },
  { value: 'weibo', label: '微博' },
  { value: 'general', label: '通用' },
]

export default function ContentGeneration() {
  const location = useLocation()
  const navigate = useNavigate()
  const [topic, setTopic] = useState('')
  const [style, setStyle] = useState('轻松幽默')
  const [platform, setPlatform] = useState('general')
  const [length, setLength] = useState(500)
  const [generatedContent, setGeneratedContent] = useState('')
  const [loading, setLoading] = useState(false)
  const [copied, setCopied] = useState(false)
  const [rewriting, setRewriting] = useState(false)
  const [publishing, setPublishing] = useState(false)

  // 接收来自热点页面的参数
  useEffect(() => {
    if (location.state) {
      const state = location.state as { topic?: string; source?: string }
      if (state.topic) {
        setTopic(state.topic)
      }
    }
  }, [location.state])

  const handleGenerate = async () => {
    if (!topic.trim()) {
      alert('请输入主题')
      return
    }

    setLoading(true)
    setGeneratedContent('')

    try {
      const response = await aiContentGenerate({
        topic: topic.trim(),
        style,
        platform,
        length,
      })

      if (response.success && response.data) {
        setGeneratedContent(response.data.content)
      } else {
        alert(response.error || '生成失败')
      }
    } catch (error) {
      console.error('Generate failed:', error)
      alert('生成失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  const handleRewrite = async () => {
    if (!generatedContent.trim()) {
      alert('没有内容可以改写')
      return
    }

    setRewriting(true)

    try {
      const response = await aiContentRewrite({
        content: generatedContent,
        style: style === '轻松幽默' ? '正式专业' : '轻松幽默',
        platform,
      })

      if (response.success && response.data) {
        setGeneratedContent(response.data.content)
      } else {
        alert(response.error || '改写失败')
      }
    } catch (error) {
      console.error('Rewrite failed:', error)
      alert('改写失败，请重试')
    } finally {
      setRewriting(false)
    }
  }

  const handleCopy = async () => {
    if (!generatedContent) return

    try {
      await navigator.clipboard.writeText(generatedContent)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (error) {
      console.error('Copy failed:', error)
    }
  }

  const handlePublish = async () => {
    if (!generatedContent.trim()) {
      alert('没有内容可以发布')
      return
    }

    setPublishing(true)

    try {
      // 根据选择的平台确定内容类型
      let contentType = 'text'
      if (platform === 'douyin' || platform === 'xiaohongshu') {
        contentType = 'images'
      } else if (platform === 'toutiao') {
        contentType = 'article'
      }

      const response = await publish({
        platform: (platform === 'general' ? 'toutiao' : platform) as Platform,
        type: contentType as ContentType,
        title: topic || 'AI生成内容',
        body: generatedContent,
        tags: [topic, style, 'AI生成'].filter(Boolean)
      })

      if (response.success) {
        alert(`内容已成功发布到${getPlatformName(platform)}!`)
        // 跳转到发布历史页面
        navigate('/history')
      } else {
        alert(response.error || '发布失败')
      }
    } catch (error) {
      console.error('Publish failed:', error)
      alert('发布失败，请重试')
    } finally {
      setPublishing(false)
    }
  }

  const getPlatformName = (platformCode: string) => {
    const platformMap: Record<string, string> = {
      'douyin': '抖音',
      'toutiao': '今日头条',
      'xiaohongshu': '小红书',
      'weibo': '微博',
      'general': '通用平台'
    }
    return platformMap[platformCode] || platformCode
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">AI 内容生成</h1>
        <p className="text-muted-foreground mt-2">输入主题，AI 为你生成高质量内容</p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>生成设置</CardTitle>
            <CardDescription>配置内容生成的参数</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label htmlFor="topic">主题 / 关键词</Label>
              <Textarea
                id="topic"
                value={topic}
                onChange={(e) => setTopic(e.target.value)}
                placeholder="输入你想要生成的主题或关键词，例如：AI 技术发展趋势"
                className="mt-1 min-h-20"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>风格</Label>
                <Select value={style} onValueChange={setStyle}>
                  <SelectTrigger className="mt-1">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {styles.map((s) => (
                      <SelectItem key={s.value} value={s.value}>
                        {s.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div>
                <Label>平台</Label>
                <Select value={platform} onValueChange={setPlatform}>
                  <SelectTrigger className="mt-1">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {platforms.map((p) => (
                      <SelectItem key={p.value} value={p.value}>
                        {p.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>

            <div>
              <Label htmlFor="length">字数要求</Label>
              <Input
                id="length"
                type="number"
                value={length}
                onChange={(e) => setLength(parseInt(e.target.value) || 500)}
                className="mt-1"
              />
            </div>

            <div className="flex gap-2">
              <Button 
                onClick={handleGenerate} 
                disabled={loading || !topic.trim()}
                className="flex-1"
              >
                {loading ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    生成中...
                  </>
                ) : (
                  <>
                    <Sparkles className="mr-2 h-4 w-4" />
                    生成内容
                  </>
                )}
              </Button>
              {generatedContent && (
                <Button 
                  onClick={handlePublish} 
                  disabled={publishing}
                  className="bg-blue-600 hover:bg-blue-700"
                >
                  {publishing ? (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  ) : (
                    <Upload className="mr-2 h-4 w-4" />
                  )}
                  发布
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>生成结果</CardTitle>
                <CardDescription>AI 生成的内容将显示在这里</CardDescription>
              </div>
              {generatedContent && (
                <div className="flex gap-2">
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={handleRewrite} 
                    disabled={rewriting}
                    title="改写内容"
                  >
                    {rewriting ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <RefreshCw className="h-4 w-4" />
                    )}
                  </Button>
                  <Button 
                    variant="outline" 
                    size="sm" 
                    onClick={handleCopy}
                    title="复制内容"
                  >
                    {copied ? (
                      <CheckCircle className="h-4 w-4 text-green-500" />
                    ) : (
                      <Copy className="h-4 w-4" />
                    )}
                  </Button>
                  <Button 
                    variant="default" 
                    size="sm" 
                    onClick={handlePublish} 
                    disabled={publishing || !generatedContent.trim()}
                    className="bg-blue-600 hover:bg-blue-700"
                    title="一键发布"
                  >
                    {publishing ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Send className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent>
            {generatedContent ? (
              <div className="whitespace-pre-wrap rounded-lg bg-muted p-4 text-sm">
                {generatedContent}
              </div>
            ) : (
              <div className="flex h-40 items-center justify-center text-muted-foreground">
                <div className="text-center">
                  <Sparkles className="mx-auto h-8 w-8 mb-2 opacity-50" />
                  <p>输入主题并点击生成</p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Card className="mt-6">
        <CardHeader>
          <CardTitle>使用说明</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3 text-sm text-muted-foreground">
            <div>
              <h4 className="font-medium text-foreground mb-1">1. 输入主题</h4>
              <p>描述你想要生成的内容主题或关键词</p>
            </div>
            <div>
              <h4 className="font-medium text-foreground mb-1">2. 选择风格</h4>
              <p>根据目标平台选择合适的内容风格</p>
            </div>
            <div>
              <h4 className="font-medium text-foreground mb-1">3. 生成内容</h4>
              <p>AI 将根据设置生成高质量内容</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

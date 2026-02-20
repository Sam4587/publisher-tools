import { useState } from 'react'
import { X, Sparkles, Loader2, Copy, Check, Download, MessageSquare, Brain, TrendingUp, AlertCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { analyzeHotTopics, generateHotTopicsBrief } from '@/lib/api'
import type { HotTopic, AIAnalysisResult } from '@/types/api'

interface EnhancedAIAnalysisPanelProps {
  topics: HotTopic[]
  onClose: () => void
}

interface AnalysisSection {
  title: string
  icon: React.ReactNode
  content: string
  type: 'summary' | 'keypoints' | 'sentiment' | 'trends' | 'recommendations'
}

export default function EnhancedAIAnalysisPanel({ topics, onClose }: EnhancedAIAnalysisPanelProps) {
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<AIAnalysisResult | null>(null)
  const [brief, setBrief] = useState<string>('')
  const [copied, setCopied] = useState(false)
  const [mode, setMode] = useState<'analyze' | 'brief'>('analyze')
  const [activeSection, setActiveSection] = useState<string>('all')

  async function handleAnalyze() {
    setLoading(true)
    try {
      const res = await analyzeHotTopics(topics, { focus: 'important' })
      if (res.success && res.data) {
        setResult(res.data)
        setBrief('')
      }
    } finally {
      setLoading(false)
    }
  }

  async function handleGenerateBrief() {
    setLoading(true)
    try {
      const res = await generateHotTopicsBrief(topics, 500)
      if (res.success && res.data) {
        setBrief(res.data.brief)
        setResult(null)
      }
    } finally {
      setLoading(false)
    }
  }

  async function handleCopy() {
    const text = mode === 'analyze' && result
      ? formatAnalysisText(result)
      : brief

    if (text) {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  async function handleDownload() {
    const text = mode === 'analyze' && result
      ? formatAnalysisText(result)
      : brief

    if (text) {
      const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `ai-analysis-${new Date().toISOString().slice(0, 10)}.txt`
      a.click()
      URL.revokeObjectURL(url)
    }
  }

  function formatAnalysisText(result: AIAnalysisResult): string {
    let text = `AI 分析结果\n${'='.repeat(40)}\n\n`
    text += `摘要:\n${result.summary}\n\n`
    if (result.keyPoints && result.keyPoints.length > 0) {
      text += `关键点:\n${result.keyPoints.map((p, i) => `${i + 1}. ${p}`).join('\n')}\n\n`
    }
    if (result.sentiment) {
      text += `情感分析: ${result.sentiment === 'positive' ? '正面' : result.sentiment === 'negative' ? '负面' : '中性'}\n`
    }
    return text
  }

  const getSections = (): AnalysisSection[] => {
    if (!result) return []

    const sections: AnalysisSection[] = []

    if (result.summary) {
      sections.push({
        title: '摘要',
        icon: <MessageSquare className="h-4 w-4" />,
        content: result.summary,
        type: 'summary'
      })
    }

    if (result.keyPoints && result.keyPoints.length > 0) {
      sections.push({
        title: '关键点',
        icon: <Brain className="h-4 w-4" />,
        content: result.keyPoints.map((p, i) => `${i + 1}. ${p}`).join('\n'),
        type: 'keypoints'
      })
    }

    if (result.sentiment) {
      const sentimentText = result.sentiment === 'positive' ? '正面' : result.sentiment === 'negative' ? '负面' : '中性'
      sections.push({
        title: '情感分析',
        icon: result.sentiment === 'positive' ? <TrendingUp className="h-4 w-4 text-green-600" /> :
              result.sentiment === 'negative' ? <AlertCircle className="h-4 w-4 text-red-600" /> :
              <MessageSquare className="h-4 w-4 text-gray-600" />,
        content: sentimentText,
        type: 'sentiment'
      })
    }

    return sections
  }

  const filteredSections = activeSection === 'all' ? getSections() : getSections().filter(s => s.type === activeSection)

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-4xl max-h-[90vh] overflow-hidden">
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-purple-500" />
            AI 智能分析
            <Badge variant="secondary" className="ml-2">
              {topics.length} 个话题
            </Badge>
          </CardTitle>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" onClick={handleCopy} title="复制">
              {copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
            </Button>
            <Button variant="ghost" size="icon" onClick={handleDownload} title="下载">
              <Download className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </CardHeader>

        <CardContent className="space-y-4 overflow-y-auto max-h-[calc(90vh-100px)]">
          {/* 话题预览 */}
          <div className="flex flex-wrap gap-2 p-3 bg-gradient-to-r from-purple-50 to-blue-50 rounded-lg border">
            {topics.slice(0, 6).map((topic) => (
              <Badge key={topic._id} variant="outline" className="bg-white truncate max-w-32">
                {topic.title}
              </Badge>
            ))}
            {topics.length > 6 && (
              <Badge variant="secondary">+{topics.length - 6} 更多...</Badge>
            )}
          </div>

          {/* 操作按钮 */}
          <div className="flex gap-2">
            <Button
              variant={mode === 'analyze' ? 'default' : 'outline'}
              onClick={() => {
                setMode('analyze')
                handleAnalyze()
              }}
              disabled={loading}
              className="flex-1"
            >
              {loading && mode === 'analyze' ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : (
                <Sparkles className="h-4 w-4 mr-2" />
              )}
              深度分析
            </Button>
            <Button
              variant={mode === 'brief' ? 'default' : 'outline'}
              onClick={() => {
                setMode('brief')
                handleGenerateBrief()
              }}
              disabled={loading}
              className="flex-1"
            >
              {loading && mode === 'brief' ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : null}
              生成简报
            </Button>
          </div>

          {/* 分析结果 */}
          {loading ? (
            <div className="flex flex-col items-center justify-center py-16 space-y-4">
              <Loader2 className="h-12 w-12 animate-spin text-purple-500" />
              <div className="text-center space-y-2">
                <p className="text-lg font-medium">AI 正在分析中...</p>
                <p className="text-sm text-muted-foreground">这可能需要几秒钟</p>
              </div>
            </div>
          ) : result ? (
            <div className="space-y-4">
              {/* 分析类型选择 */}
              <Tabs defaultValue="all" value={activeSection} onValueChange={setActiveSection}>
                <TabsList className="grid w-full grid-cols-4">
                  <TabsTrigger value="all" className="flex items-center gap-2">
                    <MessageSquare className="h-4 w-4" />
                    全部
                  </TabsTrigger>
                  <TabsTrigger value="summary">摘要</TabsTrigger>
                  <TabsTrigger value="keypoints">关键点</TabsTrigger>
                  <TabsTrigger value="sentiment">情感</TabsTrigger>
                </TabsList>

                <div className="mt-4 space-y-3">
                  {filteredSections.map((section, index) => (
                    <div key={index} className="border rounded-lg p-4 bg-white hover:shadow-md transition-shadow">
                      <div className="flex items-center gap-2 mb-3">
                        <div className="text-purple-600">{section.icon}</div>
                        <h4 className="font-semibold">{section.title}</h4>
                      </div>
                      {section.type === 'sentiment' ? (
                        <Badge
                          variant={result.sentiment === 'positive' ? 'default' :
                                  result.sentiment === 'negative' ? 'destructive' : 'secondary'}
                          className="text-sm"
                        >
                          {section.content}
                        </Badge>
                      ) : section.type === 'keypoints' ? (
                        <ul className="space-y-2">
                          {result.keyPoints?.map((point, i) => (
                            <li key={i} className="flex items-start gap-2">
                              <span className="bg-purple-100 text-purple-600 rounded-full w-5 h-5 flex items-center justify-center text-xs flex-shrink-0 mt-0.5">
                                {i + 1}
                              </span>
                              <span className="text-gray-700">{point}</span>
                            </li>
                          ))}
                        </ul>
                      ) : (
                        <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">{section.content}</p>
                      )}
                    </div>
                  ))}
                </div>
              </Tabs>

              {/* 统计信息 */}
              {result.keyPoints && (
                <div className="flex items-center justify-between p-4 bg-gradient-to-r from-blue-50 to-purple-50 rounded-lg border">
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <span>分析话题: {topics.length} 个</span>
                    <span>•</span>
                    <span>关键点: {result.keyPoints.length} 个</span>
                    <span>•</span>
                    <span>情感: {result.sentiment === 'positive' ? '正面' : result.sentiment === 'negative' ? '负面' : '中性'}</span>
                  </div>
                  <Button variant="outline" size="sm" onClick={handleCopy}>
                    <Copy className="h-4 w-4 mr-2" />
                    复制全部
                  </Button>
                </div>
              )}
            </div>
          ) : brief ? (
            <div className="space-y-4">
              <div className="border rounded-lg p-4 bg-white">
                <div className="flex items-center gap-2 mb-3">
                  <Sparkles className="h-4 w-4 text-purple-600" />
                  <h4 className="font-semibold">热点简报</h4>
                </div>
                <Textarea
                  value={brief}
                  readOnly
                  className="min-h-[300px] resize-none"
                />
              </div>
              <Button variant="outline" size="sm" onClick={handleCopy}>
                {copied ? (
                  <Check className="h-4 w-4 mr-2 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4 mr-2" />
                )}
                复制简报
              </Button>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-16 space-y-4 text-muted-foreground">
              <Sparkles className="h-16 w-16 text-purple-200" />
              <div className="text-center space-y-2">
                <p className="text-lg font-medium">点击上方按钮开始 AI 分析</p>
                <p className="text-sm">深度分析或生成简报</p>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

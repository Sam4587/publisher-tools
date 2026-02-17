import { useState } from 'react'
import { X, Sparkles, Loader2, Copy, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
// import { Textarea } from '@/components/ui/textarea'
import { analyzeHotTopics, generateHotTopicsBrief } from '@/lib/api'
import type { HotTopic, AIAnalysisResult } from '@/types/api'

interface AIAnalysisPanelProps {
  topics: HotTopic[]
  onClose: () => void
}

export default function AIAnalysisPanel({ topics, onClose }: AIAnalysisPanelProps) {
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<AIAnalysisResult | null>(null)
  const [brief, setBrief] = useState<string>('')
  const [copied, setCopied] = useState(false)
  const [mode, setMode] = useState<'analyze' | 'brief'>('analyze')

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
      ? `${result.summary}\n\n关键点:\n${result.keyPoints.map((p, i) => `${i + 1}. ${p}`).join('\n')}`
      : brief

    if (text) {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-2xl max-h-[80vh] overflow-hidden">
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-purple-500" />
            AI 分析
            <span className="text-sm font-normal text-gray-500">
              ({topics.length} 个话题)
            </span>
          </CardTitle>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* 话题预览 */}
          <div className="flex flex-wrap gap-2 p-3 bg-gray-50 rounded-lg">
            {topics.slice(0, 5).map((topic) => (
              <span
                key={topic._id}
                className="text-xs bg-white px-2 py-1 rounded border truncate max-w-32"
                title={topic.title}
              >
                {topic.title}
              </span>
            ))}
            {topics.length > 5 && (
              <span className="text-xs text-gray-500">
                +{topics.length - 5} 更多...
              </span>
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
            >
              {loading && mode === 'brief' ? (
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              ) : null}
              生成简报
            </Button>
          </div>

          {/* 结果展示 */}
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-purple-500" />
              <span className="ml-2 text-gray-500">AI 正在分析中...</span>
            </div>
          ) : result ? (
            <div className="space-y-4">
              <div>
                <h4 className="font-medium mb-2">摘要</h4>
                <p className="text-gray-600 bg-gray-50 p-3 rounded-lg">
                  {result.summary}
                </p>
              </div>

              {result.keyPoints && result.keyPoints.length > 0 && (
                <div>
                  <h4 className="font-medium mb-2">关键点</h4>
                  <ul className="space-y-2">
                    {result.keyPoints.map((point, index) => (
                      <li key={index} className="flex items-start gap-2">
                        <span className="bg-purple-100 text-purple-600 rounded-full w-5 h-5 flex items-center justify-center text-xs flex-shrink-0">
                          {index + 1}
                        </span>
                        <span className="text-gray-600">{point}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              {result.sentiment && (
                <div>
                  <h4 className="font-medium mb-2">情感分析</h4>
                  <span
                    className={`inline-block px-3 py-1 rounded-full text-sm ${
                      result.sentiment === 'positive'
                        ? 'bg-green-100 text-green-600'
                        : result.sentiment === 'negative'
                        ? 'bg-red-100 text-red-600'
                        : 'bg-gray-100 text-gray-600'
                    }`}
                  >
                    {result.sentiment === 'positive'
                      ? '正面'
                      : result.sentiment === 'negative'
                      ? '负面'
                      : '中性'}
                  </span>
                </div>
              )}

              <Button variant="outline" size="sm" onClick={handleCopy}>
                {copied ? (
                  <Check className="h-4 w-4 mr-1 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4 mr-1" />
                )}
                复制结果
              </Button>
            </div>
          ) : brief ? (
            <div className="space-y-4">
              <div>
                <h4 className="font-medium mb-2">热点简报</h4>
                <div className="bg-gray-50 p-3 rounded-lg whitespace-pre-wrap text-gray-600">
                  {brief}
                </div>
              </div>
              <Button variant="outline" size="sm" onClick={handleCopy}>
                {copied ? (
                  <Check className="h-4 w-4 mr-1 text-green-500" />
                ) : (
                  <Copy className="h-4 w-4 mr-1" />
                )}
                复制简报
              </Button>
            </div>
          ) : (
            <div className="text-center text-gray-500 py-8">
              点击上方按钮开始分析
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

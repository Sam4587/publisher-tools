import { useState, useEffect } from 'react'
import { X, TrendingUp, BarChart3, Clock } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getHotTopicTrend, getCrossPlatformAnalysis } from '@/lib/api'
import type { HotTopic, TrendDataPoint, CrossPlatformAnalysis } from '@/types/api'

interface TrendTimelineProps {
  topic: HotTopic
  onClose: () => void
}

// 简单的图表组件
function SimpleChart({ data }: { data: TrendDataPoint[] }) {
  if (!data || data.length === 0) {
    return <div className="text-center text-gray-500 py-8">暂无趋势数据</div>
  }

  const maxHeat = Math.max(...data.map(d => d.heat), 1)

  return (
    <div className="space-y-2">
      {data.map((point, index) => (
        <div key={index} className="flex items-center gap-2">
          <span className="text-xs text-gray-500 w-20">{point.date}</span>
          <div className="flex-1 bg-gray-100 rounded-full h-4 overflow-hidden">
            <div
              className="bg-blue-500 h-full transition-all duration-300"
              style={{ width: `${(point.heat / maxHeat) * 100}%` }}
            />
          </div>
          <span className="text-xs font-medium w-8">{point.heat}</span>
        </div>
      ))}
    </div>
  )
}

// 跨平台分析组件
function PlatformAnalysis({ data }: { data: CrossPlatformAnalysis | null }) {
  if (!data) {
    return <div className="text-center text-gray-500 py-4">暂无跨平台数据</div>
  }

  return (
    <div className="space-y-3">
      <div className="text-sm text-gray-600 mb-2">
        总热度: <span className="font-bold text-lg">{data.totalHeat}</span>
      </div>
      {data.platforms.map((platform, index) => (
        <div key={index} className="flex items-center gap-2 p-2 bg-gray-50 rounded">
          <span className="font-medium text-sm flex-1">{platform.name}</span>
          <span className="text-xs text-gray-500">排名 #{platform.rank}</span>
          <span className="text-sm font-bold text-blue-600">{platform.heat}</span>
        </div>
      ))}
    </div>
  )
}

export default function TrendTimeline({ topic, onClose }: TrendTimelineProps) {
  const [trendData, setTrendData] = useState<TrendDataPoint[]>([])
  const [crossPlatform, setCrossPlatform] = useState<CrossPlatformAnalysis | null>(null)
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'trend' | 'platform'>('trend')

  useEffect(() => {
    async function loadData() {
      setLoading(true)
      try {
        // 加载趋势数据
        const trendResult = await getHotTopicTrend(topic._id, 7)
        if (trendResult.success && trendResult.data) {
          setTrendData(trendResult.data.trend || [])
        }

        // 加载跨平台分析
        const crossResult = await getCrossPlatformAnalysis(topic.title)
        if (crossResult.success && crossResult.data) {
          setCrossPlatform(crossResult.data)
        }
      } finally {
        setLoading(false)
      }
    }
    loadData()
  }, [topic._id, topic.title])

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <Card className="w-full max-w-2xl max-h-[80vh] overflow-hidden">
        <CardHeader className="flex flex-row items-center justify-between">
          <div className="flex-1 pr-4">
            <CardTitle className="text-lg line-clamp-2">{topic.title}</CardTitle>
            <div className="flex items-center gap-2 mt-2">
              <span className="text-sm text-gray-500">{topic.source}</span>
              <span className="text-sm font-bold text-blue-600">热度: {topic.heat}</span>
            </div>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </CardHeader>

        <CardContent className="space-y-4">
          {/* 标签切换 */}
          <div className="flex gap-2">
            <Button
              variant={activeTab === 'trend' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setActiveTab('trend')}
            >
              <TrendingUp className="h-4 w-4 mr-1" />
              趋势变化
            </Button>
            <Button
              variant={activeTab === 'platform' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setActiveTab('platform')}
            >
              <BarChart3 className="h-4 w-4 mr-1" />
              跨平台分析
            </Button>
          </div>

          {/* 内容区域 */}
          {loading ? (
            <div className="space-y-3">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="animate-pulse flex items-center gap-2">
                  <div className="h-4 bg-gray-200 rounded w-20" />
                  <div className="h-4 bg-gray-200 rounded flex-1" />
                  <div className="h-4 bg-gray-200 rounded w-8" />
                </div>
              ))}
            </div>
          ) : activeTab === 'trend' ? (
            <SimpleChart data={trendData} />
          ) : (
            <PlatformAnalysis data={crossPlatform} />
          )}

          {/* 时间信息 */}
          {topic.publishedAt && (
            <div className="flex items-center gap-2 text-sm text-gray-500 pt-4 border-t">
              <Clock className="h-4 w-4" />
              发布时间: {new Date(topic.publishedAt).toLocaleString()}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

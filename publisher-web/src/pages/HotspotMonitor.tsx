import { useState, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { RefreshCw, TrendingUp, Sparkles, Video, Flame } from 'lucide-react'
import HotspotTrendChart from '@/components/HotspotTrendChart'
import RankTimelineChart from '@/components/RankTimelineChart'
import EnhancedAIAnalysisPanel from '@/components/EnhancedAIAnalysisPanel'
import VideoProcessingProgress from '@/components/VideoProcessingProgress'
import GlobalFilterBar from '@/components/GlobalFilterBar'
import { getHotTopics } from '@/lib/api'
import type { HotTopic } from '@/types/api'

interface TrendDataPoint {
  date: string
  value: number
}

interface RankData {
  date: string
  rank: number
  topicId: string
  topicTitle: string
}

interface FilterOptions {
  keyword?: string
  platform?: string
  category?: string
  dateRange?: string
  sortBy?: string
  trend?: string
}

export default function HotspotMonitor() {
  const [topics, setTopics] = useState<HotTopic[]>([])
  const [filteredTopics, setFilteredTopics] = useState<HotTopic[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [selectedTopics, setSelectedTopics] = useState<HotTopic[]>([])
  const [showAIPanel, setShowAIPanel] = useState(false)
  const [filters, setFilters] = useState<FilterOptions>({})

  const fetchTopics = async () => {
    try {
      const response = await getHotTopics()
      if (response.success && response.data) {
        const topicsData = Array.isArray(response.data) ? response.data : (response.data as any).topics || []
        setTopics(topicsData)
        setFilteredTopics(topicsData)
      }
    } catch (error) {
      console.error('获取热点话题失败:', error)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchTopics()
  }, [])

  // 应用筛选条件
  useEffect(() => {
    let result = [...topics]

    // 按分类筛选
    if (filters.category && filters.category !== 'all') {
      result = result.filter(topic => topic.category === filters.category)
    }

    // 按平台筛选
    if (filters.platform && filters.platform !== 'all') {
      result = result.filter(topic => topic.source === filters.platform)
    }

    // 按关键词筛选
    if (filters.keyword) {
      const keyword = filters.keyword.toLowerCase()
      result = result.filter(topic => 
        topic.title.toLowerCase().includes(keyword) ||
        (topic.description && topic.description.toLowerCase().includes(keyword)) ||
        (topic.keywords && topic.keywords.some(k => k.toLowerCase().includes(keyword)))
      )
    }

    // 按趋势筛选
    if (filters.trend && filters.trend !== 'all') {
      result = result.filter(topic => topic.trend === filters.trend)
    }

    // 排序
    if (filters.sortBy) {
      switch (filters.sortBy) {
        case 'heat':
          result.sort((a, b) => b.heat - a.heat)
          break
        case 'date':
          result.sort((a, b) => new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime())
          break
        case 'suitability':
          result.sort((a, b) => (b.suitability || 0) - (a.suitability || 0))
          break
      }
    }

    setFilteredTopics(result)
  }, [topics, filters])

  const handleRefresh = () => {
    setRefreshing(true)
    fetchTopics()
  }

  const handleFilterChange = (newFilters: FilterOptions) => {
    setFilters(newFilters)
  }

  const handleSelectTopic = (topic: HotTopic) => {
    const exists = selectedTopics.find(t => t._id === topic._id)
    if (exists) {
      setSelectedTopics(prev => prev.filter(t => t._id !== topic._id))
    } else {
      setSelectedTopics(prev => [...prev, topic])
    }
  }

  const handleSelectAll = () => {
    if (selectedTopics.length === filteredTopics.length) {
      setSelectedTopics([])
    } else {
      setSelectedTopics(filteredTopics)
    }
  }

  const handleAnalyze = () => {
    if (selectedTopics.length > 0) {
      setShowAIPanel(true)
    }
  }

  // 生成模拟趋势数据
  const getTrendData = (): TrendDataPoint[] => {
    const data: TrendDataPoint[] = []
    const now = new Date()
    for (let i = 6; i >= 0; i--) {
      const date = new Date(now.getTime() - i * 24 * 60 * 60 * 1000)
      data.push({
        date: date.toISOString().split('T')[0],
        value: Math.floor(Math.random() * 50000) + 10000,
      })
    }
    return data
  }

  // 生成模拟排名数据
  const getRankData = (): RankData[] => {
    const data: RankData[] = []
    const now = new Date()
    topics.slice(0, 5).forEach((topic) => {
      for (let i = 6; i >= 0; i--) {
        const date = new Date(now.getTime() - i * 24 * 60 * 60 * 1000)
        data.push({
          date: date.toISOString().split('T')[0],
          rank: Math.floor(Math.random() * 10) + 1,
          topicId: topic._id,
          topicTitle: topic.title,
        })
      }
    })
    return data
  }

  const trendData = getTrendData()
  const rankData = getRankData()

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8 space-y-6">
      {/* 页面头部 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">热点监控</h1>
          <p className="text-muted-foreground mt-2">实时追踪全网热点话题</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleRefresh}
            disabled={refreshing}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            刷新
          </Button>
          <Button
            onClick={handleAnalyze}
            disabled={selectedTopics.length === 0}
          >
            <Sparkles className="h-4 w-4 mr-2" />
            AI 分析 ({selectedTopics.length})
          </Button>
        </div>
      </div>

      {/* 全局筛选 */}
      <GlobalFilterBar
        onFilterChange={handleFilterChange}
        platforms={['weibo', 'douyin', 'zhihu', 'baidu']}
        categories={['news', 'entertainment', 'sports', 'tech', 'finance']}
        placeholder="搜索热点话题..."
      />

      {/* 主要内容 */}
      <Tabs defaultValue="trends" className="space-y-6">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="trends" className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4" />
            趋势分析
          </TabsTrigger>
          <TabsTrigger value="ranking" className="flex items-center gap-2">
            <Flame className="h-4 w-4" />
            排名追踪
          </TabsTrigger>
          <TabsTrigger value="topics" className="flex items-center gap-2">
            <Sparkles className="h-4 w-4" />
            热点话题
          </TabsTrigger>
          <TabsTrigger value="video" className="flex items-center gap-2">
            <Video className="h-4 w-4" />
            视频处理
          </TabsTrigger>
        </TabsList>

        {/* 趋势分析 */}
        <TabsContent value="trends" className="space-y-6">
          <HotspotTrendChart
            data={trendData}
            title="热点趋势分析"
            description="查看热点话题的总体热度变化"
          />
        </TabsContent>

        {/* 排名追踪 */}
        <TabsContent value="ranking" className="space-y-6">
          <RankTimelineChart
            data={rankData}
            title="排名时间线"
            description="追踪热点话题的排名变化轨迹"
          />
        </TabsContent>

        {/* 热点话题 */}
        <TabsContent value="topics" className="space-y-4">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={handleSelectAll}>
                {selectedTopics.length === filteredTopics.length ? '取消全选' : '全选'}
              </Button>
              <Badge variant="secondary">
                已选择 {selectedTopics.length} 个话题
              </Badge>
            </div>
          </div>

          <div className="grid gap-4">
            {filteredTopics.map((topic, index) => (
              <Card
                key={topic._id}
                className={`cursor-pointer transition-all hover:shadow-md ${
                  selectedTopics.find(t => t._id === topic._id)
                    ? 'border-blue-500 bg-blue-50'
                    : ''
                }`}
                onClick={() => handleSelectTopic(topic)}
              >
                <CardContent className="p-4">
                  <div className="flex items-start gap-4">
                    <div className="flex-shrink-0">
                      <div
                        className={`w-10 h-10 rounded-full flex items-center justify-center font-bold ${
                          index < 3
                            ? 'bg-yellow-100 text-yellow-700'
                            : 'bg-gray-100 text-gray-600'
                        }`}
                      >
                        {index + 1}
                      </div>
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="font-medium line-clamp-2">{topic.title}</div>
                      <div className="flex items-center gap-2 mt-2 text-sm text-muted-foreground">
                        <span>{topic.source}</span>
                        <span>•</span>
                        <span className="text-blue-600 font-medium">
                          热度: {topic.heat?.toLocaleString()}
                        </span>
                        {topic.publishedAt && (
                          <>
                            <span>•</span>
                            <span>
                              {new Date(topic.publishedAt).toLocaleDateString()}
                            </span>
                          </>
                        )}
                      </div>
                    </div>
                    {selectedTopics.find(t => t._id === topic._id) && (
                      <Badge variant="default" className="flex-shrink-0">
                        已选择
                      </Badge>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>

        {/* 视频处理 */}
        <TabsContent value="video" className="space-y-6">
          <VideoProcessingProgress />
        </TabsContent>
      </Tabs>

      {/* AI 分析面板 */}
      {showAIPanel && (
        <EnhancedAIAnalysisPanel
          topics={selectedTopics}
          onClose={() => setShowAIPanel(false)}
        />
      )}
    </div>
  )
}

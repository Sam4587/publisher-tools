import { useState, useEffect } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { RefreshCw, TrendingUp, Sparkles, Video, Flame, AlertCircle } from 'lucide-react'
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

// 后端数据字段映射
// GlobalFilterBar 传递的是英文，后端 category 是中文，需要映射
const categoryMapping: Record<string, string> = {
  'news': '新闻',
  'entertainment': '娱乐',
  'sports': '体育',
  'tech': '科技',
  'finance': '财经',
}

// GlobalFilterBar 传递的是英文，后端 source 也是英文，不需要映射
// 但为了显示，需要反向映射
const platformDisplayNames: Record<string, string> = {
  'weibo': '微博',
  'douyin': '抖音',
  'toutiao': '今日头条',
  'zhihu': '知乎',
  'bilibili': 'B站',
  'xiaohongshu': '小红书',
}

export default function HotspotMonitor() {
  const [topics, setTopics] = useState<HotTopic[]>([])
  const [filteredTopics, setFilteredTopics] = useState<HotTopic[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [selectedTopics, setSelectedTopics] = useState<HotTopic[]>([])
  const [showAIPanel, setShowAIPanel] = useState(false)
  const [filters, setFilters] = useState<FilterOptions>({})
  const [error, setError] = useState<string | null>(null)

  const fetchTopics = async () => {
    try {
      setError(null)
      const response = await getHotTopics()
      console.log('API响应:', response)

      if (response.success && response.data) {
        // 处理多种可能的响应格式
        let topicsData: HotTopic[] = []

        if (Array.isArray(response.data)) {
          topicsData = response.data
        } else if ((response.data as any).topics && Array.isArray((response.data as any).topics)) {
          topicsData = (response.data as any).topics
        } else if ((response.data as any).data && Array.isArray((response.data as any).data)) {
          topicsData = (response.data as any).data
        } else {
          console.warn('未知的响应格式:', response.data)
        }

        console.log('解析后的热点数据:', topicsData)

        if (topicsData.length > 0) {
          setTopics(topicsData)
          setFilteredTopics(topicsData)
        } else {
          console.warn('热点数据为空')
          setTopics([])
          setFilteredTopics([])
        }
      } else {
        console.error('API返回失败:', response)
        setError(response.message || '获取热点数据失败')
        setTopics([])
        setFilteredTopics([])
      }
    } catch (error) {
      console.error('获取热点话题失败:', error)
      setError('网络请求失败，请检查后端服务是否正常运行')
      setTopics([])
      setFilteredTopics([])
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchTopics()
    // 设置自动刷新，每5分钟刷新一次
    const interval = setInterval(() => {
      fetchTopics()
    }, 5 * 60 * 1000) // 5分钟

    return () => clearInterval(interval)
  }, [])

  // 应用筛选条件
  useEffect(() => {
    console.log('开始应用筛选条件:', filters)
    console.log('原始数据数量:', topics.length)

    let result = [...topics]

    // 按分类筛选 - GlobalFilterBar 传递的是英文，需要转换为中文匹配后端数据
    if (filters.category && filters.category !== 'all') {
      const targetCategory = categoryMapping[filters.category] || filters.category
      console.log(`分类筛选: ${filters.category} -> ${targetCategory}`)
      result = result.filter(topic => {
        const match = topic.category === targetCategory
        console.log(`话题 "${topic.title}" 的分类 "${topic.category}" 是否匹配 "${targetCategory}": ${match}`)
        return match
      })
      console.log(`分类筛选后结果数量: ${result.length}`)
    }

    // 按平台筛选 - GlobalFilterBar 传递的是英文，后端 source 也是英文，直接匹配
    if (filters.platform && filters.platform !== 'all') {
      console.log(`平台筛选: ${filters.platform}`)
      result = result.filter(topic => {
        const match = topic.source === filters.platform
        console.log(`话题 "${topic.title}" 的来源 "${topic.source}" 是否匹配 "${filters.platform}": ${match}`)
        return match
      })
      console.log(`平台筛选后结果数量: ${result.length}`)
    }

    // 按关键词筛选
    if (filters.keyword) {
      const keyword = filters.keyword.toLowerCase()
      console.log(`关键词筛选: ${keyword}`)
      result = result.filter(topic =>
        topic.title.toLowerCase().includes(keyword) ||
        (topic.description && topic.description.toLowerCase().includes(keyword)) ||
        (topic.keywords && topic.keywords.some(k => k.toLowerCase().includes(keyword)))
      )
      console.log(`关键词筛选后结果数量: ${result.length}`)
    }

    // 按趋势筛选
    if (filters.trend && filters.trend !== 'all') {
      console.log(`趋势筛选: ${filters.trend}`)
      result = result.filter(topic => topic.trend === filters.trend)
      console.log(`趋势筛选后结果数量: ${result.length}`)
    }

    // 排序
    if (filters.sortBy) {
      console.log(`排序方式: ${filters.sortBy}`)
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

    console.log('最终筛选结果数量:', result.length)
    setFilteredTopics(result)
  }, [topics, filters])

  const handleRefresh = () => {
    setRefreshing(true)
    fetchTopics()
  }

  const handleFilterChange = (newFilters: FilterOptions) => {
    console.log('筛选条件变化:', newFilters)
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

      {/* 错误提示 */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-3">
          <AlertCircle className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
          <div className="flex-1">
            <h3 className="font-medium text-red-800">数据加载失败</h3>
            <p className="text-sm text-red-700 mt-1">{error}</p>
            <Button
              variant="outline"
              size="sm"
              className="mt-2"
              onClick={() => {
                setError(null)
                setLoading(true)
                fetchTopics()
              }}
            >
              重试
            </Button>
          </div>
        </div>
      )}

      {/* 数据为空提示 */}
      {!error && topics.length === 0 && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-6 text-center">
          <AlertCircle className="h-12 w-12 text-yellow-600 mx-auto mb-3" />
          <h3 className="font-medium text-yellow-800 mb-2">暂无热点数据</h3>
          <p className="text-sm text-yellow-700 mb-4">当前没有可用的热点话题数据，请点击刷新按钮获取最新数据</p>
          <Button
            variant="outline"
            onClick={handleRefresh}
            disabled={refreshing}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            立即刷新
          </Button>
        </div>
      )}

      {/* 全局筛选 */}
      <GlobalFilterBar
        onFilterChange={handleFilterChange}
        platforms={['weibo', 'douyin', 'zhihu', 'bilibili']}
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
              {filteredTopics.length > 0 && (
                <Badge variant="outline">
                  共 {filteredTopics.length} 个话题
                </Badge>
              )}
            </div>
          </div>

          {filteredTopics.length === 0 ? (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-8 text-center">
              <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-3" />
              <h3 className="font-medium text-gray-800 mb-2">没有找到匹配的话题</h3>
              <p className="text-sm text-gray-600">请尝试调整筛选条件或刷新数据</p>
            </div>
          ) : (
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
                          <span>{platformDisplayNames[topic.source] || topic.source}</span>
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
          )}
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

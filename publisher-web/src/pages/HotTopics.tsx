import { useState, useEffect } from 'react'
import { RefreshCw, Search, TrendingUp, ExternalLink, Sparkles, BarChart3, CheckSquare, Square } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { getHotTopics, getHotSources, fetchHotTopics, refreshHotTopics, getNewHotTopics } from '@/lib/api'
import TrendTimeline from '@/components/TrendTimeline'
import AIAnalysisPanel from '@/components/AIAnalysisPanel'
import type { HotTopic, HotSource, HotTopicTrend } from '@/types/api'

// 分类选项
const categories: { value: string; label: string }[] = [
  { value: 'all', label: '全部分类' },
  { value: '娱乐', label: '娱乐' },
  { value: '科技', label: '科技' },
  { value: '财经', label: '财经' },
  { value: '体育', label: '体育' },
  { value: '社会', label: '社会' },
  { value: '国际', label: '国际' },
]

// 趋势图标
function TrendIcon({ trend }: { trend: HotTopicTrend }) {
  const colors = {
    up: 'text-red-500',
    down: 'text-green-500',
    stable: 'text-gray-500',
    new: 'text-blue-500',
    hot: 'text-orange-500',
  }
  const icons = {
    up: '↑',
    down: '↓',
    stable: '→',
    new: '★',
    hot: '🔥',
  }
  return <span className={`${colors[trend]} font-bold`}>{icons[trend]}</span>
}

// 热度条
function HeatBar({ heat }: { heat: number }) {
  const getColor = (h: number) => {
    if (h >= 80) return 'bg-red-500'
    if (h >= 60) return 'bg-orange-500'
    if (h >= 40) return 'bg-yellow-500'
    return 'bg-blue-500'
  }
  return (
    <div className="w-20 h-2 bg-gray-200 rounded-full overflow-hidden">
      <div
        className={`h-full ${getColor(heat)} transition-all duration-300`}
        style={{ width: `${heat}%` }}
      />
    </div>
  )
}

// 热点卡片
function TopicCard({ topic, onAnalyze, onGenerate, selected, onSelect }: { 
  topic: HotTopic; 
  onAnalyze: (topic: HotTopic) => void; 
  onGenerate: (topic: HotTopic) => void;
  selected: boolean; 
  onSelect: () => void 
}) {
  return (
    <Card className={`hover:shadow-lg transition-shadow cursor-pointer group ${selected ? 'ring-2 ring-blue-500' : ''}`}>
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between">
          <div className="flex items-start gap-2 flex-1 pr-2">
            <button onClick={onSelect} className="mt-1 flex-shrink-0">
              {selected ? (
                <CheckSquare className="h-4 w-4 text-blue-600" />
              ) : (
                <Square className="h-4 w-4 text-gray-400" />
              )}
            </button>
            <div className="flex-1">
              <CardTitle className="text-base line-clamp-2 group-hover:text-blue-600 transition-colors">
                {topic.title}
              </CardTitle>
              <CardDescription className="mt-1 flex items-center gap-2">
                <span className="text-xs">{topic.source}</span>
                {topic.category && (
                  <Badge variant="secondary" className="text-xs">
                    {topic.category}
                  </Badge>
                )}
              </CardDescription>
            </div>
          </div>
          <div className="flex items-center gap-1">
            <TrendIcon trend={topic.trend} />
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-500">热度</span>
            <HeatBar heat={topic.heat} />
            <span className="text-sm font-medium">{topic.heat}</span>
          </div>
          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            {topic.sourceUrl && (
              <Button
                variant="ghost"
                size="sm"
                onClick={(e) => {
                  e.stopPropagation()
                  window.open(topic.sourceUrl, '_blank')
                }}
              >
                <ExternalLink className="h-4 w-4" />
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation()
                onAnalyze(topic)
              }}
              title="AI 分析"
            >
              <BarChart3 className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => {
                e.stopPropagation()
                onGenerate(topic)
              }}
              title="生成内容"
            >
              <Sparkles className="h-4 w-4" />
            </Button>
          </div>
        </div>
        {topic.keywords && topic.keywords.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {topic.keywords.slice(0, 3).map((keyword, idx) => (
              <Badge key={idx} variant="outline" className="text-xs">
                {keyword}
              </Badge>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

export default function HotTopics() {
  const navigate = useNavigate()
  const [topics, setTopics] = useState<HotTopic[]>([])
  const [sources, setSources] = useState<HotSource[]>([])
  const [newTopics, setNewTopics] = useState<HotTopic[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('all')
  const [selectedSource, setSelectedSource] = useState('all')

  // 加载数据源
  useEffect(() => {
    async function loadSources() {
      const result = await getHotSources()
      if (result.success && result.data) {
        setSources(result.data)
      }
    }
    loadSources()
  }, [])

  // 初始加载热点数据
  useEffect(() => {
    async function initialLoad() {
      setLoading(true)
      try {
        // 直接获取热点数据（不依赖数据库）
        const result = await fetchHotTopics(undefined, 30)
        if (result.success && result.data?.topics) {
          setTopics(result.data.topics)
        }
      } finally {
        setLoading(false)
      }
    }
    initialLoad()
  }, [])

  // 加载热点数据（从数据库）
  async function loadTopics() {
    setLoading(true)
    try {
      const result = await getHotTopics({
        search,
        category: category === 'all' ? undefined : category,
        limit: 50,
      })
      if (result.success && result.data && result.data.length > 0) {
        // 前端筛选数据源
        let filtered = result.data
        if (selectedSource !== 'all') {
          filtered = result.data.filter(t => t.sourceId === selectedSource || t.source === selectedSource)
        }
        setTopics(filtered)
      }
    } finally {
      setLoading(false)
    }
  }

  // 加载新增热点
  async function loadNewTopics() {
    const result = await getNewHotTopics(24)
    if (result.success && result.data) {
      setNewTopics(result.data.slice(0, 6))
    }
  }

  useEffect(() => {
    loadTopics()
    loadNewTopics()
  }, [search, category, selectedSource])

  // 刷新数据
  async function handleRefresh() {
    setRefreshing(true)
    try {
      await refreshHotTopics()
      await loadTopics()
      await loadNewTopics()
    } finally {
      setRefreshing(false)
    }
  }

  // 从 NewsNow 拉取数据
  async function handleFetchFromNewsNow() {
    setRefreshing(true)
    try {
      const sourceIds = selectedSource === 'all' ? undefined : [selectedSource]
      const result = await fetchHotTopics(sourceIds, 30)
      if (result.success && result.data?.topics) {
        // 直接使用获取的数据，不依赖数据库
        let filtered = result.data.topics
        if (selectedSource !== 'all') {
          filtered = result.data.topics.filter((t: HotTopic) => t.sourceId === selectedSource || t.source === selectedSource)
        }
        setTopics(filtered)
      }
    } finally {
      setRefreshing(false)
    }
  }

  // 分析热点
  function handleAnalyze(topic: HotTopic) {
    setSelectedTopic(topic)
    setActiveModal('timeline')
  }

  // 生成内容
  function handleGenerate(topic: HotTopic) {
    navigate('/content-generation', { 
      state: { 
        topic: topic.title,
        source: topic.source 
      } 
    })
  }

  // 选择话题
  function handleSelectTopic(topic: HotTopic) {
    setSelectedIds(prev =>
      prev.includes(topic._id)
        ? prev.filter(id => id !== topic._id)
        : [...prev, topic._id]
    )
  }

  // 全选/取消全选
  function handleSelectAll() {
    if (selectedIds.length === topics.length) {
      setSelectedIds([])
    } else {
      setSelectedIds(topics.map(t => t._id))
    }
  }

  // 打开 AI 分析面板
  function handleOpenAIAnalysis() {
    const selectedTopics = topics.filter(t => selectedIds.includes(t._id))
    if (selectedTopics.length > 0) {
      setAnalysisTopics(selectedTopics)
      setActiveModal('ai')
    }
  }

  // 关闭模态窗口
  function handleCloseModal() {
    setActiveModal(null)
    setSelectedTopic(null)
    setAnalysisTopics([])
  }

  // 选中的话题 ID
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [selectedTopic, setSelectedTopic] = useState<HotTopic | null>(null)
  const [analysisTopics, setAnalysisTopics] = useState<HotTopic[]>([])
  const [activeModal, setActiveModal] = useState<'timeline' | 'ai' | null>(null)

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">热点监控</h1>
          <p className="text-gray-500 text-sm mt-1">实时追踪多平台热点话题</p>
        </div>
        <div className="flex items-center gap-2">
          {selectedIds.length > 0 && (
            <Button onClick={handleOpenAIAnalysis} className="bg-purple-600 hover:bg-purple-700">
              <Sparkles className="h-4 w-4 mr-2" />
              AI 分析 ({selectedIds.length})
            </Button>
          )}
          <Button variant="outline" onClick={handleFetchFromNewsNow} disabled={refreshing}>
            <TrendingUp className="h-4 w-4 mr-2" />
            获取最新热点
          </Button>
          <Button onClick={handleRefresh} disabled={refreshing}>
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            刷新
          </Button>
        </div>
      </div>

      {/* 新增热点 */}
      {newTopics.length > 0 && (
        <Card className="bg-gradient-to-r from-orange-50 to-red-50 border-orange-200">
          <CardHeader className="pb-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5 text-red-600" />
                <CardTitle className="text-lg">新增热点</CardTitle>
              </div>
              <span className="text-sm text-gray-500">近24小时</span>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              {newTopics.map((topic) => (
                <div
                  key={topic._id}
                  className="bg-white rounded-lg p-3 hover:shadow-md transition-shadow cursor-pointer"
                  onClick={() => handleAnalyze(topic)}
                >
                  <div className="flex items-start justify-between mb-2">
                    <h4 className="text-sm font-medium text-gray-900 line-clamp-2 flex-1">
                      {topic.title}
                    </h4>
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleSelectTopic(topic)
                      }}
                      className="ml-2 flex-shrink-0"
                    >
                      {selectedIds.includes(topic._id) ? (
                        <CheckSquare className="h-4 w-4 text-blue-600" />
                      ) : (
                        <Square className="h-4 w-4 text-gray-400" />
                      )}
                    </button>
                  </div>
                  <div className="flex items-center justify-between text-xs text-gray-500">
                    <span className="bg-gray-100 px-2 py-0.5 rounded">
                      {topic.source}
                    </span>
                    <span className="text-red-600 font-medium">
                      热度: {topic.heat}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* 筛选栏 */}
      <Card>
        <CardContent className="pt-4">
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <button
                onClick={handleSelectAll}
                className="flex items-center gap-1 text-sm text-gray-600 hover:text-gray-900"
              >
                {selectedIds.length === topics.length && topics.length > 0 ? (
                  <CheckSquare className="h-4 w-4 text-blue-600" />
                ) : (
                  <Square className="h-4 w-4" />
                )}
                <span>全选</span>
              </button>
            </div>
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <Input
                placeholder="搜索热点话题..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={category} onValueChange={setCategory}>
              <SelectTrigger className="w-32">
                <SelectValue placeholder="分类" />
              </SelectTrigger>
              <SelectContent>
                {categories.map((cat) => (
                  <SelectItem key={cat.value} value={cat.value}>
                    {cat.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={selectedSource} onValueChange={setSelectedSource}>
              <SelectTrigger className="w-40">
                <SelectValue placeholder="数据源" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部来源</SelectItem>
                {sources.map((source) => (
                  <SelectItem key={source.id} value={source.id}>
                    {source.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* 热点列表 */}
      {loading ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <Card key={i}>
              <CardContent className="pt-6">
                <div className="animate-pulse space-y-3">
                  <div className="h-4 bg-gray-200 rounded w-3/4" />
                  <div className="h-3 bg-gray-200 rounded w-1/2" />
                  <div className="h-8 bg-gray-200 rounded" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : topics.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <TrendingUp className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">暂无热点数据</h3>
            <p className="text-gray-500 mb-4">点击"获取最新热点"从 NewsNow 拉取数据</p>
            <Button onClick={handleFetchFromNewsNow}>
              <TrendingUp className="h-4 w-4 mr-2" />
              获取最新热点
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {topics.map((topic) => (
            <TopicCard
              key={topic._id}
              topic={topic}
              onAnalyze={handleAnalyze}
              onGenerate={handleGenerate}
              selected={selectedIds.includes(topic._id)}
              onSelect={() => handleSelectTopic(topic)}
            />
          ))}
        </div>
      )}

      {/* 模态窗口 */}
      {activeModal === 'timeline' && selectedTopic && (
        <TrendTimeline topic={selectedTopic} onClose={handleCloseModal} />
      )}
      {activeModal === 'ai' && analysisTopics.length > 0 && (
        <AIAnalysisPanel topics={analysisTopics} onClose={handleCloseModal} />
      )}
    </div>
  )
}

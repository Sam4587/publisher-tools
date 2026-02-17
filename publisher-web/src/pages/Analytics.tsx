import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { RefreshCw, TrendingUp, Eye, Heart, MessageCircle, Users } from 'lucide-react'

interface DashboardStats {
  total_posts: number
  total_views: number
  total_likes: number
  total_comments: number
  total_shares: number
  avg_engagement: number
  platform_stats: PlatformStats[]
  recent_posts: PostMetrics[]
  top_performing: PostMetrics[]
  growth_trend: TrendData[]
}

interface PlatformStats {
  platform: string
  posts: number
  views: number
  likes: number
  comments: number
  engagement: number
}

interface PostMetrics {
  post_id: string
  platform: string
  title: string
  views: number
  likes: number
  comments: number
  shares: number
  engagement: number
}

interface TrendData {
  date: string
  value: number
}

const platformNames: Record<string, string> = {
  douyin: '抖音',
  toutiao: '今日头条',
  xiaohongshu: '小红书',
  weibo: '微博',
}

export default function Analytics() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const fetchDashboard = async () => {
    try {
      const response = await fetch('/api/analytics/dashboard')
      const data = await response.json()
      if (data.success) {
        setStats(data.data)
      }
    } catch (error) {
      console.error('Failed to fetch dashboard:', error)
    } finally {
      setLoading(false)
    }
  }

  const generateMockData = async () => {
    setRefreshing(true)
    try {
      await fetch('/api/analytics/mock', { method: 'POST' })
      await fetchDashboard()
    } finally {
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchDashboard()
  }, [])

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="text-center">加载中...</div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">数据分析</h1>
          <p className="text-muted-foreground mt-2">查看内容发布效果和数据统计</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={generateMockData} disabled={refreshing}>
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            生成测试数据
          </Button>
          <Button onClick={fetchDashboard} disabled={refreshing}>
            <RefreshCw className={`h-4 w-4 mr-2 ${refreshing ? 'animate-spin' : ''}`} />
            刷新
          </Button>
        </div>
      </div>

      {/* 总览卡片 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5 mb-8">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总发布</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_posts || 0}</div>
            <p className="text-xs text-muted-foreground">篇内容</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总浏览</CardTitle>
            <Eye className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(stats?.total_views || 0)}</div>
            <p className="text-xs text-muted-foreground">次浏览</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总点赞</CardTitle>
            <Heart className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(stats?.total_likes || 0)}</div>
            <p className="text-xs text-muted-foreground">次点赞</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总评论</CardTitle>
            <MessageCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(stats?.total_comments || 0)}</div>
            <p className="text-xs text-muted-foreground">条评论</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">平均互动率</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{(stats?.avg_engagement || 0).toFixed(2)}%</div>
            <p className="text-xs text-muted-foreground">互动率</p>
          </CardContent>
        </Card>
      </div>

      {/* 平台分布 */}
      <div className="grid gap-6 lg:grid-cols-2 mb-8">
        <Card>
          <CardHeader>
            <CardTitle>平台数据分布</CardTitle>
            <CardDescription>各平台的内容发布数据</CardDescription>
          </CardHeader>
          <CardContent>
            {stats?.platform_stats && stats.platform_stats.length > 0 ? (
              <div className="space-y-4">
                {stats.platform_stats.map((p) => (
                  <div key={p.platform} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline">{platformNames[p.platform] || p.platform}</Badge>
                      <span className="text-sm text-muted-foreground">{p.posts} 篇</span>
                    </div>
                    <div className="flex items-center gap-4 text-sm">
                      <span className="flex items-center gap-1">
                        <Eye className="h-3 w-3" />
                        {formatNumber(p.views)}
                      </span>
                      <span className="flex items-center gap-1">
                        <Heart className="h-3 w-3" />
                        {formatNumber(p.likes)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center text-muted-foreground py-8">
                暂无数据，点击"生成测试数据"创建示例数据
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>互动趋势</CardTitle>
            <CardDescription>近7天数据趋势</CardDescription>
          </CardHeader>
          <CardContent>
            {stats?.growth_trend && stats.growth_trend.length > 0 ? (
              <div className="space-y-2">
                {stats.growth_trend.map((t) => (
                  <div key={t.date} className="flex items-center justify-between text-sm">
                    <span className="text-muted-foreground">{t.date}</span>
                    <span className="font-medium">{formatNumber(t.value)}</span>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center text-muted-foreground py-8">
                暂无趋势数据
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 最近发布 */}
      <Card>
        <CardHeader>
          <CardTitle>最近发布内容</CardTitle>
          <CardDescription>最新发布的内容表现</CardDescription>
        </CardHeader>
        <CardContent>
          {stats?.recent_posts && stats.recent_posts.length > 0 ? (
            <div className="space-y-4">
              {stats.recent_posts.map((post) => (
                <div key={post.post_id} className="flex items-center justify-between border-b pb-4">
                  <div className="flex-1">
                    <div className="font-medium">{post.title}</div>
                    <div className="flex items-center gap-2 mt-1">
                      <Badge variant="secondary" className="text-xs">
                        {platformNames[post.platform] || post.platform}
                      </Badge>
                      <span className="text-xs text-muted-foreground">
                        互动率 {post.engagement.toFixed(2)}%
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <Eye className="h-3 w-3" />
                      {formatNumber(post.views)}
                    </span>
                    <span className="flex items-center gap-1">
                      <Heart className="h-3 w-3" />
                      {formatNumber(post.likes)}
                    </span>
                    <span className="flex items-center gap-1">
                      <MessageCircle className="h-3 w-3" />
                      {formatNumber(post.comments)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center text-muted-foreground py-8">
              暂无发布数据
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function formatNumber(num: number): string {
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + 'w'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'k'
  }
  return num.toString()
}

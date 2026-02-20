import React, { useState } from 'react'
import ReactECharts from 'echarts-for-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Trophy, Target, Clock, TrendingUp, Filter } from 'lucide-react'

interface RankData {
  date: string
  rank: number
  topicId: string
  topicTitle: string
}

interface RankTimelineChartProps {
  data?: RankData[]
  title?: string
  description?: string
}

export default function RankTimelineChart({ data = [], title = '排名时间线', description = '追踪热点话题的排名变化' }: RankTimelineChartProps) {
  const [timeRange, setTimeRange] = useState('7d')
  const [sortBy, setSortBy] = useState<'rank' | 'date'>('rank')

  // 获取所有唯一的话题
  const topics = React.useMemo(() => {
    const topicMap = new Map<string, string>()
    data.forEach(item => {
      topicMap.set(item.topicId, item.topicTitle)
    })
    return Array.from(topicMap.entries()).slice(0, 5) // 最多显示5个话题
  }, [data])

  // 过滤数据
  const filteredData = React.useMemo(() => {
    if (!data || data.length === 0) return []

    const now = new Date()
    const days = timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : 90
    const cutoffDate = new Date(now.getTime() - days * 24 * 60 * 60 * 1000)

    return data.filter(item => new Date(item.date) >= cutoffDate)
  }, [data, timeRange])

  // 获取图表配置
  const getChartOption = () => {
    const dates = [...new Set(filteredData.map(d => d.date))].sort()

    const series = topics.map(([topicId, topicTitle]) => {
      const topicData = dates.map(date => {
        const item = filteredData.find(d => d.date === date && d.topicId === topicId)
        return item ? item.rank : null
      })

      return {
        name: topicTitle,
        type: 'line',
        data: topicData,
        smooth: true,
        symbol: 'circle',
        symbolSize: 8,
        connectNulls: false,
        lineStyle: {
          width: 2
        },
        itemStyle: {
          borderWidth: 2,
          borderColor: '#fff'
        }
      }
    })

    return {
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross'
        },
        formatter: (params: any) => {
          if (!params || params.length === 0) return ''
          const date = params[0].name
          let html = `<div style="padding: 8px;"><div style="font-weight: bold; margin-bottom: 8px;">${date}</div>`
          params.forEach((param: any) => {
            if (param.value !== null) {
              html += `
                <div style="display: flex; align-items: center; gap: 8px; margin: 4px 0;">
                  <span style="color: ${param.color};">●</span>
                  <span style="flex: 1;">${param.seriesName}</span>
                  <span style="font-weight: bold;">#${param.value}</span>
                </div>
              `
            }
          })
          html += '</div>'
          return html
        }
      },
      legend: {
        type: 'scroll',
        top: 10,
        textStyle: {
          fontSize: 10
        }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        top: 60,
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: dates,
        axisLabel: {
          rotate: 45,
          fontSize: 10
        }
      },
      yAxis: {
        type: 'value',
        inverse: true, // 排名越小越好，所以反转坐标轴
        min: 1,
        max: Math.max(...filteredData.map(d => d.rank), 10),
        axisLabel: {
          formatter: (value: number) => `#${value}`
        }
      },
      series
    }
  }

  // 获取时间线数据
  const getTimelineData = () => {
    const sortedData = [...filteredData].sort((a, b) => {
      if (sortBy === 'rank') return a.rank - b.rank
      return new Date(b.date).getTime() - new Date(a.date).getTime()
    })

    return sortedData.slice(0, 20) // 最多显示20条记录
  }

  const timelineData = getTimelineData()

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Trophy className="h-5 w-5 text-yellow-500" />
              {title}
            </CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
          <div className="flex gap-2">
            <Select value={timeRange} onValueChange={setTimeRange}>
              <SelectTrigger className="w-32">
                <Clock className="h-4 w-4 mr-2" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="7d">近7天</SelectItem>
                <SelectItem value="30d">近30天</SelectItem>
                <SelectItem value="90d">近90天</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="chart" className="w-full">
          <TabsList className="mb-4">
            <TabsTrigger value="chart" className="flex items-center gap-2">
              <TrendingUp className="h-4 w-4" />
              趋势图
            </TabsTrigger>
            <TabsTrigger value="timeline" className="flex items-center gap-2">
              <Target className="h-4 w-4" />
              时间线
            </TabsTrigger>
          </TabsList>

          <TabsContent value="chart" className="mt-0">
            {filteredData.length > 0 ? (
              <ReactECharts
                option={getChartOption()}
                style={{ height: '400px' }}
                opts={{ renderer: 'svg' }}
              />
            ) : (
              <div className="flex items-center justify-center h-[400px] text-muted-foreground">
                暂无排名数据
              </div>
            )}
          </TabsContent>

          <TabsContent value="timeline" className="mt-0">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex gap-2">
                  <Button
                    variant={sortBy === 'rank' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSortBy('rank')}
                  >
                    按排名
                  </Button>
                  <Button
                    variant={sortBy === 'date' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSortBy('date')}
                  >
                    按时间
                  </Button>
                </div>
                <Filter className="h-4 w-4 text-muted-foreground" />
              </div>

              <div className="space-y-3 max-h-[350px] overflow-y-auto">
                {timelineData.length > 0 ? (
                  timelineData.map((item) => (
                    <div
                      key={`${item.topicId}-${item.date}`}
                      className="flex items-center gap-4 p-4 border rounded-lg hover:bg-gray-50 transition-colors"
                    >
                      <div className="flex-shrink-0">
                        <div
                          className={`w-10 h-10 rounded-full flex items-center justify-center font-bold ${
                            item.rank <= 3
                              ? 'bg-yellow-100 text-yellow-700'
                              : item.rank <= 10
                              ? 'bg-blue-100 text-blue-700'
                              : 'bg-gray-100 text-gray-600'
                          }`}
                        >
                          #{item.rank}
                        </div>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="font-medium truncate">{item.topicTitle}</div>
                        <div className="text-sm text-muted-foreground">{item.date}</div>
                      </div>
                      {item.rank <= 3 && (
                        <Trophy className="h-5 w-5 text-yellow-500 flex-shrink-0" />
                      )}
                    </div>
                  ))
                ) : (
                  <div className="flex items-center justify-center h-[300px] text-muted-foreground">
                    暂无排名数据
                  </div>
                )}
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}

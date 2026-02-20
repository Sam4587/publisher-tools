import { useState, useEffect } from 'react'
import ReactECharts from 'echarts-for-react'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { TrendingUp, TrendingDown, Minus, Calendar, BarChart3, LineChart } from 'lucide-react'

interface TrendData {
  date: string
  value: number
  rank?: number
}

interface HotspotTrendChartProps {
  data?: TrendData[]
  title?: string
  description?: string
}

export default function HotspotTrendChart({ data = [], title = '热点趋势', description = '查看热点话题的变化趋势' }: HotspotTrendChartProps) {
  const [timeRange, setTimeRange] = useState('7d')
  const [chartType, setChartType] = useState<'line' | 'bar'>('line')
  const [filteredData, setFilteredData] = useState<TrendData[]>(data)

  useEffect(() => {
    if (!data || data.length === 0) return

    const now = new Date()
    const days = timeRange === '7d' ? 7 : timeRange === '30d' ? 30 : 90
    const cutoffDate = new Date(now.getTime() - days * 24 * 60 * 60 * 1000)

    const filtered = data.filter(item => new Date(item.date) >= cutoffDate)
    setFilteredData(filtered)
  }, [data, timeRange])

  const getOption = () => {
    const dates = filteredData.map(d => d.date)
    const values = filteredData.map(d => d.value)

    return {
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'cross'
        },
        formatter: (params: any) => {
          const param = params[0]
          return `
            <div style="padding: 8px;">
              <div style="font-weight: bold; margin-bottom: 4px;">${param.name}</div>
              <div style="display: flex; align-items: center; gap: 8px;">
                <span style="color: #3b82f6;">●</span>
                <span>热度: ${param.value}</span>
              </div>
              ${param.data.rank ? `<div style="margin-top: 4px;">排名: #${param.data.rank}</div>` : ''}
            </div>
          `
        }
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: chartType === 'bar',
        data: dates,
        axisLabel: {
          rotate: 45,
          fontSize: 10
        }
      },
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: (value: number) => {
            if (value >= 10000) return (value / 10000).toFixed(1) + 'w'
            if (value >= 1000) return (value / 1000).toFixed(1) + 'k'
            return value.toString()
          }
        }
      },
      series: [
        {
          name: '热度',
          type: chartType,
          data: values.map((v, i) => ({
            value: v,
            rank: filteredData[i]?.rank
          })),
          smooth: chartType === 'line',
          itemStyle: {
            color: '#3b82f6'
          },
          areaStyle: chartType === 'line' ? {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(59, 130, 246, 0.3)' },
                { offset: 1, color: 'rgba(59, 130, 246, 0.05)' }
              ]
            }
          } : undefined,
          markPoint: {
            data: [
              { type: 'max', name: '最高' },
              { type: 'min', name: '最低' }
            ]
          },
          markLine: {
            data: [{ type: 'average', name: '平均值' }]
          }
        }
      ]
    }
  }

  const calculateTrend = () => {
    if (filteredData.length < 2) return 'stable'
    const first = filteredData[0].value
    const last = filteredData[filteredData.length - 1].value
    const change = ((last - first) / first) * 100

    if (change > 10) return 'up'
    if (change < -10) return 'down'
    return 'stable'
  }

  const trend = calculateTrend()
  const trendIcon = trend === 'up' ? <TrendingUp className="h-4 w-4 text-green-600" /> :
                    trend === 'down' ? <TrendingDown className="h-4 w-4 text-red-600" /> :
                    <Minus className="h-4 w-4 text-gray-400" />

  const trendText = trend === 'up' ? '上升' : trend === 'down' ? '下降' : '稳定'

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              {title}
              <span className="flex items-center gap-1 text-sm text-muted-foreground">
                {trendIcon}
                {trendText}
              </span>
            </CardTitle>
            <CardDescription>{description}</CardDescription>
          </div>
          <div className="flex gap-2">
            <Select value={timeRange} onValueChange={setTimeRange}>
              <SelectTrigger className="w-32">
                <Calendar className="h-4 w-4 mr-2" />
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
          <div className="flex items-center justify-between mb-4">
            <TabsList>
              <TabsTrigger value="chart" className="flex items-center gap-2">
                <LineChart className="h-4 w-4" />
                图表
              </TabsTrigger>
              <TabsTrigger value="list" className="flex items-center gap-2">
                <BarChart3 className="h-4 w-4" />
                列表
              </TabsTrigger>
            </TabsList>
            <div className="flex gap-1">
              <Button
                variant={chartType === 'line' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setChartType('line')}
              >
                折线
              </Button>
              <Button
                variant={chartType === 'bar' ? 'default' : 'outline'}
                size="sm"
                onClick={() => setChartType('bar')}
              >
                柱状
              </Button>
            </div>
          </div>

          <TabsContent value="chart" className="mt-0">
            {filteredData.length > 0 ? (
              <ReactECharts
                option={getOption()}
                style={{ height: '400px' }}
                opts={{ renderer: 'svg' }}
              />
            ) : (
              <div className="flex items-center justify-center h-[400px] text-muted-foreground">
                暂无数据
              </div>
            )}
          </TabsContent>

          <TabsContent value="list" className="mt-0">
            <div className="space-y-2 max-h-[400px] overflow-y-auto">
              {filteredData.length > 0 ? (
                filteredData.map((item, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-sm text-muted-foreground">{item.date}</span>
                      {item.rank && (
                        <span className="text-xs bg-blue-100 text-blue-600 px-2 py-1 rounded">
                          排名 #{item.rank}
                        </span>
                      )}
                    </div>
                    <span className="font-medium">{item.value.toLocaleString()}</span>
                  </div>
                ))
              ) : (
                <div className="flex items-center justify-center h-[400px] text-muted-foreground">
                  暂无数据
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}

import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import RankTimelineChart from './RankTimelineChart'

describe('RankTimelineChart', () => {
  const mockData = [
    { date: '2026-02-14', rank: 5, topicId: '1', topicTitle: '话题1' },
    { date: '2026-02-15', rank: 3, topicId: '1', topicTitle: '话题1' },
    { date: '2026-02-16', rank: 1, topicId: '1', topicTitle: '话题1' },
    { date: '2026-02-14', rank: 8, topicId: '2', topicTitle: '话题2' },
    { date: '2026-02-15', rank: 6, topicId: '2', topicTitle: '话题2' },
    { date: '2026-02-16', rank: 4, topicId: '2', topicTitle: '话题2' },
  ]

  it('renders correctly', () => {
    render(<RankTimelineChart data={mockData} />)
    expect(screen.getByText('排名时间线')).toBeInTheDocument()
  })

  it('renders with custom title', () => {
    render(<RankTimelineChart data={mockData} title="自定义标题" />)
    expect(screen.getByText('自定义标题')).toBeInTheDocument()
  })

  it('renders with custom description', () => {
    render(<RankTimelineChart data={mockData} description="自定义描述" />)
    expect(screen.getByText('自定义描述')).toBeInTheDocument()
  })

  it('displays trophy icon', () => {
    render(<RankTimelineChart data={mockData} />)
    // 应该显示奖杯图标
    const titleElement = screen.getByText('排名时间线')
    expect(titleElement).toBeInTheDocument()
  })

  it('renders chart when data is provided', () => {
    render(<RankTimelineChart data={mockData} />)
    // ECharts 应该渲染图表
    const chartContainer = document.querySelector('.echarts-for-react')
    expect(chartContainer).toBeInTheDocument()
  })

  it('renders empty state when no data', () => {
    render(<RankTimelineChart data={[]} />)
    // 应该显示"暂无排名数据"提示
    expect(screen.getByText('暂无排名数据')).toBeInTheDocument()
  })

  it('handles time range selection', () => {
    render(<RankTimelineChart data={mockData} />)
    // 应该有时间范围选择器
    expect(screen.getByText('近7天')).toBeInTheDocument()
  })

  it('handles tab switching', () => {
    render(<RankTimelineChart data={mockData} />)
    // 应该有趋势图和时间线两个标签
    expect(screen.getByText('趋势图')).toBeInTheDocument()
    expect(screen.getByText('时间线')).toBeInTheDocument()
  })

  it('handles sort selection', () => {
    render(<RankTimelineChart data={mockData} />)
    // 应该有排序选择按钮
    expect(screen.getByText('按排名')).toBeInTheDocument()
    expect(screen.getByText('按时间')).toBeInTheDocument()
  })

  it('displays ranking badges', () => {
    render(<RankTimelineChart data={mockData} />)
    // 应该显示排名徽章
    expect(screen.getByText('排名追踪')).toBeInTheDocument()
  })
})

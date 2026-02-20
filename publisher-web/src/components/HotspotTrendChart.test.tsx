import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import HotspotTrendChart from './HotspotTrendChart'

describe('HotspotTrendChart', () => {
  const mockData = [
    { date: '2026-02-14', value: 10000 },
    { date: '2026-02-15', value: 20000 },
    { date: '2026-02-16', value: 15000 },
    { date: '2026-02-17', value: 25000 },
    { date: '2026-02-18', value: 30000 },
    { date: '2026-02-19', value: 35000 },
    { date: '2026-02-20', value: 40000 },
  ]

  it('renders correctly', () => {
    render(<HotspotTrendChart data={mockData} />)
    expect(screen.getByText('热点趋势分析')).toBeInTheDocument()
  })

  it('renders with custom title', () => {
    render(<HotspotTrendChart data={mockData} title="自定义标题" />)
    expect(screen.getByText('自定义标题')).toBeInTheDocument()
  })

  it('renders with custom description', () => {
    render(<HotspotTrendChart data={mockData} description="自定义描述" />)
    expect(screen.getByText('自定义描述')).toBeInTheDocument()
  })

  it('displays trend indicator', () => {
    render(<HotspotTrendChart data={mockData} />)
    // 应该显示趋势指示器（上升/下降/稳定）
    const titleElement = screen.getByText('热点趋势分析')
    expect(titleElement).toBeInTheDocument()
  })

  it('renders chart when data is provided', () => {
    render(<HotspotTrendChart data={mockData} />)
    // ECharts 应该渲染图表
    const chartContainer = document.querySelector('.echarts-for-react')
    expect(chartContainer).toBeInTheDocument()
  })

  it('renders empty state when no data', () => {
    render(<HotspotTrendChart data={[]} />)
    // 应该显示"暂无数据"提示
    expect(screen.getByText('暂无数据')).toBeInTheDocument()
  })

  it('handles time range selection', () => {
    render(<HotspotTrendChart data={mockData} />)
    // 应该有时间范围选择器
    expect(screen.getByText('近7天')).toBeInTheDocument()
  })

  it('handles chart type toggle', () => {
    render(<HotspotTrendChart data={mockData} />)
    // 应该有图表类型切换按钮
    expect(screen.getByText('折线')).toBeInTheDocument()
    expect(screen.getByText('柱状')).toBeInTheDocument()
  })

  it('handles tab switching', () => {
    render(<HotspotTrendChart data={mockData} />)
    // 应该有图表和列表两个标签
    expect(screen.getByText('图表')).toBeInTheDocument()
    expect(screen.getByText('列表')).toBeInTheDocument()
  })
})

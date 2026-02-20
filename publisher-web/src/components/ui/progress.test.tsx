import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Progress } from './progress'

describe('Progress', () => {
  it('renders correctly', () => {
    render(<Progress value={50} />)
    const progressContainer = screen.getByRole('progressbar')
    expect(progressContainer).toBeInTheDocument()
  })

  it('renders with default value', () => {
    render(<Progress />)
    const progressContainer = screen.getByRole('progressbar')
    expect(progressContainer).toBeInTheDocument()
  })

  it('applies custom className', () => {
    render(<Progress value={50} className="custom-class" />)
    const progressContainer = screen.getByRole('progressbar')
    expect(progressContainer).toHaveClass('custom-class')
  })

  it('handles value changes', () => {
    const { rerender } = render(<Progress value={25} />)
    const progressContainer = screen.getByRole('progressbar')
    expect(progressContainer).toBeInTheDocument()

    rerender(<Progress value={75} />)
    expect(screen.getByRole('progressbar')).toBeInTheDocument()
  })

  it('handles values greater than 100', () => {
    render(<Progress value={150} />)
    const progressContainer = screen.getByRole('progressbar')
    expect(progressContainer).toBeInTheDocument()
  })

  it('handles negative values', () => {
    render(<Progress value={-10} />)
    const progressContainer = screen.getByRole('progressbar')
    expect(progressContainer).toBeInTheDocument()
  })

  it('displays progress indicator visually', () => {
    render(<Progress value={50} />)
    const progressContainer = screen.getByRole('progressbar')
    // 验证进度条容器存在
    expect(progressContainer).toBeInTheDocument()
    // 验证进度条有正确的样式类
    expect(progressContainer).toHaveClass('relative')
  })

  it('has correct accessibility attributes', () => {
    render(<Progress value={50} />)
    const progressContainer = screen.getByRole('progressbar')
    // 验证角色属性
    expect(progressContainer).toHaveAttribute('role', 'progressbar')
  })
})

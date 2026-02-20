import React, { useState, useEffect } from 'react'
import { Search, Filter, X, SlidersHorizontal, Calendar, TrendingUp, Tag } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'

interface FilterOptions {
  keyword?: string
  platform?: string
  category?: string
  dateRange?: string
  sortBy?: string
  trend?: string
}

interface GlobalFilterBarProps {
  onFilterChange: (filters: FilterOptions) => void
  initialFilters?: FilterOptions
  platforms?: string[]
  categories?: string[]
  placeholder?: string
}

const platformNames: Record<string, string> = {
  douyin: '抖音',
  toutiao: '今日头条',
  xiaohongshu: '小红书',
  weibo: '微博',
  bilibili: 'B站',
}

const categoryNames: Record<string, string> = {
  news: '新闻',
  entertainment: '娱乐',
  sports: '体育',
  tech: '科技',
  finance: '财经',
}

export default function GlobalFilterBar({
  onFilterChange,
  initialFilters = {},
  platforms = ['douyin', 'toutiao', 'xiaohongshu', 'weibo', 'bilibili'],
  categories = ['news', 'entertainment', 'sports', 'tech', 'finance'],
  placeholder = '搜索话题、内容...'
}: GlobalFilterBarProps) {
  const [filters, setFilters] = useState<FilterOptions>(initialFilters)
  const [keyword, setKeyword] = useState(initialFilters.keyword || '')
  const [isFilterOpen, setIsFilterOpen] = useState(false)

  useEffect(() => {
    onFilterChange(filters)
  }, [filters, onFilterChange])

  const updateFilter = (key: keyof FilterOptions, value: string | undefined) => {
    setFilters(prev => ({
      ...prev,
      [key]: value === 'all' ? undefined : value,
    }))
  }

  const clearFilters = () => {
    setFilters({})
    setKeyword('')
  }

  const clearFilter = (key: keyof FilterOptions) => {
    setFilters(prev => {
      const newFilters = { ...prev }
      delete newFilters[key]
      return newFilters
    })
    if (key === 'keyword') {
      setKeyword('')
    }
  }

  const activeFiltersCount = Object.keys(filters).filter(key =>
    filters[key as keyof FilterOptions] !== undefined
  ).length

  const activeFilters = Object.entries(filters).filter(([_, value]) => value !== undefined)

  return (
    <div className="space-y-3">
      {/* 搜索栏 */}
      <div className="flex gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            type="text"
            placeholder={placeholder}
            value={keyword}
            onChange={(e) => {
              setKeyword(e.target.value)
              const timer = setTimeout(() => {
                setFilters(prev => ({ ...prev, keyword: e.target.value || undefined }))
              }, 300)
              return () => clearTimeout(timer)
            }}
            className="pl-9"
          />
          {keyword && (
            <Button
              variant="ghost"
              size="sm"
              className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6 p-0"
              onClick={() => {
                setKeyword('')
                clearFilter('keyword')
              }}
            >
              <X className="h-3 w-3" />
            </Button>
          )}
        </div>

        <Sheet open={isFilterOpen} onOpenChange={setIsFilterOpen}>
          <SheetTrigger asChild>
            <Button variant="outline" className="relative">
              <SlidersHorizontal className="h-4 w-4 mr-2" />
              筛选
              {activeFiltersCount > 0 && (
                <Badge className="absolute -top-2 -right-2 h-5 w-5 flex items-center justify-center p-0 text-xs">
                  {activeFiltersCount}
                </Badge>
              )}
            </Button>
          </SheetTrigger>
          <SheetContent>
            <SheetHeader>
              <SheetTitle>筛选条件</SheetTitle>
              <SheetDescription>
                设置筛选条件来查找您需要的内容
              </SheetDescription>
            </SheetHeader>

            <div className="space-y-6 mt-6">
              {/* 平台筛选 */}
              <div className="space-y-2">
                <label className="text-sm font-medium flex items-center gap-2">
                  <Tag className="h-4 w-4" />
                  平台
                </label>
                <Select
                  value={filters.platform || 'all'}
                  onValueChange={(value) => updateFilter('platform', value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择平台" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部平台</SelectItem>
                    {platforms.map(platform => (
                      <SelectItem key={platform} value={platform}>
                        {platformNames[platform] || platform}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* 分类筛选 */}
              <div className="space-y-2">
                <label className="text-sm font-medium flex items-center gap-2">
                  <Filter className="h-4 w-4" />
                  分类
                </label>
                <Select
                  value={filters.category || 'all'}
                  onValueChange={(value) => updateFilter('category', value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择分类" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部分类</SelectItem>
                    {categories.map(category => (
                      <SelectItem key={category} value={category}>
                        {categoryNames[category] || category}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* 时间范围 */}
              <div className="space-y-2">
                <label className="text-sm font-medium flex items-center gap-2">
                  <Calendar className="h-4 w-4" />
                  时间范围
                </label>
                <Select
                  value={filters.dateRange || 'all'}
                  onValueChange={(value) => updateFilter('dateRange', value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择时间范围" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部时间</SelectItem>
                    <SelectItem value="today">今天</SelectItem>
                    <SelectItem value="week">近7天</SelectItem>
                    <SelectItem value="month">近30天</SelectItem>
                    <SelectItem value="quarter">近90天</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 趋势筛选 */}
              <div className="space-y-2">
                <label className="text-sm font-medium flex items-center gap-2">
                  <TrendingUp className="h-4 w-4" />
                  趋势
                </label>
                <Select
                  value={filters.trend || 'all'}
                  onValueChange={(value) => updateFilter('trend', value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择趋势" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部趋势</SelectItem>
                    <SelectItem value="up">上升</SelectItem>
                    <SelectItem value="down">下降</SelectItem>
                    <SelectItem value="stable">稳定</SelectItem>
                    <SelectItem value="new">新话题</SelectItem>
                    <SelectItem value="hot">热门</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 排序方式 */}
              <div className="space-y-2">
                <label className="text-sm font-medium">排序方式</label>
                <Select
                  value={filters.sortBy || 'heat'}
                  onValueChange={(value) => updateFilter('sortBy', value)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="选择排序方式" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="heat">按热度</SelectItem>
                    <SelectItem value="date">按时间</SelectItem>
                    <SelectItem value="rank">按排名</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* 操作按钮 */}
              <div className="flex gap-2 pt-4">
                <Button
                  variant="outline"
                  className="flex-1"
                  onClick={clearFilters}
                >
                  清除全部
                </Button>
                <Button
                  className="flex-1"
                  onClick={() => setIsFilterOpen(false)}
                >
                  应用筛选
                </Button>
              </div>
            </div>
          </SheetContent>
        </Sheet>

        {activeFiltersCount > 0 && (
          <Button variant="ghost" size="sm" onClick={clearFilters}>
            <X className="h-4 w-4 mr-2" />
            清除
          </Button>
        )}
      </div>

      {/* 活跃筛选标签 */}
      {activeFiltersCount > 0 && (
        <div className="flex flex-wrap gap-2">
          {activeFilters.map(([key, value]) => {
            let label = value as string
            let icon: React.ReactNode | null = null

            switch (key) {
              case 'platform':
                label = platformNames[value as string] || value as string
                icon = <Tag className="h-3 w-3" />
                break
              case 'category':
                label = categoryNames[value as string] || value as string
                icon = <Filter className="h-3 w-3" />
                break
              case 'dateRange':
                const dateLabels: Record<string, string> = {
                  today: '今天',
                  week: '近7天',
                  month: '近30天',
                  quarter: '近90天',
                }
                label = dateLabels[value as string] || value as string
                icon = <Calendar className="h-3 w-3" />
                break
              case 'trend':
                const trendLabels: Record<string, string> = {
                  up: '上升',
                  down: '下降',
                  stable: '稳定',
                  new: '新话题',
                  hot: '热门',
                }
                label = trendLabels[value as string] || value as string
                icon = <TrendingUp className="h-3 w-3" />
                break
              case 'sortBy':
                const sortLabels: Record<string, string> = {
                  heat: '按热度',
                  date: '按时间',
                  rank: '按排名',
                }
                label = sortLabels[value as string] || value as string
                break
            }

            return (
              <Badge key={key} variant="secondary" className="gap-1">
                {icon}
                {label}
                <button
                  onClick={() => clearFilter(key as keyof FilterOptions)}
                  className="ml-1 hover:text-red-600"
                >
                  <X className="h-3 w-3" />
                </button>
              </Badge>
            )
          })}
        </div>
      )}
    </div>
  )
}

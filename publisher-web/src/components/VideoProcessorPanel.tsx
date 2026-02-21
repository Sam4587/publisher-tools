import { useState, useCallback } from 'react'
import { Video, Scissors, Image, RefreshCw, Settings, ChevronDown, ChevronUp, Play, Download, Info } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

// 视频信息接口
interface VideoInfo {
  file_path: string
  file_name: string
  file_size: number
  duration: number
  width: number
  height: number
  aspect_ratio: string
  fps: number
  bit_rate: number
  codec: string
  audio_codec: string
  format: string
}

// 切片结果接口
interface SliceResult {
  source_file: string
  output_dir: string
  slice_count: number
  slices: SliceInfo[]
  duration: number
  total_size: number
}

// 切片信息接口
interface SliceInfo {
  index: number
  file_path: string
  file_name: string
  start_time: number
  end_time: number
  duration: number
  size: number
}

// 缩略图结果接口
interface ThumbnailResult {
  source_file: string
  output_dir: string
  thumbnails: ThumbnailInfo[]
  sprite_sheet?: SpriteSheetInfo
  duration: number
}

// 缩略图信息接口
interface ThumbnailInfo {
  index: number
  file_path: string
  file_name: string
  timestamp: number
  width: number
  height: number
  size: number
}

// 雪碧图信息接口
interface SpriteSheetInfo {
  file_path: string
  file_name: string
  width: number
  height: number
  columns: number
  rows: number
  count: number
  size: number
}

// 转换结果接口
interface ConvertResult {
  source_file: string
  output_file: string
  source_info: VideoInfo
  output_info: VideoInfo
  duration: number
  compression_ratio: number
}

// 分辨率预设
const RESOLUTION_PRESETS = [
  { name: '4K', width: 3840, height: 2160 },
  { name: '1080p', width: 1920, height: 1080 },
  { name: '720p', width: 1280, height: 720 },
  { name: '480p', width: 854, height: 480 },
  { name: '360p', width: 640, height: 360 },
]

// 支持的格式
const SUPPORTED_FORMATS = ['mp4', 'webm', 'avi', 'mov', 'mkv', 'flv']

// 视频编码
const VIDEO_CODECS = [
  { value: 'libx264', label: 'H.264 (兼容性最好)' },
  { value: 'libx265', label: 'H.265 (高压缩率)' },
  { value: 'libvpx-vp9', label: 'VP9 (WebM)' },
  { value: 'libaom-av1', label: 'AV1 (最新)' },
]

export default function VideoProcessorPanel() {
  // 视频路径
  const [videoPath, setVideoPath] = useState('')
  const [videoInfo, setVideoInfo] = useState<VideoInfo | null>(null)

  // 切片配置
  const [sliceMode, setSliceMode] = useState<'duration' | 'count' | 'time'>('duration')
  const [sliceDuration, setSliceDuration] = useState(300)
  const [sliceCount, setSliceCount] = useState(5)
  const [sliceStartTime, setSliceStartTime] = useState(0)
  const [sliceEndTime, setSliceEndTime] = useState(0)
  const [sliceResult, setSliceResult] = useState<SliceResult | null>(null)

  // 缩略图配置
  const [thumbCount, setThumbCount] = useState(5)
  const [thumbWidth, setThumbWidth] = useState(320)
  const [thumbInterval, setThumbInterval] = useState(0)
  const [generateSprite, setGenerateSprite] = useState(false)
  const [thumbResult, setThumbResult] = useState<ThumbnailResult | null>(null)

  // 转换配置
  const [outputFormat, setOutputFormat] = useState('mp4')
  const [videoCodec, setVideoCodec] = useState('libx264')
  const [resolution, setResolution] = useState('720p')
  const [crf, setCRF] = useState(23)
  const [preset, setPreset] = useState('medium')
  const [convertResult, setConvertResult] = useState<ConvertResult | null>(null)

  // 处理状态
  const [isProcessing, setIsProcessing] = useState(false)
  const [progress, setProgress] = useState(0)
  const [activeTab, setActiveTab] = useState('info')

  // 高级设置
  const [showAdvanced, setShowAdvanced] = useState(false)

  // 获取视频信息
  const handleGetInfo = async () => {
    if (!videoPath) return

    setIsProcessing(true)
    setProgress(0)

    try {
      const response = await fetch('/api/v1/video/info', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ file_path: videoPath }),
      })

      const data = await response.json()
      if (data.success) {
        setVideoInfo(data.data)
        // 设置默认结束时间为视频时长
        if (data.data.duration) {
          setSliceEndTime(data.data.duration)
        }
      }
    } catch (err) {
      console.error('Failed to get video info:', err)
    } finally {
      setIsProcessing(false)
    }
  }

  // 切片视频
  const handleSlice = async () => {
    if (!videoPath) return

    setIsProcessing(true)
    setProgress(0)

    const progressInterval = setInterval(() => {
      setProgress(prev => Math.min(prev + 5, 90))
    }, 500)

    try {
      const body: Record<string, unknown> = {
        file_path: videoPath,
      }

      switch (sliceMode) {
        case 'duration':
          body.segment_duration = sliceDuration
          break
        case 'count':
          body.segment_count = sliceCount
          break
        case 'time':
          body.start_time = sliceStartTime
          body.end_time = sliceEndTime
          break
      }

      const response = await fetch('/api/v1/video/slice', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })

      clearInterval(progressInterval)
      setProgress(100)

      const data = await response.json()
      if (data.success) {
        setSliceResult(data.data)
      }
    } catch (err) {
      console.error('Failed to slice video:', err)
    } finally {
      clearInterval(progressInterval)
      setIsProcessing(false)
    }
  }

  // 生成缩略图
  const handleGenerateThumbnails = async () => {
    if (!videoPath) return

    setIsProcessing(true)
    setProgress(0)

    const progressInterval = setInterval(() => {
      setProgress(prev => Math.min(prev + 5, 90))
    }, 500)

    try {
      const response = await fetch('/api/v1/video/thumbnails', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          file_path: videoPath,
          count: thumbCount,
          width: thumbWidth,
          interval: thumbInterval > 0 ? thumbInterval : undefined,
          sprite_sheet: generateSprite,
        }),
      })

      clearInterval(progressInterval)
      setProgress(100)

      const data = await response.json()
      if (data.success) {
        setThumbResult(data.data)
      }
    } catch (err) {
      console.error('Failed to generate thumbnails:', err)
    } finally {
      clearInterval(progressInterval)
      setIsProcessing(false)
    }
  }

  // 转换视频
  const handleConvert = async () => {
    if (!videoPath) return

    setIsProcessing(true)
    setProgress(0)

    const progressInterval = setInterval(() => {
      setProgress(prev => Math.min(prev + 2, 90))
    }, 1000)

    try {
      const response = await fetch('/api/v1/video/convert', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          file_path: videoPath,
          format: outputFormat,
          video_codec: videoCodec,
          resolution: resolution,
          crf: crf,
          preset: preset,
        }),
      })

      clearInterval(progressInterval)
      setProgress(100)

      const data = await response.json()
      if (data.success) {
        setConvertResult(data.data)
      }
    } catch (err) {
      console.error('Failed to convert video:', err)
    } finally {
      clearInterval(progressInterval)
      setIsProcessing(false)
    }
  }

  // 格式化时长
  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = Math.floor(seconds % 60)
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  // 格式化文件大小
  const formatFileSize = (bytes: number) => {
    if (bytes >= 1024 * 1024 * 1024) {
      return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
    }
    if (bytes >= 1024 * 1024) {
      return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
    }
    if (bytes >= 1024) {
      return `${(bytes / 1024).toFixed(2)} KB`
    }
    return `${bytes} B`
  }

  return (
    <div className="space-y-6">
      {/* 视频路径输入 */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Video className="h-5 w-5 text-blue-600" />
            <CardTitle className="text-lg">视频处理工具</CardTitle>
          </div>
          <CardDescription>
            支持视频切片、缩略图生成、格式转换等功能
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label>视频文件路径</Label>
            <div className="flex gap-2">
              <Input
                value={videoPath}
                onChange={(e) => setVideoPath(e.target.value)}
                placeholder="输入视频文件路径"
                className="flex-1"
              />
              <Button onClick={handleGetInfo} disabled={isProcessing || !videoPath}>
                <Info className="h-4 w-4 mr-2" />
                获取信息
              </Button>
            </div>
          </div>

          {/* 视频信息显示 */}
          {videoInfo && (
            <div className="bg-gray-50 p-4 rounded-lg">
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div>
                  <span className="text-gray-500">文件名:</span>
                  <p className="font-medium truncate">{videoInfo.file_name}</p>
                </div>
                <div>
                  <span className="text-gray-500">时长:</span>
                  <p className="font-medium">{formatDuration(videoInfo.duration)}</p>
                </div>
                <div>
                  <span className="text-gray-500">分辨率:</span>
                  <p className="font-medium">{videoInfo.width}x{videoInfo.height}</p>
                </div>
                <div>
                  <span className="text-gray-500">大小:</span>
                  <p className="font-medium">{formatFileSize(videoInfo.file_size)}</p>
                </div>
                <div>
                  <span className="text-gray-500">编码:</span>
                  <p className="font-medium">{videoInfo.codec}</p>
                </div>
                <div>
                  <span className="text-gray-500">帧率:</span>
                  <p className="font-medium">{videoInfo.fps.toFixed(2)} fps</p>
                </div>
                <div>
                  <span className="text-gray-500">比特率:</span>
                  <p className="font-medium">{(videoInfo.bit_rate / 1000).toFixed(0)} kbps</p>
                </div>
                <div>
                  <span className="text-gray-500">格式:</span>
                  <p className="font-medium">{videoInfo.format}</p>
                </div>
              </div>
            </div>
          )}

          {/* 进度条 */}
          {isProcessing && (
            <div className="space-y-2">
              <Progress value={progress} />
              <p className="text-sm text-gray-500 text-center">处理中...</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 功能选项卡 */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="info">
            <Scissors className="h-4 w-4 mr-2" />
            视频切片
          </TabsTrigger>
          <TabsTrigger value="thumbnail">
            <Image className="h-4 w-4 mr-2" />
            缩略图
          </TabsTrigger>
          <TabsTrigger value="convert">
            <RefreshCw className="h-4 w-4 mr-2" />
            格式转换
          </TabsTrigger>
        </TabsList>

        {/* 切片选项卡 */}
        <TabsContent value="info">
          <Card>
            <CardContent className="pt-6 space-y-4">
              <div className="space-y-2">
                <Label>切片模式</Label>
                <div className="flex gap-2">
                  <Button
                    variant={sliceMode === 'duration' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSliceMode('duration')}
                  >
                    按时长
                  </Button>
                  <Button
                    variant={sliceMode === 'count' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSliceMode('count')}
                  >
                    按数量
                  </Button>
                  <Button
                    variant={sliceMode === 'time' ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setSliceMode('time')}
                  >
                    按时间范围
                  </Button>
                </div>
              </div>

              {sliceMode === 'duration' && (
                <div className="space-y-2">
                  <Label>每段时长 (秒)</Label>
                  <Input
                    type="number"
                    value={sliceDuration}
                    onChange={(e) => setSliceDuration(Number(e.target.value))}
                    min={1}
                  />
                </div>
              )}

              {sliceMode === 'count' && (
                <div className="space-y-2">
                  <Label>切片数量</Label>
                  <Input
                    type="number"
                    value={sliceCount}
                    onChange={(e) => setSliceCount(Number(e.target.value))}
                    min={1}
                    max={100}
                  />
                </div>
              )}

              {sliceMode === 'time' && (
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label>开始时间 (秒)</Label>
                    <Input
                      type="number"
                      value={sliceStartTime}
                      onChange={(e) => setSliceStartTime(Number(e.target.value))}
                      min={0}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label>结束时间 (秒)</Label>
                    <Input
                      type="number"
                      value={sliceEndTime}
                      onChange={(e) => setSliceEndTime(Number(e.target.value))}
                      min={0}
                    />
                  </div>
                </div>
              )}

              <Button onClick={handleSlice} disabled={isProcessing || !videoPath} className="w-full">
                <Scissors className="h-4 w-4 mr-2" />
                开始切片
              </Button>

              {/* 切片结果 */}
              {sliceResult && (
                <div className="border-t pt-4 space-y-4">
                  <div className="flex items-center justify-between">
                    <h4 className="font-medium">切片结果</h4>
                    <Badge>{sliceResult.slice_count} 个切片</Badge>
                  </div>
                  <div className="max-h-48 overflow-y-auto space-y-2">
                    {sliceResult.slices.map((slice) => (
                      <div key={slice.index} className="flex items-center justify-between text-sm bg-gray-50 p-2 rounded">
                        <span className="font-medium">{slice.file_name}</span>
                        <div className="flex items-center gap-4 text-gray-500">
                          <span>{formatDuration(slice.start_time)} - {formatDuration(slice.end_time)}</span>
                          <span>{formatFileSize(slice.size)}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                  <p className="text-sm text-gray-500">
                    总大小: {formatFileSize(sliceResult.total_size)} | 
                    处理时间: {(sliceResult.duration / 1000).toFixed(2)}s
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 缩略图选项卡 */}
        <TabsContent value="thumbnail">
          <Card>
            <CardContent className="pt-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>截图数量</Label>
                  <Input
                    type="number"
                    value={thumbCount}
                    onChange={(e) => setThumbCount(Number(e.target.value))}
                    min={1}
                    max={50}
                  />
                </div>
                <div className="space-y-2">
                  <Label>图片宽度</Label>
                  <Input
                    type="number"
                    value={thumbWidth}
                    onChange={(e) => setThumbWidth(Number(e.target.value))}
                    min={64}
                    max={1920}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <Label>截图间隔 (秒，0表示均匀分布)</Label>
                <Input
                  type="number"
                  value={thumbInterval}
                  onChange={(e) => setThumbInterval(Number(e.target.value))}
                  min={0}
                />
              </div>

              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={generateSprite}
                  onChange={(e) => setGenerateSprite(e.target.checked)}
                  className="rounded"
                />
                生成雪碧图
              </label>

              <Button onClick={handleGenerateThumbnails} disabled={isProcessing || !videoPath} className="w-full">
                <Image className="h-4 w-4 mr-2" />
                生成缩略图
              </Button>

              {/* 缩略图结果 */}
              {thumbResult && (
                <div className="border-t pt-4 space-y-4">
                  <div className="flex items-center justify-between">
                    <h4 className="font-medium">缩略图结果</h4>
                    <Badge>{thumbResult.thumbnails.length} 张</Badge>
                  </div>
                  <div className="grid grid-cols-5 gap-2">
                    {thumbResult.thumbnails.slice(0, 10).map((thumb) => (
                      <div key={thumb.index} className="text-center">
                        <div className="bg-gray-100 rounded aspect-video flex items-center justify-center">
                          <Image className="h-6 w-6 text-gray-400" />
                        </div>
                        <p className="text-xs text-gray-500 mt-1">{formatDuration(thumb.timestamp)}</p>
                      </div>
                    ))}
                  </div>
                  {thumbResult.sprite_sheet && (
                    <div className="bg-blue-50 p-3 rounded-lg">
                      <p className="text-sm text-blue-700">
                        雪碧图: {thumbResult.sprite_sheet.file_name} 
                        ({thumbResult.sprite_sheet.columns}x{thumbResult.sprite_sheet.rows})
                      </p>
                    </div>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* 转换选项卡 */}
        <TabsContent value="convert">
          <Card>
            <CardContent className="pt-6 space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>输出格式</Label>
                  <Select value={outputFormat} onValueChange={setOutputFormat}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {SUPPORTED_FORMATS.map((f) => (
                        <SelectItem key={f} value={f}>{f.toUpperCase()}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label>视频编码</Label>
                  <Select value={videoCodec} onValueChange={setVideoCodec}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {VIDEO_CODECS.map((c) => (
                        <SelectItem key={c.value} value={c.value}>{c.label}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label>分辨率</Label>
                  <Select value={resolution} onValueChange={setResolution}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {RESOLUTION_PRESETS.map((p) => (
                        <SelectItem key={p.name} value={p.name}>{p.name} ({p.width}x{p.height})</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label>编码预设</Label>
                  <Select value={preset} onValueChange={setPreset}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="ultrafast">超快 (低质量)</SelectItem>
                      <SelectItem value="fast">快</SelectItem>
                      <SelectItem value="medium">中等</SelectItem>
                      <SelectItem value="slow">慢 (高质量)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              <div className="space-y-2">
                <Label>质量 (CRF: 0-51, 越小质量越高)</Label>
                <Input
                  type="number"
                  value={crf}
                  onChange={(e) => setCRF(Number(e.target.value))}
                  min={0}
                  max={51}
                />
              </div>

              <Button onClick={handleConvert} disabled={isProcessing || !videoPath} className="w-full">
                <RefreshCw className="h-4 w-4 mr-2" />
                开始转换
              </Button>

              {/* 转换结果 */}
              {convertResult && (
                <div className="border-t pt-4 space-y-4">
                  <h4 className="font-medium">转换结果</h4>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div className="bg-gray-50 p-3 rounded-lg">
                      <p className="text-gray-500 mb-1">源文件</p>
                      <p>{formatFileSize(convertResult.source_info.file_size)}</p>
                      <p className="text-xs text-gray-400">
                        {convertResult.source_info.width}x{convertResult.source_info.height}
                      </p>
                    </div>
                    <div className="bg-green-50 p-3 rounded-lg">
                      <p className="text-gray-500 mb-1">输出文件</p>
                      <p>{formatFileSize(convertResult.output_info.file_size)}</p>
                      <p className="text-xs text-gray-400">
                        {convertResult.output_info.width}x{convertResult.output_info.height}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span>压缩比: {(convertResult.compression_ratio * 100).toFixed(1)}%</span>
                    <span>处理时间: {(convertResult.duration / 1000).toFixed(2)}s</span>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

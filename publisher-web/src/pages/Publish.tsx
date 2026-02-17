import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Badge } from "@/components/ui/badge"
import { useState, useEffect } from "react"
import { getPlatforms, checkLogin, publishAsync } from "@/lib/api"
import type { Platform, AccountStatus } from "@/types/api"


// 文件上传函数
async function uploadFile(file: File): Promise<string | null> {
  const formData = new FormData()
  formData.append('file', file)

  try {
    const response = await fetch('/api/v1/storage/upload', {
      method: 'POST',
      body: formData,
    })

    const result = await response.json()
    if (result.success && result.data) {
      return result.data.storage_path
    } else {
      console.error('文件上传失败:', result.error)
      return null
    }
  } catch (error) {
    console.error('文件上传失败:', error)
    return null
  }
}

const platformNames: Record<Platform, string> = {
  douyin: "抖音",
  toutiao: "今日头条",
  xiaohongshu: "小红书",
}

// 默认平台限制
const platformLimits: Record<Platform, { title_max_length: number; body_max_length: number; max_images: number }> = {
  douyin: { title_max_length: 30, body_max_length: 2000, max_images: 12 },
  toutiao: { title_max_length: 30, body_max_length: 2000, max_images: 9 },
  xiaohongshu: { title_max_length: 20, body_max_length: 1000, max_images: 18 },
}

export default function Publish() {
  const [platforms, setPlatforms] = useState<Platform[]>([])
  const [accountStatuses, setAccountStatuses] = useState<Record<Platform, AccountStatus | null>>({
    douyin: null,
    toutiao: null,
    xiaohongshu: null,
  })
  const [selectedPlatform, setSelectedPlatform] = useState<Platform | null>(null)
  const [contentType, setContentType] = useState<"images" | "video">("images")

  // 表单数据
  const [title, setTitle] = useState("")
  const [body, setBody] = useState("")
  const [images, setImages] = useState<string[]>([])
  const [video, setVideo] = useState("")
  const [tags, setTags] = useState("")

  const [_loading, _setLoading] = useState(true)
  const [publishing, setPublishing] = useState(false)

  useEffect(() => {
    async function fetchData() {
      try {
        const response = await getPlatforms()
        if (response.success && response.data) {
          const platformList = response.data.platforms as Platform[]
          setPlatforms(platformList)

          // 检查登录状态
          for (const platform of platformList) {
            const statusRes = await checkLogin(platform)
            if (statusRes.success && statusRes.data) {
              setAccountStatuses((prev) => ({
                ...prev,
                [platform]: statusRes.data!,
              }))
            }
          }

          // 默认选择第一个平台
          if (platformList.length > 0) {
            setSelectedPlatform(platformList[0])
          }
        }
      } catch (error) {
        console.error("获取数据失败:", error)
      } finally {
        _setLoading(false)
      }
    }

    fetchData()
  }, [])

  async function handlePublish() {
    if (!selectedPlatform) {
      alert("请选择发布平台")
      return
    }

    const status = accountStatuses[selectedPlatform]
    if (!status?.logged_in) {
      alert("请先登录该平台账号")
      return
    }

    if (!title.trim()) {
      alert("请输入标题")
      return
    }

    if (contentType === "images" && images.length === 0) {
      alert("请上传至少一张图片")
      return
    }

    if (contentType === "video" && !video.trim()) {
      alert("请上传视频")
      return
    }

    setPublishing(true)

    try {
      const content = {
        platform: selectedPlatform,
        type: contentType,
        title: title.trim(),
        body: body.trim(),
        images: contentType === "images" ? images : undefined,
        video: contentType === "video" ? video : undefined,
        tags: tags.split(",").map((t) => t.trim()).filter(Boolean),
      }

      const response = await publishAsync(content)
      if (response.success && response.data) {
        alert(`发布任务已创建，任务ID: ${response.data.task_id}`)
        // 重置表单
        setTitle("")
        setBody("")
        setImages([])
        setVideo("")
        setTags("")
      } else {
        alert(response.error || "发布失败")
      }
    } catch (error) {
      console.error("发布失败:", error)
      alert("发布失败，请重试")
    } finally {
      setPublishing(false)
    }
  }

  const currentLimits = selectedPlatform ? platformLimits[selectedPlatform] : null

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">内容发布</h1>
        <p className="text-muted-foreground mt-2">创建并发布内容到各平台</p>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* 平台选择 */}
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>选择平台</CardTitle>
            <CardDescription>选择要发布内容的平台</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {platforms.map((platform) => {
              const status = accountStatuses[platform]
              const isSelected = selectedPlatform === platform

              return (
                <Button
                  key={platform}
                  variant={isSelected ? "default" : "outline"}
                  className="w-full justify-between"
                  onClick={() => setSelectedPlatform(platform)}
                >
                  <span>{platformNames[platform]}</span>
                  <Badge variant={status?.logged_in ? "default" : "secondary"}>
                    {status?.logged_in ? "已登录" : "未登录"}
                  </Badge>
                </Button>
              )
            })}
          </CardContent>
        </Card>

        {/* 发布表单 */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>发布内容</CardTitle>
            <CardDescription>
              {currentLimits && (
                <span>
                  标题最多 {currentLimits.title_max_length} 字，正文最多{" "}
                  {currentLimits.body_max_length} 字
                </span>
              )}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Tabs value={contentType} onValueChange={(v) => setContentType(v as "images" | "video")}>
              <TabsList className="mb-4">
                <TabsTrigger value="images">图文</TabsTrigger>
                <TabsTrigger value="video">视频</TabsTrigger>
              </TabsList>

              <TabsContent value="images" className="space-y-4">
                <div>
                  <Label htmlFor="images">图片上传</Label>
                  <Input
                    id="images"
                    type="file"
                    multiple
                    accept="image/*"
                    className="mt-1"
                    onChange={async (e) => {
                      const files = e.target.files
                      if (files) {
                        // 上传文件到服务器
                        const uploadPromises = Array.from(files).map(file => uploadFile(file))
                        const paths = (await Promise.all(uploadPromises)).filter((p): p is string => p !== null)
                        setImages(paths)
                      }
                    }}
                  />
                  <p className="text-sm text-muted-foreground mt-1">
                    {currentLimits && `最多上传 ${currentLimits.max_images} 张图片`}
                  </p>
                </div>
              </TabsContent>

              <TabsContent value="video" className="space-y-4">
                <div>
                  <Label htmlFor="video">视频上传</Label>
                  <Input
                    id="video"
                    type="file"
                    accept="video/*"
                    className="mt-1"
                    onChange={async (e) => {
                      const file = e.target.files?.[0]
                      if (file) {
                        // 上传文件到服务器
                        const path = await uploadFile(file)
                        if (path) {
                          setVideo(path)
                        }
                      }
                    }}
                  />
                </div>
              </TabsContent>
            </Tabs>

            <div className="space-y-4 mt-4">
              <div>
                <Label htmlFor="title">标题</Label>
                <Input
                  id="title"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="请输入标题"
                  className="mt-1"
                  maxLength={currentLimits?.title_max_length}
                />
                <p className="text-sm text-muted-foreground mt-1">
                  {title.length} / {currentLimits?.title_max_length || 30}
                </p>
              </div>

              <div>
                <Label htmlFor="body">正文</Label>
                <Textarea
                  id="body"
                  value={body}
                  onChange={(e) => setBody(e.target.value)}
                  placeholder="请输入正文内容"
                  className="mt-1 min-h-32"
                  maxLength={currentLimits?.body_max_length}
                />
                <p className="text-sm text-muted-foreground mt-1">
                  {body.length} / {currentLimits?.body_max_length || 2000}
                </p>
              </div>

              <div>
                <Label htmlFor="tags">话题标签</Label>
                <Input
                  id="tags"
                  value={tags}
                  onChange={(e) => setTags(e.target.value)}
                  placeholder="多个标签用逗号分隔，如：美食,生活,旅行"
                  className="mt-1"
                />
              </div>

              <div className="flex gap-2 pt-4">
                <Button onClick={handlePublish} disabled={publishing || _loading}>
                  {publishing ? "发布中..." : "立即发布"}
                </Button>
                <Button variant="outline" disabled={publishing}>
                  定时发布
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

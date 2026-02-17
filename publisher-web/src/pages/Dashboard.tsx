import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Link } from "react-router-dom"
import { useEffect, useState } from "react"
import { getPlatforms, checkLogin } from "@/lib/api"
import type { Platform, AccountStatus } from "@/types/api"

const platformNames: Record<Platform, string> = {
  douyin: "抖音",
  toutiao: "今日头条",
  xiaohongshu: "小红书",
}

export default function Dashboard() {
  const [platforms, setPlatforms] = useState<Platform[]>([])
  const [accountStatuses, setAccountStatuses] = useState<Record<Platform, AccountStatus | null>>({
    douyin: null,
    toutiao: null,
    xiaohongshu: null,
  })
  const [_loading, _setLoading] = useState(true)

  useEffect(() => {
    async function fetchData() {
      try {
        const response = await getPlatforms()
        if (response.success && response.data) {
          const platformList = response.data.platforms as Platform[]
          setPlatforms(platformList)

          // 检查每个平台的登录状态
          for (const platform of platformList) {
            const statusRes = await checkLogin(platform)
            if (statusRes.success && statusRes.data) {
              setAccountStatuses((prev) => ({
                ...prev,
                [platform]: statusRes.data!,
              }))
            }
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

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">仪表盘</h1>
        <p className="text-muted-foreground mt-2">管理您的多平台内容发布</p>
      </div>

      {/* 快速操作 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-8">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">已连接平台</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {Object.values(accountStatuses).filter((s) => s?.logged_in).length}
            </div>
            <p className="text-xs text-muted-foreground">共 {platforms.length} 个平台</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">今日发布</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0</div>
            <p className="text-xs text-muted-foreground">条内容</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待发布</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0</div>
            <p className="text-xs text-muted-foreground">条内容</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">发布成功</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">0</div>
            <p className="text-xs text-muted-foreground">总发布数</p>
          </CardContent>
        </Card>
      </div>

      {/* 平台状态 */}
      <div className="mb-8">
        <h2 className="text-xl font-semibold mb-4">平台状态</h2>
        <div className="grid gap-4 md:grid-cols-3">
          {platforms.map((platform) => {
            const status = accountStatuses[platform]
            return (
              <Card key={platform}>
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-lg">{platformNames[platform]}</CardTitle>
                    <Badge variant={status?.logged_in ? "default" : "secondary"}>
                      {status?.logged_in ? "已登录" : "未登录"}
                    </Badge>
                  </div>
                  <CardDescription>
                    {status?.logged_in ? status.account_name || "已连接" : "点击管理账号进行登录"}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" asChild>
                      <Link to="/accounts">管理账号</Link>
                    </Button>
                    {status?.logged_in && (
                      <Button size="sm" asChild>
                        <Link to="/publish">发布内容</Link>
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      </div>

      {/* 快速发布 */}
      <Card>
        <CardHeader>
          <CardTitle>快速开始</CardTitle>
          <CardDescription>选择一个平台开始发布内容</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <Button asChild>
              <Link to="/publish">发布图文</Link>
            </Button>
            <Button variant="outline" asChild>
              <Link to="/publish">发布视频</Link>
            </Button>
            <Button variant="secondary" asChild>
              <Link to="/history">查看历史</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

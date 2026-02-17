import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { useEffect, useState } from "react"
import { getPlatforms, checkLogin, login, logout } from "@/lib/api"
import type { Platform, AccountStatus } from "@/types/api"

const platformNames: Record<Platform, string> = {
  douyin: "抖音",
  toutiao: "今日头条",
  xiaohongshu: "小红书",
}

const platformDescriptions: Record<Platform, string> = {
  douyin: "短视频平台，适合发布视频和图文内容",
  toutiao: "新闻资讯平台，适合发布文章和视频",
  xiaohongshu: "生活方式平台，适合发布图文笔记",
}

export default function Accounts() {
  const [platforms, setPlatforms] = useState<Platform[]>([])
  const [accountStatuses, setAccountStatuses] = useState<Record<Platform, AccountStatus | null>>({
    douyin: null,
    toutiao: null,
    xiaohongshu: null,
  })
  const [_loading, _setLoading] = useState(true)
  const [loginning, setLoginning] = useState<Platform | null>(null)
  const [qrcodeUrl, setQrcodeUrl] = useState<string | null>(null)

  useEffect(() => {
    async function fetchData() {
      try {
        const response = await getPlatforms()
        if (response.success && response.data) {
          const platformList = response.data.platforms as Platform[]
          setPlatforms(platformList)
          await checkAllStatuses(platformList)
        }
      } catch (error) {
        console.error("获取数据失败:", error)
      } finally {
        _setLoading(false)
      }
    }

    fetchData()
  }, [])

  async function checkAllStatuses(platforms: Platform[]) {
    for (const platform of platforms) {
      try {
        const statusRes = await checkLogin(platform)
        if (statusRes.success && statusRes.data) {
          setAccountStatuses((prev) => ({
            ...prev,
            [platform]: statusRes.data!,
          }))
        }
      } catch (error) {
        console.error(`检查 ${platform} 状态失败:`, error)
      }
    }
  }

  async function handleLogin(platform: Platform) {
    setLoginning(platform)
    setQrcodeUrl(null)

    try {
      const response = await login(platform)
      if (response.success && response.data) {
        if (response.data.qrcode_url) {
          setQrcodeUrl(response.data.qrcode_url)
        } else if (response.data.success) {
          // 已经登录成功
          await checkAllStatuses(platforms)
        }
      }
    } catch (error) {
      console.error("登录失败:", error)
      alert("登录失败，请重试")
    } finally {
      setLoginning(null)
    }
  }

  async function handleLogout(platform: Platform) {
    if (!confirm("确定要登出该账号吗？")) {
      return
    }

    try {
      const response = await logout(platform)
      if (response.success) {
        // 清除登录状态
        setAccountStatuses((prev) => ({
          ...prev,
          [platform]: null,
        }))
        alert("登出成功")
      } else {
        alert(response.error || "登出失败")
      }
    } catch (error) {
      console.error("登出失败:", error)
      alert("登出失败，请重试")
    }
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">账号管理</h1>
        <p className="text-muted-foreground mt-2">管理各平台的登录状态和账号信息</p>
      </div>

      {/* 二维码弹窗 */}
      {qrcodeUrl && (
        <Card className="mb-8 border-primary">
          <CardHeader>
            <CardTitle>扫码登录</CardTitle>
            <CardDescription>请使用手机 APP 扫描二维码完成登录</CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col items-center">
            <div className="w-64 h-64 bg-muted rounded-lg flex items-center justify-center mb-4">
              <img src={qrcodeUrl} alt="登录二维码" className="max-w-full max-h-full" />
            </div>
            <Button variant="outline" onClick={() => setQrcodeUrl(null)}>
              关闭
            </Button>
          </CardContent>
        </Card>
      )}

      {/* 平台列表 */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {platforms.map((platform) => {
          const status = accountStatuses[platform]
          const isLoginning = loginning === platform

          return (
            <Card key={platform}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-xl">{platformNames[platform]}</CardTitle>
                  <Badge variant={status?.logged_in ? "default" : "secondary"}>
                    {status?.logged_in ? "已连接" : "未连接"}
                  </Badge>
                </div>
                <CardDescription>{platformDescriptions[platform]}</CardDescription>
              </CardHeader>
              <CardContent>
                {status?.logged_in ? (
                  <div className="space-y-4">
                    <div className="flex items-center gap-3">
                      {status.avatar && (
                        <img
                          src={status.avatar}
                          alt="头像"
                          className="w-10 h-10 rounded-full"
                        />
                      )}
                      <div>
                        <p className="font-medium">{status.account_name || "已登录"}</p>
                        <p className="text-sm text-muted-foreground">
                          上次检查: {status.last_check || "-"}
                        </p>
                      </div>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => checkAllStatuses([platform])}
                      >
                        刷新状态
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => handleLogout(platform)}
                      >
                        登出
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <p className="text-sm text-muted-foreground">
                      点击下方按钮，使用二维码登录您的 {platformNames[platform]} 账号
                    </p>
                    <Button
                      className="w-full"
                      onClick={() => handleLogin(platform)}
                      disabled={isLoginning}
                    >
                      {isLoginning ? "正在生成二维码..." : "登录账号"}
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          )
        })}
      </div>
    </div>
  )
}

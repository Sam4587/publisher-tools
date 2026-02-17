import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useEffect, useState } from "react"
import { getTasks, cancelTask } from "@/lib/api"
import type { Task, PublishStatus } from "@/types/api"

const platformNames: Record<string, string> = {
  douyin: "抖音",
  toutiao: "今日头条",
  xiaohongshu: "小红书",
}

const statusColors: Record<PublishStatus, "secondary" | "default" | "destructive"> = {
  pending: "secondary",
  processing: "default",
  success: "default",
  failed: "destructive",
}

const statusText: Record<PublishStatus, string> = {
  pending: "等待中",
  processing: "处理中",
  success: "成功",
  failed: "失败",
}

export default function History() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [statusFilter, setStatusFilter] = useState<string>("")

  useEffect(() => {
    async function fetchTasks() {
      try {
        const response = await getTasks(statusFilter || undefined)
        if (response.success && response.data) {
          setTasks(response.data)
        }
      } catch (error) {
        console.error("获取任务列表失败:", error)
      } finally {
        setLoading(false)
      }
    }

    fetchTasks()
  }, [statusFilter])

  async function handleCancel(taskId: string) {
    if (!confirm("确定要取消这个任务吗？")) {
      return
    }

    try {
      const response = await cancelTask(taskId)
      if (response.success) {
        setTasks(tasks.map((t) => (t.id === taskId ? { ...t, status: "cancelled" as PublishStatus } : t)))
      } else {
        alert(response.error || "取消失败")
      }
    } catch (error) {
      console.error("取消任务失败:", error)
      alert("取消失败，请重试")
    }
  }

  function formatDate(dateStr: string) {
    const date = new Date(dateStr)
    return date.toLocaleString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    })
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">发布历史</h1>
        <p className="text-muted-foreground mt-2">查看和管理您的发布任务</p>
      </div>

      {/* 筛选器 */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>筛选</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            <Button
              variant={statusFilter === "" ? "default" : "outline"}
              size="sm"
              onClick={() => setStatusFilter("")}
            >
              全部
            </Button>
            <Button
              variant={statusFilter === "pending" ? "default" : "outline"}
              size="sm"
              onClick={() => setStatusFilter("pending")}
            >
              等待中
            </Button>
            <Button
              variant={statusFilter === "processing" ? "default" : "outline"}
              size="sm"
              onClick={() => setStatusFilter("processing")}
            >
              处理中
            </Button>
            <Button
              variant={statusFilter === "success" ? "default" : "outline"}
              size="sm"
              onClick={() => setStatusFilter("success")}
            >
              成功
            </Button>
            <Button
              variant={statusFilter === "failed" ? "default" : "outline"}
              size="sm"
              onClick={() => setStatusFilter("failed")}
            >
              失败
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* 任务列表 */}
      <Card>
        <CardHeader>
          <CardTitle>任务列表</CardTitle>
          <CardDescription>共 {tasks.length} 个任务</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-8 text-muted-foreground">加载中...</div>
          ) : tasks.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">暂无发布任务</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>任务ID</TableHead>
                  <TableHead>平台</TableHead>
                  <TableHead>类型</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>进度</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tasks.map((task) => (
                  <TableRow key={task.id}>
                    <TableCell className="font-mono text-sm">{task.id.slice(0, 8)}...</TableCell>
                    <TableCell>{platformNames[task.platform] || task.platform}</TableCell>
                    <TableCell>{task.type === "images" ? "图文" : "视频"}</TableCell>
                    <TableCell>
                      <Badge variant={statusColors[task.status as PublishStatus]}>
                        {statusText[task.status as PublishStatus]}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <div className="w-16 bg-muted rounded-full h-2">
                          <div
                            className="bg-primary h-2 rounded-full"
                            style={{ width: `${task.progress}%` }}
                          />
                        </div>
                        <span className="text-sm text-muted-foreground">{task.progress}%</span>
                      </div>
                    </TableCell>
                    <TableCell className="text-sm">{formatDate(task.created_at)}</TableCell>
                    <TableCell>
                      {task.status === "pending" || task.status === "processing" ? (
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => handleCancel(task.id)}
                        >
                          取消
                        </Button>
                      ) : task.status === "success" && task.result?.post_url ? (
                        <Button variant="outline" size="sm" asChild>
                          <a href={task.result.post_url as string} target="_blank" rel="noreferrer">
                            查看
                          </a>
                        </Button>
                      ) : task.status === "failed" ? (
                        <span className="text-sm text-destructive">{task.error || "发布失败"}</span>
                      ) : (
                        "-"
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

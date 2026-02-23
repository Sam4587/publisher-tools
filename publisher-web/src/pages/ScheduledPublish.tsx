import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger, DialogFooter } from "@/components/ui/dialog"
import { Clock, Plus, Play, Pause, Trash2, Edit } from 'lucide-react'
import { Switch } from "@/components/ui/switch"

interface ScheduledTask {
  id: number
  name: string
  task_type: string
  cron_expr: string
  queue_name: string
  is_active: boolean
  last_run_at: string | null
  next_run_at: string | null
  run_count: number
  last_error: string | null
  created_at: string
  updated_at: string
}

interface SchedulerStats {
  total_jobs: number
  active_jobs: number
  inactive_jobs: number
  running_since: string
}

export default function ScheduledPublish() {
  const [tasks, setTasks] = useState<ScheduledTask[]>([])
  const [stats, setStats] = useState<SchedulerStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [editingTask, setEditingTask] = useState<ScheduledTask | null>(null)

  // 表单数据
  const [formData, setFormData] = useState({
    name: '',
    task_type: 'publish',
    cron_expr: '',
    queue_name: 'default',
    payload: ''
  })

  useEffect(() => {
    fetchData()
  }, [])

  async function fetchData() {
    try {
      const [tasksRes, statsRes] = await Promise.all([
        fetch('/api/v1/scheduler/tasks'),
        fetch('/api/v1/scheduler/stats')
      ])

      const tasksData = await tasksRes.json()
      const statsData = await statsRes.json()

      if (tasksData.success) {
        setTasks(tasksData.data.tasks || [])
      }
      if (statsData.success) {
        setStats(statsData.data)
      }
    } catch (error) {
      console.error('Failed to fetch data:', error)
    } finally {
      setLoading(false)
    }
  }

  async function handleSubmit() {
    try {
      const response = await fetch('/api/v1/scheduler/tasks', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: formData.name,
          task_type: formData.task_type,
          cron_expr: formData.cron_expr,
          queue_name: formData.queue_name,
          payload: formData.payload ? JSON.parse(formData.payload) : {}
        })
      })

      const result = await response.json()
      if (result.success) {
        setIsDialogOpen(false)
        resetForm()
        fetchData()
      } else {
        alert(result.error || '创建失败')
      }
    } catch (error) {
      console.error('Failed to create task:', error)
      alert('创建失败,请重试')
    }
  }

  async function handlePause(name: string) {
    try {
      const response = await fetch(`/api/v1/scheduler/tasks/${name}/pause`, {
        method: 'POST'
      })
      if (response.ok) {
        fetchData()
      }
    } catch (error) {
      console.error('Failed to pause task:', error)
    }
  }

  async function handleResume(name: string) {
    try {
      const response = await fetch(`/api/v1/scheduler/tasks/${name}/resume`, {
        method: 'POST'
      })
      if (response.ok) {
        fetchData()
      }
    } catch (error) {
      console.error('Failed to resume task:', error)
    }
  }

  async function handleDelete(name: string) {
    if (!confirm('确定要删除这个定时任务吗?')) {
      return
    }

    try {
      const response = await fetch(`/api/v1/scheduler/tasks/${name}`, {
        method: 'DELETE'
      })
      if (response.ok) {
        fetchData()
      }
    } catch (error) {
      console.error('Failed to delete task:', error)
    }
  }

  async function handleRunNow(name: string) {
    if (!confirm('确定要立即执行这个任务吗?')) {
      return
    }

    try {
      const response = await fetch(`/api/v1/scheduler/tasks/${name}/run`, {
        method: 'POST'
      })
      if (response.ok) {
        alert('任务已触发执行')
      }
    } catch (error) {
      console.error('Failed to run task:', error)
    }
  }

  function resetForm() {
    setFormData({
      name: '',
      task_type: 'publish',
      cron_expr: '',
      queue_name: 'default',
      payload: ''
    })
    setEditingTask(null)
  }

  function openEditDialog(task: ScheduledTask) {
    setEditingTask(task)
    setFormData({
      name: task.name,
      task_type: task.task_type,
      cron_expr: task.cron_expr,
      queue_name: task.queue_name,
      payload: ''
    })
    setIsDialogOpen(true)
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">定时发布</h1>
        <p className="text-muted-foreground mt-2">管理定时发布任务</p>
      </div>

      {/* 统计信息 */}
      {stats && (
        <div className="grid gap-4 md:grid-cols-3 mb-8">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">总任务数</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total_jobs}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">运行中</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600">{stats.active_jobs}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">已暂停</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-gray-600">{stats.inactive_jobs}</div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* 创建任务按钮 */}
      <div className="mb-6">
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button onClick={() => { resetForm(); setEditingTask(null) }}>
              <Plus className="h-4 w-4 mr-2" />
              创建定时任务
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{editingTask ? '编辑定时任务' : '创建定时任务'}</DialogTitle>
              <DialogDescription>
                配置定时发布任务,使用Cron表达式设置执行时间
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <div>
                <Label htmlFor="name">任务名称</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="例如:每日热点内容发布"
                />
              </div>
              <div>
                <Label htmlFor="cron_expr">Cron表达式</Label>
                <Input
                  id="cron_expr"
                  value={formData.cron_expr}
                  onChange={(e) => setFormData({ ...formData, cron_expr: e.target.value })}
                  placeholder="例如: 0 9 * * * (每天9点)"
                />
                <p className="text-sm text-muted-foreground mt-1">
                  格式: 秒 分 时 日 月 周
                </p>
              </div>
              <div>
                <Label htmlFor="task_type">任务类型</Label>
                <Input
                  id="task_type"
                  value={formData.task_type}
                  onChange={(e) => setFormData({ ...formData, task_type: e.target.value })}
                  placeholder="publish"
                />
              </div>
              <div>
                <Label htmlFor="queue_name">队列名称</Label>
                <Input
                  id="queue_name"
                  value={formData.queue_name}
                  onChange={(e) => setFormData({ ...formData, queue_name: e.target.value })}
                  placeholder="default"
                />
              </div>
              <div>
                <Label htmlFor="payload">任务配置 (JSON)</Label>
                <Textarea
                  id="payload"
                  value={formData.payload}
                  onChange={(e) => setFormData({ ...formData, payload: e.target.value })}
                  placeholder={`{\n  "platform": "douyin",\n  "type": "images"\n}`}
                  className="font-mono"
                />
              </div>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setIsDialogOpen(false)}>
                取消
              </Button>
              <Button onClick={handleSubmit}>
                {editingTask ? '更新' : '创建'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {/* 任务列表 */}
      <div className="space-y-4">
        {loading ? (
          <Card>
            <CardContent className="py-8">
              <p className="text-center text-muted-foreground">加载中...</p>
            </CardContent>
          </Card>
        ) : tasks.length === 0 ? (
          <Card>
            <CardContent className="py-8">
              <p className="text-center text-muted-foreground">暂无定时任务</p>
            </CardContent>
          </Card>
        ) : (
          tasks.map((task) => (
            <Card key={task.id}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Clock className="h-5 w-5 text-blue-600" />
                    <div>
                      <CardTitle className="text-lg">{task.name}</CardTitle>
                      <CardDescription className="mt-1">
                        {task.cron_expr} • 已执行 {task.run_count} 次
                      </CardDescription>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={task.is_active ? "default" : "secondary"}>
                      {task.is_active ? "运行中" : "已暂停"}
                    </Badge>
                    <Badge variant="outline">{task.task_type}</Badge>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="grid gap-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">上次执行:</span>
                    <span>{task.last_run_at ? new Date(task.last_run_at).toLocaleString() : '未执行'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">下次执行:</span>
                    <span>{task.next_run_at ? new Date(task.next_run_at).toLocaleString() : '待定'}</span>
                  </div>
                  {task.last_error && (
                    <div className="flex justify-between text-red-600">
                      <span className="text-muted-foreground">错误:</span>
                      <span className="truncate max-w-md">{task.last_error}</span>
                    </div>
                  )}
                </div>
                <div className="flex gap-2 mt-4">
                  {task.is_active ? (
                    <Button variant="outline" size="sm" onClick={() => handlePause(task.name)}>
                      <Pause className="h-4 w-4 mr-1" />
                      暂停
                    </Button>
                  ) : (
                    <Button variant="outline" size="sm" onClick={() => handleResume(task.name)}>
                      <Play className="h-4 w-4 mr-1" />
                      恢复
                    </Button>
                  )}
                  <Button variant="outline" size="sm" onClick={() => handleRunNow(task.name)}>
                    <Play className="h-4 w-4 mr-1" />
                    立即执行
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => openEditDialog(task)}>
                    <Edit className="h-4 w-4 mr-1" />
                    编辑
                  </Button>
                  <Button variant="destructive" size="sm" onClick={() => handleDelete(task.name)}>
                    <Trash2 className="h-4 w-4 mr-1" />
                    删除
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>
    </div>
  )
}

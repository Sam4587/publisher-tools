import { useState, useEffect, useCallback, useRef } from 'react'
import { Loader2, CheckCircle2, XCircle, Clock, Wifi, WifiOff, RefreshCw, AlertCircle } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Button } from '@/components/ui/button'

// WebSocket消息类型
interface WebSocketMessage {
  type: string
  task_id: string
  payload: ProgressPayload
}

// 进度消息
interface ProgressPayload {
  task_id: string
  progress: number
  current_step: string
  total_steps: number
  completed_steps: number
  message: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  timestamp: string
}

// 任务进度状态
interface TaskProgress {
  taskId: string
  progress: number
  currentStep: string
  totalSteps: number
  completedSteps: number
  message: string
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  lastUpdated: Date
  history: ProgressPayload[]
}

// 组件属性
interface RealtimeProgressTrackerProps {
  taskId: string
  userId?: string
  projectId?: string
  wsUrl?: string
  onProgress?: (progress: TaskProgress) => void
  onComplete?: (progress: TaskProgress) => void
  onError?: (error: string) => void
  onReconnect?: (clientId: string) => void
  showHistory?: boolean
  autoReconnect?: boolean
  maxReconnectAttempts?: number
}

export default function RealtimeProgressTracker({
  taskId,
  userId = 'anonymous',
  projectId,
  wsUrl = `ws://${window.location.host}/api/v1/ws`,
  onProgress,
  onComplete,
  onError,
  onReconnect,
  showHistory = false,
  autoReconnect = true,
  maxReconnectAttempts = 5,
}: RealtimeProgressTrackerProps) {
  const [progress, setProgress] = useState<TaskProgress>({
    taskId,
    progress: 0,
    currentStep: '',
    totalSteps: 0,
    completedSteps: 0,
    message: '等待连接...',
    status: 'pending',
    lastUpdated: new Date(),
    history: [],
  })
  const [isConnected, setIsConnected] = useState(false)
  const [clientId, setClientId] = useState<string | null>(null)
  const [reconnectAttempts, setReconnectAttempts] = useState(0)
  const [error, setError] = useState<string | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  // 连接WebSocket
  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    // 构建WebSocket URL
    let url = `${wsUrl}?user_id=${userId}`
    if (projectId) {
      url += `&project_id=${projectId}`
    }
    if (clientId && autoReconnect) {
      url += `&old_client_id=${clientId}`
    }

    try {
      const ws = new WebSocket(url)
      wsRef.current = ws

      ws.onopen = () => {
        setIsConnected(true)
        setError(null)
        setReconnectAttempts(0)
        console.log('WebSocket已连接')

        // 订阅任务
        ws.send(JSON.stringify({
          action: 'subscribe',
          task_id: taskId,
        }))
      }

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data)
          handleMessage(message)
        } catch (err) {
          console.error('解析消息失败:', err)
        }
      }

      ws.onerror = (event) => {
        console.error('WebSocket错误:', event)
        setError('WebSocket连接错误')
        onError?.('WebSocket连接错误')
      }

      ws.onclose = () => {
        setIsConnected(false)
        console.log('WebSocket已断开')

        // 自动重连
        if (autoReconnect && reconnectAttempts < maxReconnectAttempts) {
          const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
          console.log(`${delay}ms后尝试重连... (${reconnectAttempts + 1}/${maxReconnectAttempts})`)
          
          reconnectTimeoutRef.current = setTimeout(() => {
            setReconnectAttempts(prev => prev + 1)
            connect()
          }, delay)
        }
      }
    } catch (err) {
      console.error('创建WebSocket失败:', err)
      setError('创建WebSocket失败')
      onError?.('创建WebSocket失败')
    }
  }, [wsUrl, userId, projectId, clientId, taskId, autoReconnect, reconnectAttempts, maxReconnectAttempts, onError])

  // 处理消息
  const handleMessage = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case 'connected':
        setClientId(message.payload.client_id)
        console.log('已连接，客户端ID:', message.payload.client_id)
        break

      case 'reconnected':
        setClientId(message.payload.client_id)
        setReconnectAttempts(0)
        onReconnect?.(message.payload.client_id)
        console.log('重连成功，客户端ID:', message.payload.client_id)
        break

      case 'progress':
        const payload = message.payload as ProgressPayload
        setProgress(prev => {
          const newProgress: TaskProgress = {
            taskId: payload.task_id,
            progress: payload.progress,
            currentStep: payload.current_step,
            totalSteps: payload.total_steps,
            completedSteps: payload.completed_steps,
            message: payload.message,
            status: payload.status,
            lastUpdated: new Date(payload.timestamp),
            history: showHistory 
              ? [...prev.history.slice(-99), payload]
              : prev.history,
          }
          onProgress?.(newProgress)
          return newProgress
        })
        break

      case 'status':
        const statusPayload = message.payload
        setProgress(prev => ({
          ...prev,
          status: statusPayload.status,
          message: statusPayload.error || prev.message,
          lastUpdated: new Date(),
        }))
        
        if (statusPayload.status === 'completed') {
          onComplete?.(progress)
        }
        break

      case 'error':
        setError(message.payload.error || '未知错误')
        onError?.(message.payload.error || '未知错误')
        break

      case 'subscribed':
        console.log('已订阅任务:', message.payload.topics)
        break

      case 'pong':
        // 心跳响应
        break

      default:
        console.log('未知消息类型:', message.type)
    }
  }, [onProgress, onComplete, onError, progress, showHistory, onReconnect])

  // 断开连接
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setIsConnected(false)
  }, [])

  // 手动重连
  const handleReconnect = useCallback(() => {
    disconnect()
    setReconnectAttempts(0)
    connect()
  }, [disconnect, connect])

  // 组件挂载时连接
  useEffect(() => {
    connect()
    return () => {
      disconnect()
    }
  }, [connect, disconnect])

  // 发送心跳
  useEffect(() => {
    if (!isConnected) return

    const interval = setInterval(() => {
      if (wsRef.current?.readyState === WebSocket.OPEN) {
        wsRef.current.send(JSON.stringify({ type: 'ping' }))
      }
    }, 30000)

    return () => clearInterval(interval)
  }, [isConnected])

  // 获取状态图标
  const getStatusIcon = () => {
    switch (progress.status) {
      case 'running':
        return <Loader2 className="h-4 w-4 animate-spin text-blue-600" />
      case 'completed':
        return <CheckCircle2 className="h-4 w-4 text-green-600" />
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-600" />
      case 'cancelled':
        return <AlertCircle className="h-4 w-4 text-orange-600" />
      default:
        return <Clock className="h-4 w-4 text-gray-400" />
    }
  }

  // 获取状态徽章
  const getStatusBadge = () => {
    switch (progress.status) {
      case 'running':
        return <Badge variant="secondary" className="animate-pulse">处理中</Badge>
      case 'completed':
        return <Badge variant="default" className="bg-green-500">已完成</Badge>
      case 'failed':
        return <Badge variant="destructive">失败</Badge>
      case 'cancelled':
        return <Badge variant="outline">已取消</Badge>
      default:
        return <Badge variant="outline">等待中</Badge>
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <CardTitle className="text-lg">任务进度</CardTitle>
            {getStatusBadge()}
            <div className="flex items-center gap-1">
              {isConnected ? (
                <Wifi className="h-4 w-4 text-green-500" />
              ) : (
                <WifiOff className="h-4 w-4 text-red-500" />
              )}
              <span className="text-xs text-muted-foreground">
                {isConnected ? '已连接' : '未连接'}
              </span>
            </div>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleReconnect}
            disabled={isConnected && progress.status === 'running'}
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            重连
          </Button>
        </div>
        <CardDescription>
          任务ID: {taskId}
          {clientId && <span className="ml-2 text-xs">客户端: {clientId.slice(0, 8)}...</span>}
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* 错误提示 */}
        {error && (
          <div className="flex items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700">
            <AlertCircle className="h-4 w-4" />
            <span className="text-sm">{error}</span>
          </div>
        )}

        {/* 重连提示 */}
        {!isConnected && reconnectAttempts > 0 && reconnectAttempts < maxReconnectAttempts && (
          <div className="flex items-center gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-lg text-yellow-700">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span className="text-sm">
              正在重连... ({reconnectAttempts}/{maxReconnectAttempts})
            </span>
          </div>
        )}

        {/* 进度条 */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <div className="flex items-center gap-2">
              {getStatusIcon()}
              <span className="font-medium">{progress.currentStep || '准备中...'}</span>
            </div>
            <span className="font-medium">{progress.progress}%</span>
          </div>
          <Progress value={progress.progress} className="h-2" />
        </div>

        {/* 步骤进度 */}
        {progress.totalSteps > 0 && (
          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <span>步骤: {progress.completedSteps} / {progress.totalSteps}</span>
            <span>
              更新于: {progress.lastUpdated.toLocaleTimeString()}
            </span>
          </div>
        )}

        {/* 消息 */}
        <p className="text-sm text-gray-600">{progress.message}</p>

        {/* 历史记录 */}
        {showHistory && progress.history.length > 0 && (
          <div className="mt-4 border-t pt-4">
            <h4 className="text-sm font-medium mb-2">进度历史</h4>
            <div className="max-h-40 overflow-y-auto space-y-1">
              {progress.history.slice().reverse().map((item, index) => (
                <div key={index} className="flex items-center justify-between text-xs text-muted-foreground">
                  <span>{item.current_step || item.message}</span>
                  <span>{item.progress}%</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

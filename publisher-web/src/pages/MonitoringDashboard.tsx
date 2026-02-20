import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Statistic, Progress, Button, Badge, Timeline, Modal } from 'antd';
import { PlayCircleOutlined, PauseCircleOutlined, StopOutlined, FullscreenOutlined } from '@ant-design/icons';

interface MonitoringStats {
  running: number;
  pending: number;
  today: number;
  week: number;
  successRate: number;
  failed: number;
}

interface Execution {
  id: string;
  pipelineId: string;
  pipelineName: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'paused' | 'cancelled';
  progress: number;
  currentStep: string;
  totalSteps: number;
  startedAt: string;
  finishedAt?: string;
  error?: string;
  steps: ExecutionStep[];
}

interface ExecutionStep {
  stepId: string;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress: number;
  startedAt: string;
  finishedAt?: string;
  error?: string;
}

const MonitoringDashboard: React.FC = () => {
  const [stats, setStats] = useState<MonitoringStats>({
    running: 0,
    pending: 0,
    today: 0,
    week: 0,
    successRate: 0,
    failed: 0,
  });
  const [runningExecutions, setRunningExecutions] = useState<Execution[]>([]);
  const [recentExecutions, setRecentExecutions] = useState<Execution[]>([]);
  const [selectedExecution, setSelectedExecution] = useState<Execution | null>(null);
  const [isDetailModalVisible, setIsDetailModalVisible] = useState(false);
  const [ws, setWs] = useState<WebSocket | null>(null);

  // WebSocket 连接
  useEffect(() => {
    const websocket = new WebSocket('ws://localhost:8080/ws/monitor');

    websocket.onopen = () => {
      console.log('WebSocket connected');
    };

    websocket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      handleRealtimeUpdate(message);
    };

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    websocket.onclose = () => {
      console.log('WebSocket disconnected');
    };

    setWs(websocket);

    return () => {
      websocket.close();
    };
  }, []);

  // 加载数据
  useEffect(() => {
    loadStats();
    loadRunningExecutions();
    loadRecentExecutions();
  }, []);

  const loadStats = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/monitoring/stats');
      const data = await response.json();
      setStats(data);
    } catch (error) {
      console.error('加载统计数据失败');
    }
  };

  const loadRunningExecutions = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/executions?status=running');
      const data = await response.json();
      setRunningExecutions(data);
    } catch (error) {
      console.error('加载执行中任务失败');
    }
  };

  const loadRecentExecutions = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/executions?limit=10');
      const data = await response.json();
      setRecentExecutions(data.filter((e: Execution) => e.status === 'completed'));
    } catch (error) {
      console.error('加载最近完成任务失败');
    }
  };

  const handleRealtimeUpdate = (message: any) => {
    switch (message.type) {
      case 'progress':
        updateExecutionProgress(message.data);
        break;
      case 'status_change':
        updateExecutionStatus(message.data);
        loadStats();
        break;
      case 'error':
        updateExecutionError(message.data);
        break;
      case 'completed':
        handleExecutionCompleted(message.data);
        loadStats();
        break;
    }
  };

  const updateExecutionProgress = (data: any) => {
    setRunningExecutions(prev => prev.map(exec =>
      exec.id === data.execution_id ? { ...exec, progress: data.progress, currentStep: data.current_step } : exec
    ));
  };

  const updateExecutionStatus = (data: any) => {
    if (data.status === 'completed' || data.status === 'failed') {
      setRunningExecutions(prev => prev.filter(e => e.id !== data.execution_id));
    }
  };

  const updateExecutionError = (data: any) => {
    setRunningExecutions(prev => prev.map(exec =>
      exec.id === data.execution_id ? { ...exec, error: data.error } : exec
    ));
  };

  const handleExecutionCompleted = (data: any) => {
    // 从运行中移除，添加到最近完成
    setRunningExecutions(prev => prev.filter(e => e.id !== data.execution_id));
    loadRecentExecutions();
  };

  const handlePauseExecution = async (executionId: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/executions/${executionId}/pause`, {
        method: 'POST',
      });

      if (response.ok) {
        loadRunningExecutions();
      }
    } catch (error) {
      console.error('暂停失败');
    }
  };

  const handleResumeExecution = async (executionId: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/executions/${executionId}/resume`, {
        method: 'POST',
      });

      if (response.ok) {
        loadRunningExecutions();
      }
    } catch (error) {
      console.error('恢复失败');
    }
  };

  const handleCancelExecution = async (executionId: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/executions/${executionId}/cancel`, {
        method: 'POST',
      });

      if (response.ok) {
        loadRunningExecutions();
      }
    } catch (error) {
      console.error('取消失败');
    }
  };

  const handleViewDetail = async (execution: Execution) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/executions/${execution.id}`);
      const data = await response.json();
      setSelectedExecution(data);
      setIsDetailModalVisible(true);
    } catch (error) {
      console.error('加载详情失败');
    }
  };

  const getStatusBadge = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      pending: { color: 'default', text: '等待中' },
      running: { color: 'processing', text: '运行中' },
      completed: { color: 'success', text: '已完成' },
      failed: { color: 'error', text: '失败' },
      paused: { color: 'warning', text: '已暂停' },
      cancelled: { color: 'default', text: '已取消' },
    };
    const { color, text } = statusMap[status] || { color: 'default', text: status };
    return <Badge color={color} text={text} />;
  };

  const calculateDuration = (startedAt: string, finishedAt?: string) => {
    const start = new Date(startedAt);
    const end = finishedAt ? new Date(finishedAt) : new Date();
    const duration = Math.floor((end.getTime() - start.getTime()) / 1000 / 60);
    return `${duration}分钟`;
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>实时监控面板</h1>
        <Button icon={<FullscreenOutlined />}>全屏</Button>
      </div>

      {/* 统计卡片 */}
      <Row gutter={16} style={{ marginBottom: '24px' }}>
        <Col span={6}>
          <Card>
            <Statistic title="运行中" value={stats.running} valueStyle={{ color: '#1890ff' }} />
            <div style={{ marginTop: '8px', fontSize: '14px', color: '#888' }}>等待中: {stats.pending}</div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="今日执行" value={stats.today} valueStyle={{ color: '#52c41a' }} />
            <div style={{ marginTop: '8px', fontSize: '14px', color: '#888' }}>本周: {stats.week}</div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="成功率" value={stats.successRate} suffix="%" valueStyle={{ color: '#52c41a' }} />
            <div style={{ marginTop: '8px', fontSize: '14px', color: '#888' }}>失败: {stats.failed}</div>
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic title="平均耗时" value={12} suffix="分钟" valueStyle={{ color: '#faad14' }} />
          </Card>
        </Col>
      </Row>

      {/* 执行中任务 */}
      <Card title="执行中任务" style={{ marginBottom: '24px' }}>
        {runningExecutions.length === 0 ? (
          <p style={{ textAlign: 'center', color: '#888' }}>暂无执行中的任务</p>
        ) : (
          <div style={{ display: 'grid', gap: '16px' }}>
            {runningExecutions.map((execution) => (
              <Card key={execution.id} size="small">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
                  <div>
                    <strong>{execution.pipelineName}</strong>
                    <span style={{ marginLeft: '12px', fontSize: '12px', color: '#888' }}>
                      {execution.id}
                    </span>
                  </div>
                  <div>
                    <Button
                      size="small"
                      icon={<PauseCircleOutlined />}
                      onClick={() => handlePauseExecution(execution.id)}
                      style={{ marginRight: '8px' }}
                    >
                      暂停
                    </Button>
                    <Button
                      size="small"
                      danger
                      icon={<StopOutlined />}
                      onClick={() => handleCancelExecution(execution.id)}
                    >
                      取消
                    </Button>
                  </div>
                </div>
                <Progress
                  percent={execution.progress}
                  status="active"
                  strokeColor={{
                    '0%': '#108ee9',
                    '100%': '#87d068',
                  }}
                />
                <div style={{ marginTop: '8px', display: 'flex', justifyContent: 'space-between', fontSize: '12px' }}>
                  <span>{execution.currentStep}</span>
                  <span>{execution.progress}%</span>
                </div>
                <div style={{ marginTop: '4px', fontSize: '12px', color: '#888' }}>
                  开始: {new Date(execution.startedAt).toLocaleString()} | 耗时: {calculateDuration(execution.startedAt)}
                </div>
                {execution.error && (
                  <div style={{ marginTop: '8px', color: '#ff4d4f', fontSize: '12px' }}>
                    错误: {execution.error}
                  </div>
                )}
              </Card>
            ))}
          </div>
        )}
      </Card>

      {/* 最近完成 */}
      <Card title="最近完成">
        {recentExecutions.length === 0 ? (
          <p style={{ textAlign: 'center', color: '#888' }}>暂无已完成的任务</p>
        ) : (
          <div style={{ display: 'grid', gap: '12px' }}>
            {recentExecutions.slice(0, 10).map((execution) => (
              <Card key={execution.id} size="small" hoverable onClick={() => handleViewDetail(execution)}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <strong>{execution.pipelineName}</strong>
                    <span style={{ marginLeft: '12px', fontSize: '12px', color: '#888' }}>
                      {execution.id}
                    </span>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
                    {getStatusBadge(execution.status)}
                    <span style={{ fontSize: '12px', color: '#888' }}>
                      耗时: {calculateDuration(execution.startedAt, execution.finishedAt)}
                    </span>
                    <span style={{ fontSize: '12px', color: '#888' }}>
                      {execution.finishedAt && new Date(execution.finishedAt).toLocaleString()}
                    </span>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </Card>

      {/* 执行详情模态框 */}
      <Modal
        title={`执行详情: ${selectedExecution?.id}`}
        open={isDetailModalVisible}
        onCancel={() => setIsDetailModalVisible(false)}
        footer={null}
        width={800}
      >
        {selectedExecution && (
          <div>
            <Card title="基本信息" size="small" style={{ marginBottom: '16px' }}>
              <p><strong>流水线:</strong> {selectedExecution.pipelineName}</p>
              <p><strong>状态:</strong> {getStatusBadge(selectedExecution.status)}</p>
              <p><strong>开始时间:</strong> {new Date(selectedExecution.startedAt).toLocaleString()}</p>
              {selectedExecution.finishedAt && (
                <p><strong>完成时间:</strong> {new Date(selectedExecution.finishedAt).toLocaleString()}</p>
              )}
              <p><strong>耗时:</strong> {calculateDuration(selectedExecution.startedAt, selectedExecution.finishedAt)}</p>
              {selectedExecution.error && (
                <p><strong>错误:</strong> <span style={{ color: '#ff4d4f' }}>{selectedExecution.error}</span></p>
              )}
            </Card>

            <Card title="步骤执行进度" size="small">
              <Timeline>
                {selectedExecution.steps.map((step, index) => (
                  <Timeline.Item
                    key={step.stepId}
                    color={
                      step.status === 'completed' ? 'green' :
                      step.status === 'running' ? 'blue' :
                      step.status === 'failed' ? 'red' : 'gray'
                    }
                  >
                    <div>
                      <strong>步骤 {index + 1}: {step.name}</strong>
                      {getStatusBadge(step.status)}
                    </div>
                    {step.status === 'running' && (
                      <Progress percent={step.progress} size="small" style={{ marginTop: '8px' }} />
                    )}
                    <div style={{ fontSize: '12px', color: '#888', marginTop: '4px' }}>
                      开始: {new Date(step.startedAt).toLocaleString()}
                      {step.finishedAt && ` | 完成: ${new Date(step.finishedAt).toLocaleString()}`}
                    </div>
                    {step.error && (
                      <div style={{ color: '#ff4d4f', fontSize: '12px', marginTop: '4px' }}>
                        错误: {step.error}
                      </div>
                    )}
                  </Timeline.Item>
                ))}
              </Timeline>
            </Card>
          </div>
        )}
      </Modal>
    </div>
  );
};

export default MonitoringDashboard;

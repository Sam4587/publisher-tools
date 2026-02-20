import React, { useState, useEffect } from 'react';
import { Card, Button, Badge, Progress, Modal, Form, Input, Select, message } from 'antd';

interface Pipeline {
  id: string;
  name: string;
  description: string;
  status: 'draft' | 'active' | 'running' | 'completed' | 'failed' | 'paused';
  steps: PipelineStep[];
  stats: {
    totalExecutions: number;
    successRate: number;
    avgDuration: number;
  };
  createdAt: string;
}

interface PipelineStep {
  id: string;
  name: string;
  type: string;
  handler: string;
}

interface Execution {
  id: string;
  pipelineId: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'paused' | 'cancelled';
  progress: number;
  currentStep: string;
  totalSteps: number;
  startedAt: string;
  finishedAt?: string;
  error?: string;
}

const PipelineManagement: React.FC = () => {
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [selectedPipeline, setSelectedPipeline] = useState<Pipeline | null>(null);
  const [isCreateModalVisible, setIsCreateModalVisible] = useState(false);
  const [isExecuteModalVisible, setIsExecuteModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [executeForm] = Form.useForm();
  const [ws, setWs] = useState<WebSocket | null>(null);

  // WebSocket 连接
  useEffect(() => {
    const websocket = new WebSocket('ws://localhost:8080/ws/monitor');

    websocket.onopen = () => {
      console.log('WebSocket connected');
      message.success('实时监控已连接');
    };

    websocket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      handleRealtimeUpdate(message);
    };

    websocket.onerror = (error) => {
      console.error('WebSocket error:', error);
      message.error('实时监控连接失败');
    };

    websocket.onclose = () => {
      console.log('WebSocket disconnected');
      message.warning('实时监控已断开');
    };

    setWs(websocket);

    return () => {
      websocket.close();
    };
  }, []);

  // 加载流水线列表
  useEffect(() => {
    loadPipelines();
    loadExecutions();
  }, []);

  const loadPipelines = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/pipelines');
      const data = await response.json();
      setPipelines(data);
    } catch (error) {
      message.error('加载流水线失败');
    }
  };

  const loadExecutions = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/executions?limit=10');
      const data = await response.json();
      setExecutions(data);
    } catch (error) {
      message.error('加载执行记录失败');
    }
  };

  const handleRealtimeUpdate = (message: any) => {
    switch (message.type) {
      case 'progress':
        updateExecutionProgress(message.data);
        break;
      case 'status_change':
        updateExecutionStatus(message.data);
        break;
      case 'error':
        message.error(`执行错误: ${message.data.error}`);
        break;
      case 'completed':
        message.success(`执行完成: ${message.data.execution_id}`);
        loadExecutions();
        break;
    }
  };

  const updateExecutionProgress = (data: any) => {
    setExecutions(prev => prev.map(exec =>
      exec.id === data.execution_id ? { ...exec, progress: data.progress, currentStep: data.current_step } : exec
    ));
  };

  const updateExecutionStatus = (data: any) => {
    setExecutions(prev => prev.map(exec =>
      exec.id === data.execution_id ? { ...exec, status: data.status } : exec
    ));
  };

  const handleCreatePipeline = async (values: any) => {
    try {
      const response = await fetch('http://localhost:8080/api/v1/pipelines', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(values),
      });

      if (response.ok) {
        message.success('流水线创建成功');
        setIsCreateModalVisible(false);
        form.resetFields();
        loadPipelines();
      } else {
        message.error('流水线创建失败');
      }
    } catch (error) {
      message.error('创建失败');
    }
  };

  const handleExecutePipeline = async (values: any) => {
    if (!selectedPipeline) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/pipelines/${selectedPipeline.id}/execute`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ input: values }),
      });

      if (response.ok) {
        const data = await response.json();
        message.success('流水线已启动');
        setIsExecuteModalVisible(false);
        executeForm.resetFields();
        loadExecutions();

        // 订阅执行进度
        if (ws) {
          ws.send(JSON.stringify({
            type: 'subscribe',
            topics: [`execution:${data.execution_id}`]
          }));
        }
      } else {
        message.error('启动失败');
      }
    } catch (error) {
      message.error('启动失败');
    }
  };

  const handlePauseExecution = async (executionId: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/executions/${executionId}/pause`, {
        method: 'POST',
      });

      if (response.ok) {
        message.success('已暂停');
        loadExecutions();
      }
    } catch (error) {
      message.error('暂停失败');
    }
  };

  const handleResumeExecution = async (executionId: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/executions/${executionId}/resume`, {
        method: 'POST',
      });

      if (response.ok) {
        message.success('已恢复');
        loadExecutions();
      }
    } catch (error) {
      message.error('恢复失败');
    }
  };

  const handleCancelExecution = async (executionId: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/v1/executions/${executionId}/cancel`, {
        method: 'POST',
      });

      if (response.ok) {
        message.success('已取消');
        loadExecutions();
      }
    } catch (error) {
      message.error('取消失败');
    }
  };

  const getStatusBadge = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      draft: { color: 'default', text: '草稿' },
      active: { color: 'success', text: '活跃' },
      running: { color: 'processing', text: '运行中' },
      completed: { color: 'success', text: '已完成' },
      failed: { color: 'error', text: '失败' },
      paused: { color: 'warning', text: '已暂停' },
    };
    const { color, text } = statusMap[status] || { color: 'default', text: status };
    return <Badge color={color} text={text} />;
  };

  return (
    <div style={{ padding: '24px' }}>
      <div style={{ marginBottom: '24px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1>流水线管理</h1>
        <Button type="primary" onClick={() => setIsCreateModalVisible(true)}>
          + 创建流水线
        </Button>
      </div>

      {/* 流水线列表 */}
      <div style={{ marginBottom: '32px' }}>
        <h2>流水线列表</h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: '16px' }}>
          {pipelines.map((pipeline) => (
            <Card
              key={pipeline.id}
              title={pipeline.name}
              extra={getStatusBadge(pipeline.status)}
              actions={[
                <Button key="execute" type="primary" onClick={() => { setSelectedPipeline(pipeline); setIsExecuteModalVisible(true); }}>
                  立即执行
                </Button>,
                <Button key="history" onClick={() => console.log('查看历史')}>
                  查看历史
                </Button>,
              ]}
            >
              <p>{pipeline.description}</p>
              <div style={{ marginTop: '16px' }}>
                <p>执行次数: {pipeline.stats.totalExecutions}</p>
                <p>成功率: {(pipeline.stats.successRate * 100).toFixed(1)}%</p>
                <p>平均耗时: {pipeline.stats.avgDuration}s</p>
                <p>步骤数: {pipeline.steps.length}</p>
              </div>
            </Card>
          ))}
        </div>
      </div>

      {/* 执行中任务 */}
      <div>
        <h2>执行中任务</h2>
        {executions.filter(e => e.status === 'running').length === 0 ? (
          <p>暂无执行中的任务</p>
        ) : (
          <div style={{ display: 'grid', gap: '16px' }}>
            {executions.filter(e => e.status === 'running').map((execution) => (
              <Card key={execution.id}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                  <strong>执行ID: {execution.id}</strong>
                  <div>
                    <Button size="small" onClick={() => handlePauseExecution(execution.id)} style={{ marginRight: '8px' }}>
                      暂停
                    </Button>
                    <Button size="small" danger onClick={() => handleCancelExecution(execution.id)}>
                      取消
                    </Button>
                  </div>
                </div>
                <Progress percent={execution.progress} status="active" />
                <p style={{ marginTop: '8px' }}>
                  {execution.currentStep} ({execution.progress}%)
                </p>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* 最近完成 */}
      <div style={{ marginTop: '32px' }}>
        <h2>最近完成</h2>
        {executions.filter(e => e.status === 'completed').length === 0 ? (
          <p>暂无已完成的任务</p>
        ) : (
          <div style={{ display: 'grid', gap: '16px' }}>
            {executions.filter(e => e.status === 'completed').slice(0, 5).map((execution) => (
              <Card key={execution.id}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <div>
                    <strong>执行ID: {execution.id}</strong>
                    {getStatusBadge(execution.status)}
                  </div>
                  <span>{new Date(execution.finishedAt!).toLocaleString()}</span>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* 创建流水线模态框 */}
      <Modal
        title="创建流水线"
        open={isCreateModalVisible}
        onCancel={() => setIsCreateModalVisible(false)}
        footer={null}
      >
        <Form form={form} layout="vertical" onFinish={handleCreatePipeline}>
          <Form.Item
            name="template_id"
            label="选择模板"
            rules={[{ required: true, message: '请选择模板' }]}
          >
            <Select>
              <Select.Option value="content-publish-v1">内容发布流水线</Select.Option>
              <Select.Option value="video-processing-v1">视频处理流水线</Select.Option>
              <Select.Option value="hotspot-analysis-v1">热点分析流水线</Select.Option>
              <Select.Option value="data-collection-v1">数据采集流水线</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="name"
            label="流水线名称"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder="输入流水线名称" />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea placeholder="输入描述" rows={3} />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              创建
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      {/* 执行流水线模态框 */}
      <Modal
        title={`执行流水线: ${selectedPipeline?.name}`}
        open={isExecuteModalVisible}
        onCancel={() => setIsExecuteModalVisible(false)}
        footer={null}
      >
        <Form form={executeForm} layout="vertical" onFinish={handleExecutePipeline}>
          <Form.Item
            name="topic"
            label="内容主题"
            rules={[{ required: true, message: '请输入主题' }]}
          >
            <Input placeholder="输入内容主题" />
          </Form.Item>
          <Form.Item
            name="keywords"
            label="关键词"
          >
            <Select mode="tags" placeholder="输入关键词" />
          </Form.Item>
          <Form.Item
            name="platforms"
            label="发布平台"
            rules={[{ required: true, message: '请选择平台' }]}
          >
            <Select mode="multiple" placeholder="选择发布平台">
              <Select.Option value="douyin">抖音</Select.Option>
              <Select.Option value="toutiao">今日头条</Select.Option>
              <Select.Option value="xiaohongshu">小红书</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              开始执行
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default PipelineManagement;
